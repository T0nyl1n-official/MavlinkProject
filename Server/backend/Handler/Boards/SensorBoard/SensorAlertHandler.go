package sensorHandler

import (
	"fmt"
	"log"

	Board "MavlinkProject/Server/backend/Shared/Boards"
)

type SensorAlertHandler struct {
	name        string
	autoExecute bool
}

func NewSensorAlertHandler() *SensorAlertHandler {
	return &SensorAlertHandler{
		name:        "SensorAlertHandler",
		autoExecute: true,
	}
}

func (h *SensorAlertHandler) GetHandlerType() string {
	return "sensor"
}

func (h *SensorAlertHandler) GetName() string {
	return h.name
}

func (h *SensorAlertHandler) CanHandle(msg *Board.BoardMessage) bool {
	if msg.FromType == "sensor" || msg.FromType == "esp32" || msg.FromType == "alarm" {
		return true
	}
	if msg.Message.Attribute == Board.MessageAttribute_Warning ||
		msg.Message.Attribute == Board.MessageAttribute_Mission {
		if msg.Message.Command == "Alert" || msg.Message.Command == "SensorAlert" {
			return true
		}
	}
	if msg.Message.Command == "SensorAlert" || msg.Message.Command == "Alert" {
		return true
	}
	return false
}

func (h *SensorAlertHandler) Handle(msg *Board.BoardMessage) error {
	log.Printf("[SensorAlertHandler] Processing sensor alert from %s", msg.FromID)

	data := msg.Message.Data
	if data == nil {
		return fmt.Errorf("[SensorAlertHandler] No data in sensor alert message")
	}

	sensorID, _ := data["sensor_id"].(string)
	if sensorID == "" {
		sensorID = msg.FromID
	}

	lat, _ := data["latitude"].(float64)
	lon, _ := data["longitude"].(float64)
	radius, _ := data["radius"].(float64)
	altitude, _ := data["altitude"].(float64)
	photoCount, _ := data["photo_count"].(int)

	req := SensorAlertReq{
		SensorID:   sensorID,
		Latitude:   lat,
		Longitude:  lon,
		Radius:     radius,
		PhotoCount: photoCount,
		Altitude:   altitude,
	}

	log.Printf("[SensorAlertHandler] Sensor alert: ID=%s, Lat=%.6f, Lon=%.6f, Radius=%.1f, Altitude=%.1f",
		sensorID, lat, lon, radius, altitude)

	if h.autoExecute {
		log.Printf("[SensorAlertHandler] Auto-executing: generating chain to central")
		return GenerateChainAndSendToCentral(req)
	}

	log.Printf("[SensorAlertHandler] Auto-execute disabled, skipping chain generation")
	return nil
}

func (h *SensorAlertHandler) SetAutoExecute(auto bool) {
	h.autoExecute = auto
}

func (h *SensorAlertHandler) IsAutoExecute() bool {
	return h.autoExecute
}

type AIAgentHandler struct {
	name           string
	enabled        bool
	analysisDepth  string
	RequireConfirm bool
}

func NewAIAgentHandler() *AIAgentHandler {
	return &AIAgentHandler{
		name:           "AIAgentHandler",
		enabled:        false,
		analysisDepth:  "full",
		RequireConfirm: false,
	}
}

func (h *AIAgentHandler) GetHandlerType() string {
	return "ai"
}

func (h *AIAgentHandler) GetName() string {
	return h.name
}

func (h *AIAgentHandler) CanHandle(msg *Board.BoardMessage) bool {
	return h.enabled
}

func (h *AIAgentHandler) Handle(msg *Board.BoardMessage) error {
	if !h.enabled {
		return fmt.Errorf("[AIAgentHandler] AI Agent is not enabled")
	}

	log.Printf("[AIAgentHandler] AI analyzing message: %s", msg.MessageID)
	log.Printf("[AIAgentHandler] Analysis depth: %s", h.analysisDepth)
	log.Printf("[AIAgentHandler] Require confirm before action: %v", h.RequireConfirm)

	return fmt.Errorf("[AIAgentHandler] AI Agent handling not yet implemented")
}

func (h *AIAgentHandler) Enable() {
	h.enabled = true
	log.Printf("[AIAgentHandler] AI Agent enabled")
}

func (h *AIAgentHandler) Disable() {
	h.enabled = false
	log.Printf("[AIAgentHandler] AI Agent disabled")
}

func (h *AIAgentHandler) IsEnabled() bool {
	return h.enabled
}

func (h *AIAgentHandler) SetAnalysisDepth(depth string) {
	h.analysisDepth = depth
}

func (h *AIAgentHandler) SetRequireConfirm(require bool) {
	h.RequireConfirm = require
}
