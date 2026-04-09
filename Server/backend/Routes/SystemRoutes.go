package Routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	Conf "MavlinkProject/Server/backend/Config"
	JwtMiddleware "MavlinkProject/Server/backend/Middles"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
)

func InitSystemRoutes(r *gin.Engine, settingManager *Conf.SettingManager, jwtManager *jwtUtils.JWTManager) {
	systemGroup := r.Group("/system")
	{
		systemGroup.GET("/setting", JwtMiddleware.JwtAuthMiddleWare(jwtManager, nil), func(c *gin.Context) {
			claims, exists := c.Get("claims")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "unauthorized"})
				return
			}
			jwtClaims := claims.(*jwtUtils.JWTClaims)
			if jwtClaims.Role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "admin permission required"})
				return
			}

			setting := settingManager.GetSetting()
			c.JSON(http.StatusOK, gin.H{
				"code": 0,
				"msg":  "success",
				"data": setting,
			})
		})

		systemGroup.GET("/setting/:category", func(c *gin.Context) {
			category := c.Param("category")
			setting := settingManager.GetSetting()

			switch category {
			case "database":
				c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": setting.Database})
			case "redis":
				c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": setting.Redis})
			case "jwt":
				c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": setting.JWT})
			case "cors":
				c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": setting.CORS})
			case "rate_limit":
				c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": setting.RateLimit})
			case "logger":
				c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": setting.Logger})
			case "resources":
				c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": setting.Resources})
			case "verification":
				c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": setting.Verification})
			case "error_listener":
				c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": setting.ErrorListener})
			case "board":
				c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "success", "data": setting.Board})
			default:
				c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "unknown category"})
			}
		})

		systemGroup.POST("/setting/reload", JwtMiddleware.JwtAuthMiddleWare(jwtManager, nil), func(c *gin.Context) {
			claims, exists := c.Get("claims")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "unauthorized"})
				return
			}
			jwtClaims := claims.(*jwtUtils.JWTClaims)
			if jwtClaims.Role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "admin permission required"})
				return
			}
			if err := settingManager.ReloadSetting(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "reload failed: " + err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "setting reloaded successfully"})
		})

		systemGroup.POST("/setting/update", JwtMiddleware.JwtAuthMiddleWare(jwtManager, nil), func(c *gin.Context) {
			claims, exists := c.Get("claims")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "unauthorized"})
				return
			}
			jwtClaims := claims.(*jwtUtils.JWTClaims)
			if jwtClaims.Role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "admin permission required"})
				return
			}
			var newSetting Conf.Setting
			if err := c.ShouldBindJSON(&newSetting); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "invalid request: " + err.Error()})
				return
			}

			if err := settingManager.UpdateSetting(&newSetting); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "update failed: " + err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "setting updated successfully"})
		})

		systemGroup.POST("/setting/:category", JwtMiddleware.JwtAuthMiddleWare(jwtManager, nil), func(c *gin.Context) {
			claims, exists := c.Get("claims")
			if !exists {
				c.JSON(http.StatusUnauthorized, gin.H{"code": 1, "msg": "unauthorized"})
				return
			}
			jwtClaims := claims.(*jwtUtils.JWTClaims)
			if jwtClaims.Role != "admin" {
				c.JSON(http.StatusForbidden, gin.H{"code": 1, "msg": "admin permission required"})
				return
			}
			category := c.Param("category")
			setting := settingManager.GetSetting()

			switch category {
			case "database":
				var config Conf.DatabaseConfig
				if err := c.ShouldBindJSON(&config); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
					return
				}
				setting.Database = config
			case "redis":
				var config Conf.RedisConfig
				if err := c.ShouldBindJSON(&config); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
					return
				}
				setting.Redis = config
			case "jwt":
				var config Conf.JWTConfig
				if err := c.ShouldBindJSON(&config); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
					return
				}
				setting.JWT = config
			case "cors":
				var config Conf.CORSConfig
				if err := c.ShouldBindJSON(&config); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
					return
				}
				setting.CORS = config
			case "rate_limit":
				var config Conf.RateLimitConfig
				if err := c.ShouldBindJSON(&config); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
					return
				}
				setting.RateLimit = config
			case "logger":
				var config Conf.LoggerConfig
				if err := c.ShouldBindJSON(&config); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
					return
				}
				setting.Logger = config
			case "resources":
				var config Conf.ResourcesConfig
				if err := c.ShouldBindJSON(&config); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
					return
				}
				setting.Resources = config
			case "verification":
				var config Conf.VerificationConfig
				if err := c.ShouldBindJSON(&config); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
					return
				}
				setting.Verification = config
			case "error_listener":
				var config Conf.ErrorListenerConfig
				if err := c.ShouldBindJSON(&config); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
					return
				}
				setting.ErrorListener = config
			case "board":
				var config Conf.BoardConfig
				if err := c.ShouldBindJSON(&config); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": err.Error()})
					return
				}
				setting.Board = config
			default:
				c.JSON(http.StatusBadRequest, gin.H{"code": 1, "msg": "unknown category"})
				return
			}

			if err := settingManager.UpdateSetting(setting); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "update failed: " + err.Error()})
				return
			}

			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "category updated and components restarted", "category": category})
		})
	}
}
