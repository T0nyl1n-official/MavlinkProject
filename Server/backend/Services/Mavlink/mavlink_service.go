package mavlink

import (
	"fmt"
	"sync"
	"time"

	"MavlinkProject/Server/backend/Shared/Drones"
)

type MavlinkConfig struct {
	ProtocolVersion  string        `json:"protocol_version"`
	SystemID         int           `json:"system_id"`
	ComponentID      int           `json:"component_id"`
	HeartbeatRate    time.Duration `json:"heartbeat_rate"`
	ConnectionTimeout time.Duration `json:"connection_timeout"`
	MaxRetries      int           `json:"max_retries"`
	RetryDelay      time.Duration `json:"retry_delay"`
}

type MessageType string

const (
	MsgTypeHeartbeat       MessageType = "heartbeat"
	MsgTypeSysStatus       MessageType = "sys_status"
	MsgTypeGPSRawInt       MessageType = "gps_raw_int"
	MsgTypeAttitude        MessageType = "attitude"
	MsgTypeVfrHud          MessageType = "vfr_hud"
	MsgTypeBatteryStatus   MessageType = "battery_status"
	MsgTypeMissionItem     MessageType = "mission_item"
	MsgTypeMissionAck      MessageType = "mission_ack"
	MsgTypeCommandAck      MessageType = "command_ack"
	MsgTypeParamRequestRead MessageType = "param_request_read"
	MsgTypeParamValue      MessageType = "param_value"
	MsgTypeSetMode        MessageType = "set_mode"
)

type MavlinkMessage struct {
	Type      MessageType
	SystemID  int
	ComponentID int
	Payload   interface{}
	Timestamp time.Time
}

type Command struct {
	ID          uint16
	Command     uint16
	Params      []float32
	TargetSystem uint8
	TargetComponent uint8
	Response    chan CommandResponse
}

type CommandResponse struct {
	Command   uint16
	Result    uint8
	Progress  uint8
	ResultParam2 int32
	Message   string
	Error     error
}

type MavlinkService struct {
	config     MavlinkConfig
	drone      *Drones.Drone
	
	messageQueue chan *MavlinkMessage
	commandChan chan *Command
	
	started    bool
	mu         sync.RWMutex
	stopChan   chan bool
	
	messageHandlers map[MessageType]MessageHandler
}

type MessageHandler func(msg *MavlinkMessage) error

func NewMavlinkService(config MavlinkConfig, drone *Drones.Drone) *MavlinkService {
	if config.HeartbeatRate == 0 {
		config.HeartbeatRate = 1 * time.Second
	}
	if config.ConnectionTimeout == 0 {
		config.ConnectionTimeout = 10 * time.Second
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 1 * time.Second
	}
	
	return &MavlinkService{
		config: config,
		drone: drone,
		messageQueue: make(chan *MavlinkMessage, 1000),
		commandChan: make(chan *Command, 100),
		messageHandlers: make(map[MessageType]MessageHandler),
		stopChan: make(chan bool),
	}
}

func (ms *MavlinkService) GetConfig() MavlinkConfig {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.config
}

func (ms *MavlinkService) SetConfig(config MavlinkConfig) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.config = config
}

func (ms *MavlinkService) GetDrone() *Drones.Drone {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.drone
}

func (ms *MavlinkService) SetDrone(drone *Drones.Drone) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.drone = drone
}

func (ms *MavlinkService) RegisterHandler(msgType MessageType, handler MessageHandler) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.messageHandlers[msgType] = handler
}

func (ms *MavlinkService) UnregisterHandler(msgType MessageType) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	delete(ms.messageHandlers, msgType)
}

func (ms *MavlinkService) Start() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if ms.started {
		return fmt.Errorf("MAVLink服务已经在运行")
	}
	
	ms.started = true
	go ms.processMessageQueue()
	go ms.processCommands()
	
	return nil
}

func (ms *MavlinkService) Stop() error {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	
	if !ms.started {
		return fmt.Errorf("MAVLink服务未在运行")
	}
	
	ms.started = false
	close(ms.stopChan)
	ms.stopChan = make(chan bool)
	
	return nil
}

func (ms *MavlinkService) IsStarted() bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return ms.started
}

func (ms *MavlinkService) SendMessage(msg *MavlinkMessage) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if !ms.started {
		return fmt.Errorf("MAVLink服务未在运行")
	}
	
	select {
	case ms.messageQueue <- msg:
		return nil
	default:
		return fmt.Errorf("消息队列已满")
	}
}

func (ms *MavlinkService) SendCommand(cmd *Command) error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	
	if !ms.started {
		return fmt.Errorf("MAVLink服务未在运行")
	}
	
	select {
	case ms.commandChan <- cmd:
		return nil
	default:
		return fmt.Errorf("命令队列已满")
	}
}

func (ms *MavlinkService) processMessageQueue() {
	for {
		select {
		case <-ms.stopChan:
			return
		case msg := <-ms.messageQueue:
			ms.handleMessage(msg)
		}
	}
}

func (ms *MavlinkService) processCommands() {
	for {
		select {
		case <-ms.stopChan:
			return
		case cmd := <-ms.commandChan:
			ms.executeCommand(cmd)
		}
	}
}

func (ms *MavlinkService) handleMessage(msg *MavlinkMessage) {
	ms.mu.RLock()
	handler, exists := ms.messageHandlers[msg.Type]
	ms.mu.RUnlock()
	
	if exists && handler != nil {
		if err := handler(msg); err != nil {
			ms.handleError(err, msg)
		}
	}
	
	switch msg.Type {
	case MsgTypeHeartbeat:
		ms.handleHeartbeat(msg)
	case MsgTypeAttitude:
		ms.handleAttitude(msg)
	case MsgTypeBatteryStatus:
		ms.handleBattery(msg)
	}
}

func (ms *MavlinkService) handleHeartbeat(msg *MavlinkMessage) {
	if ms.drone != nil {
		ms.drone.UpdateHeartbeat()
	}
}

func (ms *MavlinkService) handleAttitude(msg *MavlinkMessage) {
	if ms.drone != nil && msg.Payload != nil {
		if data, ok := msg.Payload.(map[string]float64); ok {
			att := Drones.Attitude{
				Roll:  data["roll"],
				Pitch: data["pitch"],
				Yaw:   data["yaw"],
			}
			ms.drone.SetAttitude(att)
		}
	}
}

func (ms *MavlinkService) handleBattery(msg *MavlinkMessage) {
	if ms.drone != nil && msg.Payload != nil {
		if data, ok := msg.Payload.(map[string]interface{}); ok {
			bat := Drones.BatteryStatus{}
			if v, ok := data["voltage"].(float64); ok {
				bat.Voltage = v
			}
			if v, ok := data["remaining"].(int); ok {
				bat.Remaining = v
			}
			ms.drone.SetBattery(bat)
		}
	}
}

func (ms *MavlinkService) executeCommand(cmd *Command) {
	response := CommandResponse{
		Command: cmd.Command,
		Result:  0,
	}
	
	cmd.Response <- response
}

func (ms *MavlinkService) handleError(err error, msg *MavlinkMessage) {
	fmt.Printf("MAVLink错误: %v, 消息类型: %v\n", err, msg.Type)
}

func (ms *MavlinkService) GetMessageQueueLength() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.messageQueue)
}

func (ms *MavlinkService) GetCommandQueueLength() int {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	return len(ms.commandChan)
}
