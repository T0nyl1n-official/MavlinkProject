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

var (
	pi = float32(3.141592653589793)
)

type ConnectionType string

// 连接类型
const (
	ConnectionSerial ConnectionType = "serial"
	ConnectionUDP    ConnectionType = "udp"
	ConnectionTCP    ConnectionType = "tcp"
)

// 协议版本
type ProtocolVersion string

const (
	ProtocolVersion1 ProtocolVersion = "1.0"
	ProtocolVersion2 ProtocolVersion = "2.0"
)

// 调度者类型
type DispatcherType string

const (
	DispatcherTypeHuman DispatcherType = "human"
	DispatcherTypeAI    DispatcherType = "ai_agent"
)

// 地面站位置信息
type GroundStationInfo struct {
	Name     string          `json:"name"`
	ID       string          `json:"id"`
	Position Drones.Position `json:"position"`
}

// 调度者信息
type DispatcherInfo struct {
	Type     DispatcherType `json:"type"`
	Username string         `json:"username"`
	Email    string         `json:"email,omitempty"`
}

// Mavlink协议处理handler
type MAVLinkHandler struct {
	handlerID string

	config   MAVLinkConfig
	node     *gomavlib.Node
	drone    *Drones.Drone
	mu       sync.RWMutex
	started  bool
	stopped  bool
	stopChan chan bool

	groundStation *GroundStationInfo
	dispatcher    *DispatcherInfo

	messageChan      chan *IncomingMessage
	heartbeatChan    chan *HeartbeatData
	GPSChan          chan *GPSData
	attitudeChan     chan *AttitudeData
	batteryChan      chan *BatteryData
	timestampHistory []time.Time
}

// MAVLink 配置
type MAVLinkConfig struct {
	ConnectionType ConnectionType
	SerialPort     string
	SerialBaud     int
	UDPAddr        string
	UDPPort        int
	TCPAddr        string
	TCPPort        int

	SystemID    int
	ComponentID int

	ProtocolVersion ProtocolVersion

	HeartbeatRate time.Duration
}

// 入站消息
type IncomingMessage struct {
	SystemID    int
	ComponentID int
	MessageID   int
	Message     message.Message
	Timestamp   time.Time
}

// 心跳包数据
type HeartbeatData struct {
	SystemID     int
	ComponentID  int
	Type         uint64
	Autopilot    uint64
	BaseMode     uint64
	CustomMode   uint32
	SystemStatus uint64
}

// 位置数据
type GPSData struct {
	SystemID    int
	ComponentID int
	FixType     uint64
	Lat         int32
	Lon         int32
	Alt         int32
	EPH         uint16
	EPV         uint16
	Vel         uint16
	COG         uint16
}

// 姿态数据
type AttitudeData struct {
	SystemID    int
	ComponentID int
	Roll        float32
	Pitch       float32
	Yaw         float32
	RollSpeed   float32
	PitchSpeed  float32
	YawSpeed    float32
}

// 电池数据
type BatteryData struct {
	SystemID    int
	ComponentID int
	Voltage     float64
	Current     float64
	Remaining   int8
}

// handler 创建方法
func NewMAVLinkHandler(config MAVLinkConfig, drone *Drones.Drone) *MAVLinkHandler {
	if config.SystemID == 0 {
		config.SystemID = 255
	}
	if config.ComponentID == 0 {
		config.ComponentID = 0
	}
	if config.HeartbeatRate == 0 {
		config.HeartbeatRate = 1 * time.Second
	}
	if config.ProtocolVersion == "" {
		config.ProtocolVersion = ProtocolVersion2
	}

	handlerID := generateHandlerID(5)

	return &MAVLinkHandler{
		handlerID:     handlerID,
		config:        config,
		drone:         drone,
		stopChan:      make(chan bool),
		messageChan:   make(chan *IncomingMessage, 1000),
		heartbeatChan: make(chan *HeartbeatData, 100),
		GPSChan:       make(chan *GPSData, 100),
		attitudeChan:  make(chan *AttitudeData, 100),
		batteryChan:   make(chan *BatteryData, 100),
	}
}

// 生成handlerID
func generateHandlerID(length int) string {
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

// 处理开始
func (h *MAVLinkHandler) Start() error {
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
		if h.config.SerialPort == "" {
			return fmt.Errorf("serial port not specified")
		}
		if h.config.SerialBaud == 0 {
			h.config.SerialBaud = 57600
		}
		endpointConf = &gomavlib.EndpointSerial{
			Device: h.config.SerialPort,
			Baud:   h.config.SerialBaud,
		}

	case ConnectionUDP:
		if h.config.UDPAddr == "" {
			h.config.UDPAddr = "0.0.0.0"
		}
		if h.config.UDPPort == 0 {
			h.config.UDPPort = 14550
		}
		endpointConf = &gomavlib.EndpointUDPServer{
			Address: fmt.Sprintf("%s:%d", h.config.UDPAddr, h.config.UDPPort),
		}

	case ConnectionTCP:
		if h.config.TCPAddr == "" {
			return fmt.Errorf("TCP address not specified")
		}
		if h.config.TCPPort == 0 {
			h.config.TCPPort = 5760
		}
		endpointConf = &gomavlib.EndpointTCPClient{
			Address: fmt.Sprintf("%s:%d", h.config.TCPAddr, h.config.TCPPort),
		}

	default:
		return fmt.Errorf("unsupported connection type: %s", h.config.ConnectionType)
	}

	var outVersion gomavlib.Version
	if h.config.ProtocolVersion == ProtocolVersion1 {
		outVersion = gomavlib.V1
	} else {
		outVersion = gomavlib.V2
	}

	node, err := gomavlib.NewNode(gomavlib.NodeConf{
		Endpoints:           []gomavlib.EndpointConf{endpointConf},
		OutVersion:          outVersion,
		OutSystemID:         byte(h.config.SystemID),
		OutComponentID:      byte(h.config.ComponentID),
		HeartbeatPeriod:     h.config.HeartbeatRate,
		HeartbeatDisable:    false,
		Dialect:             common.Dialect,
		StreamRequestEnable: true,
	})

	if err != nil {
		return fmt.Errorf("failed to create MAVLink node: %w", err)
	}

	h.node = node
	h.started = true

	go h.readLoop()
	go h.processLoop()

	return nil
}

// 处理关闭
func (h *MAVLinkHandler) Stop() error {
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

// handler 状态检查
func (h *MAVLinkHandler) IsStarted() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.started
}

func (h *MAVLinkHandler) readLoop() {
	for {
		select {
		case <-h.stopChan:
			return
		case evt, ok := <-h.node.Events():
			if !ok {
				return
			}
			if evt != nil {
				h.handleEvent(evt)
			}
		}
	}
}

func (h *MAVLinkHandler) handleEvent(evt interface{}) {
	switch e := evt.(type) {
	case *gomavlib.EventFrame:
		h.handleFrame(e)
	}
}

func (h *MAVLinkHandler) handleFrame(evt *gomavlib.EventFrame) {
	fr := evt.Frame

	incomingMsg := &IncomingMessage{
		SystemID:    int(fr.GetSystemID()),
		ComponentID: int(fr.GetComponentID()),
		MessageID:   int(fr.GetMessage().GetID()),
		Message:     fr.GetMessage(),
		Timestamp:   time.Now(),
	}

	h.messageChan <- incomingMsg

	switch msg := fr.GetMessage().(type) {
	case *common.MessageHeartbeat:
		h.handleHeartbeat(msg, incomingMsg)
	case *common.MessageGpsRawInt:
		h.handleGPS(msg, incomingMsg)
	case *common.MessageAttitude:
		h.handleAttitude(msg, incomingMsg)
	case *common.MessageBatteryStatus:
		h.handleBattery(msg, incomingMsg)
	case *common.MessageGlobalPositionInt:
		h.handleGlobalPosition(msg, incomingMsg)
	case *common.MessageVfrHud:
		h.handleVfrHud(msg, incomingMsg)
	}
}

func (h *MAVLinkHandler) handleHeartbeat(msg *common.MessageHeartbeat, incoming *IncomingMessage) {
	data := &HeartbeatData{
		SystemID:     incoming.SystemID,
		ComponentID:  incoming.ComponentID,
		Type:         uint64(msg.Type),
		Autopilot:    uint64(msg.Autopilot),
		BaseMode:     uint64(msg.BaseMode),
		CustomMode:   msg.CustomMode,
		SystemStatus: uint64(msg.SystemStatus),
	}

	h.heartbeatChan <- data

	if h.drone != nil {
		h.drone.UpdateHeartbeat()
		h.drone.SetConnected(true)
	}
}

func (h *MAVLinkHandler) handleGPS(msg *common.MessageGpsRawInt, incoming *IncomingMessage) {
	data := &GPSData{
		SystemID:    incoming.SystemID,
		ComponentID: incoming.ComponentID,
		FixType:     uint64(msg.FixType),
		Lat:         msg.Lat,
		Lon:         msg.Lon,
		Alt:         msg.Alt,
		EPH:         msg.Eph,
		EPV:         msg.Epv,
		Vel:         msg.Vel,
		COG:         msg.Cog,
	}

	h.GPSChan <- data

	if h.drone != nil {
		pos := Drones.Position{
			Latitude:  float64(msg.Lat) / 1e7,
			Longitude: float64(msg.Lon) / 1e7,
			Altitude:  float64(msg.Alt) / 1000,
		}
		h.drone.SetPosition(pos)
	}
}

func (h *MAVLinkHandler) handleAttitude(msg *common.MessageAttitude, incoming *IncomingMessage) {
	data := &AttitudeData{
		SystemID:    incoming.SystemID,
		ComponentID: incoming.ComponentID,
		Roll:        msg.Roll,
		Pitch:       msg.Pitch,
		Yaw:         msg.Yaw,
		RollSpeed:   msg.Rollspeed,
		PitchSpeed:  msg.Pitchspeed,
		YawSpeed:    msg.Yawspeed,
	}

	h.attitudeChan <- data

	if h.drone != nil {
		att := Drones.Attitude{
			Roll:  float64(msg.Roll * 180 / pi),
			Pitch: float64(msg.Pitch * 180 / pi),
			Yaw:   float64(msg.Yaw * 180 / pi),
		}
		h.drone.SetAttitude(att)
	}
}

func (h *MAVLinkHandler) handleBattery(msg *common.MessageBatteryStatus, incoming *IncomingMessage) {
	var voltage float64
	if len(msg.Voltages) > 0 && msg.Voltages[0] != 65535 {
		voltage = float64(msg.Voltages[0]) / 1000.0
	}

	data := &BatteryData{
		SystemID:    incoming.SystemID,
		ComponentID: incoming.ComponentID,
		Voltage:     voltage,
		Current:     float64(msg.CurrentBattery) / 100.0,
		Remaining:   msg.BatteryRemaining,
	}

	h.batteryChan <- data

	if h.drone != nil {
		bat := Drones.BatteryStatus{
			Voltage:   voltage,
			Current:   float64(msg.CurrentBattery) / 100.0,
			Remaining: int(msg.BatteryRemaining),
		}
		h.drone.SetBattery(bat)
	}
}

func (h *MAVLinkHandler) handleGlobalPosition(msg *common.MessageGlobalPositionInt, incoming *IncomingMessage) {
	if h.drone != nil {
		pos := Drones.Position{
			Latitude:  float64(msg.Lat) / 1e7,
			Longitude: float64(msg.Lon) / 1e7,
			Altitude:  float64(msg.Alt) / 1000.0,
		}
		h.drone.SetPosition(pos)
	}
}

func (h *MAVLinkHandler) handleVfrHud(msg *common.MessageVfrHud, incoming *IncomingMessage) {
	if h.drone != nil {
		vel := Drones.Velocity{
			X: float64(msg.Groundspeed),
			Y: float64(msg.Airspeed),
			Z: 0,
		}
		h.drone.SetVelocity(vel)
	}
}

func (h *MAVLinkHandler) processLoop() {
	for {
		select {
		case <-h.stopChan:
			return
		case msg := <-h.messageChan:
			h.processMessage(msg)
		case hb := <-h.heartbeatChan:
			h.processHeartbeat(hb)
		case gps := <-h.GPSChan:
			h.processGPS(gps)
		case att := <-h.attitudeChan:
			h.processAttitude(att)
		case bat := <-h.batteryChan:
			h.processBattery(bat)
		}
	}
}

func (h *MAVLinkHandler) processMessage(msg *IncomingMessage) {
}

func (h *MAVLinkHandler) processHeartbeat(data *HeartbeatData) {
}

func (h *MAVLinkHandler) processGPS(data *GPSData) {
}

func (h *MAVLinkHandler) processAttitude(data *AttitudeData) {
}

func (h *MAVLinkHandler) processBattery(data *BatteryData) {
}

func (h *MAVLinkHandler) SendMessage(msg message.Message) error {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if !h.started || h.node == nil {
		return fmt.Errorf("MAVLink handler not started")
	}

	return h.node.WriteMessageAll(msg)
}

func (h *MAVLinkHandler) getTargetSysComp() (uint8, uint8) {
	targetSys := uint8(h.config.SystemID)
	if targetSys == 0 {
		targetSys = 1
	}
	targetComp := uint8(h.config.ComponentID)
	if targetComp == 0 {
		targetComp = 1
	}
	return targetSys, targetComp
}

func (h *MAVLinkHandler) SendCommandLong(targetSys, targetComp uint8, command common.MAV_CMD, params [7]float32) error {
	msg := &common.MessageCommandLong{
		TargetSystem:    targetSys,
		TargetComponent: targetComp,
		Command:         command,
		Param1:          params[0],
		Param2:          params[1],
		Param3:          params[2],
		Param4:          params[3],
		Param5:          params[4],
		Param6:          params[5],
		Param7:          params[6],
		Confirmation:    0,
	}

	return h.SendMessage(msg)
}

func (h *MAVLinkHandler) ArmDisarm(arm bool) error {
	var param float32
	if arm {
		param = 1.0
	} else {
		param = 0.0
	}

	targetSys, targetComp := h.getTargetSysComp()

	return h.SendCommandLong(
		targetSys,
		targetComp,
		common.MAV_CMD(400),
		[7]float32{param, 0, 0, 0, 0, 0, 0},
	)
}

func (h *MAVLinkHandler) Takeoff(altitude float32) error {
	targetSys, targetComp := h.getTargetSysComp()

	return h.SendCommandLong(
		targetSys,
		targetComp,
		common.MAV_CMD(22),
		[7]float32{0, 0, 0, 0, 0, altitude, 0},
	)
}

func (h *MAVLinkHandler) Land(lat, lon, alt float32) error {
	targetSys, targetComp := h.getTargetSysComp()

	return h.SendCommandLong(
		targetSys,
		targetComp,
		common.MAV_CMD(23),
		[7]float32{0, 0, 0, 0, lat, alt, lon},
	)
}

func (h *MAVLinkHandler) ReturnToHome() error {
	targetSys, targetComp := h.getTargetSysComp()

	return h.SendCommandLong(
		targetSys,
		targetComp,
		common.MAV_CMD(20),
		[7]float32{0, 0, 0, 0, 0, 0, 0},
	)
}

func (h *MAVLinkHandler) SetMode(mode uint32) error {
	targetSys, targetComp := h.getTargetSysComp()

	return h.SendCommandLong(
		targetSys,
		targetComp,
		common.MAV_CMD(11),
		[7]float32{float32(mode), 0, 0, 0, 0, 0, 0},
	)
}

func (h *MAVLinkHandler) RequestMessage(messageID uint8, rate uint16) error {
	targetSys, targetComp := h.getTargetSysComp()

	msg := &common.MessageRequestDataStream{
		TargetSystem:    targetSys,
		TargetComponent: targetComp,
		ReqStreamId:     uint8(messageID),
		ReqMessageRate:  rate,
		StartStop:       1,
	}

	return h.SendMessage(msg)
}

// 获取消息通道
func (h *MAVLinkHandler) GetMessageChan() <-chan *IncomingMessage {
	return h.messageChan
}

// 获取心跳通道
func (h *MAVLinkHandler) GetHeartbeatChan() <-chan *HeartbeatData {
	return h.heartbeatChan
}

// 获取GPS通道
func (h *MAVLinkHandler) GetGPSChan() <-chan *GPSData {
	return h.GPSChan
}

// 获取姿态通道
func (h *MAVLinkHandler) GetAttitudeChan() <-chan *AttitudeData {
	return h.attitudeChan
}

// 获取电池通道
func (h *MAVLinkHandler) GetBatteryChan() <-chan *BatteryData {
	return h.batteryChan
}

// 获取本handlerID(本次route请求ID)
func (h *MAVLinkHandler) GetHandlerID() string {
	return h.handlerID
}

// 获取无人机(对象)
func (h *MAVLinkHandler) GetDrone() *Drones.Drone {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.drone
}

// 设置无人机(对象)
func (h *MAVLinkHandler) SetDrone(drone *Drones.Drone) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.drone = drone
}

// 获取MAVLink配置
func (h *MAVLinkHandler) GetConfig() MAVLinkConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config
}

// 获取无人机连接状态
func (h *MAVLinkHandler) IsConnected() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.drone != nil {
		return h.drone.IsConnected()
	}
	return false
}

// 获取无人机状态
func (h *MAVLinkHandler) GetDroneStatus() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.drone != nil {
		return string(h.drone.GetStatus())
	}
	return "unknown"
}

// 获取无人机位置信息
func (h *MAVLinkHandler) GetDronePosition() Drones.Position {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.drone != nil {
		return h.drone.GetPosition()
	}
	return Drones.Position{}
}

// 获取无人机姿态信息
func (h *MAVLinkHandler) GetDroneAttitude() Drones.Attitude {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.drone != nil {
		return h.drone.GetAttitude()
	}
	return Drones.Attitude{}
}

// 获取无人机电池信息
func (h *MAVLinkHandler) GetDroneBattery() Drones.BatteryStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if h.drone != nil {
		return h.drone.GetBattery()
	}
	return Drones.BatteryStatus{}
}

// 重启handler
func (h *MAVLinkHandler) Restart() error {
	if err := h.Stop(); err != nil {
		return err
	}
	time.Sleep(100 * time.Millisecond)
	return h.Start()
}

// 手动发送心跳
func (h *MAVLinkHandler) SendHeartbeat() error {
	msg := &common.MessageHeartbeat{
		Type:         common.MAV_TYPE(0),
		Autopilot:    common.MAV_AUTOPILOT(0),
		BaseMode:     common.MAV_MODE_FLAG(0),
		SystemStatus: common.MAV_STATE(0),
		CustomMode:   0,
	}
	return h.SendMessage(msg)
}

func (h *MAVLinkHandler) SetGroundStation(info *GroundStationInfo) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.groundStation = info
}

func (h *MAVLinkHandler) GetGroundStation() *GroundStationInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.groundStation
}

func (h *MAVLinkHandler) SetDispatcher(info *DispatcherInfo) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dispatcher = info
}

func (h *MAVLinkHandler) GetDispatcher() *DispatcherInfo {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.dispatcher
}

func (h *MAVLinkHandler) SetDispatcherByUser(username, email string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dispatcher = &DispatcherInfo{
		Type:     DispatcherTypeHuman,
		Username: username,
		Email:    email,
	}
}

func (h *MAVLinkHandler) SetDispatcherByAI() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.dispatcher = &DispatcherInfo{
		Type:     DispatcherTypeAI,
		Username: "AI-Agent",
		Email:    "",
	}
}

func (h *MAVLinkHandler) RecordTimestamp() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.timestampHistory = append(h.timestampHistory, time.Now())
	if len(h.timestampHistory) > 1000 {
		h.timestampHistory = h.timestampHistory[len(h.timestampHistory)-1000:]
	}
}

func (h *MAVLinkHandler) GetTimestampHistory() []time.Time {
	h.mu.RLock()
	defer h.mu.RUnlock()
	result := make([]time.Time, len(h.timestampHistory))
	copy(result, h.timestampHistory)
	return result
}

func (h *MAVLinkHandler) ClearTimestampHistory() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.timestampHistory = nil
}
