package MiddleWare

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/microcosm-cc/bluemonday"
)

// XSSConfig XSS防护配置
type XSSConfig struct {
	// 是否启用HTML输出净化
	EnableHTMLSanitization bool
	// 是否启用严格模式（移除所有HTML标签）
	StrictMode bool
	// 是否启用CSP头
	EnableCSP bool
	// CSP策略
	CSPPolicy string
}

// DefaultXSSConfig 默认XSS配置
func DefaultXSSConfig() *XSSConfig {
	return &XSSConfig{
		EnableHTMLSanitization: true,
		StrictMode:             false, // 默认使用UGCPolicy
		EnableCSP:              true,
		// 安全的CSP策略
		CSPPolicy: "default-src 'self'; " +
			"script-src 'self'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"connect-src 'self'; " +
			"font-src 'self'; " +
			"object-src 'none'; " +
			"media-src 'self'; " +
			"frame-src 'none';",
	}
}

// StrictXSSConfig 严格XSS配置
func StrictXSSConfig() *XSSConfig {
	return &XSSConfig{
		EnableHTMLSanitization: true,
		StrictMode:             true, // 移除所有HTML
		EnableCSP:              true,
		// 最严格的CSP策略
		CSPPolicy: "default-src 'self'; " +
			"script-src 'self'; " +
			"style-src 'self'; " +
			"img-src 'self'; " +
			"connect-src 'self'; " +
			"font-src 'self'; " +
			"object-src 'none'; " +
			"media-src 'none'; " +
			"frame-src 'none';",
	}
}

// XSSMiddleware XSS防护中间件
type XSSMiddleware struct {
	config    *XSSConfig
	sanitizer *bluemonday.Policy
}

// NewXSSMiddleware 创建XSS中间件
func NewXSSMiddleware(config *XSSConfig) gin.HandlerFunc {
	if config == nil {
		config = DefaultXSSConfig()
	}

	var sanitizer *bluemonday.Policy
	if config.EnableHTMLSanitization {
		if config.StrictMode {
			sanitizer = bluemonday.StrictPolicy() // 移除所有HTML
		} else {
			sanitizer = bluemonday.UGCPolicy() // 用户生成内容策略
		}
	}

	m := &XSSMiddleware{
		config:    config,
		sanitizer: sanitizer,
	}

	return func(c *gin.Context) {
		// 1. 添加安全头
		m.addSecurityHeaders(c)

		// 2. 在上下文中存储净化器，供业务逻辑在输入时使用
		if sanitizer != nil {
			c.Set("xss_sanitizer", sanitizer)
		}

		// 3. 设置响应处理器
		if sanitizer != nil {
			originalWriter := c.Writer
			xssWriter := &xssResponseWriter{
				ResponseWriter: originalWriter,
				sanitizer:      sanitizer,
				config:         config,
			}
			c.Writer = xssWriter
		}

		c.Next()
	}
}

// addSecurityHeaders 添加安全相关的HTTP头
func (m *XSSMiddleware) addSecurityHeaders(c *gin.Context) {
	// Content Security Policy (CSP) - 最重要的XSS防御
	if m.config.EnableCSP {
		c.Header("Content-Security-Policy", m.config.CSPPolicy)
	}

	// X-XSS-Protection (旧浏览器)
	c.Header("X-XSS-Protection", "1; mode=block")

	// X-Content-Type-Options
	c.Header("X-Content-Type-Options", "nosniff")

	// X-Frame-Options (防点击劫持)
	c.Header("X-Frame-Options", "DENY")

	// Referrer-Policy
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

	c.Header("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
    c.Header("Pragma", "no-cache")
    c.Header("Expires", "0")
}

// xssResponseWriter 自定义ResponseWriter，只对HTML输出进行净化
type xssResponseWriter struct {
	gin.ResponseWriter
	sanitizer *bluemonday.Policy
	config    *XSSConfig
}

// Write 重写Write方法，只对HTML输出进行净化
func (w *xssResponseWriter) Write(data []byte) (int, error) {
	// 只对HTML内容进行净化
	if w.shouldSanitize() && w.sanitizer != nil {
		sanitized := w.sanitizer.SanitizeBytes(data)
		// 只有当内容改变时才写入新数据
		if string(sanitized) != string(data) {
			return w.ResponseWriter.Write(sanitized)
		}
	}

	return w.ResponseWriter.Write(data)
}

// WriteString 重写WriteString方法
func (w *xssResponseWriter) WriteString(s string) (int, error) {
	// 只对HTML内容进行净化
	if w.shouldSanitize() && w.sanitizer != nil {
		sanitized := w.sanitizer.Sanitize(s)
		// 只有当内容改变时才写入新数据
		if sanitized != s {
			return w.ResponseWriter.WriteString(sanitized)
		}
	}

	return w.ResponseWriter.WriteString(s)
}

// shouldSanitize 检查是否需要净化输出
func (w *xssResponseWriter) shouldSanitize() bool {
    contentType := w.Header().Get("Content-Type")

    switch {
    case strings.HasPrefix(contentType, "text/html"):
        return true
    case strings.HasPrefix(contentType, "application/xhtml+xml"):
        return true
    case strings.HasPrefix(contentType, "text/xml"):
        return true
    case strings.HasPrefix(contentType, "application/xml"):
        return true
    default:
        return false
    }
}

// GetSanitizerFromContext 从上下文中获取净化器
func GetSanitizerFromContext(c *gin.Context) *bluemonday.Policy {
	if sanitizer, exists := c.Get("xss_sanitizer"); exists {
		if policy, ok := sanitizer.(*bluemonday.Policy); ok {
			return policy
		}
	}
	return nil
}

// SanitizeInput 在业务逻辑中净化输入数据（推荐在输入时使用）
func SanitizeInput(sanitizer *bluemonday.Policy, input string) string {
	if sanitizer != nil {
		return sanitizer.Sanitize(input)
	}
	return input
}

// SanitizeUserInput 在业务逻辑中净化用户输入
func SanitizeUserInput(input string) string {
	// 使用StrictPolicy净化所有用户输入
	sanitizer := bluemonday.StrictPolicy()
	return sanitizer.Sanitize(input)
}
