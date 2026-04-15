package Sensor

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	Distribute "MavlinkProject/Server/backend/Utils/CentralBoard/Distribute"
)

type SensorAlert struct {
	SensorID   string                 `json:"sensor_id"`   // 传感器ID (必填)
	SensorIP   string                 `json:"sensor_ip"`   // 传感器IP地址
	SensorName string                 `json:"sensor_name"` // 传感器名称 (可选，未填则用SensorIP)
	AlertType  string                 `json:"alert_type"`  // 警报类型 (必填)
	AlertMsg   string                 `json:"alert_msg"`   // 预警消息
	Latitude   float64                `json:"latitude"`    // GPS纬度 (必填)
	Longitude  float64                `json:"longitude"`   // GPS经度 (必填)
	Timestamp  int64                  `json:"timestamp"`   // 时间戳，默认当前时间
	Severity   string                 `json:"severity"`    // 严重程度
	Data       map[string]interface{} `json:"data"`        // 传感器数据 (GAS/TEMP/PRESS/ALT/AMB/OBJ/WiFi/RSSI等)
}

func ReceiveSensorMessage(c *gin.Context) {
	var alert SensorAlert

	if err := c.ShouldBindJSON(&alert); err != nil {
		c.JSON(400, gin.H{
			"code":    1,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	if alert.SensorID == "" {
		c.JSON(400, gin.H{
			"code":    1,
			"message": "sensor_id is required",
		})
		return
	}

	alert.SensorIP = c.ClientIP()

	if alert.SensorName == "" {
		alert.SensorName = alert.SensorIP
	}

	if alert.Timestamp == 0 {
		alert.Timestamp = time.Now().Unix()
	}

	log.Printf("[SensorRoutes] Alert received: sensor=%s, type=%s, msg=%s, lat=%.6f, lon=%.6f",
		alert.SensorName, alert.AlertType, alert.AlertMsg, alert.Latitude, alert.Longitude)

	droneSearch := Distribute.GetDroneSearch()
	if droneSearch == nil {
		log.Printf("[SensorRoutes] Error: DroneSearch not available")
		c.JSON(500, gin.H{
			"code":    1,
			"message": "DroneSearch not available",
		})
		return
	}

	availableDrones := droneSearch.GetAvailableDrones()
	if len(availableDrones) == 0 {
		log.Printf("[SensorRoutes] Warning: No available drones for alert response")
		log.Printf("[SensorRoutes] Alert logged: sensor=%s, type=%s at (%.6f, %.6f) - No drones available",
			alert.SensorName, alert.AlertType, alert.Latitude, alert.Longitude)
		c.JSON(202, gin.H{
			"code":       0,
			"message":    "Alert received, logged - No available drones",
			"sensor_id":  alert.SensorID,
			"alert_type": alert.AlertType,
			"drones":     0,
		})
		return
	}

	bestDrone, err := droneSearch.FindBestDrone()
	if err != nil {
		log.Printf("[SensorRoutes] Warning: No suitable drone found: %v", err)
		log.Printf("[SensorRoutes] Alert logged: sensor=%s, type=%s - No suitable drone: %v",
			alert.SensorName, alert.AlertType, err)
		c.JSON(202, gin.H{
			"code":       0,
			"message":    "Alert received, logged - No suitable drone",
			"sensor_id":  alert.SensorID,
			"alert_type": alert.AlertType,
			"drones":     len(availableDrones),
		})
		return
	}

	tasks := generateTasksByAlertType(alert, bestDrone)
	if len(tasks) == 0 {
		log.Printf("[SensorRoutes] No tasks generated for alert type: %s", alert.AlertType)
		c.JSON(202, gin.H{
			"code":       0,
			"message":    "Alert received - Unknown alert type",
			"sensor_id":  alert.SensorID,
			"alert_type": alert.AlertType,
		})
		return
	}

	chainID := generateChainID(alert)
	chain := &Distribute.ProgressChain{
		ChainID:       chainID,
		Tasks:         tasks,
		CurrentTask:   0,
		Status:        "pending",
		StartTime:     time.Now(),
		AssignedDrone: bestDrone.BoardID,
	}

	err = droneSearch.SubmitProgressChain(chain)
	if err != nil {
		log.Printf("[SensorRoutes] Failed to submit progress chain: %v", err)
		c.JSON(202, gin.H{
			"code":    0,
			"message": "Alert received, logged - Failed to submit task chain",
			"error":   err.Error(),
		})
		return
	}

	log.Printf("[SensorRoutes] Task chain submitted: chain_id=%s, drone=%s, tasks=%d",
		chainID, bestDrone.BoardID, len(tasks))

	c.JSON(200, gin.H{
		"code":           0,
		"message":        "Task chain created and submitted",
		"chain_id":       chainID,
		"assigned_drone": bestDrone.BoardID,
		"task_count":     len(tasks),
		"alert_type":     alert.AlertType,
		"target": map[string]float64{
			"latitude":  alert.Latitude,
			"longitude": alert.Longitude,
		},
	})
}

// 任务分拣，根据告警类型生成不同的任务链
func generateTasksByAlertType(alert SensorAlert, drone *Distribute.DroneStatus) []Distribute.Task {
	var tasks []Distribute.Task

	switch alert.AlertType {
	case "fire", "Fire", "FIRE":
		tasks = []Distribute.Task{
			{
				TaskID:  "task_0",
				Command: "takeoff",
				Data: map[string]interface{}{
					"altitude": 30.0,
				},
				Status:     "pending",
				MaxRetries: 3,
				Timeout:    30 * time.Second,
			},
			{
				TaskID:  "task_1",
				Command: "goto",
				Data: map[string]interface{}{
					"latitude":  alert.Latitude,
					"longitude": alert.Longitude,
					"altitude":  25.0,
				},
				Status:     "pending",
				MaxRetries: 3,
				Timeout:    60 * time.Second,
			},
			{
				TaskID:  "task_2",
				Command: "survey",
				Data: map[string]interface{}{
					"latitude":  alert.Latitude,
					"longitude": alert.Longitude,
					"radius":    50.0,
					"duration":  120,
				},
				Status:     "pending",
				MaxRetries: 2,
				Timeout:    180 * time.Second,
			},
		}

	case "rescue", "Rescue", "RESCUE", "missing", "Missing", "MISSING":
		tasks = []Distribute.Task{
			{
				TaskID:  "task_0",
				Command: "takeoff",
				Data: map[string]interface{}{
					"altitude": 50.0,
				},
				Status:     "pending",
				MaxRetries: 3,
				Timeout:    30 * time.Second,
			},
			{
				TaskID:  "task_1",
				Command: "goto",
				Data: map[string]interface{}{
					"latitude":  alert.Latitude,
					"longitude": alert.Longitude,
					"altitude":  40.0,
				},
				Status:     "pending",
				MaxRetries: 3,
				Timeout:    90 * time.Second,
			},
			{
				TaskID:  "task_2",
				Command: "survey_grid",
				Data: map[string]interface{}{
					"latitude":  alert.Latitude,
					"longitude": alert.Longitude,
					"width":     200.0,
					"height":    200.0,
					"altitude":  30.0,
				},
				Status:     "pending",
				MaxRetries: 2,
				Timeout:    300 * time.Second,
			},
			{
				TaskID:  "task_3",
				Command: "land",
				Data: map[string]interface{}{
					"latitude":  alert.Latitude,
					"longitude": alert.Longitude,
				},
				Status:     "pending",
				MaxRetries: 2,
				Timeout:    60 * time.Second,
			},
		}

	case "patrol", "Patrol", "PATROL":
		tasks = []Distribute.Task{
			{
				TaskID:  "task_0",
				Command: "takeoff",
				Data: map[string]interface{}{
					"altitude": 40.0,
				},
				Status:     "pending",
				MaxRetries: 3,
				Timeout:    30 * time.Second,
			},
			{
				TaskID:  "task_1",
				Command: "goto",
				Data: map[string]interface{}{
					"latitude":  alert.Latitude,
					"longitude": alert.Longitude,
					"altitude":  35.0,
				},
				Status:     "pending",
				MaxRetries: 3,
				Timeout:    60 * time.Second,
			},
			{
				TaskID:  "task_2",
				Command: "orbit",
				Data: map[string]interface{}{
					"latitude":  alert.Latitude,
					"longitude": alert.Longitude,
					"radius":    30.0,
					"duration":  60,
				},
				Status:     "pending",
				MaxRetries: 2,
				Timeout:    90 * time.Second,
			},
			{
				TaskID:     "task_3",
				Command:    "return_to_home",
				Data:       map[string]interface{}{},
				Status:     "pending",
				MaxRetries: 2,
				Timeout:    60 * time.Second,
			},
		}

	case "flood", "Flood", "FLOOD":
		tasks = []Distribute.Task{
			{
				TaskID:  "task_0",
				Command: "takeoff",
				Data: map[string]interface{}{
					"altitude": 60.0,
				},
				Status:     "pending",
				MaxRetries: 3,
				Timeout:    30 * time.Second,
			},
			{
				TaskID:  "task_1",
				Command: "goto",
				Data: map[string]interface{}{
					"latitude":  alert.Latitude,
					"longitude": alert.Longitude,
					"altitude":  50.0,
				},
				Status:     "pending",
				MaxRetries: 3,
				Timeout:    90 * time.Second,
			},
			{
				TaskID:  "task_2",
				Command: "survey",
				Data: map[string]interface{}{
					"latitude":  alert.Latitude,
					"longitude": alert.Longitude,
					"radius":    100.0,
					"duration":  180,
				},
				Status:     "pending",
				MaxRetries: 2,
				Timeout:    240 * time.Second,
			},
		}

	default:
		tasks = []Distribute.Task{
			{
				TaskID:  "task_0",
				Command: "takeoff",
				Data: map[string]interface{}{
					"altitude": 30.0,
				},
				Status:     "pending",
				MaxRetries: 3,
				Timeout:    30 * time.Second,
			},
			{
				TaskID:  "task_1",
				Command: "goto",
				Data: map[string]interface{}{
					"latitude":  alert.Latitude,
					"longitude": alert.Longitude,
					"altitude":  25.0,
				},
				Status:     "pending",
				MaxRetries: 3,
				Timeout:    60 * time.Second,
			},
			{
				TaskID:     "task_2",
				Command:    "return_to_home",
				Data:       map[string]interface{}{},
				Status:     "pending",
				MaxRetries: 2,
				Timeout:    60 * time.Second,
			},
		}
	}

	return tasks
}

func generateChainID(alert SensorAlert) string {
	return time.Now().Format("20060102150405") + "_" + alert.AlertType
}

func GetSensorStatus(c *gin.Context) {
	droneSearch := Distribute.GetDroneSearch()
	if droneSearch == nil {
		c.JSON(500, gin.H{
			"code":    1,
			"message": "DroneSearch not available",
		})
		return
	}

	drones := droneSearch.GetAvailableDrones()
	c.JSON(200, gin.H{
		"code":   0,
		"drones": drones,
	})
}

func generateMessageID() string {
	return time.Now().Format("20060102150405.000")
}
