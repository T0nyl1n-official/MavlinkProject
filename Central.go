package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	Board "MavlinkProject_Board/Shared/Boards"
	Distribute "MavlinkProject_Board/Distribute"
	MavlinkBoard "MavlinkProject_Board/MavlinkCommand"

	"github.com/bluenviron/gomavlib/v3/pkg/dialects/common"
)

// 后端配置
const (
	backendAddress = "localhost"
	backendPort    = "8080"
)

// using for Task.Status
const (
	TaskStatusPending   = "pending"
	TaskStatusRunning   = "running"
	TaskStatusCompleted = "completed"
	TaskStatusFailed    = "failed"
)

// ProgressChain 链式任务结构
type ProgressChain struct {
	ChainID       string    `json:"chain_id"`
	Tasks         []Task    `json:"tasks"`
	CurrentTask   int       `json:"current_task"`
	Status        string    `json:"status"` // pending, running, completed, failed
	StartTime     time.Time `json:"start_time"`
	EndTime       time.Time `json:"end_time"`
	AssignedDrone string    `json:"assigned_drone"`
}

// Task 单个任务结构
type Task struct {
	TaskID     string                 `json:"task_id"`
	Command    string                 `json:"command"`
	Data       map[string]interface{} `json:"data"`
	Status     string                 `json:"status"` // pending, running, completed, failed
	RetryCount int                    `json:"retry_count"`
	MaxRetries int                    `json:"max_retries"`
	Timeout    time.Duration          `json:"timeout"`
	StartTime  time.Time              `json:"start_time"`
	EndTime    time.Time              `json:"end_time"`
}

// CentralServer 中央调度服务器
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

const (
	DefaultPort = "8080"
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

	// 启动 DroneSearch
	if err := cs.droneSearch.Start(); err != nil {
		return fmt.Errorf("failed to start DroneSearch: %v", err)
	}

	// 启动网络监听
	listener, err := net.Listen("tcp", cs.address+":"+cs.port)
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	cs.listener = listener

	cs.running = true

	// 启动处理循环
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

	// 停止 DroneSearch
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
			log.Printf("[CentralServer] Read error from %s: %v", conn.RemoteAddr(), err)
			return
		}

		if n > 0 {
			var boardMsg Board.BoardMessage
			if err := json.Unmarshal(buffer[:n], &boardMsg); err != nil {
				log.Printf("[CentralServer] JSON unmarshal error: %v", err)
				continue
			}

			// 处理接收到的消息
			if err := cs.handleBoardMessage(&boardMsg); err != nil {
				log.Printf("[CentralServer] Handle message error: %v", err)
			}

			// 发送响应
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

	// 检查是否为任务链消息
	chainData, exists := msg.Message.Data["progress_chain"]
	if !exists {
		return fmt.Errorf("no progress_chain found in message data")
	}

	// 解析任务链
	chainJSON, err := json.Marshal(chainData)
	if err != nil {
		return fmt.Errorf("failed to marshal chain data: %v", err)
	}

	var progressChain ProgressChain
	if err := json.Unmarshal(chainJSON, &progressChain); err != nil {
		return fmt.Errorf("failed to unmarshal progress chain: %v", err)
	}

	// 设置任务链基本信息
	if progressChain.ChainID == "" {
		progressChain.ChainID = fmt.Sprintf("chain_%d", time.Now().UnixNano())
	}
	progressChain.Status = "pending"
	progressChain.StartTime = time.Now()

	// 初始化任务状态
	for i := range progressChain.Tasks {
		progressChain.Tasks[i].TaskID = fmt.Sprintf("task_%d", i)
		progressChain.Tasks[i].Status = "pending"
		progressChain.Tasks[i].MaxRetries = MaxRetries
		progressChain.Tasks[i].Timeout = TaskTimeout
	}

	// 保存任务链
	cs.mu.Lock()
	cs.taskChains[progressChain.ChainID] = &progressChain
	cs.mu.Unlock()

	log.Printf("[CentralServer] Received progress chain: %s with %d tasks",
		progressChain.ChainID, len(progressChain.Tasks))

	return nil
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

	// 处理待处理的任务链
	for chainID, chain := range cs.taskChains {
		if chain.Status == "pending" {
			// 分配无人机并开始执行
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

	// 处理执行中的任务链
	for chainID, chain := range cs.activeChains {
		if chain.Status == "running" {
			cs.executeCurrentTask(chain)
		}

		// 检查任务链是否完成
		if chain.Status == "completed" || chain.Status == "failed" {
			chain.EndTime = time.Now()
			delete(cs.activeChains, chainID)
			log.Printf("[CentralServer] Chain %s finished with status: %s", chainID, chain.Status)
		}
	}
}

func (cs *CentralServer) startChainExecution(chain *ProgressChain) error {
	// 查找最佳无人机
	bestDrone, err := cs.droneSearch.FindBestDrone()
	if err != nil {
		return fmt.Errorf("no available drone found: %v", err)
	}

	chain.AssignedDrone = bestDrone.BoardID
	chain.CurrentTask = 0

	// 标记无人机为忙碌状态
	cs.droneSearch.SetDroneIdle(bestDrone.BoardID, false)

	log.Printf("[CentralServer] Chain %s assigned to drone %s",
		chain.ChainID, bestDrone.BoardID)

	return nil
}

func (cs *CentralServer) executeCurrentTask(chain *ProgressChain) {
	if chain.CurrentTask >= len(chain.Tasks) {
		chain.Status = "completed"
		// 释放无人机
		cs.droneSearch.SetDroneIdle(chain.AssignedDrone, true)
		return
	}

	task := &chain.Tasks[chain.CurrentTask]

	// 检查任务状态
	switch task.Status {
	case "pending":
		// 开始执行任务
		task.StartTime = time.Now()
		task.Status = "running"

		if err := cs.executeTask(chain.AssignedDrone, task); err != nil {
			log.Printf("[CentralServer] Task %s execution failed: %v", task.TaskID, err)
			task.Status = "failed"
			task.EndTime = time.Now()

			// 重试逻辑
			if task.RetryCount < task.MaxRetries {
				task.RetryCount++
				task.Status = "pending"
				log.Printf("[CentralServer] Retrying task %s (attempt %d)",
					task.TaskID, task.RetryCount)
			} else {
				chain.Status = "failed"
				cs.droneSearch.SetDroneIdle(chain.AssignedDrone, true)
			}
		}

	case "running":
		// 检查任务超时
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
		// 任务完成，移动到下一个任务
		chain.CurrentTask++
		if chain.CurrentTask >= len(chain.Tasks) {
			chain.Status = "completed"
			cs.droneSearch.SetDroneIdle(chain.AssignedDrone, true)
		}

	case "failed":
		// 任务失败，处理重试或失败链
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

	// 根据命令类型执行相应操作
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

// 具体的命令执行函数
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

// 获取任务链状态
func (cs *CentralServer) GetChainStatus(chainID string) (*ProgressChain, error) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	chain, exists := cs.taskChains[chainID]
	if exists {
		return chain, nil
	}

	chain, exists = cs.activeChains[chainID]
	if exists {
		return chain, nil
	}

	return nil, fmt.Errorf("chain %s not found", chainID)
}

// 获取所有任务链状态
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

func main() {
	// 创建中央调度服务器
	central := NewCentralServer(backendAddress, backendPort)

	// 启动服务器
	if err := central.Start(); err != nil {
		log.Fatalf("Failed to start CentralServer: %v", err)
	}

	log.Printf("Central调度系统已启动, 监听端口 %s", backendPort)
	log.Printf("等待接收ProgressChain任务链...")

	// 等待中断信号
	<-central.stopChan

	log.Printf("Central调度系统正在关闭...")
	central.Stop()
	log.Printf("Central调度系统已关闭")
}
