package AI

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
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

// ==================== thermalLevelToSeverity 测试 ====================

func TestThermalLevelToSeverity(t *testing.T) {
	tests := []struct {
		name     string
		level    string
		expected string
	}{
		{"HIGH Lv1 -> critical", Models.TempLevelHigh1, Models.SeverityCritical},
		{"HIGH Lv2 -> high", Models.TempLevelHigh2, Models.SeverityHigh},
		{"NORMAL -> info", Models.TempLevelNormal, Models.SeverityInfo},
		{"LOW Lv2 -> info", Models.TempLevelLow2, Models.SeverityInfo},
		{"LOW Lv1 -> info", Models.TempLevelLow1, Models.SeverityInfo},
		{"未知等级 -> info (默认)", "UNKNOWN_LEVEL", Models.SeverityInfo},
		{"空字符串 -> info (默认)", "", Models.SeverityInfo},
		{"小写等级 -> info (默认)", "high lv2", Models.SeverityInfo},
		{"类似但不同的字符串 -> info (默认)", "HIGH Lv3", Models.SeverityInfo},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := thermalLevelToSeverity(tt.level)
			if result != tt.expected {
				t.Errorf("thermalLevelToSeverity(%q) = %s, 期望 %s", tt.level, result, tt.expected)
			}
		})
	}
}

func TestThermalLevelToSeverityMappingRules(t *testing.T) {
	// 验证温度等级映射规则（根据用户需求文档）
	rules := []struct {
		level       string
		expectedSev string
		isAnomaly   bool
	}{
		{Models.TempLevelHigh1, Models.SeverityCritical, true},
		{Models.TempLevelHigh2, Models.SeverityHigh, true},
		{Models.TempLevelNormal, Models.SeverityInfo, false},
		{Models.TempLevelLow2, Models.SeverityInfo, false},
		{Models.TempLevelLow1, Models.SeverityInfo, false},
	}

	for _, rule := range rules {
		sev := thermalLevelToSeverity(rule.level)
		if sev != rule.expectedSev {
			t.Errorf("规则验证失败: level=%s 期望 severity=%s, 实际 %s",
				rule.level, rule.expectedSev, sev)
		}

		isAnomaly := (sev != Models.SeverityInfo)
		if isAnomaly != rule.isAnomaly {
			t.Errorf("异常判定失败: level=%s 期望 isAnomaly=%v, 实际 %v",
				rule.level, rule.isAnomaly, isAnomaly)
		}
	}
}

// ==================== severityRank 测试 ====================

func TestSeverityRank(t *testing.T) {
	tests := []struct {
		name     string
		severity string
		expected int
	}{
		{"critical rank 5", Models.SeverityCritical, 5},
		{"high rank 4", Models.SeverityHigh, 4},
		{"medium rank 3", Models.SeverityMedium, 3},
		{"low rank 2", Models.SeverityLow, 2},
		{"info rank 1", Models.SeverityInfo, 1},
		{"unknown rank 0", "nonexistent_severity", 0},
		{"空字符串 rank 0", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := severityRank(tt.severity)
			if result != tt.expected {
				t.Errorf("severityRank(%q) = %d, 期望 %d", tt.severity, result, tt.expected)
			}
		})
	}
}

func TestSeverityRankOrdering(t *testing.T) {
	// 验证排序正确性: critical > high > medium > low > info > unknown(0)
	ranks := map[string]int{
		Models.SeverityCritical: severityRank(Models.SeverityCritical),
		Models.SeverityHigh:     severityRank(Models.SeverityHigh),
		Models.SeverityMedium:   severityRank(Models.SeverityMedium),
		Models.SeverityLow:      severityRank(Models.SeverityLow),
		Models.SeverityInfo:     severityRank(Models.SeverityInfo),
	}

	if ranks[Models.SeverityCritical] != 5 {
		t.Errorf("critical rank 应该为 5, 实际 %d", ranks[Models.SeverityCritical])
	}
	if ranks[Models.SeverityHigh] != 4 {
		t.Errorf("high rank 应该为 4, 实际 %d", ranks[Models.SeverityHigh])
	}
	if ranks[Models.SeverityMedium] != 3 {
		t.Errorf("medium rank 应该为 3, 实际 %d", ranks[Models.SeverityMedium])
	}
	if ranks[Models.SeverityLow] != 2 {
		t.Errorf("low rank 应该为 2, 实际 %d", ranks[Models.SeverityLow])
	}
	if ranks[Models.SeverityInfo] != 1 {
		t.Errorf("info rank 应该为 1, 实际 %d", ranks[Models.SeverityInfo])
	}

	// 验证递减关系
	if !(ranks[Models.SeverityCritical] > ranks[Models.SeverityHigh] &&
		ranks[Models.SeverityHigh] > ranks[Models.SeverityMedium] &&
		ranks[Models.SeverityMedium] > ranks[Models.SeverityLow] &&
		ranks[Models.SeverityLow] > ranks[Models.SeverityInfo]) {
		t.Error("severity 排序不满足 critical > high > medium > low > info")
	}
}

// ==================== Mock Server 辅助函数 ====================
// setupMockThermalClient 配置全局 ModelClient 使用 mock YOLOv8 server
// 通过公开 API (SetYOLOURL, SetYOLOParams) 配置，测试结束后恢复原始状态
func setupMockThermalClient(mockServerURL string) func() {
	client := Models.GetModelClient()
	client.SetYOLOURL(mockServerURL)
	client.SetYOLOParams(0.10, 0.60, 1024)

	return func() {
		client.SetYOLOURL("")
		client.SetYOLOParams(0.10, 0.60, 1024)
	}
}

// ==================== ProcessDronePhoto 测试 ====================

func TestProcessDronePhotoDisabled(t *testing.T) {
	service := GetAnalysisService()

	alert, rawResult, err := service.ProcessDronePhoto("drone_001", "/fake/path.jpg", 39.9, 116.4)
	if err != nil {
		t.Fatalf("模型禁用时 ProcessDronePhoto 不应该返回错误: %v", err)
	}
	if alert == nil {
		t.Fatal("alert 不应该为 nil")
	}
	if rawResult == nil {
		t.Fatal("rawResult 不应该为 nil")
	}

	// 模型禁用时应该是 normal 状态
	if alert.AlertType != "normal" {
		t.Errorf("AlertType 应该为 'normal', 实际: %s", alert.AlertType)
	}
	if alert.Severity != Models.SeverityInfo {
		t.Errorf("Severity 应该为 'info', 实际: %s", alert.Severity)
	}
	if alert.AnomalyType != "none" {
		t.Errorf("AnomalyType 应该为 'none', 实际: %s", alert.AnomalyType)
	}
	if alert.Source != Models.SourceDrone {
		t.Errorf("Source 应该为 '%s', 实际: %s", Models.SourceDrone, alert.Source)
	}
	if alert.DroneID != "drone_001" {
		t.Errorf("DroneID 不匹配: 期望 drone_001, 实际: %s", alert.DroneID)
	}
	if alert.Confidence != 1.0 {
		t.Errorf("Confidence 应该为 1.0, 实际: %f", alert.Confidence)
	}
}

func TestProcessDronePhotoWithMockServer_HighLv1Anomaly(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Models.ThermalDetectResponse{
			Success: true,
			Image:   Models.ThermalImageInfo{Width: 640, Height: 512},
			Detections: []Models.ThermalDetection{
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{100, 200, 300, 400}, Xywh: [4]float64{150, 250, 200, 200}},
					Confidence: 0.95,
					Temperature: Models.ThermalInfo{
						MeanGray: 220.0,
						Level:     Models.TempLevelHigh1,
					},
				},
			},
			ElapsedMs: 95.5,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()
	cleanup := setupMockThermalClient(mockServer.URL)
	defer cleanup()

	tmpFile, _ := os.CreateTemp("", "thermal_h1_*.jpg")
	tmpFile.Write([]byte("high lv1 test"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	service := GetAnalysisService()

	alert, rawResult, err := service.ProcessDronePhoto("drone_h1", tmpFile.Name(), 31.2304, 121.4737)
	if err != nil {
		t.Fatalf("ProcessDronePhoto 失败: %v", err)
	}
	if alert == nil || rawResult == nil {
		t.Fatal("alert 和 rawResult 不应该为 nil")
	}

	if alert.AlertType != "anomaly" {
		t.Errorf("有异常时 AlertType 应该为 'anomaly', 实际: %s", alert.AlertType)
	}
	if alert.Severity != Models.SeverityCritical {
		t.Errorf("HIGH Lv1 对应 Severity 应该为 'critical', 实际: %s", alert.Severity)
	}
	if alert.AnomalyType != Models.AnomalyThermal {
		t.Errorf("AnomalyType 应该为 '%s', 实际: %s", Models.AnomalyThermal, alert.AnomalyType)
	}
	if alert.Confidence != 0.95 {
		t.Errorf("Confidence 应该为检测框的 confidence 0.95, 实际: %f", alert.Confidence)
	}
	if alert.Latitude != 31.2304 {
		t.Errorf("Latitude 不匹配: 期望 31.2304, 实际: %f", alert.Latitude)
	}
	if alert.Longitude != 121.4737 {
		t.Errorf("Longitude 不匹配: 期望 121.4737, 实际: %f", alert.Longitude)
	}
	if !rawResult.Success {
		t.Error("rawResult.Success 应该为 true")
	}
	if len(rawResult.Detections) != 1 {
		t.Errorf("rawResult.Detections 数量不匹配: 期望 1, 实际 %d", len(rawResult.Detections))
	}
}

func TestProcessDronePhotoWithMockServer_HighLv2Anomaly(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Models.ThermalDetectResponse{
			Success: true,
			Image:   Models.ThermalImageInfo{Width: 1280, Height: 720},
			Detections: []Models.ThermalDetection{
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{50, 100, 200, 300}, Xywh: [4]float64{125, 175, 150, 200}},
					Confidence: 0.78,
					Temperature: Models.ThermalInfo{
						MeanGray: 150.0,
						Level:     Models.TempLevelHigh2,
					},
				},
			},
			ElapsedMs: 72.3,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()
	cleanup := setupMockThermalClient(mockServer.URL)
	defer cleanup()

	tmpFile, _ := os.CreateTemp("", "thermal_h2_*.jpg")
	tmpFile.Write([]byte("high lv2 test"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	service := GetAnalysisService()

	alert, _, err := service.ProcessDronePhoto("drone_h2", tmpFile.Name(), 30.0, 110.0)
	if err != nil {
		t.Fatalf("ProcessDronePhoto 失败: %v", err)
	}

	if alert.Severity != Models.SeverityHigh {
		t.Errorf("HIGH Lv2 对应 Severity 应该为 'high', 实际: %s", alert.Severity)
	}
	if alert.Confidence != 0.78 {
		t.Errorf("Confidence 不匹配: 期望 0.78, 实际: %f", alert.Confidence)
	}
}

func TestProcessDronePhotoWithMockServer_NoAnomaly(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Models.ThermalDetectResponse{
			Success: true,
			Image:   Models.ThermalImageInfo{Width: 800, Height: 600},
			Detections: []Models.ThermalDetection{
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{10, 20, 50, 80}, Xywh: [4]float64{30, 50, 40, 60}},
					Confidence: 0.45,
					Temperature: Models.ThermalInfo{
						MeanGray: 100.0,
						Level:     Models.TempLevelNormal,
					},
				},
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{200, 300, 250, 350}, Xywh: [4]float64{225, 325, 50, 50}},
					Confidence: 0.22,
					Temperature: Models.ThermalInfo{
						MeanGray: 55.0,
						Level:     Models.TempLevelLow1,
					},
				},
			},
			ElapsedMs: 55.0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()
	cleanup := setupMockThermalClient(mockServer.URL)
	defer cleanup()

	tmpFile, _ := os.CreateTemp("", "thermal_normal_*.jpg")
	tmpFile.Write([]byte("normal test"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	service := GetAnalysisService()

	alert, _, err := service.ProcessDronePhoto("drone_normal", tmpFile.Name(), 35.0, 115.0)
	if err != nil {
		t.Fatalf("ProcessDronePhoto 失败: %v", err)
	}

	if alert.AlertType != "normal" {
		t.Errorf("无异常时 AlertType 应该为 'normal', 实际: %s", alert.AlertType)
	}
	if alert.Severity != Models.SeverityInfo {
		t.Errorf("无异常时 Severity 应该为 'info', 实际: %s", alert.Severity)
	}
	if alert.AnomalyType != "none" {
		t.Errorf("无异常时 AnomalyType 应该为 'none', 实际: %s", alert.AnomalyType)
	}
}

func TestProcessDronePhotoWithMockServer_EmptyDetections(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Models.ThermalDetectResponse{
			Success:    true,
			Image:      Models.ThermalImageInfo{Width: 640, Height: 480},
			Detections: []Models.ThermalDetection{},
			ElapsedMs:  33.3,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()
	cleanup := setupMockThermalClient(mockServer.URL)
	defer cleanup()

	tmpFile, _ := os.CreateTemp("", "thermal_empty_*.jpg")
	tmpFile.Write([]byte("empty det test"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	service := GetAnalysisService()

	alert, rawResult, err := service.ProcessDronePhoto("drone_empty", tmpFile.Name(), 40.0, 116.0)
	if err != nil {
		t.Fatalf("ProcessDronePhoto 失败: %v", err)
	}

	if alert.AlertType != "normal" {
		t.Errorf("空 Detections 时 AlertType 应该为 'normal', 实际: %s", alert.AlertType)
	}
	if len(rawResult.Detections) != 0 {
		t.Errorf("rawResult.Detections 应该为空, 实际长度: %d", len(rawResult.Detections))
	}
}

func TestProcessDronePhotoWithMockServer_MultiDetectionsPickHighest(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Models.ThermalDetectResponse{
			Success: true,
			Image:   Models.ThermalImageInfo{Width: 1920, Height: 1080},
			Detections: []Models.ThermalDetection{
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{100, 100, 200, 200}, Xywh: [4]float64{150, 150, 100, 100}},
					Confidence: 0.65,
					Temperature: Models.ThermalInfo{MeanGray: 150.0, Level: Models.TempLevelHigh2},
				},
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{300, 300, 500, 500}, Xywh: [4]float64{400, 400, 200, 200}},
					Confidence: 0.92,
					Temperature: Models.ThermalInfo{MeanGray: 210.0, Level: Models.TempLevelHigh1},
				},
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{600, 50, 700, 150}, Xywh: [4]float64{650, 100, 100, 100}},
					Confidence: 0.35,
					Temperature: Models.ThermalInfo{MeanGray: 105.0, Level: Models.TempLevelNormal},
				},
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{10, 10, 50, 50}, Xywh: [4]float64{30, 30, 40, 40}},
					Confidence: 0.15,
					Temperature: Models.ThermalInfo{MeanGray: 45.0, Level: Models.TempLevelLow1},
				},
			},
			ElapsedMs: 150.8,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()
	cleanup := setupMockThermalClient(mockServer.URL)
	defer cleanup()

	tmpFile, _ := os.CreateTemp("", "thermal_multi_*.jpg")
	tmpFile.Write([]byte("multi pick highest"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	service := GetAnalysisService()

	alert, rawResult, err := service.ProcessDronePhoto("drone_multi", tmpFile.Name(), 22.5, 114.0)
	if err != nil {
		t.Fatalf("ProcessDronePhoto 失败: %v", err)
	}

	if alert.Severity != Models.SeverityCritical {
		t.Errorf("多检测框应取最高 severity=critical, 实际: %s", alert.Severity)
	}
	if alert.Confidence != 0.92 {
		t.Errorf("Confidence 应该取最高 severity 的 confidence=0.92, 实际: %f", alert.Confidence)
	}
	if alert.AnomalyType != Models.AnomalyThermal {
		t.Errorf("AnomalyType 应该为 thermal_anomaly, 实际: %s", alert.AnomalyType)
	}
	if len(rawResult.Detections) != 4 {
		t.Errorf("rawResult 应保留全部 4 个检测框, 实际: %d", len(rawResult.Detections))
	}
}

func TestProcessDronePhotoWithMockServer_SameSeverityPickHigherConfidence(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Models.ThermalDetectResponse{
			Success: true,
			Image:   Models.ThermalImageInfo{Width: 1024, Height: 768},
			Detections: []Models.ThermalDetection{
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{50, 50, 150, 250}, Xywh: [4]float64{100, 150, 100, 200}},
					Confidence: 0.55,
					Temperature: Models.ThermalInfo{MeanGray: 140.0, Level: Models.TempLevelHigh2},
				},
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{300, 100, 450, 300}, Xywh: [4]float64{375, 200, 150, 200}},
					Confidence: 0.82,
					Temperature: Models.ThermalInfo{MeanGray: 180.0, Level: Models.TempLevelHigh2},
				},
			},
			ElapsedMs: 88.8,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()
	cleanup := setupMockThermalClient(mockServer.URL)
	defer cleanup()

	tmpFile, _ := os.CreateTemp("", "thermal_same_sev_*.jpg")
	tmpFile.Write([]byte("same severity test"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	service := GetAnalysisService()

	alert, _, err := service.ProcessDronePhoto("drone_same_sev", tmpFile.Name(), 23.0, 113.0)
	if err != nil {
		t.Fatalf("ProcessDronePhoto 失败: %v", err)
	}

	if alert.Severity != Models.SeverityHigh {
		t.Errorf("Severity 应该为 'high', 实际: %s", alert.Severity)
	}
	if alert.Confidence != 0.82 {
		t.Errorf("同 severity 应选更高 confidence=0.82, 实际: %f", alert.Confidence)
	}
}

func TestProcessDronePhotoWithMockServer_AlertDetails(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Models.ThermalDetectResponse{
			Success: true,
			Image:   Models.ThermalImageInfo{Width: 640, Height: 512},
			Detections: []Models.ThermalDetection{
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{164.33, 259.13, 207.07, 279.85}, Xywh: [4]float64{185.7, 269.49, 42.74, 20.72}},
					Confidence: 0.182988,
					Temperature: Models.ThermalInfo{MeanGray: 158.8, Level: Models.TempLevelHigh2},
				},
			},
			ElapsedMs: 86.35,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()
	cleanup := setupMockThermalClient(mockServer.URL)
	defer cleanup()

	tmpFile, _ := os.CreateTemp("", "thermal_details_*.jpg")
	tmpFile.Write([]byte("details test"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	service := GetAnalysisService()

	alert, _, err := service.ProcessDronePhoto("drone_details", tmpFile.Name(), 39.9042, 116.4074)
	if err != nil {
		t.Fatalf("ProcessDronePhoto 失败: %v", err)
	}

	if alert.Details == nil {
		t.Fatal("Details 不应该为 nil")
	}
	if alert.Details["thermal_detections"] == nil {
		t.Error("Details 应该包含 thermal_detections")
	}
	if alert.Details["top_detection"] == nil {
		t.Error("Details 应该包含 top_detection")
	}
	if alert.Details["elapsed_ms"] == nil {
		t.Error("Details 应该包含 elapsed_ms")
	}
	if alert.Details["image_size"] == nil {
		t.Error("Details 应该包含 image_size")
	}
}

func TestProcessDronePhotoModelAPIError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("bad gateway error"))
	}))
	defer mockServer.Close()
	cleanup := setupMockThermalClient(mockServer.URL)
	defer cleanup()

	tmpFile, _ := os.CreateTemp("", "thermal_api_err_*.jpg")
	tmpFile.Write([]byte("api error test"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	service := GetAnalysisService()

	_, _, err := service.ProcessDronePhoto("drone_err", tmpFile.Name(), 0, 0)
	if err == nil {
		t.Error("YOLO API 错误时 ProcessDronePhoto 应该返回错误")
	}
}
