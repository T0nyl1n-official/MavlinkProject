package Routes

import (
	"github.com/gin-gonic/gin"

	SensorHandler "MavlinkProject/Server/backend/Handler/Sensor"
)

func SetupSensorRoutes(r *gin.Engine) {
	sensorGroup := r.Group("/api/sensor")
	{
		sensorGroup.POST("/message", SensorHandler.ReceiveSensorMessage)
		sensorGroup.GET("/status", SensorHandler.GetSensorStatus)
	}
}
