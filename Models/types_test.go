package Models

import (
	"encoding/json"
	"strings"
	"testing"
)

// ==================== LSTMRequest 测试 ====================

func TestLSTMRequestJSON(t *testing.T) {
	req := LSTMRequest{
		SensorID:     "sensor_001",
		DataType:     "temperature",
		TimeSeries:   []TimeSeriesPoint{
			{Timestamp: 1234567890, Value: 25.5},
			{Timestamp: 1234567891, Value: 26.0},
		},
		PredictSteps: 10,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("LSTMRequest 序列化失败: %v", err)
	}

	var parsed LSTMRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("LSTMRequest 反序列化失败: %v", err)
	}

	if parsed.SensorID != "sensor_001" {
		t.Errorf("SensorID 不匹配: 期望 sensor_001, 实际 %s", parsed.SensorID)
	}
	if parsed.DataType != "temperature" {
		t.Errorf("DataType 不匹配: 期望 temperature, 实际 %s", parsed.DataType)
	}
	if parsed.PredictSteps != 10 {
		t.Errorf("PredictSteps 不匹配: 期望 10, 实际 %d", parsed.PredictSteps)
	}
	if len(parsed.TimeSeries) != 2 {
		t.Fatalf("TimeSeries 长度不匹配: 期望 2, 实际 %d", len(parsed.TimeSeries))
	}
	if parsed.TimeSeries[0].Timestamp != 1234567890 {
		t.Errorf("TimeSeries[0].Timestamp 不匹配: 期望 1234567890, 实际 %d", parsed.TimeSeries[0].Timestamp)
	}
	if parsed.TimeSeries[0].Value != 25.5 {
		t.Errorf("TimeSeries[0].Value 不匹配: 期望 25.5, 实际 %f", parsed.TimeSeries[0].Value)
	}
}

func TestLSTMRequestEmptyTimeSeries(t *testing.T) {
	req := LSTMRequest{
		SensorID:     "sensor_002",
		DataType:     "humidity",
		TimeSeries:   []TimeSeriesPoint{},
		PredictSteps: 5,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("LSTMRequest 空时间序列序列化失败: %v", err)
	}

	var parsed LSTMRequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("LSTMRequest 空时间序列反序列化失败: %v", err)
	}

	if len(parsed.TimeSeries) != 0 {
		t.Errorf("空 TimeSeries 应该长度为 0, 实际 %d", len(parsed.TimeSeries))
	}
}

// ==================== TimeSeriesPoint 测试 ====================

func TestTimeSeriesPointJSON(t *testing.T) {
	point := TimeSeriesPoint{
		Timestamp: 1609459200,
		Value:     36.6,
	}

	data, err := json.Marshal(point)
	if err != nil {
		t.Fatalf("TimeSeriesPoint 序列化失败: %v", err)
	}

	var parsed TimeSeriesPoint
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("TimeSeriesPoint 反序列化失败: %v", err)
	}

	if parsed.Timestamp != 1609459200 {
		t.Errorf("Timestamp 不匹配: 期望 1609459200, 实际 %d", parsed.Timestamp)
	}
	if parsed.Value != 36.6 {
		t.Errorf("Value 不匹配: 期望 36.6, 实际 %f", parsed.Value)
	}
}

// ==================== LSTMResponse 测试 ====================

func TestLSTMResponseJSON(t *testing.T) {
	resp := LSTMResponse{
		Prediction:   []float64{1.0, 2.0, 3.0},
		AnomalyScore: 0.85,
		IsAnomaly:    true,
		AnomalyType:  "temperature_anomaly",
		Confidence:   0.92,
		ModelVersion: "v1.0",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("LSTMResponse 序列化失败: %v", err)
	}

	var parsed LSTMResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("LSTMResponse 反序列化失败: %v", err)
	}

	if parsed.AnomalyScore != 0.85 {
		t.Errorf("AnomalyScore 不匹配: 期望 0.85, 实际 %f", parsed.AnomalyScore)
	}
	if !parsed.IsAnomaly {
		t.Error("IsAnomaly 应该为 true")
	}
	if parsed.AnomalyType != "temperature_anomaly" {
		t.Errorf("AnomalyType 不匹配: 期望 temperature_anomaly, 实际 %s", parsed.AnomalyType)
	}
	if len(parsed.Prediction) != 3 {
		t.Errorf("Prediction 长度不匹配: 期望 3, 实际 %d", len(parsed.Prediction))
	}
}

func TestLSTMResponseOmitEmptyAnomalyType(t *testing.T) {
	resp := LSTMResponse{
		Prediction:   []float64{1.0},
		AnomalyScore: 0.1,
		IsAnomaly:    false,
		Confidence:   0.9,
		ModelVersion: "v1.0",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("LSTMResponse 序列化失败: %v", err)
	}

	jsonStr := string(data)
	if strings.Contains(jsonStr, "anomaly_type") {
		t.Error("AnomalyType 为空时应该被 omitempty 忽略")
	}
}

// ==================== YOLORequest 测试 ====================

func TestYOLORequestJSON(t *testing.T) {
	req := YOLORequest{
		ImageBase64: "base64data",
		ImageURL:    "http://example.com/image.jpg",
		Confidence:  0.5,
		Source:      "drone",
		Metadata:    map[string]string{"drone_id": "drone_001"},
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("YOLORequest 序列化失败: %v", err)
	}

	var parsed YOLORequest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("YOLORequest 反序列化失败: %v", err)
	}

	if parsed.ImageBase64 != "base64data" {
		t.Errorf("ImageBase64 不匹配: 期望 base64data, 实际 %s", parsed.ImageBase64)
	}
	if parsed.Confidence != 0.5 {
		t.Errorf("Confidence 不匹配: 期望 0.5, 实际 %f", parsed.Confidence)
	}
	if parsed.Metadata["drone_id"] != "drone_001" {
		t.Errorf("Metadata[drone_id] 不匹配: 期望 drone_001, 实际 %s", parsed.Metadata["drone_id"])
	}
}

func TestYOLORequestOmitEmpty(t *testing.T) {
	req := YOLORequest{
		ImageBase64: "base64data",
		Confidence:  0.5,
		Source:      "drone",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("YOLORequest 序列化失败: %v", err)
	}

	jsonStr := string(data)
	if strings.Contains(jsonStr, "image_url") {
		t.Error("ImageURL 为空时应该被 omitempty 忽略")
	}
	if strings.Contains(jsonStr, "metadata") {
		t.Error("Metadata 为 nil 时应该被 omitempty 忽略")
	}
}

// ==================== YOLOResponse 测试 ====================

func TestYOLOResponseJSON(t *testing.T) {
	resp := YOLOResponse{
		Detections: []Detection{
			{Class: "fire", Confidence: 0.95, BBox: [4]float64{100, 200, 300, 400}, Area: 40000},
		},
		HasAnomaly:    true,
		AnomalyType:   "fire",
		Severity:      "critical",
		ImageAnnotated: "annotated_base64",
		ModelVersion:  "v1.0",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("YOLOResponse 序列化失败: %v", err)
	}

	var parsed YOLOResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("YOLOResponse 反序列化失败: %v", err)
	}

	if !parsed.HasAnomaly {
		t.Error("HasAnomaly 应该为 true")
	}
	if parsed.Severity != "critical" {
		t.Errorf("Severity 不匹配: 期望 critical, 实际 %s", parsed.Severity)
	}
	if len(parsed.Detections) != 1 {
		t.Fatalf("Detections 长度不匹配: 期望 1, 实际 %d", len(parsed.Detections))
	}
	if parsed.Detections[0].Class != "fire" {
		t.Errorf("Detections[0].Class 不匹配: 期望 fire, 实际 %s", parsed.Detections[0].Class)
	}
	if parsed.Detections[0].BBox != [4]float64{100, 200, 300, 400} {
		t.Errorf("Detections[0].BBox 不匹配: 期望 [100,200,300,400], 实际 %v", parsed.Detections[0].BBox)
	}
}

func TestYOLOResponseOmitEmpty(t *testing.T) {
	resp := YOLOResponse{
		Detections:  []Detection{},
		HasAnomaly:  false,
		Severity:    "info",
		ModelVersion: "v1.0",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("YOLOResponse 序列化失败: %v", err)
	}

	jsonStr := string(data)
	if strings.Contains(jsonStr, "anomaly_type") {
		t.Error("AnomalyType 为空时应该被 omitempty 忽略")
	}
	if strings.Contains(jsonStr, "image_annotated") {
		t.Error("ImageAnnotated 为空时应该被 omitempty 忽略")
	}
}

// ==================== Detection 测试 ====================

func TestDetectionJSON(t *testing.T) {
	det := Detection{
		Class:      "smoke",
		Confidence: 0.88,
		BBox:       [4]float64{10.5, 20.5, 30.5, 40.5},
		Area:       1500.0,
	}

	data, err := json.Marshal(det)
	if err != nil {
		t.Fatalf("Detection 序列化失败: %v", err)
	}

	var parsed Detection
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Detection 反序列化失败: %v", err)
	}

	if parsed.Class != "smoke" {
		t.Errorf("Class 不匹配: 期望 smoke, 实际 %s", parsed.Class)
	}
	if parsed.Confidence != 0.88 {
		t.Errorf("Confidence 不匹配: 期望 0.88, 实际 %f", parsed.Confidence)
	}
	if parsed.Area != 1500.0 {
		t.Errorf("Area 不匹配: 期望 1500.0, 实际 %f", parsed.Area)
	}
	if parsed.BBox[0] != 10.5 || parsed.BBox[3] != 40.5 {
		t.Errorf("BBox 不匹配: 期望 [10.5,20.5,30.5,40.5], 实际 %v", parsed.BBox)
	}
}

// ==================== AlertJSON 测试 ====================

func TestAlertJSONSerialization(t *testing.T) {
	alert := AlertJSON{
		AlertID:     "alert_001",
		AlertType:   "anomaly",
		Severity:    SeverityCritical,
		Latitude:    39.9042,
		Longitude:   116.4074,
		AnomalyType: AnomalyFire,
		Source:      SourceSensor,
		SensorID:    "sensor_001",
		DroneID:     "drone_001",
		Timestamp:   1234567890,
		Confidence:  0.95,
		Details:     map[string]interface{}{"key": "value"},
	}

	data, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("AlertJSON 序列化失败: %v", err)
	}

	var parsed AlertJSON
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("AlertJSON 反序列化失败: %v", err)
	}

	if parsed.AlertID != "alert_001" {
		t.Errorf("AlertID 不匹配: 期望 alert_001, 实际 %s", parsed.AlertID)
	}
	if parsed.AlertType != "anomaly" {
		t.Errorf("AlertType 不匹配: 期望 anomaly, 实际 %s", parsed.AlertType)
	}
	if parsed.Severity != SeverityCritical {
		t.Errorf("Severity 不匹配: 期望 %s, 实际 %s", SeverityCritical, parsed.Severity)
	}
	if parsed.Latitude != 39.9042 {
		t.Errorf("Latitude 不匹配: 期望 39.9042, 实际 %f", parsed.Latitude)
	}
	if parsed.Longitude != 116.4074 {
		t.Errorf("Longitude 不匹配: 期望 116.4074, 实际 %f", parsed.Longitude)
	}
	if parsed.AnomalyType != AnomalyFire {
		t.Errorf("AnomalyType 不匹配: 期望 %s, 实际 %s", AnomalyFire, parsed.AnomalyType)
	}
	if parsed.Source != SourceSensor {
		t.Errorf("Source 不匹配: 期望 %s, 实际 %s", SourceSensor, parsed.Source)
	}
	if parsed.SensorID != "sensor_001" {
		t.Errorf("SensorID 不匹配: 期望 sensor_001, 实际 %s", parsed.SensorID)
	}
	if parsed.DroneID != "drone_001" {
		t.Errorf("DroneID 不匹配: 期望 drone_001, 实际 %s", parsed.DroneID)
	}
	if parsed.Timestamp != 1234567890 {
		t.Errorf("Timestamp 不匹配: 期望 1234567890, 实际 %d", parsed.Timestamp)
	}
	if parsed.Confidence != 0.95 {
		t.Errorf("Confidence 不匹配: 期望 0.95, 实际 %f", parsed.Confidence)
	}
}

func TestAlertJSONDeserialization(t *testing.T) {
	jsonStr := `{
		"alert_id": "alert_002",
		"alert_type": "normal",
		"severity": "info",
		"latitude": 0,
		"longitude": 0,
		"anomaly_type": "none",
		"source": "drone",
		"timestamp": 9876543210,
		"confidence": 0.5
	}`

	var alert AlertJSON
	if err := json.Unmarshal([]byte(jsonStr), &alert); err != nil {
		t.Fatalf("AlertJSON 反序列化失败: %v", err)
	}

	if alert.AlertID != "alert_002" {
		t.Errorf("AlertID 不匹配: 期望 alert_002, 实际 %s", alert.AlertID)
	}
	if alert.AlertType != "normal" {
		t.Errorf("AlertType 不匹配: 期望 normal, 实际 %s", alert.AlertType)
	}
	if alert.Severity != SeverityInfo {
		t.Errorf("Severity 不匹配: 期望 %s, 实际 %s", SeverityInfo, alert.Severity)
	}
	if alert.Source != SourceDrone {
		t.Errorf("Source 不匹配: 期望 %s, 实际 %s", SourceDrone, alert.Source)
	}
	if alert.Timestamp != 9876543210 {
		t.Errorf("Timestamp 不匹配: 期望 9876543210, 实际 %d", alert.Timestamp)
	}
}

func TestAlertJSONOmitEmpty(t *testing.T) {
	alert := AlertJSON{
		AlertID:     "alert_003",
		AlertType:   "anomaly",
		Severity:    SeverityHigh,
		Latitude:    39.9,
		Longitude:   116.4,
		AnomalyType: AnomalyGas,
		Source:      SourceSensor,
		Timestamp:   1111111111,
		Confidence:  0.8,
	}

	data, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("AlertJSON 序列化失败: %v", err)
	}

	jsonStr := string(data)

	if strings.Contains(jsonStr, "sensor_id") {
		t.Error("SensorID 为空字符串时应该被 omitempty 忽略")
	}
	if strings.Contains(jsonStr, "drone_id") {
		t.Error("DroneID 为空字符串时应该被 omitempty 忽略")
	}
	if strings.Contains(jsonStr, "details") {
		t.Error("Details 为 nil 时应该被 omitempty 忽略")
	}
}

func TestAlertJSONWithDetails(t *testing.T) {
	alert := AlertJSON{
		AlertID:     "alert_004",
		AlertType:   "anomaly",
		Severity:    SeverityMedium,
		AnomalyType: AnomalyTemp,
		Source:      SourceModel,
		Timestamp:   2222222222,
		Confidence:  0.7,
		Details: map[string]interface{}{
			"anomaly_score": 0.65,
			"prediction":    []interface{}{1.0, 2.0, 3.0},
		},
	}

	data, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("AlertJSON 序列化失败: %v", err)
	}

	var parsed AlertJSON
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("AlertJSON 反序列化失败: %v", err)
	}

	if parsed.Details == nil {
		t.Fatal("Details 不应该为 nil")
	}
	if score, ok := parsed.Details["anomaly_score"].(float64); !ok || score != 0.65 {
		t.Errorf("Details[anomaly_score] 不匹配: 期望 0.65, 实际 %v", parsed.Details["anomaly_score"])
	}
}

// ==================== 常量测试 ====================

func TestSeverityConstants(t *testing.T) {
	if SeverityCritical != "critical" {
		t.Errorf("SeverityCritical 应该为 'critical', 实际 %s", SeverityCritical)
	}
	if SeverityHigh != "high" {
		t.Errorf("SeverityHigh 应该为 'high', 实际 %s", SeverityHigh)
	}
	if SeverityMedium != "medium" {
		t.Errorf("SeverityMedium 应该为 'medium', 实际 %s", SeverityMedium)
	}
	if SeverityLow != "low" {
		t.Errorf("SeverityLow 应该为 'low', 实际 %s", SeverityLow)
	}
	if SeverityInfo != "info" {
		t.Errorf("SeverityInfo 应该为 'info', 实际 %s", SeverityInfo)
	}
}

func TestSourceConstants(t *testing.T) {
	if SourceSensor != "sensor" {
		t.Errorf("SourceSensor 应该为 'sensor', 实际 %s", SourceSensor)
	}
	if SourceDrone != "drone" {
		t.Errorf("SourceDrone 应该为 'drone', 实际 %s", SourceDrone)
	}
	if SourceModel != "model" {
		t.Errorf("SourceModel 应该为 'model', 实际 %s", SourceModel)
	}
}

func TestAnomalyConstants(t *testing.T) {
	anomalyConstants := map[string]string{
		"AnomalyFire":     AnomalyFire,
		"AnomalyGas":      AnomalyGas,
		"AnomalyStruct":   AnomalyStruct,
		"AnomalySmoke":    AnomalySmoke,
		"AnomalyPerson":   AnomalyPerson,
		"AnomalyTemp":     AnomalyTemp,
		"AnomalyHumidity": AnomalyHumidity,
		"AnomalyPressure": AnomalyPressure,
		"AnomalyUnknown":  AnomalyUnknown,
	}

	expectedValues := map[string]string{
		"AnomalyFire":     "fire",
		"AnomalyGas":      "gas_leak",
		"AnomalyStruct":   "structural_damage",
		"AnomalySmoke":    "smoke",
		"AnomalyPerson":   "person_detected",
		"AnomalyTemp":     "temperature_anomaly",
		"AnomalyHumidity": "humidity_anomaly",
		"AnomalyPressure": "pressure_anomaly",
		"AnomalyUnknown":  "unknown",
	}

	for name, actual := range anomalyConstants {
		if expected, ok := expectedValues[name]; !ok {
			t.Errorf("未知的异常常量: %s", name)
		} else if actual != expected {
			t.Errorf("%s 应该为 '%s', 实际 '%s'", name, expected, actual)
		}
	}
}

func TestSeverityConstantsDistinct(t *testing.T) {
	severities := []string{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}
	seen := make(map[string]bool)
	for _, s := range severities {
		if seen[s] {
			t.Errorf("严重性常量有重复值: %s", s)
		}
		seen[s] = true
	}
}

func TestAnomalyConstantsDistinct(t *testing.T) {
	anomalies := []string{AnomalyFire, AnomalyGas, AnomalyStruct, AnomalySmoke, AnomalyPerson, AnomalyTemp, AnomalyHumidity, AnomalyPressure, AnomalyUnknown}
	seen := make(map[string]bool)
	for _, a := range anomalies {
		if seen[a] {
			t.Errorf("异常常量有重复值: %s", a)
		}
		seen[a] = true
	}
}
