package tests

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	//"strings"
	"testing"
	"time"

	boardHandler "MavlinkProject/Server/backend/Handler/Boards"
	Board "MavlinkProject/Server/backend/Shared/Boards"
)

func TestTCPBoardConnection(t *testing.T) {
	manager := boardHandler.GetBoardManager()
	testBoardID := "tcp_board_listener"
	testAddr := "127.0.0.1"
	testPort := "61001"

	err := manager.StartTCPServer(testBoardID, testAddr, testPort)
	if err != nil {
		t.Fatalf("Failed to start TCP server: %v", err)
	}
	defer manager.StopBoard(testBoardID)

	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", testAddr, testPort))
	if err != nil {
		t.Fatalf("Failed to connect to TCP server: %v", err)
	}
	defer conn.Close()

	testBoardID2 := "esp32_c3_001"
	testMessage := Board.BoardMessage{
		MessageID:   "TEST_MSG_001",
		MessageTime: time.Now(),
		Message: Board.Message{
			MessageType: "Telemetry",
			Attribute:   Board.MessageAttribute_Default,
		},
		FromID:   testBoardID2,
		FromType: string(Board.Type_Drone),
	}

	msgBytes, err := json.Marshal(testMessage)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}
	msgBytes = append(msgBytes, '\n')

	_, err = conn.Write(msgBytes)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	t.Logf("Sent message: %s", string(msgBytes))

	select {
	case receivedMsg := <-manager.GetMessageChan():
		t.Logf("Received message - FromID: %s, ToID: %s, MessageID: %s",
			receivedMsg.FromID, receivedMsg.ToID, receivedMsg.MessageID)
		if receivedMsg.FromID != testBoardID2 {
			t.Errorf("Expected FromID %s, got %s", testBoardID2, receivedMsg.FromID)
		}
		if receivedMsg.ToID != testBoardID {
			t.Errorf("Expected ToID %s, got %s", testBoardID, receivedMsg.ToID)
		}
		if receivedMsg.MessageID != testMessage.MessageID {
			t.Errorf("Expected MessageID %s, got %s", testMessage.MessageID, receivedMsg.MessageID)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestUDPBoardConnection(t *testing.T) {
	manager := boardHandler.GetBoardManager()
	testBoardID := "udp_board_listener"
	testAddr := "127.0.0.1"
	testPort := "61011"

	err := manager.StartUDPServer(testBoardID, testAddr, testPort)
	if err != nil {
		t.Fatalf("Failed to start UDP server: %v", err)
	}
	defer manager.StopBoard(testBoardID)

	time.Sleep(100 * time.Millisecond)

	udpAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", testAddr, testPort))
	if err != nil {
		t.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		t.Fatalf("Failed to create UDP connection: %v", err)
	}
	defer conn.Close()

	testBoardID2 := "esp32_c3_002"
	testMessage := Board.BoardMessage{
		MessageID:   "TEST_MSG_002",
		MessageTime: time.Now(),
		Message: Board.Message{
			MessageType: "Telemetry",
			Attribute:   Board.MessageAttribute_Default,
		},
		FromID:   testBoardID2,
		FromType: string(Board.Type_Drone),
	}

	msgBytes, err := json.Marshal(testMessage)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	_, err = conn.Write(msgBytes)
	if err != nil {
		t.Fatalf("Failed to send message: %v", err)
	}

	t.Logf("Sent message: %s", string(msgBytes))

	select {
	case receivedMsg := <-manager.GetMessageChan():
		t.Logf("Received message - FromID: %s, ToID: %s, MessageID: %s",
			receivedMsg.FromID, receivedMsg.ToID, receivedMsg.MessageID)
		if receivedMsg.FromID != testBoardID2 {
			t.Errorf("Expected FromID %s, got %s", testBoardID2, receivedMsg.FromID)
		}
		if receivedMsg.ToID != testBoardID {
			t.Errorf("Expected ToID %s, got %s", testBoardID, receivedMsg.ToID)
		}
		if receivedMsg.MessageID != testMessage.MessageID {
			t.Errorf("Expected MessageID %s, got %s", testMessage.MessageID, receivedMsg.MessageID)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for message")
	}
}

func TestTCPMultipleBoardsConnection(t *testing.T) {
	manager := boardHandler.GetBoardManager()
	testBoardID := "tcp_multi_board_listener"
	testAddr := "127.0.0.1"
	testPort := "61021"

	err := manager.StartTCPServer(testBoardID, testAddr, testPort)
	if err != nil {
		t.Fatalf("Failed to start TCP server: %v", err)
	}
	defer manager.StopBoard(testBoardID)

	time.Sleep(100 * time.Millisecond)

	boardIDs := []string{"esp32_c3_001", "esp32_c3_002", "esp32_c3_003"}
	var wg sync.WaitGroup
	var receivedMsgs []Board.BoardMessage
	var mu sync.Mutex

	for _, boardID := range boardIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", testAddr, testPort))
			if err != nil {
				t.Errorf("Failed to connect: %v", err)
				return
			}
			defer conn.Close()

			time.Sleep(50 * time.Millisecond)

			msg := Board.BoardMessage{
				MessageID:   fmt.Sprintf("TEST_%s", id),
				MessageTime: time.Now(),
				Message: Board.Message{
					MessageType: "Telemetry",
					Attribute:   Board.MessageAttribute_Default,
				},
				FromID:   id,
				FromType: string(Board.Type_Drone),
			}

			msgBytes, _ := json.Marshal(msg)
			msgBytes = append(msgBytes, '\n')

			_, err = conn.Write(msgBytes)
			if err != nil {
				t.Errorf("Failed to send: %v", err)
				return
			}

			t.Logf("Board %s sent message", id)
		}(boardID)
	}

	go func() {
		wg.Wait()
	}()

	timeout := time.After(5 * time.Second)
	received := 0
	for received < len(boardIDs) {
		select {
		case msg := <-manager.GetMessageChan():
			mu.Lock()
			receivedMsgs = append(receivedMsgs, *msg)
			received++
			mu.Unlock()
			t.Logf("Received from %s: %s", msg.FromID, msg.MessageID)
		case <-timeout:
			t.Errorf("Timeout: received %d out of %d messages", received, len(boardIDs))
			mu.Lock()
			defer mu.Unlock()
			for _, expectedID := range boardIDs {
				found := false
				for _, msg := range receivedMsgs {
					if msg.FromID == expectedID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Message from %s not received", expectedID)
				}
			}
			return
		}
	}

	mu.Lock()
	defer mu.Unlock()
	if len(receivedMsgs) != len(boardIDs) {
		t.Errorf("Expected %d messages, got %d", len(boardIDs), len(receivedMsgs))
	}
	for _, expectedID := range boardIDs {
		found := false
		for _, msg := range receivedMsgs {
			if msg.FromID == expectedID {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Message from %s not received", expectedID)
		}
	}
}

func TestTCPPacketSplitting(t *testing.T) {
	manager := boardHandler.GetBoardManager()
	testBoardID := "tcp_split_listener"
	testAddr := "127.0.0.1"
	testPort := "63031"

	err := manager.StartTCPServer(testBoardID, testAddr, testPort)
	if err != nil {
		t.Fatalf("Failed to start TCP server: %v", err)
	}
	defer manager.StopBoard(testBoardID)

	time.Sleep(100 * time.Millisecond)

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", testAddr, testPort))
	if err != nil {
		t.Fatalf("Failed to connect to TCP server: %v", err)
	}
	defer conn.Close()

	fullMessage := Board.BoardMessage{
		MessageID:   "SPLIT_MSG_001",
		MessageTime: time.Now(),
		Message: Board.Message{
			MessageType: "LargeData",
			Attribute:   Board.MessageAttribute_Default,
			Data:        map[string]interface{}{"key": "value", "data": "abcdefghijklmnopqrstuvwxyz"},
		},
		FromID:   "esp32_c3_split",
		FromType: string(Board.Type_Drone),
	}

	msgBytes, err := json.Marshal(fullMessage)
	if err != nil {
		t.Fatalf("Failed to marshal message: %v", err)
	}

	half := len(msgBytes) / 2
	part1 := msgBytes[:half]
	part2 := msgBytes[half:]
	part1 = append(part1, '\n')
	part2 = append(part2, '\n')

	t.Logf("Sending part 1 (%d bytes): %s", len(part1), string(part1))
	_, err = conn.Write(part1)
	if err != nil {
		t.Fatalf("Failed to send part 1: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	t.Logf("Sending part 2 (%d bytes): %s", len(part2), string(part2))
	_, err = conn.Write(part2)
	if err != nil {
		t.Fatalf("Failed to send part 2: %v", err)
	}

	select {
	case receivedMsg := <-manager.GetMessageChan():
		t.Logf("Received complete message - FromID: %s, MessageID: %s",
			receivedMsg.FromID, receivedMsg.MessageID)
		if receivedMsg.MessageID != fullMessage.MessageID {
			t.Errorf("Expected MessageID %s, got %s", fullMessage.MessageID, receivedMsg.MessageID)
		}
		if receivedMsg.FromID != fullMessage.FromID {
			t.Errorf("Expected FromID %s, got %s", fullMessage.FromID, receivedMsg.FromID)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for message - packet splitting may have failed")
	}
}

func TestSendMessageToBoardByFromID(t *testing.T) {
	manager := boardHandler.GetBoardManager()
	testBoardID := "tcp_send_listener"
	testAddr := "127.0.0.1"
	testPort := "61041"

	err := manager.StartTCPServer(testBoardID, testAddr, testPort)
	if err != nil {
		t.Fatalf("Failed to start TCP server: %v", err)
	}
	defer manager.StopBoard(testBoardID)

	time.Sleep(100 * time.Millisecond)

	boardFromID := "esp32_c3_sender"

	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", testAddr, testPort))
	if err != nil {
		t.Fatalf("Failed to connect to TCP server: %v", err)
	}
	defer conn.Close()

	handshakeMsg := Board.BoardMessage{
		MessageID:   "HANDSHAKE",
		MessageTime: time.Now(),
		FromID:      boardFromID,
		FromType:    string(Board.Type_Drone),
	}

	handshakeBytes, _ := json.Marshal(handshakeMsg)
	handshakeBytes = append(handshakeBytes, '\n')
	conn.Write(handshakeBytes)

	time.Sleep(200 * time.Millisecond)

	outMsg := Board.BoardMessage{
		MessageID:   "RESPONSE",
		MessageTime: time.Now(),
		Message: Board.Message{
			MessageType: "Response",
			Attribute:   Board.MessageAttribute_Default,
		},
		FromID:   testBoardID,
		ToID:     boardFromID,
		FromType: string(Board.Type_Control),
		ToType:   string(Board.Type_Drone),
	}

	err = manager.SendMessageToBoardByFromID(boardFromID, &outMsg)
	if err != nil {
		t.Fatalf("Failed to send message by FromID: %v", err)
	}

	buffer := make([]byte, 4096)
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	n, err := conn.Read(buffer)
	if err != nil {
		t.Fatalf("Failed to read response: %v", err)
	}

	var received Board.BoardMessage
	err = json.Unmarshal(buffer[:n], &received)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	t.Logf("Received response: %s", string(buffer[:n]))
	if received.MessageID != outMsg.MessageID {
		t.Errorf("Expected MessageID %s, got %s", outMsg.MessageID, received.MessageID)
	}
}

func TestSensorAlertToCentral(t *testing.T) {
	reqPayload := map[string]interface{}{
		"sensor_id":   "esp32_sensor_001",
		"latitude":    39.9042,
		"longitude":   116.4074,
		"radius":      50.0,
		"photo_count": 5,
		"altitude":    100.0,
	}

	body, err := json.Marshal(reqPayload)
	if err != nil {
		t.Fatalf("Failed to marshal request payload: %v", err)
	}

	apiURL := "https://localhost:8888/mavlink/v2/sensor-alert"
	t.Logf("Sending POST request to %s with payload: %s", apiURL, string(body))

	// 因为本地是自签发测试TLS，为了防止报错，可以忽略证书校验
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	
	// 先进行用户注册和登录获取 Token 
    apiBaseUrl := "https://api.deeppluse.dpdns.org"
	username := fmt.Sprintf("test_sensor_%d", time.Now().Unix())
	email := fmt.Sprintf("%s@example.com", username)
	password := "TestPass@123"
	
	// 1. 注册
	regData, _ := json.Marshal(map[string]string{"username": username, "email": email, "password": password})
	regResp, err := client.Post(apiBaseUrl+"/users/register", "application/json", bytes.NewBuffer(regData))
	if err != nil {
		t.Logf("注册请求警告 (可能已存在): %v", err)
	} else if regResp != nil {
		regResp.Body.Close()
	}
	
	// 2. 登录
	loginData, _ := json.Marshal(map[string]string{"email": email, "password": password})
	loginResp, err := client.Post(apiBaseUrl+"/users/login", "application/json", bytes.NewBuffer(loginData))
	if err != nil {
		t.Fatalf("测试预先登录失败，请求报错: %v", err)
	}
	if loginResp.StatusCode != 200 {
		var errData map[string]interface{}
		json.NewDecoder(loginResp.Body).Decode(&errData)
		t.Fatalf("登录返回非 200: %v, Body: %v", loginResp.StatusCode, errData)
	}
	
	var loginRes map[string]interface{}
	json.NewDecoder(loginResp.Body).Decode(&loginRes)
	
	var token string
	if data, ok := loginRes["data"].(map[string]interface{}); ok {
		if tVal, ok := data["token"].(string); ok {
			token = tVal
		}
	} else if tVal, ok := loginRes["token"].(string); ok {
		token = tVal
	}
	
	if token == "" {
		t.Fatalf("未能从登录响应中提取到 Token")
	}
	
	t.Logf("获取 Token 成功: %s...", token[:10])

	// -------- 核心修改：换成纯 HTTP POST 请求，完全模拟现在 ESP32 的行为 --------
    // 注意：替换为你后端真正接收该 JSON 的实际接口路径
    // 如果你为 ESP32 专属开了一个新接口，请填在这里；如果复用 v2/sensor-alert，请确保后端能解析这个嵌套结构
    apiURL = apiBaseUrl + "/api/sensor/message" 

    // 构造目前 ESP32 正在发送的完整 JSON 格式（包含 Heartbeat 结构或 BoardMessage 结构）
    testMessage := Board.BoardMessage{
        MessageID:   fmt.Sprintf("ALERT_%d", time.Now().Unix()),
        MessageTime: time.Now(),
        Message: Board.Message{
            MessageType: "Request",  // 注意跟你后端处理的类型匹配
            Command:     "SensorAlert",
            Attribute:   Board.MessageAttribute_Default,
            // 数据载荷对应 ESP32 发送的内容
            Data: map[string]interface{}{
                "sensor_id":   "esp32_sensor_001",
                "latitude":    39.9042,
                "longitude":   116.4074,
                "radius":      50.0,
                "photo_count": 5,
                "altitude":    100.0,
            },
        },
        FromID:   "esp32_sensor_001",
        FromType: "sensor", // 如果之前定义了具体的标识常量，或者强制为 "ESP32" 等
        ToID:     "cloud_backend", 
    }

    msgBytes, err := json.Marshal(testMessage)
    if err != nil {
        t.Fatalf("Failed to marshal BoardMessage: %v", err)
    }

    t.Logf("Sending pure HTTPS POST to %s...", apiURL)

    // 构造 HTTP POST 请求
    req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(msgBytes))
    if err != nil {
        t.Fatalf("Failed to create HTTP POST request: %v", err)
    }
    
    // 设置 Content-Type 和鉴权 Header (这里与 ESP32 的 addHeader 行为一致)
    req.Header.Set("Content-Type", "application/json")
    if token != "" {
        req.Header.Set("Authorization", "Bearer "+token) 
    }

    // 设定超时并发送（利用外部定义的客户端 `client`，具备你之前写好的 InsecureSkipVerify 等特性）
    client.Timeout = 10 * time.Second
    resp, err := client.Do(req)
    if err != nil {
        t.Fatalf("HTTPS POST failed (could be Cloudflare block, DNS, or server down): %v", err)
    }
    defer resp.Body.Close()

    // 读取后端返回的响应内容
    respBuffer := new(bytes.Buffer)
    respBuffer.ReadFrom(resp.Body)

    if resp.StatusCode != 200 {
        t.Fatalf("Server rejected the request. HTTP Status: %d, Response: %s", resp.StatusCode, respBuffer.String())
    }

    // 解析并验证返回结果
    var receivedResponse map[string]interface{}
    if err := json.Unmarshal(respBuffer.Bytes(), &receivedResponse); err != nil {
        t.Logf("Warning: Response is not standard JSON: %s", respBuffer.String())
    } else {
        t.Logf("Success! Received valid JSON response from backend: %+v", receivedResponse)
    }
}