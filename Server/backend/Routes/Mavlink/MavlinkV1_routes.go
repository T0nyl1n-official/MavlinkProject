package MavlinkRoute

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	Mavlink "MavlinkProject/Server/backend/Handler/Mavlink"
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

func SetupMavlinkV1Routes(router *gin.Engine) {
	v1Group := router.Group("/mavlink/v1")
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

func getHandler(id string) *Mavlink.MAVLinkHandlerV1 {
	return Mavlink.GetHandlerV1(id)
}

func setHandler(id string, handler *Mavlink.MAVLinkHandlerV1) {
	// 使用 handler 内部的 handlerID 存储
	// 这里暂时保留用于兼容
}

func removeHandler(id string) {
	Mavlink.DeleteHandlerV1(id)
}

func parseConfig(c *gin.Context) (string, *Mavlink.MAVLinkConfigV1, error) {
	handlerID := c.GetString("handler_id")
	if handlerID == "" {
		handlerID = c.Query("handler_id")
	}

	if handlerID != "" {
		handler := getHandler(handlerID)
		if handler != nil {
			cfg := handler.GetConfig()
			return handlerID, &cfg, nil
		}
	}

	var config ConnectionConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		var queryErr error
		handlerID = c.Query("handler_id")
		if handlerID != "" {
			handler := getHandler(handlerID)
			if handler != nil {
				cfg := handler.GetConfig()
				return handlerID, &cfg, nil
			}
			queryErr = nil
		}
		if queryErr != nil {
			return "", nil, queryErr
		}
		return "", nil, err
	}

	mavlinkConfig := &Mavlink.MAVLinkConfigV1{
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

	return handlerID, mavlinkConfig, nil
}

func getOrCreateHandler(c *gin.Context) (*Mavlink.MAVLinkHandlerV1, string, error) {
	handlerID, config, err := parseConfig(c)
	if err != nil {
		return nil, "", err
	}

	var handler *Mavlink.MAVLinkHandlerV1

	if handlerID != "" {
		handler = getHandler(handlerID)
		if handler != nil {
			return handler, handlerID, nil
		}
	}

	handler = Mavlink.NewMAVLinkHandlerV1(*config)
	handlerID = handler.GetHandlerID()
	setHandler(handlerID, handler)

	return handler, handlerID, nil
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
	handlerID := handler.GetHandlerID()
	setHandler(handlerID, handler)

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Message:   "Handler创建成功",
	})
}

func deleteHandler(c *gin.Context) {
	handlerID := c.Param("id")
	handler := getHandler(handlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler不存在",
		})
		return
	}

	handler.Stop()
	removeHandler(handlerID)

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success: true,
		Message: "Handler已删除",
	})
}

func getHandlerStatus(c *gin.Context) {
	handlerID := c.Param("id")
	handler := getHandler(handlerID)
	if handler == nil {
		c.JSON(http.StatusNotFound, MavlinkV1Response{
			Success: false,
			Error:   "Handler不存在",
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success: true,
		Data: gin.H{
			"handler_id":        handler.GetHandlerID(),
			"connection_status": handler.GetConnectionStatus(),
			"config":            handler.GetConfig(),
		},
	})
}

func startConnection(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := handler.Start(); err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Message:   "连接已启动",
	})
}

func stopConnection(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := handler.Stop(); err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Message:   "连接已停止",
	})
}

func restartConnection(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := handler.Restart(); err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Message:   "连接已重启",
	})
}

func droneTakeoff(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req struct {
		Altitude float32 `json:"altitude" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	if err := handler.SendTakeoff(req.Altitude); err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Message:   "起飞指令已发送",
	})
}

func droneLand(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Altitude  float64 `json:"altitude"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	if err := handler.SendLand(req.Latitude, req.Longitude, req.Altitude); err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Message:   "降落指令已发送",
	})
}

func droneMove(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req struct {
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
		Altitude  float64 `json:"altitude" binding:"required"`
		Speed     float32 `json:"speed"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	speed := req.Speed
	if speed == 0 {
		speed = 5.0
	}

	if err := handler.SendMoveToPosition(req.Latitude, req.Longitude, req.Altitude, speed); err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Message:   "移动指令已发送",
	})
}

func droneReturn(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := handler.SendReturnToLaunch(); err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Message:   "返航指令已发送",
	})
}

func droneMode(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req struct {
		Mode string `json:"mode" binding:"required,oneof=manual stabilize auto guided rtl land"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	mode := Mavlink.FlightMode(req.Mode)
	if err := handler.SetFlightMode(mode); err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Message:   "飞行模式已设置",
	})
}

func getDroneStatus(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Data: gin.H{
			"drone_status": handler.GetDroneStatus(),
			"position":     handler.GetDronePosition(),
			"attitude":     handler.GetDroneAttitude(),
			"battery":      handler.GetDroneBattery(),
		},
	})
}

func getDronePosition(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Data:      handler.GetDronePosition(),
	})
}

func getDroneAttitude(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Data:      handler.GetDroneAttitude(),
	})
}

func getDroneBattery(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Data:      handler.GetDroneBattery(),
	})
}

func setGroundStation(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req struct {
		Name      string  `json:"name" binding:"required"`
		ID        string  `json:"id" binding:"required"`
		Latitude  float64 `json:"latitude" binding:"required"`
		Longitude float64 `json:"longitude" binding:"required"`
		Altitude  float64 `json:"altitude" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	handler.SetGroundStation(req.Name, req.ID, req.Latitude, req.Longitude, req.Altitude)

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Message:   "地面站已设置",
	})
}

func getGroundStation(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Data:      handler.GetGroundStation(),
	})
}

func requestStream(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	var req struct {
		MessageID int `json:"message_id" binding:"required"`
		Rate      int `json:"rate" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	if err := handler.RequestMessageStream(req.MessageID, req.Rate); err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Message:   "数据流请求已发送",
	})
}

func sendHeartbeat(c *gin.Context) {
	handler, handlerID, err := getOrCreateHandler(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, MavlinkV1Response{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	if err := handler.SendHeartbeat(); err != nil {
		c.JSON(http.StatusInternalServerError, MavlinkV1Response{
			Success:   false,
			HandlerID: handlerID,
			Error:     err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, MavlinkV1Response{
		Success:   true,
		HandlerID: handlerID,
		Message:   "心跳已发送",
	})
}
