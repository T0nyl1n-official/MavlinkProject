package AI

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	Conf "MavlinkProject/Server/backend/Config"
	Models "MavlinkProject/Models"
)

type GasKickPredictRequest struct {
	WellName   string                   `json:"well_name" binding:"required"`
	SensorID   string                   `json:"sensor_id,omitempty"`
	TimeSeries []Models.TimeSeriesPoint `json:"time_series" binding:"required"`
	Latitude   float64                  `json:"latitude"`
	Longitude  float64                  `json:"longitude"`
}

type GasKickPredictResponse struct {
	Code      int                      `json:"code"`
	Message   string                   `json:"message"`
	Alert     *Models.AlertJSON         `json:"alert,omitempty"`
	RawResult *Models.GasKickResponse   `json:"raw_result,omitempty"`
}

var gasKickHistory struct {
	results []gasKickRecord
	mu      sync.RWMutex
	maxLen  int
}

type gasKickRecord struct {
	Timestamp    time.Time              `json:"timestamp"`
	WellName     string                 `json:"well_name"`
	SensorID     string                 `json:"sensor_id"`
	Result       Models.GasKickResponse `json:"result"`
	Alert        *Models.AlertJSON       `json:"alert,omitempty"`
}

func init() {
	gasKickHistory.results = make([]gasKickRecord, 0)
	gasKickHistory.maxLen = 500
}

func HandleSensorGasKick(c *gin.Context) {
	var req GasKickPredictRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, GasKickPredictResponse{
			Code:    1,
			Message: "Invalid request body: " + err.Error(),
		})
		return
	}

	if req.WellName == "" {
		c.JSON(http.StatusBadRequest, GasKickPredictResponse{
			Code:    1,
			Message: "well_name is required",
		})
		return
	}

	if len(req.TimeSeries) == 0 {
		c.JSON(http.StatusBadRequest, GasKickPredictResponse{
			Code:    1,
			Message: "time_series is required and must not be empty",
		})
		return
	}

	service := GetAnalysisService()

	setting := Conf.GetSettingManager().GetSetting()
	alertWindow := setting.AI.LSTM.AlertWindow
	alertThreshold := setting.AI.LSTM.AlertThreshold
	if alertWindow <= 0 {
		alertWindow = 5
	}
	if alertThreshold <= 0 {
		alertThreshold = 3
	}

	alert, rawResult, err := service.ProcessGasKick(
		req.WellName,
		req.SensorID,
		req.TimeSeries,
		req.Latitude,
		req.Longitude,
		alertWindow,
		alertThreshold,
	)
	if err != nil {
		log.Printf("[GasKickHandler] 气侵预测失败: well=%s, err=%v", req.WellName, err)
		c.JSON(http.StatusInternalServerError, GasKickPredictResponse{
			Code:    2,
			Message: "Gas kick prediction failed: " + err.Error(),
		})
		return
	}

	record := gasKickRecord{
		Timestamp: time.Now(),
		WellName:  req.WellName,
		SensorID:  req.SensorID,
		Result:    *rawResult,
		Alert:     alert,
	}
	gasKickHistory.mu.Lock()
	gasKickHistory.results = append(gasKickHistory.results, record)
	if len(gasKickHistory.results) > gasKickHistory.maxLen {
		gasKickHistory.results = gasKickHistory.results[len(gasKickHistory.results)-gasKickHistory.maxLen:]
	}
	gasKickHistory.mu.Unlock()

	c.JSON(http.StatusOK, GasKickPredictResponse{
		Code:      0,
		Message:   "Gas kick prediction completed",
		Alert:     alert,
		RawResult: rawResult,
	})
}

func HandleGasKickHistory(c *gin.Context) {
	wellName := c.Query("well_name")
	limitStr := c.DefaultQuery("limit", "50")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}

	gasKickHistory.mu.RLock()
	defer gasKickHistory.mu.RUnlock()

	var records []gasKickRecord
	if wellName != "" {
		for _, r := range gasKickHistory.results {
			if r.WellName == wellName {
				records = append(records, r)
			}
		}
	} else {
		records = make([]gasKickRecord, len(gasKickHistory.results))
		copy(records, gasKickHistory.results)
	}

	if len(records) > limit {
		records = records[len(records)-limit:]
	}

	c.JSON(http.StatusOK, gin.H{
		"code":   0,
		"count":  len(records),
		"results": records,
	})
}
