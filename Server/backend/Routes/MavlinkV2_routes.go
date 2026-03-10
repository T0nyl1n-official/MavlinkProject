package routes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	Mavlink "MavlinkProject/Server/backend/Handler/Mavlink"
)

func SetupMavlinkV2Routes(router *gin.Engine, handler *Mavlink.MAVLinkHandlerV1) {
	packageHandler := Mavlink.NewPackageHandler(handler)

	v2Group := router.Group("/mavlink/v2/api")
	{
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

		v2Group.POST("/return", func(c *gin.Context) {
			err := handler.SendReturnToLaunch()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Return to launch initiated",
			})
		})

		v2Group.POST("/mode", func(c *gin.Context) {
			var req struct {
				Mode string `json:"mode" binding:"required"`
			}
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			err := handler.SetFlightMode(Mavlink.FlightMode(req.Mode))
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "Flight mode set to " + req.Mode,
			})
		})

		v2Group.GET("/status", func(c *gin.Context) {
			resp := packageHandler.GetStatus()
			c.JSON(http.StatusOK, resp)
		})

		v2Group.GET("/position", func(c *gin.Context) {
			position := handler.GetDronePosition()
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    position,
			})
		})

		v2Group.GET("/battery", func(c *gin.Context) {
			battery := handler.GetDroneBattery()
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    battery,
			})
		})

		v2Group.POST("/ground-station", func(c *gin.Context) {
			var req struct {
				Name      string  `json:"name" binding:"required"`
				ID        string  `json:"id"`
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

			packageHandler.SetGroundStation(req.Name, req.ID, req.Latitude, req.Longitude, req.Altitude)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "地面站信息已设置",
			})
		})
	}
}

func SetupDefaultMavlinkRoutesV2(router *gin.Engine) {
	config := Mavlink.MAVLinkConfigV1{
		ConnectionType:  Mavlink.ConnectionUDP,
		UDPAddr:         "0.0.0.0",
		UDPPort:         14550,
		SystemID:        255,
		ComponentID:     1,
		ProtocolVersion: Mavlink.ProtocolVersionV2,
		HeartbeatRate:   1 * time.Second,
	}
	handler := Mavlink.NewMAVLinkHandlerV1(config)

	SetupMavlinkV2Routes(router, handler)
}
