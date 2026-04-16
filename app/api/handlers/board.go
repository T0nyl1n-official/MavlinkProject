package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"MavlinkProject_Board/app/services/backend"
	"MavlinkProject_Board/app/services/task"
)

func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Unix(),
	})
}

func HandleBoardMessage(c *gin.Context) {
	var req struct {
		MessageID   string      `json:"message_id"`
		MessageTime int64       `json:"message_time"`
		Message     MessageData `json:"message"`
		FromID      string      `json:"from_id"`
		FromType    string      `json:"from_type"`
		ToID        string      `json:"to_id"`
		ToType      string      `json:"to_type"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	// 处理消息
	if err := backend.SendBoardMessage(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "Failed to send message",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Message received",
	})
}

func GetBoardStatus(c *gin.Context) {
	status := backend.GetBoardStatus()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": status,
	})
}

func CreateTaskChain(c *gin.Context) {
	var req struct {
		Tasks []struct {
			Command string                 `json:"command"`
			Data    map[string]interface{} `json:"data"`
			Timeout int                    `json:"timeout"`
		}
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request",
			"error":   err.Error(),
		})
		return
	}

	chain, err := task.CreateTaskChain(req.Tasks)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "Failed to create task chain",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":    0,
		"chain_id": chain.ChainID,
		"status":   chain.Status,
	})
}

func ListTaskChains(c *gin.Context) {
	chains := task.ListTaskChains()
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": chains,
	})
}

func GetTaskChain(c *gin.Context) {
	chainID := c.Param("id")
	chain, err := task.GetTaskChain(chainID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    1,
			"message": "Task chain not found",
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": chain,
	})
}

type MessageData struct {
	MessageType string                 `json:"message_type"`
	Attribute   string                 `json:"attribute"`
	Connection  string                 `json:"connection"`
	Command     string                 `json:"command"`
	Data        map[string]interface{} `json:"data"`
}
