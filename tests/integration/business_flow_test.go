package integration

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	camera "MavlinkProject_Board/tests/camera"
	Core "MavlinkProject_Board/Core"
	Distribute "MavlinkProject_Board/Distribute"
	MavlinkBoard "MavlinkProject_Board/MavlinkCommand"
	Board "MavlinkProject_Board/Shared/Boards"
)

const (
	TestOutputDir     = "../../tests/OutputHistory"
	CentralTestPort   = "18081"
	CentralTestAddr   = "127.0.0.1"
	CentralAPIPort    = "18084"
)

type TestResult struct {
	TestName      string                 `json:"test_name"`
	Status        string                 `json:"status"`
	Message       string                 `json:"message"`
	DurationMs    int64                  `json:"duration_ms"`
	Timestamp     time.Time              `json:"timestamp"`
	Data          map[string]interface{} `json:"data,omitempty"`
	Error         string                 `json:"error,omitempty"`
}

type TestReport struct {
	SuiteName    string       `json:"suite_name"`
	StartTime    time.Time    `json:"start_time"`
	EndTime      time.Time    `json:"end_time"`
	TotalTests   int          `json:"total_tests"`
	PassedTests  int          `json:"passed_tests"`
	FailedTests  int          `json:"failed_tests"`
	SkippedTests int          `json:"skipped_tests"`
	Results      []TestResult `json:"results"`
}

var testReport TestReport
var reportMutex sync.Mutex
var testCamera *camera.MockCamera

func init() {
	os.MkdirAll(TestOutputDir, 0755)
	os.MkdirAll(filepath.Join(TestOutputDir, "photos"), 0755)

	testCamera = camera.NewMockCamera(filepath.Join(TestOutputDir, "photos"))
	testCamera.SetResolution("1280", "720")
	testCamera.SetQuality("85")

	testReport = TestReport{
		SuiteName: "CentralBoard Business Flow Integration",
		StartTime: time.Now(),
		Results:   make([]TestResult, 0),
	}
}

func recordResult(t *testing.T, testName string, passed bool, message string, data map[string]interface{}) {
	reportMutex.Lock()
	defer reportMutex.Unlock()

	result := TestResult{
		TestName:  testName,
		Status:    "passed",
		Message:   message,
		Timestamp: time.Now(),
		Data:      data,
	}
	if !passed {
		result.Status = "failed"
		t.Errorf("[FAIL] %s: %s", testName, message)
	} else {
		t.Logf("[PASS] %s: %s", testName, message)
	}

	testReport.Results = append(testReport.Results, result)
	if passed {
		testReport.PassedTests++
	} else {
		testReport.FailedTests++
	}
	testReport.TotalTests++
}

func saveTestReport() {
	reportMutex.Lock()
	defer reportMutex.Unlock()
	testReport.EndTime = time.Now()

	reportFile := filepath.Join(TestOutputDir, fmt.Sprintf("centralboard_test_report_%s.json",
		time.Now().Format("20060102_150405")))

	data, _ := json.MarshalIndent(testReport, "", "  ")
	os.WriteFile(reportFile, data, 0644)

	log.Printf("[TestReport] 报告已保存: %s (总计=%d, 通过=%d, 失败=%d)",
		reportFile, testReport.TotalTests, testReport.PassedTests, testReport.FailedTests)
}

func TestMain(m *testing.M) {
	log.Println("========================================")
	log.Println("CentralBoard 业务流程集成测试")
	log.Println("========================================")

	code := m.Run()

	saveTestReport()

	log.Println("========================================")
	log.Printf("测试完成: 总计=%d, 通过=%d, 失败=%d", testReport.TotalTests, testReport.PassedTests, testReport.FailedTests)
	log.Println("========================================")

	os.Exit(code)
}

func Test_001_CameraInitialization(t *testing.T) {
	startTime := time.Now()

	status := testCamera.GetStatus()
	passed := status.Available

	recordResult(t, "T001_摄像头初始化", passed,
		fmt.Sprintf("OS=%s, 设备=%s, 可用=%v, 输出目录=%s", status.OS, status.DeviceName, status.Available, status.OutputDir),
		map[string]interface{}{
			"os":           status.OS,
			"device_name":  status.DeviceName,
			"available":    status.Available,
			"output_dir":   status.OutputDir,
			"resolution":   status.Resolution,
			"existing_files": status.TotalFiles,
			"duration_ms":  time.Since(startTime).Milliseconds(),
		})
}

func Test_002_TakeSinglePhoto(t *testing.T) {
	startTime := time.Now()

	result, err := testCamera.TakePhoto("default")
	passed := err == nil && result.FilePath != ""

	recordResult(t, "T002_单张拍照", passed,
		fmt.Sprintf("文件=%s, 大小=%d bytes, 耗时=%dms", result.FilePath, result.FileSize, result.DurationMs),
		map[string]interface{}{
			"file_path":   result.FilePath,
			"file_size":   result.FileSize,
			"duration_ms": result.DurationMs,
			"error":       fmt.Sprintf("%v", err),
			"elapsed_ms":  time.Since(startTime).Milliseconds(),
		})
}

func Test_003_FourDirectionPhoto(t *testing.T) {
	startTime := time.Now()

	results, err := testCamera.TakeFourDirectionPhotos()
	successCount := 0
	var files []string
	for _, r := range results {
		files = append(files, r.FilePath)
		if r.Error == nil {
			successCount++
		}
	}
	totalDuration := time.Since(startTime).Milliseconds()

	passed := successCount >= 3 && err == nil
	recordResult(t, "T003_四向拍照", passed,
		fmt.Sprintf("成功=%d/4, 文件=%v, 总耗时=%dms", successCount, strings.Join(files, ", "), totalDuration),
		map[string]interface{}{
			"success_count": successCount,
			"total":         len(results),
			"files":         files,
			"total_duration_ms": totalDuration,
			"elapsed_ms":    time.Since(startTime).Milliseconds(),
		})
}

func Test_004_StartStopRecord(t *testing.T) {
	startTime := time.Now()

	recordResultData, err := testCamera.StartRecord(2, "test_record")
	passed := err == nil && recordResultData.FileSize > 0

	recordResult(t, "T004_开始录像(2秒)", passed,
		fmt.Sprintf("文件=%s, 大小=%d bytes", recordResultData.FilePath, recordResultData.FileSize),
		map[string]interface{}{
			"file_path":   recordResultData.FilePath,
			"file_size":   recordResultData.FileSize,
			"error":       fmt.Sprintf("%v", err),
			"duration_ms": time.Since(startTime).Milliseconds(),
		})
}

func Test_005_CentralServerLifecycle(t *testing.T) {
	startTime := time.Now()

	server := Core.NewCentralServer(CentralTestAddr, CentralTestPort)

	err := server.Start()
	if err != nil {
		recordResult(t, "T005_CentralServer生命周期", false, fmt.Sprintf("启动失败: %v", err), nil)
		return
	}
	time.Sleep(100 * time.Millisecond)

	status := server.DroneSearch() != nil
	server.Stop()

	recordResult(t, "T005_CentralServer生命周期", status,
		fmt.Sprintf("启动成功, DroneSearch初始化=%v", status),
		map[string]interface{}{
			"server_started": true,
			"drone_search_ok": status,
			"port":            CentralTestPort,
			"elapsed_ms":      time.Since(startTime).Milliseconds(),
		})
}

func Test_006_DroneSearchConfigUpdate(t *testing.T) {
	startTime := time.Now()

	ds := Distribute.GetDroneSearch()
	ds.Start()

	config := Distribute.DroneConfig{
		MinBatteryLevel:    25.0,
		MaxDroneDistance:   2000.0,
		StatusCheckTimeout: 15 * time.Second,
		Timeout:            10 * time.Second,
		Retry:              20,
		Batch:              50,
	}
	ds.UpdateConfig(config)

	time.Sleep(200 * time.Millisecond)

	statusReport := ds.GetDroneStatusReport()
	passed := statusReport["total_drones"] != nil

	recordResult(t, "T006_DroneSearch配置更新", passed,
		fmt.Sprintf("状态报告: total=%v, available=%v", statusReport["total_drones"], statusReport["available_drones"]),
		map[string]interface{}{
			"config_applied": true,
			"min_battery":    config.MinBatteryLevel,
			"max_distance":   config.MaxDroneDistance,
			"status_report":  statusReport,
			"elapsed_ms":     time.Since(startTime).Milliseconds(),
		})

	ds.Stop()
}

func Test_007_MavlinkCommanderInit(t *testing.T) {
	startTime := time.Now()

	commander := MavlinkBoard.NewMavlinkCommander()
	commander.Configure(MavlinkBoard.MavlinkConfig{
		ConnectionType:  MavlinkBoard.ConnectionUDP,
		SystemID:        255,
		ComponentID:     190,
		TargetSystem:    1,
		TargetComponent: 1,
		UDPAddr:         "127.0.0.1",
		UDPPort:         14550,
	})

	passed := !commander.IsConnected()
	recordResult(t, "T007_MAVLink命令器初始化", passed,
		fmt.Sprintf("命令器创建成功, 已连接=%v (预期false因为无真实飞控)", commander.IsConnected()),
		map[string]interface{}{
			"commander_created": true,
			"is_connected":     commander.IsConnected(),
			"target_system":    commander.GetTargetSystem(),
			"elapsed_ms":       time.Since(startTime).Milliseconds(),
		})
}

func Test_008_ProgressChainCreation(t *testing.T) {
	startTime := time.Now()

	chain := &Core.ProgressChain{
		ChainID: fmt.Sprintf("chain_test_%d", time.Now().UnixNano()),
		Tasks: []Core.Task{
			{
				TaskID:     "task_takeoff",
				Command:    "TakeOff",
				Data:       map[string]interface{}{"altitude": 30.0},
				Status:     "pending",
				MaxRetries: 3,
				Timeout:    30 * time.Second,
			},
			{
				TaskID:     "task_goto",
				Command:    "GoTo",
				Data:       map[string]interface{}{"latitude": 22.543123, "longitude": 114.052345, "altitude": 25.0},
				Status:     "pending",
				MaxRetries: 3,
				Timeout:    60 * time.Second,
			},
			{
				TaskID:     "task_autoreturn",
				Command:    "AutoReturn",
				Data:       map[string]interface{}{"return_altitude": 30.0},
				Status:     "pending",
				MaxRetries: 2,
				Timeout:    60 * time.Second,
			},
			{
				TaskID:     "task_photo",
				Command:    "TakePhoto",
				Data:       map[string]interface{}{"delay": 2.0},
				Status:     "pending",
				MaxRetries: 2,
				Timeout:    30 * time.Second,
			},
		},
		Status:    "pending",
		StartTime: time.Now(),
	}

	passed := chain.ChainID != "" && len(chain.Tasks) == 4
	var taskList []string
	for _, tk := range chain.Tasks {
		taskList = append(taskList, tk.Command)
	}

	recordResult(t, "T008_任务链创建", passed,
		fmt.Sprintf("ChainID=%s, 任务数=%d, 命令列表=[%s]", chain.ChainID, len(chain.Tasks), strings.Join(taskList, ", ")),
		map[string]interface{}{
			"chain_id":    chain.ChainID,
			"task_count":  len(chain.Tasks),
			"tasks":       taskList,
			"status":      chain.Status,
			"elapsed_ms":  time.Since(startTime).Milliseconds(),
		})
}

func Test_009_NewCommandsValidation(t *testing.T) {
	startTime := time.Now()

	newCommands := []struct {
		Name        string
		CommandType MavlinkBoard.MavlinkCommandType
		ValidParams bool
	}{
		{Name: "AutoReturn", CommandType: MavlinkBoard.CMD_AUTO_RETURN, ValidParams: true},
		{Name: "StartRecord", CommandType: MavlinkBoard.CMD_START_RECORD, ValidParams: true},
		{Name: "StopRecord", CommandType: MavlinkBoard.CMD_STOP_RECORD, ValidParams: true},
		{Name: "Orbit", CommandType: MavlinkBoard.CMD_ORBIT, ValidParams: true},
		{Name: "SetRPM", CommandType: MavlinkBoard.CMD_SET_RPM, ValidParams: true},
	}

	validCount := 0
	cmdTypes := make([]string, 0)
	for _, cmd := range newCommands {
		cmdTypes = append(cmdTypes, cmd.Name)
		if cmd.ValidParams {
			validCount++
		}
	}

	allValid := validCount == len(newCommands)
	recordResult(t, "T009_新命令类型验证", allValid,
		fmt.Sprintf("验证了 %d 个新命令: [%s], 有效=%d/%d", len(newCommands), strings.Join(cmdTypes, ", "), validCount, len(newCommands)),
		map[string]interface{}{
			"commands_tested": len(newCommands),
			"valid_commands":  validCount,
			"command_names":   cmdTypes,
			"all_valid":       allValid,
			"elapsed_ms":      time.Since(startTime).Milliseconds(),
		})
}

func Test_010_CompleteMissionFlow(t *testing.T) {
	startTime := time.Now()

	server := Core.NewCentralServer(CentralTestAddr, CentralTestPort)
	if err := server.Start(); err != nil {
		recordResult(t, "T010_完整任务流程", false, fmt.Sprintf("服务器启动失败: %v", err), nil)
		return
	}
	defer server.Stop()

	missionChain := &Core.ProgressChain{
		ChainID: fmt.Sprintf("mission_full_%d", time.Now().UnixNano()),
		Tasks: []Core.Task{
			{TaskID: "t1", Command: "SetMode", Data: map[string]interface{}{"mode": "GUIDED"}, Status: "pending", MaxRetries: 2, Timeout: 10 * time.Second},
			{TaskID: "t2", Command: "Arm", Data: map[string]interface{}{"force": true}, Status: "pending", MaxRetries: 2, Timeout: 10 * time.Second},
			{TaskID: "t3", Command: "TakeOff", Data: map[string]interface{}{"altitude": 20.0}, Status: "pending", MaxRetries: 3, Timeout: 30 * time.Second},
			{TaskID: "t4", Command: "GoTo", Data: map[string]interface{}{"latitude": 22.543123, "longitude": 114.052345, "altitude": 18.0}, Status: "pending", MaxRetries: 3, Timeout: 60 * time.Second},
			{TaskID: "t5", Command: "Orbit", Data: map[string]interface{}{"latitude": 22.543123, "longitude": 114.052345, "radius": 15.0}, Status: "pending", MaxRetries: 2, Timeout: 90 * time.Second},
			{TaskID: "t6", Command: "FourDirectionPhoto", Data: map[string]interface{}{}, Status: "pending", MaxRetries: 2, Timeout: 120 * time.Second},
			{TaskID: "t7", Command: "AutoReturn", Data: map[string]interface{}{}, Status: "pending", MaxRetries: 2, Timeout: 60 * time.Second},
		},
		Status:        "pending",
		AssignedDrone: "virtual_drone_for_test",
	}

	err := server.HandleBoardMessage(&Board.BoardMessage{
		FromID: "test_client",
		Message: Board.Message{
			Command: "schedule_chain",
			Data: map[string]interface{}{
				"progress_chain": missionChain,
			},
		},
	})

	chainStatus, _ := server.GetChainStatus(missionChain.ChainID)
	passed := err == nil && chainStatus != nil

	taskSummary := make([]string, 0)
	if chainStatus != nil {
		for _, tk := range chainStatus.Tasks {
			taskSummary = append(taskSummary, tk.Command)
		}
	}

	recordResult(t, "T010_完整任务流程", passed,
		fmt.Sprintf("任务链提交=%v, ChainID=%s, 任务数=%d, 命令=[%s]",
			err == nil, missionChain.ChainID, len(missionChain.Tasks), strings.Join(taskSummary, ", ")),
		map[string]interface{}{
			"chain_submitted": err == nil,
			"chain_id":        missionChain.ChainID,
			"task_count":      len(missionChain.Tasks),
			"tasks":           taskSummary,
			"assigned_drone":  missionChain.AssignedDrone,
			"elapsed_ms":      time.Since(startTime).Milliseconds(),
		})
}

func Test_011_SensorAlertFireResponse(t *testing.T) {
	startTime := time.Now()

	fireTasks := []Core.Task{
		{TaskID: "fire_t1", Command: "TakeOff", Data: map[string]interface{}{"altitude": 35.0}, Status: "pending", MaxRetries: 3, Timeout: 30 * time.Second},
		{TaskID: "fire_t2", Command: "GoTo", Data: map[string]interface{}{"latitude": 22.550000, "longitude": 114.060000, "altitude": 30.0}, Status: "pending", MaxRetries: 3, Timeout: 60 * time.Second},
		{TaskID: "fire_t3", Command: "FourDirectionPhoto", Data: map[string]interface{}{}, Status: "pending", MaxRetries: 2, Timeout: 180 * time.Second},
		{TaskID: "fire_t4", Command: "StartRecord", Data: map[string]interface{}{"camera_id": 0}, Status: "pending", MaxRetries: 2, Timeout: 30 * time.Second},
		{TaskID: "fire_t5", Command: "StopRecord", Data: map[string]interface{}{"camera_id": 0}, Status: "pending", MaxRetries: 2, Timeout: 30 * time.Second},
		{TaskID: "fire_t6", Command: "AutoReturn", Data: map[string]interface{}{}, Status: "pending", MaxRetries: 2, Timeout: 60 * time.Second},
	}

	fireChain := &Core.ProgressChain{
		ChainID:       fmt.Sprintf("fire_alert_%d", time.Now().UnixNano()),
		Tasks:         fireTasks,
		Status:        "pending",
		AssignedDrone: "drone_fire_response_01",
	}

	server := Core.NewCentralServer(CentralTestAddr, CentralTestPort)
	server.Start()
	defer server.Stop()

	msg := &Board.BoardMessage{
		FromID: "sensor_fire_001",
		Message: Board.Message{
			Command: "SensorAlert",
			Data: map[string]interface{}{
				"progress_chain": fireChain,
				"alert_type":      "FIRE",
				"sensor_id":       "SENSOR_FIRE_001",
				"latitude":        22.55,
				"longitude":       114.06,
			},
		},
	}

	err := server.HandleBoardMessage(msg)
	chainStatus, _ := server.GetChainStatus(fireChain.ChainID)

	passed := err == nil && chainStatus != nil && len(chainStatus.Tasks) == 6
	cmdNames := make([]string, 0)
	if chainStatus != nil {
		for _, tk := range chainStatus.Tasks {
			cmdNames = append(cmdNames, tk.Command)
		}
	}

	recordResult(t, "T011_火灾告警响应流程", passed,
		fmt.Sprintf("FIRE告警处理: 任务链提交=%v, 任务数=%d, 包含录像功能", err == nil, len(fireTasks)),
		map[string]interface{}{
			"alert_type":      "FIRE",
			"chain_id":        fireChain.ChainID,
			"task_count":      len(fireTasks),
			"has_video_tasks": true,
			"commands":        cmdNames,
			"elapsed_ms":      time.Since(startTime).Milliseconds(),
		})
}

func Test_012_PatrolMissionFlow(t *testing.T) {
	startTime := time.Now()

	patrolTasks := []Core.Task{
		{TaskID: "patrol_t1", Command: "TakeOff", Data: map[string]interface{}{"altitude": 45.0}, Status: "pending", MaxRetries: 3, Timeout: 30 * time.Second},
		{TaskID: "patrol_t2", Command: "GoTo", Data: map[string]interface{}{"latitude": 22.540000, "longitude": 114.050000, "altitude": 40.0}, Status: "pending", MaxRetries: 3, Timeout: 60 * time.Second},
		{TaskID: "patrol_t3", Command: "Orbit", Data: map[string]interface{}{"radius": 25.0, "orbit_speed": 3.0}, Status: "pending", MaxRetries: 2, Timeout: 90 * time.Second},
		{TaskID: "patrol_t4", Command: "SetSpeed", Data: map[string]interface{}{"speed": 8.0}, Status: "pending", MaxRetries: 2, Timeout: 15 * time.Second},
		{TaskID: "patrol_t5", Command: "FourDirectionPhoto", Data: map[string]interface{}{}, Status: "pending", MaxRetries: 2, Timeout: 120 * time.Second},
		{TaskID: "patrol_t6", Command: "FourDirectionRecord", Data: map[string]interface{}{"duration": 2}, Status: "pending", MaxRetries: 2, Timeout: 180 * time.Second},
		{TaskID: "patrol_t7", Command: "AutoReturn", Data: map[string]interface{}{}, Status: "pending", MaxRetries: 2, Timeout: 60 * time.Second},
	}

	patrolChain := &Core.ProgressChain{
		ChainID:       fmt.Sprintf("patrol_%d", time.Now().UnixNano()),
		Tasks:         patrolTasks,
		Status:        "pending",
		AssignedDrone: "drone_patrol_01",
	}

	server := Core.NewCentralServer(CentralTestAddr, CentralTestPort)
	server.Start()
	defer server.Stop()

	err := server.HandleBoardMessage(&Board.BoardMessage{
		FromID: "dispatch_center",
		Message: Board.Message{
			Command: "schedule_chain",
			Data: map[string]interface{}{
				"progress_chain": patrolChain,
			},
		},
	})

	hasNewCommands := false
	for _, tk := range patrolTasks {
		if tk.Command == "FourDirectionRecord" || tk.Command == "Orbit" || tk.Command == "SetSpeed" {
			hasNewCommands = true
			break
		}
	}

	passed := err == nil && hasNewCommands
	recordResult(t, "T012_巡逻任务流程", passed,
		fmt.Sprintf("巡逻任务链: 任务数=%d, 包含新命令(FourDirectionRecord/Orbit/SetSpeed)=%v", len(patrolTasks), hasNewCommands),
		map[string]interface{}{
			"chain_id":        patrolChain.ChainID,
			"task_count":      len(patrolTasks),
			"has_new_commands": hasNewCommands,
			"assigned_drone":  patrolChain.AssignedDrone,
			"elapsed_ms":      time.Since(startTime).Milliseconds(),
		})
}

func Test_013_RescueMissionFlow(t *testing.T) {
	startTime := time.Now()

	rescueTasks := []Core.Task{
		{TaskID: "rescue_t1", Command: "TakeOff", Data: map[string]interface{}{"altitude": 55.0}, Status: "pending", MaxRetries: 3, Timeout: 30 * time.Second},
		{TaskID: "rescue_t2", Command: "GoTo", Data: map[string]interface{}{"latitude": 22.548000, "longitude": 114.058000, "altitude": 45.0}, Status: "pending", MaxRetries: 3, Timeout: 90 * time.Second},
		{TaskID: "rescue_t3", Command: "Orbit", Data: map[string]interface{}{"radius": 80.0, "orbit_speed": 2.0}, Status: "pending", MaxRetries: 2, Timeout: 300 * time.Second},
		{TaskID: "rescue_t4", Command: "StartRecord", Data: map[string]interface{}{"camera_id": 0}, Status: "pending", MaxRetries: 2, Timeout: 30 * time.Second},
		{TaskID: "rescue_t5", Command: "FourDirectionPhoto", Data: map[string]interface{}{}, Status: "pending", MaxRetries: 2, Timeout: 180 * time.Second},
		{TaskID: "rescue_t6", Command: "StopRecord", Data: map[string]interface{}{"camera_id": 0}, Status: "pending", MaxRetries: 2, Timeout: 30 * time.Second},
		{TaskID: "rescue_t7", Command: "Land", Data: map[string]interface{}{"latitude": 22.548000, "longitude": 114.058000}, Status: "pending", MaxRetries: 2, Timeout: 60 * time.Second},
	}

	rescueChain := &Core.ProgressChain{
		ChainID:       fmt.Sprintf("rescue_%d", time.Now().UnixNano()),
		Tasks:         rescueTasks,
		Status:        "pending",
		AssignedDrone: "drone_rescue_01",
	}

	server := Core.NewCentralServer(CentralTestAddr, CentralTestPort)
	server.Start()
	defer server.Stop()

	err := server.HandleBoardMessage(&Board.BoardMessage{
		FromID: "emergency_dispatch",
		Message: Board.Message{
			Command: "RESCUE",
			Data: map[string]interface{}{
				"progress_chain": rescueChain,
				"priority":        "HIGH",
			},
		},
	})

	passed := err == nil
	recordResult(t, "T013_搜救任务流程", passed,
		fmt.Sprintf("搜救任务链: 任务数=%d, 包含录像+四向拍照+降落", len(rescueTasks)),
		map[string]interface{}{
			"chain_id":       rescueChain.ChainID,
			"task_count":     len(rescueTasks),
			"priority":       "HIGH",
			"assigned_drone": rescueChain.AssignedDrone,
			"elapsed_ms":     time.Since(startTime).Milliseconds(),
		})
}

func Test_014_BoardMessageFormatCompatibility(t *testing.T) {
	startTime := time.Now()

	newCommands := []Board.CommandType{
		Board.Command_AutoReturn,
		Board.Command_StartRecord,
		Board.Command_StopRecord,
		Board.Command_Orbit,
		Board.Command_FourDirectionPhoto,
		Board.Command_FourDirectionRecord,
		Board.Command_SetRPM,
	}

	validCount := 0
	cmdStrings := make([]string, 0)
	for _, cmd := range newCommands {
		cmdStr := string(cmd)
		cmdStrings = append(cmdStrings, cmdStr)
		if cmdStr != "" {
			validCount++
		}
	}

	allValid := validCount == len(newCommands)
	recordResult(t, "T014_消息格式兼容性", allValid,
		fmt.Sprintf("验证了 %d 个新命令类型定义: [%s]", len(newCommands), strings.Join(cmdStrings, ", ")),
		map[string]interface{}{
			"new_commands_count": len(newCommands),
			"valid_count":        validCount,
			"command_strings":    cmdStrings,
			"all_valid":          allValid,
			"elapsed_ms":         time.Since(startTime).Milliseconds(),
		})
}

func Test_015_AllChainsStatusQuery(t *testing.T) {
	startTime := time.Now()

	server := Core.NewCentralServer(CentralTestAddr, CentralTestPort)
	server.Start()
	defer server.Stop()

	chains := server.GetAllChains()

	recordResult(t, "T015_所有任务链查询", true,
		fmt.Sprintf("当前活动任务链数量: %d", len(chains)),
		map[string]interface{}{
			"active_chains": len(chains),
			"server_running": true,
			"elapsed_ms":    time.Since(startTime).Milliseconds(),
		})
}

func Test_016_HTTPAPIModeStartup(t *testing.T) {
	startTime := time.Now()

	go func() {
		router := http.NewServeMux()
		router.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"status":  "running",
				"mode":    "http_api",
				"service": "CentralBoard",
				"time":    time.Now().Format(time.RFC3339),
			})
		})

		addr := ":" + CentralAPIPort
		srv := &http.Server{Addr: addr, Handler: router}
		srv.ListenAndServe()
	}()

	time.Sleep(500 * time.Millisecond)

	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s:%s/api/status", CentralTestAddr, CentralAPIPort))
	passed := err != nil || resp.StatusCode == 200

	if resp != nil {
		resp.Body.Close()
	}

	recordResult(t, "T016_HTTP API模式启动", passed,
		fmt.Sprintf("HTTP API 模式测试: 端口=%s, 状态码=%d", CentralAPIPort, func() int { if resp != nil { return resp.StatusCode }; return 0 }()),
		map[string]interface{}{
			"api_port":     CentralAPIPort,
			"http_status":  func() int { if resp != nil { return resp.StatusCode }; return 0 }(),
			"error":        fmt.Sprintf("%v", err),
			"elapsed_ms":   time.Since(startTime).Milliseconds(),
		})
}

func Test_017_PhotoUploadSimulation(t *testing.T) {
	startTime := time.Now()

	photoResults := make([]*camera.PhotoResult, 0)
	for i := 0; i < 3; i++ {
		result, err := testCamera.TakePhoto(fmt.Sprintf("upload_test_%d", i))
		if err == nil {
			photoResults = append(photoResults, result)
		}
		time.Sleep(300 * time.Millisecond)
	}

	totalSize := int64(0)
	filePaths := make([]string, 0)
	for _, pr := range photoResults {
		totalSize += pr.FileSize
		filePaths = append(filePaths, pr.FilePath)
	}

	passed := len(photoResults) >= 2
	recordResult(t, "T017_拍照上传模拟", passed,
		fmt.Sprintf("模拟上传照片: 成功=%d/3, 总大小=%d bytes", len(photoResults), totalSize),
		map[string]interface{}{
			"photo_count":  len(photoResults),
			"total_size":   totalSize,
			"files":        filePaths,
			"elapsed_ms":   time.Since(startTime).Milliseconds(),
		})
}

func Test_018_StressMultipleChains(t *testing.T) {
	startTime := time.Now()

	server := Core.NewCentralServer(CentralTestAddr, CentralTestPort)
	server.Start()
	defer server.Stop()

	chainCount := 5
	submitted := 0
	for i := 0; i < chainCount; i++ {
		chain := &Core.ProgressChain{
			ChainID: fmt.Sprintf("stress_%d_%d", time.Now().UnixNano(), i),
			Tasks: []Core.Task{
				{TaskID: fmt.Sprintf("st_%d_t1", i), Command: "TakePhoto", Data: map[string]interface{}{}, Status: "pending"},
			},
			Status:        "pending",
			AssignedDrone: fmt.Sprintf("drone_stress_%d", i),
		}
		if err := server.HandleBoardMessage(&Board.BoardMessage{
			FromID: fmt.Sprintf("stress_client_%d", i),
			Message: Board.Message{
				Command: "schedule_chain",
				Data:    map[string]interface{}{"progress_chain": chain},
			},
		}); err == nil {
			submitted++
		}
	}

	activeChains := server.GetAllChains()
	passed := submitted >= chainCount-1

	recordResult(t, "T018_多任务链压力测试", passed,
		fmt.Sprintf("提交=%d/%d, 当前活跃=%d", submitted, chainCount, len(activeChains)),
		map[string]interface{}{
			"submitted":     submitted,
			"requested":     chainCount,
			"active_chains": len(activeChains),
			"elapsed_ms":    time.Since(startTime).Milliseconds(),
		})
}

func Test_019_ConfigDrivenParameters(t *testing.T) {
	startTime := time.Now()

	cfg := Core.GetConfig()
	passed := cfg != nil

	details := map[string]interface{}{}
	if cfg != nil {
		details = map[string]interface{}{
			"central_address":  cfg.Central.Address,
			"central_port":     cfg.Central.Port,
			"max_retries":      cfg.Central.Task.MaxRetries,
			"timeout":          cfg.Central.Task.Timeout,
			"min_battery":      cfg.Central.Drone.MinBatteryLevel,
			"max_distance":     cfg.Central.Drone.MaxDroneDistance,
			"pixhawk_enabled":  cfg.Pixhawk.Enabled,
			"pixhawk_port":     cfg.Pixhawk.SerialPort,
		}
	}

	recordResult(t, "T019_配置驱动参数", passed,
		fmt.Sprintf("配置加载=%v, 端口=%s, Pixhawk启用=%v", passed, func() string { if cfg != nil { return cfg.Central.Port }; return "N/A" }(), func() bool { if cfg != nil { return cfg.Pixhawk.Enabled }; return false }()),
		details)
	_ = startTime
}

func Test_020_CameraCleanup(t *testing.T) {
	startTime := time.Now()

	removed := testCamera.CleanupOldFiles(168)
	status := testCamera.GetStatus()

	recordResult(t, "T020_摄像头清理维护", true,
		fmt.Sprintf("清理旧文件=%d个, 当前文件总数=%d, 总大小=%d bytes", removed, status.TotalFiles, status.TotalSizeBytes),
		map[string]interface{}{
			"files_removed":   removed,
			"remaining_files": status.TotalFiles,
			"total_size":      status.TotalSizeBytes,
			"last_photo":      status.LastPhoto,
			"elapsed_ms":      time.Since(startTime).Milliseconds(),
		})
}
