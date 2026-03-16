package Warning_Drone

const (
	WarningType_Drone_Connection   = "connectionError"
	WarningType_Drone_Communication = "communicationError"
	WarningType_Drone_Battery      = "batteryError"
	WarningType_Drone_GPS         = "gpsError"
	WarningType_Drone_Motor       = "motorError"
	WarningType_Drone_Sensor      = "sensorError"
	WarningType_Drone_Command     = "commandError"
	WarningType_Drone_Flight      = "flightError"
	WarningType_Drone_Unknown     = "unknownError"
)

type Warning_Drone struct {
	WarningID      int64  `json:"warning_id"`
	WarningType    string `json:"warning_type"`
	WarningContent string `json:"warning_content"`
	Timestamp      int64  `json:"timestamp"`
	DroneID        string `json:"drone_id"`
	HandlerID      string `json:"handler_id"`
	Details        string `json:"details"`
}
