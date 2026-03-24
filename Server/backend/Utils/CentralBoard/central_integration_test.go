package main

import (
	"encoding/json"
	"fmt"
	"net"
	"testing"
	"time"

	boardHandler "MavlinkProject/Server/backend/Handler/Boards"
	Board "MavlinkProject/Server/backend/Shared/Boards"
)

// 辅助函数：发送 UDP 模拟心跳
// 这里发送的是 JSON 格式的 BoardMessage，因为 BoardConnectionManager 会解析 JSON
func sendSimulatedHeartbeat(t *testing.T, address, port, boardID string) {
	serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", address, port))
	if err != nil {
		t.Fatalf("无法解析 UDP 地址: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		t.Fatalf("无法连接 UDP 服务器: %v", err)
	}
	defer conn.Close()

	// 构造 BoardMessage
	// DroneSearch 逻辑只提取 Data 中的 status, battery, latitude, longitude, altitude 等字段
	msgPayload := Board.BoardMessage{
		FromID:      boardID,
		ToID:        "Central",
		MessageTime: time.Now(),
		Message: Board.Message{
			MessageType: "Status",
			Attribute:   Board.MessageAttribute_Status,
			Connection:  Board.Connection_UDP,
			Command:     "Heartbeat", // 只要是合法的 CommandType 即可，或者自定义
			Data: map[string]interface{}{
				"status":       "idle",
				"battery":      95.5,
				"latitude":     39.9042,
				"longitude":    116.4074,
				"altitude":     100.0,
				"system_id":    1.0,
				"component_id": 1.0,
			},
		},
	}
	
	msgBytes, _ := json.Marshal(msgPayload)
	_, err = conn.Write(msgBytes)
	if err != nil {
		t.Fatalf("无法发送 UDP 包: %v", err)
	}
	t.Logf("已发送模拟心跳给 Board %s", boardID)
}

func TestCentralBoardSystem(t *testing.T) {
	// 1. 设置测试配置
	testPort := "18088" // 使用非标准端口
	testDronePort := "14598"
	testBoardID := "test_drone_01"

	// 2. 启动 CentralServer
	cs := NewCentralServer("127.0.0.1", testPort)
	go func() {
		if err := cs.Start(); err != nil {
			// 在实际测试中 Start 可能会阻塞，这里简单处理
			t.Logf("CentralServer stopped or failed: %v", err)
		}
	}()
	defer cs.Stop()
	
	// 等待服务器启动
	time.Sleep(1 * time.Second)
	if !cs.running {
		t.Fatal("CentralServer 未能成功启动")
	}
	t.Log("CentralServer 启动成功")

	// 3. 启动 Board Manager 并模拟一个 Drone 连接 (UDP)
	// DroneSearch (在 CentralServer 初始化时已获取) 依赖 BoardConnectionManager
	bm := boardHandler.GetBoardManager()
	// 如果之前的测试未能清理，可能需要先清理
	bm.StopAll()
	
	err := bm.StartUDPServer(testBoardID, "127.0.0.1", testDronePort)
	if err != nil {
		t.Fatalf("无法启动模拟 Board UDP Server: %v", err)
	}
	t.Logf("模拟 Board UDP Server 启动在端口 %s", testDronePort)

	// 4. 发送模拟心跳，让 DroneSearch '发现' 这个 Drone
	// 需要多次发送以确状态被更新 (StatusCheckTimeout 默认为 5s)
	for i := 0; i < 3; i++ {
		sendSimulatedHeartbeat(t, "127.0.0.1", testDronePort, testBoardID)
		time.Sleep(500 * time.Millisecond)
	}

	// 5. 模拟客户端连接 CentralServer 并发送任务链
	conn, err := net.Dial("tcp", "127.0.0.1:"+testPort)
	if err != nil {
		t.Fatalf("无法连接到 CentralServer: %v", err)
	}
	defer conn.Close()

	// 构造任务链请求
	taskChain := ProgressChain{
		ChainID: "chain_test_001",
		Status:  TaskStatusPending,
		Tasks: []Task{
			{
				TaskID:  "task_01",
				Command: "TAKEOFF",
				Status:  TaskStatusPending,
				Data: map[string]interface{}{
					"altitude": 10,
				},
			},
		},
	}
	
	// 包装在 BoardMessage 中发送
	reqMsg := Board.BoardMessage{
		FromID:   "GroundStation",
		ToID:     "Central",
		MessageTime: time.Now(),
		Message: Board.Message{
			MessageType: "Request", 
			Attribute:   Board.MessageAttribute_Command,
			Command:     "ExecuteChain", // 假设有这样的 Command
			Data: map[string]interface{}{
				"progress_chain": taskChain, 
			},
		},
	}

	reqBytes, _ := json.Marshal(reqMsg)
	_, err = conn.Write(reqBytes)
	if err != nil {
		t.Fatalf("发送任务链失败: %v", err)
	}

	// 6. 读取响应
	buffer := make([]byte, 1024)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buffer)
	// 即使超时也不一定致命，因为我们更关心后台逻辑是否运行
	if err != nil {
		t.Logf("读取响应超时 (可能正常): %v", err)
	} else {
		t.Logf("收到服务器响应: %s", string(buffer[:n]))
	}

	// 7. 等待一段时间让后台任务处理器运行
	time.Sleep(2 * time.Second)

	// 8. 验证任务链状态
	cs.mu.RLock()
	// 检查 activeChains (因为如果分配成功，它会从 taskChains 移到 activeChains)
	var foundChain *ProgressChain
	if chain, ok := cs.activeChains["chain_test_001"]; ok {
		foundChain = chain
	} else if chain, ok := cs.taskChains["chain_test_001"]; ok {
		foundChain = chain
	}
	cs.mu.RUnlock()

	if foundChain == nil {
		t.Errorf("测试失败: 任务链 chain_test_001 未能在 CentralServer 中找到 (既不在 taskChains 也不在 activeChains)")
	} else {
		t.Logf("验证成功: 任务链已注册，当前状态: %s，分配无人机: %s", foundChain.Status, foundChain.AssignedDrone)
		if foundChain.AssignedDrone == testBoardID {
			t.Logf("PASS: 任务成功分配给了预期的无人机 %s", testBoardID)
		} else if foundChain.Status == "failed" {
			t.Logf("WARN: 任务链状态为 failed")
		}
	}
}
