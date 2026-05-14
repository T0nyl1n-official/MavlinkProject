package AI

import (
	"github.com/gin-gonic/gin"

	AIHandler "MavlinkProject/Server/backend/Handler/AI"
	Models "MavlinkProject/Models"
)

func SetupAIRoutes(r *gin.Engine) {
	aiGroup := r.Group("/api/ai")
	{
		aiGroup.POST("/analyze/sensor", AIHandler.HandleSensorAnalysis)
		aiGroup.POST("/analyze/drone", AIHandler.HandleDroneImageAnalysis)
		aiGroup.POST("/drone/photo", AIHandler.HandleDronePhotoUpload)
		aiGroup.GET("/drone/photo/generated/:filename", AIHandler.HandleDownloadGeneratedImage)
		aiGroup.GET("/alerts/ws", AIHandler.HandleAlertWebSocket)
		aiGroup.GET("/alerts/sse", AIHandler.HandleAlertSSE)
		aiGroup.GET("/alerts/history", AIHandler.HandleAlertHistory)
		aiGroup.GET("/model/status", HandleModelStatus)
	}
}

func HandleModelStatus(c *gin.Context) {
	client := Models.GetModelClient()
	status := client.HealthCheck()
	c.JSON(200, gin.H{
		"code":   0,
		"status": status,
	})
}
