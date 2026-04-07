package boardHandler

import (
	"fmt"
	"time"

	Board "MavlinkProject/Server/backend/Shared/Boards"
)

// SensorAlertReq 传感器上传警报请求 (例如ESP32)
type SensorAlertReq struct {
	SensorID   string  `json:"sensor_id"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	Radius     float64 `json:"radius"`
	PhotoCount int     `json:"photo_count"`
	Altitude   float64 `json:"altitude"`
}

// GenerateChainAndSendToCentral 接受传感器警报信息，生成任务链并行使向树莓派发送请求的过程
func GenerateChainAndSendToCentral(req SensorAlertReq, centralFrpAddress string) error {
	// 构建给树莓派的简单进度链结构
	chainID := fmt.Sprintf("chain_sensor_%v", time.Now().Unix())

	radius := req.Radius
	if radius <= 0 {
		radius = 50 // 默认值
	}

	altitude := req.Altitude
	if altitude <= 0 {
		altitude = 100 // 默认值
	}

	photoCount := req.PhotoCount
	if photoCount <= 0 {
		photoCount = 10
	}

	tasks := []Board.CentralTask{
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
			},
			Status: "pending",
		},
	}

	for i := 0; i < photoCount; i++ {
		tasks = append(tasks, Board.CentralTask{
			TaskID:  fmt.Sprintf("task_%d_photo", 3+i),
			Command: "TakePhoto",
			Data:    map[string]interface{}{},
			Status:  "pending",
		})
	}

	tasks = append(tasks, Board.CentralTask{
		TaskID:  "task_last_land",
		Command: "Land",
		Data:    map[string]interface{}{},
		Status:  "pending",
	})

	centralChain := Board.CentralProgressChain{
		ChainID: chainID,
		Tasks:   tasks,
		Status:  "pending",
	}

	// 构建下发的BoardMessage
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

	// 此时可以替换为你最终想设定的内网穿透 FRP 地址，目前默认由参数传入
	response, err := PushMessageToCentral(centralFrpAddress, &boardMsg)
	if err != nil {
		return fmt.Errorf("推送任务链至树莓派失败: %v", err)
	}
	_ = response // 得到回复后的处理可以扩展

	return nil
}
