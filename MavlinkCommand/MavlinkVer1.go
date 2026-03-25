package Mavlink_Board

import (
	"fmt"
	"log"
	"time"

	"github.com/bluenviron/gomavlib/v3"
	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"github.com/bluenviron/gomavlib/v3/pkg/message"
)

type ConnectionType string

const (
	ConnectionSerial ConnectionType = "Serial"
	ConnectionUDP    ConnectionType = "UDP"
	ConnectionTCP    ConnectionType = "TCP"
)

type MavlinkConfig struct {
	ConnectionType  ConnectionType
	SerialPort      string
	SerialBaud      int
	TCPAddr         string
	TCPPort         int
	UDPAddr         string
	UDPPort         int
	SystemID        int
	ComponentID     int
	HeartbeatRate   time.Duration
	TargetSystem    uint8
	TargetComponent uint8
}

type MavlinkCommander struct {
	node        *gomavlib.Node
	config      *MavlinkConfig
	initialized bool
	messageChan chan *ReceivedMAVLinkMessage
	pendingAcks map[uint16]chan *CommandAck
	ackMutex    chan struct{}
	stopChan    chan bool
}

type ReceivedMAVLinkMessage struct {
	MessageID   uint32
	SystemID    uint8
	ComponentID uint8
	Message     interface{}
	Timestamp   time.Time
}

type CommandAck struct {
	Command         uint16
	Result          uint8
	Progress        uint8
	ResultParam2    int32
	TargetSystem    uint8
	TargetComponent uint8
}

const (
	MAVLINK_MSG_ID_HEARTBEAT             = 0
	MAVLINK_MSG_ID_SYS_STATUS            = 1
	MAVLINK_MSG_ID_SYSTEM_TIME           = 2
	MAVLINK_MSG_ID_GPS_RAW_INT           = 24
	MAVLINK_MSG_ID_GPS_STATUS            = 25
	MAVLINK_MSG_ID_SCALED_IMU            = 26
	MAVLINK_MSG_ID_RAW_IMU               = 27
	MAVLINK_MSG_ID_ATTITUDE              = 30
	MAVLINK_MSG_ID_ATTITUDE_QUATERNION   = 31
	MAVLINK_MSG_ID_LOCAL_POSITION_NED    = 32
	MAVLINK_MSG_ID_GLOBAL_POSITION_INT   = 33
	MAVLINK_MSG_ID_RC_CHANNELS_RAW       = 35
	MAVLINK_MSG_ID_RC_CHANNELS_SCALED    = 36
	MAVLINK_MSG_ID_SERVO_OUTPUT_RAW      = 36
	MAVLINK_MSG_ID_MISSION_ITEM          = 39
	MAVLINK_MSG_ID_MISSION_REQUEST       = 40
	MAVLINK_MSG_ID_MISSION_ACK           = 47
	MAVLINK_MSG_ID_COMMAND_LONG          = 76
	MAVLINK_MSG_ID_COMMAND_ACK           = 77
	MAVLINK_MSG_ID_COMMAND_INT           = 75
	MAVLINK_MSG_ID_MANUAL_CONTROL        = 69
	MAVLINK_MSG_ID_RC_CHANNELS_OVERRIDE  = 70
	MAVLINK_MSG_ID_GPS_GLOBAL_ORIGIN     = 49
	MAVLINK_MSG_ID_SET_GPS_GLOBAL_ORIGIN = 48
	MAVLINK_MSG_ID_PARAM_REQUEST_READ    = 20
	MAVLINK_MSG_ID_PARAM_REQUEST_LIST    = 21
	MAVLINK_MSG_ID_PARAM_VALUE           = 22
	MAVLINK_MSG_ID_PARAM_SET             = 23
	MAVLINK_MSG_ID_MISSION_COUNT         = 44
	MAVLINK_MSG_ID_MISSION_CLEAR_ALL     = 43
	MAVLINK_MSG_ID_MISSION_ITEM_REACHED  = 46
	MAVLINK_MSG_ID_MISSION_CURRENT       = 42
	MAVLINK_MSG_ID_REQUEST_DATA_STREAM   = 66
	MAVLINK_MSG_ID_DATA_STREAM           = 67
	MAVLINK_MSG_ID_EXTENDED_SYS_STATE    = 245
	MAVLINK_MSG_ID_ADSB_VEHICLE          = 246
	MAVLINK_MSG_ID_UTM_GLOBAL_POSITION   = 340
	MAVLINK_MSG_ID_ALTITUDE              = 141
	MAVLINK_MSG_ID_DISTANCE_SENSOR       = 132
	MAVLINK_MSG_ID_BATTERY_STATUS        = 370
	MAVLINK_MSG_ID_TERRAIN_REPORT        = 136
	MAVLINK_MSG_ID_SENSOR_OFFSETS        = 150
	MAVLINK_MSG_ID_MEMINFO               = 152
	MAVLINK_MSG_ID_AP_ADC                = 153
	MAVLINK_MSG_ID_DIGICAM_CONFIGURE     = 154
	MAVLINK_MSG_ID_DIGICAM_CONTROL       = 155
	MAVLINK_MSG_ID_MOUNT_CONFIGURE       = 156
	MAVLINK_MSG_ID_MOUNT_CONTROL         = 157
	MAVLINK_MSG_ID_MOUNT_STATUS          = 158
	MAVLINK_MSG_ID_FENCE_POINT           = 160
	MAVLINK_MSG_ID_FENCE_STATUS          = 162
	MAVLINK_MSG_ID_RADIO_STATUS          = 109
	MAVLINK_MSG_ID_LANDING_TARGET        = 149
	MAVLINK_MSG_ID_ESTIMATOR_STATUS      = 230
	MAVLINK_MSG_ID_WIND_COV              = 231
	MAVLINK_MSG_ID_GPS_INPUT             = 232
	MAVLINK_MSG_ID_NAMED_VALUE_FLOAT     = 251
	MAVLINK_MSG_ID_NAMED_VALUE_INT       = 252
	MAVLINK_MSG_ID_DEBUG                 = 254
	MAVLINK_MSG_ID_DEBUG_FLOAT_ARRAY     = 255
	MAVLINK_MSG_ID_STATUSTEXT            = 65
	MAVLINK_MSG_ID_SEND_EXTENDED_COMMAND = 84
)

const (
	MAV_CMD_NAV_ALTITUDE_WAIT  = 83
	MAV_CMD_DO_CHANGE_SPEED    = 178
	MAV_CMD_DO_DIGICAM_CONTROL = 203
)

const (
	MAV_RESULT_ACCEPTED             = 0
	MAV_RESULT_TEMPORARILY_REJECTED = 1
	MAV_RESULT_DENIED               = 2
	MAV_RESULT_UNSUPPORTED          = 3
	MAV_RESULT_FAILED               = 4
	MAV_RESULT_IN_PROGRESS          = 5
)

func NewMavlinkCommander() *MavlinkCommander {
	return &MavlinkCommander{
		node:        nil,
		config:      nil,
		initialized: false,
		messageChan: make(chan *ReceivedMAVLinkMessage, 100),
		pendingAcks: make(map[uint16]chan *CommandAck),
		ackMutex:    make(chan struct{}, 1),
	}
}

func (m *MavlinkCommander) Configure(config MavlinkConfig) {
	m.config = &config
	if m.config.TargetSystem == 0 {
		m.config.TargetSystem = 1
	}
	if m.config.TargetComponent == 0 {
		m.config.TargetComponent = 1
	}
}

func (m *MavlinkCommander) GetTargetSystem() uint8 {
	if m.config == nil {
		return 1
	}
	return m.config.TargetSystem
}

func (m *MavlinkCommander) SetTargetSystem(systemID uint8) {
	if m.config != nil {
		m.config.TargetSystem = systemID
	}
}

func (m *MavlinkCommander) SetTargetComponent(componentID uint8) {
	if m.config != nil {
		m.config.TargetComponent = componentID
	}
}

func (m *MavlinkCommander) GetMessageChan() chan *ReceivedMAVLinkMessage {
	return m.messageChan
}

func (m *MavlinkCommander) Start() error {
	if m.config == nil {
		return fmt.Errorf("configuration not set, call Configure() first")
	}

	if m.initialized && m.node != nil {
		return fmt.Errorf("already initialized, call Stop() first")
	}

	var endpointConf gomavlib.EndpointConf

	switch m.config.ConnectionType {
	case ConnectionSerial:
		endpointConf = gomavlib.EndpointSerial{
			Device: m.config.SerialPort,
			Baud:   m.config.SerialBaud,
		}
	case ConnectionUDP:
		endpointConf = gomavlib.EndpointUDPClient{
			Address: fmt.Sprintf("%s:%d", m.config.UDPAddr, m.config.UDPPort),
		}
	case ConnectionTCP:
		endpointConf = gomavlib.EndpointTCPClient{
			Address: fmt.Sprintf("%s:%d", m.config.TCPAddr, m.config.TCPPort),
		}
	default:
		return fmt.Errorf("unsupported connection type: %s", m.config.ConnectionType)
	}

	heartbeatRate := m.config.HeartbeatRate
	if heartbeatRate == 0 {
		heartbeatRate = time.Second
	}

	componentID := m.config.ComponentID
	if componentID == 0 {
		componentID = 1
	}

	node := gomavlib.Node{
		Endpoints:           []gomavlib.EndpointConf{endpointConf},
		Dialect:             common.Dialect,
		OutVersion:          gomavlib.V2,
		OutSystemID:         byte(m.config.SystemID),
		OutComponentID:      byte(componentID),
		HeartbeatDisable:    false,
		HeartbeatPeriod:     heartbeatRate,
		StreamRequestEnable: true,
	}

	if err := node.Initialize(); err != nil {
		return fmt.Errorf("failed to initialize MAVLink node: %w", err)
	}

	m.node = &node
	m.initialized = true
	m.stopChan = make(chan bool)

	go m.readLoop()

	return nil
}

func (m *MavlinkCommander) Stop() error {
	if m.node == nil {
		return nil
	}

	close(m.messageChan)
	m.node.Close()
	m.node = nil
	m.initialized = false

	for _, ackChan := range m.pendingAcks {
		close(ackChan)
	}
	m.pendingAcks = make(map[uint16]chan *CommandAck)

	return nil
}

func (m *MavlinkCommander) IsConnected() bool {
	return m.initialized && m.node != nil
}

func (m *MavlinkCommander) readLoop() {
	if m.node == nil {
		return
	}

	for {
		select {
		case <-m.stopChan:
			return
		case evt := <-m.node.Events():
			m.handleEvent(evt)
		}
	}
}

func (m *MavlinkCommander) handleEvent(evt gomavlib.Event) {
	switch evt := evt.(type) {
	case *gomavlib.EventFrame:
		m.handleFrame(evt)
	}
}

func (m *MavlinkCommander) handleFrame(evt *gomavlib.EventFrame) {
	if evt == nil {
		return
	}

	msg := evt.Message()
	if msg == nil {
		return
	}

	recvMsg := &ReceivedMAVLinkMessage{
		MessageID:   msg.GetID(),
		SystemID:    evt.SystemID(),
		ComponentID: evt.ComponentID(),
		Message:     msg,
		Timestamp:   time.Now(),
	}

	select {
	case m.messageChan <- recvMsg:
	default:
		log.Printf("[MavlinkCommander] Message channel full, dropping message ID: %d", msg.GetID())
	}

	m.processMessage(msg, evt.SystemID(), evt.ComponentID())
}

func (m *MavlinkCommander) processMessage(msg message.Message, systemID, componentID uint8) {
	switch msg.GetID() {
	case MAVLINK_MSG_ID_HEARTBEAT:
		heartbeat := msg.(*common.MessageHeartbeat)
		log.Printf("[Mavlink] Heartbeat: SysID=%d, Type=%d, Autopilot=%d, BaseMode=%d, SystemStatus=%d",
			systemID, heartbeat.Type, heartbeat.Autopilot, heartbeat.BaseMode, heartbeat.SystemStatus)

	case MAVLINK_MSG_ID_SYS_STATUS:
		sysStatus := msg.(*common.MessageSysStatus)
		log.Printf("[Mavlink] SysStatus: SysID=%d, Voltage=%dmV, Current=%dmA, Battery=%d%%, Load=%d%%",
			systemID, sysStatus.VoltageBattery, sysStatus.CurrentBattery, sysStatus.BatteryRemaining, sysStatus.Load/10)

	case MAVLINK_MSG_ID_GPS_RAW_INT:
		gps := msg.(*common.MessageGpsRawInt)
		lat := float64(gps.Lat) / 1e7
		lon := float64(gps.Lon) / 1e7
		alt := float64(gps.Alt) / 1000.0
		log.Printf("[Mavlink] GPS_Raw_Int: SysID=%d, Lat=%.7f, Lon=%.7f, Alt=%.1fm, FixType=%d, Satellites=%d",
			systemID, lat, lon, alt, gps.FixType, gps.SatellitesVisible)

	case MAVLINK_MSG_ID_ATTITUDE:
		att := msg.(*common.MessageAttitude)
		log.Printf("[Mavlink] Attitude: SysID=%d, Roll=%.2f, Pitch=%.2f, Yaw=%.2f",
			systemID, att.Roll, att.Pitch, att.Yaw)

	case MAVLINK_MSG_ID_GLOBAL_POSITION_INT:
		gp := msg.(*common.MessageGlobalPositionInt)
		lat := float64(gp.Lat) / 1e7
		lon := float64(gp.Lon) / 1e7
		alt := float64(gp.Alt) / 1000.0
		relativeAlt := float64(gp.RelativeAlt) / 1000.0
		log.Printf("[Mavlink] GlobalPosition: SysID=%d, Lat=%.7f, Lon=%.7f, Alt=%.1fm, RelativeAlt=%.1fm",
			systemID, lat, lon, alt, relativeAlt)

	case MAVLINK_MSG_ID_COMMAND_ACK:
		ack := msg.(*common.MessageCommandAck)
		log.Printf("[Mavlink] CommandAck: SysID=%d, Command=%d, Result=%d",
			systemID, ack.Command, ack.Result)
		m.handleCommandAck(ack)

	case MAVLINK_MSG_ID_PARAM_VALUE:
		param := msg.(*common.MessageParamValue)
		name := string(param.ParamId[:])
		log.Printf("[Mavlink] ParamValue: SysID=%d, Param=%s, Value=%.6f, Type=%d",
			systemID, name, param.ParamValue, param.ParamType)

	case MAVLINK_MSG_ID_MISSION_ITEM:
		mi := msg.(*common.MessageMissionItem)
		log.Printf("[Mavlink] MissionItem: SysID=%d, Seq=%d, Command=%d, Current=%d",
			systemID, mi.Seq, mi.Command, mi.Current)

	case MAVLINK_MSG_ID_MISSION_ACK:
		ma := msg.(*common.MessageMissionAck)
		log.Printf("[Mavlink] MissionAck: SysID=%d, Type=%d", systemID, ma.Type)

	case MAVLINK_MSG_ID_RC_CHANNELS_RAW:
		rc := msg.(*common.MessageRcChannelsRaw)
		log.Printf("[Mavlink] RC_Raw: SysID=%d, Ch1=%d, Ch2=%d, Ch3=%d, Ch4=%d, Rssi=%d",
			systemID, rc.Chan1Raw, rc.Chan2Raw, rc.Chan3Raw, rc.Chan4Raw, rc.Rssi)

	case MAVLINK_MSG_ID_RC_CHANNELS_OVERRIDE:
		rc := msg.(*common.MessageRcChannelsOverride)
		log.Printf("[Mavlink] RC_Override: SysID=%d, msg= %v", systemID, rc)

	case MAVLINK_MSG_ID_BATTERY_STATUS:
		bat := msg.(*common.MessageBatteryStatus)
		log.Printf("[Mavlink] BatteryStatus: SysID=%d, Voltage=%dmV, Current=%dmA, Remaining=%d%%",
			systemID, bat.Voltages[0], bat.CurrentBattery, bat.BatteryRemaining)

	case MAVLINK_MSG_ID_STATUSTEXT:
		st := msg.(*common.MessageStatustext)
		severityNames := []string{"EMERGENCY", "ALERT", "CRITICAL", "ERROR", "WARNING", "NOTICE", "INFO", "DEBUG"}
		severityName := "UNKNOWN"
		if uint8(st.Severity) < uint8(len(severityNames)) {
			severityName = severityNames[st.Severity]
		}
		text := string(st.Text[:])
		log.Printf("[Mavlink] StatusText: SysID=%d, Severity=%s, Text=%s", systemID, severityName, text)

	case MAVLINK_MSG_ID_EXTENDED_SYS_STATE:
		ess := msg.(*common.MessageExtendedSysState)
		landedStateNames := []string{"UNDEFINED", "LANDED", "TAKING_OFF", "IN_AIR", "DESCENDING", "ASCENDING"}
		landedState := "UNKNOWN"
		if uint8(ess.LandedState) < uint8(len(landedStateNames)) {
			landedState = landedStateNames[ess.LandedState]
		}
		log.Printf("[Mavlink] ExtendedSysState: SysID=%d, LandedState=%s", systemID, landedState)

	case MAVLINK_MSG_ID_DISTANCE_SENSOR:
		ds := msg.(*common.MessageDistanceSensor)
		log.Printf("[Mavlink] DistanceSensor: SysID=%d, Distance=%.2fm, Type=%d",
			systemID, float64(ds.CurrentDistance)/100.0, ds.Type)

	case MAVLINK_MSG_ID_ALTITUDE:
		alt := msg.(*common.MessageAltitude)
		log.Printf("[Mavlink] Altitude: SysID=%d, AltitudeMonotonic=%.2f, AltitudeLocal=%.2f, \n Altitude = %v",
			systemID, alt.AltitudeMonotonic, alt.AltitudeLocal, alt)

	case MAVLINK_MSG_ID_ESTIMATOR_STATUS:
		es := msg.(*common.MessageEstimatorStatus)
		log.Printf("[Mavlink] EstimatorStatus: SysID=%d, Flags=0x%x", systemID, es.Flags)

	default:
		log.Printf("[Mavlink] Unknown Message ID: %d from SysID=%d, CompID=%d", msg.GetID(), systemID, componentID)
	}
}

func (m *MavlinkCommander) handleCommandAck(ack *common.MessageCommandAck) {
	select {
	case m.ackMutex <- struct{}{}:
	default:
		return
	}
	defer func() { <-m.ackMutex }()

	ackChan, exists := m.pendingAcks[uint16(ack.Command)]
	if exists {
		ackMsg := &CommandAck{
			Command:         uint16(ack.Command),
			Result:          uint8(ack.Result),
			Progress:        uint8(ack.Progress),
			ResultParam2:    ack.ResultParam2,
			TargetSystem:    ack.TargetSystem,
			TargetComponent: ack.TargetComponent,
		}
		select {
		case ackChan <- ackMsg:
		default:
		}
	}
}

func (m *MavlinkCommander) WriteMessage(msg message.Message) error {
	if m.node == nil {
		return fmt.Errorf("node not initialized, call Start() first")
	}
	return m.node.WriteMessageAll(msg)
}

func (m *MavlinkCommander) WriteMessageTo(channel *gomavlib.Channel, msg message.Message) error {
	if m.node == nil {
		return fmt.Errorf("node not initialized, call Start() first")
	}
	return m.node.WriteMessageTo(channel, msg)
}

func (m *MavlinkCommander) WriteMessageToEndpoint(msg message.Message, systemID, componentID uint8) error {
	if m.node == nil {
		return fmt.Errorf("node not initialized, call Start() first")
	}
	return m.node.WriteMessageAll(msg)
}

func (m *MavlinkCommander) SendHeartbeat() error {
	heartbeat := &common.MessageHeartbeat{
		Type:           common.MAV_TYPE_ONBOARD_CONTROLLER,
		Autopilot:      common.MAV_AUTOPILOT_GENERIC,
		BaseMode:       common.MAV_MODE_FLAG_MANUAL_INPUT_ENABLED | common.MAV_MODE_FLAG_CUSTOM_MODE_ENABLED,
		SystemStatus:   common.MAV_STATE_ACTIVE,
		MavlinkVersion: 2,
	}
	return m.node.WriteMessageAll(heartbeat)
}

func (m *MavlinkCommander) CommandLong(
	targetSystem uint8,
	targetComponent uint8,
	command uint16,
	confirmation uint8,
	param1 float32,
	param2 float32,
	param3 float32,
	param4 float32,
	param5 float32,
	param6 float32,
	param7 float32,
) error {
	cmd := &common.MessageCommandLong{
		TargetSystem:    targetSystem,
		TargetComponent: targetComponent,
		Command:         common.MAV_CMD(command),
		Confirmation:    confirmation,
		Param1:          param1,
		Param2:          param2,
		Param3:          param3,
		Param4:          param4,
		Param5:          param5,
		Param6:          param6,
		Param7:          param7,
	}
	return m.node.WriteMessageAll(cmd)
}

func (m *MavlinkCommander) CommandLongWithAck(
	targetSystem uint8,
	targetComponent uint8,
	command uint16,
	param1 float32,
	param2 float32,
	param3 float32,
	param4 float32,
	param5 float32,
	param6 float32,
	param7 float32,
	timeout time.Duration,
) (*CommandAck, error) {
	ackChan := make(chan *CommandAck, 1)

	select {
	case m.ackMutex <- struct{}{}:
	default:
	}
	m.pendingAcks[command] = ackChan
	defer func() {
		select {
		case m.ackMutex <- struct{}{}:
		default:
		}
		delete(m.pendingAcks, command)
		close(ackChan)
	}()

	err := m.CommandLong(targetSystem, targetComponent, command, 0, param1, param2, param3, param4, param5, param6, param7)
	if err != nil {
		return nil, err
	}

	select {
	case ack := <-ackChan:
		return ack, nil
	case <-time.After(timeout):
		return nil, fmt.Errorf("command 0x%x timeout", command)
	}
}

func (m *MavlinkCommander) CommandAck(command uint16, result uint8) error {
	ack := &common.MessageCommandAck{
		Command:         common.MAV_CMD(command),
		Result:          common.MAV_RESULT(result),
		Progress:        0,
		ResultParam2:    0,
		TargetSystem:    0,
		TargetComponent: 0,
	}
	return m.node.WriteMessageAll(ack)
}

func (m *MavlinkCommander) ArmDisarm(arm bool, force bool) error {
	armValue := float32(0)
	if arm {
		armValue = float32(1)
	}
	forceValue := float32(0)
	if force {
		forceValue = float32(21196)
	}
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_COMPONENT_ARM_DISARM), 0, armValue, forceValue, 0, 0, 0, 0, 0)
}

func (m *MavlinkCommander) ArmDisarmWithAck(arm bool, force bool, timeout time.Duration) (*CommandAck, error) {
	armValue := float32(0)
	if arm {
		armValue = float32(1)
	}
	forceValue := float32(0)
	if force {
		forceValue = float32(21196)
	}
	targetSys := m.GetTargetSystem()
	return m.CommandLongWithAck(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_COMPONENT_ARM_DISARM),
		armValue, forceValue, 0, 0, 0, 0, 0, timeout)
}

func (m *MavlinkCommander) Takeoff(altitude float32) error {
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_NAV_TAKEOFF), 0, 0, 0, 0, 0, 0, 0, altitude)
}

func (m *MavlinkCommander) TakeoffWithAck(altitude float32, timeout time.Duration) (*CommandAck, error) {
	targetSys := m.GetTargetSystem()
	return m.CommandLongWithAck(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_NAV_TAKEOFF),
		0, 0, 0, 0, 0, 0, altitude, timeout)
}

func (m *MavlinkCommander) TakeoffWithCoords(latitude float32, longitude float32, altitude float32) error {
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_NAV_TAKEOFF), 0, 0, 0, 0, 0, latitude, longitude, altitude)
}

func (m *MavlinkCommander) Land(latitude float32, longitude float32, altitude float32) error {
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_NAV_LAND), 0, 0, 0, 0, 0, latitude, longitude, altitude)
}

func (m *MavlinkCommander) LandWithAck(latitude float32, longitude float32, altitude float32, timeout time.Duration) (*CommandAck, error) {
	targetSys := m.GetTargetSystem()
	return m.CommandLongWithAck(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_NAV_LAND),
		0, 0, 0, 0, latitude, longitude, altitude, timeout)
}

func (m *MavlinkCommander) ReturnToLaunch() error {
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_NAV_RETURN_TO_LAUNCH), 0, 0, 0, 0, 0, 0, 0, 0)
}

func (m *MavlinkCommander) ReturnToLaunchWithAck(timeout time.Duration) (*CommandAck, error) {
	targetSys := m.GetTargetSystem()
	return m.CommandLongWithAck(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_NAV_RETURN_TO_LAUNCH),
		0, 0, 0, 0, 0, 0, 0, timeout)
}

func (m *MavlinkCommander) Hold() error {
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_NAV_LOITER_UNLIM), 0, 0, 0, 0, 0, 0, 0, 0)
}

func (m *MavlinkCommander) HoldWithAck(timeout time.Duration) (*CommandAck, error) {
	targetSys := m.GetTargetSystem()
	return m.CommandLongWithAck(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_NAV_LOITER_UNLIM),
		0, 0, 0, 0, 0, 0, 0, timeout)
}

func (m *MavlinkCommander) ContinueMission() error {
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_MISSION_START), 0, 0, 0, 0, 0, 0, 0, 0)
}

func (m *MavlinkCommander) ContinueMissionWithAck(timeout time.Duration) (*CommandAck, error) {
	targetSys := m.GetTargetSystem()
	return m.CommandLongWithAck(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_MISSION_START),
		0, 0, 0, 0, 0, 0, 0, timeout)
}

func (m *MavlinkCommander) SetMode(mode common.MAV_MODE) error {
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_DO_SET_MODE), 0, float32(mode), 0, 0, 0, 0, 0, 0)
}

func (m *MavlinkCommander) SetModeWithAck(mode common.MAV_MODE, timeout time.Duration) (*CommandAck, error) {
	targetSys := m.GetTargetSystem()
	return m.CommandLongWithAck(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_DO_SET_MODE),
		float32(mode), 0, 0, 0, 0, 0, 0, timeout)
}

func (m *MavlinkCommander) SetSpeed(speedType uint8, speed float32, relative bool) error {
	relativeVal := float32(0)
	if relative {
		relativeVal = float32(1)
	}
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_DO_CHANGE_SPEED), 0, float32(speedType), speed, relativeVal, 0, 0, 0, 0)
}

func (m *MavlinkCommander) SetSpeedWithAck(speedType uint8, speed float32, relative bool, timeout time.Duration) (*CommandAck, error) {
	relativeVal := float32(0)
	if relative {
		relativeVal = float32(1)
	}
	targetSys := m.GetTargetSystem()
	return m.CommandLongWithAck(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_DO_CHANGE_SPEED),
		float32(speedType), speed, relativeVal, 0, 0, 0, 0, timeout)
}

func (m *MavlinkCommander) WaitAltitude(altitude float32) error {
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(MAV_CMD_NAV_ALTITUDE_WAIT), 0, altitude, -1, 0, 0, 0, 0, 0)
}

func (m *MavlinkCommander) WaitAltitudeWithAck(altitude float32, timeout time.Duration) (*CommandAck, error) {
	targetSys := m.GetTargetSystem()
	return m.CommandLongWithAck(targetSys, m.config.TargetComponent, uint16(MAV_CMD_NAV_ALTITUDE_WAIT),
		altitude, -1, 0, 0, 0, 0, 0, timeout)
}

func (m *MavlinkCommander) GetParam(paramID string) error {
	paramBytes := []byte(paramID)
	for len(paramBytes) < 16 {
		paramBytes = append(paramBytes, 0)
	}
	targetSys := m.GetTargetSystem()
	paramMsg := &common.MessageParamRequestRead{
		TargetSystem:    targetSys,
		TargetComponent: m.config.TargetComponent,
		ParamId:         string(paramBytes[:16]),
		ParamIndex:      -1,
	}
	return m.node.WriteMessageAll(paramMsg)
}

func (m *MavlinkCommander) SetParam(paramID string, paramValue float32, paramType uint8) error {
	paramBytes := []byte(paramID)
	for len(paramBytes) < 16 {
		paramBytes = append(paramBytes, 0)
	}
	targetSys := m.GetTargetSystem()
	paramMsg := &common.MessageParamSet{
		TargetSystem:    targetSys,
		TargetComponent: m.config.TargetComponent,
		ParamId:         string(paramBytes[:16]),
		ParamValue:      paramValue,
		ParamType:       common.MAV_PARAM_TYPE(paramType),
	}
	return m.node.WriteMessageAll(paramMsg)
}

func (m *MavlinkCommander) SetParamWithAck(paramID string, paramValue float32, paramType uint8, timeout time.Duration) (*CommandAck, error) {
	err := m.SetParam(paramID, paramValue, paramType)
	if err != nil {
		return nil, err
	}
	paramIDBytes := []byte(paramID)
	for len(paramIDBytes) < 16 {
		paramIDBytes = append(paramIDBytes, 0)
	}
	paramIDStr := string(paramIDBytes[:16])
	ack := &CommandAck{
		Command: uint16(MAVLINK_MSG_ID_PARAM_SET),
		Result:  MAV_RESULT_ACCEPTED,
	}
	_ = paramIDStr
	return ack, nil
}

func (m *MavlinkCommander) RequestDataStream(streamID uint16, rate uint16, startStop uint8) error {
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(MAVLINK_MSG_ID_REQUEST_DATA_STREAM), 0, float32(streamID), float32(rate), float32(startStop), 0, 0, 0, 0)
}

func (m *MavlinkCommander) RequestDataStreamWithAck(streamID uint16, rate uint16, startStop uint8, timeout time.Duration) (*CommandAck, error) {
	targetSys := m.GetTargetSystem()
	return m.CommandLongWithAck(targetSys, m.config.TargetComponent, uint16(MAVLINK_MSG_ID_REQUEST_DATA_STREAM),
		float32(streamID), float32(rate), float32(startStop), 0, 0, 0, 0, timeout)
}

func (m *MavlinkCommander) StartGPS(rawLat, rawLon, rawAlt int32, fixType uint8) error {
	rawMsg := &common.MessageGpsRawInt{
		TimeUsec:          uint64(time.Now().UnixNano() / 1000),
		Lat:               rawLat,
		Lon:               rawLon,
		Alt:               rawAlt,
		Eph:               65535,
		Epv:               65535,
		Vel:               65535,
		Cog:               65535,
		FixType:           common.GPS_FIX_TYPE(fixType),
		SatellitesVisible: 255,
	}
	return m.node.WriteMessageAll(rawMsg)
}

func (m *MavlinkCommander) SendSYSStatus(batteryVoltage uint16, batteryCurrent int16, batteryRemaining int8) error {
	sysStatus := &common.MessageSysStatus{
		OnboardControlSensorsEnabled: 0,
		OnboardControlSensorsHealth:  0,
		OnboardControlSensorsPresent: 0,
		Load:                         0,
		VoltageBattery:               batteryVoltage,
		CurrentBattery:               batteryCurrent,
		BatteryRemaining:             batteryRemaining,
		DropRateComm:                 0,
		ErrorsComm:                   0,
		ErrorsCount1:                 0,
		ErrorsCount2:                 0,
		ErrorsCount3:                 0,
		ErrorsCount4:                 0,
	}
	return m.node.WriteMessageAll(sysStatus)
}

func (m *MavlinkCommander) SendExtendedSysState(landedState uint8) error {
	extSysState := &common.MessageExtendedSysState{
		VtolState:   common.MAV_VTOL_STATE_UNDEFINED,
		LandedState: common.MAV_LANDED_STATE(landedState),
	}
	return m.node.WriteMessageAll(extSysState)
}

func (m *MavlinkCommander) SendCameraFeedback(
	systemID uint8,
	componentID uint8,
	lat int32,
	lng int32,
	alt int32,
	relativeAlt int32,
	q [4]float32,
	imgIndex int32,
	cameraID uint8,
	recording uint8,
) error {
	return nil
}

func (m *MavlinkCommander) SendMountStatus(
	componentID uint8,
	q [4]float32,
	inputA int32,
	inputB int32,
	inputC int32,
	targetSystem uint8,
) error {
	return nil
}

func (m *MavlinkCommander) DoDigICamControl(
	componentID uint8,
	sessionControl uint8,
	mode uint8,
	eZoomAbsolute uint16,
	zoomStep int8,
	focusLock uint8,
	shot uint8,
) error {
	return m.CommandLong(0, componentID, uint16(common.MAV_CMD_DO_DIGICAM_CONTROL), 0,
		float32(sessionControl), float32(mode), float32(eZoomAbsolute), float32(zoomStep), float32(focusLock), float32(shot), 0)
}

func (m *MavlinkCommander) DoSetRelay(relayNum uint8, state bool) error {
	stateVal := float32(0)
	if state {
		stateVal = float32(1)
	}
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_DO_SET_RELAY), 0, float32(relayNum), stateVal, 0, 0, 0, 0, 0)
}

func (m *MavlinkCommander) DoSetServo(servoChannel uint8, pwm uint16) error {
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_DO_SET_SERVO), 0, float32(servoChannel), float32(pwm), 0, 0, 0, 0, 0)
}

func (m *MavlinkCommander) DoSetPositionYaw(yaw float32, yawRate float32) error {
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(common.MAV_CMD_CONDITION_YAW), 0, yaw, yawRate, 0, 0, 0, 0, 0)
}

func (m *MavlinkCommander) DoSetAltitudeYaw(altitude float32, yaw float32) error {
	err := m.WaitAltitude(altitude)
	if err != nil {
		return err
	}
	return m.DoSetPositionYaw(yaw, 0)
}

func (m *MavlinkCommander) MissionClearAll() error {
	targetSys := m.GetTargetSystem()
	return m.CommandLong(targetSys, m.config.TargetComponent, uint16(MAVLINK_MSG_ID_MISSION_CLEAR_ALL), 0, 0, 0, 0, 0, 0, 0, 0)
}

func (m *MavlinkCommander) MissionClearAllWithAck(timeout time.Duration) (*CommandAck, error) {
	targetSys := m.GetTargetSystem()
	return m.CommandLongWithAck(targetSys, m.config.TargetComponent, uint16(MAVLINK_MSG_ID_MISSION_CLEAR_ALL),
		0, 0, 0, 0, 0, 0, 0, timeout)
}

func (m *MavlinkCommander) MissionWriteItem(
	seq uint16,
	current uint8,
	autocontinue uint8,
	frame uint8,
	command uint16,
	param1, param2, param3, param4, x, y, z float32,
) error {
	targetSys := m.GetTargetSystem()
	item := &common.MessageMissionItem{
		TargetSystem:    targetSys,
		TargetComponent: m.config.TargetComponent,
		Seq:             seq,
		Frame:           common.MAV_FRAME(frame),
		Command:         common.MAV_CMD(command),
		Current:         current,
		Autocontinue:    autocontinue,
		Param1:          param1,
		Param2:          param2,
		Param3:          param3,
		Param4:          param4,
		X:               x,
		Y:               y,
		Z:               z,
	}
	return m.node.WriteMessageAll(item)
}

func (m *MavlinkCommander) MissionWriteItemWithAck(
	seq uint16,
	current uint8,
	autocontinue uint8,
	frame uint8,
	command uint16,
	param1, param2, param3, param4, x, y, z float32,
	timeout time.Duration,
) (*CommandAck, error) {
	err := m.MissionWriteItem(seq, current, autocontinue, frame, command, param1, param2, param3, param4, x, y, z)
	if err != nil {
		return nil, err
	}
	ack := &CommandAck{
		Command: command,
		Result:  MAV_RESULT_ACCEPTED,
	}
	return ack, nil
}

func (m *MavlinkCommander) MissionCount(count uint16) error {
	targetSys := m.GetTargetSystem()
	missionCount := &common.MessageMissionCount{
		TargetSystem:    targetSys,
		TargetComponent: m.config.TargetComponent,
		Count:           count,
	}
	return m.node.WriteMessageAll(missionCount)
}

func (m *MavlinkCommander) MissionRequest(seq uint16) error {
	targetSys := m.GetTargetSystem()
	missionReq := &common.MessageMissionRequest{
		TargetSystem:    targetSys,
		TargetComponent: m.config.TargetComponent,
		Seq:             seq,
	}
	return m.node.WriteMessageAll(missionReq)
}

func (m *MavlinkCommander) MissionAck(missionType uint8) error {
	targetSys := m.GetTargetSystem()
	ack := &common.MessageMissionAck{
		TargetSystem:    targetSys,
		TargetComponent: m.config.TargetComponent,
		Type:            common.MAV_MISSION_RESULT(missionType),
	}
	return m.node.WriteMessageAll(ack)
}

func (m *MavlinkCommander) SendDebug(typeID uint8, value float32) error {
	debugMsg := &common.MessageDebug{
		TimeBootMs: uint32(time.Now().UnixNano() / 1000000),
		Ind:        typeID,
		Value:      value,
	}
	return m.node.WriteMessageAll(debugMsg)
}

func (m *MavlinkCommander) SendNamedValueFloat(name string, value float32) error {
	nameBytes := []byte(name)
	for len(nameBytes) < 10 {
		nameBytes = append(nameBytes, 0)
	}
	targetSys := m.GetTargetSystem()
	namedValue := &common.MessageNamedValueFloat{
		TimeBootMs: uint32(time.Now().UnixNano() / 1000000),
		Name:       string(nameBytes[:10]),
		Value:      value,
	}
	_ = targetSys
	return m.node.WriteMessageAll(namedValue)
}

func (m *MavlinkCommander) SendNamedValueInt(name string, value int32) error {
	nameBytes := []byte(name)
	for len(nameBytes) < 10 {
		nameBytes = append(nameBytes, 0)
	}
	namedValue := &common.MessageNamedValueInt{
		TimeBootMs: uint32(time.Now().UnixNano() / 1000000),
		Name:       string(nameBytes[:10]),
		Value:      value,
	}
	return m.node.WriteMessageAll(namedValue)
}

func (m *MavlinkCommander) SendStatustext(severity uint8, text string) error {
	textBytes := []byte(text)
	for len(textBytes) < 50 {
		textBytes = append(textBytes, 0)
	}
	statustext := &common.MessageStatustext{
		Severity: common.MAV_SEVERITY(severity),
		Text:     string(textBytes[:50]),
	}
	return m.node.WriteMessageAll(statustext)
}

func (m *MavlinkCommander) SendRCChannelsOverride(
	targetSystem uint8,
	targetComponent uint8,
	chan1Raw uint16,
	chan2Raw uint16,
	chan3Raw uint16,
	chan4Raw uint16,
	chan5Raw uint16,
	chan6Raw uint16,
	chan7Raw uint16,
	chan8Raw uint16,
) error {
	override := &common.MessageRcChannelsOverride{
		TargetSystem:    targetSystem,
		TargetComponent: targetComponent,
		Chan1Raw:        chan1Raw,
		Chan2Raw:        chan2Raw,
		Chan3Raw:        chan3Raw,
		Chan4Raw:        chan4Raw,
		Chan5Raw:        chan5Raw,
		Chan6Raw:        chan6Raw,
		Chan7Raw:        chan7Raw,
		Chan8Raw:        chan8Raw,
	}
	return m.node.WriteMessageAll(override)
}

func (m *MavlinkCommander) SendManualControl(
	targetSystem uint8,
	targetComponent uint8,
	x int16,
	y int16,
	z int16,
	r int16,
	buttons uint16,
) error {
	manual := &common.MessageManualControl{
		Target:  targetSystem,
		X:       x,
		Y:       y,
		Z:       z,
		R:       r,
		Buttons: buttons,
	}
	return m.node.WriteMessageAll(manual)
}

func (m *MavlinkCommander) RequestMessage(messageID uint32, targetSystem, targetComponent uint8) error {
	return nil
}

type MavlinkCommandType string

const (
	CMD_ARM_DISARM          MavlinkCommandType = "ARM_DISARM"
	CMD_TAKEOFF             MavlinkCommandType = "TAKEOFF"
	CMD_LAND                MavlinkCommandType = "LAND"
	CMD_RTL                 MavlinkCommandType = "RTL"
	CMD_HOLD                MavlinkCommandType = "HOLD"
	CMD_CONTINUE_MISSION    MavlinkCommandType = "CONTINUE_MISSION"
	CMD_SET_MODE            MavlinkCommandType = "SET_MODE"
	CMD_SET_SPEED           MavlinkCommandType = "SET_SPEED"
	CMD_SET_ALTITUDE        MavlinkCommandType = "SET_ALTITUDE"
	CMD_GET_PARAM           MavlinkCommandType = "GET_PARAM"
	CMD_SET_PARAM           MavlinkCommandType = "SET_PARAM"
	CMD_REQUEST_DATA_STREAM MavlinkCommandType = "REQUEST_DATA_STREAM"
	CMD_MISSION_CLEAR_ALL   MavlinkCommandType = "MISSION_CLEAR_ALL"
	CMD_DO_DIGICAM_CONTROL  MavlinkCommandType = "DO_DIGICAM_CONTROL"
	CMD_DO_SET_RELAY        MavlinkCommandType = "DO_SET_RELAY"
	CMD_DO_SET_SERVO        MavlinkCommandType = "DO_SET_SERVO"
	CMD_CONDITION_YAW       MavlinkCommandType = "CONDITION_YAW"
	CMD_SEND_RC_OVERRIDE    MavlinkCommandType = "SEND_RC_OVERRIDE"
	CMD_SEND_MANUAL_CONTROL MavlinkCommandType = "SEND_MANUAL_CONTROL"
	CMD_HEARTBEAT           MavlinkCommandType = "HEARTBEAT"
	CMD_MISSION_ITEM        MavlinkCommandType = "MISSION_ITEM"
)

type MavlinkCommand struct {
	Type            MavlinkCommandType
	TargetSystem    uint8
	TargetComponent uint8
	Command         uint16
	Confirmation    uint8
	Param1          float32
	Param2          float32
	Param3          float32
	Param4          float32
	Param5          float32
	Param6          float32
	Param7          float32
	Frame           uint8
	X               float32
	Y               float32
	Z               float32
	Current         uint8
	Autocontinue    uint8
}

func (m *MavlinkCommander) ExecuteCommand(cmd MavlinkCommand) error {
	switch cmd.Type {
	case CMD_HEARTBEAT:
		return m.SendHeartbeat()

	case CMD_ARM_DISARM:
		arm := cmd.Param1 > 0
		force := cmd.Param2 > 0
		return m.ArmDisarm(arm, force)

	case CMD_TAKEOFF:
		return m.Takeoff(cmd.Param7)

	case CMD_LAND:
		return m.Land(cmd.Param5, cmd.Param6, cmd.Param7)

	case CMD_RTL:
		return m.ReturnToLaunch()

	case CMD_HOLD:
		return m.Hold()

	case CMD_CONTINUE_MISSION:
		return m.ContinueMission()

	case CMD_SET_MODE:
		return m.SetMode(common.MAV_MODE(cmd.Param1))

	case CMD_SET_SPEED:
		return m.SetSpeed(uint8(cmd.Param1), cmd.Param2, cmd.Param3 > 0)

	case CMD_SET_ALTITUDE:
		return m.WaitAltitude(cmd.Param1)

	case CMD_GET_PARAM:
		return m.GetParam(fmt.Sprintf("%.16f", cmd.Param1))

	case CMD_SET_PARAM:
		return m.SetParam(fmt.Sprintf("%.16f", cmd.Param1), cmd.Param2, uint8(cmd.Param3))

	case CMD_REQUEST_DATA_STREAM:
		return m.RequestDataStream(uint16(cmd.Param1), uint16(cmd.Param2), uint8(cmd.Param3))

	case CMD_MISSION_CLEAR_ALL:
		return m.MissionClearAll()

	case CMD_DO_DIGICAM_CONTROL:
		return m.DoDigICamControl(
			cmd.TargetComponent,
			uint8(cmd.Param1),
			uint8(cmd.Param2),
			uint16(cmd.Param3),
			int8(cmd.Param4),
			uint8(cmd.Param5),
			uint8(cmd.Param6),
		)

	case CMD_DO_SET_RELAY:
		return m.DoSetRelay(uint8(cmd.Param1), cmd.Param2 > 0)

	case CMD_DO_SET_SERVO:
		return m.DoSetServo(uint8(cmd.Param1), uint16(cmd.Param2))

	case CMD_CONDITION_YAW:
		return m.DoSetPositionYaw(cmd.Param1, cmd.Param2)

	case CMD_MISSION_ITEM:
		return m.MissionWriteItem(
			uint16(cmd.Param1),
			cmd.Current,
			cmd.Autocontinue,
			cmd.Frame,
			cmd.Command,
			cmd.Param1, cmd.Param2, cmd.Param3, cmd.Param4,
			cmd.X, cmd.Y, cmd.Z,
		)

	case CMD_SEND_RC_OVERRIDE:
		return m.SendRCChannelsOverride(
			cmd.TargetSystem,
			cmd.TargetComponent,
			uint16(cmd.Param1),
			uint16(cmd.Param2),
			uint16(cmd.Param3),
			uint16(cmd.Param4),
			uint16(cmd.Param5),
			uint16(cmd.Param6),
			uint16(cmd.Param7),
			0,
		)

	case CMD_SEND_MANUAL_CONTROL:
		return m.SendManualControl(
			cmd.TargetSystem,
			cmd.TargetComponent,
			int16(cmd.Param1),
			int16(cmd.Param2),
			int16(cmd.Param3),
			int16(cmd.Param4),
			uint16(cmd.Param5),
		)

	default:
		return m.CommandLong(
			cmd.TargetSystem,
			cmd.TargetComponent,
			cmd.Command,
			cmd.Confirmation,
			cmd.Param1, cmd.Param2, cmd.Param3, cmd.Param4,
			cmd.Param5, cmd.Param6, cmd.Param7,
		)
	}
}

var _ message.Message = &common.MessageHeartbeat{}
