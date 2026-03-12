package ProgressChain

import (
	"fmt"
	"time"

	gin "github.com/gin-gonic/gin"

	Mavlink "MavlinkProject/Server/backend/Handler/Mavlink"
)

// =============================================================================
// 常量定义
// =============================================================================

// NodeStatus 节点状态
// 记录节点的当前状态, 用于区分不同的状态
type NodeStatus string

const (
	NodeStatusWaiting  NodeStatus = "waiting"
	NodeStatusRunning  NodeStatus = "running"
	NodeStatusFinished NodeStatus = "finished"
	NodeStatusError    NodeStatus = "error"
)

// NodeType 节点类型
// 记录节点的操作类型, 用于区分不同的操作
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

// ChainConfig 默认配置
const (
	DefaultMaxNodes     = 1000
	DefaultTimeout      = 300 * time.Second
	DefaultAutoContinue = true
)

// Redis 配置常量
const (
	RedisDBChainErrors = 4 // 用于存储链错误的 Redis 数据库编号
)

// =============================================================================
// 结构体定义
// =============================================================================

// handlerConfig Mavlink协议-处理器配置
// 同理于结构体: MavlinkV1_Config
// 其记录json传输参数名称, 适配于Node参数 "HandlerConfig"
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

// ProgressChain内重要对象 - 节点(Node)
// 其记录操作事项的各项参数
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

// ProgressChain对象.
// 本质为带指针数组链表, 记录链的所有节点, 并通过Head/Tail指针指向链头/链尾
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

// ProgressChain 的配置信息.
// 此结构体独立出来, 同Chain, Mutex共同整合为 Chain.go 中的ChainManager对象
type ChainConfig struct {
	MaxNodes     int           `json:"max_nodes"`
	Timeout      time.Duration `json:"timeout"`
	AutoContinue bool          `json:"auto_continue"`
}

// =============================================================================
// HandlerConfig 方法 - Default/Create/Update/Delete
// =============================================================================

// Default 为 HandlerConfig 设置默认值
// 用于在没有提供配置时使用默认参数创建 HandlerConfig
//
// 默认值说明:
//   - ConnectionType: "UDP" - 默认使用 UDP 连接
//   - SystemID: 1 - 默认系统 ID
//   - ComponentID: 1 - 默认组件 ID
//   - ProtocolVersion: "v2" - 默认使用 MAVLink v2 协议
//   - HeartbeatRate: 1s - 默认心跳间隔 1 秒
func (hc *HandlerConfig) Default() {
	hc.ConnectionType = "UDP"
	hc.SystemID = 1
	hc.ComponentID = 1
	hc.ProtocolVersion = "v2"
	hc.HeartbeatRate = 1 * time.Second

	if hc.UDPPort == 0 {
		hc.UDPPort = 14550
	}
	if hc.TCPPort == 0 {
		hc.TCPPort = 5760
	}
	if hc.SerialBaud == 0 {
		hc.SerialBaud = 115200
	}
}

// Create 从 gin.Context 中解析并创建 HandlerConfig
// 如果解析失败或缺少必需字段, 自动使用默认值填充
//
// 参数说明:
//   - ctx: gin.Context 指针, 包含请求的 JSON body
//
// 返回值:
//   - error: 如果请求体为空或解析失败返回错误信息
//
// 注意: 此方法会自动调用 Default() 填充缺失的字段
func (hc *HandlerConfig) Create(ctx *gin.Context) error {
	// 尝试从请求体解析 JSON
	if err := ctx.ShouldBindJSON(hc); err != nil {
		// 解析失败时使用默认值
		hc.Default()
		return fmt.Errorf("failed to parse HandlerConfig from request body: %v, using default config", err)
	}

	// 验证必需字段, 如果缺失则使用默认值填充
	if hc.SystemID == 0 {
		hc.SystemID = 1
	}
	if hc.ComponentID == 0 {
		hc.ComponentID = 1
	}
	if hc.ProtocolVersion == "" {
		hc.ProtocolVersion = "v2"
	}
	if hc.ConnectionType == "" {
		hc.ConnectionType = "UDP"
	}

	return nil
}

// Update 从 gin.Context 中更新 HandlerConfig
// 如果请求中缺少某些字段, 保留原值不变
// 如果发现请求中有对应的结构体但解析失败, 将错误记录到 Redis
//
// 参数说明:
//   - ctx: gin.Context 指针, 包含请求的 JSON body
//
// 返回值:
//   - error: 如果请求体格式错误返回错误信息, 同时将错误存入 Redis
//
// 注意:
//   - 与 Create 不同, Update 不会自动使用默认值填充缺失字段
//   - 错误信息会被包装后存入 Redis 的 DB4
func (hc *HandlerConfig) Update(ctx *gin.Context) error {
	// 创建临时结构体用于解析请求
	var temp HandlerConfig

	// 尝试解析 JSON
	if err := ctx.ShouldBindJSON(&temp); err != nil {
		// 解析失败, 将错误存入 Redis
		errMsg := fmt.Sprintf("[HandlerConfig Update Error] Time: %s, Error: %v",
			time.Now().Format(time.RFC3339), err)
		reportToRedisWithDB("handler_config_update", errMsg, RedisDBChainErrors)

		return fmt.Errorf("invalid request body for HandlerConfig update: %v", err)
	}

	// 只有当请求中提供的字段才更新, 其他保持原值
	if temp.ConnectionType != "" {
		hc.ConnectionType = temp.ConnectionType
	}
	if temp.SerialPort != "" {
		hc.SerialPort = temp.SerialPort
	}
	if temp.SerialBaud != 0 {
		hc.SerialBaud = temp.SerialBaud
	}
	if temp.UDPAddr != "" {
		hc.UDPAddr = temp.UDPAddr
	}
	if temp.UDPPort != 0 {
		hc.UDPPort = temp.UDPPort
	}
	if temp.TCPAddr != "" {
		hc.TCPAddr = temp.TCPAddr
	}
	if temp.TCPPort != 0 {
		hc.TCPPort = temp.TCPPort
	}
	if temp.SystemID != 0 {
		hc.SystemID = temp.SystemID
	}
	if temp.ComponentID != 0 {
		hc.ComponentID = temp.ComponentID
	}
	if temp.ProtocolVersion != "" {
		hc.ProtocolVersion = temp.ProtocolVersion
	}
	if temp.HeartbeatRate != 0 {
		hc.HeartbeatRate = temp.HeartbeatRate
	}

	return nil
}

// Delete 重置 HandlerConfig 为空状态
// 调用后 HandlerConfig 的所有字段将被清空
func (hc *HandlerConfig) Delete() {
	hc.ConnectionType = ""
	hc.SerialPort = ""
	hc.SerialBaud = 0
	hc.UDPAddr = ""
	hc.UDPPort = 0
	hc.TCPAddr = ""
	hc.TCPPort = 0
	hc.SystemID = 0
	hc.ComponentID = 0
	hc.ProtocolVersion = ""
	hc.HeartbeatRate = 0
}

// 转换函数 ToMavlinkV1Config 将 HandlerConfig 转换为 MAVLink v1 配置
//
// 返回值:
//   - *Mavlink.MAVLinkConfigV1: 转换后的 MAVLink v1 配置指针
func (hc *HandlerConfig) ToMavlinkV1Config() *Mavlink.MAVLinkConfigV1 {
	return &Mavlink.MAVLinkConfigV1{
		ConnectionType:  Mavlink.ConnectionType(hc.ConnectionType),
		SerialPort:      hc.SerialPort,
		SerialBaud:      hc.SerialBaud,
		UDPAddr:         hc.UDPAddr,
		UDPPort:         hc.UDPPort,
		TCPAddr:         hc.TCPAddr,
		TCPPort:         hc.TCPPort,
		SystemID:        hc.SystemID,
		ComponentID:     hc.ComponentID,
		ProtocolVersion: Mavlink.ProtocolVersion(hc.ProtocolVersion),
		HeartbeatRate:   hc.HeartbeatRate,
	}
}

// =============================================================================
// Node 方法 - Default/Create/Update/Delete
// =============================================================================

// Default 为 Node 设置默认值
// 用于在没有提供参数时创建 Node
//
// 默认值说明:
//   - ID: 自动生成唯一节点 ID
//   - Status: NodeStatusWaiting - 默认等待状态
//   - HandlerConfig: 创建新的默认 HandlerConfig
func (n *Node) Default() {
	n.ID = GenerateNodeID()
	n.Status = NodeStatusWaiting

	if n.HandlerConfig == nil {
		n.HandlerConfig = &HandlerConfig{}
	}
	n.HandlerConfig.Default()

	if n.Params == nil {
		n.Params = make(map[string]interface{})
	}
}

// Create 从 gin.Context 中解析并创建 Node
// 如果解析失败或缺少必需字段, 自动使用默认值填充
//
// 参数说明:
//   - ctx: gin.Context 指针, 包含请求的 JSON body
//
// 返回值:
//   - error: 如果请求体为空或解析失败返回错误信息
//
// 注意: 此方法会自动调用 Default() 填充缺失的字段
func (n *Node) Create(ctx *gin.Context) error {
	// 尝试从请求体解析 JSON
	if err := ctx.ShouldBindJSON(n); err != nil {
		// 解析失败时使用默认值
		n.Default()
		return fmt.Errorf("failed to parse Node from request body: %v, using default node", err)
	}

	// 验证必需字段
	if n.ID == "" {
		n.ID = GenerateNodeID()
	}
	if n.Status == "" {
		n.Status = NodeStatusWaiting
	}
	if n.HandlerConfig == nil {
		n.HandlerConfig = &HandlerConfig{}
		n.HandlerConfig.Default()
	}
	if n.Params == nil {
		n.Params = make(map[string]interface{})
	}

	return nil
}

// Update 从 gin.Context 中更新 Node
// 如果请求中缺少某些字段, 保留原值不变
// 如果发现请求中有对应的结构体但解析失败, 将错误记录到 Redis
//
// 参数说明:
//   - ctx: gin.Context 指针, 包含请求的 JSON body
//
// 返回值:
//   - error: 如果请求体格式错误返回错误信息, 同时将错误存入 Redis
//
// 注意:
//   - 与 Create 不同, Update 不会自动使用默认值填充缺失字段
//   - 错误信息会被包装后存入 Redis 的 DB4
func (n *Node) Update(ctx *gin.Context) error {
	// 创建临时结构体用于解析请求
	var temp Node

	// 尝试解析 JSON
	if err := ctx.ShouldBindJSON(&temp); err != nil {
		// 解析失败, 将错误存入 Redis
		errMsg := fmt.Sprintf("[Node Update Error] Time: %s, Error: %v",
			time.Now().Format(time.RFC3339), err)
		reportToRedisWithDB("node_update", errMsg, RedisDBChainErrors)

		return fmt.Errorf("invalid request body for Node update: %v", err)
	}

	// 只有当请求中提供的字段才更新, 其他保持原值
	if temp.Type != "" {
		n.Type = temp.Type
	}
	if temp.Status != "" {
		n.Status = temp.Status
	}
	if temp.HandlerConfig != nil {
		n.HandlerConfig = temp.HandlerConfig
	}
	if temp.Params != nil {
		n.Params = temp.Params
	}
	if temp.Result != "" {
		n.Result = temp.Result
	}
	if temp.Error != "" {
		n.Error = temp.Error
	}

	return nil
}

// Delete 重置 Node 为空状态
// 调用后 Node 的所有字段将被重置
func (n *Node) Delete() {
	n.ID = ""
	n.Type = ""
	n.Status = ""
	n.HandlerConfig = nil
	n.Params = nil
	n.Result = ""
	n.Error = ""
	n.StartedAt = nil
	n.FinishedAt = nil
	n.Next = nil
	n.Prev = nil
}

// =============================================================================
// ChainConfig 方法
// =============================================================================

// NewChainConfig 创建默认的ChainConfig
func NewChainConfig() *ChainConfig {
	return &ChainConfig{
		MaxNodes:     DefaultMaxNodes,
		Timeout:      DefaultTimeout,
		AutoContinue: DefaultAutoContinue,
	}
}

// NewChainConfig_Custom 创建自定义的ChainConfig
func NewChainConfig_Custom(maxNodes int, timeout time.Duration, autoContinue bool) *ChainConfig {
	return &ChainConfig{
		MaxNodes:     maxNodes,
		Timeout:      timeout,
		AutoContinue: autoContinue,
	}
}

// =============================================================================
// Redis 错误报告方法
// =============================================================================

// reportToRedisWithDB 将错误信息存入指定 Redis 数据库
// 用于记录链执行过程中的错误信息
//
// 参数说明:
//   - key: Redis 键名, 通常使用有意义的标识符
//   - message: 错误信息内容
//   - db: Redis 数据库编号 (0-15)
//
// 注意: 这是一个 TODO 实现, 目前仅打印日志
//
//	实际实现需要连接 Redis 服务器
func reportToRedisWithDB(key string, message string, db int) {
	// TODO: 实现 Redis 错误报告
	// 实现思路:
	// 1. 连接 Redis 服务器
	//    rdb := redis.NewClient(&redis.Options{
	//        Addr: "localhost:6379",
	//        DB:   db,
	//    })
	//
	// 2. 使用 LPUSH 将错误信息推送到列表
	//    rdb.LPush(key, message)
	//
	// 3. 可选: 设置过期时间 (例如 7 天)
	//    rdb.Expire(key, 7*24*time.Hour)
	//
	// 4. 可选: 限制列表长度, 防止无限增长
	//    rdb.LTrim(key, 0, 999)

	// 临时实现: 打印到控制台
	fmt.Printf("[Redis DB%d] %s: %s\n", db, key, message)
}
