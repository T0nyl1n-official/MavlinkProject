package BoardsRoutes

import (
	"github.com/gin-gonic/gin"

	LiveStreamHandler "MavlinkProject/Server/backend/Handler/Boards"
	MiddleWare "MavlinkProject/Server/backend/Middles"
	Middleware "MavlinkProject/Server/backend/Middles"
	UserAgentMiddleware "MavlinkProject/Server/backend/Middles"
	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
)

var liveHandler *LiveStreamHandler.LiveStreamHandler

func SetLiveStreamRoutes(r *gin.Engine, jwtManager *jwtUtils.JWTManager, tokenManager *Jwt.RedisTokenManager) {
	liveHandler = LiveStreamHandler.NewLiveStreamHandler()

	boardLive := r.Group("/api/board/live")
	boardLive.Use(Middleware.DeviceJwtAuthMiddleware())
	boardLive.Use(UserAgentMiddleware.UserAgentCheckMiddleware())
	{
		boardLive.POST("", liveHandler.HandleCentralUploadWithMetadata)
		boardLive.POST("/raw", liveHandler.HandleCentralUpload)
		boardLive.POST("/rtmp/start", liveHandler.HandleStartRTMPTranslator)
		boardLive.POST("/rtmp/stop", liveHandler.HandleStopRTMPTranslator)
		boardLive.GET("/rtmp/status", liveHandler.HandleRTMPStatus)
		boardLive.POST("/ffmpeg", liveHandler.HandleFFmpegDirect)
	}

	frontendLive := r.Group("/api/backend/live")
	frontendLive.Use(MiddleWare.JwtAuthMiddleWareWithRedis(jwtManager, tokenManager, nil))
	{
		frontendLive.GET("", liveHandler.HandleFrontendGetStream)
		frontendLive.GET("/ws", liveHandler.HandleFrontendWebSocket)
		frontendLive.GET("/list", liveHandler.HandleListStreams)
		frontendLive.GET("/info/:stream_id", liveHandler.HandleGetStreamInfo)
		frontendLive.DELETE("/:stream_id", liveHandler.HandleStopStream)
	}
}
