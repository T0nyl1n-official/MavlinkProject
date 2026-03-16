package Warning_General

const (
	WarningType_General_Functions = "functionError"
	WarningType_General_Parameters = "parametersError"
	WarningType_General_Log = "logError"
	WarningType_General_Server = "serverError"
	WarningType_General_Connection = "connectionError"
	WarningType_General_Script = "scriptError"
	WarningType_General_Verification = "verificationError"
	WarningType_General_Unknown = "unknownError"
)

type Warning_General struct {
	WarningID int64 `json:"warning_id"`
	WarningType string `json:"warning_type"`
	WarningContent string `json:"warning_content"`
}
