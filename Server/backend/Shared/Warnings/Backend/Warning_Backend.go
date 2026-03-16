package Warning_Backend

const (
	WarningType_Backend_Database   = "databaseError"
	WarningType_Backend_API        = "apiError"
	WarningType_Backend_Auth       = "authError"
	WarningType_Backend_Route      = "routeError"
	WarningType_Backend_Middleware = "middlewareError"
	WarningType_Backend_Config     = "configError"
	WarningType_Backend_Timeout    = "timeoutError"
	WarningType_Backend_Unknown    = "unknownError"
)

type Warning_Backend struct {
	WarningID      int64  `json:"warning_id"`
	WarningType    string `json:"warning_type"`
	WarningContent string `json:"warning_content"`
	Timestamp      int64  `json:"timestamp"`
	Source         string `json:"source"`
	Details        string `json:"details"`
}
