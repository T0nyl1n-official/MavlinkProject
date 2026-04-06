package listening

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/gin-gonic/gin"

	BoardHandler "MavlinkProject/Server/Backend/Handler/Boards"
	BoardClassifier "MavlinkProject/Server/Backend/Utils/BoardClassifier"
	Conf "MavlinkProject/Server/backend/Config"
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

func GetDefaultBoardListenerConfig() BoardListenerConfig {
	setting := Conf.GetSetting()
	boardCfg := setting.Board

	return BoardListenerConfig{
		EnableTCP:     boardCfg.TCP.Enabled,
		EnableUDP:     boardCfg.UDP.Enabled,
		TCPAddress:    boardCfg.TCP.Address,
		TCPPort:       boardCfg.TCP.Port,
		UDPAddress:    boardCfg.UDP.Address,
		UDPPort:       boardCfg.UDP.Port,
		MaxBufferSize: boardCfg.TCP.MaxBufferSize,
	}
}

var (
	boardListener     *BoardListener
	boardListenerOnce sync.Once
	boardManager      *BoardHandler.BoardConnectionManager
	classifier        *BoardClassifier.BoardMessageClassifier
)

func BoardListenerMiddleware() gin.HandlerFunc {
	return BoardListenerWithConfig(GetDefaultBoardListenerConfig())
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

	bl.running = true
	return nil
}

func (bl *BoardListener) Stop() error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if !bl.running {
		return fmt.Errorf("board listener not running")
	}

	bl.running = false
	log.Printf("[BoardListener] Stopped")
	return nil
}

func (bl *BoardListener) IsRunning() bool {
	bl.mu.RLock()
	defer bl.mu.RUnlock()
	return bl.running
}

func (bl *BoardListener) GetConfig() BoardListenerConfig {
	return bl.config
}

func (bl *BoardListener) SetConfig(config BoardListenerConfig) error {
	bl.mu.Lock()
	defer bl.mu.Unlock()

	if bl.running {
		return fmt.Errorf("cannot update config while running")
	}

	bl.config = config
	log.Printf("[BoardListener] Config updated - TCP: %s:%s, UDP: %s:%s",
		config.TCPAddress, config.TCPPort, config.UDPAddress, config.UDPPort)
	return nil
}

func (bl *BoardListener) GetStatus() string {
	bl.mu.RLock()
	defer bl.mu.RUnlock()

	config := bl.config
	running := bl.running

	return fmt.Sprintf(`{"running":%v,"tcp":{"enabled":%v,"address":"%s","port":"%s"},"udp":{"enabled":%v,"address":"%s","port":"%s"}}`,
		running,
		config.EnableTCP, config.TCPAddress, config.TCPPort,
		config.EnableUDP, config.UDPAddress, config.UDPPort)
}

func GetBoardListenerStatus() string {
	if boardListener == nil {
		return `{"error":"board listener not initialized"}`
	}
	return boardListener.GetStatus()
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

func UpdateBoardListenerConfig(config BoardListenerConfig) error {
	if boardListener == nil {
		return fmt.Errorf("board listener not initialized")
	}
	return boardListener.SetConfig(config)
}

func BoardListenerJSON(c *gin.Context) {
	status := GetBoardListenerStatus()
	var jsonData interface{}
	if err := json.Unmarshal([]byte(status), &jsonData); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, jsonData)
}
