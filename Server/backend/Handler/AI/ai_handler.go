package AI

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"

	Models "MavlinkProject/Models"
)

type SensorAnalysisRequest struct {
	SensorID   string                   `json:"sensor_id" binding:"required"`
	SensorType string                   `json:"sensor_type" binding:"required"`
	TimeSeries []Models.TimeSeriesPoint `json:"time_series" binding:"required"`
	Latitude   float64                  `json:"latitude"`
	Longitude  float64                  `json:"longitude"`
	Metadata   map[string]interface{}   `json:"metadata,omitempty"`
}

type DroneImageAnalysisRequest struct {
	DroneID     string  `json:"drone_id" binding:"required"`
	ImageBase64 string  `json:"image_base64"`
	ImageURL    string  `json:"image_url"`
	Latitude    float64 `json:"latitude"`
	Longitude   float64 `json:"longitude"`
	Confidence  float64 `json:"confidence"`
}

type AlertHistoryStore struct {
	alerts []Models.AlertJSON
	mu     sync.RWMutex
	maxLen int
}

var alertHistory *AlertHistoryStore

func init() {
	alertHistory = &AlertHistoryStore{
		alerts: make([]Models.AlertJSON, 0),
		maxLen: 1000,
	}
}

func (s *AlertHistoryStore) Add(alert Models.AlertJSON) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts = append(s.alerts, alert)
	if len(s.alerts) > s.maxLen {
		s.alerts = s.alerts[len(s.alerts)-s.maxLen:]
	}
}

func (s *AlertHistoryStore) GetAll() []Models.AlertJSON {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]Models.AlertJSON, len(s.alerts))
	copy(result, s.alerts)
	return result
}

func (s *AlertHistoryStore) GetBySeverity(severity string) []Models.AlertJSON {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []Models.AlertJSON
	for _, a := range s.alerts {
		if a.Severity == severity {
			result = append(result, a)
		}
	}
	return result
}

func (s *AlertHistoryStore) GetRecent(count int) []Models.AlertJSON {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if count > len(s.alerts) {
		count = len(s.alerts)
	}
	start := len(s.alerts) - count
	result := make([]Models.AlertJSON, count)
	copy(result, s.alerts[start:])
	return result
}

func HandleSensorAnalysis(c *gin.Context) {
	var req SensorAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	if req.SensorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "sensor_id is required",
		})
		return
	}

	if len(req.TimeSeries) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "time_series is required and must not be empty",
		})
		return
	}

	service := GetAnalysisService()

	alert, err := service.ProcessSensorData(
		req.SensorID,
		req.SensorType,
		req.TimeSeries,
		req.Latitude,
		req.Longitude,
	)
	if err != nil {
		log.Printf("[AI-Handler] 传感器分析失败: sensor=%s, err=%v", req.SensorID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    2,
			"message": "Analysis failed",
			"error":   err.Error(),
		})
		return
	}

	alertHistory.Add(*alert)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Analysis completed",
		"alert":   alert,
	})
}

func HandleDroneImageAnalysis(c *gin.Context) {
	var req DroneImageAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	if req.DroneID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "drone_id is required",
		})
		return
	}

	if req.ImageBase64 == "" && req.ImageURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Either image_base64 or image_url must be provided",
		})
		return
	}

	service := GetAnalysisService()

	alert, err := service.ProcessDroneImage(
		req.DroneID,
		req.ImageBase64,
		req.ImageURL,
		req.Latitude,
		req.Longitude,
	)
	if err != nil {
		log.Printf("[AI-Handler] 无人机图像分析失败: drone=%s, err=%v", req.DroneID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    2,
			"message": "Analysis failed",
			"error":   err.Error(),
		})
		return
	}

	alertHistory.Add(*alert)

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "Analysis completed",
		"alert":   alert,
	})
}

func HandleAlertHistory(c *gin.Context) {
	severity := c.Query("severity")
	limitStr := c.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	var alerts []Models.AlertJSON

	if severity != "" {
		alerts = alertHistory.GetBySeverity(severity)
		if len(alerts) > limit {
			alerts = alerts[len(alerts)-limit:]
		}
	} else {
		alerts = alertHistory.GetRecent(limit)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   0,
		"count":  len(alerts),
		"alerts": alerts,
	})
}
