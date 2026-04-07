package boardHandler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	WarningHandler "MavlinkProject/Server/backend/Utils/WarningHandle"
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

	if keepAliveInterval <= 0 {
		keepAliveInterval = 30 * time.Second
	}
	if connectionTimeout <= 0 {
		connectionTimeout = 10 * time.Second
	}
	if maxRetryAttempts <= 0 {
		maxRetryAttempts = 3
	}
	if retryDelay <= 0 {
		retryDelay = 5 * time.Second
	}

	return
}

type BoardConnectionManager struct {
	boards      map[string]*BoardServer
	tcpConns    map[string]net.Conn
	udpAddrs    map[string]*net.UDPAddr
	fromIDToKey map[string]string
	mu          sync.RWMutex
	messageChan chan *Board.BoardMessage
	running     bool
	stopChan    chan bool
}

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

type Message_HeartBeat struct {
	Type      string    `json:"type"`
	BoardID   string    `json:"board_id"`
	Timestamp time.Time `json:"timestamp"`
}

var (
	boardManager *BoardConnectionManager
	boardMgrOnce sync.Once
)

func GetBoardManager() *BoardConnectionManager {
	boardMgrOnce.Do(func() {
		boardManager = &BoardConnectionManager{
			boards:      make(map[string]*BoardServer),
			tcpConns:    make(map[string]net.Conn),
			udpAddrs:    make(map[string]*net.UDPAddr),
			fromIDToKey: make(map[string]string),
			messageChan: make(chan *Board.BoardMessage, 1000),
			stopChan:    make(chan bool),
		}
	})
	return boardManager
}

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

func (bm *BoardConnectionManager) RegisterBoardFromID(fromID, connKey string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.fromIDToKey[fromID] = connKey
}

func (bm *BoardConnectionManager) GetConnKeyByFromID(fromID string) (string, bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	key, ok := bm.fromIDToKey[fromID]
	return key, ok
}

func (bm *BoardConnectionManager) SendMessageToBoardByFromID(fromID string, msg *Board.BoardMessage) error {
	bm.mu.RLock()
	key, ok := bm.fromIDToKey[fromID]
	bm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("board %s not registered", fromID)
	}

	return bm.SendMessageToBoard(key, msg)
}

func (bm *BoardConnectionManager) SendMessageToBoard(boardID string, msg *Board.BoardMessage) error {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %v", err)
	}

	if conn, ok := bm.tcpConns[boardID]; ok && conn != nil {
		_, err = conn.Write(data)
		if err != nil {
			return fmt.Errorf("failed to send TCP message to board %s: %v", boardID, err)
		}
		log.Printf("[BoardManager] Sent TCP message to %s: %d bytes", boardID, len(data))
		return nil
	}

	if addr, ok := bm.udpAddrs[boardID]; ok && addr != nil {
		bs := bm.boards["udp_board_listener"]
		if bs != nil && bs.udpConn != nil {
			_, err = bs.udpConn.WriteToUDP(data, addr)
			if err != nil {
				return fmt.Errorf("failed to send UDP message to board %s: %v", boardID, err)
			}
			log.Printf("[BoardManager] Sent UDP message to %s: %d bytes", boardID, len(data))
			return nil
		}
	}

	return fmt.Errorf("board %s connection not found", boardID)
}

func (bm *BoardConnectionManager) RemoveBoardConnection(boardID string) {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if conn, ok := bm.tcpConns[boardID]; ok && conn != nil {
		conn.Close()
	}
	delete(bm.tcpConns, boardID)
	delete(bm.udpAddrs, boardID)

	for fromID, key := range bm.fromIDToKey {
		if key == boardID {
			delete(bm.fromIDToKey, fromID)
		}
	}
}

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

func (bs *BoardServer) keepAliveWriter(conn net.Conn, boardKey string, stopChan chan bool) {
	ticker := time.NewTicker(bs.keepAliveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-stopChan:
			return
		case <-ticker.C:
			keepAliveMsg := Message_HeartBeat{
				Type:      "keepalive",
				BoardID:   boardKey,
				Timestamp: time.Now(),
			}

			data, err := json.Marshal(keepAliveMsg)
			if err != nil {
				log.Printf("[BoardServer] Failed to marshal keepalive message: %v", err)
				continue
			}

			_, err = conn.Write(data)
			if err != nil {
				log.Printf("[BoardServer] Failed to send keepalive to %s: %v", boardKey, err)
				return
			}

			log.Printf("[BoardServer] Sent keepalive to %s", boardKey)
		}
	}
}

func (bs *BoardServer) handleConnection(conn net.Conn) {
	defer bs.wg.Done()
	defer conn.Close()

	manager := GetBoardManager()
	connKey := conn.RemoteAddr().String()

	manager.mu.Lock()
	manager.tcpConns[connKey] = conn
	manager.mu.Unlock()

	retryCount := 0

	keepAliveStop := make(chan bool)
	go bs.keepAliveWriter(conn, connKey, keepAliveStop)

	defer func() {
		close(keepAliveStop)
	}()

	defer conn.Close()

	reader := bufio.NewReader(conn)
	buffer := make([]byte, 0, 4096)

	for {
		select {
		case <-bs.stopChan:
			return
		default:
		}

		conn.SetDeadline(time.Now().Add(bs.connectionTimeout))

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF || strings.Contains(err.Error(), "EOF") {
				log.Printf("[BoardServer] Connection closed normally for %s", connKey)
				return
			} else if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				log.Printf("[BoardServer] Connection timeout for %s, retrying...", connKey)

				for retryCount < bs.maxRetryAttempts {
					retryCount++
					log.Printf("[BoardServer] Reconnection attempt %d/%d for %s", retryCount, bs.maxRetryAttempts, connKey)

					time.Sleep(bs.retryDelay)

					newConn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%s", bs.addr, bs.port), bs.connectionTimeout)
					if err != nil {
						log.Printf("[BoardServer] Reconnection failed for %s: %v", connKey, err)
						continue
					}

					manager.mu.Lock()
					manager.tcpConns[connKey] = newConn
					manager.mu.Unlock()
					conn = newConn
					reader = bufio.NewReader(conn)
					log.Printf("[BoardServer] Reconnection successful for %s", connKey)
					break
				}

				if retryCount >= bs.maxRetryAttempts {
					log.Printf("[BoardServer] Max retry attempts reached for %s, reporting to WarningHandler", connKey)
					WarningHandler.HandleBackendError(
						fmt.Sprintf("Board connection lost after %d reconnection attempts", bs.maxRetryAttempts),
						"BoardConnection",
						fmt.Sprintf("board_id=%s, addr=%s:%s", connKey, bs.addr, bs.port),
					)
					return
				}
				continue
			}

			log.Printf("[BoardServer] Connection closed for %s: %v", connKey, err)
			return
		}

		lineWithoutNewline := trimNewline(line)
		buffer = append(buffer, lineWithoutNewline...)

		if data, ok := bs.extractMessage(&buffer); ok {
			bs.processMessage(data, connKey)
			buffer = buffer[:0]
		}

		conn.SetDeadline(time.Now().Add(bs.connectionTimeout))
	}
}

func trimNewline(data []byte) []byte {
	for len(data) > 0 && (data[len(data)-1] == '\n' || data[len(data)-1] == '\r') {
		data = data[:len(data)-1]
	}
	return data
}

func (bs *BoardServer) extractMessage(buffer *[]byte) ([]byte, bool) {
	data := *buffer
	if len(data) == 0 {
		return nil, false
	}

	var msg Board.BoardMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, false
	}

	if msg.FromID == "" {
		return nil, false
	}

	return data, true
}

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

		addrKey := addr.String()
		manager.mu.Lock()
		manager.udpAddrs[addrKey] = addr
		manager.mu.Unlock()

		log.Printf("[BoardServer] Received UDP packet from %s", addrKey)
		bs.processMessage(buffer[:n], addrKey)
	}
}

func (bs *BoardServer) processMessage(data []byte, connKey string) {
	var msg Board.BoardMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[BoardServer] Failed to unmarshal message from %s: %v", connKey, err)
		return
	}

	manager := GetBoardManager()

	if msg.FromID != "" {
		manager.RegisterBoardFromID(msg.FromID, connKey)
		log.Printf("[BoardServer] Registered board %s with key %s", msg.FromID, connKey)
	}

	msg.ToID = bs.boardID
	msg.ToType = string(Board.Type_Control)

	bs.messageChan <- &msg
}

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

func (bm *BoardConnectionManager) GetMessageChan() chan *Board.BoardMessage {
	return bm.messageChan
}

func (bm *BoardConnectionManager) GetAllBoards() []BoardServerInfo {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var result []BoardServerInfo
	for _, bs := range bm.boards {
		_, connExists := bm.tcpConns[bs.boardID]
		_, udpExists := bm.udpAddrs[bs.boardID]

		result = append(result, BoardServerInfo{
			BoardID:    bs.boardID,
			Addr:       bs.addr,
			Port:       bs.port,
			Connection: bs.connection,
			Connected:  connExists || udpExists,
		})
	}
	return result
}

func (bm *BoardConnectionManager) ForwardMessageToBoard(fromBoardID, toBoardID string, msg *Board.BoardMessage) error {
	bm.mu.RLock()
	_, fromExists := bm.tcpConns[fromBoardID]
	if !fromExists {
		_, fromExists = bm.udpAddrs[fromBoardID]
	}
	_, toExists := bm.tcpConns[toBoardID]
	if !toExists {
		_, toExists = bm.udpAddrs[toBoardID]
	}
	bm.mu.RUnlock()

	if !fromExists {
		return fmt.Errorf("source board %s not connected", fromBoardID)
	}
	if !toExists {
		return fmt.Errorf("target board %s not connected", toBoardID)
	}

	msg.FromID = fromBoardID
	msg.ToID = toBoardID

	return bm.SendMessageToBoard(toBoardID, msg)
}

type BoardServerInfo struct {
	BoardID    string
	Addr       string
	Port       string
	Connection string
	Connected  bool
}

var autoForwardEnabled = false
var autoForwardMu sync.RWMutex

func (bm *BoardConnectionManager) EnableAutoForward() {
	autoForwardMu.Lock()
	autoForwardEnabled = true
	autoForwardMu.Unlock()
	log.Printf("[BoardManager] Auto-forward enabled")
}

func (bm *BoardConnectionManager) DisableAutoForward() {
	autoForwardMu.Lock()
	autoForwardEnabled = false
	autoForwardMu.Unlock()
	log.Printf("[BoardManager] Auto-forward disabled")
}

func IsAutoForwardEnabled() bool {
	autoForwardMu.RLock()
	defer autoForwardMu.RUnlock()
	return autoForwardEnabled
}
