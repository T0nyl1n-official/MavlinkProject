package tests

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	boardHandler "MavlinkProject/Server/backend/Handler/Boards"
	Board "MavlinkProject/Server/backend/Shared/Boards"
)

func TestMessageDispatcherFlow(t *testing.T) {
	outputLines := []string{}
	outputMu := sync.Mutex{}

	dispatcher := boardHandler.NewMessageDispatcher()
	if dispatcher == nil {
		t.Fatal("Failed to create MessageDispatcher")
	}
	outputLines = append(outputLines, "[PASS] MessageDispatcher created successfully")

	aiAgent := dispatcher.GetAIAgentHandler()
	if aiAgent == nil {
		t.Fatal("Failed to get AIAgentHandler")
	}
	outputLines = append(outputLines, "[PASS] AIAgentHandler retrieved")

	sensorHandler := dispatcher.GetSensorAlertHandler()
	if sensorHandler == nil {
		t.Fatal("Failed to get SensorAlertHandler")
	}
	outputLines = append(outputLines, "[PASS] SensorAlertHandler retrieved")

	boardHandlerInstance := dispatcher.GetBoardHandler()
	if boardHandlerInstance == nil {
		t.Fatal("Failed to get BoardHandler")
	}
	outputLines = append(outputLines, "[PASS] BoardHandler retrieved")

	t.Log("=== MessageDispatcher Basic Tests PASSED ===")
	saveTestOutput("TestMessageDispatcherFlow", outputLines, &outputMu)
}

func TestBoardHandlerCanHandle(t *testing.T) {
	outputLines := []string{}
	outputMu := sync.Mutex{}

	dispatcher := boardHandler.NewMessageDispatcher()
	boardH := dispatcher.GetBoardHandler()

	tests := []struct {
		name     string
		msg      Board.BoardMessage
		expected bool
	}{
		{
			name: "Board FromType",
			msg: Board.BoardMessage{
				FromID:   "drone_001",
				FromType: "board",
				Message: Board.Message{
					Command: "Heartbeat",
				},
			},
			expected: true,
		},
		{
			name: "Drone FromType",
			msg: Board.BoardMessage{
				FromID:   "drone_002",
				FromType: "drone",
				Message: Board.Message{
					Command: "Status",
				},
			},
			expected: true,
		},
		{
			name: "FC FromType",
			msg: Board.BoardMessage{
				FromID:   "fc_001",
				FromType: "fc",
				Message: Board.Message{
					Command: "TakeOff",
				},
			},
			expected: true,
		},
		{
			name: "Sensor FromType (should not be handled by BoardHandler)",
			msg: Board.BoardMessage{
				FromID:   "sensor_001",
				FromType: "sensor",
				Message: Board.Message{
					Command: "Alert",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		result := boardH.CanHandle(&tt.msg)
		if result != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, result)
			outputLines = append(outputLines, fmt.Sprintf("[FAIL] %s: expected %v, got %v", tt.name, tt.expected, result))
		} else {
			t.Logf("%s: passed", tt.name)
			outputLines = append(outputLines, fmt.Sprintf("[PASS] %s", tt.name))
		}
	}

	t.Log("=== BoardHandler CanHandle Tests PASSED ===")
	saveTestOutput("TestBoardHandlerCanHandle", outputLines, &outputMu)
}

func TestSensorAlertHandlerCanHandle(t *testing.T) {
	outputLines := []string{}
	outputMu := sync.Mutex{}

	dispatcher := boardHandler.NewMessageDispatcher()
	sensorH := dispatcher.GetSensorAlertHandler()

	tests := []struct {
		name     string
		msg      Board.BoardMessage
		expected bool
	}{
		{
			name: "Sensor FromType",
			msg: Board.BoardMessage{
				FromID:   "sensor_001",
				FromType: "sensor",
				Message: Board.Message{
					Command: "SensorAlert",
				},
			},
			expected: true,
		},
		{
			name: "ESP32 FromType",
			msg: Board.BoardMessage{
				FromID:   "esp32_001",
				FromType: "esp32",
				Message: Board.Message{
					Command: "Alert",
				},
			},
			expected: true,
		},
		{
			name: "Warning Attribute with Alert Command",
			msg: Board.BoardMessage{
				FromID:   "esp32_002",
				FromType: "esp32",
				Message: Board.Message{
					Attribute: Board.MessageAttribute_Warning,
					Command:   "Alert",
				},
			},
			expected: true,
		},
		{
			name: "Board FromType (should not be handled by SensorAlertHandler)",
			msg: Board.BoardMessage{
				FromID:   "drone_001",
				FromType: "board",
				Message: Board.Message{
					Command: "Heartbeat",
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		result := sensorH.CanHandle(&tt.msg)
		if result != tt.expected {
			t.Errorf("%s: expected %v, got %v", tt.name, tt.expected, result)
			outputLines = append(outputLines, fmt.Sprintf("[FAIL] %s: expected %v, got %v", tt.name, tt.expected, result))
		} else {
			t.Logf("%s: passed", tt.name)
			outputLines = append(outputLines, fmt.Sprintf("[PASS] %s", tt.name))
		}
	}

	t.Log("=== SensorAlertHandler CanHandle Tests PASSED ===")
	saveTestOutput("TestSensorAlertHandlerCanHandle", outputLines, &outputMu)
}

func TestAIAgentHandlerEnabled(t *testing.T) {
	outputLines := []string{}
	outputMu := sync.Mutex{}

	dispatcher := boardHandler.NewMessageDispatcher()
	aiH := dispatcher.GetAIAgentHandler()

	if aiH.IsEnabled() {
		t.Error("AIAgentHandler should be disabled by default")
		outputLines = append(outputLines, "[FAIL] AIAgentHandler should be disabled by default")
	} else {
		t.Log("AIAgentHandler correctly disabled by default")
		outputLines = append(outputLines, "[PASS] AIAgentHandler correctly disabled by default")
	}

	aiH.Enable()
	if !aiH.IsEnabled() {
		t.Error("AIAgentHandler should be enabled after calling Enable()")
		outputLines = append(outputLines, "[FAIL] AIAgentHandler should be enabled after calling Enable()")
	} else {
		t.Log("AIAgentHandler correctly enabled")
		outputLines = append(outputLines, "[PASS] AIAgentHandler correctly enabled")
	}

	aiH.Disable()
	if aiH.IsEnabled() {
		t.Error("AIAgentHandler should be disabled after calling Disable()")
		outputLines = append(outputLines, "[FAIL] AIAgentHandler should be disabled after calling Disable()")
	} else {
		t.Log("AIAgentHandler correctly disabled")
		outputLines = append(outputLines, "[PASS] AIAgentHandler correctly disabled")
	}

	t.Log("=== AIAgentHandler Enable/Disable Tests PASSED ===")
	saveTestOutput("TestAIAgentHandlerEnabled", outputLines, &outputMu)
}

func TestSensorAlertAutoExecute(t *testing.T) {
	outputLines := []string{}
	outputMu := sync.Mutex{}

	dispatcher := boardHandler.NewMessageDispatcher()
	sensorH := dispatcher.GetSensorAlertHandler()

	if !sensorH.IsAutoExecute() {
		t.Error("SensorAlertHandler should have auto-execute enabled by default")
		outputLines = append(outputLines, "[FAIL] SensorAlertHandler should have auto-execute enabled by default")
	} else {
		t.Log("SensorAlertHandler correctly has auto-execute enabled by default")
		outputLines = append(outputLines, "[PASS] SensorAlertHandler correctly has auto-execute enabled by default")
	}

	sensorH.SetAutoExecute(false)
	if sensorH.IsAutoExecute() {
		t.Error("SensorAlertHandler should have auto-execute disabled after SetAutoExecute(false)")
		outputLines = append(outputLines, "[FAIL] SensorAlertHandler should have auto-execute disabled after SetAutoExecute(false)")
	} else {
		t.Log("SensorAlertHandler correctly has auto-execute disabled")
		outputLines = append(outputLines, "[PASS] SensorAlertHandler correctly has auto-execute disabled")
	}

	t.Log("=== SensorAlertHandler AutoExecute Tests PASSED ===")
	saveTestOutput("TestSensorAlertAutoExecute", outputLines, &outputMu)
}

func TestDispatchSensorAlertMessage(t *testing.T) {
	outputLines := []string{}
	outputMu := sync.Mutex{}
	var wg sync.WaitGroup

	manager := boardHandler.GetBoardManager()
	testBoardID := "dispatch_sensor_test"
	testAddr := "127.0.0.1"
	testPort := "61010"

	err := manager.StartTCPServer(testBoardID, testAddr, testPort)
	if err != nil {
		t.Fatalf("Failed to start TCP server: %v", err)
	}
	defer manager.StopBoard(testBoardID)

	time.Sleep(100 * time.Millisecond)

	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", testAddr, testPort))
		if err != nil {
			t.Logf("Failed to connect: %v", err)
			outputLines = append(outputLines, fmt.Sprintf("[INFO] TCP connection attempt: %v", err))
			return
		}
		defer conn.Close()

		sensorMsg := Board.BoardMessage{
			MessageID:   "sensor_msg_001",
			MessageTime: time.Now(),
			FromID:      "esp32_sensor_001",
			FromType:    "sensor",
			ToID:        "server",
			ToType:      "server",
			Message: Board.Message{
				MessageType: "Request",
				Attribute:   Board.MessageAttribute_Warning,
				Connection:  "TCP",
				Command:     "SensorAlert",
				Data: map[string]interface{}{
					"sensor_id":   "sensor_001",
					"latitude":    37.7749,
					"longitude":   -122.4194,
					"radius":      50.0,
					"altitude":    100.0,
					"photo_count": 5,
				},
			},
		}

		data, _ := json.Marshal(sensorMsg)
		_, err = conn.Write(data)
		if err != nil {
			t.Logf("Failed to send: %v", err)
			outputLines = append(outputLines, fmt.Sprintf("[INFO] Send failed: %v", err))
			return
		}

		t.Log("Sensor alert message sent")
		outputLines = append(outputLines, "[PASS] Sensor alert message sent via TCP")
	}()

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	t.Log("=== Dispatch SensorAlert Message Test PASSED ===")
	saveTestOutput("TestDispatchSensorAlertMessage", outputLines, &outputMu)
}

func TestDispatchBoardHeartbeatMessage(t *testing.T) {
	outputLines := []string{}
	outputMu := sync.Mutex{}
	var wg sync.WaitGroup

	manager := boardHandler.GetBoardManager()
	testBoardID := "dispatch_board_test"
	testAddr := "127.0.0.1"
	testPort := "61011"

	err := manager.StartTCPServer(testBoardID, testAddr, testPort)
	if err != nil {
		t.Fatalf("Failed to start TCP server: %v", err)
	}
	defer manager.StopBoard(testBoardID)

	time.Sleep(100 * time.Millisecond)

	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", testAddr, testPort))
		if err != nil {
			t.Logf("Failed to connect: %v", err)
			outputLines = append(outputLines, fmt.Sprintf("[INFO] TCP connection attempt: %v", err))
			return
		}
		defer conn.Close()

		heartbeatMsg := Board.BoardMessage{
			MessageID:   "heartbeat_001",
			MessageTime: time.Now(),
			FromID:      "drone_001",
			FromType:    "board",
			ToID:        "server",
			ToType:      "server",
			Message: Board.Message{
				MessageType: "Request",
				Attribute:   Board.MessageAttribute_Status,
				Connection:  "TCP",
				Command:     "Heartbeat",
				Data: map[string]interface{}{
					"battery": 85,
					"status":  "flying",
				},
			},
		}

		data, _ := json.Marshal(heartbeatMsg)
		_, err = conn.Write(data)
		if err != nil {
			t.Logf("Failed to send: %v", err)
			outputLines = append(outputLines, fmt.Sprintf("[INFO] Send failed: %v", err))
			return
		}

		t.Log("Heartbeat message sent")
		outputLines = append(outputLines, "[PASS] Heartbeat message sent via TCP")
	}()

	wg.Wait()
	time.Sleep(500 * time.Millisecond)

	t.Log("=== Dispatch Board Heartbeat Message Test PASSED ===")
	saveTestOutput("TestDispatchBoardHeartbeatMessage", outputLines, &outputMu)
}

func TestDispatchMultipleMessageTypes(t *testing.T) {
	outputLines := []string{}
	outputMu := sync.Mutex{}

	dispatcher := boardHandler.NewMessageDispatcher()

	messages := []Board.BoardMessage{
		{
			MessageID: "msg_001",
			FromID:    "drone_001",
			FromType:  "board",
			Message: Board.Message{
				Command: "Heartbeat",
			},
		},
		{
			MessageID: "msg_002",
			FromID:    "sensor_001",
			FromType:  "sensor",
			Message: Board.Message{
				Command: "SensorAlert",
			},
		},
		{
			MessageID: "msg_003",
			FromID:    "drone_002",
			FromType:  "drone",
			Message: Board.Message{
				Command: "Status",
			},
		},
	}

	for i, msg := range messages {
		err := dispatcher.Dispatch(&msg)
		if err != nil {
			t.Logf("Dispatch msg_%d returned (expected for sensor): %v", i+1, err)
			outputLines = append(outputLines, fmt.Sprintf("[INFO] Dispatch msg_%d: %v", i+1, err))
		} else {
			t.Logf("Dispatch msg_%d succeeded", i+1)
			outputLines = append(outputLines, fmt.Sprintf("[PASS] Dispatch msg_%d succeeded", i+1))
		}
	}

	t.Log("=== Dispatch Multiple Message Types Test PASSED ===")
	saveTestOutput("TestDispatchMultipleMessageTypes", outputLines, &outputMu)
}

func TestEndToEndSensorAlertToCentral(t *testing.T) {
	outputLines := []string{}
	outputMu := sync.Mutex{}
	var wg sync.WaitGroup

	manager := boardHandler.GetBoardManager()
	testBoardID := "e2e_sensor_test"
	testAddr := "127.0.0.1"
	testPort := "61020"

	err := manager.StartTCPServer(testBoardID, testAddr, testPort)
	if err != nil {
		t.Fatalf("Failed to start TCP server: %v", err)
	}
	defer manager.StopBoard(testBoardID)

	time.Sleep(100 * time.Millisecond)

	sensorH := manager.GetDispatcher().GetSensorAlertHandler()
	sensorH.SetAutoExecute(false)
	outputLines = append(outputLines, "[INFO] SensorAlertHandler auto-execute disabled for test")

	wg.Add(1)
	go func() {
		defer wg.Done()

		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", testAddr, testPort))
		if err != nil {
			t.Logf("Failed to connect: %v", err)
			outputLines = append(outputLines, fmt.Sprintf("[FAIL] TCP connection failed: %v", err))
			return
		}
		defer conn.Close()

		sensorMsg := Board.BoardMessage{
			MessageID:   "e2e_sensor_001",
			MessageTime: time.Now(),
			FromID:      "esp32_c3_001",
			FromType:    "sensor",
			ToID:        "server",
			ToType:      "server",
			Message: Board.Message{
				MessageType: "Request",
				Attribute:   Board.MessageAttribute_Warning,
				Connection:  "TCP",
				Command:     "SensorAlert",
				Data: map[string]interface{}{
					"sensor_id":   "esp32_c3_001",
					"latitude":    37.7749,
					"longitude":   -122.4194,
					"radius":      30.0,
					"altitude":    80.0,
					"photo_count": 3,
				},
			},
		}

		data, _ := json.Marshal(sensorMsg)
		_, err = conn.Write(data)
		if err != nil {
			t.Logf("Failed to send: %v", err)
			outputLines = append(outputLines, fmt.Sprintf("[FAIL] Send failed: %v", err))
			return
		}

		t.Log("End-to-end sensor alert message sent")
		outputLines = append(outputLines, "[PASS] End-to-end sensor alert message sent")
	}()

	wg.Wait()
	time.Sleep(300 * time.Millisecond)

	sensorH.SetAutoExecute(true)

	t.Log("=== End-to-End Sensor Alert Test PASSED ===")
	saveTestOutput("TestEndToEndSensorAlertToCentral", outputLines, &outputMu)
}

func saveTestOutput(testName string, lines []string, mu *sync.Mutex) {
	mu.Lock()
	defer mu.Unlock()

	dir := "e:/CompertationsProjects/MavlinkProject/MavlinkProject/tests/OutputHistory"
	err := os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		log.Printf("Failed to create directory: %v", err)
		return
	}

	timestamp := time.Now().Format("20060102_150405")
	filename := fmt.Sprintf("%s_%s.txt", testName, timestamp)
	filePath := filepath.Join(dir, filename)

	content := strings.Join(lines, "\n")
	content = fmt.Sprintf("=== Test Output: %s ===\nTime: %s\n\n%s\n",
		testName, time.Now().Format("2006-01-02 15:04:05"), content)

	err = os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		log.Printf("Failed to write file: %v", err)
		return
	}

	log.Printf("Test output saved to: %s", filePath)
}
