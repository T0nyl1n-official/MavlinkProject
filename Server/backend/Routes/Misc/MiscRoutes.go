package miscRoutes

import (
	gin "github.com/gin-gonic/gin"
)

func SetMiscRoutes(r *gin.Engine) {
	r.StaticFile("/robots.txt", "./Resources/robots.txt")
}
