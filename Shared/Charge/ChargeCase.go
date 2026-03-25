package Charging

import (
	"sync"
	"time"
)

type ChargingStatus string

const (
	StatusIdle     ChargingStatus = "idle"
	StatusCharging ChargingStatus = "charging"
	StatusOccupied ChargingStatus = "occupied"
	StatusError    ChargingStatus = "error"
	StatusOffline  ChargingStatus = "offline"
)

type ChargingCase struct {
	ID         string         `json:"id" gorm:"primaryKey;uniqueIndex"`
	Name       string         `json:"name"`
	Status     ChargingStatus `json:"status"`
	Latitude   float64        `json:"latitude"`
	Longitude  float64        `json:"longitude"`
	Altitude   float64        `json:"altitude"`
	Voltage    float64        `json:"voltage"`
	Current    float64        `json:"current"`
	Power      float64        `json:"power"`
	DroneID    string         `json:"drone_id"`
	LastUpdate time.Time      `json:"last_update"`

	mu sync.RWMutex
}

func NewChargingCase(id, name string, lat, lng, alt float64) *ChargingCase {
	return &ChargingCase{
		ID:         id,
		Name:       name,
		Status:     StatusIdle,
		Latitude:   lat,
		Longitude:  lng,
		Altitude:   alt,
		LastUpdate: time.Now(),
	}
}

func (c *ChargingCase) SetStatus(status ChargingStatus) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Status = status
	c.LastUpdate = time.Now()
}

func (c *ChargingCase) Occupy(droneID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.Status != StatusIdle {
		return ErrChargingCaseNotAvailable
	}

	c.Status = StatusOccupied
	c.DroneID = droneID
	c.LastUpdate = time.Now()
	return nil
}

func (c *ChargingCase) Release() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Status = StatusIdle
	c.DroneID = ""
	c.LastUpdate = time.Now()
}

func (c *ChargingCase) IsAvailable() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Status == StatusIdle
}

func (c *ChargingCase) GetInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"id":          c.ID,
		"name":        c.Name,
		"status":      c.Status,
		"latitude":    c.Latitude,
		"longitude":   c.Longitude,
		"altitude":    c.Altitude,
		"voltage":     c.Voltage,
		"current":     c.Current,
		"power":       c.Power,
		"drone_id":    c.DroneID,
		"last_update": c.LastUpdate,
	}
}

var ErrChargingCaseNotAvailable = &ChargingError{"charging case is not available"}
var ErrChargingCaseNotFound = &ChargingError{"charging case not found"}

type ChargingError struct {
	Message string
}

func (e *ChargingError) Error() string {
	return e.Message
}

type ChargingCaseManager struct {
	cases map[string]*ChargingCase
	mu    sync.RWMutex
}

var defaultChargingManager *ChargingCaseManager

func init() {
	defaultChargingManager = NewChargingCaseManager()
}

func NewChargingCaseManager() *ChargingCaseManager {
	return &ChargingCaseManager{
		cases: make(map[string]*ChargingCase),
	}
}

func GetChargingManager() *ChargingCaseManager {
	return defaultChargingManager
}

func (m *ChargingCaseManager) Add(caseInfo *ChargingCase) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cases[caseInfo.ID] = caseInfo
}

func (m *ChargingCaseManager) Get(id string) *ChargingCase {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cases[id]
}

func (m *ChargingCaseManager) GetAll() []*ChargingCase {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ChargingCase, 0, len(m.cases))
	for _, c := range m.cases {
		result = append(result, c)
	}
	return result
}

func (m *ChargingCaseManager) GetAvailable() []*ChargingCase {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*ChargingCase, 0)
	for _, c := range m.cases {
		if c.Status == StatusIdle {
			result = append(result, c)
		}
	}
	return result
}

func (m *ChargingCaseManager) Remove(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.cases, id)
}

func (m *ChargingCaseManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.cases)
}
