package mavlink

import (
	"fmt"
	"log"
	"time"

	MavlinkBoard "MavlinkProject_Board/MavlinkCommand"
)

var commander *MavlinkBoard.MavlinkCommander

func InitMavlinkCommander(serialPort string, baud int, targetSystem uint8) error {
	commander = MavlinkBoard.NewMavlinkCommander()

	config := MavlinkBoard.MavlinkConfig{
		ConnectionType: MavlinkBoard.ConnectionSerial,
		SerialPort:     serialPort,
		SerialBaud:     baud,
		TargetSystem:   targetSystem,
		HeartbeatRate:  1,
	}

	commander.Configure(config)

	if err := commander.Start(); err != nil {
		return fmt.Errorf("failed to start MAVLink commander: %v", err)
	}

	log.Printf("[MAVLink] Initialized: %s @ %d baud", serialPort, baud)
	return nil
}

func TakeOff(data map[string]interface{}) error {
	altitude, ok := data["altitude"].(float64)
	if !ok {
		altitude = 10.0
	}

	log.Printf("[MAVLink] TakeOff to %.1f meters", altitude)
	return commander.CommandLong(
		commander.GetTargetSystem(), 1,
		22, // MAV_CMD_NAV_TAKEOFF
		0,
		0, 0, 0, 0, 0, 0,
		float32(altitude),
	)
}

func Land(data map[string]interface{}) error {
	lat, _ := data["latitude"].(float64)
	lon, _ := data["longitude"].(float64)
	alt, _ := data["altitude"].(float64)

	log.Printf("[MAVLink] Land at (%.6f, %.6f, %.1f)", lat, lon, alt)
	return commander.CommandLong(
		commander.GetTargetSystem(), 1,
		21, // MAV_CMD_NAV_LAND
		0,
		0, 0, 0, 0,
		float32(lat), float32(lon),
		float32(alt),
	)
}

func GoTo(data map[string]interface{}) error {
	lat, ok := data["latitude"].(float64)
	if !ok {
		return fmt.Errorf("missing latitude")
	}
	lon, ok := data["longitude"].(float64)
	if !ok {
		return fmt.Errorf("missing longitude")
	}
	alt, _ := data["altitude"].(float64)

	log.Printf("[MAVLink] GoTo (%.6f, %.6f, %.1f)", lat, lon, alt)
	return commander.CommandLong(
		commander.GetTargetSystem(), 1,
		16, // MAV_CMD_NAV_WAYPOINT
		0,
		0, 0, 0, 0,
		float32(lat), float32(lon),
		float32(alt),
	)
}

func ReturnToHome() error {
	log.Printf("[MAVLink] Return to home")
	return commander.CommandLong(
		commander.GetTargetSystem(), 1,
		20, // MAV_CMD_NAV_RETURN_TO_LAUNCH
		0,
		0, 0, 0, 0, 0, 0, 0,
	)
}

func Survey(data map[string]interface{}) error {
	lat, ok := data["latitude"].(float64)
	if !ok {
		return fmt.Errorf("missing latitude")
	}
	lon, ok := data["longitude"].(float64)
	if !ok {
		return fmt.Errorf("missing longitude")
	}
	radius, _ := data["radius"].(float64)
	duration, _ := data["duration"].(float64)

	log.Printf("[MAVLink] Survey area (%.6f, %.6f) radius %.1f for %.0f seconds", lat, lon, radius, duration)
	// 这里应该实现区域侦察逻辑
	return nil
}

func SurveyGrid(data map[string]interface{}) error {
	lat, ok := data["latitude"].(float64)
	if !ok {
		return fmt.Errorf("missing latitude")
	}
	lon, ok := data["longitude"].(float64)
	if !ok {
		return fmt.Errorf("missing longitude")
	}
	width, _ := data["width"].(float64)
	height, _ := data["height"].(float64)
	alt, _ := data["altitude"].(float64)

	log.Printf("[MAVLink] Survey grid (%.6f, %.6f) width %.1f height %.1f at %.1f meters", lat, lon, width, height, alt)
	// 这里应该实现网格搜索逻辑
	return nil
}

func Orbit(data map[string]interface{}) error {
	lat, ok := data["latitude"].(float64)
	if !ok {
		return fmt.Errorf("missing latitude")
	}
	lon, ok := data["longitude"].(float64)
	if !ok {
		return fmt.Errorf("missing longitude")
	}
	radius, _ := data["radius"].(float64)
	duration, _ := data["duration"].(float64)

	log.Printf("[MAVLink] Orbit around (%.6f, %.6f) radius %.1f for %.0f seconds", lat, lon, radius, duration)
	// 这里应该实现盘旋巡逻逻辑
	return nil
}

func TakePhoto() error {
	log.Printf("[MAVLink] Take photo")
	// 这里应该实现拍照逻辑
	return nil
}

func StartVideo() error {
	log.Printf("[MAVLink] Start video recording")
	// 这里应该实现开始录像逻辑
	return nil
}

func StopVideo() error {
	log.Printf("[MAVLink] Stop video recording")
	// 这里应该实现停止录像逻辑
	return nil
}

func SetMode(data map[string]interface{}) error {
	mode, ok := data["mode"].(string)
	if !ok {
		return fmt.Errorf("missing mode")
	}

	log.Printf("[MAVLink] Set mode: %s", mode)
	// 这里应该实现设置飞行模式逻辑
	return nil
}

func GetDroneStatus() (map[string]interface{}, error) {
	if commander == nil {
		return nil, fmt.Errorf("MAVLink commander not initialized")
	}

	// 这里应该从 MAVLink 获取实际的无人机状态
	// 由于 MavlinkCommander 可能没有直接的状态获取方法
	// 我们可以返回一些基本信息
	status := map[string]interface{}{
		"board_id":       "Error-NoDroneStatus",
		"system_id":      commander.GetTargetSystem(),
		"battery_level":  85.5, // 模拟数据，实际应该从 MAVLink 获取
		"latitude":       22.543123, // 模拟数据，实际应该从 MAVLink 获取
		"longitude":      114.052345, // 模拟数据，实际应该从 MAVLink 获取
		"altitude":       0.0, // 模拟数据，实际应该从 MAVLink 获取
		"is_armed":       false, // 模拟数据，实际应该从 MAVLink 获取
		"flight_mode":    "STABILIZE", // 模拟数据，实际应该从 MAVLink 获取
		"last_update":    time.Now().Unix(),
	}

	return status, nil
}
