package boards

import (
	"log"

	Board "MavlinkProject/Server/backend/Shared/Boards"
)

type BoardHandler struct {
	name string
}

func NewBoardHandler() *BoardHandler {
	return &BoardHandler{
		name: "BoardHandler",
	}
}

func (h *BoardHandler) GetHandlerType() string {
	return "board"
}

func (h *BoardHandler) GetName() string {
	return h.name
}

func (h *BoardHandler) CanHandle(msg *Board.BoardMessage) bool {
	if msg.FromType == "board" || msg.FromType == "drone" || msg.FromType == "fc" {
		return true
	}
	if msg.Message.Attribute == Board.MessageAttribute_Status ||
		msg.Message.Attribute == Board.MessageAttribute_Control ||
		msg.Message.Attribute == Board.MessageAttribute_Command {
		return true
	}
	if msg.Message.Command == "Heartbeat" || msg.Message.Command == "Status" {
		return true
	}
	return false
}

func (h *BoardHandler) Handle(msg *Board.BoardMessage) error {
	log.Printf("[BoardHandler] Handling board message: From=%s, Command=%s, Attribute=%s",
		msg.FromID, msg.Message.Command, msg.Message.Attribute)

	switch msg.Message.Command {
	case "Heartbeat":
		return h.handleHeartbeat(msg)
	case "Status":
		return h.handleStatus(msg)
	case "TakeOff", "Land", "GoTo", "SetMode", "Arm", "Disarm":
		return h.handleCommand(msg)
	default:
		return h.handleGenericBoardMessage(msg)
	}
}

func (h *BoardHandler) handleHeartbeat(msg *Board.BoardMessage) error {
	log.Printf("[BoardHandler] Heartbeat from %s", msg.FromID)
	return nil
}

func (h *BoardHandler) handleStatus(msg *Board.BoardMessage) error {
	log.Printf("[BoardHandler] Status update from %s: %+v", msg.FromID, msg.Message.Data)
	return nil
}

func (h *BoardHandler) handleCommand(msg *Board.BoardMessage) error {
	log.Printf("[BoardHandler] Command '%s' from %s", msg.Message.Command, msg.FromID)
	return nil
}

func (h *BoardHandler) handleGenericBoardMessage(msg *Board.BoardMessage) error {
	log.Printf("[BoardHandler] Generic board message processed: %s", msg.MessageID)
	return nil
}
