package Routes

import (
	gin "github.com/gin-gonic/gin"
	gorm "gorm.io/gorm"

	Conf "MavlinkProject/Server/backend/Config"
	BoardsRoutes "MavlinkProject/Server/backend/Routes/Boards"
	MavlinkRoutes "MavlinkProject/Server/backend/Routes/Mavlink"
	MiscRoutes "MavlinkProject/Server/backend/Routes/Misc"
	SensorRoutes "MavlinkProject/Server/backend/Routes/Sensor"
	UsersRoutes "MavlinkProject/Server/backend/Routes/User"

	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
)

func InitAllRoutes(r *gin.Engine, jwtManager *jwtUtils.JWTManager, tokenManager *Jwt.RedisTokenManager, mysqlDB *gorm.DB, settingManager *Conf.SettingManager) {
	r.StaticFile("/favicon.ico", "./Resources/favicon.ico")
	Test_Routes(r)

	InitSystemRoutes(r, settingManager, jwtManager)

	MiscRoutes.SetMiscRoutes(r)
	UsersRoutes.SetUsersRoutes(r, jwtManager, tokenManager, mysqlDB)
	BoardsRoutes.SetupBoardRoutes(r, jwtManager, tokenManager)
	MavlinkRoutes.SetupChainRoutes(r, jwtManager, tokenManager)
	MavlinkRoutes.SetupDefaultMavlinkRoutesV2(r, jwtManager, tokenManager)
	MavlinkRoutes.SetupMavlinkV1Routes(r, jwtManager, tokenManager)

	SensorRoutes.SetupSensorRoutes(r)
}

func Test_Routes(r *gin.Engine) {
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "success",
			"message": "Hello world! - Welcome to The Mavlink Project!",
			"version": "Pre-Release 0.1.6",
		})
	})
}
