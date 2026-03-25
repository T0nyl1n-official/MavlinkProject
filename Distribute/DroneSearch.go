package Distribute

import (
	"fmt"
	"log"
	"sync"
	"time"

	boardHandler "MavlinkProject_Board/Handler/Boards"
	Board "MavlinkProject_Board/Shared/Boards"
	MavlinkBoard "MavlinkProject_Board/MavlinkCommand"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/message"
)

type MavlinkCommander = MavlinkBoard.MavlinkCommander

// 无人机状态阈值
const (
	MinBatteryLevel    float64       = 20.0
	MaxDroneDistance   float64       = 1000.0
	StatusCheckTimeout time.Duration = 5 * time.Second
)

type DroneStatus struct {
	BoardID      string
	SystemID     uint8
	ComponentID  uint8
	IsIdle       bool
	BatteryLevel float64
	Latitude     float64
	Longitude    float64
	Altitude     float64
	LastUpdate   time.Time
	Channel      *gomavlib.Channel
	Commander    *MavlinkCommander
}

type DroneSearch struct {
	boardManager *boardHandler.BoardConnectionManager
	drones       map[string]*DroneStatus
	activeTasks  map[string]string // boardID -> taskChainID
	mu           sync.RWMutex
	messageChan  chan *Board.BoardMessage
	stopChan     chan bool
	running      bool
}

var (
	droneSearch     *DroneSearch
	droneSearchOnce sync.Once
)

func GetDroneSearch() *DroneSearch {
	droneSearchOnce.Do(func() {
		droneSearch = &DroneSearch{
			boardManager: boardHandler.GetBoardManager(),
			drones:       make(map[string]*DroneStatus),
			activeTasks:  make(map[string]string),
			messageChan:  make(chan *Board.BoardMessage, 1000),
			stopChan:     make(chan bool),
		}
	})
	return droneSearch
}

func (ds *DroneSearch) Start() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if ds.running {
		return fmt.Errorf("DroneSearch already running")
	}

	ds.running = true
	ds.messageChan = ds.boardManager.GetMessageChan()

	go ds.statusUpdateLoop()

	log.Printf("[DroneSearch] Started")
	return nil
}

func (ds *DroneSearch) Stop() error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	if !ds.running {
		return nil
	}

	close(ds.stopChan)
	ds.running = false

	log.Printf("[DroneSearch] Stopped")
	return nil
}

func (ds *DroneSearch) statusUpdateLoop() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ds.stopChan:
			return
		case msg := <-ds.messageChan:
			ds.handleBoardMessage(msg)
		case <-ticker.C:
			ds.checkDroneStatus()
		}
	}
}

func (ds *DroneSearch) handleBoardMessage(msg *Board.BoardMessage) {
	if msg == nil || msg.Message.Data == nil {
		return
	}

	boardID := msg.FromID
	ds.mu.Lock()
	defer ds.mu.Unlock()

	drone, exists := ds.drones[boardID]
	if !exists {
		drone = &DroneStatus{
			BoardID:    boardID,
			LastUpdate: time.Now(),
		}
		ds.drones[boardID] = drone
	}

	data := msg.Message.Data

	if status, ok := data["status"].(string); ok {
		drone.IsIdle = (status == "idle" || status == "IDLE")
	}

	if battery, ok := data["battery"].(float64); ok {
		drone.BatteryLevel = battery
	}

	if lat, ok := data["latitude"].(float64); ok {
		drone.Latitude = lat
	}

	if lon, ok := data["longitude"].(float64); ok {
		drone.Longitude = lon
	}

	if alt, ok := data["altitude"].(float64); ok {
		drone.Altitude = alt
	}

	if sysID, ok := data["system_id"].(float64); ok {
		drone.SystemID = uint8(sysID)
	}

	if compID, ok := data["component_id"].(float64); ok {
		drone.ComponentID = uint8(compID)
	}

	drone.LastUpdate = time.Now()

	log.Printf("[DroneSearch] Updated drone %s: Idle=%v, Battery=%.1f%%, Lat=%.6f, Lon=%.6f",
		boardID, drone.IsIdle, drone.BatteryLevel, drone.Latitude, drone.Longitude)
}

func (ds *DroneSearch) checkDroneStatus() {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	now := time.Now()
	for boardID, drone := range ds.drones {
		if now.Sub(drone.LastUpdate) > StatusCheckTimeout {
			log.Printf("[DroneSearch] Drone %s timeout, marking as unavailable", boardID)
			drone.IsIdle = false
		}
	}
}

func (ds *DroneSearch) FindBestDrone() (*DroneStatus, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var bestDrone *DroneStatus
	var bestScore float64 = -1

	for _, drone := range ds.drones {
		if !drone.IsIdle {
			continue
		}

		if drone.BatteryLevel < MinBatteryLevel {
			continue
		}

		score := drone.BatteryLevel * 10

		if score > bestScore {
			bestScore = score
			bestDrone = drone
		}
	}

	if bestDrone == nil {
		return nil, fmt.Errorf("no available drone found")
	}

	log.Printf("[DroneSearch] Selected best drone: %s (Battery: %.1f%%, Score: %.1f)",
		bestDrone.BoardID, bestDrone.BatteryLevel, bestScore)

	return bestDrone, nil
}

func (ds *DroneSearch) GetDroneChannel(boardID string) (*gomavlib.Channel, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	drone, exists := ds.drones[boardID]
	if !exists {
		return nil, fmt.Errorf("drone %s not found", boardID)
	}

	if drone.Channel == nil {
		return nil, fmt.Errorf("drone %s channel not available", boardID)
	}

	return drone.Channel, nil
}

func (ds *DroneSearch) GetDroneCommander(boardID string) (*MavlinkCommander, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	drone, exists := ds.drones[boardID]
	if !exists {
		return nil, fmt.Errorf("drone %s not found", boardID)
	}

	if drone.Commander == nil {
		return nil, fmt.Errorf("drone %s commander not available", boardID)
	}

	return drone.Commander, nil
}

func (ds *DroneSearch) RegisterDroneCommander(boardID string, commander *MavlinkCommander) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	drone, exists := ds.drones[boardID]
	if !exists {
		drone = &DroneStatus{
			BoardID: boardID,
		}
		ds.drones[boardID] = drone
	}

	drone.Commander = commander
	log.Printf("[DroneSearch] Registered commander for drone %s", boardID)
}

func (ds *DroneSearch) RegisterDroneChannel(boardID string, channel *gomavlib.Channel) {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	drone, exists := ds.drones[boardID]
	if !exists {
		drone = &DroneStatus{
			BoardID: boardID,
		}
		ds.drones[boardID] = drone
	}

	drone.Channel = channel
	log.Printf("[DroneSearch] Registered channel for drone %s", boardID)
}

func (ds *DroneSearch) SendMessageToDrone(boardID string, msg message.Message) error {
	ds.mu.RLock()
	drone, exists := ds.drones[boardID]
	ds.mu.RUnlock()

	if !exists {
		return fmt.Errorf("drone %s not found", boardID)
	}

	if drone.Commander == nil {
		return fmt.Errorf("drone %s commander not available", boardID)
	}

	if drone.Channel != nil {
		return drone.Commander.WriteMessageTo(drone.Channel, msg)
	}

	return drone.Commander.WriteMessage(msg)
}

func (ds *DroneSearch) GetAllDrones() []*DroneStatus {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	drones := make([]*DroneStatus, 0, len(ds.drones))
	for _, drone := range ds.drones {
		drones = append(drones, drone)
	}

	return drones
}

func (ds *DroneSearch) GetAvailableDrones() []*DroneStatus {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	var drones []*DroneStatus
	for _, drone := range ds.drones {
		if drone.IsIdle && drone.BatteryLevel >= MinBatteryLevel {
			drones = append(drones, drone)
		}
	}

	return drones
}

func (ds *DroneSearch) GetDroneStatus(boardID string) (*DroneStatus, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	drone, exists := ds.drones[boardID]
	if !exists {
		return nil, fmt.Errorf("drone %s not found", boardID)
	}

	return drone, nil
}

func (ds *DroneSearch) SetDroneIdle(boardID string, isIdle bool) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	drone, exists := ds.drones[boardID]
	if !exists {
		return fmt.Errorf("drone %s not found", boardID)
	}

	drone.IsIdle = isIdle
	log.Printf("[DroneSearch] Set drone %s idle status to %v", boardID, isIdle)

	return nil
}

// 任务链调度相关方法
func (ds *DroneSearch) AssignTaskToDrone(boardID string, taskChainID string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	drone, exists := ds.drones[boardID]
	if !exists {
		return fmt.Errorf("drone %s not found", boardID)
	}

	// 检查无人机是否可用
	if !drone.IsIdle {
		return fmt.Errorf("drone %s is not idle", boardID)
	}

	if drone.BatteryLevel < MinBatteryLevel {
		return fmt.Errorf("drone %s battery level too low: %.1f%%", boardID, drone.BatteryLevel)
	}

	// 分配任务
	ds.activeTasks[boardID] = taskChainID
	drone.IsIdle = false

	log.Printf("[DroneSearch] Assigned task chain %s to drone %s", taskChainID, boardID)
	return nil
}

func (ds *DroneSearch) ReleaseDroneFromTask(boardID string) error {
	ds.mu.Lock()
	defer ds.mu.Unlock()

	drone, exists := ds.drones[boardID]
	if !exists {
		return fmt.Errorf("drone %s not found", boardID)
	}

	// 释放任务
	delete(ds.activeTasks, boardID)
	drone.IsIdle = true

	log.Printf("[DroneSearch] Released drone %s from task", boardID)
	return nil
}

func (ds *DroneSearch) GetDroneTaskChain(boardID string) (string, error) {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	taskChainID, exists := ds.activeTasks[boardID]
	if !exists {
		return "", fmt.Errorf("drone %s has no active task chain", boardID)
	}

	return taskChainID, nil
}

func (ds *DroneSearch) IsDroneAvailable(boardID string) bool {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	drone, exists := ds.drones[boardID]
	if !exists {
		return false
	}

	// 检查无人机是否空闲且电量充足
	return drone.IsIdle && drone.BatteryLevel >= MinBatteryLevel
}

func (ds *DroneSearch) GetAvailableDroneCount() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	count := 0
	for _, drone := range ds.drones {
		if drone.IsIdle && drone.BatteryLevel >= MinBatteryLevel {
			count++
		}
	}

	return count
}

func (ds *DroneSearch) GetActiveTaskCount() int {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	return len(ds.activeTasks)
}

func (ds *DroneSearch) GetMessageChan() chan *Board.BoardMessage {
	return ds.messageChan
}

// 检查无人机状态并返回详细报告
func (ds *DroneSearch) GetDroneStatusReport() map[string]interface{} {
	ds.mu.RLock()
	defer ds.mu.RUnlock()

	report := make(map[string]interface{})
	report["total_drones"] = len(ds.drones)
	report["available_drones"] = ds.GetAvailableDroneCount()
	report["active_tasks"] = len(ds.activeTasks)

	droneDetails := make(map[string]interface{})
	for boardID, drone := range ds.drones {
		droneDetails[boardID] = map[string]interface{}{
			"idle":          drone.IsIdle,
			"battery":       drone.BatteryLevel,
			"latitude":      drone.Latitude,
			"longitude":     drone.Longitude,
			"altitude":      drone.Altitude,
			"last_update":   drone.LastUpdate.Format(time.RFC3339),
			"system_id":     drone.SystemID,
			"component_id":  drone.ComponentID,
			"has_commander": drone.Commander != nil,
			"has_channel":   drone.Channel != nil,
		}
	}

	report["drone_details"] = droneDetails

	activeTaskDetails := make(map[string]string)
	for boardID, taskChainID := range ds.activeTasks {
		activeTaskDetails[boardID] = taskChainID
	}

	report["active_task_details"] = activeTaskDetails

	return report
}
