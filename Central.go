package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"
	"gopkg.in/yaml.v3"

	MavlinkBoard "MavlinkProject_Board/MavlinkCommand"
)

type Config struct {
	Server struct {
		Port     string `yaml:"port"`
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
		Domain   string `yaml:"domain"`
		Email    string `yaml:"email"`
	} `yaml:"server"`
	Mavlink struct {
		SerialPort   string `yaml:"serial_port"`
		SerialBaud   int    `yaml:"serial_baud"`
		TargetSystem uint8  `yaml:"target_system"`
	} `yaml:"mavlink"`
	Drone struct {
		Search struct {
			Interval int `yaml:"interval"`
			Timeout  int `yaml:"timeout"`
		} `yaml:"search"`
	} `yaml:"drone"`
}

type CentralServer struct {
	config      *Config
	commander   *MavlinkBoard.MavlinkCommander
	router      *gin.Engine
	server      *http.Server
	taskChains  map[string]*ProgressChain
	activeTasks map[string]*Task
	mu          sync.RWMutex
	stopChan    chan bool
	running     bool
}

type ProgressChain struct {
	ChainID     string    `json:"chain_id"`
	Tasks       []Task    `json:"tasks"`
	CurrentTask int       `json:"current_task"`
	Status      string    `json:"status"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
}

type Task struct {
	TaskID     string                 `json:"task_id"`
	Command    string                 `json:"command"`
	Data       map[string]interface{} `json:"data"`
	Status     string                 `json:"status"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	Timeout    time.Duration          `json:"timeout"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
}

type MessageRequest struct {
	MessageID   string    `json:"message_id"`
	MessageTime time.Time `json:"message_time"`
	Message     Message   `json:"message"`
}

type Message struct {
	MessageType string                 `json:"message_type"`
	Attribute   string                 `json:"attribute"`
	Command     string                 `json:"command"`
	Data        map[string]interface{} `json:"data"`
}

type MessageResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	ChainID string `json:"chain_id,omitempty"`
}

const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"

	MaxRetries  = 3
	TaskTimeout = 30 * time.Second
)

var appConfig *Config

func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	if cfg.Server.Port == "" {
		cfg.Server.Port = "8084"
	}
	if cfg.Mavlink.SerialBaud == 0 {
		cfg.Mavlink.SerialBaud = 115200
	}
	if cfg.Mavlink.TargetSystem == 0 {
		cfg.Mavlink.TargetSystem = 1
	}

	return &cfg, nil
}

func NewCentralServer(cfg *Config) *CentralServer {
	return &CentralServer{
		config:      cfg,
		commander:   MavlinkBoard.NewMavlinkCommander(),
		taskChains:  make(map[string]*ProgressChain),
		activeTasks: make(map[string]*Task),
		stopChan:    make(chan bool),
	}
}

func (cs *CentralServer) initializeMavlink() error {
	mavConfig := MavlinkBoard.MavlinkConfig{
		ConnectionType: MavlinkBoard.ConnectionSerial,
		SerialPort:     cs.config.Mavlink.SerialPort,
		SerialBaud:     cs.config.Mavlink.SerialBaud,
		TargetSystem:   cs.config.Mavlink.TargetSystem,
		HeartbeatRate:  time.Second,
	}

	cs.commander.Configure(mavConfig)

	if err := cs.commander.Start(); err != nil {
		return fmt.Errorf("failed to start MAVLink: %v", err)
	}

	log.Printf("[Central] MAVLink initialized: %s @ %d baud", cs.config.Mavlink.SerialPort, cs.config.Mavlink.SerialBaud)
	return nil
}

func (cs *CentralServer) setupRouter() {
	gin.SetMode(gin.ReleaseMode)
	cs.router = gin.New()
	cs.router.Use(gin.Recovery())
	cs.router.Use(gin.Logger())

	cs.router.GET("/health", cs.healthCheck)

	central := cs.router.Group("/central")
	{
		central.POST("/message", cs.handleMessage)
		central.GET("/status", cs.getStatus)
		central.GET("/chain/:chain_id", cs.getChainStatus)
	}
}

func (cs *CentralServer) Start() error {
	cs.mu.Lock()
	if cs.running {
		cs.mu.Unlock()
		return fmt.Errorf("CentralServer already running")
	}

	if err := cs.initializeMavlink(); err != nil {
		cs.mu.Unlock()
		return fmt.Errorf("failed to initialize MAVLink: %v", err)
	}

	cs.setupRouter()

	addr := fmt.Sprintf("0.0.0.0:%s", cs.config.Server.Port)
	cs.server = &http.Server{
		Addr:    addr,
		Handler: cs.router,
	}

	cs.running = true
	cs.mu.Unlock()

	go cs.startHTTPServer()

	log.Printf("[Central] HTTPS Server started on %s", addr)
	return nil
}

func (cs *CentralServer) startHTTPServer() {
	var err error

	if cs.config.Server.Domain != "" {
		log.Printf("[Central] Using Let's Encrypt for domain: %s", cs.config.Server.Domain)

		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(cs.config.Server.Domain),
			Cache:      autocert.DirCache("letsencrypt-cache"),
			Email:      cs.config.Server.Email,
		}

		tlsConfig := &tls.Config{
			GetCertificate: certManager.GetCertificate,
		}

		tlsListener, err := tls.Listen("tcp", cs.server.Addr, tlsConfig)
		if err != nil {
			log.Printf("[Central] TLS listener error: %v", err)
			return
		}

		go func() {
			err = cs.server.Serve(tlsListener)
			if err != nil && err != http.ErrServerClosed {
				log.Printf("[Central] HTTPS Server error: %v", err)
			}
		}()

		go func() {
			httpServer := &http.Server{
				Addr:    ":80",
				Handler: certManager.HTTPHandler(nil),
			}
			log.Printf("[Central] ACME HTTP server started on :80 for certificate verification")
			err = httpServer.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Printf("[Central] ACME HTTP server error: %v", err)
			}
		}()

		log.Printf("[Central] HTTPS Server started with Let's Encrypt on %s", cs.server.Addr)
	} else if cs.config.Server.CertFile != "" && cs.config.Server.KeyFile != "" {
		err = cs.server.ListenAndServeTLS(cs.config.Server.CertFile, cs.config.Server.KeyFile)
		if err != nil && err != http.ErrServerClosed {
			log.Printf("[Central] HTTPS Server error: %v", err)
		}
	} else {
		log.Printf("[Central] Warning: Running without TLS (no cert/key configured)")
		err = cs.server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Printf("[Central] HTTP Server error: %v", err)
		}
	}
}

func (cs *CentralServer) Stop() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if !cs.running {
		return nil
	}

	cs.running = false
	close(cs.stopChan)

	if cs.commander != nil {
		cs.commander.Stop()
	}

	if cs.server != nil {
		cs.server.Close()
	}

	log.Printf("[Central] Stopped")
	return nil
}

func (cs *CentralServer) healthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

func (cs *CentralServer) handleMessage(c *gin.Context) {
	var req MessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MessageResponse{
			Status:  "error",
			Message: fmt.Sprintf("Invalid request: %v", err),
		})
		return
	}

	log.Printf("[Central] Received message: type=%s, command=%s", req.Message.MessageType, req.Message.Command)

	chainData, exists := req.Message.Data["progress_chain"]
	if !exists {
		c.JSON(http.StatusBadRequest, MessageResponse{
			Status:  "error",
			Message: "No progress_chain found in message data",
		})
		return
	}

	chainJSON, err := json.Marshal(chainData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to marshal chain data: %v", err),
		})
		return
	}

	var progressChain ProgressChain
	if err := json.Unmarshal(chainJSON, &progressChain); err != nil {
		c.JSON(http.StatusInternalServerError, MessageResponse{
			Status:  "error",
			Message: fmt.Sprintf("Failed to unmarshal progress chain: %v", err),
		})
		return
	}

	if progressChain.ChainID == "" {
		progressChain.ChainID = fmt.Sprintf("chain_%d", time.Now().UnixNano())
	}
	progressChain.Status = TaskStatusPending
	progressChain.StartTime = time.Now()

	for i := range progressChain.Tasks {
		progressChain.Tasks[i].TaskID = fmt.Sprintf("task_%s_%d", progressChain.ChainID, i)
		progressChain.Tasks[i].Status = TaskStatusPending
		progressChain.Tasks[i].MaxRetries = MaxRetries
		progressChain.Tasks[i].Timeout = TaskTimeout
	}

	cs.mu.Lock()
	cs.taskChains[progressChain.ChainID] = &progressChain
	cs.mu.Unlock()

	go cs.processChain(&progressChain)

	c.JSON(http.StatusOK, MessageResponse{
		Status:  "ok",
		Message: "Task chain received and queued",
		ChainID: progressChain.ChainID,
	})
}

func (cs *CentralServer) processChain(chain *ProgressChain) {
	log.Printf("[Central] Processing chain %s with %d tasks", chain.ChainID, len(chain.Tasks))

	cs.mu.Lock()
	chain.Status = TaskStatusRunning
	cs.mu.Unlock()

	for i := range chain.Tasks {
		task := &chain.Tasks[i]
		task.Status = TaskStatusRunning
		task.StartTime = time.Now()

		cs.mu.Lock()
		cs.activeTasks[task.TaskID] = task
		cs.mu.Unlock()

		log.Printf("[Central] Executing task %s: command=%s", task.TaskID, task.Command)

		err := cs.executeTask(task)

		cs.mu.Lock()
		delete(cs.activeTasks, task.TaskID)
		cs.mu.Unlock()

		if err != nil {
			log.Printf("[Central] Task %s failed: %v", task.TaskID, err)
			task.Status = TaskStatusFailed

			cs.mu.Lock()
			chain.Status = TaskStatusFailed
			cs.mu.Unlock()
			return
		}

		task.Status = TaskStatusCompleted
		task.EndTime = time.Now()
		log.Printf("[Central] Task %s completed", task.TaskID)
	}

	cs.mu.Lock()
	chain.Status = TaskStatusCompleted
	chain.EndTime = time.Now()
	cs.mu.Unlock()

	log.Printf("[Central] Chain %s completed", chain.ChainID)
}

func (cs *CentralServer) executeTask(task *Task) error {
	command := task.Command
	params := task.Data

	switch command {
	case "TakeOff", "takeoff":
		return cs.cmdTakeOff(params)
	case "Land", "land":
		return cs.cmdLand(params)
	case "GoTo", "goto":
		return cs.cmdGoto(params)
	case "SetSpeed", "set_speed":
		return cs.cmdSetSpeed(params)
	case "SetPosition", "set_position":
		return cs.cmdSetPosition(params)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func (cs *CentralServer) cmdTakeOff(params map[string]interface{}) error {
	altitude, ok := params["altitude"].(float64)
	if !ok {
		altitude = 10.0
	}

	targetSystem := cs.config.Mavlink.TargetSystem

	log.Printf("[Central] MAVLink: TakeOff to %.1f meters", altitude)
	return cs.commander.CommandLong(
		targetSystem, 1,
		MAV_CMD_NAV_TAKEOFF,
		0,
		0, 0, 0, 0, 0, 0,
		float32(altitude),
	)
}

func (cs *CentralServer) cmdLand(params map[string]interface{}) error {
	lat, _ := params["latitude"].(float64)
	lon, _ := params["longitude"].(float64)
	alt, _ := params["altitude"].(float64)

	targetSystem := cs.config.Mavlink.TargetSystem

	log.Printf("[Central] MAVLink: Land at (%.6f, %.6f, %.1f)", lat, lon, alt)
	return cs.commander.CommandLong(
		targetSystem, 1,
		MAV_CMD_NAV_LAND,
		0,
		0, 0, 0, 0,
		float32(lat), float32(lon),
		float32(alt),
	)
}

func (cs *CentralServer) cmdGoto(params map[string]interface{}) error {
	lat, ok := params["latitude"].(float64)
	if !ok {
		return fmt.Errorf("missing latitude")
	}
	lon, ok := params["longitude"].(float64)
	if !ok {
		return fmt.Errorf("missing longitude")
	}
	alt, _ := params["altitude"].(float64)

	targetSystem := cs.config.Mavlink.TargetSystem

	log.Printf("[Central] MAVLink: GoTo (%.6f, %.6f, %.1f)", lat, lon, alt)
	return cs.commander.CommandLong(
		targetSystem, 1,
		MAV_CMD_NAV_WAYPOINT,
		0,
		0, 0, 0, 0,
		float32(lat), float32(lon),
		float32(alt),
	)
}

func (cs *CentralServer) cmdSetSpeed(params map[string]interface{}) error {
	speed, ok := params["speed"].(float64)
	if !ok {
		return fmt.Errorf("missing speed")
	}
	speedType, _ := params["type"].(float64)

	targetSystem := cs.config.Mavlink.TargetSystem

	log.Printf("[Central] MAVLink: SetSpeed %.1f (type %.0f)", speed, speedType)
	return cs.commander.CommandLong(
		targetSystem, 1,
		MAV_CMD_DO_CHANGE_SPEED,
		0,
		float32(speedType), float32(speed), 0, 0, 0, 0, 0,
	)
}

func (cs *CentralServer) cmdSetPosition(params map[string]interface{}) error {
	lat, ok := params["latitude"].(float64)
	if !ok {
		return fmt.Errorf("missing latitude")
	}
	lon, ok := params["longitude"].(float64)
	if !ok {
		return fmt.Errorf("missing longitude")
	}
	alt, _ := params["altitude"].(float64)

	targetSystem := cs.config.Mavlink.TargetSystem

	log.Printf("[Central] MAVLink: SetPosition (%.6f, %.6f, %.1f)", lat, lon, alt)
	return cs.commander.CommandLong(
		targetSystem, 1,
		MAV_CMD_DO_SET_MODE,
		0,
		float32(MAV_FRAME_GLOBAL_RELATIVE_ALT), 0, 0, 0,
		float32(lat), float32(lon),
		float32(alt),
	)
}

func (cs *CentralServer) getStatus(c *gin.Context) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	c.JSON(http.StatusOK, gin.H{
		"status":        "running",
		"active_chains": len(cs.taskChains),
		"active_tasks":  len(cs.activeTasks),
		"timestamp":     time.Now().Unix(),
	})
}

func (cs *CentralServer) getChainStatus(c *gin.Context) {
	chainID := c.Param("chain_id")

	cs.mu.RLock()
	chain, exists := cs.taskChains[chainID]
	cs.mu.RUnlock()

	if !exists {
		c.JSON(http.StatusNotFound, MessageResponse{
			Status:  "error",
			Message: fmt.Sprintf("Chain %s not found", chainID),
		})
		return
	}

	c.JSON(http.StatusOK, chain)
}

const (
	MAV_CMD_NAV_TAKEOFF           = 22
	MAV_CMD_NAV_LAND              = 21
	MAV_CMD_NAV_WAYPOINT          = 16
	MAV_CMD_DO_CHANGE_SPEED       = 178
	MAV_CMD_DO_SET_MODE           = 176
	MAV_FRAME_GLOBAL_RELATIVE_ALT = 3
)

func main() {
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	appConfig = cfg

	central := NewCentralServer(cfg)

	if err := central.Start(); err != nil {
		log.Fatalf("Failed to start CentralServer: %v", err)
	}

	log.Printf("[Central] Central调度系统已启动")
	log.Printf("[Central] 等待接收任务链...")

	<-central.stopChan

	log.Printf("[Central] Central调度系统正在关闭...")
	central.Stop()
}
