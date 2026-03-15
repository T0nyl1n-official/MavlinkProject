package Drones

import (
	"sync"
	"time"
)

type DroneStatus string

const (
	StatusIdle         DroneStatus = "idle"
	StatusArmed        DroneStatus = "armed"
	StatusFlying       DroneStatus = "flying"
	StatusReturning    DroneStatus = "returning"
	StatusLanding      DroneStatus = "landing"
	StatusEmergency    DroneStatus = "emergency"
	StatusDisconnected DroneStatus = "disconnected"
)

type Position struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude"`
}

type Attitude struct {
	Roll  float64 `json:"roll"`
	Pitch float64 `json:"pitch"`
	Yaw   float64 `json:"yaw"`
}

type Velocity struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
	Z float64 `json:"z"`
}

type BatteryStatus struct {
	Voltage     float64 `json:"voltage"`
	Current     float64 `json:"current"`
	Remaining   int     `json:"remaining"`
	Temperature float64 `json:"temperature"`
	CellCount   int     `json:"cell_count"`
}

// 摄像头子对象 (备用)
type Camera struct {
	Model        string `json:"model"`
	Resolution   string `json:"resolution"`
	StorageGB    int    `json:"storage_gb"`
	PhotoCount   int    `json:"photo_count"`
	VideoSupport bool   `json:"video_support"`
}

type DroneConfig struct {
	SystemID        int           `json:"system_id"`
	ComponentID     int           `json:"component_id"`
	ProtocolVersion string        `json:"protocol_version"`
	HeartbeatRate   time.Duration `json:"heartbeat_rate"`
	Timeout         time.Duration `json:"timeout"`
}

type Drone struct {
	ID     string      `json:"id" gorm:"primaryKey;uniqueIndex"`
	Name   string      `json:"name"`
	Model  string      `json:"model"`
	Config DroneConfig `json:"config"`

	Status   DroneStatus   `json:"status"`
	Position Position      `json:"position"`
	Attitude Attitude      `json:"attitude"`
	Velocity Velocity      `json:"velocity"`
	Battery  BatteryStatus `json:"battery"`
	Camera   Camera        `json:"camera"`

	LastHeartbeat time.Time `json:"last_heartbeat"`
	Connected     bool      `json:"connected"`

	mu             sync.RWMutex
	eventCallbacks map[string]EventCallback
}

type EventCallback func(event DroneEvent)

type DroneEvent struct {
	Type    string
	DroneID string
	Data    interface{}
	Time    time.Time
}

func NewDrone(id, name, model string, config DroneConfig) *Drone {
	return &Drone{
		ID:             id,
		Name:           name,
		Model:          model,
		Config:         config,
		Status:         StatusIdle,
		Connected:      false,
		eventCallbacks: make(map[string]EventCallback),
		Position:       Position{},
		Attitude:       Attitude{},
		Velocity:       Velocity{},
		Battery:        BatteryStatus{},
	}
}

func (d *Drone) GetID() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.ID
}

func (d *Drone) GetName() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Name
}

func (d *Drone) GetStatus() DroneStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Status
}

func (d *Drone) SetStatus(status DroneStatus) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Status = status
	d.emitEvent(DroneEvent{
		Type:    "status_changed",
		DroneID: d.ID,
		Data:    status,
		Time:    time.Now(),
	})
}

func (d *Drone) GetPosition() Position {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Position
}

func (d *Drone) SetPosition(pos Position) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Position = pos
	d.emitEvent(DroneEvent{
		Type:    "position_updated",
		DroneID: d.ID,
		Data:    pos,
		Time:    time.Now(),
	})
}

func (d *Drone) GetAttitude() Attitude {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Attitude
}

func (d *Drone) SetAttitude(att Attitude) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Attitude = att
	d.emitEvent(DroneEvent{
		Type:    "attitude_updated",
		DroneID: d.ID,
		Data:    att,
		Time:    time.Now(),
	})
}

func (d *Drone) GetVelocity() Velocity {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Velocity
}

func (d *Drone) SetVelocity(vel Velocity) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Velocity = vel
	d.emitEvent(DroneEvent{
		Type:    "velocity_updated",
		DroneID: d.ID,
		Data:    vel,
		Time:    time.Now(),
	})
}

func (d *Drone) GetBattery() BatteryStatus {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Battery
}

func (d *Drone) SetBattery(bat BatteryStatus) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Battery = bat
	d.emitEvent(DroneEvent{
		Type:    "battery_updated",
		DroneID: d.ID,
		Data:    bat,
		Time:    time.Now(),
	})
}

func (d *Drone) IsConnected() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Connected
}

func (d *Drone) SetConnected(connected bool) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Connected = connected
	d.emitEvent(DroneEvent{
		Type:    "connection_changed",
		DroneID: d.ID,
		Data:    connected,
		Time:    time.Now(),
	})
}

func (d *Drone) UpdateHeartbeat() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.LastHeartbeat = time.Now()
	if !d.Connected {
		d.Connected = true
	}
}

func (d *Drone) IsTimedOut(timeout time.Duration) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return time.Since(d.LastHeartbeat) > timeout
}

func (d *Drone) RegisterCallback(id string, callback EventCallback) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.eventCallbacks[id] = callback
}

func (d *Drone) UnregisterCallback(id string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.eventCallbacks, id)
}

func (d *Drone) emitEvent(event DroneEvent) {
	d.mu.RLock()
	callbacks := make([]EventCallback, 0, len(d.eventCallbacks))
	for _, cb := range d.eventCallbacks {
		callbacks = append(callbacks, cb)
	}
	d.mu.RUnlock()

	for _, callback := range callbacks {
		callback(event)
	}
}

func (d *Drone) GetConfig() DroneConfig {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Config
}

func (d *Drone) SetConfig(config DroneConfig) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Config = config
}

func (d *Drone) HasCamera() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Camera.Model != ""
}

func (d *Drone) GetCamera() Camera {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Camera
}

func (d *Drone) SetCamera(camera Camera) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.Camera = camera
}

func (d *Drone) CanTakePhoto() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.Camera.Model != "" && d.Camera.PhotoCount < d.Camera.StorageGB*1000
}

func (d *Drone) TakePhoto() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.Camera.Model == "" {
		return ErrNoCamera
	}
	if d.Camera.PhotoCount >= d.Camera.StorageGB*1000 {
		return ErrStorageFull
	}
	d.Camera.PhotoCount++
	return nil
}

var ErrNoCamera = &DroneError{"drone has no camera"}
var ErrStorageFull = &DroneError{"camera storage is full"}

type DroneError struct {
	Message string
}

func (e *DroneError) Error() string {
	return e.Message
}
