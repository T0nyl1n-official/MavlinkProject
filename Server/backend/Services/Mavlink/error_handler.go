package MavlinkService

import (
	"fmt"
	"sync"
	"time"
)

type ErrorLevel string

const (
	ErrorLevelInfo     ErrorLevel = "info"
	ErrorLevelWarning  ErrorLevel = "warning"
	ErrorLevelError    ErrorLevel = "error"
	ErrorLevelCritical ErrorLevel = "critical"
)

type ErrorCategory string

const (
	ErrCategoryConnection ErrorCategory = "connection"
	ErrCategoryProtocol   ErrorCategory = "protocol"
	ErrCategoryCommand    ErrorCategory = "command"
	ErrCategoryTimeout    ErrorCategory = "timeout"
	ErrCategoryValidation ErrorCategory = "validation"
	ErrCategoryResource   ErrorCategory = "resource"
	ErrCategoryUnknown    ErrorCategory = "unknown"
)

type MavlinkError struct {
	ID          string
	Category    ErrorCategory
	Level       ErrorLevel
	Message     string
	Details     string
	Source      string
	Timestamp   time.Time
	Recoverable bool
	DroneID     string
}

type ErrorRecovery struct {
	ErrorID      string
	Strategy     string
	AttemptCount int
	MaxAttempts  int
	LastAttempt  time.Time
	Success      bool
}

type ErrorHandler struct {
	mu           sync.RWMutex
	errors       map[string]*MavlinkError
	errorHistory []*MavlinkError
	recoveries   map[string]*ErrorRecovery

	maxHistorySize int
	errorChan      chan *MavlinkError
	alertChan      chan *MavlinkError

	errorListeners map[string]ErrorListener
}

type ErrorListener func(err *MavlinkError)

func NewErrorHandler(maxHistorySize int) *ErrorHandler {
	if maxHistorySize == 0 {
		maxHistorySize = 1000
	}

	return &ErrorHandler{
		errors:         make(map[string]*MavlinkError),
		errorHistory:   make([]*MavlinkError, 0),
		recoveries:     make(map[string]*ErrorRecovery),
		maxHistorySize: maxHistorySize,
		errorChan:      make(chan *MavlinkError, 100),
		alertChan:      make(chan *MavlinkError, 50),
		errorListeners: make(map[string]ErrorListener),
	}
}

func (eh *ErrorHandler) generateErrorID() string {
	return fmt.Sprintf("err_%d", time.Now().UnixNano())
}

func (eh *ErrorHandler) HandleError(category ErrorCategory, level ErrorLevel, message, details, source, droneID string, recoverable bool) *MavlinkError {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	err := &MavlinkError{
		ID:          eh.generateErrorID(),
		Category:    category,
		Level:       level,
		Message:     message,
		Details:     details,
		Source:      source,
		Timestamp:   time.Now(),
		Recoverable: recoverable,
		DroneID:     droneID,
	}

	eh.errors[err.ID] = err
	eh.errorHistory = append(eh.errorHistory, err)

	if len(eh.errorHistory) > eh.maxHistorySize {
		eh.errorHistory = eh.errorHistory[1:]
	}

	select {
	case eh.errorChan <- err:
	default:
	}

	if level == ErrorLevelCritical || level == ErrorLevelError {
		select {
		case eh.alertChan <- err:
		default:
		}
	}

	for _, listener := range eh.errorListeners {
		listener(err)
	}

	return err
}

func (eh *ErrorHandler) GetError(errorID string) (*MavlinkError, error) {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	err, exists := eh.errors[errorID]
	if !exists {
		return nil, fmt.Errorf("错误不存在: %s", errorID)
	}

	return err, nil
}

func (eh *ErrorHandler) GetErrorsByCategory(category ErrorCategory) []*MavlinkError {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	var result []*MavlinkError
	for _, err := range eh.errors {
		if err.Category == category {
			result = append(result, err)
		}
	}

	return result
}

func (eh *ErrorHandler) GetErrorsByLevel(level ErrorLevel) []*MavlinkError {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	var result []*MavlinkError
	for _, err := range eh.errors {
		if err.Level == level {
			result = append(result, err)
		}
	}

	return result
}

func (eh *ErrorHandler) GetErrorsByDrone(droneID string) []*MavlinkError {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	var result []*MavlinkError
	for _, err := range eh.errors {
		if err.DroneID == droneID {
			result = append(result, err)
		}
	}

	return result
}

func (eh *ErrorHandler) GetErrorHistory() []*MavlinkError {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	result := make([]*MavlinkError, len(eh.errorHistory))
	copy(result, eh.errorHistory)

	return result
}

func (eh *ErrorHandler) ClearErrors() {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	eh.errors = make(map[string]*MavlinkError)
	eh.errorHistory = make([]*MavlinkError, 0)
}

func (eh *ErrorHandler) ClearDroneErrors(droneID string) {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	for id, err := range eh.errors {
		if err.DroneID == droneID {
			delete(eh.errors, id)
		}
	}

	newHistory := make([]*MavlinkError, 0)
	for _, err := range eh.errorHistory {
		if err.DroneID != droneID {
			newHistory = append(newHistory, err)
		}
	}
	eh.errorHistory = newHistory
}

func (eh *ErrorHandler) StartRecovery(errorID, strategy string, maxAttempts int) *ErrorRecovery {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	recovery := &ErrorRecovery{
		ErrorID:      errorID,
		Strategy:     strategy,
		MaxAttempts:  maxAttempts,
		AttemptCount: 0,
	}

	eh.recoveries[errorID] = recovery
	return recovery
}

func (eh *ErrorHandler) UpdateRecovery(errorID string, success bool) {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	if recovery, exists := eh.recoveries[errorID]; exists {
		recovery.AttemptCount++
		recovery.LastAttempt = time.Now()
		recovery.Success = success

		if success || recovery.AttemptCount >= recovery.MaxAttempts {
			delete(eh.recoveries, errorID)
		}
	}
}

func (eh *ErrorHandler) GetActiveRecoveries() []*ErrorRecovery {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	result := make([]*ErrorRecovery, 0, len(eh.recoveries))
	for _, recovery := range eh.recoveries {
		result = append(result, recovery)
	}

	return result
}

func (eh *ErrorHandler) RegisterListener(id string, listener ErrorListener) {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	eh.errorListeners[id] = listener
}

func (eh *ErrorHandler) UnregisterListener(id string) {
	eh.mu.Lock()
	defer eh.mu.Unlock()

	delete(eh.errorListeners, id)
}

func (eh *ErrorHandler) GetErrorChan() <-chan *MavlinkError {
	return eh.errorChan
}

func (eh *ErrorHandler) GetAlertChan() <-chan *MavlinkError {
	return eh.alertChan
}

func (eh *ErrorHandler) GetErrorCount() int {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	return len(eh.errors)
}

func (eh *ErrorHandler) GetCriticalErrorCount() int {
	eh.mu.RLock()
	defer eh.mu.RUnlock()

	count := 0
	for _, err := range eh.errors {
		if err.Level == ErrorLevelCritical {
			count++
		}
	}

	return count
}

func (eh *ErrorHandler) CreateConnectionError(details, source, droneID string, recoverable bool) *MavlinkError {
	return eh.HandleError(
		ErrCategoryConnection,
		ErrorLevelError,
		"连接错误",
		details,
		source,
		droneID,
		recoverable,
	)
}

func (eh *ErrorHandler) CreateProtocolError(details, source, droneID string, recoverable bool) *MavlinkError {
	return eh.HandleError(
		ErrCategoryProtocol,
		ErrorLevelError,
		"协议错误",
		details,
		source,
		droneID,
		recoverable,
	)
}

func (eh *ErrorHandler) CreateTimeoutError(source, droneID string, recoverable bool) *MavlinkError {
	return eh.HandleError(
		ErrCategoryTimeout,
		ErrorLevelWarning,
		"操作超时",
		"请求在规定时间内未收到响应",
		source,
		droneID,
		recoverable,
	)
}

func (eh *ErrorHandler) CreateCommandError(details, source, droneID string, recoverable bool) *MavlinkError {
	return eh.HandleError(
		ErrCategoryCommand,
		ErrorLevelError,
		"命令执行失败",
		details,
		source,
		droneID,
		recoverable,
	)
}
