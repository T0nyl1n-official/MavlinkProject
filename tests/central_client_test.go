package tests

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"testing"
	"time"
)

const CentralServerAddress = "frp-put.com:14465"
//const CentralServerAddress = "localhost:8081"

type BoardMessage struct {
	MessageID   string      `json:"message_id"`
	MessageTime time.Time   `json:"message_time"`
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

type ProgressChain struct {
	ChainID     string `json:"chain_id"`
	Tasks       []Task `json:"tasks"`
	CurrentTask int    `json:"current_task"`
	Status      string `json:"status"`
}

type Task struct {
	TaskID  string                 `json:"task_id"`
	Command string                 `json:"command"`
	Data    map[string]interface{} `json:"data"`
	Status  string                 `json:"status"`
}

func SendProgressChainToCentral(chain ProgressChain) error {
	conn, err := net.Dial("tcp", CentralServerAddress)
	if err != nil {
		return fmt.Errorf("连接中央服务器失败: %v", err)
	}
	defer conn.Close()

	boardMsg := BoardMessage{
		MessageID:   fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		MessageTime: time.Now(),
		FromID:      "test_client",
		FromType:    "test",
		ToID:        "central",
		ToType:      "server",
		Message: MessageData{
			MessageType: "Request",
			Attribute:   "Default",
			Connection:  "tcp",
			Command:     "schedule_chain",
			Data: map[string]interface{}{
				"progress_chain": chain,
			},
		},
	}

	data, err := json.Marshal(boardMsg)
	if err != nil {
		return fmt.Errorf("序列化消息失败: %v", err)
	}

	n, err := conn.Write(data)
	if err != nil {
		return fmt.Errorf("发送消息失败: %v", err)
	}
	log.Printf("已发送 %d 字节到中央服务器", n)

	buffer := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	n, err = conn.Read(buffer)
	if err != nil {
		return fmt.Errorf("读取响应失败: %v", err)
	}

	responseStr := string(buffer[:n])
	log.Printf("服务器原始响应: %s", responseStr)

	var response map[string]interface{}
	if err := json.Unmarshal(buffer[:n], &response); err != nil {
		return fmt.Errorf("解析响应失败: %v", err)
	}

	log.Printf("服务器响应: %v", response)
	return nil
}

func TestForceArmTakeOffHoverLandChain(t *testing.T) {
	chain := ProgressChain{
		ChainID: "test_force_arm_takeoff_001",
		Tasks: []Task{
			{
				TaskID:  "task_setmode",
				Command: "SetMode", // 切换为引导模式
				Data: map[string]interface{}{
					"mode": "GUIDED",
				},
				Status: "pending",
			},
			{
				TaskID:  "task_force_arm",
				Command: "Arm", // 强制解锁
				Data: map[string]interface{}{
					"force": true,
				},
				Status: "pending",
			},
			{
				TaskID:  "task_takeoff",
				Command: "TakeOff", // 起飞并悬停
				Data: map[string]interface{}{
					"altitude": 10.0,
					"delay":    10.0, // 悬停 10 秒
				},
				Status: "pending",
			},
			{
				TaskID:  "task_land",
				Command: "Land", // 降落
				Data:    map[string]interface{}{},
				Status:  "pending",
			},
		},
		Status: "pending",
	}

	log.Printf("发送强制解锁-起飞-悬停-降落任务链: %s", chain.ChainID)
	if err := SendProgressChainToCentral(chain); err != nil {
		t.Errorf("发送失败: %v", err)
	}
}
