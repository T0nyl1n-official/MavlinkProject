package Warning_Agent

const (
	WarningType_Agent_Execution = "executionError"
	WarningType_Agent_Planning  = "planningError"
	WarningType_Agent_Timeout   = "timeoutError"
	WarningType_Agent_Resource  = "resourceError"
	WarningType_Agent_API       = "apiError"
	WarningType_Agent_Logic     = "logicError"
	WarningType_Agent_Memory    = "memoryError"
	WarningType_Agent_Unknown   = "unknownError"
)

type Warning_Agent struct {
	WarningID      int64  `json:"warning_id"`
	WarningType    string `json:"warning_type"`
	WarningContent string `json:"warning_content"`
	Timestamp      int64  `json:"timestamp"`
	AgentID        string `json:"agent_id"`
	TaskID         string `json:"task_id"`
	Details        string `json:"details"`
}
