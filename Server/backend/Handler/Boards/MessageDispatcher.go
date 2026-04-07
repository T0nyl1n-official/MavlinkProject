package boardHandler

import (
	"fmt"
	"log"

	sensorHandler "MavlinkProject/Server/backend/Handler/Sensor"
	Board "MavlinkProject/Server/backend/Shared/Boards"
)

type MessageType string

const (
	MessageType_Board MessageType = "board"
)

type MessageHandler interface {
	CanHandle(msg *Board.BoardMessage) bool
	Handle(msg *Board.BoardMessage) error
	GetHandlerType() string
	GetName() string
}

type MessageDispatcher struct {
	handlers      []MessageHandler
	aiAgent       *sensorHandler.AIAgentHandler
	sensorHandler *sensorHandler.SensorAlertHandler
	boardHandler  *BoardHandler
}

func NewMessageDispatcher() *MessageDispatcher {
	d := &MessageDispatcher{
		handlers: make([]MessageHandler, 0),
	}

	d.aiAgent = sensorHandler.NewAIAgentHandler()
	d.sensorHandler = sensorHandler.NewSensorAlertHandler()
	d.boardHandler = NewBoardHandler()

	d.RegisterHandler(d.boardHandler)
	d.RegisterHandler(d.sensorHandler)
	d.RegisterHandler(d.aiAgent)

	return d
}

func (d *MessageDispatcher) RegisterHandler(handler MessageHandler) {
	d.handlers = append(d.handlers, handler)
}

func (d *MessageDispatcher) Dispatch(msg *Board.BoardMessage) error {
	log.Printf("[Dispatcher] Received message: Type=%s, From=%s, Command=%s",
		msg.Message.Attribute, msg.FromID, msg.Message.Command)

	for _, handler := range d.handlers {
		if handler.CanHandle(msg) {
			log.Printf("[Dispatcher] Routing to handler: %s (%s)", handler.GetName(), handler.GetHandlerType())
			return handler.Handle(msg)
		}
	}

	return fmt.Errorf("[Dispatcher] No handler found for message: %+v", msg)
}

func (d *MessageDispatcher) GetAIAgentHandler() *sensorHandler.AIAgentHandler {
	return d.aiAgent
}

func (d *MessageDispatcher) GetSensorAlertHandler() *sensorHandler.SensorAlertHandler {
	return d.sensorHandler
}

func (d *MessageDispatcher) GetBoardHandler() *BoardHandler {
	return d.boardHandler
}
