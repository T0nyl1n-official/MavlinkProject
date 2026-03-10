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

	Drones "MavlinkProject/Server/backend/Shared/Drones"
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

// ==================== 连接管理 ====================

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

	if !h.started {
		return ConnectionStatusDisconnected
	}
	if h.stopped {
		return ConnectionStatusDisconnected
	}
	if h.node == nil {
		return ConnectionStatusError
	}
	return ConnectionStatusConnected
}

func (h *MAVLinkHandlerV1) GetDroneStatus() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.drone != nil {
		return string(h.drone.GetStatus())
	}
	return "unknown"
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
