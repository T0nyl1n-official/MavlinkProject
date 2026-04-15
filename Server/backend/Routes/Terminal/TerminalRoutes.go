package TerminalRoutes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	JwtMiddleware "MavlinkProject/Server/backend/Middles"
	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"

	Conf "MavlinkProject/Server/backend/Config"
	Terminal "MavlinkProject/Server/backend/Config/Terminal"
	User "MavlinkProject/Server/backend/Shared/User"
)

type BackendOperations interface {
	Restart()
	Shutdown() error
}

type TerminalHandler struct {
	DB             *gorm.DB
	SettingManager *Conf.SettingManager
	BackendOps     BackendOperations
}

func (h *TerminalHandler) HandleTerminalMessage(c *gin.Context) {
	var tm Terminal.TerminalManager

	if err := c.ShouldBindJSON(&tm); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "User not authenticated",
		})
		return
	}

	user := &User.User{}
	user.ID = userID.(uint)

	userclaims, _ := c.Get("claims")
	if claims, ok := userclaims.(*jwtUtils.JWTClaims); ok {
		user.Username = claims.Username
		if claims.Role == "admin" {
			user.IsAdmin = true
		}
	}

	tm.User = user
	tm.DB = h.DB
	tm.SettingManager = h.SettingManager

	response := tm.Handle()

	if response.Message != nil {
		if msg, ok := response.Message.(map[string]interface{}); ok {
			cmd, _ := msg["command"].(string)
			if cmd == "reboot" || cmd == "shutdown" {
				if h.BackendOps != nil {
					go func() {
						if cmd == "reboot" {
							h.BackendOps.Restart()
						} else {
							h.BackendOps.Shutdown()
						}
					}()
				}
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

type TerminalRouteConfig struct {
	BackendOps BackendOperations
}

func SetTerminalRoutes(r *gin.Engine, jwtManager *jwtUtils.JWTManager, tokenManager *Jwt.RedisTokenManager, mysqlDB *gorm.DB, settingManager *Conf.SettingManager, config TerminalRouteConfig) {
	terminalHandler := &TerminalHandler{
		DB:             mysqlDB,
		SettingManager: settingManager,
		BackendOps:     config.BackendOps,
	}

	terminal := r.Group("/terminal")
	terminal.Use(JwtMiddleware.JwtAuthMiddleWareWithRedis(jwtManager, tokenManager, nil))
	{
		terminal.POST("/message", terminalHandler.HandleTerminalMessage)
	}
}
