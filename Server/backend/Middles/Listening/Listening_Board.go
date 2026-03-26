package listening

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/gin-gonic/gin"

	BoardHandler "MavlinkProject/Server/Backend/Handler/Boards"
	Board "MavlinkProject/Server/backend/Shared/Boards"
	BoardClassifier "MavlinkProject/Server/Backend/Utils/BoardClassifier"
)

type BoardListenerConfig struct {
	EnableTCP     bool
	EnableUDP     bool
	TCPAddress    string
	TCPPort       string
	UDPAddress    string
	UDPPort       string
	MaxBufferSize int
}

var defaultBoardListenerConfig = BoardListenerConfig{
	EnableTCP:     true,
	EnableUDP:     true,
	TCPAddress:    "localhost",
	TCPPort:       "14550",
	UDPAddress:    "0.0.0.0",
	UDPPort:       "14551",
	MaxBufferSize: 4096,
}

var (
	boardListener     *BoardListener
	boardListenerOnce sync.Once
	boardManager      *BoardHandler.BoardConnectionManager
	classifier        *BoardClassifier.BoardMessageClassifier
)

func BoardListenerMiddleware() gin.HandlerFunc {
	return BoardListenerWithConfig(defaultBoardListenerConfig)
}

func BoardListenerWithConfig(config BoardListenerConfig) gin.HandlerFunc {
	boardListenerOnce.Do(func() {
		boardManager = BoardHandler.GetBoardManager()
		classifier = BoardClassifier.NewBoardMessageClassifier()

		boardListener = &BoardListener{
			config:     config,
			manager:    boardManager,
			classifier: classifier,
			running:    false,
		}

		log.Printf("[BoardListener] Initialized - TCP: %s:%s, UDP: %s:%s",
			config.TCPAddress, config.TCPPort, config.UDPAddress, config.UDPPort)
	})

	return func(c *gin.Context) {
		c.Next()
	}
}

type BoardListener struct {
	config     BoardListenerConfig
	manager    *BoardHandler.BoardConnectionManager
	classifier *BoardClassifier.BoardMessageClassifier
	running    bool
	mu         sync.RWMutex
}

func (bl *BoardListener) Start() error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if bl.running {
		return fmt.Errorf("board listener already running")
	}

	if bl.config.EnableTCP {
		boardID := "tcp_board_listener"
		if err := bl.manager.StartTCPServer(boardID, bl.config.TCPAddress, bl.config.TCPPort); err != nil {
			log.Printf("[BoardListener] Failed to start TCP server: %v", err)
		} else {
			log.Printf("[BoardListener] TCP server started on %s:%s", bl.config.TCPAddress, bl.config.TCPPort)
		}
	}

	if bl.config.EnableUDP {
		boardID := "udp_board_listener"
		if err := bl.manager.StartUDPServer(boardID, bl.config.UDPAddress, bl.config.UDPPort); err != nil {
			log.Printf("[BoardListener] Failed to start UDP server: %v", err)
		} else {
			log.Printf("[BoardListener] UDP server started on %s:%s", bl.config.UDPAddress, bl.config.UDPPort)
		}
	}

	bl.classifier.StartProcessor(bl.manager.GetMessageChan())

	bl.running = true
	log.Printf("[BoardListener] All listeners started successfully")
	return nil
}

func (bl *BoardListener) Stop() error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if !bl.running {
		return fmt.Errorf("board listener not running")
	}

	bl.manager.StopAll()
	bl.running = false
	log.Printf("[BoardListener] All listeners stopped")
	return nil
}

func (bl *BoardListener) IsRunning() bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return bl.running
}

func GetBoardListener() *BoardListener {
	return boardListener
}

func StartBoardListener() error {
	if boardListener == nil {
		return fmt.Errorf("board listener not initialized")
	}
	return boardListener.Start()
}

func StopBoardListener() error {
	if boardListener == nil {
		return fmt.Errorf("board listener not initialized")
	}
	return boardListener.Stop()
}

type BoardTCPHandler struct {
	conn       net.Conn
	boardID    string
	parser     *json.Decoder
	buffer     []byte
	classifier *BoardClassifier.BoardMessageClassifier
}

func NewBoardTCPHandler(conn net.Conn, classifier *BoardClassifier.BoardMessageClassifier) *BoardTCPHandler {
	return &BoardTCPHandler{
		conn:       conn,
		parser:     json.NewDecoder(conn),
		buffer:     make([]byte, 4096),
		classifier: classifier,
	}
}

func (h *BoardTCPHandler) Start() {
	defer h.conn.Close()

	log.Printf("[BoardTCPHandler] New connection from %s", h.conn.RemoteAddr().String())

	for {
		var msg Board.BoardMessage
		if err := h.parser.Decode(&msg); err != nil {
			log.Printf("[BoardTCPHandler] Decode error: %v", err)
			return
		}

		log.Printf("[BoardTCPHandler] Received: From=%s, Command=%s", msg.FromID, msg.Message.Command)

		h.classifier.ClassifyAndProcess(&msg)
	}
}

type BoardUDPHandler struct {
	addr       *net.UDPAddr
	boardID    string
	classifier *BoardClassifier.BoardMessageClassifier
}

func (h *BoardUDPHandler) Start(listener *net.UDPConn) {
	defer listener.Close()

	buffer := make([]byte, 4096)
	for {
		n, addr, err := listener.ReadFromUDP(buffer)
		if err != nil {
			log.Printf("[BoardUDPHandler] Read error: %v", err)
			return
		}

		var msg Board.BoardMessage
		if err := json.Unmarshal(buffer[:n], &msg); err != nil {
			log.Printf("[BoardUDPHandler] Unmarshal error: %v", err)
			continue
		}

		log.Printf("[BoardUDPHandler] Received from %s: From=%s, Command=%s",
			addr.String(), msg.FromID, msg.Message.Command)

		h.classifier.ClassifyAndProcess(&msg)
	}
}
