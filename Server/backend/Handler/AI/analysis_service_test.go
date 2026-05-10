package AI

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	Models "MavlinkProject/Models"
)

// ==================== scoreToSeverity 测试 ====================

func TestScoreToSeverity(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected string
	}{
		{"极高分数 0.95", 0.95, Models.SeverityCritical},
		{"临界分数 0.9", 0.9, Models.SeverityCritical},
		{"高分数 0.89", 0.89, Models.SeverityHigh},
		{"高分数 0.75", 0.75, Models.SeverityHigh},
		{"临界分数 0.7", 0.7, Models.SeverityHigh},
		{"中分数 0.5", 0.5, Models.SeverityMedium},
		{"中分数 0.4", 0.4, Models.SeverityMedium},
		{"低分数 0.3", 0.3, Models.SeverityLow},
		{"低分数 0.2", 0.2, Models.SeverityLow},
		{"信息分数 0.1", 0.1, Models.SeverityInfo},
		{"零分数 0.0", 0.0, Models.SeverityInfo},
		{"极低分数 0.01", 0.01, Models.SeverityInfo},
		{"刚好低于 0.2", 0.19, Models.SeverityInfo},
		{"刚好低于 0.4", 0.39, Models.SeverityLow},
		{"刚好低于 0.7", 0.69, Models.SeverityMedium},
		{"刚好低于 0.9", 0.89, Models.SeverityHigh},
		{"满分 1.0", 1.0, Models.SeverityCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := scoreToSeverity(tt.score)
			if result != tt.expected {
				t.Errorf("scoreToSeverity(%f) = %s, 期望 %s", tt.score, result, tt.expected)
			}
		})
	}
}

func TestScoreToSeverityBoundaryValues(t *testing.T) {
	// 精确测试边界值
	boundaryTests := []struct {
		score    float64
		expected string
	}{
		{0.9, Models.SeverityCritical},
		{0.8999999, Models.SeverityHigh},
		{0.7, Models.SeverityHigh},
		{0.6999999, Models.SeverityMedium},
		{0.4, Models.SeverityMedium},
		{0.3999999, Models.SeverityLow},
		{0.2, Models.SeverityLow},
		{0.1999999, Models.SeverityInfo},
	}

	for _, tt := range boundaryTests {
		result := scoreToSeverity(tt.score)
		if result != tt.expected {
			t.Errorf("scoreToSeverity(%f) = %s, 期望 %s", tt.score, result, tt.expected)
		}
	}
}

// ==================== mapLSTMAnomalyType 测试 ====================

func TestMapLSTMAnomalyType(t *testing.T) {
	tests := []struct {
		name       string
		modelType  string
		sensorType string
		expected   string
	}{
		// 模型返回有效类型时，直接使用
		{"模型返回 fire", "fire", "temperature", "fire"},
		{"模型返回 gas_leak", "gas_leak", "humidity", "gas_leak"},
		{"模型返回 smoke", "smoke", "pressure", "smoke"},
		{"模型返回非空非unknown", "custom_anomaly", "temperature", "custom_anomaly"},

		// 模型返回空或 unknown 时，根据传感器类型映射
		{"空模型类型-温度", "", "temperature", Models.AnomalyTemp},
		{"空模型类型-temp", "", "temp", Models.AnomalyTemp},
		{"空模型类型-湿度", "", "humidity", Models.AnomalyHumidity},
		{"空模型类型-气压", "", "pressure", Models.AnomalyPressure},
		{"空模型类型-气体", "", "gas", Models.AnomalyGas},
		{"空模型类型-CO", "", "co", Models.AnomalyGas},
		{"空模型类型-CH4", "", "ch4", Models.AnomalyGas},
		{"空模型类型-CO2", "", "co2", Models.AnomalyGas},
		{"空模型类型-未知传感器", "", "unknown_sensor", Models.AnomalyUnknown},

		// 模型返回 unknown 时，根据传感器类型映射
		{"unknown模型类型-温度", "unknown", "temperature", Models.AnomalyTemp},
		{"unknown模型类型-湿度", "unknown", "humidity", Models.AnomalyHumidity},
		{"unknown模型类型-气压", "unknown", "pressure", Models.AnomalyPressure},
		{"unknown模型类型-气体", "unknown", "gas", Models.AnomalyGas},
		{"unknown模型类型-未知传感器", "unknown", "other", Models.AnomalyUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mapLSTMAnomalyType(tt.modelType, tt.sensorType)
			if result != tt.expected {
				t.Errorf("mapLSTMAnomalyType(%q, %q) = %s, 期望 %s", tt.modelType, tt.sensorType, result, tt.expected)
			}
		})
	}
}

// ==================== generateAlertID 测试 ====================

func TestGenerateAlertID(t *testing.T) {
	id := generateAlertID()

	if id == "" {
		t.Error("generateAlertID 不应该返回空字符串")
	}

	if !strings.HasPrefix(id, "alert_") {
		t.Errorf("Alert ID 应该以 'alert_' 开头, 实际: %s", id)
	}
}

func TestGenerateAlertIDUniqueness(t *testing.T) {
	ids := make(map[string]bool)
	iterations := 1000

	for i := 0; i < iterations; i++ {
		id := generateAlertID()
		if ids[id] {
			t.Errorf("生成重复的 Alert ID: %s (第 %d 次迭代)", id, i)
		}
		ids[id] = true
	}

	if len(ids) != iterations {
		t.Errorf("生成了 %d 个唯一 ID, 期望 %d 个", len(ids), iterations)
	}
}

func TestGenerateAlertIDFormat(t *testing.T) {
	id := generateAlertID()

	// 格式应该是 alert_<hex>_<timestamp>
	parts := strings.Split(id, "_")
	if len(parts) < 3 {
		t.Errorf("Alert ID 格式不正确, 应该至少有 3 个下划线分隔的部分, 实际: %s", id)
	}

	if parts[0] != "alert" {
		t.Errorf("Alert ID 第一部分应该是 'alert', 实际: %s", parts[0])
	}
}

// ==================== AlertToJSON 测试 ====================

func TestAlertToJSON(t *testing.T) {
	alert := &Models.AlertJSON{
		AlertID:     "alert_json_test",
		AlertType:   "anomaly",
		Severity:    Models.SeverityHigh,
		Latitude:    39.9042,
		Longitude:   116.4074,
		AnomalyType: Models.AnomalyFire,
		Source:      Models.SourceSensor,
		SensorID:    "sensor_001",
		Timestamp:   1234567890,
		Confidence:  0.95,
		Details:     map[string]interface{}{"anomaly_score": 0.85},
	}

	jsonStr := AlertToJSON(alert)

	if jsonStr == "" {
		t.Error("AlertToJSON 不应该返回空字符串")
	}

	if jsonStr == "{}" {
		t.Error("AlertToJSON 不应该返回空 JSON 对象")
	}

	// 验证可以解析回 AlertJSON
	var parsed Models.AlertJSON
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Fatalf("AlertToJSON 结果应该是有效的 JSON: %v", err)
	}

	if parsed.AlertID != "alert_json_test" {
		t.Errorf("解析后 AlertID 不匹配: 期望 alert_json_test, 实际 %s", parsed.AlertID)
	}
	if parsed.Severity != Models.SeverityHigh {
		t.Errorf("解析后 Severity 不匹配: 期望 %s, 实际 %s", Models.SeverityHigh, parsed.Severity)
	}
	if parsed.Source != Models.SourceSensor {
		t.Errorf("解析后 Source 不匹配: 期望 %s, 实际 %s", Models.SourceSensor, parsed.Source)
	}
}

func TestAlertToJSONContainsFields(t *testing.T) {
	alert := &Models.AlertJSON{
		AlertID:     "field_check",
		AlertType:   "anomaly",
		Severity:    Models.SeverityCritical,
		AnomalyType: Models.AnomalyGas,
		Source:      Models.SourceDrone,
		DroneID:     "drone_001",
		Timestamp:   9999999999,
		Confidence:  0.77,
	}

	jsonStr := AlertToJSON(alert)

	expectedFields := []string{
		"alert_id", "alert_type", "severity", "anomaly_type",
		"source", "drone_id", "timestamp", "confidence",
	}

	for _, field := range expectedFields {
		if !strings.Contains(jsonStr, field) {
			t.Errorf("JSON 输出应该包含字段 '%s'", field)
		}
	}
}

// ==================== ProcessSensorData 测试（模型禁用时的降级行为） ====================

func TestProcessSensorDataDisabled(t *testing.T) {
	// 确保 AnalysisService 已初始化
	service := GetAnalysisService()

	sensorID := "test_sensor_001"
	sensorType := "temperature"
	values := []Models.TimeSeriesPoint{
		{Timestamp: time.Now().Unix(), Value: 25.5},
		{Timestamp: time.Now().Unix() + 1, Value: 26.0},
	}

	alert, err := service.ProcessSensorData(sensorID, sensorType, values, 39.9, 116.4)
	if err != nil {
		t.Fatalf("模型禁用时 ProcessSensorData 不应该返回错误: %v", err)
	}

	if alert == nil {
		t.Fatal("模型禁用时 ProcessSensorData 不应该返回 nil")
	}

	// 模型禁用时，LSTM 返回 IsAnomaly=false，所以应该是 normal 类型
	if alert.AlertType != "normal" {
		t.Errorf("模型禁用时 AlertType 应该为 'normal', 实际: %s", alert.AlertType)
	}

	if alert.Severity != Models.SeverityInfo {
		t.Errorf("模型禁用时 Severity 应该为 '%s', 实际: %s", Models.SeverityInfo, alert.Severity)
	}

	if alert.Source != Models.SourceSensor {
		t.Errorf("Source 应该为 '%s', 实际: %s", Models.SourceSensor, alert.Source)
	}

	if alert.SensorID != sensorID {
		t.Errorf("SensorID 不匹配: 期望 %s, 实际 %s", sensorID, alert.SensorID)
	}

	if !strings.HasPrefix(alert.AlertID, "alert_") {
		t.Errorf("AlertID 格式不正确: %s", alert.AlertID)
	}
}

func TestProcessSensorDataAnomalyType(t *testing.T) {
	service := GetAnalysisService()

	// 模型禁用时，AnomalyType 应该为 "none"（因为 IsAnomaly=false）
	alert, err := service.ProcessSensorData("sensor_002", "humidity", []Models.TimeSeriesPoint{
		{Timestamp: time.Now().Unix(), Value: 60.0},
	}, 40.0, 117.0)

	if err != nil {
		t.Fatalf("ProcessSensorData 不应该返回错误: %v", err)
	}

	if alert.AnomalyType != "none" {
		t.Errorf("模型禁用时 AnomalyType 应该为 'none', 实际: %s", alert.AnomalyType)
	}
}

// ==================== ProcessDroneImage 测试（模型禁用时的降级行为） ====================

func TestProcessDroneImageDisabled(t *testing.T) {
	service := GetAnalysisService()

	droneID := "test_drone_001"
	imageBase64 := "dGVzdGltYWdl"
	imageURL := ""

	alert, err := service.ProcessDroneImage(droneID, imageBase64, imageURL, 39.9, 116.4)
	if err != nil {
		t.Fatalf("模型禁用时 ProcessDroneImage 不应该返回错误: %v", err)
	}

	if alert == nil {
		t.Fatal("模型禁用时 ProcessDroneImage 不应该返回 nil")
	}

	// 模型禁用时，YOLO 返回 HasAnomaly=false，所以应该是 normal 类型
	if alert.AlertType != "normal" {
		t.Errorf("模型禁用时 AlertType 应该为 'normal', 实际: %s", alert.AlertType)
	}

	if alert.Severity != Models.SeverityInfo {
		t.Errorf("模型禁用时 Severity 应该为 '%s', 实际: %s", Models.SeverityInfo, alert.Severity)
	}

	if alert.Source != Models.SourceDrone {
		t.Errorf("Source 应该为 '%s', 实际: %s", Models.SourceDrone, alert.Source)
	}

	if alert.DroneID != droneID {
		t.Errorf("DroneID 不匹配: 期望 %s, 实际 %s", droneID, alert.DroneID)
	}

	if !strings.HasPrefix(alert.AlertID, "alert_") {
		t.Errorf("AlertID 格式不正确: %s", alert.AlertID)
	}
}

func TestProcessDroneImageAnomalyType(t *testing.T) {
	service := GetAnalysisService()

	alert, err := service.ProcessDroneImage("drone_002", "base64data", "", 40.0, 117.0)
	if err != nil {
		t.Fatalf("ProcessDroneImage 不应该返回错误: %v", err)
	}

	if alert.AnomalyType != "none" {
		t.Errorf("模型禁用时 AnomalyType 应该为 'none', 实际: %s", alert.AnomalyType)
	}
}

func TestProcessDroneImageConfidence(t *testing.T) {
	service := GetAnalysisService()

	alert, err := service.ProcessDroneImage("drone_003", "base64data", "", 39.9, 116.4)
	if err != nil {
		t.Fatalf("ProcessDroneImage 不应该返回错误: %v", err)
	}

	// 模型禁用时，Confidence 应该为 1.0
	if alert.Confidence != 1.0 {
		t.Errorf("模型禁用时 Confidence 应该为 1.0, 实际: %f", alert.Confidence)
	}
}
