package Routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	Conf "MavlinkProject/Server/backend/Config"
)

func InitSystemRoutes(r *gin.Engine, settingManager *Conf.SettingManager) {
	systemGroup := r.Group("/api/system")
	{
		systemGroup.GET("/setting", func(c *gin.Context) {
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

		systemGroup.POST("/setting/reload", func(c *gin.Context) {
			if err := settingManager.ReloadSetting(); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"code": 1, "msg": "reload failed: " + err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"code": 0, "msg": "setting reloaded successfully"})
		})

		systemGroup.POST("/setting/update", func(c *gin.Context) {
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

		systemGroup.POST("/setting/:category", func(c *gin.Context) {
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
