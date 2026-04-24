package sensorHandler

import (
	"fmt"
	"time"

	Board "MavlinkProject/Server/backend/Shared/Boards"
	FRP "MavlinkProject/Server/backend/Utils/FRPHelper"
)

type SensorAlertReq struct {
	SensorID   string  `json:"sensor_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Radius     float64 `json:"radius"`
	PhotoCount int     `json:"photo_count"`
	Altitude   float64 `json:"altitude"`
}

func GenerateChainAndSendToCentral(req SensorAlertReq) error {
	centrals := FRP.GetFRPCentrals()
	if len(centrals) == 0 {
		return fmt.Errorf("no FRP central servers configured")
	}

	chainID := fmt.Sprintf("chain_sensor_%v", time.Now().Unix())

	radius := req.Radius
	if radius <= 0 {
		radius = 50
	}

	altitude := req.Altitude
	if altitude <= 0 {
		altitude = 100
	}

	photoCount := req.PhotoCount
	if photoCount <= 0 {
		photoCount = 10
	}

	tasks := []Board.CentralTask{
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
			TaskID:  "task_1_takeoff",
			Command: "TakeOff",
			Data: map[string]interface{}{
				"altitude": altitude,
			},
			Status: "pending",
		},
		{
			TaskID:  "task_2_goto_sensor",
			Command: "GoTo",
			Data: map[string]interface{}{
				"latitude":  req.Latitude,
				"longitude": req.Longitude,
				"altitude":  altitude,
				"delay":     5.0, // 给予一定时间到达位置并悬停
			},
			Status: "pending",
		},
	}

	for i := 0; i < photoCount; i++ {
		tasks = append(tasks, Board.CentralTask{
			TaskID:  fmt.Sprintf("task_%d_photo", 3+i),
			Command: "TakePhoto",
			Data: map[string]interface{}{
				"delay": 5.0, // 给予树莓派充足的调用摄像头和网络上传时间
			},
			Status:  "pending",
		})
	}

	tasks = append(tasks, Board.CentralTask{
		TaskID:  "task_last_rtl",
		Command: "SetMode", // 切换为返航模式 RTL
		Data: map[string]interface{}{
			"mode": "RTL", // 注意要在 central.go 中的 executeSetMode 处理 RTL 模式
		},
		Status: "pending",
	})

	centralChain := Board.CentralProgressChain{
		ChainID: chainID,
		Tasks:   tasks,
		Status:  "pending",
	}

	boardMsg := Board.BoardMessage{
		MessageID:   fmt.Sprintf("msg_%d", time.Now().UnixNano()),
		MessageTime: time.Now(),
		FromID:      "cloud_backend",
		FromType:    "server",
		ToID:        "central_board",
		ToType:      "server",
		Message: Board.Message{
			MessageType: "Request",
			Attribute:   Board.MessageAttribute_Mission,
			Connection:  "TCP",
			Command:     "schedule_chain",
			Data: map[string]interface{}{
				"progress_chain": centralChain,
			},
		},
	}

	var lastErr error
	for _, central := range centrals {
		frpAddr := fmt.Sprintf("%s:%d", central.Address, central.Port)
		response, err := FRP.PushMessageToCentral(frpAddr, central.Timeout, central.MaxRetryAttempts, &boardMsg)
		if err != nil {
			lastErr = err
			continue
		}
		_ = response
		return nil
	}

	return fmt.Errorf("failed to push chain to central: %v", lastErr)
}
