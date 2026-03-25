package Warning_Sensor

const (
	WarningType_Sensor_Temperature = "temperatureError"
	WarningType_Sensor_Humidity    = "humidityError"
	WarningType_Sensor_Pressure    = "pressureError"
	WarningType_Sensor_Gas         = "gasError"
	WarningType_Sensor_Motion      = "motionError"
	WarningType_Sensor_Proximity   = "proximityError"
	WarningType_Sensor_Light       = "lightError"
	WarningType_Sensor_Unknown     = "unknownError"
)

type Warning_Sensor struct {
	WarningID      int64  `json:"warning_id"`
	WarningType    string `json:"warning_type"`
	WarningContent string `json:"warning_content"`
	Timestamp      int64  `json:"timestamp"`
	SensorID       string `json:"sensor_id"`
	SensorType     string `json:"sensor_type"`
	Value          string `json:"value"`
	Threshold      string `json:"threshold"`
	Details        string `json:"details"`
}
