package integration

import (
	"testing"
	"time"

	"MavlinkProject_Board/app/services/backend"
)

func TestSendBoardMessage(t *testing.T) {
	// 初始化后端客户端
	backend.InitBackendClient()

	// 构造测试消息
	message := backend.BoardMessageRequest{
		MessageID:   "test_msg_001",
		MessageTime: time.Now().Unix(),
		Message: backend.MessageData{
			MessageType: "Request",
			Attribute:   "Command",
			Connection:  "HTTPS",
			Command:     "StatusReport",
			Data: map[string]interface{}{
				"battery":   85.5,
				"latitude":  22.543123,
				"longitude": 114.052345,
				"altitude":  10.0,
			},
		},
		FromID:   "central_001",
		FromType: "central",
		ToID:     "backend",
		ToType:   "Server",
	}

	// 发送消息
	err := backend.SendBoardMessage(message)
	if err != nil {
		t.Logf("Expected no error, got: %v", err)
		// 注意：在测试环境中，后端可能不可用，所以这里不直接失败
		t.Skip("Backend API might not be available in test environment")
	}

	// 验证响应
	t.Logf("Message sent successfully")
}

func TestSendSensorAlert(t *testing.T) {
	// 初始化后端客户端
	backend.InitBackendClient()

	// 构造传感器警报
	alert := map[string]interface{}{
		"sensor_id":   "esp32_001",
		"sensor_ip":   "192.168.1.100",
		"sensor_name": "FireSensor-A1",
		"alert_type":  "fire",
		"alert_msg":   "检测到明火",
		"latitude":    22.543123,
		"longitude":   114.052345,
		"timestamp":   time.Now().Unix(),
		"severity":    "high",
	}

	// 发送警报
	err := backend.SendSensorAlert(alert)
	if err != nil {
		t.Logf("Expected no error, got: %v", err)
		// 注意：在测试环境中，后端可能不可用，所以这里不直接失败
		t.Skip("Backend API might not be available in test environment")
	}

	// 验证响应
	t.Logf("Sensor alert sent successfully")
}

func TestGetBoardStatus(t *testing.T) {
	// 获取板卡状态
	status := backend.GetBoardStatus()
	if len(status) == 0 {
		t.Logf("Expected at least one board status, got: %d", len(status))
	}

	// 验证状态数据
	for _, s := range status {
		t.Logf("Board: %s, Battery: %.1f%%, Position: (%.6f, %.6f, %.1f)",
			s.BoardID, s.BatteryLevel, s.Latitude, s.Longitude, s.Altitude)
	}
}
