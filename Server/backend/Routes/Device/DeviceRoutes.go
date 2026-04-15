package DeviceRoutes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	DeviceHandler "MavlinkProject/Server/backend/Handler/Device"
	JwtMiddleWare "MavlinkProject/Server/backend/Middles"
	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
)

type DeviceRoutes struct {
	DB           *gorm.DB
	JWTManager   *jwtUtils.JWTManager
	TokenManager *Jwt.RedisTokenManager
}

func SetDeviceRoutes(r *gin.Engine, jwtManager *jwtUtils.JWTManager, tokenManager *Jwt.RedisTokenManager, mysqlDB *gorm.DB) {
	JwtMiddleWare.SetDeviceJWTManagers(jwtManager, tokenManager)

	deviceHandler := &DeviceHandler.DeviceHandler{
		Mysql:      mysqlDB,
		JWTManager: jwtManager,
	}

	DeviceHandler.SetDeviceJWTManager(jwtManager)
	DeviceHandler.SetDeviceRedisTokenManager(tokenManager)

	device := r.Group("/device")
	{
		device.POST("/login", deviceHandler.LoginDevice)
		device.POST("/logout", deviceHandler.LogoutDevice)
	}
}
