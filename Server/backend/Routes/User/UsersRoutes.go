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
    
    UsersHandler.SetJWTManager(jwtManager)
    UsersHandler.SetRedisTokenManager(tokenManager)

    h := UsersHandler.UserHandler{Mysql: mysqlDB}

    users := r.Group("/users")
    {
        users.POST("/register", h.RegisterUser)
        users.POST("/login", h.LoginUser)
    }

    usersAuth := r.Group("/users")
    usersAuth.Use(JwtMiddleware.JwtAuthMiddleWareWithRedis(jwtManager, tokenManager, nil))
    {
        usersAuth.GET("/profile/:userID", h.GetUserInfo)
        usersAuth.POST("/update/:id", h.UpdateUserInfo) 
        usersAuth.POST("/delete/:id", h.DeleteUser)
        usersAuth.POST("/logout/:id", h.LogoutUser)
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
        
        // 尝试从请求体中绑定(不仅限于Body，虽然在GET请求中不常见，但作为中间件逻辑保留)
        // 注意: 实际上 Admin 权限通常通过 Token 中的 Claims.Role 判断，前面的 JwtAuthMiddleWareWithRedis 已经做了。
        // 这个 AdminAuthMiddleware 看起来像是额外的二次验证逻辑，如果不需要可以根据实际情况简化。
        if err := c.ShouldBindJSON(&req); err == nil {
             if !req.IsAdmin {
                 c.JSON(http.StatusForbidden, gin.H{"error": "需要管理员权限"})
                 c.Abort()
                 return
             }
        }
        
        c.Next()
    }
}