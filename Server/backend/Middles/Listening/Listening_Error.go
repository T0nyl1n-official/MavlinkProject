package MiddleWare

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"

	WarningHandler "MavlinkProject/Server/Backend/Utils/WarningHandle"
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

var defaultListeningConfig = ListeningConfig{
	EnablePanicRecovery: true,
	EnableErrorLogging:  true,
	EnableWarningPush:   true,
	Sources:             []string{"handler", "middleware", "route", "database", "chain"},
}

func ListeningErrorMiddleWare() gin.HandlerFunc {
	return ListeningErrorWithConfig(defaultListeningConfig)
}

func ListeningErrorWithConfig(config ListeningConfig) gin.HandlerFunc {
	listenerInit.Do(func() {
		if config.EnablePanicRecovery {
			log.Println("[ListeningError] 全局错误监听已启动 - Panic恢复已启用")
		}
	})

	return func(c *gin.Context) {
		defer func() {
			if config.EnablePanicRecovery {
				if panicValue := recover(); panicValue != nil {
					stack := string(debug.Stack())
					errorMsg := fmt.Sprintf("Panic recovered: %v", panicValue)

					if config.EnableErrorLogging {
						log.Printf("[PANIC] %s\nStack Trace:\n%s", errorMsg, stack)
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

						if config.EnableErrorLogging {
							log.Printf("[ERROR] Source: %s, Error: %s", source, errorMsg)
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
