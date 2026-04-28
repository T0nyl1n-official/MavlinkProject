package Core

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	Board "MavlinkProject_Board/Shared/Boards"
	Distribute "MavlinkProject_Board/Distribute"
	MavlinkBoard "MavlinkProject_Board/MavlinkCommand"

	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
	"gopkg.in/yaml.v3"
)

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
	Pixhawk struct {
		SerialPort string `yaml:"serial_port"`
		SerialBaud int    `yaml:"serial_baud"`
		Enabled    bool   `yaml:"enabled"`
	} `yaml:"pixhawk"`
}

var config ServerConfig

func GetConfig() *ServerConfig {
	return &config
}

func LoadConfig(configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		fallbackPaths := []string{
			"config.yaml",
			"/home/admin/MavlinkProject/config/Server_Config.yaml",
			"/home/pi/MavlinkProject/config/Server_Config.yaml",
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
	if config.Pixhawk.Enabled {
		log.Printf("[CentralServer] Pixhawk物理连接已启用: 端口=%s, 波特率=%d", config.Pixhawk.SerialPort, config.Pixhawk.SerialBaud)
	}

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

func (cs *CentralServer) WaitForShutdown() {
	cs.mu.RLock()
	stopChan := cs.stopChan
	cs.mu.RUnlock()
	<-stopChan
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
		log.Printf("[CentralServer] 警告: 当前没有可用的飞控设备，但仍将尝试进行虚拟分配放行")
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
		log.Printf("[CentralServer] 无法找到无人机: %v，指派给本机虚拟执行节点", err)
		chain.AssignedDrone = "virtual_drone_for_test"
		chain.CurrentTask = 0
		return nil
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

		log.Printf("[CentralServer] 开始执行任务 %s: 发送命令 %s", task.TaskID, task.Command)
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
		delaySeconds := 3.0
		if d, ok := task.Data["delay"].(float64); ok {
			delaySeconds = d
		}

		waitTime := time.Duration(delaySeconds * float64(time.Second))
		if time.Since(task.StartTime) >= waitTime {
			log.Printf("[CentralServer] 任务 %s (%s) 缓冲时间结束，标记为完成", task.TaskID, task.Command)
			task.Status = "completed"
			task.EndTime = time.Now()
			return
		}

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
	if droneID == "virtual_drone_for_test" {
		switch task.Command {
		case "TakePhoto":
			return cs.executeTakePhoto(nil, droneID, task.Data)
		case "FourDirectionPhoto":
			return cs.executeFourDirectionPhoto(nil, droneID, task.Data)
		default:
			return fmt.Errorf("虚拟节点仅支持执行部分任务(如TakePhoto/FourDirectionPhoto)")
		}
	}
	commander, err := cs.droneSearch.GetDroneCommander(droneID)
	if err != nil {
		return fmt.Errorf("failed to get drone commander: %v", err)
	}

	switch task.Command {
	case "TakeOff":
		return cs.executeTakeOff(commander, droneID, task.Data)
	case "Land":
		return cs.executeLand(commander, droneID, task.Data)
	case "AutoReturn":
		return cs.executeAutoReturn(commander, droneID, task.Data)
	case "GoTo":
		return cs.executeGoTo(commander, droneID, task.Data)
	case "SetSpeed":
		return cs.executeSetSpeed(commander, droneID, task.Data)
	case "TakePhoto":
		return cs.executeTakePhoto(commander, droneID, task.Data)
	case "SetPosition":
		return cs.executeSetPosition(commander, droneID, task.Data)
	case "Arm":
		return cs.executeArm(commander, droneID, task.Data)
	case "SetMode":
		return cs.executeSetMode(commander, droneID, task.Data)
	case "StartRecord":
		return cs.executeStartRecord(commander, droneID, task.Data)
	case "StopRecord":
		return cs.executeStopRecord(commander, droneID, task.Data)
	case "Orbit":
		return cs.executeOrbit(commander, droneID, task.Data)
	case "FourDirectionPhoto":
		return cs.executeFourDirectionPhoto(commander, droneID, task.Data)
	case "FourDirectionRecord":
		return cs.executeFourDirectionRecord(commander, droneID, task.Data)
	case "SetRPM":
		return cs.executeSetRPM(commander, droneID, task.Data)
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

func (cs *CentralServer) executeAutoReturn(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}
	returnHomeAltitude := float32(0)
	if alt, ok := data["return_altitude"].(float64); ok {
		returnHomeAltitude = float32(alt)
	}
	autoReturnMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         common.MAV_CMD_NAV_RETURN_TO_LAUNCH,
		Param1:          returnHomeAltitude,
	}
	log.Printf("[CentralServer] 正在发送自动返航(AutoReturn)指令到无人机 %s, 返航高度=%.1fm", droneID, returnHomeAltitude)
	return commander.WriteMessage(autoReturnMsg)
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
	log.Printf("[CentralServer] 正在调用USB摄像头拍照...")

	fileName := fmt.Sprintf("/tmp/drone_photo_%d.jpg", time.Now().Unix())
	cmd := exec.Command("fswebcam", "-r", "1280x720", "--no-banner", fileName)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("调用USB摄像头拍照失败: %v", err)
	}

	err := cs.uploadPhotoToBackend(fileName, droneID)
	if err != nil {
		return fmt.Errorf("上传照片失败: %v", err)
	}

	log.Printf("[CentralServer] 拍照任务完成并上传成功: %s", fileName)
	return nil
}

func (cs *CentralServer) executeStartRecord(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}

	cameraID := uint8(0)
	if cid, ok := data["camera_id"].(float64); ok {
		cameraID = uint8(cid)
	}

	recordMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         common.MAV_CMD_DO_DIGICAM_CONTROL,
		Param1:          0,
		Param2:          0,
		Param3:          0,
		Param4:          0,
		Param5:          1,
		Param6:          0,
		Param7:          float32(cameraID),
	}
	log.Printf("[CentralServer] 正在发送开始录像(StartRecord)指令到无人机 %s, 摄像头ID=%d", droneID, cameraID)
	return commander.WriteMessage(recordMsg)
}

func (cs *CentralServer) executeStopRecord(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}

	cameraID := uint8(0)
	if cid, ok := data["camera_id"].(float64); ok {
		cameraID = uint8(cid)
	}

	stopRecordMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         common.MAV_CMD_DO_DIGICAM_CONTROL,
		Param1:          0,
		Param2:          0,
		Param3:          0,
		Param4:          0,
		Param5:          0,
		Param6:          0,
		Param7:          float32(cameraID),
	}
	log.Printf("[CentralServer] 正在发送停止录像(StopRecord)指令到无人机 %s, 摄像头ID=%d", droneID, cameraID)
	return commander.WriteMessage(stopRecordMsg)
}

func (cs *CentralServer) executeOrbit(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}

	radius := 10.0
	if r, ok := data["radius"].(float64); ok {
		radius = r
	}
	speed := 5.0
	if s, ok := data["orbit_speed"].(float64); ok {
		speed = s
	}

	lat, _ := data["latitude"].(float64)
	lon, _ := data["longitude"].(float64)
	alt, _ := data["altitude"].(float64)

	gotoMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         common.MAV_CMD_DO_SET_ROI_LOCATION,
		Param5:          float32(lat),
		Param6:          float32(lon),
		Param7:          float32(alt),
	}

	if err := commander.WriteMessage(gotoMsg); err != nil {
		return err
	}

	speedMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         common.MAV_CMD_DO_CHANGE_SPEED,
		Param2:          float32(speed),
	}
	commander.WriteMessage(speedMsg)

	log.Printf("[CentralServer] 正在执行环绕飞行(Orbit): 目标位置=(%.6f, %.6f), 半径=%.1fm, 速度=%.1fm/s", lat, lon, radius, speed)
	return nil
}

func (cs *CentralServer) executeFourDirectionPhoto(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	log.Printf("[CentralServer] 开始执行四向拍照(FourDirectionPhoto)...")

	directions := []string{"front", "right", "back", "left"}
	yawAngles := []float32{0, 90, 180, 270}

	for i, direction := range directions {
		log.Printf("[CentralServer] 四向拍照 - 方向 %d/%d: %s (偏航角 %.0f°)", i+1, len(directions), direction, yawAngles[i])

		if commander != nil {
			drone, err := cs.droneSearch.GetDroneStatus(droneID)
			if err == nil {
				yawMsg := &common.MessageCommandLong{
					TargetSystem:    drone.SystemID,
					TargetComponent: drone.ComponentID,
					Command:         common.MAV_CMD_CONDITION_YAW,
					Param1:          yawAngles[i],
					Param2:          10,
					Param3:          0,
					Param4:          0,
					Param5:          0,
					Param6:          0,
					Param7:          0,
				}
				commander.WriteMessage(yawMsg)

				time.Sleep(2 * time.Second)
			}
		}

		fileName := fmt.Sprintf("/tmp/drone_4dir_%s_%d.jpg", direction, time.Now().Unix())
		cmd := exec.Command("fswebcam", "-r", "1280x720", "--no-banner", fileName)
		if err := cmd.Run(); err != nil {
			log.Printf("[CentralServer] 四向拍照方向 %s 失败: %v", direction, err)
			continue
		}

		cs.uploadPhotoToBackend(fileName, droneID)
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("[CentralServer] 四向拍照(FourDirectionPhoto)完成")
	return nil
}

func (cs *CentralServer) executeFourDirectionRecord(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	log.Printf("[CentralServer] 开始执行四向录像(FourDirectionRecord)...")

	recordDuration := 3
	if dur, ok := data["duration"].(float64); ok {
		recordDuration = int(dur)
	}

	directions := []string{"front", "right", "back", "left"}
	yawAngles := []float32{0, 90, 180, 270}

	for i, direction := range directions {
		log.Printf("[CentralServer] 四向录像 - 方向 %d/%d: %s (偏航角 %.0f°)", i+1, len(directions), direction, yawAngles[i])

		if commander != nil {
			drone, err := cs.droneSearch.GetDroneStatus(droneID)
			if err == nil {
				yawMsg := &common.MessageCommandLong{
					TargetSystem:    drone.SystemID,
					TargetComponent: drone.ComponentID,
					Command:         common.MAV_CMD_CONDITION_YAW,
					Param1:          yawAngles[i],
					Param2:          10,
				}
				commander.WriteMessage(yawMsg)
				time.Sleep(2 * time.Second)
			}
		}

		startRecordMsg := &common.MessageCommandLong{
			TargetSystem:    1,
			TargetComponent: 1,
			Command:         common.MAV_CMD_DO_DIGICAM_CONTROL,
			Param5:          1,
		}
		if commander != nil {
			commander.WriteMessage(startRecordMsg)
		}

		time.Sleep(time.Duration(recordDuration) * time.Second)

		stopRecordMsg := &common.MessageCommandLong{
			TargetSystem:    1,
			TargetComponent: 1,
			Command:         common.MAV_CMD_DO_DIGICAM_CONTROL,
			Param5:          0,
		}
		if commander != nil {
			commander.WriteMessage(stopRecordMsg)
		}

		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("[CentralServer] 四向录像(FourDirectionRecord)完成")
	return nil
}

func (cs *CentralServer) executeSetRPM(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}

	rpmValue := float32(0)
	if rpm, ok := data["rpm"].(float64); ok {
		rpmValue = float32(rpm)
	}

	motorIndex := float32(-1)
	if idx, ok := data["motor_index"].(float64); ok {
		motorIndex = float32(idx)
	}

	rpmMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         0,
		Confirmation:    0,
		Param1:          motorIndex,
		Param2:          rpmValue,
		Param3:          0,
		Param4:          0,
		Param5:          0,
		Param6:          0,
		Param7:          0,
	}
	log.Printf("[CentralServer] 正在设置转速(SetRPM): 马达索引=%.0f, RPM=%.0f", motorIndex, rpmValue)
	return commander.WriteMessage(rpmMsg)
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

func (cs *CentralServer) executeArm(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}

	armMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         common.MAV_CMD_COMPONENT_ARM_DISARM,
		Param1:          1,
		Param2:          21196,
	}
	log.Printf("[CentralServer] 正在发送强制解锁(Arm)指令...")
	return commander.WriteMessage(armMsg)
}

func (cs *CentralServer) executeSetMode(commander *MavlinkBoard.MavlinkCommander, droneID string, data map[string]interface{}) error {
	drone, err := cs.droneSearch.GetDroneStatus(droneID)
	if err != nil {
		return err
	}

	modeMsg := &common.MessageCommandLong{
		TargetSystem:    drone.SystemID,
		TargetComponent: drone.ComponentID,
		Command:         common.MAV_CMD_DO_SET_MODE,
		Param1:          1,
		Param2:          4,
	}
	log.Printf("[CentralServer] 正在发送切换 GUIDED(引导) 模式指令...")
	return commander.WriteMessage(modeMsg)
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

func StartLocalPixhawk(cs *CentralServer, serialPort string, baudRate int) {
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
			log.Printf("[LocalPixhawk] ==========================================")
			log.Printf("[LocalPixhawk] 错误: 物理飞控连接失败!")
			log.Printf("[LocalPixhawk] 端口: %s, 波特率: %d", serialPort, baudRate)
			log.Printf("[LocalPixhawk] 原因: %v", err)
			log.Printf("[LocalPixhawk] ==========================================")
			return
		}

		boardID := "pixhawk_local_0"

		cs.droneSearch.RegisterDroneCommander(boardID, commander)
		log.Printf("[LocalPixhawk] 成功连接并接管本地飞控物理流: %s", serialPort)

		go func() {
			heartbeatTicker := time.NewTicker(1 * time.Second)
			defer heartbeatTicker.Stop()

			hbMsg := &common.MessageHeartbeat{
				Type:           common.MAV_TYPE_GCS,
				Autopilot:      common.MAV_AUTOPILOT_INVALID,
				BaseMode:       common.MAV_MODE_FLAG(common.MAV_MODE_FLAG_CUSTOM_MODE_ENABLED),
				SystemStatus:   common.MAV_STATE_ACTIVE,
				MavlinkVersion: 3,
			}

			for range heartbeatTicker.C {
				err := commander.WriteMessage(hbMsg)
				if err != nil {
				}
			}
		}()

		droneData := make(map[string]interface{})
		var lastUpdate time.Time

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
					droneData["status"] = "idle"
				}
				updated = true

			case *common.MessageSysStatus:
				if m.BatteryRemaining <= 0 {
					droneData["battery"] = 100.0
				} else {
					droneData["battery"] = float64(m.BatteryRemaining)
				}
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

					cs.droneSearch.InjectDroneStatus(msg)
					lastUpdate = time.Now()
				}
			}
		}

		log.Printf("[LocalPixhawk] 警告: 飞控硬件 %s 数据流已强制断开或退出连接", serialPort)
	}()
}

func (cs *CentralServer) uploadPhotoToBackend(filePath string, droneID string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	defer os.Remove(filePath)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("photo", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}
	_ = writer.WriteField("drone_id", droneID)
	writer.Close()

	uploadURL := fmt.Sprintf("https://%s:%s/api/upload/photo", "api.deeppluse.dpdns.org", "8080")

	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: tr,
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("后端返回错误状态码: %d", resp.StatusCode)
	}
	return nil
}
