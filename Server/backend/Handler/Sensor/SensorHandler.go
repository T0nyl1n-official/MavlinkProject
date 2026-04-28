package Sensor

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	BoardSensorHandler "MavlinkProject/Server/backend/Handler/Boards/SensorBoard"
	FRPHelper "MavlinkProject/Server/backend/Utils/FRPHelper"
)

type SensorAlert struct {
	SensorID   string                 `json:"sensor_id"`
	SensorIP   string                 `json:"sensor_ip"`
	SensorName string                 `json:"sensor_name"`
	AlertType  string                 `json:"alert_type"`
	AlertMsg   string                 `json:"alert_msg"`
	Latitude   float64                `json:"latitude"`
	Longitude  float64                `json:"longitude"`
	Timestamp  int64                  `json:"timestamp"`
	Severity   string                 `json:"severity"`
	Data       map[string]interface{} `json:"data"`
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

	if alert.AlertType == "none" || alert.AlertType == "None" || alert.AlertType == "NONE" {
		log.Printf("[SensorRoutes] Alert type is 'none', indicating no actual alarm. Drone scheduling skipped.")
		c.JSON(200, gin.H{
			"code":       0,
			"message":    "Alert received - type is none, no drone dispatched",
			"sensor_id":  alert.SensorID,
			"alert_type": alert.AlertType,
		})
		return
	}

	tasks := generateTasksByAlertType(alert)
	if len(tasks) == 0 {
		log.Printf("[SensorRoutes] No tasks generated for alert type: %s (Ignored)", alert.AlertType)
		c.JSON(202, gin.H{
			"code":       0,
			"message":    "Alert received - Ignored or Unknown alert type",
			"sensor_id":  alert.SensorID,
			"alert_type": alert.AlertType,
		})
		return
	}

	client := FRPHelper.GetCentralClient()
	if client == nil {
		log.Printf("[SensorRoutes] Warning: CentralBoard client not available")
	} else {
		drones, err := client.GetAvailableDrones()
		if err != nil || len(drones) == 0 {
			log.Printf("[SensorRoutes] No available drones from CentralBoard: %v", err)
		} else if len(drones) > 0 {
			bestDrone := drones[0]
			frpReq := BoardSensorHandler.SensorAlertReq{
				SensorID:   alert.SensorID,
				Latitude:   alert.Latitude,
				Longitude:  alert.Longitude,
				Radius:     50,
				PhotoCount: 1,
				Altitude:   20,
			}

			err := BoardSensorHandler.GenerateChainAndSendToCentral(frpReq)
			if err != nil {
				log.Printf("[SensorRoutes] Failed to send chain to central: %v", err)
				c.JSON(202, gin.H{
					"code":       0,
					"message":    "Alert received, logged - Failed to send to Central",
					"sensor_id":  alert.SensorID,
					"alert_type": alert.AlertType,
					"drones":     len(drones),
					"error":      err.Error(),
				})
				return
			}

			log.Printf("[SensorRoutes] Task chain sent to CentralBoard via FRP")
			c.JSON(200, gin.H{
				"code":           0,
				"message":        "Task chain created and sent to CentralBoard",
				"sensor_id":      alert.SensorID,
				"alert_type":     alert.AlertType,
				"assigned_drone": bestDrone.BoardID,
				"task_count":     len(tasks),
			})
			return
		}
	}

	frpReq := BoardSensorHandler.SensorAlertReq{
		SensorID:   alert.SensorID,
		Latitude:   alert.Latitude,
		Longitude:  alert.Longitude,
		Radius:     50,
		PhotoCount: 1,
		Altitude:   20,
	}

	err := BoardSensorHandler.GenerateChainAndSendToCentral(frpReq)
	if err != nil {
		log.Printf("[SensorRoutes] Failed to send chain to central: %v", err)
		c.JSON(202, gin.H{
			"code":       0,
			"message":    "Alert received, logged - fail to send to central",
			"sensor_id":  alert.SensorID,
			"alert_type": alert.AlertType,
			"drones":     0,
			"error":      err.Error(),
		})
		return
	}

	log.Printf("[SensorRoutes] Successfully generated and sent mission chain to Central via FRP.")
	c.JSON(200, gin.H{
		"code":       0,
		"message":    "Task chain created and sent to Central via FRP",
		"sensor_id":  alert.SensorID,
		"alert_type": alert.AlertType,
		"drones":     0,
	})
	return
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

func generateTasksByAlertType(alert SensorAlert) []Task {
	var tasks []Task

	switch alert.AlertType {
	case "fire", "Fire", "FIRE":
		tasks = []Task{
			{TaskID: "task_0", Command: "TakeOff", Data: map[string]interface{}{"altitude": 30.0}, Status: "pending", MaxRetries: 3, Timeout: 30 * time.Second},
			{TaskID: "task_1", Command: "GoTo", Data: map[string]interface{}{"latitude": alert.Latitude, "longitude": alert.Longitude, "altitude": 25.0}, Status: "pending", MaxRetries: 3, Timeout: 60 * time.Second},
			{TaskID: "task_2", Command: "FourDirectionPhoto", Data: map[string]interface{}{"latitude": alert.Latitude, "longitude": alert.Longitude, "radius": 50.0}, Status: "pending", MaxRetries: 2, Timeout: 180 * time.Second},
		}
	case "rescue", "Rescue", "RESCUE", "missing", "Missing", "MISSING":
		tasks = []Task{
			{TaskID: "task_0", Command: "TakeOff", Data: map[string]interface{}{"altitude": 50.0}, Status: "pending", MaxRetries: 3, Timeout: 30 * time.Second},
			{TaskID: "task_1", Command: "GoTo", Data: map[string]interface{}{"latitude": alert.Latitude, "longitude": alert.Longitude, "altitude": 40.0}, Status: "pending", MaxRetries: 3, Timeout: 90 * time.Second},
			{TaskID: "task_2", Command: "Orbit", Data: map[string]interface{}{"latitude": alert.Latitude, "longitude": alert.Longitude, "radius": 100.0}, Status: "pending", MaxRetries: 2, Timeout: 300 * time.Second},
			{TaskID: "task_3", Command: "Land", Data: map[string]interface{}{}, Status: "pending", MaxRetries: 2, Timeout: 60 * time.Second},
		}
	case "patrol", "Patrol", "PATROL":
		tasks = []Task{
			{TaskID: "task_0", Command: "TakeOff", Data: map[string]interface{}{"altitude": 40.0}, Status: "pending", MaxRetries: 3, Timeout: 30 * time.Second},
			{TaskID: "task_1", Command: "GoTo", Data: map[string]interface{}{"latitude": alert.Latitude, "longitude": alert.Longitude, "altitude": 35.0}, Status: "pending", MaxRetries: 3, Timeout: 60 * time.Second},
			{TaskID: "task_2", Command: "Orbit", Data: map[string]interface{}{"latitude": alert.Latitude, "longitude": alert.Longitude, "radius": 30.0}, Status: "pending", MaxRetries: 2, Timeout: 90 * time.Second},
			{TaskID: "task_3", Command: "AutoReturn", Data: map[string]interface{}{}, Status: "pending", MaxRetries: 2, Timeout: 60 * time.Second},
		}
	default:
		tasks = []Task{
			{TaskID: "task_0", Command: "TakeOff", Data: map[string]interface{}{"altitude": 30.0}, Status: "pending", MaxRetries: 3, Timeout: 30 * time.Second},
			{TaskID: "task_1", Command: "GoTo", Data: map[string]interface{}{"latitude": alert.Latitude, "longitude": alert.Longitude, "altitude": 25.0}, Status: "pending", MaxRetries: 3, Timeout: 60 * time.Second},
			{TaskID: "task_2", Command: "AutoReturn", Data: map[string]interface{}{}, Status: "pending", MaxRetries: 2, Timeout: 60 * time.Second},
		}
	}

	return tasks
}

func GetSensorStatus(c *gin.Context) {
	client := FRPHelper.GetCentralClient()
	if client == nil {
		c.JSON(500, gin.H{
			"code":    1,
			"message": "CentralBoard client not available",
		})
		return
	}

	drones, err := client.GetAvailableDrones()
	if err != nil {
		c.JSON(500, gin.H{
			"code":    1,
			"message": "Failed to get drone status",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"code":   0,
		"drones": drones,
	})
}
