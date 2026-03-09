package routes

import (
	"time"
	"net/http"

	"github.com/gin-gonic/gin"

	Mavlink "MavlinkProject/Server/backend/Handler/Mavlink"
)

func SetupMavlinkV1Routes(router *gin.Engine, handler *Mavlink.MAVLinkHandlerV1) {
	v1Group := router.Group("/mavlink/v1")
	{
		// ==================== 状态查询接口 ====================
		v1Group.GET("/status/handler", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success":            true,
				"handler_id":         handler.GetHandlerID(),
				"connection_status":  handler.GetConnectionStatus(),
				"config":             handler.GetConfig(),
			})
		})

		v1Group.GET("/status/drone", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"success":     true,
				"drone_status": handler.GetDroneStatus(),
				"position":    handler.GetDronePosition(),
				"attitude":    handler.GetDroneAttitude(),
				"battery":     handler.GetDroneBattery(),
			})
		})

		v1Group.GET("/status/ground-station", func(c *gin.Context) {
			groundStation := handler.GetGroundStation()
			c.JSON(http.StatusOK, gin.H{
				"success":        true,
				"ground_station": groundStation,
			})
		})

		// ==================== 连接管理接口 ====================
		v1Group.POST("/connection/start", func(c *gin.Context) {
			if err := handler.Start(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "MAVLink连接已启动",
			})
		})

		v1Group.POST("/connection/stop", func(c *gin.Context) {
			if err := handler.Stop(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "MAVLink连接已停止",
			})
		})

		v1Group.POST("/connection/restart", func(c *gin.Context) {
			if err := handler.Restart(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "MAVLink连接已重启",
			})
		})

		// ==================== 无人机控制接口 ====================
		v1Group.POST("/control/takeoff", func(c *gin.Context) {
			var req struct {
				Altitude float32 `json:"altitude" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			if err := handler.SendTakeoff(req.Altitude); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "起飞指令已发送",
			})
		})

		v1Group.POST("/control/land", func(c *gin.Context) {
			var req struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
				Altitude  float64 `json:"altitude"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			if err := handler.SendLand(req.Latitude, req.Longitude, req.Altitude); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "降落指令已发送",
			})
		})

		v1Group.POST("/control/move", func(c *gin.Context) {
			var req struct {
				Latitude  float64 `json:"latitude" binding:"required"`
				Longitude float64 `json:"longitude" binding:"required"`
				Altitude  float64 `json:"altitude" binding:"required"`
				Speed     float32 `json:"speed"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			if err := handler.SendMoveToPosition(req.Latitude, req.Longitude, req.Altitude, req.Speed); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "移动指令已发送",
			})
		})

		v1Group.POST("/control/mode", func(c *gin.Context) {
			var req struct {
				Mode Mavlink.FlightMode `json:"mode" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			if err := handler.SetFlightMode(req.Mode); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "飞行模式已设置",
			})
		})

		v1Group.POST("/control/heartbeat", func(c *gin.Context) {
			if err := handler.SendHeartbeat(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "心跳包已发送",
			})
		})

		// ==================== 地面站管理接口 ====================
		v1Group.POST("/ground-station/set", func(c *gin.Context) {
			var req struct {
				Name      string  `json:"name" binding:"required"`
				ID        string  `json:"id" binding:"required"`
				Latitude  float64 `json:"latitude" binding:"required"`
				Longitude float64 `json:"longitude" binding:"required"`
				Altitude  float64 `json:"altitude" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			handler.SetGroundStation(req.Name, req.ID, req.Latitude, req.Longitude, req.Altitude)

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "地面站信息已设置",
			})
		})

		// ==================== 连接配置调整接口 ====================
		v1Group.PUT("/config/connection", func(c *gin.Context) {
			var req struct {
				Type   Mavlink.ConnectionType `json:"type" binding:"required"`
				Params map[string]interface{} `json:"params"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			if err := handler.UpdateConnectionType(req.Type, req.Params); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "连接类型已更新",
			})
		})

		v1Group.PUT("/config/system-id", func(c *gin.Context) {
			var req struct {
				SystemID int `json:"system_id" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			if err := handler.UpdateSystemID(req.SystemID); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "系统ID已更新",
			})
		})

		v1Group.PUT("/config/heartbeat-rate", func(c *gin.Context) {
			var req struct {
				RateSeconds int `json:"rate_seconds" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			if err := handler.UpdateHeartbeatRate(time.Duration(req.RateSeconds) * time.Second); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "心跳频率已更新",
			})
		})

		// ==================== 消息流接口 ====================
		v1Group.POST("/stream/request", func(c *gin.Context) {
			var req struct {
				MessageID int `json:"message_id" binding:"required"`
				Rate      int `json:"rate" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			if err := handler.RequestMessageStream(req.MessageID, req.Rate); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "消息流请求已发送",
			})
		})
	}
}

// 简化版本的路由设置函数
func SetupMavlinkRoutesV1(router *gin.Engine) {
	// 创建默认的 MAVLink v1 handler
	handler := Mavlink.NewMAVLinkHandlerV1(Mavlink.MAVLinkConfigV1{
		ConnectionType: Mavlink.ConnectionUDP,
		UDPAddr:        "0.0.0.0",
		UDPPort:        14550,
		SystemID:       255,
		ComponentID:    1,
		ProtocolVersion: Mavlink.ProtocolVersionV2,
		HeartbeatRate:   1 * time.Second,
	})
	
	SetupMavlinkV1Routes(router, handler)
}