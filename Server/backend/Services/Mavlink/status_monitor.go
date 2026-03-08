package MavlinkService

import (
	"fmt"
	"sync"
	"time"

	"MavlinkProject/Server/backend/Shared/Drones"
)

type StatusChangeEvent struct {
	DroneID   string             `json:"drone_id"`
	EventType string             `json:"event_type"`
	OldStatus Drones.DroneStatus `json:"old_status"`
	NewStatus Drones.DroneStatus `json:"new_status"`
	Timestamp time.Time          `json:"timestamp"`
	Details   string             `json:"details"`
}

type DroneStatusLog struct {
	ID        string               `json:"id"`
	DroneID   string               `json:"drone_id"`
	Status    Drones.DroneStatus   `json:"status"`
	Position  Drones.Position      `json:"position,omitempty"`
	Battery   Drones.BatteryStatus `json:"battery,omitempty"`
	Message   string               `json:"message"`
	Timestamp time.Time            `json:"timestamp"`
}

type StatusMonitor struct {
	mu              sync.RWMutex
	droneStatus     map[string]Drones.DroneStatus
	statusHistory   map[string][]*DroneStatusLog
	changeCallbacks map[string]StatusChangeCallback

	alertThresholds map[string]*AlertThreshold

	logChan   chan *DroneStatusLog
	eventChan chan *StatusChangeEvent

	maxHistorySize int
}

type StatusChangeCallback func(event *StatusChangeEvent)

type AlertThreshold struct {
	BatteryLow      int
	BatteryCritical int
	SignalWeak      int
	Timeout         time.Duration
}

func NewStatusMonitor(maxHistorySize int) *StatusMonitor {
	if maxHistorySize == 0 {
		maxHistorySize = 1000
	}

	return &StatusMonitor{
		droneStatus:     make(map[string]Drones.DroneStatus),
		statusHistory:   make(map[string][]*DroneStatusLog),
		changeCallbacks: make(map[string]StatusChangeCallback),
		alertThresholds: make(map[string]*AlertThreshold),
		logChan:         make(chan *DroneStatusLog, 100),
		eventChan:       make(chan *StatusChangeEvent, 50),
		maxHistorySize:  maxHistorySize,
	}
}

func (sm *StatusMonitor) SetAlertThreshold(droneID string, threshold *AlertThreshold) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.alertThresholds[droneID] = threshold
}

func (sm *StatusMonitor) GetAlertThreshold(droneID string) *AlertThreshold {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	return sm.alertThresholds[droneID]
}

func (sm *StatusMonitor) UpdateDroneStatus(droneID string, status Drones.DroneStatus, message string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	oldStatus, exists := sm.droneStatus[droneID]
	newStatus := status

	sm.droneStatus[droneID] = newStatus

	log := &DroneStatusLog{
		ID:        fmt.Sprintf("log_%d", time.Now().UnixNano()),
		DroneID:   droneID,
		Status:    newStatus,
		Message:   message,
		Timestamp: time.Now(),
	}

	sm.addLog(droneID, log)

	if exists && oldStatus != newStatus {
		event := &StatusChangeEvent{
			DroneID:   droneID,
			EventType: "status_changed",
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: time.Now(),
			Details:   message,
		}

		sm.eventChan <- event

		if callback, ok := sm.changeCallbacks[droneID]; ok {
			go callback(event)
		}

		if callback, ok := sm.changeCallbacks["*"]; ok {
			go callback(event)
		}
	}

	return nil
}

func (sm *StatusMonitor) UpdateDroneStatusWithData(droneID string, status Drones.DroneStatus, position *Drones.Position, battery *Drones.BatteryStatus, message string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	oldStatus, exists := sm.droneStatus[droneID]
	newStatus := status

	sm.droneStatus[droneID] = newStatus

	log := &DroneStatusLog{
		ID:        fmt.Sprintf("log_%d", time.Now().UnixNano()),
		DroneID:   droneID,
		Status:    newStatus,
		Position:  Drones.Position{},
		Battery:   Drones.BatteryStatus{},
		Message:   message,
		Timestamp: time.Now(),
	}

	if position != nil {
		log.Position = *position
	}
	if battery != nil {
		log.Battery = *battery
	}

	sm.addLog(droneID, log)

	if exists && oldStatus != newStatus {
		event := &StatusChangeEvent{
			DroneID:   droneID,
			EventType: "status_changed",
			OldStatus: oldStatus,
			NewStatus: newStatus,
			Timestamp: time.Now(),
			Details:   message,
		}

		sm.eventChan <- event

		if callback, ok := sm.changeCallbacks[droneID]; ok {
			go callback(event)
		}

		if callback, ok := sm.changeCallbacks["*"]; ok {
			go callback(event)
		}
	}

	sm.checkAlerts(droneID, battery)

	return nil
}

func (sm *StatusMonitor) GetDroneStatus(droneID string) (Drones.DroneStatus, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	status, exists := sm.droneStatus[droneID]
	if !exists {
		return "", fmt.Errorf("无人机状态不存在: %s", droneID)
	}

	return status, nil
}

func (sm *StatusMonitor) GetAllDroneStatuses() map[string]Drones.DroneStatus {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string]Drones.DroneStatus)
	for droneID, status := range sm.droneStatus {
		result[droneID] = status
	}

	return result
}

func (sm *StatusMonitor) GetStatusHistory(droneID string) []*DroneStatusLog {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	logs, exists := sm.statusHistory[droneID]
	if !exists {
		return nil
	}

	result := make([]*DroneStatusLog, len(logs))
	copy(result, logs)

	return result
}

func (sm *StatusMonitor) RegisterStatusChangeCallback(droneID string, callback StatusChangeCallback) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.changeCallbacks[droneID] = callback
}

func (sm *StatusMonitor) UnregisterStatusChangeCallback(droneID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.changeCallbacks, droneID)
}

func (sm *StatusMonitor) GetLogChan() <-chan *DroneStatusLog {
	return sm.logChan
}

func (sm *StatusMonitor) GetEventChan() <-chan *StatusChangeEvent {
	return sm.eventChan
}

func (sm *StatusMonitor) StartLogging() {
	go func() {
		for {
			select {
			case log := <-sm.logChan:
				sm.logStatus(log)
			}
		}
	}()
}

func (sm *StatusMonitor) logStatus(log *DroneStatusLog) {
	fmt.Printf("[%s] Drone %s: %s - %s\n",
		log.Timestamp.Format("2006-01-02 15:04:05"),
		log.DroneID,
		log.Status,
		log.Message)
}

func (sm *StatusMonitor) ClearHistory(droneID string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.statusHistory, droneID)
}

func (sm *StatusMonitor) ClearAllHistory() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.statusHistory = make(map[string][]*DroneStatusLog)
}

func (sm *StatusMonitor) GetHistorySize(droneID string) int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	if logs, exists := sm.statusHistory[droneID]; exists {
		return len(logs)
	}

	return 0
}

func (sm *StatusMonitor) addLog(droneID string, log *DroneStatusLog) {
	if _, exists := sm.statusHistory[droneID]; !exists {
		sm.statusHistory[droneID] = make([]*DroneStatusLog, 0)
	}

	sm.statusHistory[droneID] = append(sm.statusHistory[droneID], log)

	if len(sm.statusHistory[droneID]) > sm.maxHistorySize {
		sm.statusHistory[droneID] = sm.statusHistory[droneID][1:]
	}
}

func (sm *StatusMonitor) checkAlerts(droneID string, battery *Drones.BatteryStatus) {
	if battery == nil {
		return
	}

	threshold, exists := sm.alertThresholds[droneID]
	if !exists {
		return
	}

	if battery.Remaining <= threshold.BatteryCritical {
		event := &StatusChangeEvent{
			DroneID:   droneID,
			EventType: "battery_critical",
			OldStatus: "",
			NewStatus: "",
			Timestamp: time.Now(),
			Details:   fmt.Sprintf("电池电量严重不足: %d%%", battery.Remaining),
		}
		sm.eventChan <- event
	} else if battery.Remaining <= threshold.BatteryLow {
		event := &StatusChangeEvent{
			DroneID:   droneID,
			EventType: "battery_low",
			OldStatus: "",
			NewStatus: "",
			Timestamp: time.Now(),
			Details:   fmt.Sprintf("电池电量过低: %d%%", battery.Remaining),
		}
		sm.eventChan <- event
	}
}

func (sm *StatusMonitor) AnalyzeStatus(droneID string) (map[string]interface{}, error) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	logs, exists := sm.statusHistory[droneID]
	if !exists || len(logs) == 0 {
		return nil, fmt.Errorf("没有状态历史记录: %s", droneID)
	}

	analysis := make(map[string]interface{})

	statusCounts := make(map[string]int)
	for _, log := range logs {
		statusCounts[string(log.Status)]++
	}
	analysis["status_counts"] = statusCounts

	oldest := logs[0]
	newest := logs[len(logs)-1]
	duration := newest.Timestamp.Sub(oldest.Timestamp)
	analysis["monitoring_duration"] = duration.Seconds()
	analysis["log_count"] = len(logs)

	return analysis, nil
}
