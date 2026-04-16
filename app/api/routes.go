package api

import (
	"github.com/gin-gonic/gin"
	"MavlinkProject_Board/app/api/handlers"
)

func SetupRoutes(router *gin.Engine) {
	// 健康检查
	router.GET("/health", handlers.HealthCheck)

	// API 路由组
	api := router.Group("/api")
	{
		// Board 路由
		board := api.Group("/board")
		{
			board.POST("/message", handlers.HandleBoardMessage)
			board.GET("/status", handlers.GetBoardStatus)
		}

		// 任务链路由
		chain := api.Group("/chain")
		{
			chain.POST("/create", handlers.CreateTaskChain)
			chain.GET("/list", handlers.ListTaskChains)
			chain.GET("/:id", handlers.GetTaskChain)
		}
	}
}
