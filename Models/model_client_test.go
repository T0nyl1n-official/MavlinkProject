package Models

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// ==================== ModelClient 初始化测试 ====================

func TestGetModelClient(t *testing.T) {
	client := GetModelClient()
	if client == nil {
		t.Fatal("GetModelClient 不应该返回 nil")
	}

	// 再次调用应该返回同一个实例
	client2 := GetModelClient()
	if client != client2 {
		t.Error("GetModelClient 应该返回同一个单例实例")
	}
}

func TestInitModelClient(t *testing.T) {
	// InitModelClient 使用 sync.Once，所以只能初始化一次
	// 这里验证 GetModelClient 在未初始化时自动调用 InitModelClient("", "")
	// 由于 sync.Once 的特性，如果已经初始化过，再次调用不会重新初始化
	client := GetModelClient()
	if client == nil {
		t.Fatal("InitModelClient 后 GetModelClient 不应该返回 nil")
	}
}

// ==================== IsLSTMEnabled / IsYOLOEnabled 测试 ====================

func TestIsLSTMEnabledDefault(t *testing.T) {
	// 使用新实例测试，避免影响全局单例
	client := &ModelClient{
		lstmBaseURL: "",
		yoloBaseURL: "",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		lstmEnabled: false,
		yoloEnabled: false,
	}

	if client.IsLSTMEnabled() {
		t.Error("新创建的 ModelClient LSTM 不应该启用")
	}
	if client.IsYOLOEnabled() {
		t.Error("新创建的 ModelClient YOLO 不应该启用")
	}
}

func TestIsEnabledWithURL(t *testing.T) {
	client := &ModelClient{
		lstmBaseURL: "http://localhost:5000",
		yoloBaseURL: "http://localhost:5001",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		lstmEnabled: true,
		yoloEnabled: true,
	}

	if !client.IsLSTMEnabled() {
		t.Error("设置了 LSTM URL 后应该启用")
	}
	if !client.IsYOLOEnabled() {
		t.Error("设置了 YOLO URL 后应该启用")
	}
}

// ==================== SetLSTMURL / SetYOLOURL 测试 ====================

func TestSetLSTMURL(t *testing.T) {
	client := &ModelClient{
		lstmBaseURL: "",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		lstmEnabled: false,
	}

	// 设置非空 URL 应该启用
	client.SetLSTMURL("http://localhost:5000")
	if !client.IsLSTMEnabled() {
		t.Error("SetLSTMURL 设置非空 URL 后应该启用")
	}

	// 设置空 URL 应该禁用
	client.SetLSTMURL("")
	if client.IsLSTMEnabled() {
		t.Error("SetLSTMURL 设置空 URL 后应该禁用")
	}
}

func TestSetYOLOURL(t *testing.T) {
	client := &ModelClient{
		yoloBaseURL: "",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		yoloEnabled: false,
	}

	// 设置非空 URL 应该启用
	client.SetYOLOURL("http://localhost:5001")
	if !client.IsYOLOEnabled() {
		t.Error("SetYOLOURL 设置非空 URL 后应该启用")
	}

	// 设置空 URL 应该禁用
	client.SetYOLOURL("")
	if client.IsYOLOEnabled() {
		t.Error("SetYOLOURL 设置空 URL 后应该禁用")
	}
}

// ==================== AnalyzeSensorData 测试 ====================

func TestAnalyzeSensorDataDisabled(t *testing.T) {
	client := &ModelClient{
		lstmBaseURL: "",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		lstmEnabled: false,
	}

	req := LSTMRequest{
		SensorID:     "sensor_001",
		DataType:     "temperature",
		TimeSeries:   []TimeSeriesPoint{{Timestamp: 1234567890, Value: 25.5}},
		PredictSteps: 10,
	}

	resp, err := client.AnalyzeSensorData(req)
	if err != nil {
		t.Fatalf("LSTM 禁用时 AnalyzeSensorData 不应该返回错误: %v", err)
	}
	if resp == nil {
		t.Fatal("LSTM 禁用时 AnalyzeSensorData 不应该返回 nil")
	}
	if resp.IsAnomaly {
		t.Error("LSTM 禁用时 IsAnomaly 应该为 false")
	}
	if resp.AnomalyType != AnomalyUnknown {
		t.Errorf("LSTM 禁用时 AnomalyType 应该为 '%s', 实际 '%s'", AnomalyUnknown, resp.AnomalyType)
	}
	if resp.Confidence != 0 {
		t.Errorf("LSTM 禁用时 Confidence 应该为 0, 实际 %f", resp.Confidence)
	}
}

func TestAnalyzeSensorDataWithMockServer(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/predict" {
			resp := LSTMResponse{
				Prediction:   []float64{1.0, 2.0, 3.0},
				AnomalyScore: 0.85,
				IsAnomaly:    true,
				AnomalyType:  "temperature_anomaly",
				Confidence:   0.92,
				ModelVersion: "v1.0",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	client := &ModelClient{
		lstmBaseURL: mockServer.URL,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		maxRetry:    1,
		retryDelay:  10 * time.Millisecond,
		lstmEnabled: true,
	}

	req := LSTMRequest{
		SensorID:     "sensor_001",
		DataType:     "temperature",
		TimeSeries:   []TimeSeriesPoint{{Timestamp: 1234567890, Value: 25.5}},
		PredictSteps: 10,
	}

	resp, err := client.AnalyzeSensorData(req)
	if err != nil {
		t.Fatalf("AnalyzeSensorData 不应该返回错误: %v", err)
	}
	if !resp.IsAnomaly {
		t.Error("Mock 服务器返回 IsAnomaly=true, 应该为 true")
	}
	if resp.AnomalyScore != 0.85 {
		t.Errorf("AnomalyScore 不匹配: 期望 0.85, 实际 %f", resp.AnomalyScore)
	}
	if resp.AnomalyType != "temperature_anomaly" {
		t.Errorf("AnomalyType 不匹配: 期望 temperature_anomaly, 实际 %s", resp.AnomalyType)
	}
	if resp.Confidence != 0.92 {
		t.Errorf("Confidence 不匹配: 期望 0.92, 实际 %f", resp.Confidence)
	}
	if len(resp.Prediction) != 3 {
		t.Errorf("Prediction 长度不匹配: 期望 3, 实际 %d", len(resp.Prediction))
	}
}

func TestAnalyzeSensorDataServerUnavailable(t *testing.T) {
	client := &ModelClient{
		lstmBaseURL: "http://localhost:1", // 不可达的端口
		httpClient:  &http.Client{Timeout: 500 * time.Millisecond},
		maxRetry:    1,
		retryDelay:  10 * time.Millisecond,
		lstmEnabled: true,
	}

	req := LSTMRequest{
		SensorID:     "sensor_001",
		DataType:     "temperature",
		TimeSeries:   []TimeSeriesPoint{{Timestamp: 1234567890, Value: 25.5}},
		PredictSteps: 10,
	}

	_, err := client.AnalyzeSensorData(req)
	if err == nil {
		t.Error("服务器不可用时 AnalyzeSensorData 应该返回错误")
	}
}

// ==================== AnalyzeImage 测试 ====================

func TestAnalyzeImageDisabled(t *testing.T) {
	client := &ModelClient{
		yoloBaseURL: "",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		yoloEnabled: false,
	}

	req := YOLORequest{
		ImageBase64: "base64data",
		Confidence:  0.5,
		Source:      SourceDrone,
	}

	resp, err := client.AnalyzeImage(req)
	if err != nil {
		t.Fatalf("YOLO 禁用时 AnalyzeImage 不应该返回错误: %v", err)
	}
	if resp == nil {
		t.Fatal("YOLO 禁用时 AnalyzeImage 不应该返回 nil")
	}
	if resp.HasAnomaly {
		t.Error("YOLO 禁用时 HasAnomaly 应该为 false")
	}
	if resp.AnomalyType != AnomalyUnknown {
		t.Errorf("YOLO 禁用时 AnomalyType 应该为 '%s', 实际 '%s'", AnomalyUnknown, resp.AnomalyType)
	}
	if resp.Severity != SeverityInfo {
		t.Errorf("YOLO 禁用时 Severity 应该为 '%s', 实际 '%s'", SeverityInfo, resp.Severity)
	}
}

func TestAnalyzeImageWithMockServer(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/detect" {
			resp := YOLOResponse{
				Detections: []Detection{
					{Class: "fire", Confidence: 0.95, BBox: [4]float64{100, 200, 300, 400}, Area: 40000},
				},
				HasAnomaly:   true,
				AnomalyType:  "fire",
				Severity:     "critical",
				ModelVersion: "v1.0",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	client := &ModelClient{
		yoloBaseURL: mockServer.URL,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		maxRetry:    1,
		retryDelay:  10 * time.Millisecond,
		yoloEnabled: true,
	}

	req := YOLORequest{
		ImageBase64: "base64data",
		Confidence:  0.5,
		Source:      SourceDrone,
	}

	resp, err := client.AnalyzeImage(req)
	if err != nil {
		t.Fatalf("AnalyzeImage 不应该返回错误: %v", err)
	}
	if !resp.HasAnomaly {
		t.Error("Mock 服务器返回 HasAnomaly=true, 应该为 true")
	}
	if resp.Severity != "critical" {
		t.Errorf("Severity 不匹配: 期望 critical, 实际 %s", resp.Severity)
	}
	if len(resp.Detections) != 1 {
		t.Fatalf("Detections 长度不匹配: 期望 1, 实际 %d", len(resp.Detections))
	}
	if resp.Detections[0].Class != "fire" {
		t.Errorf("Detections[0].Class 不匹配: 期望 fire, 实际 %s", resp.Detections[0].Class)
	}
}

// ==================== AnalyzeImageFile 测试 ====================

func TestAnalyzeImageFileDisabled(t *testing.T) {
	client := &ModelClient{
		yoloBaseURL: "",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		yoloEnabled: false,
	}

	resp, err := client.AnalyzeImageFile("/nonexistent/file.jpg", "drone", nil)
	if err != nil {
		t.Fatalf("YOLO 禁用时 AnalyzeImageFile 不应该返回错误: %v", err)
	}
	if resp == nil {
		t.Fatal("YOLO 禁用时 AnalyzeImageFile 不应该返回 nil")
	}
	if resp.HasAnomaly {
		t.Error("YOLO 禁用时 HasAnomaly 应该为 false")
	}
	if resp.Severity != SeverityInfo {
		t.Errorf("YOLO 禁用时 Severity 应该为 '%s', 实际 '%s'", SeverityInfo, resp.Severity)
	}
}

func TestAnalyzeImageFileWithMockServer(t *testing.T) {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "test_image_*.jpg")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFile.WriteString("fake image data for testing")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/detect/file" {
			// 验证是否为 multipart form 请求
			contentType := r.Header.Get("Content-Type")
			if len(contentType) == 0 || len(contentType) < 10 {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			resp := YOLOResponse{
				Detections: []Detection{
					{Class: "smoke", Confidence: 0.88, BBox: [4]float64{50, 100, 200, 300}, Area: 30000},
				},
				HasAnomaly:   true,
				AnomalyType:  "smoke",
				Severity:     "high",
				ModelVersion: "v1.0",
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	client := &ModelClient{
		yoloBaseURL: mockServer.URL,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		maxRetry:    1,
		retryDelay:  10 * time.Millisecond,
		yoloEnabled: true,
	}

	resp, err := client.AnalyzeImageFile(tmpFile.Name(), "drone", map[string]string{"drone_id": "drone_001"})
	if err != nil {
		t.Fatalf("AnalyzeImageFile 不应该返回错误: %v", err)
	}
	if !resp.HasAnomaly {
		t.Error("Mock 服务器返回 HasAnomaly=true, 应该为 true")
	}
	if resp.Severity != "high" {
		t.Errorf("Severity 不匹配: 期望 high, 实际 %s", resp.Severity)
	}
	if len(resp.Detections) != 1 {
		t.Fatalf("Detections 长度不匹配: 期望 1, 实际 %d", len(resp.Detections))
	}
	if resp.Detections[0].Class != "smoke" {
		t.Errorf("Detections[0].Class 不匹配: 期望 smoke, 实际 %s", resp.Detections[0].Class)
	}
}

func TestAnalyzeImageFileNotFound(t *testing.T) {
	client := &ModelClient{
		yoloBaseURL: "http://localhost:5001",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		maxRetry:    1,
		retryDelay:  10 * time.Millisecond,
		yoloEnabled: true,
	}

	_, err := client.AnalyzeImageFile("/nonexistent/path/image.jpg", "drone", nil)
	if err == nil {
		t.Error("文件不存在时 AnalyzeImageFile 应该返回错误")
	}
}

// ==================== HealthCheck 测试 ====================

func TestHealthCheckDisabled(t *testing.T) {
	client := &ModelClient{
		lstmBaseURL: "",
		yoloBaseURL: "",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		lstmEnabled: false,
		yoloEnabled: false,
	}

	status := client.HealthCheck()
	if status["lstm_enabled"] != false {
		t.Error("LSTM 禁用时 lstm_enabled 应该为 false")
	}
	if status["yolo_enabled"] != false {
		t.Error("YOLO 禁用时 yolo_enabled 应该为 false")
	}
	if status["lstm_status"] != "disabled" {
		t.Errorf("LSTM 禁用时 lstm_status 应该为 'disabled', 实际 %v", status["lstm_status"])
	}
	if status["yolo_status"] != "disabled" {
		t.Errorf("YOLO 禁用时 yolo_status 应该为 'disabled', 实际 %v", status["yolo_status"])
	}
}

func TestHealthCheckWithMockServer(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			w.WriteHeader(200)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer mockServer.Close()

	client := &ModelClient{
		lstmBaseURL: mockServer.URL,
		yoloBaseURL: mockServer.URL,
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		lstmEnabled: true,
		yoloEnabled: true,
	}

	status := client.HealthCheck()
	if status["lstm_status"] != "healthy" {
		t.Errorf("Mock 服务器可用时 lstm_status 应该为 'healthy', 实际 %v", status["lstm_status"])
	}
	if status["yolo_status"] != "healthy" {
		t.Errorf("Mock 服务器可用时 yolo_status 应该为 'healthy', 实际 %v", status["yolo_status"])
	}
	if status["lstm_enabled"] != true {
		t.Error("LSTM 启用时 lstm_enabled 应该为 true")
	}
	if status["yolo_enabled"] != true {
		t.Error("YOLO 启用时 yolo_enabled 应该为 true")
	}
}

func TestHealthCheckUnhealthy(t *testing.T) {
	client := &ModelClient{
		lstmBaseURL: "http://localhost:1", // 不可达的端口
		yoloBaseURL: "http://localhost:1",
		httpClient:  &http.Client{Timeout: 500 * time.Millisecond},
		lstmEnabled: true,
		yoloEnabled: true,
	}

	status := client.HealthCheck()
	if status["lstm_status"] != "unhealthy" {
		t.Errorf("服务器不可用时 lstm_status 应该为 'unhealthy', 实际 %v", status["lstm_status"])
	}
	if status["yolo_status"] != "unhealthy" {
		t.Errorf("服务器不可用时 yolo_status 应该为 'unhealthy', 实际 %v", status["yolo_status"])
	}
}

func TestHealthCheckMixedStatus(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	}))
	defer mockServer.Close()

	client := &ModelClient{
		lstmBaseURL: mockServer.URL,
		yoloBaseURL: "http://localhost:1", // 不可达
		httpClient:  &http.Client{Timeout: 500 * time.Millisecond},
		lstmEnabled: true,
		yoloEnabled: true,
	}

	status := client.HealthCheck()
	if status["lstm_status"] != "healthy" {
		t.Errorf("LSTM 服务器可用时 lstm_status 应该为 'healthy', 实际 %v", status["lstm_status"])
	}
	if status["yolo_status"] != "unhealthy" {
		t.Errorf("YOLO 服务器不可用时 yolo_status 应该为 'unhealthy', 实际 %v", status["yolo_status"])
	}
}

// ==================== 并发安全测试 ====================

func TestModelClientConcurrentAccess(t *testing.T) {
	client := &ModelClient{
		lstmBaseURL: "",
		yoloBaseURL: "",
		httpClient:  &http.Client{Timeout: 5 * time.Second},
		lstmEnabled: false,
		yoloEnabled: false,
	}

	done := make(chan bool, 4)

	// 并发读取
	go func() {
		for i := 0; i < 100; i++ {
			client.IsLSTMEnabled()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			client.IsYOLOEnabled()
		}
		done <- true
	}()

	// 并发写入
	go func() {
		for i := 0; i < 100; i++ {
			client.SetLSTMURL("http://localhost:5000")
			client.SetLSTMURL("")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			client.SetYOLOURL("http://localhost:5001")
			client.SetYOLOURL("")
		}
		done <- true
	}()

	for i := 0; i < 4; i++ {
		<-done
	}
}
