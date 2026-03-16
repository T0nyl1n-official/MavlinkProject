package Mavlink

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"
)

type ChainStatus string

const (
	ChainStatusActive   ChainStatus = "active"
	ChainStatusComplete ChainStatus = "complete"
	ChainStatusForceEnd ChainStatus = "force_end"
)

type DispatchRecord struct {
	HandlerID string    `json:"handler_id"`
	Route     string    `json:"route"`
	Timestamp time.Time `json:"timestamp"`
	Params    string    `json:"params,omitempty"`
	Result    string    `json:"result,omitempty"`
	Success   bool      `json:"success"`
}

type DispatchChain struct {
	ChainID   string           `json:"chain_id"`
	Status    ChainStatus      `json:"status"`
	Records   []DispatchRecord `json:"records"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	mu        sync.RWMutex
}

func (c *DispatchChain) GetID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.ChainID
}

func generateChainID() string {
	const digits = "0123456789"
	var result string

	for i := 0; i < 16; i++ {
		d, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			result += "0"
			continue
		}
		result += string(digits[d.Int64()])
	}
	return result
}

func NewDispatchChain() *DispatchChain {
	return &DispatchChain{
		ChainID:   generateChainID(),
		Status:    ChainStatusActive,
		Records:   make([]DispatchRecord, 0),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func (c *DispatchChain) AddRecord(handlerID, route, params, result string, success bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	record := DispatchRecord{
		HandlerID: handlerID,
		Route:     route,
		Timestamp: time.Now(),
		Params:    params,
		Result:    result,
		Success:   success,
	}

	c.Records = append(c.Records, record)
	c.UpdatedAt = time.Now()

	if len(c.Records) >= 1000 {
		c.Status = ChainStatusForceEnd
	}
}

func (c *DispatchChain) GetRecords() []DispatchRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]DispatchRecord, len(c.Records))
	copy(result, c.Records)
	return result
}

func (c *DispatchChain) GetRecordCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.Records)
}

func (c *DispatchChain) IsFull() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.Records) >= 1000
}

func (c *DispatchChain) IsActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Status == ChainStatusActive
}

func (c *DispatchChain) Complete() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Status = ChainStatusComplete
	c.UpdatedAt = time.Now()
}

func (c *DispatchChain) ForceEnd() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Status = ChainStatusForceEnd
	c.UpdatedAt = time.Now()
}

func (c *DispatchChain) GetInfo() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return map[string]interface{}{
		"chain_id":     c.ChainID,
		"status":       c.Status,
		"record_count": len(c.Records),
		"created_at":   c.CreatedAt,
		"updated_at":   c.UpdatedAt,
	}
}

type ChainManager struct {
	chains    map[string]*DispatchChain
	ID 		  string
	mu        sync.RWMutex
	maxSize   int
}

func NewChainManager() *ChainManager {
	return &ChainManager{
		chains:  make(map[string]*DispatchChain),
		maxSize: 1000,
	}
}

func (cm *ChainManager) CreateChain() string {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	chain := NewDispatchChain()
	cm.chains[chain.ChainID] = chain
	cm.ID = chain.ChainID

	return chain.ChainID
}

func (cm *ChainManager) GetCurrentChain() *DispatchChain {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if cm.ID == "" {
		return nil
	}
	return cm.chains[cm.ID]
}

func (cm *ChainManager) GetChain(chainID string) *DispatchChain {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.chains[chainID]
}

func (cm *ChainManager) GetChainID() string {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.ID
}

func (cm *ChainManager) AddRecordToCurrentChain(handlerID, route, params, result string, success bool) error {
	cm.mu.RLock()
	chain, ok := cm.chains[cm.ID]
	cm.mu.RUnlock()

	if !ok || chain == nil {
		return fmt.Errorf("no active chain")
	}

	if chain.IsFull() {
		return fmt.Errorf("chain is full, please create new chain")
	}

	chain.AddRecord(handlerID, route, params, result, success)

	if chain.IsFull() {
		return fmt.Errorf("chain reached max size (1000), new chain required")
	}

	return nil
}

func (cm *ChainManager) SwitchChain(chainID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, ok := cm.chains[chainID]; !ok {
		return fmt.Errorf("chain not found: %s", chainID)
	}

	cm.ID = chainID
	return nil
}

func (cm *ChainManager) CreateNewChainAndSwitch() string {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	chain := NewDispatchChain()
	cm.chains[chain.ChainID] = chain
	cm.ID = chain.ChainID

	return chain.ChainID
}

func (cm *ChainManager) DeleteChain(chainID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, ok := cm.chains[chainID]; !ok {
		return fmt.Errorf("chain not found")
	}

	if cm.ID == chainID {
		cm.ID = ""
	}

	delete(cm.chains, chainID)
	return nil
}

func (cm *ChainManager) GetAllChains() []*DispatchChain {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*DispatchChain, 0, len(cm.chains))
	for _, chain := range cm.chains {
		result = append(result, chain)
	}
	return result
}

func (cm *ChainManager) GetChainCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.chains)
}

var globalChainManager = NewChainManager()

func GetChainManager() *ChainManager {
	return globalChainManager
}
