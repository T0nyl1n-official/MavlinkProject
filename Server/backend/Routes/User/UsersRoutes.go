package UsersRoutes

import (
	"net/http"

	gin "github.com/gin-gonic/gin"

	Server "MavlinkProject/Server"
	UsersHandler "MavlinkProject/Server/Backend/Handler/Users"
	JwtMiddleware "MavlinkProject/Server/Backend/Middles"
	Jwt "MavlinkProject/Server/Backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/Backend/Middles/Jwt/Claims-Manager"
)

var Backend = Server.BackendServer
var h = UsersHandler.UserHandler{
	Mysql: Backend.Mysql,
}

var jwtManager *jwtUtils.JWTManager
var tokenManager *Jwt.RedisTokenManager

func init() {
	if Backend.TokenRedis != nil {
		tokenManager = Jwt.NewRedisTokenManager(Backend.TokenRedis)
	}
	jwtManager = (*jwtUtils.JWTManager)(JwtMiddleware.NewDefaultJWTManager())
}

func SetUsersRoutes(r *gin.Engine) {
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
