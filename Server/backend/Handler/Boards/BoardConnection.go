package boardHandler

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	WarningHandler "MavlinkProject/Server/Backend/Utils/WarningHandle"
	Conf "MavlinkProject/Server/backend/Config"
	Board "MavlinkProject/Server/backend/Shared/Boards"
)

func GetBoardConnectionConfig() (keepAliveInterval, connectionTimeout time.Duration, maxRetryAttempts int, retryDelay time.Duration) {
	setting := Conf.GetSetting()
	cfg := setting.Board.Connection

	keepAliveInterval = time.Duration(cfg.KeepaliveInterval) * time.Second
	connectionTimeout = time.Duration(cfg.Timeout) * time.Second
	maxRetryAttempts = cfg.MaxRetryAttempts
	retryDelay = time.Duration(cfg.RetryDelay) * time.Second

	return
}

// BoardConnectionManager 板连接-通道管理器 (总控管理)
type BoardConnectionManager struct {
	boards      map[string]*BoardServer
	tcpConns    map[string]net.Conn
	udpAddrs    map[string]*net.UDPAddr
	mu          sync.RWMutex
	messageChan chan *Board.BoardMessage
	running     bool
	stopChan    chan bool
}

// BoardServer 后端-主控板服务器
type BoardServer struct {
	boardID           string
	listener          net.Listener
	udpConn           *net.UDPConn
	connection        string
	protocol          string
	addr              string
	port              string
	messageChan       chan *Board.BoardMessage
	stopChan          chan bool
	wg                sync.WaitGroup
	keepAliveInterval time.Duration
	connectionTimeout time.Duration
	maxRetryAttempts  int
	retryDelay        time.Duration
}

// 心跳包
type Message_HeartBeat struct {
	Type      string    `json:"type"`
	BoardID   string    `json:"board_id"`
	Timestamp time.Time `json:"timestamp"`
}

// 全局变量 定义
var (
	// 总控管理
	boardManager *BoardConnectionManager
	boardMgrOnce sync.Once
)

// 总控获取函数 生成Default
func GetBoardManager() *BoardConnectionManager {
	boardMgrOnce.Do(func() {
		boardManager = &BoardConnectionManager{
			boards:      make(map[string]*BoardServer),
			tcpConns:    make(map[string]net.Conn),
			udpAddrs:    make(map[string]*net.UDPAddr),
			messageChan: make(chan *Board.BoardMessage, 1000),
			stopChan:    make(chan bool),
		}
	})
	return boardManager
}

// 开启TCP服务器 后端-板连接
func (bm *BoardConnectionManager) StartTCPServer(boardID, addr string, port string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.boards[boardID]; exists {
		return fmt.Errorf("board %s already running", boardID)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%s", addr, port))
	if err != nil {
		return fmt.Errorf("failed to start TCP server: %v", err)
	}

	keepAliveInterval, connectionTimeout, maxRetryAttempts, retryDelay := GetBoardConnectionConfig()

	bs := &BoardServer{
		boardID:           boardID,
		listener:          listener,
		connection:        Board.Connection_TCP,
		addr:              addr,
		port:              port,
		messageChan:       bm.messageChan,
		stopChan:          make(chan bool),
		keepAliveInterval: keepAliveInterval,
		connectionTimeout: connectionTimeout,
		maxRetryAttempts:  maxRetryAttempts,
		retryDelay:        retryDelay,
	}

	bm.boards[boardID] = bs
	bs.wg.Add(1)
	go bs.acceptConnections()

	log.Printf("[BoardManager] TCP server started for board %s on %s:%s", boardID, addr, port)
	return nil
}

// 开启UDP服务器 后端-板连接
func (bm *BoardConnectionManager) StartUDPServer(boardID, addr string, port string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.boards[boardID]; exists {
		return fmt.Errorf("board %s already running", boardID)
	}

	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", addr, port))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	listener, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		return fmt.Errorf("failed to start UDP server: %v", err)
	}

	keepAliveInterval, connectionTimeout, maxRetryAttempts, retryDelay := GetBoardConnectionConfig()

	bs := &BoardServer{
		boardID:           boardID,
		udpConn:           listener,
		connection:        Board.Connection_UDP,
		addr:              addr,
		port:              port,
		messageChan:       bm.messageChan,
		stopChan:          make(chan bool),
		keepAliveInterval: keepAliveInterval,
		connectionTimeout: connectionTimeout,
		maxRetryAttempts:  maxRetryAttempts,
		retryDelay:        retryDelay,
	}

	bm.boards[boardID] = bs
	bs.wg.Add(1)
	go bs.readUDPPackets(listener)

	log.Printf("[BoardManager] UDP server started for board %s on %s:%s", boardID, addr, port)
	return nil
}

// BoardServerInfo 板服务器信息
type BoardServerInfo struct {
	BoardID    string
	Addr       string
	Port       string
	Connection string
	Connected  bool
}

// 获取所有板信息
func (bm *BoardConnectionManager) GetAllBoards() []BoardServerInfo {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	boardList := make([]BoardServerInfo, 0, len(bm.boards))
	for boardID, bs := range bm.boards {
		_, connExists := bm.tcpConns[boardID]
		_, udpExists := bm.udpAddrs[boardID]

		boardList = append(boardList, BoardServerInfo{
			BoardID:    boardID,
			Addr:       bs.addr,
			Port:       bs.port,
			Connection: bs.connection,
			Connected:  connExists || udpExists,
		})
	}
	return boardList
}

// 前馈消息到指定板
func (bm *BoardConnectionManager) ForwardMessageToBoard(fromBoardID, toBoardID string, msg *Board.BoardMessage) error {
	msg.FromID = fromBoardID
	msg.ToID = toBoardID

	if err := bm.SendMessageToBoard(toBoardID, msg); err != nil {
		return fmt.Errorf("failed to forward message from %s to %s: %v", fromBoardID, toBoardID, err)
	}

	log.Printf("[BoardManager] Forwarded message from board %s to board %s", fromBoardID, toBoardID)
	return nil
}

// 获取指定连接信息
func (bm *BoardConnectionManager) GetBoardConnectionInfo(boardID string) (connection string, addr string, port string, connected bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	bs, exists := bm.boards[boardID]
	if !exists {
		return "", "", "", false
	}

	_, connExists := bm.tcpConns[boardID]
	_, udpExists := bm.udpAddrs[boardID]

	return bs.connection, bs.addr, bs.port, connExists || udpExists
}

// 接受连接, 获取信息函数(子函数)
func (bs *BoardServer) acceptConnections() {
	defer bs.wg.Done()

	for {
		select {
		case <-bs.stopChan:
			return
		default:
		}

		bs.listener.(*net.TCPListener).SetDeadline(time.Now().Add(1 * time.Second))
		conn, err := bs.listener.Accept()
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return
		}

		bs.wg.Add(1)
		go bs.handleConnection(conn)
	}
}

// 心跳包发送函数 获取信息函数(子函数)
func (bs *BoardServer) keepAliveWriter(conn net.Conn, stopChan chan bool) {
	ticker := time.NewTicker(bs.keepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			keepAliveMsg := Message_HeartBeat{
				Type:      "keepalive",
				BoardID:   bs.boardID,
				Timestamp: time.Now(),
			}

			data, err := json.Marshal(keepAliveMsg)
			if err != nil {
				log.Printf("[BoardServer] Failed to marshal keepalive message: %v", err)
				continue
			}

			_, err = conn.Write(data)
			if err != nil {
				log.Printf("[BoardServer] Failed to send keepalive to board %s: %v", bs.boardID, err)
				return
			}

			log.Printf("[BoardServer] Sent keepalive to board %s", bs.boardID)
		}
	}
}

// 处理连接函数(imp), 获取信息函数(子函数)
func (bs *BoardServer) handleConnection(conn net.Conn) {
	defer bs.wg.Done()
	defer conn.Close()

	manager := GetBoardManager()
	manager.mu.Lock()
	manager.tcpConns[bs.boardID] = conn
	manager.mu.Unlock()

	retryCount := 0

	keepAliveStop := make(chan bool)
	go bs.keepAliveWriter(conn, keepAliveStop)

	defer func() {
		close(keepAliveStop)
	}()

	buffer := make([]byte, 4096)
	for {
		select {
		case <-bs.stopChan:
			return
		default:
		}

		conn.SetDeadline(time.Now().Add(bs.connectionTimeout))
		n, err := conn.Read(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("[BoardServer] Connection timeout for board %s, retrying...", bs.boardID)

				for retryCount < bs.maxRetryAttempts {
					retryCount++
					log.Printf("[BoardServer] Reconnection attempt %d/%d for board %s", retryCount, bs.maxRetryAttempts, bs.boardID)

					time.Sleep(bs.retryDelay)

					newConn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", bs.addr, bs.port), bs.connectionTimeout)
					if err != nil {
						log.Printf("[BoardServer] Reconnection failed for board %s: %v", bs.boardID, err)
						continue
					}

					manager.mu.Lock()
					manager.tcpConns[bs.boardID] = newConn
					manager.mu.Unlock()
					conn = newConn
					log.Printf("[BoardServer] Reconnection successful for board %s", bs.boardID)
					break
				}

				if retryCount >= bs.maxRetryAttempts {
					log.Printf("[BoardServer] Max retry attempts reached for board %s, reporting to WarningHandler", bs.boardID)
					WarningHandler.HandleBackendError(
						fmt.Sprintf("Board connection lost after %d reconnection attempts", bs.maxRetryAttempts),
						"BoardConnection",
						fmt.Sprintf("board_id=%s, addr=%s:%s", bs.boardID, bs.addr, bs.port),
					)

					manager.mu.Lock()
					delete(manager.tcpConns, bs.boardID)
					manager.mu.Unlock()
					return
				}
				continue
			}

			log.Printf("[BoardServer] Connection closed for board %s: %v", bs.boardID, err)
			manager.mu.Lock()
			delete(manager.tcpConns, bs.boardID)
			manager.mu.Unlock()
			return
		}

		bs.processMessage(buffer[:n])
		conn.SetDeadline(time.Now().Add(bs.connectionTimeout))
	}
}

// 读取UDP数据包
func (bs *BoardServer) readUDPPackets(listener *net.UDPConn) {
	defer bs.wg.Done()

	manager := GetBoardManager()

	buffer := make([]byte, 4096)
	for {
		select {
		case <-bs.stopChan:
			return
		default:
		}

		listener.SetDeadline(time.Now().Add(1 * time.Second))
		n, addr, err := listener.ReadFromUDP(buffer)
		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}
			return
		}

		manager.mu.Lock()
		manager.udpAddrs[bs.boardID] = addr
		manager.mu.Unlock()

		log.Printf("[BoardServer] Received UDP packet from %s for board %s", addr.String(), bs.boardID)
		bs.processMessage(buffer[:n])
	}
}

// 处理 BoardMessage.Message.Data (any)
func (bs *BoardServer) processMessage(data []byte) {
	var msg Board.BoardMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[BoardServer] Failed to unmarshal message from board %s: %v", bs.boardID, err)
		return
	}

	msg.ToID = bs.boardID
	msg.ToType = string(Board.Type_Control)

	bs.messageChan <- &msg
}

// 停止指定板连接
func (bm *BoardConnectionManager) StopBoard(boardID string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	bs, exists := bm.boards[boardID]
	if !exists {
		return fmt.Errorf("board %s not found", boardID)
	}

	close(bs.stopChan)
	bs.wg.Wait()

	if bs.listener != nil {
		bs.listener.Close()
	}

	delete(bm.boards, boardID)
	log.Printf("[BoardManager] Stopped board %s", boardID)
	return nil
}

// 停止所有板连接 封装函数
func (bm *BoardConnectionManager) StopAll() {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	for boardID := range bm.boards {
		bs := bm.boards[boardID]
		close(bs.stopChan)
		if bs.listener != nil {
			bs.listener.Close()
		}
		delete(bm.boards, boardID)
	}

	bm.running = false
	close(bm.stopChan)
	log.Printf("[BoardManager] All boards stopped")
}

// 获取消息通道
func (bm *BoardConnectionManager) GetMessageChan() chan *Board.BoardMessage {
	return bm.messageChan
}

// 发送消息到指定板 (子函数)
func (bm *BoardConnectionManager) SendMessageToBoard(boardID string, msg *Board.BoardMessage) error {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	bs, exists := bm.boards[boardID]
	if !exists {
		return fmt.Errorf("board %s not found", boardID)
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	if bs.connection == Board.Connection_TCP {
		conn, connExists := bm.tcpConns[boardID]
		if !connExists || conn == nil {
			return fmt.Errorf("TCP connection for board %s not available", boardID)
		}
		_, err = conn.Write(data)
		if err != nil {
			return fmt.Errorf("failed to send TCP message to board %s: %v", boardID, err)
		}
		log.Printf("[BoardManager] Sent TCP message to board %s: %d bytes", boardID, len(data))
	} else if bs.connection == Board.Connection_UDP {
		addr, addrExists := bm.udpAddrs[boardID]
		if !addrExists || addr == nil {
			return fmt.Errorf("UDP address for board %s not available", boardID)
		}
		_, err = bs.udpConn.WriteToUDP(data, addr)
		if err != nil {
			return fmt.Errorf("failed to send UDP message to board %s: %v", boardID, err)
		}
		log.Printf("[BoardManager] Sent UDP message to board %s: %d bytes", boardID, len(data))
	}

	return nil
}

// 启用自动转发 (封装函数)
func (bm *BoardConnectionManager) EnableAutoForward() {
	go func() {
		for msg := range bm.messageChan {
			if msg.ToID != "" && msg.ToID != "Control" {
				if err := bm.SendMessageToBoard(msg.ToID, msg); err != nil {
					log.Printf("[BoardManager] Auto-forward failed: %v", err)
				} else {
					log.Printf("[BoardManager] Auto-forwarded message to %s", msg.ToID)
				}
			}
		}
	}()
	log.Printf("[BoardManager] Auto-forward enabled")
}
