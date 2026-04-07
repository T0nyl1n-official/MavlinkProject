package listening

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	Conf "MavlinkProject/Server/backend/Config"
	WarningHandler "MavlinkProject/Server/backend/Utils/WarningHandle"
)

var (
	listenerInit sync.Once
)

type ListeningConfig struct {
	EnablePanicRecovery bool
	EnableErrorLogging  bool
	EnableWarningPush   bool
	Sources             []string
}

func GetDefaultListeningConfig() ListeningConfig {
	setting := Conf.GetSetting()
	errCfg := setting.ErrorListener

	return ListeningConfig{
		EnablePanicRecovery: errCfg.EnablePanicRecovery,
		EnableErrorLogging:  errCfg.EnableErrorLogging,
		EnableWarningPush:   errCfg.EnableWarningPush,
		Sources:             errCfg.Sources,
	}
}

func ListeningErrorMiddleWare() gin.HandlerFunc {
	return ListeningErrorWithConfig(GetDefaultListeningConfig())
}

// ListeningErrorWithConfig 配置监听错误中间件
func ListeningErrorWithConfig(config ListeningConfig) gin.HandlerFunc {
	listenerInit.Do(func() {
		if config.EnablePanicRecovery {
			log.Printf("[%s] [ListeningError] 全局错误监听已启动 - Panic恢复已启用", time.Now().Format("2006-01-02 15:04:05"))
		}
	})

	return func(c *gin.Context) {
		defer func() {
			if config.EnablePanicRecovery {
				if panicValue := recover(); panicValue != nil {
					stack := string(debug.Stack())
					errorMsg := fmt.Sprintf("Panic recovered: %v", panicValue)
					timestamp := time.Now().Format("2006-01-02 15:04:05")

					if config.EnableErrorLogging {
						log.Printf("[%s] [PANIC] %s\nStack Trace:\n%s", timestamp, errorMsg, stack)
					}

					if config.EnableWarningPush {
						source := detectErrorSource(c)
						WarningHandler.HandleBackendError(errorMsg, source, stack)
					}

					c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
						"success": false,
						"error":   "服务器内部错误，请稍后重试",
						"code":    500,
					})
				}
			}
		}()

		c.Next()

		if config.EnableErrorLogging || config.EnableWarningPush {
			if len(c.Errors) > 0 {
				for _, err := range c.Errors {
					if err != nil {
						errorMsg := err.Error()
						source := detectErrorSource(c)
						timestamp := time.Now().Format("2006-01-02 15:04:05")

						if config.EnableErrorLogging {
							log.Printf("[%s] [ERROR] Source: %s, Error: %s", timestamp, source, errorMsg)
						}

						if config.EnableWarningPush {
							WarningHandler.DistributeError(errorMsg, source)
						}
					}
				}
			}
		}
	}
}

// 检查错误来源路径
func detectErrorSource(c *gin.Context) string {
	path := c.Request.URL.Path

	switch {
	case strings.HasPrefix(path, "/mavlink/v"):
		return "mavlink"
	case strings.HasPrefix(path, "/api/chain"):
		return "agent"
	case strings.HasPrefix(path, "/users"):
		return "handler"
	case strings.HasPrefix(path, "/admin"):
		return "middleware"
	default:
		return "route"
	}
}
