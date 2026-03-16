package Mavlink

import (
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	Charging "MavlinkProject/Server/backend/Shared/Charge"

	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
)

/*
	前名为 Mavlink_Packaged_Handler, 其主旨为分装V1API内较为原子化的操作, 以便于人工调度或者人工介入操作. 现为方便记忆, 改名为 Mavlink_Handler_v2

	MavlinkHandler ver2的理念是因为ver1的API太过于原子化，不利于人工编排, 而ver2的API则是为了方便人工编排而设计的, 将一些常见的情况和需求连续的包装起来的API

*/

type MavlinkHandlerV2 struct {
	handler      *MAVLinkHandlerV1
	chainManager *ChainManager
	mu           sync.RWMutex
}

func (h *MavlinkHandlerV2) New(handler *MAVLinkHandlerV1) *MavlinkHandlerV2 {
	if handler == nil {
		fmt.Errorf("MavlinkHandlerV2 - ERROR: ver1 handler is nil")
		return nil
	}
	return &MavlinkHandlerV2{
		handler:      handler,
		chainManager: GetChainManager(),
		mu:           sync.RWMutex{},
	}
}

type TakeoffRequest struct {
	Altitude float32 `json:"altitude" binding:"required"`
}

type TakeoffResponse struct {
	Success   bool   `json:"success"`
	HandlerID string `json:"handler_id"`
	ChainID   string `json:"chain_id"`
	Message   string `json:"message"`
}

type LandRequest struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	Altitude  float64 `json:"altitude,omitempty"`
}

type LandResponse struct {
	Success   bool   `json:"success"`
	HandlerID string `json:"handler_id"`
	ChainID   string `json:"chain_id"`
	Message   string `json:"message"`
}

type MoveRequest struct {
	Latitude  float64 `json:"latitude" binding:"required"`
	Longitude float64 `json:"longitude" binding:"required"`
	Altitude  float64 `json:"altitude" binding:"required"`
	Speed     float32 `json:"speed,omitempty"`
}

type MoveResponse struct {
	Success   bool   `json:"success"`
	HandlerID string `json:"handler_id"`
	ChainID   string `json:"chain_id"`
	Message   string `json:"message"`
}

type StatusResponse struct {
	Success   bool                   `json:"success"`
	HandlerID string                 `json:"handler_id"`
	ChainID   string                 `json:"chain_id"`
	Status    map[string]interface{} `json:"status"`
}

func (p *MavlinkHandlerV2) Takeoff(req TakeoffRequest) TakeoffResponse {
	p.mu.Lock()
	defer p.mu.Unlock()

	handlerID := p.handler.GetHandlerID()
	chain := p.chainManager.GetCurrentChain()
	// 鲁棒性纠错
	if chain == nil {
		chainID := p.chainManager.CreateChain()
		return TakeoffResponse{
			Success:   false,
			HandlerID: handlerID,
			ChainID:   chainID,
			Message:   "创建了新调度链，请重新尝试",
		}
	}
	if chain.IsFull() {
		chainID := p.chainManager.CreateNewChainAndSwitch()
		return TakeoffResponse{
			Success:   false,
			HandlerID: handlerID,
			ChainID:   chainID,
			Message:   "调度链已满，创建了新链，请重新尝试",
		}
	}

	// 发送起飞命令
	msg := &common.MessageCommandLong{
		Command:      common.MAV_CMD_NAV_TAKEOFF,
		Param1:       0,
		Param2:       0,
		Param3:       0,
		Param4:       0,
		Param5:       0,
		Param6:       0,
		Param7:       req.Altitude,
		Confirmation: 0,
	}

	err := p.handler.SendMessage(msg)

	// 记录调度
	params, _ := json.Marshal(req)
	var result string
	if err != nil {
		result = fmt.Sprintf("起飞失败: %s", err.Error())
	} else {
		result = "起飞命令已发送"
	}

	p.chainManager.AddRecordToCurrentChain(
		handlerID,
		"/mavlink/v2/api/takeoff",
		string(params),
		result,
		err == nil,
	)

	return TakeoffResponse{
		Success:   err == nil,
		HandlerID: handlerID,
		ChainID:   chain.ChainID,
		Message:   result,
	}
}

func (p *MavlinkHandlerV2) Land(req LandRequest) LandResponse {
	p.mu.Lock()
	defer p.mu.Unlock()

	handlerID := p.handler.GetHandlerID()
	chain := p.chainManager.GetCurrentChain()
	if chain == nil {
		chainID := p.chainManager.CreateChain()
		return LandResponse{
			Success:   false,
			HandlerID: handlerID,
			ChainID:   chainID,
			Message:   "创建了新调度链，请重新尝试",
		}
	}

	if chain.IsFull() {
		chainID := p.chainManager.CreateNewChainAndSwitch()
		return LandResponse{
			Success:   false,
			HandlerID: handlerID,
			ChainID:   chainID,
			Message:   "调度链已满，创建了新链，请重新尝试",
		}
	}

	// 发送降落命令
	msg := &common.MessageCommandLong{
		Command:      common.MAV_CMD_NAV_LAND,
		Param1:       0,
		Param2:       0,
		Param3:       0,
		Param4:       0,
		Param5:       float32(req.Latitude),
		Param6:       float32(req.Longitude),
		Param7:       float32(req.Altitude),
		Confirmation: 0,
	}

	err := p.handler.SendMessage(msg)

	// 记录调度
	params, _ := json.Marshal(req)
	var result string
	if err != nil {
		result = fmt.Sprintf("降落失败: %s", err.Error())
	} else {
		result = "降落命令已发送"
	}

	p.chainManager.AddRecordToCurrentChain(
		handlerID,
		"/mavlink/v2/api/land",
		string(params),
		result,
		err == nil,
	)

	return LandResponse{
		Success:   err == nil,
		HandlerID: handlerID,
		ChainID:   chain.ChainID,
		Message:   result,
	}
}

func (p *MavlinkHandlerV2) Move(req MoveRequest) MoveResponse {
	p.mu.Lock()
	defer p.mu.Unlock()

	handlerID := p.handler.GetHandlerID()
	chain := p.chainManager.GetCurrentChain()
	if chain == nil {
		chainID := p.chainManager.CreateChain()
		return MoveResponse{
			Success:   false,
			HandlerID: handlerID,
			ChainID:   chainID,
			Message:   "创建了新调度链，请重新尝试",
		}
	}

	if chain.IsFull() {
		chainID := p.chainManager.CreateNewChainAndSwitch()
		return MoveResponse{
			Success:   false,
			HandlerID: handlerID,
			ChainID:   chainID,
			Message:   "调度链已满，创建了新链，请重新尝试",
		}
	}

	// 发送移动命令
	msg := &common.MessageSetPositionTargetGlobalInt{
		TimeBootMs:      uint32(time.Now().UnixMilli()),
		CoordinateFrame: common.MAV_FRAME_GLOBAL_RELATIVE_ALT,
		TypeMask:        0b0000111111111000, // 忽略速度，使用位置控制
		LatInt:          int32(req.Latitude * 1e7),
		LonInt:          int32(req.Longitude * 1e7),
		Alt:             float32(req.Altitude),
		Vx:              0,
		Vy:              0,
		Vz:              0,
		Afx:             0,
		Afy:             0,
		Afz:             0,
		Yaw:             0,
		YawRate:         0,
	}

	err := p.handler.SendMessage(msg)

	// 记录调度
	params, _ := json.Marshal(req)
	var result string
	if err != nil {
		result = fmt.Sprintf("移动失败: %s", err.Error())
	} else {
		result = "移动命令已发送"
	}

	p.chainManager.AddRecordToCurrentChain(
		handlerID,
		"/mavlink/v2/api/move",
		string(params),
		result,
		err == nil,
	)

	return MoveResponse{
		Success:   err == nil,
		HandlerID: handlerID,
		ChainID:   chain.ChainID,
		Message:   result,
	}
}

func (p *MavlinkHandlerV2) GetStatus() StatusResponse {
	p.mu.RLock()
	defer p.mu.RUnlock()

	handlerID := p.handler.GetHandlerID()
	chain := p.chainManager.GetCurrentChain()

	status := map[string]interface{}{
		"connected":      p.handler.GetConnectionStatus() == "connected",
		"drone_status":   p.handler.GetDroneStatus(),
		"position":       p.handler.GetDronePosition(),
		"attitude":       p.handler.GetDroneAttitude(),
		"battery":        p.handler.GetDroneBattery(),
		"ground_station": p.handler.GetGroundStation(),
	}

	var chainID string
	if chain != nil {
		chainID = chain.ChainID
		status["chain_info"] = chain.GetInfo()
	}

	return StatusResponse{
		Success:   true,
		HandlerID: handlerID,
		ChainID:   chainID,
		Status:    status,
	}
}

// SetGroundStation - 设置当前无人机的地面站信息
func (p *MavlinkHandlerV2) SetGroundStation(name, id string, lat, lon, alt float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.handler.SetGroundStation(name, id, lat, lon, alt)
}

func (p *MavlinkHandlerV2) GetHandler() *MAVLinkHandlerV1 {
	return p.handler
}

func (p *MavlinkHandlerV2) GetGroundStationInfo() GroundStationInfoV1 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.handler != nil {
		return p.handler.GetGroundStationInfo()
	}

	return GroundStationInfoV1{}
}

func (p *MavlinkHandlerV2) GetChainManager() *ChainManager {
	return p.chainManager
}

// =============================================================================
// 传感器警报响应 - 调度无人机到警报位置周围待命并拍照
// =============================================================================

type SensorAlertRequest struct {
	Latitude     float64 `json:"latitude" binding:"required"`
	Longitude    float64 `json:"longitude" binding:"required"`
	Altitude     float64 `json:"altitude"`
	Radius       float64 `json:"radius"`
	PhotoCount   int     `json:"photo_count"`
	AlertType    string  `json:"alert_type"`
	AlertMessage string  `json:"alert_message"`
}

type SensorAlertResponse struct {
	Success     bool                   `json:"success"`
	HandlerID   string                 `json:"handler_id"`
	ChainID     string                 `json:"chain_id"`
	Message     string                 `json:"message"`
	PhotosTaken int                    `json:"photos_taken"`
	FinalPos    map[string]interface{} `json:"final_position"`
	HasCamera   bool                   `json:"has_camera"`
}

// RespondToSensorAlert - 快速处理传感器警报请求, 使用指定的通用处理方法来处理传感器传来的警报(警报来自redis DB5)
func (p *MavlinkHandlerV2) RespondToSensorAlert(req SensorAlertRequest) SensorAlertResponse {
	p.mu.Lock()
	defer p.mu.Unlock()

	handlerID := p.handler.GetHandlerID()
	drone := p.handler.GetDrone()

	hasCamera := drone.HasCamera()
	photoCount := req.PhotoCount
	if photoCount <= 0 {
		photoCount = 10
	}

	radius := req.Radius
	if radius <= 0 {
		radius = 50
	}

	altitude := req.Altitude
	if altitude <= 0 {
		altitude = 100
	}

	chain := p.chainManager.GetCurrentChain()
	if chain == nil {
		chainID := p.chainManager.CreateChain()
		return SensorAlertResponse{
			Success:   false,
			HandlerID: handlerID,
			ChainID:   chainID,
			Message:   "创建了新调度链，请重新尝试",
		}
	}

	p.handler.Start()

	p.handler.SetFlightMode(FlightModeGuided)

	currentPos := p.handler.GetDronePosition()
	if currentPos.Latitude == 0 && currentPos.Longitude == 0 {
		return SensorAlertResponse{
			Success:   false,
			HandlerID: handlerID,
			ChainID:   chain.ChainID,
			Message:   "无法获取无人机当前位置",
		}
	}

	orbitPoints := calculateOrbitPoints(req.Latitude, req.Longitude, radius, 8)
	for i, point := range orbitPoints {
		moveReq := MoveRequest{
			Latitude:  point.Latitude,
			Longitude: point.Longitude,
			Altitude:  altitude,
			Speed:     5,
		}
		p.Move(moveReq)
		time.Sleep(2 * time.Second)

		if hasCamera {
			for j := 0; j < photoCount/8; j++ {
				if drone.CanTakePhoto() {
					drone.TakePhoto()
				}
			}
		}
		_ = i
	}

	holdReq := MoveRequest{
		Latitude:  req.Latitude,
		Longitude: req.Longitude,
		Altitude:  altitude,
		Speed:     0,
	}
	p.Move(holdReq)

	finalPos := p.handler.GetDronePosition()

	return SensorAlertResponse{
		Success:     true,
		HandlerID:   handlerID,
		ChainID:     chain.ChainID,
		Message:     fmt.Sprintf("已到达警报位置(%f, %f)周围待命", req.Latitude, req.Longitude),
		PhotosTaken: photoCount,
		FinalPos: map[string]interface{}{
			"latitude":  finalPos.Latitude,
			"longitude": finalPos.Longitude,
			"altitude":  finalPos.Altitude,
		},
		HasCamera: hasCamera,
	}
}

type OrbitPoint struct {
	Latitude  float64
	Longitude float64
}

func calculateOrbitPoints(centerLat, centerLon, radius float64, numPoints int) []OrbitPoint {
	points := make([]OrbitPoint, numPoints)
	earthRadius := 6371000.0

	for i := 0; i < numPoints; i++ {
		angle := float64(i) * (360.0 / float64(numPoints))
		angleRad := angle * math.Pi / 180.0

		latOffset := (radius / earthRadius) * 180.0 / math.Pi
		lonOffset := (radius / (earthRadius * math.Cos(centerLat*math.Pi/180.0))) * 180.0 / math.Pi

		points[i] = OrbitPoint{
			Latitude:  centerLat + latOffset*math.Sin(angleRad),
			Longitude: centerLon + lonOffset*math.Cos(angleRad),
		}
	}
	return points
}

// =============================================================================
// 返回充电地点 - 查找最近的空闲充电仓
// =============================================================================

type ReturnToChargeRequest struct {
	Priority       string `json:"priority"`
	ChargingCaseID string `json:"charging_case_id"`
}

type ReturnToChargeResponse struct {
	Success       bool                   `json:"success"`
	HandlerID     string                 `json:"handler_id"`
	ChainID       string                 `json:"chain_id"`
	Message       string                 `json:"message"`
	ChargingCase  map[string]interface{} `json:"charging_case"`
	RouteDistance float64                `json:"route_distance"`
}

// ReturnToCharge - 指定目前操控的无人机(从handler中获取)返回充电地点(从req中获取地点, 如果为nil, 则查找最近的空闲充电仓)
func (p *MavlinkHandlerV2) ReturnToCharge(req ReturnToChargeRequest) ReturnToChargeResponse {
	p.mu.Lock()
	defer p.mu.Unlock()

	handlerID := p.handler.GetHandlerID()
	drone := p.handler.GetDrone()

	chain := p.chainManager.GetCurrentChain()
	if chain == nil {
		chainID := p.chainManager.CreateChain()
		return ReturnToChargeResponse{
			Success:   false,
			HandlerID: handlerID,
			ChainID:   chainID,
			Message:   "创建了新调度链，请重新尝试",
		}
	}

	chargingManager := Charging.GetChargingManager()
	currentPos := p.handler.GetDronePosition()

	if currentPos.Latitude == 0 && currentPos.Longitude == 0 {
		return ReturnToChargeResponse{
			Success:   false,
			HandlerID: handlerID,
			ChainID:   chain.GetID(),
			Message:   "无法获取无人机当前位置",
		}
	}

	var targetCase *Charging.ChargingCase

	if req.ChargingCaseID != "" {
		specificCase := chargingManager.Get(req.ChargingCaseID)
		if specificCase == nil {
			return ReturnToChargeResponse{
				Success:   false,
				HandlerID: handlerID,
				ChainID:   chain.GetID(),
				Message:   fmt.Sprintf("指定的充电仓 %s 不存在", req.ChargingCaseID),
			}
		}

		if !specificCase.IsAvailable() {
			return ReturnToChargeResponse{
				Success:   false,
				HandlerID: handlerID,
				ChainID:   chain.GetID(),
				Message:   fmt.Sprintf("指定的充电仓 %s 当前不可用 (状态: %s)", req.ChargingCaseID, specificCase.Status),
			}
		}

		targetCase = specificCase
	} else {
		availableCases := chargingManager.GetAvailable()
		if len(availableCases) == 0 {
			return ReturnToChargeResponse{
				Success:   false,
				HandlerID: handlerID,
				ChainID:   chain.GetID(),
				Message:   "没有可用的充电仓",
			}
		}

		minDistance := float64(0)
		for _, c := range availableCases {
			dist := calculateDistance(currentPos.Latitude, currentPos.Longitude, c.Latitude, c.Longitude)
			if targetCase == nil || dist < minDistance {
				minDistance = dist
				targetCase = c
			}
		}
	}

	if targetCase == nil {
		return ReturnToChargeResponse{
			Success:   false,
			HandlerID: handlerID,
			ChainID:   chain.GetID(),
			Message:   "未找到合适的充电仓",
		}
	}

	if err := targetCase.Occupy(drone.GetID()); err != nil {
		return ReturnToChargeResponse{
			Success:   false,
			HandlerID: handlerID,
			ChainID:   chain.GetID(),
			Message:   fmt.Sprintf("占领充电仓失败: %v", err),
		}
	}

	routeDistance := calculateDistance(currentPos.Latitude, currentPos.Longitude, targetCase.Latitude, targetCase.Longitude)

	p.handler.SetFlightMode(FlightModeRTL)

	landReq := LandRequest{
		Latitude:  targetCase.Latitude,
		Longitude: targetCase.Longitude,
		Altitude:  targetCase.Altitude,
	}
	p.Land(landReq)

	return ReturnToChargeResponse{
		Success:       true,
		HandlerID:     handlerID,
		ChainID:       chain.GetID(),
		Message:       fmt.Sprintf("正在返回充电仓: %s", targetCase.Name),
		ChargingCase:  targetCase.GetInfo(),
		RouteDistance: routeDistance,
	}
}

// returnToCharge - 子方法, 计算无人机到充电仓的距离
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371000.0

	dLat := (lat2 - lat1) * math.Pi / 180.0
	dLon := (lon2 - lon1) * math.Pi / 180.0

	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180.0)*math.Cos(lat2*math.Pi/180.0)*
			math.Sin(dLon/2)*math.Sin(dLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}
