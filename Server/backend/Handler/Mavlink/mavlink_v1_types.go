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
