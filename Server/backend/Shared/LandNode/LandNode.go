package LandNode

import (
	"fmt"
	"sync"
	"time"

	Drones "MavlinkProject/Server/backend/Shared/Drones"
)

type LandNodeConfig struct {
	SystemID        int           `json:"system_id"`
	ComponentID     int           `json:"component_id"`
	ProtocolVersion string        `json:"protocol_version"`
	HeartbeatRate   time.Duration `json:"heartbeat_rate"`
	LogLevel        string        `json:"log_level"`
	MaxDrones       int           `json:"max_drones"`
}

type MessageHandler func(droneID string, message interface{})

type LandNode struct {
	ID     string         `json:"id"`
	Name   string         `json:"name"`
	Config LandNodeConfig `json:"config"`

	drones map[string]*Drones.Drone `json:"-"`
	mu     sync.RWMutex

	messageHandlers map[string]MessageHandler

	started  bool
	stopChan chan bool
}

func NewLandNode(id, name string, config LandNodeConfig) *LandNode {
	if config.HeartbeatRate == 0 {
		config.HeartbeatRate = 1 * time.Second
	}
	if config.MaxDrones == 0 {
		config.MaxDrones = 10
	}

	return &LandNode{
		ID:              id,
		Name:            name,
		Config:          config,
		drones:          make(map[string]*Drones.Drone),
		messageHandlers: make(map[string]MessageHandler),
		stopChan:        make(chan bool),
	}
}

// 无人机 容器操作-内部方法
func (ln *LandNode) GetID() string {
	ln.mu.RLock()
	defer ln.mu.RUnlock()
	return ln.ID
}

func (ln *LandNode) GetName() string {
	ln.mu.RLock()
	defer ln.mu.RUnlock()
	return ln.Name
}

func (ln *LandNode) SetName(name string) {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	ln.Name = name
}

func (ln *LandNode) GetConfig() LandNodeConfig {
	ln.mu.RLock()
	defer ln.mu.RUnlock()
	return ln.Config
}

func (ln *LandNode) SetConfig(config LandNodeConfig) {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	ln.Config = config
}

// 无人机 外部方法
func (ln *LandNode) AddDrone(drone *Drones.Drone) error {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if len(ln.drones) >= ln.Config.MaxDrones {
		return fmt.Errorf("已达到最大无人机数量限制: %d", ln.Config.MaxDrones)
	}

	if _, exists := ln.drones[drone.GetID()]; exists {
		return fmt.Errorf("无人机已存在: %s", drone.GetID())
	}

	ln.drones[drone.GetID()] = drone
	return nil
}

func (ln *LandNode) RemoveDrone(droneID string) error {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if _, exists := ln.drones[droneID]; !exists {
		return fmt.Errorf("无人机不存在: %s", droneID)
	}

	delete(ln.drones, droneID)
	return nil
}

func (ln *LandNode) GetDrone(droneID string) (*Drones.Drone, error) {
	ln.mu.RLock()
	defer ln.mu.RUnlock()

	drone, exists := ln.drones[droneID]
	if !exists {
		return nil, fmt.Errorf("无人机不存在: %s", droneID)
	}

	return drone, nil
}

func (ln *LandNode) GetAllDrones() []*Drones.Drone {
	ln.mu.RLock()
	defer ln.mu.RUnlock()

	drones := make([]*Drones.Drone, 0, len(ln.drones))
	for _, drone := range ln.drones {
		drones = append(drones, drone)
	}

	return drones
}

func (ln *LandNode) GetDroneCount() int {
	ln.mu.RLock()
	defer ln.mu.RUnlock()
	return len(ln.drones)
}

func (ln *LandNode) RegisterMessageHandler(messageType string, handler MessageHandler) {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	ln.messageHandlers[messageType] = handler
}

func (ln *LandNode) UnregisterMessageHandler(messageType string) {
	ln.mu.Lock()
	defer ln.mu.Unlock()
	delete(ln.messageHandlers, messageType)
}

func (ln *LandNode) HandleMessage(droneID string, message interface{}) {
	ln.mu.RLock()
	defer ln.mu.RUnlock()

	if handler, exists := ln.messageHandlers["*"]; exists {
		handler(droneID, message)
	}

	messageType := fmt.Sprintf("%T", message)
	if handler, exists := ln.messageHandlers[messageType]; exists {
		handler(droneID, message)
	}
}

func (ln *LandNode) Start() error {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if ln.started {
		return fmt.Errorf("地面站已经在运行")
	}

	ln.started = true
	go ln.monitorDrones()

	return nil
}

func (ln *LandNode) Stop() error {
	ln.mu.Lock()
	defer ln.mu.Unlock()

	if !ln.started {
		return fmt.Errorf("地面站未在运行")
	}

	ln.started = false
	close(ln.stopChan)
	ln.stopChan = make(chan bool)

	return nil
}

func (ln *LandNode) IsStarted() bool {
	ln.mu.RLock()
	defer ln.mu.RUnlock()
	return ln.started
}

func (ln *LandNode) monitorDrones() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ln.stopChan:
			return
		case <-ticker.C:
			ln.checkDroneConnections()
		}
	}
}

func (ln *LandNode) checkDroneConnections() {
	ln.mu.RLock()
	timeout := 10 * time.Second
	ln.mu.RUnlock()

	for _, drone := range ln.GetAllDrones() {
		if drone.IsTimedOut(timeout) {
			drone.SetConnected(false)
			drone.SetStatus(Drones.StatusDisconnected)
		}
	}
}

func (ln *LandNode) BroadcastMessage(message interface{}) error {
	ln.mu.RLock()
	defer ln.mu.RUnlock()

	if !ln.started {
		return fmt.Errorf("地面站未在运行")
	}

	for droneID, _ := range ln.drones {
		ln.HandleMessage(droneID, message)
	}

	return nil
}
