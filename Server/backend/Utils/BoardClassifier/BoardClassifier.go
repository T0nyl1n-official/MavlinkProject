package BoardClassifier

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	Board "MavlinkProject/Server/Backend/Shared/Boards"
	WarningHandler "MavlinkProject/Server/Backend/Utils/WarningHandle"
)

type MessageCategory string

const (
	Category_Heartbeat MessageCategory = "Heartbeat"
	Category_Status    MessageCategory = "Status"
	Category_Mission   MessageCategory = "Mission"
	Category_Control   MessageCategory = "Control"
	Category_Command   MessageCategory = "Command"
	Category_Warning   MessageCategory = "Warning"
	Category_Error     MessageCategory = "Error"
	Category_Unknown   MessageCategory = "Unknown"
)

type ProcessAction string

const (
	Action_Ignore       ProcessAction = "Ignore"
	Action_Log          ProcessAction = "Log"
	Action_DispatchNext ProcessAction = "DispatchNext"
	Action_ReportError  ProcessAction = "ReportError"
	Action_Response     ProcessAction = "Response"
	Action_Reschedule   ProcessAction = "Reschedule"
)

type BoardMessageClassifier struct {
	logDir        string
	messageChan   chan *Board.BoardMessage
	processorChan chan *ProcessedMessage
	wg            sync.WaitGroup
	running       bool
	stopChan      chan bool
}

type ProcessedMessage struct {
	Original  *Board.BoardMessage
	Category  MessageCategory
	Action    ProcessAction
	Details   string
	Timestamp time.Time
}

var (
	classifier     *BoardMessageClassifier
	classifierOnce sync.Once
)

func NewBoardMessageClassifier() *BoardMessageClassifier {
	classifierOnce.Do(func() {
		logDir := filepath.Join(".", "logs", "boards")
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Printf("[BoardClassifier] Failed to create log directory: %v", err)
			logDir = "."
		}

		classifier = &BoardMessageClassifier{
			logDir:        logDir,
			messageChan:   make(chan *Board.BoardMessage, 1000),
			processorChan: make(chan *ProcessedMessage, 1000),
			stopChan:      make(chan bool),
		}
	})

	return classifier
}

func (bc *BoardMessageClassifier) ClassifyAndProcess(msg *Board.BoardMessage) {
	category := bc.classifyMessage(msg)
	action, details := bc.determineAction(msg, category)

	processed := &ProcessedMessage{
		Original:  msg,
		Category:  category,
		Action:    action,
		Details:   details,
		Timestamp: time.Now(),
	}

	bc.logToFile(processed)

	bc.processorChan <- processed

	bc.executeAction(processed)
}

func (bc *BoardMessageClassifier) classifyMessage(msg *Board.BoardMessage) MessageCategory {
	if msg.Message.Attribute != "" {
		switch msg.Message.Attribute {
		case Board.MessageAttribute_Status:
			return Category_Status
		case Board.MessageAttribute_Mission:
			return Category_Mission
		case Board.MessageAttribute_Control:
			return Category_Control
		case Board.MessageAttribute_Command:
			return Category_Command
		case Board.MessageAttribute_Warning:
			return Category_Warning
		}
	}

	command := msg.Message.Command
	switch command {
	case "Heartbeat", "Ping":
		return Category_Heartbeat
	case "Status", "GetStatus", "DetailResponse":
		return Category_Status
	case "Mission", "MissionAck", "MissionItem":
		return Category_Mission
	case "TakeOff", "Land", "GoTo", "SetSpeed", "SetPosition":
		return Category_Control
	case "TakePhoto", "SetConfig", "SetCamera":
		return Category_Command
	case "Error", "Warning", "LowBattery", "GPSLost", "MotorError":
		return Category_Warning
	default:
		return Category_Unknown
	}
}

func (bc *BoardMessageClassifier) determineAction(msg *Board.BoardMessage, category MessageCategory) (ProcessAction, string) {
	switch category {
	case Category_Heartbeat:
		if msg.Message.Data != nil {
			if status, ok := msg.Message.Data["status"].(string); ok {
				if status == "idle" || status == "waiting" {
					return Action_DispatchNext, "Board is waiting for next task"
				}
			}
		}
		return Action_Ignore, "Normal heartbeat"

	case Category_Status:
		return Action_Log, "Status update logged"

	case Category_Mission:
		if msg.Message.Data != nil {
			if missionStatus, ok := msg.Message.Data["mission_status"].(string); ok {
				if missionStatus == "completed" {
					return Action_DispatchNext, "Mission completed, dispatching next"
				} else if missionStatus == "failed" {
					return Action_Reschedule, "Mission failed, needs rescheduling"
				}
			}
		}
		return Action_Log, "Mission update logged"

	case Category_Control:
		return Action_Log, "Control command logged"

	case Category_Command:
		return Action_Response, "Command executed, sending response"

	case Category_Warning:
		return Action_ReportError, "Warning detected, reporting to agent"

	case Category_Error:
		return Action_ReportError, "Error detected, reporting to agent"

	default:
		return Action_Log, "Unknown message type"
	}
}

func (bc *BoardMessageClassifier) executeAction(processed *ProcessedMessage) {
	switch processed.Action {
	case Action_Ignore:
		log.Printf("[BoardClassifier] IGNORE: Board=%s, Command=%s",
			processed.Original.FromID, processed.Original.Message.Command)

	case Action_Log:
		log.Printf("[BoardClassifier] LOG: Board=%s, Category=%s, Details=%s",
			processed.Original.FromID, processed.Category, processed.Details)

	case Action_DispatchNext:
		log.Printf("[BoardClassifier] DISPATCH: Board=%s - Dispatching next task",
			processed.Original.FromID)
		bc.dispatchNextTask(processed)

	case Action_ReportError:
		log.Printf("[BoardClassifier] ERROR: Board=%s, Details=%s",
			processed.Original.FromID, processed.Details)
		bc.reportToAgent(processed)

	case Action_Response:
		log.Printf("[BoardClassifier] RESPONSE: Board=%s - Sending response",
			processed.Original.FromID)
		bc.sendResponse(processed)

	case Action_Reschedule:
		log.Printf("[BoardClassifier] RESCHEDULE: Board=%s - Re-scheduling tasks",
			processed.Original.FromID)
		bc.rescheduleTasks(processed)
	}
}

func (bc *BoardMessageClassifier) dispatchNextTask(processed *ProcessedMessage) {
	if processed.Original.Message.Data != nil {
		if taskID, ok := processed.Original.Message.Data["next_task_id"].(string); ok {
			log.Printf("[BoardClassifier] Next task ID: %s for board %s",
				taskID, processed.Original.FromID)
		}
	}
}

func (bc *BoardMessageClassifier) reportToAgent(processed *ProcessedMessage) {
	errMsg := fmt.Sprintf("Board %s reported: %s", processed.Original.FromID, processed.Details)

	warningDetail := map[string]interface{}{
		"board_id":  processed.Original.FromID,
		"command":   processed.Original.Message.Command,
		"category":  string(processed.Category),
		"action":    string(processed.Action),
		"timestamp": processed.Timestamp.Unix(),
	}

	WarningHandler.HandleAgentError(errMsg, processed.Original.FromID, "", processed.Details)
	_ = warningDetail
}

func (bc *BoardMessageClassifier) sendResponse(processed *ProcessedMessage) {
	log.Printf("[BoardClassifier] Sending response to board %s for command %s",
		processed.Original.FromID, processed.Original.Message.Command)
}

func (bc *BoardMessageClassifier) rescheduleTasks(processed *ProcessedMessage) {
	errMsg := fmt.Sprintf("Board %s needs task rescheduling: %s",
		processed.Original.FromID, processed.Details)

	WarningHandler.HandleAgentError(errMsg, processed.Original.FromID, "", "Task rescheduling required")
}

func (bc *BoardMessageClassifier) logToFile(processed *ProcessedMessage) {
	dateStr := time.Now().Format("2006-01-02")
	logFile := filepath.Join(bc.logDir, fmt.Sprintf("board_messages_%s.log", dateStr))

	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("[BoardClassifier] Failed to open log file: %v", err)
		return
	}
	defer file.Close()

	logEntry := fmt.Sprintf("[%s] Category=%s | Action=%s | FromID=%s | Command=%s | Details=%s\n",
		processed.Timestamp.Format("2006-01-02 15:04:05"),
		processed.Category,
		processed.Action,
		processed.Original.FromID,
		processed.Original.Message.Command,
		processed.Details,
	)

	if _, err := file.WriteString(logEntry); err != nil {
		log.Printf("[BoardClassifier] Failed to write log: %v", err)
	}
}

func (bc *BoardMessageClassifier) StartProcessor(msgChan chan *Board.BoardMessage) {
	bc.running = true
	bc.wg.Add(1)
	go func() {
		defer bc.wg.Done()
		for {
			select {
			case msg := <-msgChan:
				bc.ClassifyAndProcess(msg)
			case <-bc.stopChan:
				return
			}
		}
	}()

	bc.wg.Add(1)
	go func() {
		defer bc.wg.Done()
		for {
			select {
			case processed := <-bc.processorChan:
				bc.logToFile(processed)
			case <-bc.stopChan:
				return
			}
		}
	}()

	log.Printf("[BoardClassifier] Message processor started")
}

func (bc *BoardMessageClassifier) StopProcessor() {
	bc.running = false
	close(bc.stopChan)
	bc.wg.Wait()
	log.Printf("[BoardClassifier] Message processor stopped")
}

func (bc *BoardMessageClassifier) GetLogDir() string {
	return bc.logDir
}

func (bc *BoardMessageClassifier) SetLogDir(dir string) {
	bc.logDir = dir
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("[BoardClassifier] Failed to create log directory: %v", err)
	}
}

func ClassifyMessage(msg *Board.BoardMessage) MessageCategory {
	c := NewBoardMessageClassifier()
	return c.classifyMessage(msg)
}

func GetClassifier() *BoardMessageClassifier {
	return classifier
}
