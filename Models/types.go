package Models

type LSTMRequest struct {
	SensorID  string             `json:"sensor_id"`
	DataType  string             `json:"data_type"`
	TimeSeries []TimeSeriesPoint `json:"time_series"`
	PredictSteps int             `json:"predict_steps"`
}

type TimeSeriesPoint struct {
	Timestamp int64   `json:"timestamp"`
	Value     float64 `json:"value"`
}

type LSTMResponse struct {
	Prediction    []float64 `json:"prediction"`
	AnomalyScore  float64   `json:"anomaly_score"`
	IsAnomaly     bool      `json:"is_anomaly"`
	AnomalyType   string    `json:"anomaly_type,omitempty"`
	Confidence    float64   `json:"confidence"`
	ModelVersion  string    `json:"model_version"`
}

type YOLORequest struct {
	ImageBase64 string            `json:"image_base64"`
	ImageURL    string            `json:"image_url,omitempty"`
	Confidence  float64           `json:"confidence"`
	Source      string            `json:"source"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type YOLOResponse struct {
	Detections   []Detection `json:"detections"`
	HasAnomaly   bool        `json:"has_anomaly"`
	AnomalyType  string      `json:"anomaly_type,omitempty"`
	Severity     string      `json:"severity"`
	ImageAnnotated string    `json:"image_annotated,omitempty"`
	ModelVersion string      `json:"model_version"`
}

type Detection struct {
	Class      string  `json:"class"`
	Confidence float64 `json:"confidence"`
	BBox       [4]float64 `json:"bbox"`
	Area       float64 `json:"area"`
}

type AlertJSON struct {
	AlertID      string                 `json:"alert_id"`
	AlertType    string                 `json:"alert_type"`
	Severity     string                 `json:"severity"`
	Latitude     float64                `json:"latitude"`
	Longitude    float64                `json:"longitude"`
	AnomalyType  string                 `json:"anomaly_type"`
	Source       string                 `json:"source"`
	SensorID     string                 `json:"sensor_id,omitempty"`
	DroneID      string                 `json:"drone_id,omitempty"`
	Timestamp    int64                  `json:"timestamp"`
	Confidence   float64                `json:"confidence"`
	Details      map[string]interface{} `json:"details,omitempty"`
}

const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"

	SourceSensor = "sensor"
	SourceDrone  = "drone"
	SourceModel  = "model"

	AnomalyFire      = "fire"
	AnomalyGas       = "gas_leak"
	AnomalyStruct    = "structural_damage"
	AnomalySmoke     = "smoke"
	AnomalyPerson    = "person_detected"
	AnomalyTemp      = "temperature_anomaly"
	AnomalyHumidity  = "humidity_anomaly"
	AnomalyPressure  = "pressure_anomaly"
	AnomalyUnknown   = "unknown"
)
