package tests

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"testing"
	"time"
)

const CentralServerAddress = "frp-sea.com:34565"

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

func TestSimpleTakeOffChain(t *testing.T) {
	chain := ProgressChain{
		ChainID: "test_takeoff_001",
		Tasks: []Task{
			{
				TaskID:  "task_0",
				Command: "TakeOff",
				Data: map[string]interface{}{
					"altitude": 10.0,
				},
				Status: "pending",
			},
		},
		Status: "pending",
	}

	log.Printf("发送简单起飞任务链: %s", chain.ChainID)
	if err := SendProgressChainToCentral(chain); err != nil {
		t.Errorf("发送失败: %v", err)
	}
}

func TestFullMissionChain(t *testing.T) {
	chain := ProgressChain{
		ChainID: "test_mission_001",
		Tasks: []Task{
			{
				TaskID:  "task_0",
				Command: "TakeOff",
				Data: map[string]interface{}{
					"altitude": 15.0,
				},
				Status: "pending",
			},
			{
				TaskID:  "task_1",
				Command: "GoTo",
				Data: map[string]interface{}{
					"latitude":  39.9042,
					"longitude": 116.4074,
					"altitude":  50.0,
				},
				Status: "pending",
			},
			{
				TaskID:  "task_2",
				Command: "TakePhoto",
				Data:    map[string]interface{}{},
				Status:  "pending",
			},
			{
				TaskID:  "task_3",
				Command: "GoTo",
				Data: map[string]interface{}{
					"latitude":  39.9085,
					"longitude": 116.4090,
					"altitude":  30.0,
				},
				Status: "pending",
			},
			{
				TaskID:  "task_4",
				Command: "Land",
				Data:    map[string]interface{}{},
				Status:  "pending",
			},
		},
		Status: "pending",
	}

	log.Printf("发送完整任务链: %s (共 %d 个任务)", chain.ChainID, len(chain.Tasks))
	if err := SendProgressChainToCentral(chain); err != nil {
		t.Errorf("发送失败: %v", err)
	}
}

func TestMultipleChains(t *testing.T) {
	chains := []ProgressChain{
		{
			ChainID: "test_multi_001",
			Tasks: []Task{
				{
					TaskID:  "task_0",
					Command: "TakeOff",
					Data:    map[string]interface{}{"altitude": 10.0},
					Status:  "pending",
				},
				{
					TaskID:  "task_1",
					Command: "Land",
					Data:    map[string]interface{}{},
					Status:  "pending",
				},
			},
			Status: "pending",
		},
		{
			ChainID: "test_multi_002",
			Tasks: []Task{
				{
					TaskID:  "task_0",
					Command: "TakeOff",
					Data:    map[string]interface{}{"altitude": 20.0},
					Status:  "pending",
				},
				{
					TaskID:  "task_1",
					Command: "SetSpeed",
					Data:    map[string]interface{}{"speed": 5.0},
					Status:  "pending",
				},
				{
					TaskID:  "task_2",
					Command: "Land",
					Data:    map[string]interface{}{},
					Status:  "pending",
				},
			},
			Status: "pending",
		},
	}

	for i, chain := range chains {
		log.Printf("发送多任务链 %d/%d: %s", i+1, len(chains), chain.ChainID)
		if err := SendProgressChainToCentral(chain); err != nil {
			t.Errorf("发送任务链 %s 失败: %v", chain.ChainID, err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func TestChainWithPosition(t *testing.T) {
	chain := ProgressChain{
		ChainID: "test_position_001",
		Tasks: []Task{
			{
				TaskID:  "task_0",
				Command: "TakeOff",
				Data: map[string]interface{}{
					"altitude": 12.0,
				},
				Status: "pending",
			},
			{
				TaskID:  "task_1",
				Command: "SetPosition",
				Data: map[string]interface{}{
					"latitude":  39.9042,
					"longitude": 116.4074,
					"altitude":  40.0,
				},
				Status: "pending",
			},
			{
				TaskID:  "task_2",
				Command: "SetPosition",
				Data: map[string]interface{}{
					"latitude":  39.9085,
					"longitude": 116.4090,
					"altitude":  35.0,
				},
				Status: "pending",
			},
			{
				TaskID:  "task_3",
				Command: "Land",
				Data:    map[string]interface{}{},
				Status:  "pending",
			},
		},
		Status: "pending",
	}

	log.Printf("发送位置控制任务链: %s", chain.ChainID)
	if err := SendProgressChainToCentral(chain); err != nil {
		t.Errorf("发送失败: %v", err)
	}
}
