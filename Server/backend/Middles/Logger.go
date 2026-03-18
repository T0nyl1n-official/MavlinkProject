package MiddleWare

import (
	
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	ErrorsMgr "MavlinkProject/Server/backend/Middles/ErrorMiddleHandle/ErrorsMgr"
)

// 访问频率监控结构
type AccessMonitor struct {
	mu            sync.RWMutex
	ipAccessCount map[string]int // IP访问次数
	urlErrorCount map[string]int // URL错误次数
	lastResetTime time.Time      // 上次重置时间
	logFile       *os.File       // 日志文件句柄
	logFileDate   string         // 当前日志文件日期
}

var (
	accessMonitor *AccessMonitor
	monitorOnce   sync.Once
)

// 日志级别
const (
	LogLevelDebug = "DEBUG"
	LogLevelInfo  = "INFO"
	LogLevelWarn  = "WARN"
	LogLevelError = "ERROR"
)

// 阈值配置
const (
	IPAccessThreshold    = 50                // IP访问阈值
	URLErrorThreshold    = 50                // URL错误阈值
	SlowRequestThreshold = 1 * time.Second   // 慢请求阈值
	MonitorWindow        = 5 * time.Minute   // 监控窗口
	MaxLogFileSize       = 100 * 1024 * 1024 // 100MB 最大日志文件大小
)

var (
	logDir = "./OutputLogs"
)

// 初始化访问监控器
func initAccessMonitor() {
	monitorOnce.Do(func() {
		accessMonitor = &AccessMonitor{
			ipAccessCount: make(map[string]int),
			urlErrorCount: make(map[string]int),
			lastResetTime: time.Now(),
		}

		// 创建日志目录
		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Printf("创建日志目录失败: %v", err)
		}

		// 启动定时清理任务
		go accessMonitor.cleanupTask()
	})
}

// 定时清理过期数据
func (am *AccessMonitor) cleanupTask() {
	ticker := time.NewTicker(MonitorWindow)
	defer ticker.Stop()

	for range ticker.C {
		am.mu.Lock()
		am.ipAccessCount = make(map[string]int)
		am.urlErrorCount = make(map[string]int)
		am.lastResetTime = time.Now()
		am.mu.Unlock()

		am.logToFile("INFO", "监控数据已重置")
	}
}

// 记录到文件
func (am *AccessMonitor) logToFile(level, message string) {
	currentDate := time.Now().Format("2006-01-02")

	// 检查是否需要创建新的日志文件（日期变更或文件过大）
	if am.logFileDate != currentDate || am.logFile == nil || am.shouldRotateLogFile() {
		if am.logFile != nil {
			am.logFile.Close()
		}

		logPath := filepath.Join(logDir,
			fmt.Sprintf("log_%s.log", currentDate))

		file, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Printf("打开日志文件失败: %v", err)
			return
		}

		am.logFile = file
		am.logFileDate = currentDate
	}

	logEntry := fmt.Sprintf("[%s] %s %s\n",
		time.Now().Format("2006-01-02 15:04:05"), level, message)

	if _, err := am.logFile.WriteString(logEntry); err != nil {
		log.Printf("写入日志文件失败: %v", err)
	}
}

// 检查是否需要轮转日志文件
func (am *AccessMonitor) shouldRotateLogFile() bool {
	if am.logFile == nil {
		return false
	}

	fileInfo, err := am.logFile.Stat()
	if err != nil {
		return false
	}

	return fileInfo.Size() >= MaxLogFileSize
}

// 检查IP访问频率
func (am *AccessMonitor) checkIPAccess(ip string) {
	am.mu.Lock()
	defer am.mu.Unlock()

	am.ipAccessCount[ip]++

	if am.ipAccessCount[ip] >= IPAccessThreshold {
		warningMsg := fmt.Sprintf("🚨 IP频率警告: %s 在5分钟内访问了%d次", ip, am.ipAccessCount[ip])
		log.Printf("%s", warningMsg)
		am.logToFile("WARN", warningMsg)
	}
}

// 检查URL错误频率
func (am *AccessMonitor) checkURLError(url string, statusCode int) {
	if statusCode < 400 {
		return // 只监控错误状态码
	}

	am.mu.Lock()
	defer am.mu.Unlock()

	am.urlErrorCount[url]++

	if am.urlErrorCount[url] >= URLErrorThreshold {
		warningMsg := fmt.Sprintf("🚨 URL错误警告: %s 在5分钟内出现%d次错误(状态码: %d)",
			url, am.urlErrorCount[url], statusCode)
		log.Printf("%s", warningMsg)
		am.logToFile("WARN", warningMsg)
	}
}

// logErrorMgrError 记录ErrorMgr错误
func (am *AccessMonitor) logErrorMgrError(c *gin.Context) {
	// 检查是否有ErrorMgr错误
	if len(c.Errors) > 0 {
		for _, ginErr := range c.Errors {
			err := ginErr.Err

			// 检查是否是ErrorMgr的错误详情类型
			if errorDetail, ok := err.(*ErrorsMgr.ErrorDetail); ok {
				// 确定错误状态（已解决/未解决）
				status := "UNSOLVED"
				if errorDetail.HTTPStatus < 500 {
					status = "SOLVED" // 4xx错误通常表示已处理的客户端错误
				}

				// 构建错误描述
				errorDesc := fmt.Sprintf("[%s][%s] [%s] - %s (Code: %s, HTTP: %d, Module: %s)",
					time.Now().Format("2006-01-02 15:04:05"),
					"ERROR",
					status,
					errorDetail.Message,
					errorDetail.Code,
					errorDetail.HTTPStatus,
					errorDetail.Module)

				// 记录到控制台
				log.Printf("%s", errorDesc)

				// 记录到文件
				am.logToFile("ERROR", errorDesc)

				// 如果有堆栈跟踪，也记录下来
				if len(errorDetail.StackTrace) > 0 {
					stackTrace := fmt.Sprintf("Stack Trace for %s:", errorDetail.Code)
					for _, line := range errorDetail.StackTrace {
						stackTrace += "\n" + line
					}
					am.logToFile("DEBUG", stackTrace)
				}

				// 如果有上下文信息，记录下来
				if len(errorDetail.Context) > 0 {
					contextJSON, _ := json.Marshal(errorDetail.Context)
					am.logToFile("DEBUG", fmt.Sprintf("Context for %s: %s",
						errorDetail.Code, string(contextJSON)))
				}
			}
		}
	}
}

// 结构化日志中间件
func Logger(mysql *gorm.DB) gin.HandlerFunc {
	// 初始化监控器
	initAccessMonitor()

	return func(c *gin.Context) {
		start := time.Now()

		// 记录请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 继续处理请求
		c.Next()

		// 记录响应
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		urlPath := c.Request.URL.Path

		logEntry := map[string]interface{}{
			"timestamp":        start.Format(time.RFC3339),
			"method":           c.Request.Method,
			"path":             urlPath,
			"query":            c.Request.URL.RawQuery,
			"ip":               clientIP,
			"user_agent":       c.Request.UserAgent(),
			"status_code":      statusCode,
			"response_time_ms": duration.Milliseconds(),
			"request_size":     c.Request.ContentLength,
			"response_size":    c.Writer.Size(),
			"errors":           c.Errors.String(),
		}

		// 只在开发环境记录请求体
		if gin.Mode() == gin.DebugMode && len(requestBody) > 0 {
			var body map[string]interface{}
			if json.Unmarshal(requestBody, &body) == nil {
				logEntry["request_body"] = body
			}
		}

		// 结构化日志输出
		logData, _ := json.Marshal(logEntry)
		log.Printf("REQUEST: %s", string(logData))

		// 同时记录到文件
		accessMonitor.logToFile("INFO", fmt.Sprintf("%s %s %d %vms",
			c.Request.Method, urlPath, statusCode, duration.Milliseconds()))

		// 慢请求警告
		if duration > time.Second {
			warningMsg := fmt.Sprintf("🚨 慢请求警告: %s %s 耗时 %v",
				c.Request.Method, urlPath, duration)
			log.Printf("%s", warningMsg)
			accessMonitor.logToFile("WARN", warningMsg)
		}

		// 检查IP访问频率
		accessMonitor.checkIPAccess(clientIP)

		// 检查URL错误频率
		accessMonitor.checkURLError(urlPath, statusCode)

		// 记录ErrorMgr错误
		accessMonitor.logErrorMgrError(c)
	}
}
