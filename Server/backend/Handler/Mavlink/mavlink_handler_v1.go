package Mavlink

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"sync"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"github.com/bluenviron/gomavlib/v3/pkg/message"
	"github.com/gin-gonic/gin"

	Drones "MavlinkProject/Server/backend/Shared/Drones"
)

// =============================================================================
// 全局 Handler 池 - 用于管理多个 MAVLinkHandlerV1 实例
// =============================================================================

var (
	handlerPool    = make(map[string]*MAVLinkHandlerV1)
	handlerPoolMux sync.RWMutex
)

// ==================== MAVLink v1 Handler ====================
type MAVLinkHandlerV1 struct {
	handlerID string

	config   MAVLinkConfigV1
	node     *gomavlib.Node
	drone    *Drones.Drone
	mu       sync.RWMutex
	started  bool
	stopped  bool
	stopChan chan bool

	groundStation *GroundStationInfoV1

	messageChan   chan *IncomingMessageV1
	heartbeatChan chan *HeartbeatDataV1
	GPSChan       chan *GPSDataV1
	attitudeChan  chan *AttitudeDataV1
	batteryChan   chan *BatteryDataV1
}

// =============================================================================
// MAVLinkHandlerV1 对象方法 - Create/Update/Delete
// =============================================================================

// Create 从 gin.Context 解析配置并初始化当前 MAVLinkHandlerV1
// 解析请求中的配置, 设置到当前 handler 实例
//
// 参数说明:
//   - ctx: gin.Context 指针, 包含请求的 JSON body
//
// 返回值:
//   - error: 如果解析失败返回错误信息
//
// 注意:
//   - 此方法会覆盖当前 handler 的配置
//   - 如果 handler 正在运行, 需要重启才能生效
func (h *MAVLinkHandlerV1) Create(ctx *gin.Context) error {
	var config MAVLinkConfigV1

	// 尝试从请求体解析配置
	if err := ctx.ShouldBindJSON(&config); err != nil {
		return fmt.Errorf("failed to parse handler config from request body: %v", err)
	}

	// 验证必需字段, 使用默认值填充
	if config.ConnectionType == "" {
		config.ConnectionType = ConnectionUDP
	}
	if config.SystemID == 0 {
		config.SystemID = 1
	}
	if config.ComponentID == 0 {
		config.ComponentID = 1
	}
	if config.ProtocolVersion == "" {
		config.ProtocolVersion = ProtocolVersionV2
	}
	if config.HeartbeatRate == 0 {
		config.HeartbeatRate = 1 * time.Second
	}

	// 根据连接类型验证参数
	switch config.ConnectionType {
	case ConnectionSerial:
		if config.SerialPort == "" {
			return fmt.Errorf("serial_port is required for serial connection")
		}
		if config.SerialBaud == 0 {
			config.SerialBaud = 115200
		}
	case ConnectionUDP:
		if config.UDPPort == 0 {
			config.UDPPort = 14550
		}
	case ConnectionTCP:
		if config.TCPAddr == "" {
			return fmt.Errorf("tcp_addr is required for tcp connection")
		}
		if config.TCPPort == 0 {
			config.TCPPort = 5760
		}
	}

	// 更新当前 handler 的配置
	h.config = config
	h.handlerID = generateHandlerIDV1(5)

	return nil
}

// Update 从 gin.Context 解析配置并更新当前 MAVLinkHandlerV1
// 解析请求中的配置, 只更新提供的字段, 保留其他字段的原值
//
// 参数说明:
//   - ctx: gin.Context 指针, 包含请求的 JSON body
//
// 返回值:
//   - error: 如果解析失败返回错误信息
//
// 注意:
//   - 此方法仅更新提供的字段, 其他字段保持不变
//   - 如果 handler 正在运行, 需要重启才能生效
func (h *MAVLinkHandlerV1) Update(ctx *gin.Context) error {
	// 解析新配置
	var newConfig MAVLinkConfigV1
	if err := ctx.ShouldBindJSON(&newConfig); err != nil {
		return fmt.Errorf("failed to parse updated config: %v", err)
	}

	// 更新配置 (只有提供的字段才会被更新)
	if newConfig.ConnectionType != "" {
		h.config.ConnectionType = newConfig.ConnectionType
	}
	if newConfig.SerialPort != "" {
		h.config.SerialPort = newConfig.SerialPort
	}
	if newConfig.SerialBaud != 0 {
		h.config.SerialBaud = newConfig.SerialBaud
	}
	if newConfig.UDPAddr != "" {
		h.config.UDPAddr = newConfig.UDPAddr
	}
	if newConfig.UDPPort != 0 {
		h.config.UDPPort = newConfig.UDPPort
	}
	if newConfig.TCPAddr != "" {
		h.config.TCPAddr = newConfig.TCPAddr
	}
	if newConfig.TCPPort != 0 {
		h.config.TCPPort = newConfig.TCPPort
	}
	if newConfig.SystemID != 0 {
		h.config.SystemID = newConfig.SystemID
	}
	if newConfig.ComponentID != 0 {
		h.config.ComponentID = newConfig.ComponentID
	}
	if newConfig.ProtocolVersion != "" {
		h.config.ProtocolVersion = newConfig.ProtocolVersion
	}
	if newConfig.HeartbeatRate != 0 {
		h.config.HeartbeatRate = newConfig.HeartbeatRate
	}

	return nil
}

// Delete 停止当前 handler 并从池中移除
// 先尝试停止 handler (如果正在运行), 然后从池中移除
//
// 返回值:
//   - error: 如果停止失败返回错误信息
func (h *MAVLinkHandlerV1) Delete() error {
	handlerID := h.handlerID

	// 如果 handler 正在运行, 先停止
	if h.started && !h.stopped {
		if err := h.Stop(); err != nil {
			return err
		}
	}

	// 从池中删除
	handlerPoolMux.Lock()
	defer handlerPoolMux.Unlock()
	delete(handlerPool, handlerID)

	return nil
}

// GetID 获取当前 handler 的唯一标识符
//
// 返回值:
//   - string: handler 的 ID
func (h *MAVLinkHandlerV1) GetID() string {
	return h.handlerID
}

// IsRunning 检查 handler 是否正在运行
//
// 返回值:
//   - bool: 运行状态
func (h *MAVLinkHandlerV1) IsRunning() bool {
	return h.started && !h.stopped
}

// GetInfo 获取当前 handler 的基本信息
//
// 返回值:
//   - map[string]interface{}: handler 的基本信息
func (h *MAVLinkHandlerV1) GetInfo() map[string]interface{} {
	return map[string]interface{}{
		"handler_id":       h.handlerID,
		"connection_type":  h.config.ConnectionType,
		"system_id":        h.config.SystemID,
		"component_id":     h.config.ComponentID,
		"protocol_version": h.config.ProtocolVersion,
		"started":          h.started,
		"stopped":          h.stopped,
	}
}

// =============================================================================
// handlerPool 全局 CRUD 函数 (供路由调用)
// =============================================================================

// CreateHandlerV1 从 gin.Context 创建新的 MAVLinkHandlerV1
// 解析请求中的配置, 创建 handler 并添加到池中
//
// 参数说明:
//   - ctx: gin.Context 指针, 包含请求的 JSON body
//
// 返回值:
//   - *MAVLinkHandlerV1: 创建的 handler 实例
//   - string: handler 的唯一 ID
//   - error: 如果创建失败返回错误信息
func CreateHandlerV1(ctx *gin.Context) (*MAVLinkHandlerV1, string, error) {
	var config MAVLinkConfigV1

	if err := ctx.ShouldBindJSON(&config); err != nil {
		return nil, "", fmt.Errorf("failed to parse handler config: %v", err)
	}

	// 验证必需字段
	if config.ConnectionType == "" {
		config.ConnectionType = ConnectionUDP
	}
	if config.SystemID == 0 {
		config.SystemID = 1
	}
	if config.ComponentID == 0 {
		config.ComponentID = 1
	}
	if config.ProtocolVersion == "" {
		config.ProtocolVersion = ProtocolVersionV2
	}
	if config.HeartbeatRate == 0 {
		config.HeartbeatRate = 1 * time.Second
	}

	// 根据连接类型验证参数
	switch config.ConnectionType {
	case ConnectionSerial:
		if config.SerialPort == "" {
			return nil, "", fmt.Errorf("serial_port is required for serial connection")
		}
		if config.SerialBaud == 0 {
			config.SerialBaud = 115200
		}
	case ConnectionUDP:
		if config.UDPPort == 0 {
			config.UDPPort = 14550
		}
	case ConnectionTCP:
		if config.TCPAddr == "" {
			return nil, "", fmt.Errorf("tcp_addr is required for tcp connection")
		}
		if config.TCPPort == 0 {
			config.TCPPort = 5760
		}
	}

	// 创建 handler
	handler := NewMAVLinkHandlerV1(config)
	handlerID := handler.GetHandlerID()

	// 添加到池中
	handlerPoolMux.Lock()
	handlerPool[handlerID] = handler
	handlerPoolMux.Unlock()

	return handler, handlerID, nil
}

// GetHandlerV1 根据 handlerID 获取 MAVLinkHandlerV1
func GetHandlerV1(handlerID string) *MAVLinkHandlerV1 {
	handlerPoolMux.RLock()
	defer handlerPoolMux.RUnlock()
	return handlerPool[handlerID]
}

// UpdateHandlerV1 根据 handlerID 更新 MAVLinkHandlerV1 的配置
func UpdateHandlerV1(handlerID string, ctx *gin.Context) error {
	handlerPoolMux.RLock()
	handler, exists := handlerPool[handlerID]
	handlerPoolMux.RUnlock()

	if !exists {
		return fmt.Errorf("handler with id %s not found", handlerID)
	}

	return handler.Update(ctx)
}

// DeleteHandlerV1 根据 handlerID 删除 MAVLinkHandlerV1
func DeleteHandlerV1(handlerID string) error {
	handler := GetHandlerV1(handlerID)
	if handler == nil {
		return fmt.Errorf("handler with id %s not found", handlerID)
	}
	return handler.Delete()
}

// ListHandlerV1 列出所有已创建的 MAVLinkHandlerV1
func ListHandlerV1() []map[string]interface{} {
	handlerPoolMux.RLock()
	defer handlerPoolMux.RUnlock()

	result := make([]map[string]interface{}, 0, len(handlerPool))
	for _, handler := range handlerPool {
		result = append(result, handler.GetInfo())
	}

	return result
}

// GetHandlerCountV1 获取 handler 池中的 handler 数量
func GetHandlerCountV1() int {
	handlerPoolMux.RLock()
	defer handlerPoolMux.RUnlock()
	return len(handlerPool)
}

// ClearAllHandlersV1 清除 handler 池中的所有 handler
func ClearAllHandlersV1() {
	handlerPoolMux.Lock()
	defer handlerPoolMux.Unlock()

	for _, handler := range handlerPool {
		if handler.started && !handler.stopped {
			handler.Stop()
		}
	}

	handlerPool = make(map[string]*MAVLinkHandlerV1)
}

// ==================== 构造函数 ====================

func NewMAVLinkHandlerV1(config MAVLinkConfigV1) *MAVLinkHandlerV1 {
	// 创建默认的无人机配置
	droneConfig := Drones.DroneConfig{
		SystemID:        config.SystemID,
		ComponentID:     config.ComponentID,
		ProtocolVersion: string(config.ProtocolVersion),
		HeartbeatRate:   config.HeartbeatRate,
		Timeout:         30 * time.Second,
	}

	return &MAVLinkHandlerV1{
		handlerID:     generateHandlerIDV1(5),
		config:        config,
		drone:         Drones.NewDrone("default", "Default Drone", "Generic", droneConfig),
		stopChan:      make(chan bool),
		messageChan:   make(chan *IncomingMessageV1, 100),
		heartbeatChan: make(chan *HeartbeatDataV1, 100),
		GPSChan:       make(chan *GPSDataV1, 100),
		attitudeChan:  make(chan *AttitudeDataV1, 100),
		batteryChan:   make(chan *BatteryDataV1, 100),
	}
}

func generateHandlerIDV1(length int) string {
	const digits = "0123456789"
	var result string

	firstDigit, err := rand.Int(rand.Reader, big.NewInt(9))
	if err != nil {
		return "10000"
	}
	result += fmt.Sprintf("%d", firstDigit.Int64()+1)

	for i := 1; i < length; i++ {
		d, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			break
		}
		result += string(digits[d.Int64()])
	}

	return result
}

// ==================== 连接管理 (MAVLink v1) ====================

func (h *MAVLinkHandlerV1) Start() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.started {
		return fmt.Errorf("MAVLink handler already started")
	}

	if h.stopped {
		h.stopChan = make(chan bool)
		h.stopped = false
	}

	var endpointConf gomavlib.EndpointConf

	switch h.config.ConnectionType {
	case ConnectionSerial:
		endpointConf = gomavlib.EndpointSerial{
			Device: h.config.SerialPort,
			Baud:   h.config.SerialBaud,
		}
	case ConnectionUDP:
		endpointConf = gomavlib.EndpointUDPClient{
			Address: fmt.Sprintf("%s:%d", h.config.UDPAddr, h.config.UDPPort),
		}
	case ConnectionTCP:
		endpointConf = gomavlib.EndpointTCPClient{
			Address: fmt.Sprintf("%s:%d", h.config.TCPAddr, h.config.TCPPort),
		}
	default:
		return fmt.Errorf("unsupported connection type: %s", h.config.ConnectionType)
	}

	node := gomavlib.Node{}
	err := node.Initialize()

	node.Endpoints = []gomavlib.EndpointConf{endpointConf}
	node.Dialect = common.Dialect
	node.OutVersion = gomavlib.V2
	node.OutSystemID = byte(h.config.SystemID)
	node.OutComponentID = 1
	node.HeartbeatDisable = false
	node.HeartbeatPeriod = h.config.HeartbeatRate
	node.StreamRequestEnable = true

	if err != nil {
		return fmt.Errorf("failed to create MAVLink node: %w", err)
	}

	h.node = &node
	h.started = true

	go h.readLoop()
	go h.processLoop()

	return nil
}

func (h *MAVLinkHandlerV1) Stop() error {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.stopped {
		return nil
	}

	h.stopped = true

	if !h.started {
		return fmt.Errorf("MAVLink handler not started")
	}

	close(h.stopChan)

	if h.node != nil {
		h.node.Close()
		h.node = nil
	}

	h.started = false
	return nil
}

func (h *MAVLinkHandlerV1) Restart() error {
	if err := h.Stop(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return h.Start()
}

func (h *MAVLinkHandlerV1) RestartWithTimeout(timeout time.Duration) error {
	done := make(chan error, 1)

	go func() {
		done <- h.Restart()
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("restart timeout after %v", timeout)
	}
}

// ==================== 状态查询 ====================

func (h *MAVLinkHandlerV1) GetHandlerID() string {
	return h.handlerID
}

func (h *MAVLinkHandlerV1) GetConnectionStatus() ConnectionStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.started || h.stopped {
		return ConnectionStatusDisconnected
	}
	if h.node == nil {
		return ConnectionStatusError
	}

	return ConnectionStatusConnected
}

func (h *MAVLinkHandlerV1) GetDrone() *Drones.Drone {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.drone
}

func (h *MAVLinkHandlerV1) GetDroneStatus() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.drone != nil {
		return string(h.drone.GetStatus())
	}
	return "Error : unknown"
}

func (h *MAVLinkHandlerV1) GetDronePosition() Drones.Position {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.drone != nil {
		return h.drone.GetPosition()
	}
	return Drones.Position{}
}

func (h *MAVLinkHandlerV1) GetDroneAttitude() Drones.Attitude {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.drone != nil {
		return h.drone.GetAttitude()
	}
	return Drones.Attitude{}
}

func (h *MAVLinkHandlerV1) GetDroneBattery() Drones.BatteryStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.drone != nil {
		return h.drone.GetBattery()
	}
	return Drones.BatteryStatus{}
}

func (h *MAVLinkHandlerV1) GetGroundStation() *GroundStationInfoV1 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.groundStation
}

func (h *MAVLinkHandlerV1) GetGroundStationInfo() GroundStationInfoV1 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.groundStation != nil {
		return *h.groundStation
	}
	return GroundStationInfoV1{}
}

func (h *MAVLinkHandlerV1) GetConfig() MAVLinkConfigV1 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config
}

// ==================== 地面站管理 ====================

func (h *MAVLinkHandlerV1) SetGroundStation(name, id string, lat, lon, alt float64) {
	h.mu.Lock()
	defer h.mu.Unlock()

	position := Drones.Position{
		Latitude:  lat,
		Longitude: lon,
		Altitude:  alt,
	}

	h.groundStation = &GroundStationInfoV1{
		Name:     name,
		ID:       id,
		Position: position,
	}
}

// ==================== 无人机控制 ====================

func (h *MAVLinkHandlerV1) SendTakeoff(altitude float32) error {
	msg := &common.MessageCommandLong{
		Command:      common.MAV_CMD_NAV_TAKEOFF,
		Param1:       0,
		Param2:       0,
		Param3:       0,
		Param4:       0,
		Param5:       0,
		Param6:       0,
		Param7:       altitude,
		Confirmation: 0,
	}
	return h.SendMessage(msg)
}

func (h *MAVLinkHandlerV1) SendLand(lat, lon, alt float64) error {
	msg := &common.MessageCommandLong{
		Command:      common.MAV_CMD_NAV_LAND,
		Param1:       0,
		Param2:       0,
		Param3:       0,
		Param4:       0,
		Param5:       float32(lat),
		Param6:       float32(lon),
		Param7:       float32(alt),
		Confirmation: 0,
	}
	return h.SendMessage(msg)
}

func (h *MAVLinkHandlerV1) SendMoveToPosition(lat, lon, alt float64, speed float32) error {
	msg := &common.MessageSetPositionTargetGlobalInt{
		TimeBootMs:      uint32(time.Now().UnixMilli()),
		CoordinateFrame: common.MAV_FRAME_GLOBAL_RELATIVE_ALT,
		TypeMask:        0b0000111111111000,
		LatInt:          int32(lat * 1e7),
		LonInt:          int32(lon * 1e7),
		Alt:             float32(alt),
		Vx:              0,
		Vy:              0,
		Vz:              0,
		Afx:             0,
		Afy:             0,
		Afz:             0,
		Yaw:             0,
		YawRate:         0,
	}
	return h.SendMessage(msg)
}

func (h *MAVLinkHandlerV1) SetFlightMode(mode FlightMode) error {
	var baseMode common.MAV_MODE
	var customMode uint32

	switch mode {
	case FlightModeManual:
		baseMode = 1 // MAV_MODE_MANUAL
	case FlightModeStabilize:
		baseMode = 2 // MAV_MODE_STABILIZE_DISARMED
	case FlightModeAuto:
		baseMode = 4 // MAV_MODE_AUTO
	case FlightModeGuided:
		baseMode = 8 // MAV_MODE_GUIDED
	case FlightModeRTL:
		baseMode = 16 // MAV_MODE_RETURN_HOME
	case FlightModeLand:
		baseMode = 32 // MAV_MODE_LAND
	default:
		return fmt.Errorf("unsupported flight mode: %s", mode)
	}

	msg := &common.MessageSetMode{
		BaseMode:   baseMode,
		CustomMode: customMode,
	}
	return h.SendMessage(msg)
}

func (h *MAVLinkHandlerV1) SendReturnToLaunch() error {
	return h.SetFlightMode(FlightModeRTL)
}

func (h *MAVLinkHandlerV1) SendHeartbeat() error {
	msg := &common.MessageHeartbeat{
		Type:         common.MAV_TYPE(0),
		Autopilot:    common.MAV_AUTOPILOT(0),
		BaseMode:     common.MAV_MODE_FLAG(0),
		SystemStatus: common.MAV_STATE(0),
		CustomMode:   0,
	}
	return h.SendMessage(msg)
}

// ==================== 连接配置调整 ====================

func (h *MAVLinkHandlerV1) UpdateConnectionType(connType ConnectionType, params map[string]interface{}) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.config.ConnectionType = connType

	switch connType {
	case ConnectionSerial:
		if port, ok := params["port"].(string); ok {
			h.config.SerialPort = port
		}
		if baud, ok := params["baud"].(int); ok {
			h.config.SerialBaud = baud
		}
	case ConnectionUDP:
		if addr, ok := params["addr"].(string); ok {
			h.config.UDPAddr = addr
		}
		if port, ok := params["port"].(int); ok {
			h.config.UDPPort = port
		}
	case ConnectionTCP:
		if addr, ok := params["addr"].(string); ok {
			h.config.TCPAddr = addr
		}
		if port, ok := params["port"].(int); ok {
			h.config.TCPPort = port
		}
	}

	return h.Restart()
}

func (h *MAVLinkHandlerV1) UpdateSystemID(systemID int) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.config.SystemID = systemID
	return h.Restart()
}

func (h *MAVLinkHandlerV1) UpdateHeartbeatRate(rate time.Duration) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.config.HeartbeatRate = rate
	return h.Restart()
}

// ==================== 消息发送 ====================

func (h *MAVLinkHandlerV1) SendMessage(msg message.Message) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if h.node == nil {
		return fmt.Errorf("MAVLink node not initialized")
	}

	return h.node.WriteMessageAll(msg)
}

func (h *MAVLinkHandlerV1) RequestMessageStream(messageID int, rate int) error {
	msg := &common.MessageRequestDataStream{
		ReqStreamId:    uint8(messageID),
		ReqMessageRate: uint16(rate),
		StartStop:      1,
	}
	return h.SendMessage(msg)
}

// ==================== 事件处理循环 ====================

func (h *MAVLinkHandlerV1) readLoop() {
	for {
		select {
		case <-h.stopChan:
			return
		case evt := <-h.node.Events():
			h.handleEvent(evt)
		}
	}
}

func (h *MAVLinkHandlerV1) processLoop() {
	// 处理消息队列
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-h.stopChan:
			return
		case <-ticker.C:
			h.processMessages()
		}
	}
}

func (h *MAVLinkHandlerV1) handleEvent(evt gomavlib.Event) {
	switch evt := evt.(type) {
	case *gomavlib.EventFrame:
		h.handleFrame(evt)
	case *gomavlib.EventChannelOpen:
		h.handleChannelOpen(evt)
	case *gomavlib.EventChannelClose:
		h.handleChannelClose(evt)
	case *gomavlib.EventParseError:
		h.handleParseError(evt)
	}
}

func (h *MAVLinkHandlerV1) handleFrame(evt *gomavlib.EventFrame) {
	// 处理接收到的帧
	msg := evt.Message()

	switch msg := msg.(type) {
	case *common.MessageHeartbeat:
		h.handleHeartbeat(msg, int(evt.SystemID()), int(evt.ComponentID()))
	case *common.MessageGlobalPositionInt:
		h.handleGPS(msg, int(evt.SystemID()), int(evt.ComponentID()))
	case *common.MessageAttitude:
		h.handleAttitude(msg, int(evt.SystemID()), int(evt.ComponentID()))
	case *common.MessageBatteryStatus:
		h.handleBattery(msg, int(evt.SystemID()), int(evt.ComponentID()))
	}
}

func (h *MAVLinkHandlerV1) handleHeartbeat(msg *common.MessageHeartbeat, systemID, componentID int) {
	data := &HeartbeatDataV1{
		SystemID:     systemID,
		ComponentID:  componentID,
		Type:         uint64(msg.Type),
		Autopilot:    uint64(msg.Autopilot),
		BaseMode:     uint64(msg.BaseMode),
		CustomMode:   msg.CustomMode,
		SystemStatus: uint64(msg.SystemStatus),
	}

	select {
	case h.heartbeatChan <- data:
	default:
	}
}

func (h *MAVLinkHandlerV1) handleGPS(msg *common.MessageGlobalPositionInt, systemID, componentID int) {
	data := &GPSDataV1{
		SystemID:    systemID,
		ComponentID: componentID,
		Latitude:    float64(msg.Lat) / 1e7,
		Longitude:   float64(msg.Lon) / 1e7,
		Altitude:    float64(msg.Alt) / 1000,
		Timestamp:   time.Now(),
	}

	select {
	case h.GPSChan <- data:
	default:
		// 通道已满，丢弃数据
	}

	// 更新无人机位置
	h.mu.Lock()
	if h.drone != nil {
		position := Drones.Position{
			Latitude:  data.Latitude,
			Longitude: data.Longitude,
			Altitude:  data.Altitude,
		}
		h.drone.SetPosition(position)
	}
	h.mu.Unlock()
}

func (h *MAVLinkHandlerV1) handleAttitude(msg *common.MessageAttitude, systemID, componentID int) {
	data := &AttitudeDataV1{
		SystemID:    systemID,
		ComponentID: componentID,
		Roll:        msg.Roll,
		Pitch:       msg.Pitch,
		Yaw:         msg.Yaw,
		RollSpeed:   msg.Rollspeed,
		PitchSpeed:  msg.Pitchspeed,
		YawSpeed:    msg.Yawspeed,
		Timestamp:   time.Now(),
	}

	select {
	case h.attitudeChan <- data:
	default:
		// 通道已满，丢弃数据
	}

	// 更新无人机姿态
	h.mu.Lock()
	if h.drone != nil {
		attitude := Drones.Attitude{
			Roll:  float64(data.Roll),
			Pitch: float64(data.Pitch),
			Yaw:   float64(data.Yaw),
		}
		h.drone.SetAttitude(attitude)
	}
	h.mu.Unlock()
}

func (h *MAVLinkHandlerV1) handleBattery(msg *common.MessageBatteryStatus, systemID, componentID int) {
	// 计算平均电压
	var totalVoltage float32
	cellCount := 0
	for i, voltage := range msg.Voltages {
		if voltage != 0xFFFF && voltage > 0 {
			totalVoltage += float32(voltage) / 1000.0 // 转换为伏特
			cellCount++
		}
		if i >= 0 && cellCount > 0 {
			break
		}
	}

	var avgVoltage float32
	if cellCount > 0 {
		avgVoltage = totalVoltage / float32(cellCount)
	}

	data := &BatteryDataV1{
		SystemID:    systemID,
		ComponentID: componentID,
		Voltage:     avgVoltage,
		Current:     float32(msg.CurrentBattery) / 100.0, // 转换为安培
		Remaining:   int(msg.BatteryRemaining),
		Temperature: float32(msg.Temperature) / 100.0, // 转换为摄氏度
		Timestamp:   time.Now(),
	}

	select {
	case h.batteryChan <- data:
	default:
		// 通道已满，丢弃数据
	}

	// 更新无人机电池状态
	h.mu.Lock()
	if h.drone != nil {
		battery := Drones.BatteryStatus{
			Voltage:     float64(data.Voltage),
			Current:     float64(data.Current),
			Remaining:   data.Remaining,
			Temperature: float64(data.Temperature),
			CellCount:   cellCount,
		}
		h.drone.SetBattery(battery)
	}
	h.mu.Unlock()
}

func (h *MAVLinkHandlerV1) handleChannelOpen(evt *gomavlib.EventChannelOpen) {
	// 通道打开处理
}

func (h *MAVLinkHandlerV1) handleChannelClose(evt *gomavlib.EventChannelClose) {
	// 通道关闭处理
}

func (h *MAVLinkHandlerV1) handleParseError(evt *gomavlib.EventParseError) {
	// 解析错误处理
}

func (h *MAVLinkHandlerV1) processMessages() {
	// 处理消息队列
}

// ==================== 通道获取 ====================

func (h *MAVLinkHandlerV1) GetMessageChan() <-chan *IncomingMessageV1 {
	return h.messageChan
}

func (h *MAVLinkHandlerV1) GetHeartbeatChan() <-chan *HeartbeatDataV1 {
	return h.heartbeatChan
}

func (h *MAVLinkHandlerV1) GetGPSChan() <-chan *GPSDataV1 {
	return h.GPSChan
}

func (h *MAVLinkHandlerV1) GetAttitudeChan() <-chan *AttitudeDataV1 {
	return h.attitudeChan
}

func (h *MAVLinkHandlerV1) GetBatteryChan() <-chan *BatteryDataV1 {
	return h.batteryChan
}
