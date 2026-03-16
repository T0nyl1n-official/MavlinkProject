package Routes

import (
	gin "github.com/gin-gonic/gin"

	MavlinkRoutes "MavlinkProject/Server/backend/Routes/Mavlink"
	UsersRoutes "MavlinkProject/Server/backend/Routes/User"

	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
)

func InitAllRoutes(r *gin.Engine, jwtManager *jwtUtils.JWTManager, tokenManager *Jwt.RedisTokenManager) {
	UsersRoutes.SetUsersRoutes(r, jwtManager, tokenManager)
	MavlinkRoutes.SetupChainRoutes(r, jwtManager, tokenManager)
	MavlinkRoutes.SetupDefaultMavlinkRoutesV2(r, jwtManager, tokenManager)
	MavlinkRoutes.SetupMavlinkV1Routes(r, jwtManager, tokenManager)
}
