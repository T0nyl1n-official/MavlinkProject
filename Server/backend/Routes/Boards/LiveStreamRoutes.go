package BoardsRoutes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	LiveStreamHandler "MavlinkProject/Server/backend/Handler/Boards"
	MiddleWare "MavlinkProject/Server/backend/Middles"
	UserAgentMiddleware "MavlinkProject/Server/backend/Middles"
	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
)

var liveHandler *LiveStreamHandler.LiveStreamHandler

func SetLiveStreamRoutes(r *gin.Engine, jwtManager *jwtUtils.JWTManager, tokenManager *Jwt.RedisTokenManager) {
	liveHandler = LiveStreamHandler.NewLiveStreamHandler()

	boardLive := r.Group("/api/board/live")
	boardLive.Use(MiddleWare.DeviceJwtAuthMiddleware())
	boardLive.Use(UserAgentMiddleware.UserAgentCheckMiddleware())
	{
		boardLive.POST("", liveHandler.HandleCentralUploadWithMetadata)
		boardLive.POST("/raw", liveHandler.HandleCentralUpload)
	}

	frontendLive := r.Group("/api/backend/live")
	frontendLive.Use(MiddleWare.JwtAuthMiddleWareWithRedis(jwtManager, tokenManager, nil))
	{
		frontendLive.GET("", liveHandler.HandleFrontendGetStream)
		frontendLive.GET("/ws", liveHandler.HandleFrontendWebSocket)
		frontendLive.GET("/list", liveHandler.HandleListStreams)
		frontendLive.GET("/info/:stream_id", liveHandler.HandleGetStreamInfo)
		frontendLive.DELETE("/:stream_id", liveHandler.HandleStopStream)

		frontendLive.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "视频流接口不存在",
				"code":    "LIVE_ENDPOINT_NOT_FOUND",
				"available_endpoints": []string{
					"GET  /api/backend/live          - 获取视频流 (mjpeg/raw/flv)",
					"GET  /api/backend/live/ws       - WebSocket 视频流",
					"GET  /api/backend/live/list     - 获取活跃流列表",
					"GET  /api/backend/live/info/:id - 获取流详情",
					"DELETE /api/backend/live/:id    - 停止指定流",
				},
			})
		})
	}
}
