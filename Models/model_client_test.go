package Models

import (
	"encoding/json"
	"fmt"
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

// ==================== SetYOLOParams / GetYOLOParams 测试 ====================

func TestSetYOLOParamsDefaultValues(t *testing.T) {
	client := &ModelClient{
		yoloConf:  0.10,
		yoloIOU:   0.60,
		yoloImgSz: 1024,
	}

	conf, iou, imgsz := client.GetYOLOParams()
	if conf != 0.10 {
		t.Errorf("默认 yoloConf 应该为 0.10, 实际 %f", conf)
	}
	if iou != 0.60 {
		t.Errorf("默认 yoloIOU 应该为 0.60, 实际 %f", iou)
	}
	if imgsz != 1024 {
		t.Errorf("默认 yoloImgSz 应该为 1024, 实际 %d", imgsz)
	}
}

func TestSetYOLOParamsAllFields(t *testing.T) {
	client := &ModelClient{
		yoloConf:  0.10,
		yoloIOU:   0.60,
		yoloImgSz: 1024,
	}

	client.SetYOLOParams(0.25, 0.45, 640)

	conf, iou, imgsz := client.GetYOLOParams()
	if conf != 0.25 {
		t.Errorf("设置后 yoloConf 应该为 0.25, 实际 %f", conf)
	}
	if iou != 0.45 {
		t.Errorf("设置后 yoloIOU 应该为 0.45, 实际 %f", iou)
	}
	if imgsz != 640 {
		t.Errorf("设置后 yoloImgSz 应该为 640, 实际 %d", imgsz)
	}
}

func TestSetYOLOParamsPartialUpdate(t *testing.T) {
	client := &ModelClient{
		yoloConf:  0.10,
		yoloIOU:   0.60,
		yoloImgSz: 1024,
	}

	// 只更新 conf，其他保持不变
	client.SetYOLOParams(0.30, 0, 0)

	conf, iou, imgsz := client.GetYOLOParams()
	if conf != 0.30 {
		t.Errorf("conf 应该被更新为 0.30, 实际 %f", conf)
	}
	if iou != 0.60 {
		t.Errorf("iou 应该保持默认值 0.60, 实际 %f", iou)
	}
	if imgsz != 1024 {
		t.Errorf("imgsz 应该保持默认值 1024, 实际 %d", imgsz)
	}
}

func TestSetYOLOParamsZeroValuesIgnored(t *testing.T) {
	client := &ModelClient{
		yoloConf:  0.10,
		yoloIOU:   0.60,
		yoloImgSz: 1024,
	}

	// 传入零值或负值应该被忽略
	client.SetYOLOParams(0, 0, 0)

	conf, iou, imgsz := client.GetYOLOParams()
	if conf != 0.10 {
		t.Errorf("传入 0 时 conf 应该保持不变, 实际 %f", conf)
	}
	if iou != 0.60 {
		t.Errorf("传入 0 时 iou 应该保持不变, 实际 %f", iou)
	}
	if imgsz != 1024 {
		t.Errorf("传入 0 时 imgsz 应该保持不变, 实际 %d", imgsz)
	}
}

func TestSetYOLOParamsNegativeValuesIgnored(t *testing.T) {
	client := &ModelClient{
		yoloConf:  0.10,
		yoloIOU:   0.60,
		yoloImgSz: 1024,
	}

	client.SetYOLOParams(-0.1, -0.5, -100)

	conf, iou, imgsz := client.GetYOLOParams()
	if conf != 0.10 {
		t.Errorf("负值不应更新 conf, 实际 %f", conf)
	}
	if iou != 0.60 {
		t.Errorf("负值不应更新 iou, 实际 %f", iou)
	}
	if imgsz != 1024 {
		t.Errorf("负值不应更新 imgsz, 实际 %d", imgsz)
	}
}

func TestGetYOLOParamsConcurrentSafety(t *testing.T) {
	client := &ModelClient{
		yoloConf:  0.10,
		yoloIOU:   0.60,
		yoloImgSz: 1024,
	}

	done := make(chan bool, 2)

	go func() {
		for i := 0; i < 100; i++ {
			client.SetYOLOParams(0.20+float64(i)*0.001, 0.50+float64(i)*0.001, 640+i)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			client.GetYOLOParams()
		}
		done <- true
	}()

	for i := 0; i < 2; i++ {
		<-done
	}
}

// ==================== ThermalDetect 测试 ====================

func TestThermalDetectDisabled(t *testing.T) {
	client := &ModelClient{
		yoloBaseURL: "",
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		yoloEnabled: false,
	}

	resp, err := client.ThermalDetect("/nonexistent/image.jpg")
	if err != nil {
		t.Fatalf("YOLO 禁用时 ThermalDetect 不应该返回错误: %v", err)
	}
	if resp == nil {
		t.Fatal("YOLO 禁用时 ThermalDetect 不应该返回 nil")
	}
	if resp.Success {
		t.Error("YOLO 禁用时 Success 应该为 false")
	}
	if len(resp.Detections) != 0 {
		t.Errorf("YOLO 禁用时 Detections 应该为空, 实际长度 %d", len(resp.Detections))
	}
	if resp.ElapsedMs != 0 {
		t.Errorf("YOLO 禁用时 ElapsedMs 应该为 0, 实际 %f", resp.ElapsedMs)
	}
}

func TestThermalDetectWithMockServer(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 验证请求路径和查询参数
		if r.URL.Path != "/api/v1/detect" {
			t.Errorf("请求路径不匹配: 期望 /api/v1/detect, 实际 %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}

		// 验证查询参数
		conf := r.URL.Query().Get("conf")
		iou := r.URL.Query().Get("iou")
		imgsz := r.URL.Query().Get("imgsz")
		if conf == "" || iou == "" || imgsz == "" {
			t.Errorf("缺少必要的查询参数: conf=%s, iou=%s, imgsz=%s", conf, iou, imgsz)
		}

		// 验证 multipart form-data
		contentType := r.Header.Get("Content-Type")
		if !contains(contentType, "multipart/form-data") {
			t.Errorf("Content-Type 应该包含 multipart/form-data, 实际: %s", contentType)
		}

		// 返回标准 YOLOv8 响应
		resp := ThermalDetectResponse{
			Success: true,
			Image: ThermalImageInfo{
				Width:  640,
				Height: 512,
			},
			Detections: []ThermalDetection{
				{
					Box: ThermalBox{
						Xyxy: [4]float64{164.33, 259.13, 207.07, 279.85},
						Xywh: [4]float64{185.7, 269.49, 42.74, 20.72},
					},
					Confidence: 0.182988,
					Temperature: ThermalInfo{
						MeanGray: 158.8,
						Level:     TempLevelHigh2,
					},
				},
			},
			ElapsedMs: 86.35,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// 创建临时图片文件
	tmpFile, err := os.CreateTemp("", "thermal_test_*.jpg")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFile.WriteString("fake thermal image data")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	client := &ModelClient{
		yoloBaseURL: mockServer.URL,
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		maxRetry:     1,
		retryDelay:   10 * time.Millisecond,
		yoloEnabled:  true,
		yoloConf:     0.10,
		yoloIOU:      0.60,
		yoloImgSz:    1024,
	}

	resp, err := client.ThermalDetect(tmpFile.Name())
	if err != nil {
		t.Fatalf("ThermalDetect 不应该返回错误: %v", err)
	}
	if !resp.Success {
		t.Error("Success 应该为 true")
	}
	if resp.Image.Width != 640 {
		t.Errorf("Image.Width 不匹配: 期望 640, 实际 %d", resp.Image.Width)
	}
	if resp.Image.Height != 512 {
		t.Errorf("Image.Height 不匹配: 期望 512, 实际 %d", resp.Image.Height)
	}
	if len(resp.Detections) != 1 {
		t.Fatalf("Detections 长度不匹配: 期望 1, 实际 %d", len(resp.Detections))
	}
	det := resp.Detections[0]
	if det.Confidence != 0.182988 {
		t.Errorf("Detection Confidence 不匹配: 期望 0.182988, 实际 %f", det.Confidence)
	}
	if det.Temperature.MeanGray != 158.8 {
		t.Errorf("Temperature.MeanGray 不匹配: 期望 158.8, 实际 %f", det.Temperature.MeanGray)
	}
	if det.Temperature.Level != TempLevelHigh2 {
		t.Errorf("Temperature.Level 不匹配: 期望 %s, 实际 %s", TempLevelHigh2, det.Temperature.Level)
	}
	if det.Box.Xyxy[0] != 164.33 || det.Box.Xyxy[3] != 279.85 {
		t.Errorf("Box.Xyxy 不匹配: 期望 [164.33,259.13,207.07,279.85], 实际 %v", det.Box.Xyxy)
	}
	if det.Box.Xywh[0] != 185.7 || det.Box.Xywh[3] != 20.72 {
		t.Errorf("Box.Xywh 不匹配: 期望 [185.7,269.49,42.74,20.72], 实际 %v", det.Box.Xywh)
	}
	if resp.ElapsedMs != 86.35 {
		t.Errorf("ElapsedMs 不匹配: 期望 86.35, 实际 %f", resp.ElapsedMs)
	}
}

func TestThermalDetectMultipleDetections(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ThermalDetectResponse{
			Success: true,
			Image: ThermalImageInfo{
				Width:  1280,
				Height: 720,
			},
			Detections: []ThermalDetection{
				{
					Box:        ThermalBox{Xyxy: [4]float64{100, 200, 300, 400}, Xywh: [4]float64{150, 250, 200, 200}},
					Confidence: 0.95,
					Temperature: ThermalInfo{
						MeanGray: 220.0,
						Level:     TempLevelHigh1,
					},
				},
				{
					Box:        ThermalBox{Xyxy: [4]float64{400, 100, 600, 300}, Xywh: [4]float64{500, 150, 200, 200}},
					Confidence: 0.75,
					Temperature: ThermalInfo{
						MeanGray: 150.0,
						Level:     TempLevelHigh2,
					},
				},
				{
					Box:        ThermalBox{Xyxy: [4]float64{50, 50, 100, 100}, Xywh: [4]float64{62.5, 62.5, 50, 50}},
					Confidence: 0.30,
					Temperature: ThermalInfo{
						MeanGray: 80.0,
						Level:     TempLevelLow2,
					},
				},
				{
					Box:        ThermalBox{Xyxy: [4]float64{700, 500, 800, 600}, Xywh: [4]float64{737.5, 537.5, 100, 100}},
					Confidence: 0.15,
					Temperature: ThermalInfo{
						MeanGray: 40.0,
						Level:     TempLevelLow1,
					},
				},
				{
					Box:        ThermalBox{Xyxy: [4]float64{200, 400, 350, 550}, Xywh: [4]float64{262.5, 462.5, 150, 150}},
					Confidence: 0.55,
					Temperature: ThermalInfo{
						MeanGray: 100.0,
						Level:     TempLevelNormal,
					},
				},
			},
			ElapsedMs: 125.6,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	tmpFile, err := os.CreateTemp("", "thermal_multi_*.jpg")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFile.Write([]byte("multi detection test"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	client := &ModelClient{
		yoloBaseURL: mockServer.URL,
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		maxRetry:     1,
		retryDelay:   10 * time.Millisecond,
		yoloEnabled:  true,
		yoloConf:     0.15,
		yoloIOU:      0.50,
		yoloImgSz:    1280,
	}

	resp, err := client.ThermalDetect(tmpFile.Name())
	if err != nil {
		t.Fatalf("ThermalDetect 不应该返回错误: %v", err)
	}
	if !resp.Success {
		t.Fatal("Success 应该为 true")
	}
	if len(resp.Detections) != 5 {
		t.Fatalf("Detections 数量不匹配: 期望 5, 实际 %d", len(resp.Detections))
	}

	// 验证每个检测框
	levelOrder := []string{TempLevelHigh1, TempLevelHigh2, TempLevelLow2, TempLevelLow1, TempLevelNormal}
	for i, det := range resp.Detections {
		if det.Temperature.Level != levelOrder[i] {
			t.Errorf("Detections[%d].Level 不匹配: 期望 %s, 实际 %s", i, levelOrder[i], det.Temperature.Level)
		}
	}
}

func TestThermalDetectNoDetections(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := ThermalDetectResponse{
			Success:    true,
			Image:      ThermalImageInfo{Width: 800, Height: 600},
			Detections: []ThermalDetection{},
			ElapsedMs:  42.5,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	tmpFile, _ := os.CreateTemp("", "thermal_empty_*.jpg")
	tmpFile.Write([]byte("empty detections"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	client := &ModelClient{
		yoloBaseURL: mockServer.URL,
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		maxRetry:     1,
		retryDelay:   10 * time.Millisecond,
		yoloEnabled:  true,
		yoloConf:     0.10,
		yoloIOU:      0.60,
		yoloImgSz:    1024,
	}

	resp, err := client.ThermalDetect(tmpFile.Name())
	if err != nil {
		t.Fatalf("ThermalDetect 不应该返回错误: %v", err)
	}
	if !resp.Success {
		t.Error("Success 应该为 true (即使无检测结果)")
	}
	if len(resp.Detections) != 0 {
		t.Errorf("无检测时 Detections 应该为空, 实际长度 %d", len(resp.Detections))
	}
}

func TestThermalDetectFileNotFound(t *testing.T) {
	client := &ModelClient{
		yoloBaseURL: "http://localhost:5001",
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		maxRetry:     1,
		retryDelay:   10 * time.Millisecond,
		yoloEnabled:  true,
	}

	_, err := client.ThermalDetect("/nonexistent/path/thermal_image.jpg")
	if err == nil {
		t.Error("文件不存在时 ThermalDetect 应该返回错误")
	}
	if !contains(err.Error(), "open image file failed") {
		t.Errorf("错误信息应包含 'open image file failed', 实际: %v", err)
	}
}

func TestThermalDetectHTTPError(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer mockServer.Close()

	tmpFile, _ := os.CreateTemp("", "thermal_err_*.jpg")
	tmpFile.Write([]byte("error test"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	client := &ModelClient{
		yoloBaseURL: mockServer.URL,
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		maxRetry:     1,
		retryDelay:   10 * time.Millisecond,
		yoloEnabled:  true,
		yoloConf:     0.10,
		yoloIOU:      0.60,
		yoloImgSz:    1024,
	}

	_, err := client.ThermalDetect(tmpFile.Name())
	if err == nil {
		t.Error("HTTP 错误时 ThermalDetect 应该返回错误")
	}
	if !contains(err.Error(), "returned status 500") {
		t.Errorf("错误信息应包含状态码 500, 实际: %v", err)
	}
}

func TestThermalDetectInvalidJSON(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte("invalid json response"))
	}))
	defer mockServer.Close()

	tmpFile, _ := os.CreateTemp("", "thermal_badjson_*.jpg")
	tmpFile.Write([]byte("bad json test"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	client := &ModelClient{
		yoloBaseURL: mockServer.URL,
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		maxRetry:     1,
		retryDelay:   10 * time.Millisecond,
		yoloEnabled:  true,
		yoloConf:     0.10,
		yoloIOU:      0.60,
		yoloImgSz:    1024,
	}

	_, err := client.ThermalDetect(tmpFile.Name())
	if err == nil {
		t.Error("无效 JSON 时 ThermalDetect 应该返回错误")
	}
	if !contains(err.Error(), "response decode failed") {
		t.Errorf("错误信息应包含 'response decode failed', 实际: %v", err)
	}
}

func TestThermalDetectServerUnavailable(t *testing.T) {
	client := &ModelClient{
		yoloBaseURL: "http://localhost:1", // 不可达端口
		httpClient:   &http.Client{Timeout: 500 * time.Millisecond},
		maxRetry:     1,
		retryDelay:   10 * time.Millisecond,
		yoloEnabled:  true,
	}

	tmpFile, _ := os.CreateTemp("", "thermal_unavail_*.jpg")
	tmpFile.Write([]byte("unavailable test"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	_, err := client.ThermalDetect(tmpFile.Name())
	if err == nil {
		t.Error("服务器不可用时 ThermalDetect 应该返回错误")
	}
}

func TestThermalDetectQueryParameters(t *testing.T) {
	var receivedConf, receivedIou string
	var receivedImgSz int

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedConf = r.URL.Query().Get("conf")
		receivedIou = r.URL.Query().Get("iou")
		imgszStr := r.URL.Query().Get("imgsz")
		fmt.Sscanf(imgszStr, "%d", &receivedImgSz)

		resp := ThermalDetectResponse{
			Success:    true,
			Image:      ThermalImageInfo{Width: 640, Height: 480},
			Detections: []ThermalDetection{},
			ElapsedMs:  10.0,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	tmpFile, _ := os.CreateTemp("", "thermal_params_*.jpg")
	tmpFile.Write([]byte("params test"))
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	client := &ModelClient{
		yoloBaseURL: mockServer.URL,
		httpClient:   &http.Client{Timeout: 5 * time.Second},
		maxRetry:     1,
		retryDelay:   10 * time.Millisecond,
		yoloEnabled:  true,
		yoloConf:     0.25,
		yoloIOU:      0.45,
		yoloImgSz:    640,
	}

	_, err := client.ThermalDetect(tmpFile.Name())
	if err != nil {
		t.Fatalf("ThermalDetect 失败: %v", err)
	}

	// 验证查询参数是否正确传递
	if receivedConf != "0.250000" && receivedConf != "0.25" {
		t.Errorf("查询参数 conf 不匹配: 期望包含 0.25, 实际 %s", receivedConf)
	}
	if receivedIou != "0.450000" && receivedIou != "0.45" {
		t.Errorf("查询参数 iou 不匹配: 期望包含 0.45, 实际 %s", receivedIou)
	}
	if receivedImgSz != 640 {
		t.Errorf("查询参数 imgsz 不匹配: 期望 640, 实际 %d", receivedImgSz)
	}
}

// 辅助函数：检查字符串是否包含子串
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
