package boardHandler

import (
	gin "github.com/gin-gonic/gin"

	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"

	Board "MavlinkProject/Server/Backend/Shared/Boards"
)

type BoardHandler struct {
	Board      *Board.Board
	Router     *gin.Engine
	Connection string
	Message    *Board.BoardMessage
	JWTManager *jwtUtils.JWTManager
}
