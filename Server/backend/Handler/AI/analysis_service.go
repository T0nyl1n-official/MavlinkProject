package AI

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	Models "MavlinkProject/Models"
	BoardSensorHandler "MavlinkProject/Server/backend/Handler/Boards/SensorBoard"
	FRPHelper "MavlinkProject/Server/backend/Utils/FRPHelper"
	WarningHandler "MavlinkProject/Server/backend/Utils/WarningHandle"
)

type AnalysisService struct {
	client *Models.ModelClient
	hub    *AlertHub
	mu     sync.RWMutex
}

var (
	globalService *AnalysisService
	serviceOnce   sync.Once
)

func GetAnalysisService() *AnalysisService {
	serviceOnce.Do(func() {
		globalService = &AnalysisService{
			client: Models.GetModelClient(),
			hub:    GetAlertHub(),
		}
	})
	return globalService
}

func (s *AnalysisService) ProcessSensorData(sensorID string, sensorType string, values []Models.TimeSeriesPoint, lat, lon float64) (*Models.AlertJSON, error) {
	startTime := time.Now()

	lstmReq := Models.LSTMRequest{
		SensorID:     sensorID,
		DataType:     sensorType,
		TimeSeries:   values,
		PredictSteps: 10,
	}

	lstmResp, err := s.client.AnalyzeSensorData(lstmReq)
	if err != nil {
		log.Printf("[AI] LSTM 分析失败: sensor=%s, err=%v", sensorID, err)
		WarningHandler.HandleAgentError("LSTM analysis failed", "analysis-service", sensorID, err.Error())
		return nil, err
	}

	log.Printf("[AI] LSTM 分析完成: sensor=%s, anomaly=%v, score=%.4f, type=%s, 耗时=%v",
		sensorID, lstmResp.IsAnomaly, lstmResp.AnomalyScore, lstmResp.AnomalyType, time.Since(startTime))

	if !lstmResp.IsAnomaly {
		statusAlert := &Models.AlertJSON{
			AlertID:     generateAlertID(),
			AlertType:   "normal",
			Severity:    Models.SeverityInfo,
			Latitude:    lat,
			Longitude:   lon,
			AnomalyType: "none",
			Source:      Models.SourceSensor,
			SensorID:    sensorID,
			Timestamp:   time.Now().Unix(),
			Confidence:  1.0 - lstmResp.AnomalyScore,
			Details: map[string]interface{}{
				"anomaly_score": lstmResp.AnomalyScore,
				"predict_steps": lstmReq.PredictSteps,
			},
		}
		s.hub.Broadcast(statusAlert)
		return statusAlert, nil
	}

	severity := scoreToSeverity(lstmResp.AnomalyScore)
	anomalyType := mapLSTMAnomalyType(lstmResp.AnomalyType, sensorType)

	alert := &Models.AlertJSON{
		AlertID:     generateAlertID(),
		AlertType:   "anomaly",
		Severity:    severity,
		Latitude:    lat,
		Longitude:   lon,
		AnomalyType: anomalyType,
		Source:      Models.SourceSensor,
		SensorID:    sensorID,
		Timestamp:   time.Now().Unix(),
		Confidence:  lstmResp.Confidence,
		Details: map[string]interface{}{
			"anomaly_score": lstmResp.AnomalyScore,
			"prediction":    lstmResp.Prediction,
			"model_version": lstmResp.ModelVersion,
		},
	}

	s.hub.Broadcast(alert)

	log.Printf("[AI] 🚨 传感器异常告警: alert_id=%s, sensor=%s, type=%s, severity=%s, score=%.4f",
		alert.AlertID, sensorID, anomalyType, severity, lstmResp.AnomalyScore)

	if severity == Models.SeverityCritical || severity == Models.SeverityHigh {
		go s.triggerDroneDispatch(sensorID, lat, lon, anomalyType, severity)
	}

	return alert, nil
}

func (s *AnalysisService) ProcessDroneImage(droneID string, imageBase64 string, imageURL string, lat, lon float64) (*Models.AlertJSON, error) {
	startTime := time.Now()

	yoloReq := Models.YOLORequest{
		ImageBase64: imageBase64,
		ImageURL:    imageURL,
		Confidence:  0.5,
		Source:      Models.SourceDrone,
		Metadata: map[string]string{
			"drone_id":  droneID,
			"latitude":  fmt.Sprintf("%.6f", lat),
			"longitude": fmt.Sprintf("%.6f", lon),
		},
	}

	yoloResp, err := s.client.AnalyzeImage(yoloReq)
	if err != nil {
		log.Printf("[AI] YOLO 分析失败: drone=%s, err=%v", droneID, err)
		WarningHandler.HandleAgentError("YOLO analysis failed", "analysis-service", droneID, err.Error())
		return nil, err
	}

	log.Printf("[AI] YOLO 分析完成: drone=%s, anomaly=%v, type=%s, severity=%s, detections=%d, 耗时=%v",
		droneID, yoloResp.HasAnomaly, yoloResp.AnomalyType, yoloResp.Severity, len(yoloResp.Detections), time.Since(startTime))

	if !yoloResp.HasAnomaly {
		statusAlert := &Models.AlertJSON{
			AlertID:     generateAlertID(),
			AlertType:   "normal",
			Severity:    Models.SeverityInfo,
			Latitude:    lat,
			Longitude:   lon,
			AnomalyType: "none",
			Source:      Models.SourceDrone,
			DroneID:     droneID,
			Timestamp:   time.Now().Unix(),
			Confidence:  1.0,
			Details: map[string]interface{}{
				"detections": len(yoloResp.Detections),
			},
		}
		s.hub.Broadcast(statusAlert)
		return statusAlert, nil
	}

	alert := &Models.AlertJSON{
		AlertID:     generateAlertID(),
		AlertType:   "anomaly",
		Severity:    yoloResp.Severity,
		Latitude:    lat,
		Longitude:   lon,
		AnomalyType: yoloResp.AnomalyType,
		Source:      Models.SourceDrone,
		DroneID:     droneID,
		Timestamp:   time.Now().Unix(),
		Confidence:  yoloResp.Detections[0].Confidence,
		Details: map[string]interface{}{
			"detections":    yoloResp.Detections,
			"model_version": yoloResp.ModelVersion,
		},
	}

	s.hub.Broadcast(alert)

	log.Printf("[AI] 🚨 无人机图像告警: alert_id=%s, drone=%s, type=%s, severity=%s",
		alert.AlertID, droneID, yoloResp.AnomalyType, yoloResp.Severity)

	return alert, nil
}

func (s *AnalysisService) triggerDroneDispatch(sensorID string, lat, lon float64, anomalyType, severity string) {
	log.Printf("[AI] 触发无人机调度: sensor=%s, type=%s, severity=%s", sensorID, anomalyType, severity)

	client := FRPHelper.GetCentralClient()
	if client != nil {
		drones, err := client.GetAvailableDrones()
		if err != nil || len(drones) == 0 {
			log.Printf("[AI] 无可用无人机: %v", err)
			return
		}

		altitude := 30.0
		radius := 50.0
		switch anomalyType {
		case Models.AnomalyFire, Models.AnomalySmoke:
			altitude = 25.0
			radius = 50.0
		case Models.AnomalyPerson:
			altitude = 40.0
			radius = 100.0
		case Models.AnomalyGas:
			altitude = 20.0
			radius = 30.0
		}

		frpReq := BoardSensorHandler.SensorAlertReq{
			SensorID:   sensorID,
			Latitude:   lat,
			Longitude:  lon,
			Radius:     radius,
			PhotoCount: 4,
			Altitude:   altitude,
		}

		if err := BoardSensorHandler.GenerateChainAndSendToCentral(frpReq); err != nil {
			log.Printf("[AI] 无人机调度失败: %v", err)
			WarningHandler.HandleDroneError("drone dispatch failed", drones[0].BoardID, "", err.Error())
		} else {
			log.Printf("[AI] ✅ 无人机调度成功: drone=%s", drones[0].BoardID)
		}
	}
}

func scoreToSeverity(score float64) string {
	switch {
	case score >= 0.9:
		return Models.SeverityCritical
	case score >= 0.7:
		return Models.SeverityHigh
	case score >= 0.4:
		return Models.SeverityMedium
	case score >= 0.2:
		return Models.SeverityLow
	default:
		return Models.SeverityInfo
	}
}

func mapLSTMAnomalyType(modelType string, sensorType string) string {
	if modelType != "" && modelType != "unknown" {
		return modelType
	}

	switch sensorType {
	case "temperature", "temp":
		return Models.AnomalyTemp
	case "humidity":
		return Models.AnomalyHumidity
	case "pressure":
		return Models.AnomalyPressure
	case "gas", "co", "ch4", "co2":
		return Models.AnomalyGas
	default:
		return Models.AnomalyUnknown
	}
}

func generateAlertID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("alert_%s_%d", hex.EncodeToString(b), time.Now().UnixMilli())
}

func AlertToJSON(alert *Models.AlertJSON) string {
	data, err := json.Marshal(alert)
	if err != nil {
		log.Printf("[AI] AlertJSON 序列化失败: %v", err)
		return "{}"
	}
	return string(data)
}

func (s *AnalysisService) ProcessDronePhoto(droneID string, imagePath string, lat, lon float64) (*Models.AlertJSON, *Models.ThermalDetectResponse, error) {
	startTime := time.Now()

	thermalResp, err := s.client.ThermalDetect(imagePath)
	if err != nil {
		log.Printf("[AI] YOLOv8 热源检测失败: drone=%s, err=%v", droneID, err)
		WarningHandler.HandleAgentError("YOLOv8 thermal detect failed", "analysis-service", droneID, err.Error())
		return nil, nil, err
	}

	log.Printf("[AI] YOLOv8 热源检测完成: drone=%s, success=%v, detections=%d, elapsed=%.2fms, 耗时=%v",
		droneID, thermalResp.Success, len(thermalResp.Detections), thermalResp.ElapsedMs, time.Since(startTime))

	hasAnomaly := false
	maxSeverity := Models.SeverityInfo
	highestConfidence := 0.0
	var topDetection *Models.ThermalDetection

	for i := range thermalResp.Detections {
		det := &thermalResp.Detections[i]
		detSeverity := thermalLevelToSeverity(det.Temperature.Level)
		if detSeverity != Models.SeverityInfo {
			hasAnomaly = true
		}
		if severityRank(detSeverity) > severityRank(maxSeverity) ||
			(detSeverity == maxSeverity && det.Confidence > highestConfidence) {
			maxSeverity = detSeverity
			highestConfidence = det.Confidence
			topDetection = det
		}
	}

	if !hasAnomaly || thermalResp.Detections == nil || len(thermalResp.Detections) == 0 {
		statusAlert := &Models.AlertJSON{
			AlertID:    generateAlertID(),
			AlertType:  "normal",
			Severity:   Models.SeverityInfo,
			Latitude:   lat,
			Longitude:  lon,
			AnomalyType: "none",
			Source:     Models.SourceDrone,
			DroneID:    droneID,
			Timestamp:  time.Now().Unix(),
			Confidence: 1.0,
			Details: map[string]interface{}{
				"detection_count": len(thermalResp.Detections),
				"elapsed_ms":     thermalResp.ElapsedMs,
				"image_size":      fmt.Sprintf("%dx%d", thermalResp.Image.Width, thermalResp.Image.Height),
			},
		}
		s.hub.Broadcast(statusAlert)
		return statusAlert, thermalResp, nil
	}

	alert := &Models.AlertJSON{
		AlertID:    generateAlertID(),
		AlertType:  "anomaly",
		Severity:   maxSeverity,
		Latitude:   lat,
		Longitude:  lon,
		AnomalyType: Models.AnomalyThermal,
		Source:     Models.SourceDrone,
		DroneID:    droneID,
		Timestamp:  time.Now().Unix(),
		Confidence: highestConfidence,
		Details: map[string]interface{}{
			"thermal_detections": thermalResp.Detections,
			"top_detection":      topDetection,
			"elapsed_ms":         thermalResp.ElapsedMs,
			"image_size":          fmt.Sprintf("%dx%d", thermalResp.Image.Width, thermalResp.Image.Height),
		},
	}

	s.hub.Broadcast(alert)

	log.Printf("[AI] 🌡️ 无人机热源异常告警: alert_id=%s, drone=%s, severity=%s, detections=%d, top_level=%s",
		alert.AlertID, droneID, maxSeverity, len(thermalResp.Detections), topDetection.Temperature.Level)

	if maxSeverity == Models.SeverityCritical || maxSeverity == Models.SeverityHigh {
		log.Printf("[AI] 热源异常严重等级 %s，建议关注区域坐标 (%.6f, %.6f)", maxSeverity, lat, lon)
	}

	return alert, thermalResp, nil
}

func thermalLevelToSeverity(level string) string {
	switch level {
	case Models.TempLevelHigh1:
		return Models.SeverityCritical
	case Models.TempLevelHigh2:
		return Models.SeverityHigh
	case Models.TempLevelNormal:
		return Models.SeverityInfo
	case Models.TempLevelLow2:
		return Models.SeverityInfo
	case Models.TempLevelLow1:
		return Models.SeverityInfo
	default:
		return Models.SeverityInfo
	}
}

func severityRank(severity string) int {
	switch severity {
	case Models.SeverityCritical:
		return 5
	case Models.SeverityHigh:
		return 4
	case Models.SeverityMedium:
		return 3
	case Models.SeverityLow:
		return 2
	case Models.SeverityInfo:
		return 1
	default:
		return 0
	}
}

func (s *AnalysisService) ProcessGasKick(wellName string, sensorID string, values []Models.TimeSeriesPoint, lat, lon float64, alertWindow int, alertThreshold int) (*Models.AlertJSON, *Models.GasKickResponse, error) {
	gasReq := Models.GasKickRequest{
		WellName:   wellName,
		SensorID:   sensorID,
		TimeSeries: values,
		Latitude:   lat,
		Longitude:  lon,
	}

	gasResp, err := s.client.PredictGasKick(gasReq)
	if err != nil {
		log.Printf("[AI] Gas Kick 预测失败: well=%s, err=%v", wellName, err)
		WarningHandler.HandleAgentError("Gas kick prediction failed", "analysis-service", wellName, err.Error())
		return nil, nil, err
	}

	log.Printf("[AI] Gas Kick 预测完成: well=%s, total=%d, kicks=%d, ratio=%.2f%%, elapsed=%.2fms",
		wellName, gasResp.Summary.TotalPoints, gasResp.Summary.GasKickCount,
		gasResp.Summary.GasKickRatio*100, gasResp.ElapsedMs)

	if gasResp.Summary.GasKickCount == 0 || gasResp.Summary.GasKickRatio == 0 {
		swTriggered, swKicks, swWindow := checkSlidingWindowAlert(gasResp.Predictions, alertWindow, alertThreshold)

		statusAlert := &Models.AlertJSON{
			AlertID:    generateAlertID(),
			AlertType:  "normal",
			Severity:   Models.SeverityInfo,
			Latitude:   lat,
			Longitude:  lon,
			AnomalyType: "none",
			Source:     Models.SourceSensor,
			SensorID:   sensorID,
			Timestamp:  time.Now().Unix(),
			Confidence: 1.0 - gasResp.Summary.GasKickRatio,
			Details: map[string]interface{}{
				"well_name":          wellName,
				"total_points":       gasResp.Summary.TotalPoints,
				"gas_kick_count":     gasResp.Summary.GasKickCount,
				"gas_kick_ratio":     gasResp.Summary.GasKickRatio,
				"elapsed_ms":         gasResp.ElapsedMs,
				"sliding_window":     map[string]interface{}{"window": swWindow, "threshold": alertThreshold, "kicks_in_window": swKicks, "triggered": swTriggered},
				"predictions":        gasResp.Predictions,
			},
		}
		s.hub.Broadcast(statusAlert)
		return statusAlert, gasResp, nil
	}

	alertSeverity := gasKickSeverity(gasResp)

	swTriggered, swKicks, swWindow := checkSlidingWindowAlert(gasResp.Predictions, alertWindow, alertThreshold)
	if swTriggered && alertSeverity == Models.SeverityInfo {
		alertSeverity = Models.SeverityLow
	}
	if swKicks >= alertThreshold + 1 {
		alertSeverity = Models.SeverityMedium
	}
	if swKicks >= swWindow {
		alertSeverity = Models.SeverityHigh
	}

	alert := &Models.AlertJSON{
		AlertID:    generateAlertID(),
		AlertType:  "anomaly",
		Severity:   alertSeverity,
		Latitude:   lat,
		Longitude:  lon,
		AnomalyType: Models.AnomalyGasKick,
		Source:     Models.SourceSensor,
		SensorID:   sensorID,
		Timestamp:  time.Now().Unix(),
		Confidence: gasResp.Summary.GasKickRatio,
		Details: map[string]interface{}{
			"well_name":           wellName,
			"summary":             gasResp.Summary,
			"predictions_count":   len(gasResp.Predictions),
			"predictions":         gasResp.Predictions,
			"elapsed_ms":          gasResp.ElapsedMs,
			"sliding_window":      map[string]interface{}{"window": swWindow, "threshold": alertThreshold, "kicks_in_window": swKicks, "triggered": swTriggered},
		},
	}

	s.hub.Broadcast(alert)

	log.Printf("[AI] ⚠️ 气侵告警: alert_id=%s, well=%s, severity=%s, kicks=%d/%d (%.1f%%), consecutive_max=%d",
		alert.AlertID, wellName, alertSeverity,
		gasResp.Summary.GasKickCount, gasResp.Summary.TotalPoints,
		gasResp.Summary.GasKickRatio*100, gasResp.Summary.ConsecutiveMax)

	return alert, gasResp, nil
}

func gasKickSeverity(resp *Models.GasKickResponse) string {
	ratio := resp.Summary.GasKickRatio
	switch {
	case ratio >= 0.5:
		return Models.SeverityCritical
	case ratio >= 0.3:
		return Models.SeverityHigh
	case ratio >= 0.15:
		return Models.SeverityMedium
	case ratio >= 0.05:
		return Models.SeverityLow
	default:
		return Models.SeverityInfo
	}
}

func checkSlidingWindowAlert(predictions []Models.GasKickPrediction, windowSize, threshold int) (triggered bool, kickCount int, actualWindow int) {
	if len(predictions) == 0 || windowSize <= 0 || threshold <= 0 {
		return false, 0, windowSize
	}

	actualWindow = windowSize
	if len(predictions) < windowSize {
		actualWindow = len(predictions)
	}

	startIdx := len(predictions) - actualWindow
	kickCount = 0

	for i := startIdx; i < len(predictions); i++ {
		if predictions[i].Predicted == 1 {
			kickCount++
		}
	}

	triggered = kickCount >= threshold
	return triggered, kickCount, actualWindow
}
