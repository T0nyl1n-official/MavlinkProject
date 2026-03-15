package MavlinkRoute

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	Mavlink "MavlinkProject/Server/backend/Handler/Mavlink"
)

var v2HandlerPoolMux sync.Mutex

func getHandlerFromContext(c *gin.Context) *Mavlink.MAVLinkHandlerV1 {
	handlerID := c.GetString("handler_id")
	if handlerID == "" {
		handlerID = c.Query("handler_id")
	}
	if handlerID == "" {
		return nil
	}
	return Mavlink.GetHandlerV1(handlerID)
}

func SetupMavlinkV2Routes(router *gin.Engine) {
	v2Group := router.Group("/mavlink/v2")
	{
		v2Group.POST("/takeoff", func(c *gin.Context) {
			handler := getHandlerFromContext(c)
			if handler == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "handler not found",
				})
				return
			}

			var req Mavlink.TakeoffRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			handlerV2 := Mavlink.MavlinkHandlerV2{}
			handlerV2.New(handler)
			resp := handlerV2.Takeoff(req)
			c.JSON(http.StatusOK, resp)
		})

		v2Group.POST("/land", func(c *gin.Context) {
			handler := getHandlerFromContext(c)
			if handler == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "handler not found",
				})
				return
			}

			var req Mavlink.LandRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			handlerV2 := Mavlink.MavlinkHandlerV2{}
			handlerV2.New(handler)
			resp := handlerV2.Land(req)
			c.JSON(http.StatusOK, resp)
		})

		v2Group.POST("/move", func(c *gin.Context) {
			handler := getHandlerFromContext(c)
			if handler == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "handler not found",
				})
				return
			}

			var req Mavlink.MoveRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			handlerV2 := Mavlink.MavlinkHandlerV2{}
			handlerV2.New(handler)
			resp := handlerV2.Move(req)
			c.JSON(http.StatusOK, resp)
		})

		v2Group.POST("/return", func(c *gin.Context) {
			handler := getHandlerFromContext(c)
			if handler == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "handler not found",
				})
				return
			}

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
			handler := getHandlerFromContext(c)
			if handler == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "handler not found",
				})
				return
			}

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
			handler := getHandlerFromContext(c)
			if handler == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "handler not found",
				})
				return
			}

			handlerV2 := Mavlink.MavlinkHandlerV2{}
			handlerV2.New(handler)
			resp := handlerV2.GetStatus()
			c.JSON(http.StatusOK, resp)
		})

		v2Group.GET("/position", func(c *gin.Context) {
			handler := getHandlerFromContext(c)
			if handler == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "handler not found",
				})
				return
			}

			position := handler.GetDronePosition()
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    position,
			})
		})

		v2Group.GET("/battery", func(c *gin.Context) {
			handler := getHandlerFromContext(c)
			if handler == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "handler not found",
				})
				return
			}

			battery := handler.GetDroneBattery()
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"data":    battery,
			})
		})

		v2Group.POST("/ground-station", func(c *gin.Context) {
			handler := getHandlerFromContext(c)
			if handler == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "handler not found",
				})
				return
			}

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

			handlerV2 := Mavlink.MavlinkHandlerV2{}
			handlerV2.New(handler)
			handlerV2.SetGroundStation(req.Name, req.ID, req.Latitude, req.Longitude, req.Altitude)
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "地面站信息已设置",
			})
		})

		v2Group.POST("/sensor-alert", func(c *gin.Context) {
			handler := getHandlerFromContext(c)
			if handler == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "handler not found",
				})
				return
			}

			var req Mavlink.SensorAlertRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			handlerV2 := Mavlink.MavlinkHandlerV2{}
			handlerV2.New(handler)
			resp := handlerV2.RespondToSensorAlert(req)
			c.JSON(http.StatusOK, resp)
		})

		v2Group.POST("/return-charge", func(c *gin.Context) {
			handler := getHandlerFromContext(c)
			if handler == nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   "handler not found",
				})
				return
			}

			var req Mavlink.ReturnToChargeRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{
					"success": false,
					"error":   err.Error(),
				})
				return
			}

			handlerV2 := Mavlink.MavlinkHandlerV2{}
			handlerV2.New(handler)
			resp := handlerV2.ReturnToCharge(req)
			c.JSON(http.StatusOK, resp)
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
	_ = Mavlink.NewMAVLinkHandlerV1(config)

	SetupMavlinkV2Routes(router)
}
