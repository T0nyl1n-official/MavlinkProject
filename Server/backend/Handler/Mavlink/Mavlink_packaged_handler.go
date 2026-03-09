package Mavlink

import (
	Drones "MavlinkProject/Server/backend/Shared/Drones"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
)

type PackageHandler struct {
	handler      *MAVLinkHandlerV1
	chainManager *ChainManager
	drone        *Drones.Drone
	mu           sync.RWMutex
}

func NewPackageHandler(handler *MAVLinkHandlerV1) *PackageHandler {
	// 创建默认的无人机配置
	droneConfig := Drones.DroneConfig{
		SystemID:        1,
		ComponentID:     1,
		ProtocolVersion: "2.0",
		HeartbeatRate:   1 * time.Second,
		Timeout:         30 * time.Second,
	}

	return &PackageHandler{
		handler:      handler,
		chainManager: GetChainManager(),
		drone:        Drones.NewDrone("default", "Default Drone", "Generic", droneConfig),
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

func (p *PackageHandler) Takeoff(req TakeoffRequest) TakeoffResponse {
	p.mu.Lock()
	defer p.mu.Unlock()

	handlerID := p.handler.GetHandlerID()
	chain := p.chainManager.GetCurrentChain()
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

func (p *PackageHandler) Land(req LandRequest) LandResponse {
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

func (p *PackageHandler) Move(req MoveRequest) MoveResponse {
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

func (p *PackageHandler) GetStatus() StatusResponse {
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

func (p *PackageHandler) SetGroundStation(name, id string, lat, lon, alt float64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	position := Drones.Position{
		Latitude:  lat,
		Longitude: lon,
		Altitude:  alt,
	}

	p.handler.SetGroundStation(name, id, lat, lon, alt)
}

func (p *PackageHandler) SetAIAsDispatcher() {
	p.mu.Lock()
	defer p.mu.Unlock()
	// v1 handler 不包含调度器功能，跳过
}

func (p *PackageHandler) SetUserAsDispatcher(username, email string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	// v1 handler 不包含调度器功能，跳过
}

func (p *PackageHandler) GetHandler() *MAVLinkHandler {
	return p.handler
}

func (p *PackageHandler) GetGroundStationInfo() GroundStationInfoV1 {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if p.handler != nil {
		return p.handler.GetGroundStationInfo()
	}

	return GroundStationInfoV1{}
}

func (p *PackageHandler) GetChainManager() *ChainManager {
	return p.chainManager
}
