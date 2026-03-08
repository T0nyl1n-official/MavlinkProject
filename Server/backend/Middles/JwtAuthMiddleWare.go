package MiddleWare

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	Config "MavlinkProject/Server/backend/Middles/Jwt/Config"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
	JwtRedis "MavlinkProject/Server/backend/Middles/Jwt"
)

// JwtAuthMiddleWareWithRedis 支持Redis token验证的JWT中间件
func JwtAuthMiddleWareWithRedis(jwtManager *jwtUtils.JWTManager, tokenManager *JwtRedis.RedisTokenManager, requiredRole []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := ExtractToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":  http.StatusBadRequest,
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		// 首先验证Redis中的Token状态
		if valid, err2 := tokenManager.ValidateToken(tokenString); !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Invalid token in Redis: " + err2.Error(),
			})
			c.Abort()
			return
		}

		// 然后验证JWT Token本身
		if valid, err3 := jwtManager.ValidateToken(tokenString); !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Invalid token claims: " + err3.Error(),
			})
			c.Abort()
			return
		}

		claims, err4 := jwtManager.ParseToken(tokenString)
		if err4 != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		// 检查角色权限
		if len(requiredRole) > 0 {
			if !contains(requiredRole, claims.Role) {
				c.JSON(http.StatusForbidden, gin.H{
					"code":  http.StatusForbidden,
					"error": fmt.Sprintf("Insufficient permissions. Required role: %v", requiredRole),
				})
				c.Abort()
				return
			}
		}

		setContext(c, claims)
		c.Next() // JWT_AUTH 结束
	}
}

// JwtAuthMiddleWare 原始JWT中间件（向后兼容）
func JwtAuthMiddleWare(jwtManager *jwtUtils.JWTManager, requiredRole []string) gin.HandlerFunc {
	if jwtManager == nil {
		jwtManager = NewDefaultJWTManager()
	}
	return func(c *gin.Context) {
		tokenString, err := ExtractToken(c)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":  http.StatusBadRequest,
				"error": "Authorization header is required",
			})
			c.Abort()
			return
		}

		if valid, err2 := jwtManager.ValidateToken(tokenString); !valid {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Invalid token claims: " + err2.Error(),
			})
			c.Abort()
			return
		}

		claims, err := jwtManager.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Invalid token",
			})
			c.Abort()
			return
		}

		setContext(c, claims)
		c.Next() // JWT_AUTH 结束
	}
}

func ExtractToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("authorization header is required")
	}

	if !strings.HasPrefix(authHeader, "Bearer ") {
		return "", errors.New("invalid Authorization header format")
	}

	tokenString := strings.TrimPrefix(authHeader, "Bearer ")
	return tokenString, nil
}

func RoleRequired(requiredRole []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || !contains(requiredRole, role.(string)) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":  http.StatusUnauthorized,
				"error": "Role not found in context",
			})
			c.Abort()
			return
		}
	}
}

func setContext(c *gin.Context, claims *jwtUtils.JWTClaims) {
	c.Set("userID", claims.UserID)
	c.Set("email", claims.UserEmail)
	c.Set("username", claims.Username)
	c.Set("role", claims.Role)
	c.Set("claims", claims)
}

func NewDefaultJWTManager() *jwtUtils.JWTManager {
	return jwtUtils.NewJWTManager(Config.DefaultJWTConfig)
}

func NewCustomJWTManager(config Config.JWTConfig) *jwtUtils.JWTManager {
	return jwtUtils.NewJWTManager(config)
}

// NewRedisClient 创建Redis客户端
func NewRedisClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // 默认Redis地址
		Password: "",               // 无密码
		DB:       0,                 // 默认数据库
	})
}

// NewRedisTokenManager 创建Redis Token管理器
func NewRedisTokenManager() *JwtRedis.RedisTokenManager {
	redisClient := NewRedisClient()
	return JwtRedis.NewRedisTokenManager(redisClient)
}

// NewDefaultJWTManagerWithRedis 创建默认的JWT管理器（带Redis支持）
func NewDefaultJWTManagerWithRedis() (*jwtUtils.JWTManager, *JwtRedis.RedisTokenManager) {
	jwtManager := NewDefaultJWTManager()
	tokenManager := NewRedisTokenManager()
	return jwtManager, tokenManager
}

func contains[T comparable](slice []T, target T) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}
