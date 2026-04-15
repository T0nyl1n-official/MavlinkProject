package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"

	Board "MavlinkProject/Server/backend/Shared/Boards"
	Distribute "MavlinkProject/Server/backend/Utils/CentralBoard/Distribute"
	MavlinkBoard "MavlinkProject/Server/backend/Utils/CentralBoard/MavlinkCommand"

	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"gopkg.in/yaml.v3"
)

// 配置文件结构
type ServerConfig struct {
	Central struct {
		Address string `yaml:"address"`
		Port    string `yaml:"port"`
		Task    struct {
			MaxRetries int `yaml:"max_retries"`
			Timeout    int `yaml:"timeout"`
		} `yaml:"task"`
		Drone struct {
			MinBatteryLevel    float64 `yaml:"min_battery_level"`
			MaxDroneDistance   float64 `yaml:"max_drone_distance"`
			StatusCheckTimeout int     `yaml:"status_check_timeout"`
		} `yaml:"drone"`
	} `yaml:"central"`
	Backend struct {
		Address string `yaml:"address"`
		Port    string `yaml:"port"`
	} `yaml:"backend"`
}

var config ServerConfig

// GetConfig 获取全局配置
func GetConfig() *ServerConfig {
	return &config
}

func loadConfig(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fallbackPaths := []string{
			"/home/admin/MavlinkProject/config/Server_Config.yaml",
			"/home/pi/MavlinkProject/config/Server_Config.yaml",
			"/home/Guocj/MavlinkProject/config/Server_Config.yaml",
			"/root/MavlinkProject/config/Server_Config.yaml",
		}
		for _, path := range fallbackPaths {
			if _, err := os.Stat(path); err == nil {
				configPath = path
				break
			}
		}
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return fmt.Errorf("解析配置文件失败: %v", err)
	}

	log.Printf("[CentralServer] 配置文件加载成功: %s", configPath)
	log.Printf("[CentralServer] Central监听地址: %s:%s", config.Central.Address, config.Central.Port)
	log.Printf("[CentralServer] 任务配置: 最大重试=%d, 超时=%d秒", config.Central.Task.MaxRetries, config.Central.Task.Timeout)
	log.Printf("[CentralServer] 无人机配置: 最小电量=%.1f%%, 最大距离=%.1f米", config.Central.Drone.MinBatteryLevel, config.Central.Drone.MaxDroneDistance)

	return nil
}

const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)

type ProgressChain struct {
	ChainID       string    `json:"chain_id"`
	Tasks         []Task    `json:"tasks"`
	CurrentTask   int       `json:"current_task"`
	Status        string    `json:"status"`
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	AssignedDrone string    `json:"assigned_drone"`
}

type Task struct {
	TaskID     string                 `json:"task_id"`
	Command    string                 `json:"command"`
	Data       map[string]interface{} `json:"data"`
	Status     string                 `json:"status"`
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	Timeout    time.Duration          `json:"timeout"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
}

type CentralServer struct {
	droneSearch  *Distribute.DroneSearch
	taskChains   map[string]*ProgressChain
	activeChains map[string]*ProgressChain
	mu           sync.RWMutex
	listener     net.Listener
	address      string
	port         string
	running      bool
	stopChan     chan bool
}

func (cs *CentralServer) DroneSearch() *Distribute.DroneSearch {
	return cs.droneSearch
}

const (
	DefaultPort = "8081"
	MaxRetries  = 3
	TaskTimeout = 30 * time.Second
)

func NewCentralServer(address, port string) *CentralServer {
	if port == "" {
		port = DefaultPort
	}
	return &CentralServer{
		droneSearch:  Distribute.GetDroneSearch(),
		taskChains:   make(map[string]*ProgressChain),
		activeChains: make(map[string]*ProgressChain),
		address:      address,
		port:         port,
		stopChan:     make(chan bool),
	}
}

func (cs *CentralServer) Start() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if cs.running {
		return fmt.Errorf("CentralServer already running")
	}

	if err := cs.droneSearch.Start(); err != nil {
		return fmt.Errorf("failed to start DroneSearch: %v", err)
	}

	listener, err := net.Listen("tcp", cs.address+":"+cs.port)
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	cs.listener = listener
	cs.running = true

	go cs.acceptConnections()
	go cs.taskProcessor()

	log.Printf("[CentralServer] Started on port %s", cs.port)
	return nil
}

func (cs *CentralServer) Stop() error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if !cs.running {
		return nil
	}

	cs.running = false
	close(cs.stopChan)
	if cs.listener != nil {
		cs.listener.Close()
	}
	cs.droneSearch.Stop()
	log.Printf("[CentralServer] Stopped")
	return nil
}

func (cs *CentralServer) acceptConnections() {
	for {
		select {
		case <-cs.stopChan:
			return
		default:
			conn, err := cs.listener.Accept()
			if err != nil {
				if cs.running {
					log.Printf("[CentralServer] Accept error: %v", err)
				}
				continue
			}
			go cs.handleConnection(conn)
		}
	}
}

func (cs *CentralServer) handleConnection(conn net.Conn) {
	defer conn.Close()
	log.Printf("[CentralServer] New connection from %s", conn.RemoteAddr())
	buffer := make([]byte, 4096)
	for {
		n, err := conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Printf("[CentralServer] 客户端 %s 正常断开连接", conn.RemoteAddr())
				return
			}
			log.Printf("[CentralServer] Read error from %s: %v", conn.RemoteAddr(), err)
			return
		}

		if n > 0 {
			var boardMsg Board.BoardMessage
			if err := json.Unmarshal(buffer[:n], &boardMsg); err != nil {
				resp := map[string]interface{}{"status": "error", "message": err.Error()}
				respData, _ := json.Marshal(resp)
				conn.Write(respData)
				continue
			}

			log.Printf("[CentralServer] Received message: Command=%s, FromID=%s", boardMsg.Message.Command, boardMsg.FromID)

			if err := cs.handleBoardMessage(&boardMsg); err != nil {
				resp := map[string]interface{}{"status": "error", "message": err.Error()}
				respData, _ := json.Marshal(resp)
				conn.Write(respData)
				continue
			}

			response := map[string]interface{}{
				"status":  "received",
				"message": "Task chain received and queued",
			}
			respData, _ := json.Marshal(response)
			conn.Write(respData)
		}
	}
}

func (cs *CentralServer) handleBoardMessage(msg *Board.BoardMessage) error {
	if msg.Message.Data == nil {
		return fmt.Errorf("message data is nil")
	}

	chainData, exists := msg.Message.Data["progress_chain"]
	if !exists {
		return fmt.Errorf("no progress_chain found in message data")
	}

	availableDrones := cs.droneSearch.GetAvailableDrones()
	if len(availableDrones) == 0 {
		return fmt.Errorf("调度失败: 当前没有成功连接的可用飞控设备或所有设备电量不足")
	}

	chainJSON, err := json.Marshal(chainData)
	if err != nil {
		return fmt.Errorf("failed to marshal chain data: %v", err)
	}

	var progressChain ProgressChain
	if err := json.Unmarshal(chainJSON, &progressChain); err != nil {
		return fmt.Errorf("failed to unmarshal progress chain: %v", err)
	}

	if progressChain.ChainID == "" {
		progressChain.ChainID = fmt.Sprintf("chain_%d", time.Now().UnixNano())
	}
	progressChain.Status = "pending"
	progressChain.StartTime = time.Now()

	for i := range progressChain.Tasks {
		progressChain.Tasks[i].TaskID = fmt.Sprintf("task_%d", i)
		progressChain.Tasks[i].Status = "pending"
		progressChain.Tasks[i].MaxRetries = MaxRetries
		progressChain.Tasks[i].Timeout = TaskTimeout
	}

	cs.mu.Lock()
	cs.taskChains[progressChain.ChainID] = &progressChain
	cs.mu.Unlock()

	log.Printf("[CentralServer] Received progress chain: %s with %d tasks", progressChain.ChainID, len(progressChain.Tasks))
	return nil
}

func (cs *CentralServer) HandleBoardMessage(msg *Board.BoardMessage) error {
	return cs.handleBoardMessage(msg)
}

func (cs *CentralServer) taskProcessor() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-cs.stopChan:
			return
		case <-ticker.C:
			cs.processTaskChains()
		}
	}
}

func (cs *CentralServer) processTaskChains() {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for chainID, chain := range cs.taskChains {
		if chain.Status == "pending" {
			if err := cs.startChainExecution(chain); err != nil {
				log.Printf("[CentralServer] Failed to start chain %s: %v", chainID, err)
				chain.Status = "failed"
				chain.EndTime = time.Now()
			} else {
				chain.Status = "running"
				cs.activeChains[chainID] = chain
				delete(cs.taskChains, chainID)
			}
		}
	}

	for chainID, chain := range cs.activeChains {
		if chain.Status == "running" {
			cs.executeCurrentTask(chain)
		}

		if chain.Status == "completed" || chain.Status == "failed" {
			chain.EndTime = time.Now()
			delete(cs.activeChains, chainID)
			log.Printf("[CentralServer] Chain %s finished with status: %s", chainID, chain.Status)
		}
	}
}

func (cs *CentralServer) startChainExecution(chain *ProgressChain) error {
	bestDrone, err := cs.droneSearch.FindBestDrone()
	if err != nil {
		return fmt.Errorf("no available drone found: %v", err)
	}

	chain.AssignedDrone = bestDrone.BoardID
	chain.CurrentTask = 0
	cs.droneSearch.SetDroneIdle(bestDrone.BoardID, false)
	log.Printf("[CentralServer] Chain %s assigned to drone %s", chain.ChainID, bestDrone.BoardID)
	return nil
}

func (cs *CentralServer) executeCurrentTask(chain *ProgressChain) {
	if chain.CurrentTask >= len(chain.Tasks) {
		chain.Status = "completed"
		cs.droneSearch.SetDroneIdle(chain.AssignedDrone, true)
		return
	}

	task := &chain.Tasks[chain.CurrentTask]

	switch task.Status {
	case "pending":
		task.StartTime = time.Now()
		task.Status = "running"

		if err := cs.executeTask(chain.AssignedDrone, task); err != nil {
			log.Printf("[CentralServer] Task %s execution failed: %v", task.TaskID, err)
			task.Status = "failed"
			task.EndTime = time.Now()

			if task.RetryCount < task.MaxRetries {
				task.RetryCount++
				task.Status = "pending"
				log.Printf("[CentralServer] Retrying task %s (attempt %d)", task.TaskID, task.RetryCount)
			} else {
				chain.Status = "failed"
				cs.droneSearch.SetDroneIdle(chain.AssignedDrone, true)
			}
		}

	case "running":
		if time.Since(task.StartTime) > task.Timeout {
			log.Printf("[CentralServer] Task %s timeout", task.TaskID)
			task.Status = "failed"
			task.EndTime = time.Now()

			if task.RetryCount < task.MaxRetries {
				task.RetryCount++
				task.Status = "pending"
			} else {
				chain.Status = "failed"
				cs.droneSearch.SetDroneIdle(chain.AssignedDrone, true)
			}
		}

	case "completed":
		chain.CurrentTask++
		if chain.CurrentTask >= len(chain.Tasks) {
			chain.Status = "completed"
			cs.droneSearch.SetDroneIdle(chain.AssignedDrone, true)
		}

	case "failed":
		if task.RetryCount < task.MaxRetries {
			task.RetryCount++
			task.Status = "pending"
		} else {
			chain.Status = "failed"
			cs.droneSearch.SetDroneIdle(chain.AssignedDrone, true)
		}
	}
}

func (cs *CentralServer) executeTask(droneID string, task *Task) error {
	commander, err := cs.droneSearch.GetDroneCommander(droneID)
	if err != nil {
		return fmt.Errorf("failed to get drone commander: %v", err)
	}

	switch task.Command {
	case "TakeOff":
		return cs.executeTakeOff(commander, droneID, task.Data)
	case "Land":
		return cs.executeLand(commander, droneID, task.Data)
	case "GoTo":
		return cs.executeGoTo(commander, droneID, task.Data)
	case "SetSpeed":
		return cs.executeSetSpeed(commander, droneID, task.Data)
	case "TakePhoto":
		return cs.executeTakePhoto(commander, droneID, task.Data)
	case "SetPosition":
		return cs.executeSetPosition(commander, droneID, task.Data)
	default:
		return fmt.Errorf("unknown command: %s", task.Command)
	}
}

func (cs *CentralServer) executeTakeOff(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}
	altitude := 10.0
	if alt, ok := data["altitude"].(float64); ok {
		altitude = alt
	}
	takeoffMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         common.MAV_CMD_NAV_TAKEOFF,
		Param7:          float32(altitude),
	}
	return commander.WriteMessage(takeoffMsg)
}

func (cs *CentralServer) executeLand(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}
	landMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         common.MAV_CMD_NAV_LAND,
	}
	return commander.WriteMessage(landMsg)
}

func (cs *CentralServer) executeGoTo(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}
	lat, _ := data["latitude"].(float64)
	lon, _ := data["longitude"].(float64)
	alt, _ := data["altitude"].(float64)

	gotoMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         common.MAV_CMD_NAV_WAYPOINT,
		Param5:          float32(lat),
		Param6:          float32(lon),
		Param7:          float32(alt),
	}
	return commander.WriteMessage(gotoMsg)
}

func (cs *CentralServer) executeSetSpeed(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}
	speed, _ := data["speed"].(float64)
	speedMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         common.MAV_CMD_DO_CHANGE_SPEED,
		Param2:          float32(speed),
	}
	return commander.WriteMessage(speedMsg)
}

func (cs *CentralServer) executeTakePhoto(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}
	photoMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         common.MAV_CMD_IMAGE_START_CAPTURE,
	}
	return commander.WriteMessage(photoMsg)
}

func (cs *CentralServer) executeSetPosition(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}
	lat, _ := data["latitude"].(float64)
	lon, _ := data["longitude"].(float64)
	alt, _ := data["altitude"].(float64)

	positionMsg := &common.MessageSetPositionTargetGlobalInt{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		LatInt:          int32(lat * 1e7),
		LonInt:          int32(lon * 1e7),
		Alt:             float32(alt),
		CoordinateFrame: common.MAV_FRAME_GLOBAL_RELATIVE_ALT,
	}
	return commander.WriteMessage(positionMsg)
}

func (cs *CentralServer) GetChainStatus(chainID string) (*ProgressChain, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if chain, exists := cs.taskChains[chainID]; exists {
		return chain, nil
	}
	if chain, exists := cs.activeChains[chainID]; exists {
		return chain, nil
	}
	return nil, fmt.Errorf("chain %s not found", chainID)
}

func (cs *CentralServer) GetAllChains() []*ProgressChain {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	chains := make([]*ProgressChain, 0)
	for _, chain := range cs.taskChains {
		chains = append(chains, chain)
	}
	for _, chain := range cs.activeChains {
		chains = append(chains, chain)
	}
	return chains
}

// StartLocalPixhawk 启动本地物理串口无人机连接
func StartLocalPixhawk(cs *CentralServer, serialPort string, baudRate int) {
	// 用独立的协程启动，坚决防止底层串口库死锁卡住中央服务器的 Main 循环
	go func() {
		log.Printf("[LocalPixhawk] 准备连接本地物理飞控... 端口: %s 波特率: %d", serialPort, baudRate)

		commander := MavlinkBoard.NewMavlinkCommander()
		commander.Configure(MavlinkBoard.MavlinkConfig{
			ConnectionType:  MavlinkBoard.ConnectionSerial,
			SerialPort:      serialPort,
			SerialBaud:      baudRate,
			SystemID:        255,
			ComponentID:     190,
			TargetSystem:    1,
			TargetComponent: 1,
		})

		startErrChan := make(chan error, 1)
		go func() {
			startErrChan <- commander.Start()
		}()

		var err error
		select {
		case err = <-startErrChan:
		case <-time.After(10 * time.Second):
			err = fmt.Errorf("打开串口建立连接超时 (设备可能已死锁)")
		}

		if err != nil {
			log.Printf("[LocalPixhawk] =========================================")
			log.Printf("[LocalPixhawk] 错误: 物理飞控连接失败!")
			log.Printf("[LocalPixhawk] 端口: %s, 波特率: %d", serialPort, baudRate)
			log.Printf("[LocalPixhawk] 原因: %v", err)
			log.Printf("[LocalPixhawk] =========================================")
			return
		}

		boardID := "pixhawk_local_0"

		cs.droneSearch.RegisterDroneCommander(boardID, commander)
		log.Printf("[LocalPixhawk] 成功连接并接管本地飞控物理流: %s", serialPort)

		droneData := make(map[string]interface{})
		var lastUpdate time.Time

		// 处理从物理接口获得的数据流
		for recvMsg := range commander.GetMessageChan() {
			updated := false
			droneData["system_id"] = float64(recvMsg.SystemID)
			droneData["component_id"] = float64(recvMsg.ComponentID)

			switch m := recvMsg.Message.(type) {
			case *common.MessageHeartbeat:
				isArmed := (int(m.BaseMode) & int(common.MAV_MODE_FLAG_SAFETY_ARMED)) != 0
				if isArmed {
					droneData["status"] = "running"
				} else {
					droneData["status"] = "idle" // 真实收到心跳包方可标记为 available (idle)
				}
				updated = true

			case *common.MessageSysStatus:
				droneData["battery"] = float64(m.BatteryRemaining)
				updated = true

			case *common.MessageGlobalPositionInt:
				droneData["latitude"] = float64(m.Lat) / 1e7
				droneData["longitude"] = float64(m.Lon) / 1e7
				droneData["altitude"] = float64(m.RelativeAlt) / 1000.0
				updated = true
			}

			if updated && time.Since(lastUpdate) > time.Second {
				dataCopy := make(map[string]interface{})
				for k, v := range droneData {
					dataCopy[k] = v
				}

				if len(dataCopy) > 0 {
					msg := &Board.BoardMessage{
						FromID: boardID,
						Message: Board.Message{
							Data: dataCopy,
						},
					}

					// 将数据注入全局 DroneSearch 集群
					cs.droneSearch.InjectDroneStatus(msg)
					lastUpdate = time.Now()
				}
			}
		}

		// 如果执行到这里说明物理通道关闭了
		log.Printf("[LocalPixhawk] 警告: 飞控硬件 %s 数据流已强制断开或退出连接", serialPort)
	}()
}

func main() {
	configPath := "config/Server_Config.yaml"
	if err := loadConfig(configPath); err != nil {
		log.Printf("警告: 加载配置文件失败，请确保 yaml 存在 (%v)", err)
	}

	port := config.Central.Port
	if port == "" {
		port = "8081"
	}
	address := config.Central.Address

	central := NewCentralServer(address, port)

	timeout := time.Duration(config.Central.Drone.StatusCheckTimeout) * time.Second
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	central.droneSearch.UpdateConfig(Distribute.DroneConfig{
		MinBatteryLevel:    config.Central.Drone.MinBatteryLevel,
		MaxDroneDistance:   config.Central.Drone.MaxDroneDistance,
		StatusCheckTimeout: timeout,
	})

	if err := central.Start(); err != nil {
		log.Fatalf("Failed to start CentralServer: %v", err)
	}

	StartLocalPixhawk(central, "/dev/ttyACM0", 115200)

	log.Printf("Central调度系统已启动, 监听地址 %s:%s", address, port)
	log.Printf("等待接收ProgressChain任务链...")

	// 阻塞守候以提供持续服务
	<-central.stopChan

	log.Printf("Central调度系统正在关闭...")
	central.Stop()
	log.Printf("Central调度系统已关闭")
}
