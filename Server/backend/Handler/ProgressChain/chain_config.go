package ProgressChain

import (
	"time"
)

type NodeStatus string

const (
	NodeStatusWaiting  NodeStatus = "waiting"
	NodeStatusRunning  NodeStatus = "running"
	NodeStatusFinished NodeStatus = "finished"
	NodeStatusError    NodeStatus = "error"
)

type NodeType string

const (
	NodeTypeCreateHandler     NodeType = "createHandler"
	NodeTypeDeleteHandler     NodeType = "deleteHandler"
	NodeTypeUpdateHandler     NodeType = "updateHandler"
	NodeTypeConnectionStart   NodeType = "connectionStart"
	NodeTypeConnectionStop    NodeType = "connectionStop"
	NodeTypeConnectionRestart NodeType = "connectionRestart"
	NodeTypeDroneStatus       NodeType = "droneStatus"
	NodeTypeDroneTakeoff      NodeType = "droneTakeoff"
	NodeTypeDroneLand         NodeType = "droneLand"
	NodeTypeDroneMove         NodeType = "droneMove"
	NodeTypeDroneReturn       NodeType = "droneReturn"
	NodeTypeDroneMode         NodeType = "droneMode"
	NodeTypeGroundStationSet  NodeType = "groundStationSet"
	NodeTypeTaskVerification  NodeType = "taskVerification"
	NodeTypeStreamRequest     NodeType = "streamRequest"
	NodeTypeHeartbeatSend     NodeType = "heartbeatSend"
)

type HandlerConfig struct {
	ConnectionType  string        `json:"connection_type"`
	SerialPort      string        `json:"serial_port,omitempty"`
	SerialBaud      int           `json:"serial_baud,omitempty"`
	UDPAddr         string        `json:"udp_addr,omitempty"`
	UDPPort         int           `json:"udp_port,omitempty"`
	TCPAddr         string        `json:"tcp_addr,omitempty"`
	TCPPort         int           `json:"tcp_port,omitempty"`
	SystemID        int           `json:"system_id"`
	ComponentID     int           `json:"component_id"`
	ProtocolVersion string        `json:"protocol_version"`
	HeartbeatRate   time.Duration `json:"heartbeat_rate,omitempty"`
}

type Node struct {
	ID            string                 `json:"id"`
	Type          NodeType               `json:"type"`
	Status        NodeStatus             `json:"status"`
	HandlerConfig *HandlerConfig         `json:"handler_config,omitempty"`
	Params        map[string]interface{} `json:"params,omitempty"`
	Result        string                 `json:"result,omitempty"`
	Error         string                 `json:"error,omitempty"`
	StartedAt     *time.Time             `json:"started_at,omitempty"`
	FinishedAt    *time.Time             `json:"finished_at,omitempty"`
	Next          *Node                  `json:"-"`
	Prev          *Node                  `json:"-"`
}

type Chain struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Nodes        []*Node   `json:"nodes"`
	Head         *Node     `json:"-"`
	Tail         *Node     `json:"-"`
	CurrentNode  *Node     `json:"current_node,omitempty"`
	CurrentIndex int       `json:"current_index"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ChainConfig struct {
	MaxNodes     int           `json:"max_nodes"`
	Timeout      time.Duration `json:"timeout"`
	AutoContinue bool          `json:"auto_continue"`
}

const (
	DefaultMaxNodes     = 1000
	DefaultTimeout      = 300 * time.Second
	DefaultAutoContinue = true
)

func NewChainConfig() *ChainConfig {
	return &ChainConfig{
		MaxNodes:     DefaultMaxNodes,
		Timeout:      DefaultTimeout,
		AutoContinue: DefaultAutoContinue,
	}
}
