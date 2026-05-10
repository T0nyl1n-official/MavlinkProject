package AI

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	Models "MavlinkProject/Models"
)

// ==================== AlertHistoryStore 测试 ====================

func TestAlertHistoryStoreAdd(t *testing.T) {
	store := &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}

	alert := Models.AlertJSON{
		AlertID:     "alert_001",
		AlertType:   "anomaly",
		Severity:    Models.SeverityHigh,
		AnomalyType: Models.AnomalyFire,
		Source:      Models.SourceSensor,
		Timestamp:   1234567890,
	}

	store.Add(alert)

	all := store.GetAll()
	if len(all) != 1 {
		t.Fatalf("添加 1 条告警后应该有 1 条记录, 实际 %d 条", len(all))
	}
	if all[0].AlertID != "alert_001" {
		t.Errorf("AlertID 不匹配: 期望 alert_001, 实际 %s", all[0].AlertID)
	}
}

func TestAlertHistoryStoreAddMultiple(t *testing.T) {
	store := &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}

	for i := 0; i < 5; i++ {
		store.Add(Models.AlertJSON{
			AlertID:   fmt.Sprintf("alert_%03d", i),
			Severity:  Models.SeverityMedium,
			Timestamp: int64(i),
		})
	}

	all := store.GetAll()
	if len(all) != 5 {
		t.Fatalf("添加 5 条告警后应该有 5 条记录, 实际 %d 条", len(all))
	}
}

func TestAlertHistoryStoreGetAll(t *testing.T) {
	store := &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}

	// 空存储
	all := store.GetAll()
	if len(all) != 0 {
		t.Errorf("空存储应该返回 0 条记录, 实际 %d 条", len(all))
	}

	// 添加数据
	store.Add(Models.AlertJSON{AlertID: "alert_001"})
	store.Add(Models.AlertJSON{AlertID: "alert_002"})

	all = store.GetAll()
	if len(all) != 2 {
		t.Fatalf("应该返回 2 条记录, 实际 %d 条", len(all))
	}

	// 验证返回的是副本，修改不影响原数据
	all[0].AlertID = "modified"
	original := store.GetAll()
	if original[0].AlertID != "alert_001" {
		t.Error("GetAll 应该返回副本，修改返回值不应该影响原数据")
	}
}

func TestAlertHistoryStoreGetBySeverity(t *testing.T) {
	store := &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}

	store.Add(Models.AlertJSON{AlertID: "alert_001", Severity: Models.SeverityCritical})
	store.Add(Models.AlertJSON{AlertID: "alert_002", Severity: Models.SeverityHigh})
	store.Add(Models.AlertJSON{AlertID: "alert_003", Severity: Models.SeverityCritical})
	store.Add(Models.AlertJSON{AlertID: "alert_004", Severity: Models.SeverityLow})
	store.Add(Models.AlertJSON{AlertID: "alert_005", Severity: Models.SeverityInfo})

	// 查询 critical
	critical := store.GetBySeverity(Models.SeverityCritical)
	if len(critical) != 2 {
		t.Fatalf("critical 告警应该有 2 条, 实际 %d 条", len(critical))
	}

	// 查询 high
	high := store.GetBySeverity(Models.SeverityHigh)
	if len(high) != 1 {
		t.Fatalf("high 告警应该有 1 条, 实际 %d 条", len(high))
	}

	// 查询不存在的严重级别
	nonexistent := store.GetBySeverity("nonexistent")
	if len(nonexistent) != 0 {
		t.Errorf("不存在的严重级别应该返回 0 条记录, 实际 %d 条", len(nonexistent))
	}
}

func TestAlertHistoryStoreGetRecent(t *testing.T) {
	store := &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}

	for i := 0; i < 10; i++ {
		store.Add(Models.AlertJSON{
			AlertID:   fmt.Sprintf("alert_%03d", i),
			Timestamp: int64(i),
		})
	}

	// 获取最近 3 条
	recent := store.GetRecent(3)
	if len(recent) != 3 {
		t.Fatalf("获取最近 3 条应该返回 3 条, 实际 %d 条", len(recent))
	}
	if recent[0].AlertID != "alert_007" {
		t.Errorf("最近 3 条的第一条应该是 alert_007, 实际 %s", recent[0].AlertID)
	}
	if recent[2].AlertID != "alert_009" {
		t.Errorf("最近 3 条的最后一条应该是 alert_009, 实际 %s", recent[2].AlertID)
	}

	// 获取数量超过总数
	allRecent := store.GetRecent(100)
	if len(allRecent) != 10 {
		t.Fatalf("获取数量超过总数时应该返回全部 10 条, 实际 %d 条", len(allRecent))
	}

	// 获取 0 条
	zero := store.GetRecent(0)
	if len(zero) != 0 {
		t.Errorf("获取 0 条应该返回 0 条, 实际 %d 条", len(zero))
	}
}

func TestAlertHistoryStoreMaxLen(t *testing.T) {
	store := &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 5,
	}

	// 添加 10 条告警，超过 maxLen=5
	for i := 0; i < 10; i++ {
		store.Add(Models.AlertJSON{
			AlertID:   fmt.Sprintf("alert_%03d", i),
			Timestamp: int64(i),
		})
	}

	all := store.GetAll()
	if len(all) != 5 {
		t.Fatalf("超过 maxLen 后应该只保留 5 条, 实际 %d 条", len(all))
	}

	// 验证保留的是最新的 5 条
	if all[0].AlertID != "alert_005" {
		t.Errorf("第一条应该是 alert_005, 实际 %s", all[0].AlertID)
	}
	if all[4].AlertID != "alert_009" {
		t.Errorf("最后一条应该是 alert_009, 实际 %s", all[4].AlertID)
	}
}

// ==================== HTTP Handler 测试 ====================

func TestHandleSensorAnalysis(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 保存并替换全局 alertHistory
	originalHistory := alertHistory
	alertHistory = &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	defer func() { alertHistory = originalHistory }()

	// 确保 AnalysisService 已初始化（模型禁用状态）
	_ = GetAnalysisService()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"sensor_id": "sensor_001",
		"sensor_type": "temperature",
		"time_series": [{"timestamp": 1234567890, "value": 25.5}],
		"latitude": 39.9,
		"longitude": 116.4
	}`

	c.Request = httptest.NewRequest("POST", "/api/ai/analyze/sensor", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	HandleSensorAnalysis(c)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 实际 %d, 响应: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if code, ok := response["code"].(float64); !ok || code != 0 {
		t.Errorf("期望 code=0, 实际 %v", response["code"])
	}

	if _, ok := response["alert"]; !ok {
		t.Error("响应应该包含 alert 字段")
	}
}

func TestHandleSensorAnalysisInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	alertHistory = &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	defer func() { alertHistory = originalHistory }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("POST", "/api/ai/analyze/sensor", strings.NewReader("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	HandleSensorAnalysis(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("无效 JSON 应该返回 400, 实际 %d", w.Code)
	}
}

func TestHandleSensorAnalysisEmptyTimeSeries(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	alertHistory = &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	defer func() { alertHistory = originalHistory }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"sensor_id": "sensor_001",
		"sensor_type": "temperature",
		"time_series": []
	}`

	c.Request = httptest.NewRequest("POST", "/api/ai/analyze/sensor", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	HandleSensorAnalysis(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("空 time_series 应该返回 400, 实际 %d", w.Code)
	}
}

func TestHandleDroneImageAnalysis(t *testing.T) {
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

	body := `{
		"drone_id": "drone_001",
		"image_base64": "dGVzdGltYWdl",
		"latitude": 39.9,
		"longitude": 116.4
	}`

	c.Request = httptest.NewRequest("POST", "/api/ai/analyze/drone", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	HandleDroneImageAnalysis(c)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 实际 %d, 响应: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if code, ok := response["code"].(float64); !ok || code != 0 {
		t.Errorf("期望 code=0, 实际 %v", response["code"])
	}
}

func TestHandleDroneImageAnalysisNoImage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	alertHistory = &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	defer func() { alertHistory = originalHistory }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// 没有 image_base64 也没有 image_url
	body := `{
		"drone_id": "drone_001"
	}`

	c.Request = httptest.NewRequest("POST", "/api/ai/analyze/drone", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	HandleDroneImageAnalysis(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("没有图片数据应该返回 400, 实际 %d", w.Code)
	}
}

func TestHandleDroneImageAnalysisMissingDroneID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	alertHistory = &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	defer func() { alertHistory = originalHistory }()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"image_base64": "dGVzdGltYWdl"
	}`

	c.Request = httptest.NewRequest("POST", "/api/ai/analyze/drone", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	HandleDroneImageAnalysis(c)

	if w.Code != http.StatusBadRequest {
		t.Errorf("缺少 drone_id 应该返回 400, 实际 %d", w.Code)
	}
}

func TestHandleDroneImageAnalysisWithImageURL(t *testing.T) {
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

	body := `{
		"drone_id": "drone_002",
		"image_url": "http://example.com/image.jpg",
		"latitude": 40.0,
		"longitude": 117.0
	}`

	c.Request = httptest.NewRequest("POST", "/api/ai/analyze/drone", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	HandleDroneImageAnalysis(c)

	if w.Code != http.StatusOK {
		t.Errorf("使用 image_url 应该返回 200, 实际 %d, 响应: %s", w.Code, w.Body.String())
	}
}

func TestHandleAlertHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	testHistory := &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	alertHistory = testHistory
	defer func() { alertHistory = originalHistory }()

	// 添加测试数据
	testHistory.Add(Models.AlertJSON{AlertID: "alert_001", Severity: Models.SeverityCritical, Timestamp: 1})
	testHistory.Add(Models.AlertJSON{AlertID: "alert_002", Severity: Models.SeverityLow, Timestamp: 2})
	testHistory.Add(Models.AlertJSON{AlertID: "alert_003", Severity: Models.SeverityHigh, Timestamp: 3})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/ai/alerts/history", nil)

	HandleAlertHistory(c)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 实际 %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if code, ok := response["code"].(float64); !ok || code != 0 {
		t.Errorf("期望 code=0, 实际 %v", response["code"])
	}

	if count, ok := response["count"].(float64); !ok || int(count) != 3 {
		t.Errorf("期望 count=3, 实际 %v", response["count"])
	}
}

func TestHandleAlertHistoryWithSeverity(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	testHistory := &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	alertHistory = testHistory
	defer func() { alertHistory = originalHistory }()

	testHistory.Add(Models.AlertJSON{AlertID: "alert_001", Severity: Models.SeverityCritical, Timestamp: 1})
	testHistory.Add(Models.AlertJSON{AlertID: "alert_002", Severity: Models.SeverityLow, Timestamp: 2})
	testHistory.Add(Models.AlertJSON{AlertID: "alert_003", Severity: Models.SeverityCritical, Timestamp: 3})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/ai/alerts/history?severity=critical", nil)

	HandleAlertHistory(c)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 实际 %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if count, ok := response["count"].(float64); !ok || int(count) != 2 {
		t.Errorf("severity=critical 应该返回 2 条, 实际 %v", response["count"])
	}
}

func TestHandleAlertHistoryWithLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	testHistory := &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	alertHistory = testHistory
	defer func() { alertHistory = originalHistory }()

	for i := 0; i < 20; i++ {
		testHistory.Add(Models.AlertJSON{
			AlertID:   fmt.Sprintf("alert_%03d", i),
			Severity:  Models.SeverityMedium,
			Timestamp: int64(i),
		})
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/ai/alerts/history?limit=5", nil)

	HandleAlertHistory(c)

	if w.Code != http.StatusOK {
		t.Errorf("期望状态码 200, 实际 %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("解析响应失败: %v", err)
	}

	if count, ok := response["count"].(float64); !ok || int(count) != 5 {
		t.Errorf("limit=5 应该返回 5 条, 实际 %v", response["count"])
	}
}

func TestHandleAlertHistoryInvalidLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	testHistory := &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	alertHistory = testHistory
	defer func() { alertHistory = originalHistory }()

	testHistory.Add(Models.AlertJSON{AlertID: "alert_001", Severity: Models.SeverityInfo, Timestamp: 1})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/ai/alerts/history?limit=invalid", nil)

	HandleAlertHistory(c)

	if w.Code != http.StatusOK {
		t.Errorf("无效 limit 应该使用默认值并返回 200, 实际 %d", w.Code)
	}
}

func TestHandleAlertHistoryLimitOverMax(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	testHistory := &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	alertHistory = testHistory
	defer func() { alertHistory = originalHistory }()

	for i := 0; i < 600; i++ {
		testHistory.Add(Models.AlertJSON{
			AlertID:   fmt.Sprintf("alert_%03d", i),
			Severity:  Models.SeverityInfo,
			Timestamp: int64(i),
		})
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/ai/alerts/history?limit=600", nil)

	HandleAlertHistory(c)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// limit 超过 500 应该被限制为 500
	if count, ok := response["count"].(float64); !ok || int(count) != 500 {
		t.Errorf("limit 超过 500 应该被限制为 500, 实际 %v", response["count"])
	}
}

// ==================== SensorAnalysisRequest 测试 ====================

func TestSensorAnalysisRequestBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	originalHistory := alertHistory
	alertHistory = &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
	defer func() { alertHistory = originalHistory }()

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "缺少 sensor_id",
			body:       `{"sensor_type": "temperature", "time_series": [{"timestamp": 1, "value": 25.5}]}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "缺少 sensor_type",
			body:       `{"sensor_id": "sensor_001", "time_series": [{"timestamp": 1, "value": 25.5}]}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "缺少 time_series",
			body:       `{"sensor_id": "sensor_001", "sensor_type": "temperature"}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "完整请求",
			body:       `{"sensor_id": "sensor_001", "sensor_type": "temperature", "time_series": [{"timestamp": 1, "value": 25.5}]}`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/api/ai/analyze/sensor", bytes.NewReader([]byte(tt.body)))
			c.Request.Header.Set("Content-Type", "application/json")

			HandleSensorAnalysis(c)

			if w.Code != tt.wantStatus {
				t.Errorf("期望状态码 %d, 实际 %d, 响应: %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}
