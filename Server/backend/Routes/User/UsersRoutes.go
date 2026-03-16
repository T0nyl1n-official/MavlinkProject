package UsersRoutes

import (
	"net/http"

	gin "github.com/gin-gonic/gin"
	gorm "gorm.io/gorm"

	UsersHandler "MavlinkProject/Server/backend/Handler/Users"
	JwtMiddleware "MavlinkProject/Server/backend/Middles"
	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
)

func SetUsersRoutes(r *gin.Engine, jwtManager *jwtUtils.JWTManager, tokenManager *Jwt.RedisTokenManager, mysqlDB *gorm.DB) {
	h := UsersHandler.UserHandler{Mysql: mysqlDB}

	users := r.Group("/users")
	{
		users.POST("/register", h.RegisterUser)
		users.POST("/login", h.LoginUser)
		users.GET("/profile", h.GetUserInfo)
	}

	usersAuth := r.Group("/users")
	usersAuth.Use(JwtMiddleware.JwtAuthMiddleWareWithRedis(jwtManager, tokenManager, nil))
	{
		usersAuth.POST("/update", h.UpdateUserInfo)
		usersAuth.POST("/delete", h.DeleteUser)
		usersAuth.POST("/logout", h.LogoutUser)
		usersAuth.POST("/send-email-verification", h.SendEmailVerification)
	}

	admin := r.Group("/admin")
	admin.Use(JwtMiddleware.JwtAuthMiddleWareWithRedis(jwtManager, tokenManager, []string{"admin"}))
	admin.Use(AdminAuthMiddleware())
	{
		admin.GET("/all-profile", h.GetAllUsers)
	}
}

func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			IsAdmin bool `json:"isadmin"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid request body",
			})
			c.Abort()
			return
		}

		if !req.IsAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "Admin permission required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
