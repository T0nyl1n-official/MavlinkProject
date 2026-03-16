package MavlinkRoute

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	Mavlink "MavlinkProject/Server/backend/Handler/Mavlink"
	JwtMiddleware "MavlinkProject/Server/backend/Middles"
	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
)

type ConnectionConfig struct {
	ConnectionType  string        `json:"connection_type" binding:"required,oneof=serial udp tcp"`
	SerialPort      string        `json:"serial_port,omitempty"`
	SerialBaud      int           `json:"serial_baud,omitempty"`
	UDPAddr         string        `json:"udp_addr,omitempty"`
	UDPPort         int           `json:"udp_port,omitempty"`
	TCPAddr         string        `json:"tcp_addr,omitempty"`
	TCPPort         int           `json:"tcp_port,omitempty"`
	SystemID        int           `json:"system_id" binding:"required,min=1,max=255"`
	ComponentID     int           `json:"component_id" binding:"required,min=1,max=255"`
	ProtocolVersion string        `json:"protocol_version" binding:"required,oneof=1.0 2.0"`
	HeartbeatRate   time.Duration `json:"heartbeat_rate,omitempty"`
}

type MavlinkV1Response struct {
	Success   bool        `json:"success"`
	HandlerID string      `json:"handler_id,omitempty"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
}

func SetupMavlinkV1Routes(router *gin.Engine, jwtManager interface{}, tokenManager interface{}) {
	v1Group := router.Group("/mavlink/v1")
	v1Group.Use(JwtMiddleware.JwtAuthMiddleWareWithRedis(jwtManager.(*jwtUtils.JWTManager), tokenManager.(*Jwt.RedisTokenManager), nil))

	{
		v1Group.POST("/handler/create", createHandler)
		v1Group.DELETE("/handler/:id", deleteHandler)
		v1Group.GET("/handler/:id", getHandlerStatus)

		v1Group.POST("/connection/start", startConnection)
		v1Group.POST("/connection/stop", stopConnection)
		v1Group.POST("/connection/restart", restartConnection)

		v1Group.POST("/drone/takeoff", droneTakeoff)
		v1Group.POST("/drone/land", droneLand)
		v1Group.POST("/drone/move", droneMove)
		v1Group.POST("/drone/return", droneReturn)
		v1Group.POST("/drone/mode", droneMode)

		v1Group.GET("/drone/status", getDroneStatus)
		v1Group.GET("/drone/position", getDronePosition)
		v1Group.GET("/drone/attitude", getDroneAttitude)
		v1Group.GET("/drone/battery", getDroneBattery)

		v1Group.POST("/ground-station/set", setGroundStation)
		v1Group.GET("/ground-station", getGroundStation)

		v1Group.POST("/stream/request", requestStream)
		v1Group.POST("/heartbeat/send", sendHeartbeat)
	}
}

func createHandler(c *gin.Context) {
	var config ConnectionConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	mavlinkConfig := Mavlink.MAVLinkConfigV1{
		ConnectionType:  Mavlink.ConnectionType(config.ConnectionType),
		SerialPort:      config.SerialPort,
		SerialBaud:      config.SerialBaud,
		UDPAddr:         config.UDPAddr,
		UDPPort:         config.UDPPort,
		TCPAddr:         config.TCPAddr,
		TCPPort:         config.TCPPort,
		SystemID:        config.SystemID,
		ComponentID:     config.ComponentID,
		ProtocolVersion: Mavlink.ProtocolVersion(config.ProtocolVersion),
		HeartbeatRate:   config.HeartbeatRate,
	}

	handler := Mavlink.NewMAVLinkHandlerV1(mavlinkConfig)
	if handler == nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success: false,
			Error:   "Failed to create handler",
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handler.GetHandlerID(),
		Message:   "Handler created successfully",
	})
}

func deleteHandler(c *gin.Context) {
	handlerID := c.Param("id")
	handler := Mavlink.GetHandlerV1(handlerID)

	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	err := handler.Delete()
	if err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success: true,
		Message: "Handler deleted successfully",
	})
}

func getHandlerStatus(c *gin.Context) {
	handlerID := c.Param("id")
	handler := Mavlink.GetHandlerV1(handlerID)

	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	info := handler.GetInfo()
	c.JSON(http.StatusOK, MavlinkV1Response{
		Success: true,
		Data:    info,
	})
}

func startConnection(c *gin.Context) {
	var req struct {
		HandlerID string `json:"handler_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	handler := Mavlink.GetHandlerV1(req.HandlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	err := handler.Start()
	if err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: req.HandlerID,
		Message:   "Connection started",
	})
}

func stopConnection(c *gin.Context) {
	var req struct {
		HandlerID string `json:"handler_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	handler := Mavlink.GetHandlerV1(req.HandlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	err := handler.Stop()
	if err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: req.HandlerID,
		Message:   "Connection stopped",
	})
}

func restartConnection(c *gin.Context) {
	var req struct {
		HandlerID string `json:"handler_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	handler := Mavlink.GetHandlerV1(req.HandlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	err := handler.Restart()
	if err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: req.HandlerID,
		Message:   "Connection restarted",
	})
}

func droneTakeoff(c *gin.Context) {
	var req struct {
		HandlerID string  `json:"handler_id" binding:"required"`
		Altitude  float32 `json:"altitude" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	handler := Mavlink.GetHandlerV1(req.HandlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	err := handler.SendTakeoff(req.Altitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: req.HandlerID,
		Message:   "Drone taking off",
	})
}

func droneLand(c *gin.Context) {
	var req struct {
		HandlerID string  `json:"handler_id" binding:"required"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Altitude  float64 `json:"altitude"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	handler := Mavlink.GetHandlerV1(req.HandlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	err := handler.SendLand(req.Latitude, req.Longitude, req.Altitude)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: req.HandlerID,
		Message:   "Drone landing",
	})
}

func droneMove(c *gin.Context) {
	var req struct {
		HandlerID string  `json:"handler_id" binding:"required"`
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
		Altitude  float64 `json:"altitude" binding:"required"`
		Speed     float32 `json:"speed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	handler := Mavlink.GetHandlerV1(req.HandlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	err := handler.SendMoveToPosition(req.Latitude, req.Longitude, req.Altitude, req.Speed)
	if err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: req.HandlerID,
		Message:   "Drone moving",
	})
}

func droneReturn(c *gin.Context) {
	var req struct {
		HandlerID string `json:"handler_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	handler := Mavlink.GetHandlerV1(req.HandlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	err := handler.SendReturnToLaunch()
	if err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: req.HandlerID,
		Message:   "Drone returning to launch",
	})
}

func droneMode(c *gin.Context) {
	var req struct {
		HandlerID string `json:"handler_id" binding:"required"`
		Mode      string `json:"mode" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	handler := Mavlink.GetHandlerV1(req.HandlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	err := handler.SetFlightMode(Mavlink.FlightMode(req.Mode))
	if err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: req.HandlerID,
		Message:   "Flight mode set",
	})
}

func getDroneStatus(c *gin.Context) {
	handlerID := c.Query("handler_id")
	if handlerID == "" {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   "handler_id is required",
		})
		return
	}

	handler := Mavlink.GetHandlerV1(handlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	status := handler.GetDroneStatus()
	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Data:      status,
	})
}

func getDronePosition(c *gin.Context) {
	handlerID := c.Query("handler_id")
	if handlerID == "" {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   "handler_id is required",
		})
		return
	}

	handler := Mavlink.GetHandlerV1(handlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	position := handler.GetDronePosition()
	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Data:      position,
	})
}

func getDroneAttitude(c *gin.Context) {
	handlerID := c.Query("handler_id")
	if handlerID == "" {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   "handler_id is required",
		})
		return
	}

	handler := Mavlink.GetHandlerV1(handlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	attitude := handler.GetDroneAttitude()
	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Data:      attitude,
	})
}

func getDroneBattery(c *gin.Context) {
	handlerID := c.Query("handler_id")
	if handlerID == "" {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   "handler_id is required",
		})
		return
	}

	handler := Mavlink.GetHandlerV1(handlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	drone := handler.GetDrone()
	if drone == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Drone not found",
		})
		return
	}

	battery := drone.GetBattery()
	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Data:      battery,
	})
}

func setGroundStation(c *gin.Context) {
	var req struct {
		HandlerID string  `json:"handler_id" binding:"required"`
		Name      string  `json:"name" binding:"required"`
		ID        string  `json:"id"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Altitude  float64 `json:"altitude"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	handler := Mavlink.GetHandlerV1(req.HandlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	handler.SetGroundStation(req.Name, req.ID, req.Latitude, req.Longitude, req.Altitude)
	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: req.HandlerID,
		Message:   "Ground station set",
	})
}

func getGroundStation(c *gin.Context) {
	handlerID := c.Query("handler_id")
	if handlerID == "" {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   "handler_id is required",
		})
		return
	}

	handler := Mavlink.GetHandlerV1(handlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	info := handler.GetGroundStationInfo()
	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Data:      info,
	})
}

func requestStream(c *gin.Context) {
	var req struct {
		HandlerID string `json:"handler_id" binding:"required"`
		StreamID  int    `json:"stream_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	handler := Mavlink.GetHandlerV1(req.HandlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: req.HandlerID,
		Message:   "Stream requested",
	})
}

func sendHeartbeat(c *gin.Context) {
	var req struct {
		HandlerID string `json:"handler_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	handler := Mavlink.GetHandlerV1(req.HandlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler not found",
		})
		return
	}

	err := handler.SendHeartbeat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: req.HandlerID,
		Message:   "Heartbeat sent",
	})
}
