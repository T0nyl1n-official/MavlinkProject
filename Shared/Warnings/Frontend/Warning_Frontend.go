package Warning_Frontend

const (
	WarningType_Frontend_Render  = "renderError"
	WarningType_Frontend_State   = "stateError"
	WarningType_Frontend_Component = "componentError"
	WarningType_Frontend_API      = "apiError"
	WarningType_Frontend_Auth     = "authError"
	WarningType_Frontend_Router   = "routerError"
	WarningType_Frontend_Storage  = "storageError"
	WarningType_Frontend_Unknown  = "unknownError"
)

type Warning_Frontend struct {
	WarningID      int64  `json:"warning_id"`
	WarningType    string `json:"warning_type"`
	WarningContent string `json:"warning_content"`
	Timestamp      int64  `json:"timestamp"`
	Source         string `json:"source"`
	Component      string `json:"component"`
	Details        string `json:"details"`
}
