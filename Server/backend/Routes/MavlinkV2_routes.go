package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	Mavlink "MavlinkProject/Server/backend/Handler/Mavlink"
)

func SetupMavlinkV2Routes(router *gin.Engine, handler *Mavlink.MAVLinkHandler) {
	packageHandler := Mavlink.NewPackageHandler(handler)

	v2Group := router.Group("/mavlink/v2/api")
	{
		// 无人机控制接口
		v2Group.POST("/takeoff", func(c *gin.Context) {
			var req Mavlink.TakeoffRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			resp := packageHandler.Takeoff(req)
			c.JSON(http.StatusOK, resp)
		})

		v2Group.POST("/land", func(c *gin.Context) {
			var req Mavlink.LandRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			resp := packageHandler.Land(req)
			c.JSON(http.StatusOK, resp)
		})

		v2Group.POST("/move", func(c *gin.Context) {
			var req Mavlink.MoveRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			resp := packageHandler.Move(req)
			c.JSON(http.StatusOK, resp)
		})

		// 状态查询接口
		v2Group.GET("/status", func(c *gin.Context) {
			resp := packageHandler.GetStatus()
			c.JSON(http.StatusOK, resp)
		})

		// 调度者设置接口
		v2Group.POST("/dispatcher/ai", func(c *gin.Context) {
			packageHandler.SetAIAsDispatcher()
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "调度者已设置为AI",
			})
		})

		v2Group.POST("/dispatcher/user", func(c *gin.Context) {
			var req struct {
				Username string `json:"username" binding:"required"`
				Email    string `json:"email" binding:"required"`
			}

			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			packageHandler.SetUserAsDispatcher(req.Username, req.Email)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "调度者已设置为用户",
			})
		})

		// 地面站设置接口
		v2Group.POST("/ground-station", func(c *gin.Context) {
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

			packageHandler.SetGroundStation(req.Name, req.ID, req.Latitude, req.Longitude, req.Altitude)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "地面站信息已设置",
			})
		})
	}
}

// 简化的路由设置函数（用于快速集成）
func SetupDefaultMavlinkRoutesV2(router *gin.Engine) {
	// 创建默认的 MAVLink handler
	handler := Mavlink.NewMAVLinkHandler(Mavlink.MAVLinkConfig{
		ConnectionType:  Mavlink.ConnectionUDP,
		UDPAddr:         "0.0.0.0",
		UDPPort:         14550,
		SystemID:        255,
		ComponentID:     1,
		ProtocolVersion: Mavlink.ProtocolVersion2,
		HeartbeatRate:   1 * time.Second,
	}, nil)

	SetupMavlinkV2Routes(router, handler)
}
