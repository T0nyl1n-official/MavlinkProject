package AI

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	Models "MavlinkProject/Models"
)

// ==================== HandleDronePhotoUpload 测试 ====================

func TestHandleDronePhotoUploadSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	alertHistory = &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	defer func() { alertHistory = originalHistory }()

	_ = GetAnalysisService()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("photo", "test_thermal.jpg")
	if err != nil {
		t.Fatalf("创建 form file 失败: %v", err)
	}
	part.Write([]byte("fake thermal image data for upload test"))
	writer.WriteField("drone_id", "test_drone_001")
	writer.WriteField("latitude", "39.9042")
	writer.WriteField("longitude", "116.4074")
	writer.Close()

	c.Request = httptest.NewRequest("POST", "/api/ai/drone/photo", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	HandleDronePhotoUpload(c)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 实际 %d, 响应: %s", w.Code, w.Body.String())
	}

	var response DronePhotoResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if response.Code != 0 {
		t.Errorf("期望 code=0, 实际 %d, message: %s", response.Code, response.Message)
	}
	if response.Alert == nil {
		t.Error("响应应该包含 alert 字段")
	}
	if response.RawResult == nil {
		t.Error("响应应该包含 raw_result 字段")
	}
	if response.PhotoPath == "" {
		t.Error("响应应该包含 photo_path 字段")
	}
	if !strings.Contains(response.PhotoPath, "test_drone_001") {
		t.Errorf("photo_path 应该包含 drone_id: %s", response.PhotoPath)
	}
}

func TestHandleDronePhotoUploadMissingPhoto(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("drone_id", "drone_no_photo")
	writer.Close()

	c.Request = httptest.NewRequest("POST", "/api/ai/drone/photo", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	HandleDronePhotoUpload(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("缺少 photo 文件应返回 400, 实际 %d, 响应: %s", w.Code, w.Body.String())
	}

	var response DronePhotoResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	if response.Code != 1 {
		t.Errorf("缺少 photo 时 code 应该为 1, 实际 %d", response.Code)
	}
	if !strings.Contains(response.Message, "照片文件") {
		t.Errorf("错误消息应包含 '照片文件', 实际: %s", response.Message)
	}
}

func TestHandleDronePhotoUploadNoDroneID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("photo", "no_drone_id.jpg")
	part.Write([]byte("test data"))
	writer.Close()

	c.Request = httptest.NewRequest("POST", "/api/ai/drone/photo", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	HandleDronePhotoUpload(c)

	if w.Code != http.StatusOK {
		t.Errorf("无 drone_id 应使用默认值并返回 200, 实际 %d, 响应: %s", w.Code, w.Body.String())
	}
}

func TestHandleDronePhotoUploadDefaultCoordinates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	alertHistory = &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	defer func() { alertHistory = originalHistory }()

	_ = GetAnalysisService()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("photo", "default_coords.jpg")
	part.Write([]byte("default coords data"))
	writer.WriteField("drone_id", "drone_coords")
	writer.Close()

	c.Request = httptest.NewRequest("POST", "/api/ai/drone/photo", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	HandleDronePhotoUpload(c)

	if w.Code != http.StatusOK {
		t.Errorf("使用默认坐标应返回 200, 实际 %d", w.Code)
	}

	var response DronePhotoResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	if response.Alert == nil {
		t.Fatal("alert 不应为 nil")
	}
	if response.Alert.Latitude != 0 || response.Alert.Longitude != 0 {
		t.Errorf("默认坐标不匹配: 期望 (0, 0), 实际 (%f, %f)", response.Alert.Latitude, response.Alert.Longitude)
	}
}

func TestHandleDronePhotoUploadWithMockServer(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	// 通过公开 API 配置全局 client 使用 mock server
	cleanup := func() {
		client := Models.GetModelClient()
		client.SetYOLOURL("")
		client.SetYOLOParams(0.10, 0.60, 1024)
	}
	Models.GetModelClient().SetYOLOURL(mockServer.URL)
	Models.GetModelClient().SetYOLOParams(0.10, 0.60, 1024)
	defer cleanup()

	originalHistory := alertHistory
	alertHistory = &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	defer func() { alertHistory = originalHistory }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("photo", "mock_server_test.jpg")
	part.Write([]byte("mock server image"))
	writer.WriteField("drone_id", "drone_mock")
	writer.WriteField("latitude", "31.2304")
	writer.WriteField("longitude", "121.4737")
	writer.Close()

	c.Request = httptest.NewRequest("POST", "/api/ai/drone/photo", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	HandleDronePhotoUpload(c)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 实际 %d, 响应: %s", w.Code, w.Body.String())
	}

	var response DronePhotoResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	if response.Code != 0 {
		t.Errorf("期望 code=0, 实际 %d, message: %s", response.Code, response.Message)
	}
	if response.RawResult == nil {
		t.Error("响应应该包含 raw_result 字段 (来自 mock server)")
	}
}

// ==================== 完整链路集成测试 ====================
// Mock YOLO Server -> thermal_handler -> analysis_service -> AlertJSON output

func TestFullThermalPipeline_MockServerToAlert(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/detect" {
			t.Errorf("请求路径错误: 期望 /api/v1/detect, 实际 %s", r.URL.Path)
		}

		conf := r.URL.Query().Get("conf")
		iou := r.URL.Query().Get("iou")
		imgsz := r.URL.Query().Get("imgsz")
		if conf == "" || iou == "" || imgsz == "" {
			t.Errorf("缺少查询参数: conf=%s, iou=%s, imgsz=%s", conf, iou, imgsz)
		}

		ct := r.Header.Get("Content-Type")
		if !strings.Contains(ct, "multipart/form-data") {
			t.Errorf("Content-Type 错误: %s", ct)
		}

		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			t.Fatalf("解析 multipart form 失败: %v", err)
		}
		_, _, err = r.FormFile("image")
		if err != nil {
			t.Errorf("form 中缺少 'image' 字段: %v", err)
		}

		resp := Models.ThermalDetectResponse{
			Success: true,
			Image:   Models.ThermalImageInfo{Width: 1280, Height: 720},
			Detections: []Models.ThermalDetection{
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{100, 200, 350, 450}, Xywh: [4]float64{225, 325, 250, 250}},
					Confidence: 0.91,
					Temperature: Models.ThermalInfo{
						MeanGray: 215.5,
						Level:     Models.TempLevelHigh1,
					},
				},
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{500, 50, 650, 200}, Xywh: [4]float64{575, 125, 150, 150}},
					Confidence: 0.63,
					Temperature: Models.ThermalInfo{
						MeanGray: 135.0,
						Level:     Models.TempLevelHigh2,
					},
				},
			},
			ElapsedMs: 102.7,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	// 使用公开 API 配置全局 client
	restore := func() {
		Models.GetModelClient().SetYOLOURL("")
		Models.GetModelClient().SetYOLOParams(0.10, 0.60, 1024)
	}
	Models.GetModelClient().SetYOLOURL(mockServer.URL)
	Models.GetModelClient().SetYOLOParams(0.15, 0.50, 1280)
	defer restore()

	service := GetAnalysisService()

	tmpFile, err := os.CreateTemp("", "pipeline_test_*.jpg")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	tmpFile.WriteString("pipeline integration test image")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	alert, rawResult, err := service.ProcessDronePhoto("drone_pipeline", tmpFile.Name(), 31.2304, 121.4737)
	if err != nil {
		t.Fatalf("完整链路 ProcessDronePhoto 失败: %v", err)
	}

	if alert == nil {
		t.Fatal("完整链路: alert 不应该为 nil")
	}
	if rawResult == nil {
		t.Fatal("完整链路: rawResult 不应该为 nil")
	}

	if alert.AlertType != "anomaly" {
		t.Errorf("完整链路: AlertType 应该为 'anomaly', 实际 '%s'", alert.AlertType)
	}
	if alert.Severity != Models.SeverityCritical {
		t.Errorf("完整链路: Severity 应该为 'critical'(来自 HIGH Lv1), 实际 '%s'", alert.Severity)
	}
	if alert.AnomalyType != Models.AnomalyThermal {
		t.Errorf("完整链路: AnomalyType 应该为 '%s', 实际 '%s'", Models.AnomalyThermal, alert.AnomalyType)
	}
	if alert.Source != Models.SourceDrone {
		t.Errorf("完整链路: Source 应该为 '%s', 实际 '%s'", Models.SourceDrone, alert.Source)
	}
	if alert.DroneID != "drone_pipeline" {
		t.Errorf("完整链路: DroneID 不匹配: 期望 drone_pipeline, 实际 '%s'", alert.DroneID)
	}
	if alert.Latitude != 31.2304 {
		t.Errorf("完整链路: Latitude 不匹配: 期望 31.2304, 实际 %f", alert.Latitude)
	}
	if alert.Longitude != 121.4737 {
		t.Errorf("完整链路: Longitude 不匹配: 期望 121.4737, 实际 %f", alert.Longitude)
	}
	if alert.Confidence != 0.91 {
		t.Errorf("完整链路: Confidence 应该取最高 severity 的 0.91, 实际 %f", alert.Confidence)
	}

	if alert.Details == nil {
		t.Fatal("完整链路: Details 不应该为 nil")
	}
	if alert.Details["thermal_detections"] == nil {
		t.Error("完整链路: Details 缺少 thermal_detections")
	}
	if alert.Details["top_detection"] == nil {
		t.Error("完整链路: Details 缺少 top_detection")
	}
	if alert.Details["elapsed_ms"] == nil {
		t.Error("完整链路: Details 缺少 elapsed_ms")
	}

	if !rawResult.Success {
		t.Error("完整链路: rawResult.Success 应该为 true")
	}
	if len(rawResult.Detections) != 2 {
		t.Errorf("完整链路: rawResult.Detections 数量不匹配: 期望 2, 实际 %d", len(rawResult.Detections))
	}
	if rawResult.Image.Width != 1280 {
		t.Errorf("完整链路: Image.Width 不匹配: 期望 1280, 实际 %d", rawResult.Image.Width)
	}
	if rawResult.ElapsedMs != 102.7 {
		t.Errorf("完整链路: ElapsedMs 不匹配: 期望 102.7, 实际 %f", rawResult.ElapsedMs)
	}

	topDet := rawResult.Detections[0]
	if topDet.Temperature.Level != Models.TempLevelHigh1 {
		t.Errorf("完整链路: 最高 severity 检测框 Level 应该为 HIGH Lv1, 实际 %s", topDet.Temperature.Level)
	}
	if topDet.Temperature.MeanGray != 215.5 {
		t.Errorf("完整链路: MeanGray 不匹配: 期望 215.5, 实际 %f", topDet.Temperature.MeanGray)
	}
}

func TestFullThermalPipeline_NoAnomalyResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Models.ThermalDetectResponse{
			Success: true,
			Image:   Models.ThermalImageInfo{Width: 800, Height: 600},
			Detections: []Models.ThermalDetection{
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{50, 50, 150, 250}, Xywh: [4]float64{100, 150, 100, 200}},
					Confidence: 0.42,
					Temperature: Models.ThermalInfo{MeanGray: 95.0, Level: Models.TempLevelNormal},
				},
			},
			ElapsedMs: 45.6,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	restore := func() {
		Models.GetModelClient().SetYOLOURL("")
		Models.GetModelClient().SetYOLOParams(0.10, 0.60, 1024)
	}
	Models.GetModelClient().SetYOLOURL(mockServer.URL)
	Models.GetModelClient().SetYOLOParams(0.10, 0.60, 1024)
	defer restore()

	service := GetAnalysisService()

	tmpFile, _ := os.CreateTemp("", "pipeline_normal_*.jpg")
	tmpFile.WriteString("normal pipeline test")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	alert, _, err := service.ProcessDronePhoto("drone_safe", tmpFile.Name(), 22.5, 114.0)
	if err != nil {
		t.Fatalf("无异常链路失败: %v", err)
	}

	if alert.AlertType != "normal" {
		t.Errorf("无异常链路: AlertType 应该为 'normal', 实际 '%s'", alert.AlertType)
	}
	if alert.Severity != Models.SeverityInfo {
		t.Errorf("无异常链路: Severity 应该为 'info', 实际 '%s'", alert.Severity)
	}
	if alert.Confidence != 1.0 {
		t.Errorf("无异常链路: Confidence 应该为 1.0, 实际 %f", alert.Confidence)
	}
}

func TestFullThermalPipeline_MixedLevelsPickCritical(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := Models.ThermalDetectResponse{
			Success: true,
			Image:   Models.ThermalImageInfo{Width: 1920, Height: 1080},
			Detections: []Models.ThermalDetection{
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{10, 10, 80, 80}, Xywh: [4]float64{45, 45, 70, 70}},
					Confidence: 0.20,
					Temperature: Models.ThermalInfo{MeanGray: 70.0, Level: Models.TempLevelLow2},
				},
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{200, 200, 300, 400}, Xywh: [4]float64{250, 300, 100, 200}},
					Confidence: 0.38,
					Temperature: Models.ThermalInfo{MeanGray: 105.0, Level: Models.TempLevelNormal},
				},
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{500, 100, 700, 300}, Xywh: [4]float64{600, 200, 200, 200}},
					Confidence: 0.72,
					Temperature: Models.ThermalInfo{MeanGray: 165.0, Level: Models.TempLevelHigh2},
				},
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{900, 400, 1200, 700}, Xywh: [4]float64{1050, 550, 300, 300}},
					Confidence: 0.88,
					Temperature: Models.ThermalInfo{MeanGray: 205.0, Level: Models.TempLevelHigh1},
				},
				{
					Box:        Models.ThermalBox{Xyxy: [4]float64{1300, 50, 1380, 130}, Xywh: [4]float64{1340, 90, 80, 80}},
					Confidence: 0.12,
					Temperature: Models.ThermalInfo{MeanGray: 30.0, Level: Models.TempLevelLow1},
				},
			},
			ElapsedMs: 188.9,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockServer.Close()

	restore := func() {
		Models.GetModelClient().SetYOLOURL("")
		Models.GetModelClient().SetYOLOParams(0.10, 0.60, 1024)
	}
	Models.GetModelClient().SetYOLOURL(mockServer.URL)
	Models.GetModelClient().SetYOLOParams(0.10, 0.60, 1024)
	defer restore()

	service := GetAnalysisService()

	tmpFile, _ := os.CreateTemp("", "pipeline_mixed_*.jpg")
	tmpFile.WriteString("mixed levels pipeline test")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	alert, rawResult, err := service.ProcessDronePhoto("drone_mixed", tmpFile.Name(), 34.05, 118.28)
	if err != nil {
		t.Fatalf("混合级别链路失败: %v", err)
	}

	if alert.Severity != Models.SeverityCritical {
		t.Errorf("混合级别链路: 应取最高 critical, 实际 %s", alert.Severity)
	}
	if alert.Confidence != 0.88 {
		t.Errorf("混合级别链路: 应取 critical 对应的 confidence=0.88, 实际 %f", alert.Confidence)
	}
	if alert.AlertType != "anomaly" {
		t.Errorf("混合级别链路: AlertType 应该为 'anomaly', 实际 %s", alert.AlertType)
	}
	if len(rawResult.Detections) != 5 {
		t.Errorf("混合级别链路: rawResult 应保留全部 5 个检测框, 实际 %d", len(rawResult.Detections))
	}
}

func TestFullThermalPipeline_ResponseSerializationRoundTrip(t *testing.T) {
	gin.SetMode(gin.TestMode)

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

	restore := func() {
		Models.GetModelClient().SetYOLOURL("")
		Models.GetModelClient().SetYOLOParams(0.10, 0.60, 1024)
	}
	Models.GetModelClient().SetYOLOURL(mockServer.URL)
	Models.GetModelClient().SetYOLOParams(0.10, 0.60, 1024)
	defer restore()

	service := GetAnalysisService()

	tmpFile, _ := os.CreateTemp("", "pipeline_json_*.jpg")
	tmpFile.WriteString("json roundtrip test")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	alert, rawResult, _ := service.ProcessDronePhoto("drone_json", tmpFile.Name(), 39.9042, 116.4074)

	alertData, err := json.Marshal(alert)
	if err != nil {
		t.Fatalf("序列化 AlertJSON 失败: %v", err)
	}

	var parsedAlert Models.AlertJSON
	if err := json.Unmarshal(alertData, &parsedAlert); err != nil {
		t.Fatalf("反序列化 AlertJSON 失败: %v", err)
	}

	if parsedAlert.AlertID != alert.AlertID {
		t.Errorf("往返后 AlertID 不一致")
	}
	if parsedAlert.Severity != alert.Severity {
		t.Errorf("往返后 Severity 不一致: 期望 %s, 实际 %s", alert.Severity, parsedAlert.Severity)
	}
	if parsedAlert.Confidence != alert.Confidence {
		t.Errorf("往返后 Confidence 不一致: 期望 %f, 实际 %f", alert.Confidence, parsedAlert.Confidence)
	}
	if parsedAlert.AnomalyType != alert.AnomalyType {
		t.Errorf("往返后 AnomalyType 不一致: 期望 %s, 实际 %s", alert.AnomalyType, parsedAlert.AnomalyType)
	}

	rawData, err := json.Marshal(rawResult)
	if err != nil {
		t.Fatalf("序列化 ThermalDetectResponse 失败: %v", err)
	}

	var parsedRaw Models.ThermalDetectResponse
	if err := json.Unmarshal(rawData, &parsedRaw); err != nil {
		t.Fatalf("反序列化 ThermalDetectResponse 失败: %v", err)
	}

	if !parsedRaw.Success {
		t.Error("往返后 Success 不一致")
	}
	if len(parsedRaw.Detections) != len(rawResult.Detections) {
		t.Errorf("往返后 Detections 数量不一致: 期望 %d, 实际 %d", len(rawResult.Detections), len(parsedRaw.Detections))
	}
	if parsedRaw.ElapsedMs != rawResult.ElapsedMs {
		t.Errorf("往返后 ElapsedMs 不一致: 期望 %f, 实际 %f", rawResult.ElapsedMs, parsedRaw.ElapsedMs)
	}
}

func TestHandleDronePhotoUploadSavedFilePath(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	alertHistory = &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	defer func() { alertHistory = originalHistory }()

	_ = GetAnalysisService()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	droneID := "drone_path_test"
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("photo", "my_photo.jpg")
	part.Write([]byte("path test image data"))
	writer.WriteField("drone_id", droneID)
	writer.Close()

	c.Request = httptest.NewRequest("POST", "/api/ai/drone/photo", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	HandleDronePhotoUpload(c)

	var response DronePhotoResponse
	json.Unmarshal(w.Body.Bytes(), &response)

	if response.PhotoPath == "" {
		t.Fatal("photo_path 不应为空")
	}

	if !strings.HasPrefix(response.PhotoPath, thermalUploadDir) && !strings.Contains(response.PhotoPath, "output"+string(os.PathSeparator)+"thermal_photos") {
		t.Errorf("photo_path 应包含 thermal_photos 目录, 实际: %s", response.PhotoPath)
	}
	if !strings.Contains(response.PhotoPath, droneID) {
		t.Errorf("photo_path 应包含 drone_id: %s", droneID)
	}

	if _, err := os.Stat(response.PhotoPath); os.IsNotExist(err) {
		t.Errorf("照片文件未被保存到路径: %s", response.PhotoPath)
	} else {
		os.Remove(response.PhotoPath)
	}
}

func TestHandleDronePhotoUploadAlertRecorded(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	testHistory := &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	alertHistory = testHistory
	defer func() { alertHistory = originalHistory }()

	_ = GetAnalysisService()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("photo", "alert_record.jpg")
	part.Write([]byte("alert record test"))
	writer.WriteField("drone_id", "drone_alert_rec")
	writer.Close()

	c.Request = httptest.NewRequest("POST", "/api/ai/drone/photo", body)
	c.Request.Header.Set("Content-Type", writer.FormDataContentType())

	HandleDronePhotoUpload(c)

	allAlerts := testHistory.GetAll()
	found := false
	for _, a := range allAlerts {
		if a.DroneID == "drone_alert_rec" {
			found = true
			break
		}
	}
	if !found {
		t.Error("上传后的 alert 应该被记录到 alertHistory")
	}

	var response DronePhotoResponse
	json.Unmarshal(w.Body.Bytes(), &response)
	if response.PhotoPath != "" {
		os.Remove(response.PhotoPath)
	}
}
