package MiddleWare

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	Conf "MavlinkProject/Server/backend/Config"
)

func CORSMiddleware() gin.HandlerFunc {
	setting := Conf.GetSetting()
	cfg := setting.CORS

	return cors.New(cors.Config{
		AllowOrigins:     cfg.AllowOrigins,
		AllowMethods:     cfg.AllowMethods,
		AllowHeaders:     cfg.AllowHeaders,
		ExposeHeaders:    cfg.ExposeHeaders,
		AllowCredentials: cfg.AllowCredentials,
		MaxAge:           time.Duration(cfg.MaxAge) * time.Second,
	})
}
