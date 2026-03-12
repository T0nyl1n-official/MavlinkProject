package Mavlink

import (
	"time"

	Drones "MavlinkProject/Server/backend/Shared/Drones"
)

// ==================== v1 专用类型定义 ====================

type ConnectionStatus string

const (
	ConnectionStatusDisconnected ConnectionStatus = "disconnected"
	ConnectionStatusConnecting   ConnectionStatus = "connecting"
	ConnectionStatusConnected    ConnectionStatus = "connected"
	ConnectionStatusError        ConnectionStatus = "error"
)

type FlightMode string

const (
	FlightModeManual    FlightMode = "manual"
	FlightModeStabilize FlightMode = "stabilize"
	FlightModeAuto      FlightMode = "auto"
	FlightModeGuided    FlightMode = "guided"
	FlightModeRTL       FlightMode = "rtl"
	FlightModeLand      FlightMode = "land"
)

// v1 专用连接类型
type ConnectionType string

const (
	ConnectionSerial ConnectionType = "serial"
	ConnectionUDP    ConnectionType = "udp"
	ConnectionTCP    ConnectionType = "tcp"
)

// v1 专用协议版本
type ProtocolVersion string

const (
	ProtocolVersionV1 ProtocolVersion = "1.0"
	ProtocolVersionV2 ProtocolVersion = "2.0"
)

// v1 专用配置结构
type MAVLinkConfigV1 struct {
	ConnectionType  ConnectionType
	SerialPort      string
	SerialBaud      int
	UDPAddr         string
	UDPPort         int
	TCPAddr         string
	TCPPort         int
	SystemID        int
	ComponentID     int
	ProtocolVersion ProtocolVersion
	HeartbeatRate   time.Duration
}

// v1 专用地面站信息
type GroundStationInfoV1 struct {
	Name     string          `json:"name"`
	ID       string          `json:"id"`
	Position Drones.Position `json:"position"`
}

// v1 专用入站消息
type IncomingMessageV1 struct {
	SystemID    int
	ComponentID int
	MessageID   int
	Message     interface{}
	Timestamp   time.Time
}

// v1 专用心跳包数据
type HeartbeatDataV1 struct {
	SystemID     int
	ComponentID  int
	Type         uint64
	Autopilot    uint64
	BaseMode     uint64
	CustomMode   uint32
	SystemStatus uint64
}

// v1 专用位置数据
type GPSDataV1 struct {
	SystemID    int
	ComponentID int
	Latitude    float64
	Longitude   float64
	Altitude    float64
	Timestamp   time.Time
}

// v1 专用姿态数据
type AttitudeDataV1 struct {
	SystemID    int
	ComponentID int
	Roll        float32
	Pitch       float32
	Yaw         float32
	RollSpeed   float32
	PitchSpeed  float32
	YawSpeed    float32
	Timestamp   time.Time
}

// v1 专用电池数据
type BatteryDataV1 struct {
	SystemID    int
	ComponentID int
	Voltage     float32
	Current     float32
	Remaining   int
	Temperature float32
	Timestamp   time.Time
}

// =============================================================================
// 类型转换函数 - 用于 HandlerConfig 数据交换
// =============================================================================

// HandlerConfigData 用于链配置传递的中间数据结构
// 不引用 ProgressChain 包以避免循环依赖
type HandlerConfigData struct {
	ConnectionType  string
	SerialPort      string
	SerialBaud      int
	UDPAddr         string
	UDPPort         int
	TCPAddr         string
	TCPPort         int
	SystemID        int
	ComponentID     int
	ProtocolVersion string
	HeartbeatRate   time.Duration
}

// ToConfigV1 将 HandlerConfigData 转换为 MAVLinkConfigV1
func (hc *HandlerConfigData) ToConfigV1() *MAVLinkConfigV1 {
	if hc == nil {
		return &MAVLinkConfigV1{}
	}

	return &MAVLinkConfigV1{
		ConnectionType:  ConnectionType(hc.ConnectionType),
		SerialPort:      hc.SerialPort,
		SerialBaud:      hc.SerialBaud,
		UDPAddr:         hc.UDPAddr,
		UDPPort:         hc.UDPPort,
		TCPAddr:         hc.TCPAddr,
		TCPPort:         hc.TCPPort,
		SystemID:        hc.SystemID,
		ComponentID:     hc.ComponentID,
		ProtocolVersion: ProtocolVersion(hc.ProtocolVersion),
		HeartbeatRate:   hc.HeartbeatRate,
	}
}

// ToHandlerConfigData 将 MAVLinkConfigV1 转换为 HandlerConfigData
func (v1Config *MAVLinkConfigV1) ToHandlerConfigData() *HandlerConfigData {
	if v1Config == nil {
		return &HandlerConfigData{}
	}

	return &HandlerConfigData{
		ConnectionType:  string(v1Config.ConnectionType),
		SerialPort:      v1Config.SerialPort,
		SerialBaud:      v1Config.SerialBaud,
		UDPAddr:         v1Config.UDPAddr,
		UDPPort:         v1Config.UDPPort,
		TCPAddr:         v1Config.TCPAddr,
		TCPPort:         v1Config.TCPPort,
		SystemID:        v1Config.SystemID,
		ComponentID:     v1Config.ComponentID,
		ProtocolVersion: string(v1Config.ProtocolVersion),
		HeartbeatRate:   v1Config.HeartbeatRate,
	}
}
