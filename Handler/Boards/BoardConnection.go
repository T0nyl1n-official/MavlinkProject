package boardHandler

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	Board "MavlinkProject_Board/Shared/Boards"
)

type BoardConnectionManager struct {
	boards      map[string]*BoardServer
	tcpConns    map[string]net.Conn
	udpAddrs    map[string]*net.UDPAddr
	mu          sync.RWMutex
	messageChan chan *Board.BoardMessage
	running     bool
	stopChan    chan bool
}

type BoardServer struct {
	boardID     string
	listener    net.Listener
	udpConn     *net.UDPConn
	connection  string
	protocol    string
	addr        string
	port        string
	messageChan chan *Board.BoardMessage
	stopChan    chan bool
	wg          sync.WaitGroup
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

	bs := &BoardServer{
		boardID:     boardID,
		listener:    listener,
		connection:  Board.Connection_TCP,
		addr:        addr,
		port:        port,
		messageChan: bm.messageChan,
		stopChan:    make(chan bool),
	}

	bm.boards[boardID] = bs
	bs.wg.Add(1)
	go bs.acceptConnections()

	log.Printf("[BoardManager] TCP server started for board %s on %s:%s", boardID, addr, port)
	return nil
}

func (bm *BoardConnectionManager) ForwardMessageToBoard(fromBoardID, toBoardID string, msg *Board.BoardMessage) error {
	msg.FromID = fromBoardID
	msg.ToID = toBoardID

	if err := bm.SendMessageToBoard(toBoardID, msg); err != nil {
		return fmt.Errorf("failed to forward message from %s to %s: %v", fromBoardID, toBoardID, err)
	}

	log.Printf("[BoardManager] Forwarded message from board %s to board %s", fromBoardID, toBoardID)
	return nil
}

func (bm *BoardConnectionManager) GetBoardConnectionInfo(boardID string) (connection string, addr string, port string, connected bool) {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	if bs, exists := bm.boards[boardID]; exists {
		connection = bs.connection
		addr = bs.addr
		port = bs.port
		connected = true
	}

	return
}

func (bm *BoardConnectionManager) GetAllBoards() []BoardServerInfo {
	bm.mu.RLock()
	defer bm.mu.RUnlock()

	var boards []BoardServerInfo
	for boardID, bs := range bm.boards {
		boards = append(boards, BoardServerInfo{
			BoardID:    boardID,
			Connection: bs.connection,
			Addr:       bs.addr,
			Port:       bs.port,
			Connected:  true,
		})
	}
	return boards
}

type BoardServerInfo struct {
	BoardID    string `json:"board_id"`
	Connection string `json:"connection"`
	Addr       string `json:"addr"`
	Port       string `json:"port"`
	Connected  bool   `json:"connected"`
}

func (bm *BoardConnectionManager) StartUDPServer(boardID, addr string, port string) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if _, exists := bm.boards[boardID]; exists {
		return fmt.Errorf("board %s already running", boardID)
	}

	addrUDP, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", addr, port))
	if err != nil {
		return fmt.Errorf("failed to resolve UDP address: %v", err)
	}

	listener, err := net.ListenUDP("udp", addrUDP)
	if err != nil {
		return fmt.Errorf("failed to start UDP server: %v", err)
	}

	bs := &BoardServer{
		boardID:     boardID,
		listener:    nil,
		udpConn:     listener,
		connection:  Board.Connection_UDP,
		addr:        addr,
		port:        port,
		messageChan: bm.messageChan,
		stopChan:    make(chan bool),
	}

	bm.boards[boardID] = bs
	bs.wg.Add(1)
	go bs.readUDPPackets(listener)

	log.Printf("[BoardManager] UDP server started for board %s on %s:%s", boardID, addr, port)
	return nil
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

func (bs *BoardServer) handleConnection(conn net.Conn) {
	defer bs.wg.Done()
	defer conn.Close()

	manager := GetBoardManager()
	manager.mu.Lock()
	manager.tcpConns[bs.boardID] = conn
	manager.mu.Unlock()

	buffer := make([]byte, 4096)
	for {
		select {
		case <-bs.stopChan:
			return
		default:
		}

		conn.SetDeadline(time.Now().Add(30 * time.Second))
		n, err := conn.Read(buffer)
		if err != nil {
			log.Printf("[BoardServer] Connection closed for board %s: %v", bs.boardID, err)
			manager.mu.Lock()
			delete(manager.tcpConns, bs.boardID)
			manager.mu.Unlock()
			return
		}

		bs.processMessage(buffer[:n])
	}
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

		manager.mu.Lock()
		manager.udpAddrs[bs.boardID] = addr
		manager.mu.Unlock()

		log.Printf("[BoardServer] Received UDP packet from %s for board %s", addr.String(), bs.boardID)
		bs.processMessage(buffer[:n])
	}
}

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
