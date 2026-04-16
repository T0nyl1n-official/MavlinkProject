package backend

import (
	"MavlinkProject_Board/app/services/mavlink"
	"fmt"
	"log"
	"time"

	"github.com/go-resty/resty/v2"
)

var client *resty.Client

func InitBackendClient() {
	client = resty.New()
	client.SetTimeout(30 * time.Second)
	client.SetRetryCount(3)
	client.SetRetryWaitTime(2 * time.Second)
	client.SetRetryMaxWaitTime(10 * time.Second)
}

type BoardMessageRequest struct {
	MessageID   string      `json:"message_id"`
	MessageTime int64       `json:"message_time"`
	Message     MessageData `json:"message"`
	FromID      string      `json:"from_id"`
	FromType    string      `json:"from_type"`
	ToID        string      `json:"to_id"`
	ToType      string      `json:"to_type"`
}

type MessageData struct {
	MessageType string                 `json:"message_type"`
	Attribute   string                 `json:"attribute"`
	Connection  string                 `json:"connection"`
	Command     string                 `json:"command"`
	Data        map[string]interface{} `json:"data"`
}

type BoardStatus struct {
	BoardID      string  `json:"board_id"`
	SystemID     uint8   `json:"system_id"`
	IsIdle       bool    `json:"is_idle"`
	BatteryLevel float64 `json:"battery_level"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Altitude     float64 `json:"altitude"`
	LastUpdate   int64   `json:"last_update"`
}

func SendBoardMessage(req interface{}) error {
	if client == nil {
		InitBackendClient()
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetHeader("Authorization", "Bearer your-device-token").
		SetHeader("X-Device-ID", "central_001").
		SetHeader("X-Device-Type", "central").
		SetBody(req).
		Post("https://api.deeppluse.dpdns.org/api/board/send-message")

	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("backend returned non-200 status: %d, body: %s", resp.StatusCode(), resp.String())
	}

	log.Printf("[Backend] Message sent successfully: %s", resp.String())
	return nil
}

func GetBoardStatus() []BoardStatus {
	// 通过 MAVLink 获取实际的无人机状态
	droneStatus, err := mavlink.GetDroneStatus()
	if err != nil {
		log.Printf("[Backend] Failed to get drone status: %v", err)
		// 如果获取失败，返回默认数据
		return []BoardStatus{
			{
				BoardID:      "Error-NoDroneStatus",
				SystemID:     1,
				IsIdle:       true,
				BatteryLevel: 85.5,
				Latitude:     22.543123,
				Longitude:    114.052345,
				Altitude:     0,
				LastUpdate:   time.Now().Unix(),
			},
		}
	}

	// 构建 BoardStatus 响应
	status := BoardStatus{
		BoardID:      droneStatus["board_id"].(string),
		SystemID:     uint8(droneStatus["system_id"].(uint8)),
		IsIdle:       true, // 这里应该从 MAVLink 状态中判断
		BatteryLevel: droneStatus["battery_level"].(float64),
		Latitude:     droneStatus["latitude"].(float64),
		Longitude:    droneStatus["longitude"].(float64),
		Altitude:     droneStatus["altitude"].(float64),
		LastUpdate:   droneStatus["last_update"].(int64),
	}

	log.Printf("[Backend] Got drone status: BoardID=%s, Battery=%.1f%%, Position=(%.6f, %.6f, %.1f)",
		status.BoardID, status.BatteryLevel, status.Latitude, status.Longitude, status.Altitude)

	return []BoardStatus{status}
}

func SendSensorAlert(alert map[string]interface{}) error {
	if client == nil {
		InitBackendClient()
	}

	resp, err := client.R().
		SetHeader("Content-Type", "application/json").
		SetBody(alert).
		Post("https://api.deeppluse.dpdns.org/api/sensor/message")

	if err != nil {
		return fmt.Errorf("failed to send sensor alert: %v", err)
	}

	if resp.StatusCode() != 200 && resp.StatusCode() != 202 {
		return fmt.Errorf("backend returned non-200/202 status: %d, body: %s", resp.StatusCode(), resp.String())
	}

	log.Printf("[Backend] Sensor alert sent successfully: %s", resp.String())
	return nil
}
