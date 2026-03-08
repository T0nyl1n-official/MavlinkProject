package mavlink

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

type DataPacket struct {
	ID        string    `json:"id"`
	Sequence  uint32    `json:"sequence"`
	Timestamp time.Time `json:"timestamp"`
	Data      []byte    `json:"data"`
	Checksum  string    `json:"checksum"`
	RetryCount int     `json:"retry_count"`
}

type CorrectionResult struct {
	Corrected bool
	Data      []byte
	Errors    []string
}

type ConnectionState struct {
	Connected    bool
	LastAttempt  time.Time
	RetryCount   int
	MaxRetries   int
	BackoffDelay time.Duration
}

type ErrorCorrector struct {
	mu             sync.RWMutex
	receivedPackets map[uint32]*DataPacket
	sequenceNumber  uint32
	packetBuffer    map[string][]*DataPacket
	
	connectionStates map[string]*ConnectionState
	defaultMaxRetries int
	defaultBackoff   time.Duration
	
	reconnectHandlers map[string]ReconnectHandler
}

type ReconnectHandler func(droneID string) error

func NewErrorCorrector(maxRetries int, backoff time.Duration) *ErrorCorrector {
	if maxRetries == 0 {
		maxRetries = 3
	}
	if backoff == 0 {
		backoff = 1 * time.Second
	}
	
	return &ErrorCorrector{
		receivedPackets: make(map[uint32]*DataPacket),
		packetBuffer: make(map[string][]*DataPacket),
		connectionStates: make(map[string]*ConnectionState),
		defaultMaxRetries: maxRetries,
		defaultBackoff: backoff,
		reconnectHandlers: make(map[string]ReconnectHandler),
	}
}

func (ec *ErrorCorrector) CalculateChecksum(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func (ec *ErrorCorrector) VerifyChecksum(packet *DataPacket) bool {
	calculated := ec.CalculateChecksum(packet.Data)
	return calculated == packet.Checksum
}

func (ec *ErrorCorrector) GenerateChecksum(data []byte) string {
	return ec.CalculateChecksum(data)
}

func (ec *ErrorCorrector) CreatePacket(data []byte) *DataPacket {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	ec.sequenceNumber++
	
	packet := &DataPacket{
		ID:        fmt.Sprintf("pkt_%d", ec.sequenceNumber),
		Sequence:  ec.sequenceNumber,
		Timestamp: time.Now(),
		Data:      data,
		Checksum:  ec.CalculateChecksum(data),
		RetryCount: 0,
	}
	
	return packet
}

func (ec *ErrorCorrector) ValidatePacket(packet *DataPacket) *CorrectionResult {
	result := &CorrectionResult{
		Corrected: false,
		Data:      packet.Data,
		Errors:    make([]string, 0),
	}
	
	if packet.Data == nil || len(packet.Data) == 0 {
		result.Errors = append(result.Errors, "数据包为空")
		return result
	}
	
	if !ec.VerifyChecksum(packet) {
		result.Errors = append(result.Errors, "校验和不匹配")
	}
	
	if time.Since(packet.Timestamp) > 30*time.Second {
		result.Errors = append(result.Errors, "数据包已过期")
	}
	
	if len(result.Errors) == 0 {
		result.Corrected = true
	}
	
	return result
}

func (ec *ErrorCorrector) ProcessReceivedPacket(packet *DataPacket) (*DataPacket, error) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	if _, exists := ec.receivedPackets[packet.Sequence]; exists {
		return nil, fmt.Errorf("重复数据包: 序列号 %d", packet.Sequence)
	}
	
	validation := ec.ValidatePacket(packet)
	if !validation.Corrected {
		return nil, fmt.Errorf("数据包验证失败: %v", validation.Errors)
	}
	
	ec.receivedPackets[packet.Sequence] = packet
	return packet, nil
}

func (ec *ErrorCorrector) BufferPacket(droneID string, packet *DataPacket) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	ec.packetBuffer[droneID] = append(ec.packetBuffer[droneID], packet)
	
	if len(ec.packetBuffer[droneID]) > 100 {
		ec.packetBuffer[droneID] = ec.packetBuffer[droneID][1:]
	}
}

func (ec *ErrorCorrector) GetBufferedPackets(droneID string) []*DataPacket {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	
	packets := ec.packetBuffer[droneID]
	result := make([]*DataPacket, len(packets))
	copy(result, packets)
	
	return result
}

func (ec *ErrorCorrector) ClearBuffer(droneID string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	delete(ec.packetBuffer, droneID)
}

func (ec *ErrorCorrector) InitConnection(droneID string, maxRetries int, backoff time.Duration) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	if maxRetries == 0 {
		maxRetries = ec.defaultMaxRetries
	}
	if backoff == 0 {
		backoff = ec.defaultBackoff
	}
	
	ec.connectionStates[droneID] = &ConnectionState{
		Connected:    true,
		LastAttempt:  time.Now(),
		RetryCount:   0,
		MaxRetries:   maxRetries,
		BackoffDelay: backoff,
	}
}

func (ec *ErrorCorrector) SetConnectionState(droneID string, connected bool) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	if state, exists := ec.connectionStates[droneID]; exists {
		state.Connected = connected
		if connected {
			state.RetryCount = 0
		}
	}
}

func (ec *ErrorCorrector) GetConnectionState(droneID string) (*ConnectionState, error) {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	
	state, exists := ec.connectionStates[droneID]
	if !exists {
		return nil, fmt.Errorf("连接状态不存在: %s", droneID)
	}
	
	return state, nil
}

func (ec *ErrorCorrector) IsConnected(droneID string) bool {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	
	if state, exists := ec.connectionStates[droneID]; exists {
		return state.Connected
	}
	
	return false
}

func (ec *ErrorCorrector) AttemptReconnect(droneID string) error {
	ec.mu.Lock()
	state, exists := ec.connectionStates[droneID]
	ec.mu.Unlock()
	
	if !exists {
		return fmt.Errorf("连接状态不存在: %s", droneID)
	}
	
	if state.RetryCount >= state.MaxRetries {
		return fmt.Errorf("超过最大重试次数: %d", state.MaxRetries)
	}
	
	state.RetryCount++
	state.LastAttempt = time.Now()
	
	if handler, ok := ec.reconnectHandlers[droneID]; ok {
		return handler(droneID)
	}
	
	return nil
}

func (ec *ErrorCorrector) RegisterReconnectHandler(droneID string, handler ReconnectHandler) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	ec.reconnectHandlers[droneID] = handler
}

func (ec *ErrorCorrector) UnregisterReconnectHandler(droneID string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	delete(ec.reconnectHandlers, droneID)
}

func (ec *ErrorCorrector) GetRetryCount(droneID string) int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	
	if state, exists := ec.connectionStates[droneID]; exists {
		return state.RetryCount
	}
	
	return 0
}

func (ec *ErrorCorrector) CanRetry(droneID string) bool {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	
	if state, exists := ec.connectionStates[droneID]; exists {
		return state.RetryCount < state.MaxRetries
	}
	
	return false
}

func (ec *ErrorCorrector) CalculateBackoff(droneID string) time.Duration {
	ec.mu.RLock()
	state, _ := ec.connectionStates[droneID]
	ec.mu.RUnlock()
	
	if state == nil {
		return ec.defaultBackoff
	}
	
	backoff := state.BackoffDelay * time.Duration(state.RetryCount+1)
	if backoff > 30*time.Second {
		return 30 * time.Second
	}
	
	return backoff
}

func (ec *ErrorCorrector) ResetConnection(droneID string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	if state, exists := ec.connectionStates[droneID]; exists {
		state.RetryCount = 0
		state.Connected = true
		state.LastAttempt = time.Now()
	}
}

func (ec *ErrorCorrector) CloseConnection(droneID string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	if state, exists := ec.connectionStates[droneID]; exists {
		state.Connected = false
	}
}

func (ec *ErrorCorrector) RemoveConnection(droneID string) {
	ec.mu.Lock()
	defer ec.mu.Unlock()
	
	delete(ec.connectionStates, droneID)
	delete(ec.packetBuffer, droneID)
	delete(ec.reconnectHandlers, droneID)
}

func (ec *ErrorCorrector) GetConnectionCount() int {
	ec.mu.RLock()
	defer ec.mu.RUnlock()
	
	count := 0
	for _, state := range ec.connectionStates {
		if state.Connected {
			count++
		}
	}
	
	return count
}

type CommandBackup struct {
	CommandID    string    `json:"command_id"`
	CommandData  []byte    `json:"command_data"`
	CreatedAt    time.Time `json:"created_at"`
	DroneID      string    `json:"drone_id"`
	Executed     bool      `json:"executed"`
	ExecutionAt  time.Time `json:"execution_at,omitempty"`
}

type CommandBackupManager struct {
	mu       sync.RWMutex
	backups  map[string]*CommandBackup
	maxSize  int
}

func NewCommandBackupManager(maxSize int) *CommandBackupManager {
	if maxSize == 0 {
		maxSize = 100
	}
	
	return &CommandBackupManager{
		backups: make(map[string]*CommandBackup),
		maxSize: maxSize,
	}
}

func (cbm *CommandBackupManager) BackupCommand(cmdID string, data []byte, droneID string) error {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()
	
	if len(cbm.backups) >= cbm.maxSize {
		cbm.removeOldest()
	}
	
	backup := &CommandBackup{
		CommandID: cmdID,
		CommandData: data,
		CreatedAt: time.Now(),
		DroneID:   droneID,
		Executed:  false,
	}
	
	cbm.backups[cmdID] = backup
	return nil
}

func (cbm *CommandBackupManager) GetBackup(cmdID string) (*CommandBackup, error) {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()
	
	backup, exists := cbm.backups[cmdID]
	if !exists {
		return nil, fmt.Errorf("命令备份不存在: %s", cmdID)
	}
	
	return backup, nil
}

func (cbm *CommandBackupManager) MarkExecuted(cmdID string) error {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()
	
	backup, exists := cbm.backups[cmdID]
	if !exists {
		return fmt.Errorf("命令备份不存在: %s", cmdID)
	}
	
	backup.Executed = true
	backup.ExecutionAt = time.Now()
	
	return nil
}

func (cbm *CommandBackupManager) RemoveBackup(cmdID string) error {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()
	
	if _, exists := cbm.backups[cmdID]; !exists {
		return fmt.Errorf("命令备份不存在: %s", cmdID)
	}
	
	delete(cbm.backups, cmdID)
	return nil
}

func (cbm *CommandBackupManager) GetPendingBackups(droneID string) []*CommandBackup {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()
	
	var result []*CommandBackup
	for _, backup := range cbm.backups {
		if backup.DroneID == droneID && !backup.Executed {
			result = append(result, backup)
		}
	}
	
	return result
}

func (cbm *CommandBackupManager) removeOldest() {
	var oldest *CommandBackup
	var oldestID string
	
	for id, backup := range cbm.backups {
		if oldest == nil || backup.CreatedAt.Before(oldest.CreatedAt) {
			oldest = backup
			oldestID = id
		}
	}
	
	if oldestID != "" {
		delete(cbm.backups, oldestID)
	}
}

func (cbm *CommandBackupManager) Clear() {
	cbm.mu.Lock()
	defer cbm.mu.Unlock()
	
	cbm.backups = make(map[string]*CommandBackup)
}

func (cbm *CommandBackupManager) GetBackupCount() int {
	cbm.mu.RLock()
	defer cbm.mu.RUnlock()
	
	return len(cbm.backups)
}
