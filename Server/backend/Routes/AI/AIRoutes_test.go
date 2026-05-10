package AI

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	Models "MavlinkProject/Models"
)

// ==================== SetupAIRoutes 测试 ====================

func TestSetupAIRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupAIRoutes(r)

	routes := r.Routes()

	expectedRoutes := map[string]string{
		"POST:/api/ai/analyze/sensor":  "not_found",
		"POST:/api/ai/analyze/drone":   "not_found",
		"GET:/api/ai/alerts/ws":        "not_found",
		"GET:/api/ai/alerts/sse":       "not_found",
		"GET:/api/ai/alerts/history":   "not_found",
		"GET:/api/ai/model/status":     "not_found",
	}

	for _, route := range routes {
		key := route.Method + ":" + route.Path
		if _, ok := expectedRoutes[key]; ok {
			expectedRoutes[key] = "found"
		}
	}

	for routeKey, status := range expectedRoutes {
		if status == "not_found" {
			t.Errorf("路由 %s 未注册", routeKey)
		}
	}
}

func TestSetupAIRoutesCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupAIRoutes(r)

	routes := r.Routes()
	aiRoutes := 0
	for _, route := range routes {
		if len(route.Path) >= len("/api/ai/") && route.Path[:len("/api/ai/")] == "/api/ai/" {
			aiRoutes++
		}
	}

	if aiRoutes != 6 {
		t.Errorf("AI 路由数量不匹配: 期望 6, 实际 %d", aiRoutes)
	}
}

// ==================== HandleModelStatus 测试 ====================

func TestHandleModelStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 确保 ModelClient 已初始化
	_ = Models.GetModelClient()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/ai/model/status", nil)

	HandleModelStatus(c)

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

	status, ok := response["status"].(map[string]interface{})
	if !ok {
		t.Fatal("响应应该包含 status 字段且为 map 类型")
	}

	// 验证 status 包含必要字段
	requiredFields := []string{"lstm_enabled", "yolo_enabled", "lstm_status", "yolo_status"}
	for _, field := range requiredFields {
		if _, exists := status[field]; !exists {
			t.Errorf("status 应该包含字段 '%s'", field)
		}
	}
}

func TestHandleModelStatusDisabledModels(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 使用全局单例，先确保 URL 为空（模型禁用）
	client := Models.GetModelClient()
	client.SetLSTMURL("")
	client.SetYOLOURL("")

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

// ==================== 路由集成测试 ====================

func TestAIRoutesIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	SetupAIRoutes(r)

	// 测试 /api/ai/model/status 路由
	req := httptest.NewRequest("GET", "/api/ai/model/status", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /api/ai/model/status 期望状态码 200, 实际 %d", w.Code)
	}
}

func TestAIRoutesAlertHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	SetupAIRoutes(r)

	req := httptest.NewRequest("GET", "/api/ai/alerts/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("GET /api/ai/alerts/history 期望状态码 200, 实际 %d", w.Code)
	}
}
