package MiddleWare

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000", // 前端开发服务器地址(默认)
		},

		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{
			"Origin",
			"Content-Type", // context-JSON
			"Accept",
			"Authorization",
			"X-Requested-With", // 跨域请求时需要的头信息
		},

		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,           // 必须带cookies
		MaxAge:           12 * time.Hour, // 控制浏览器缓存预检请求（OPTIONS）结果的时间
	})
}
