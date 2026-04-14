package Sensor

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"

	BoardsHandler "MavlinkProject/Server/backend/Handler/Boards"
	Board "MavlinkProject/Server/backend/Shared/Boards"
)

func ReceiveSensorMessage(c *gin.Context) {
	var msg Board.BoardMessage

	if err := c.ShouldBindJSON(&msg); err != nil {
		c.JSON(400, gin.H{
			"code":    1,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
		return
	}

	if msg.MessageID == "" {
		msg.MessageID = generateMessageID()
	}
	msg.MessageTime = time.Now()

	if msg.FromID == "" {
		msg.FromID = c.ClientIP()
	}
	if msg.FromType == "" {
		msg.FromType = "ESP32"
	}

	boardMgr := BoardsHandler.GetBoardManager()
	if boardMgr == nil {
		c.JSON(500, gin.H{
			"code":    1,
			"message": "Board manager not available",
		})
		return
	}

	select {
	case boardMgr.GetMessageChan() <- &msg:
		log.Printf("[Sensor HTTP] Message received from %s, type: %s", msg.FromID, msg.Message.MessageType)
		c.JSON(200, gin.H{
			"code":       0,
			"message":    "Message received",
			"message_id": msg.MessageID,
		})
	default:
		c.JSON(503, gin.H{
			"code":    1,
			"message": "Message channel full, try again later",
		})
	}
}

func GetSensorStatus(c *gin.Context) {
	boardMgr := BoardsHandler.GetBoardManager()
	if boardMgr == nil {
		c.JSON(500, gin.H{
			"code":    1,
			"message": "Board manager not available",
		})
		return
	}

	boards := boardMgr.GetAllBoards()
	c.JSON(200, gin.H{
		"code":   0,
		"boards": boards,
	})
}

func generateMessageID() string {
	return time.Now().Format("20060102150405.000")
}
