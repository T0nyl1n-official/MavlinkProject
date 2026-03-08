package RateLimit

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimitMiddleware 速率限制中间件
type RateLimitMiddleware struct {
	tokenBucket *TokenBucket
	config      *RateLimitConfig
	enabled     bool // 中间件启用状态
}

// NewRateLimitMiddleware 创建速率限制中间件
func NewRateLimitMiddleware(redisClient *redis.Client, config *RateLimitConfig) *RateLimitMiddleware {
	if config == nil {
		config = DefaultRateLimitConfig()
	}

	tokenBucket := NewTokenBucket(redisClient, config)

	return &RateLimitMiddleware{
		tokenBucket: tokenBucket,
		config:      config,
		enabled:     true, // 默认启用
	}
}

// Middleware 速率限制中间件函数
func (m *RateLimitMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查中间件是否启用
		if !m.enabled {
			c.Next()
			return
		}

		// 获取请求信息
		path := c.Request.URL.Path
		method := c.Request.Method

		// 获取匹配的速率限制规则
		rule := m.config.GetRuleForRequest(path, method)

		if !rule.Enabled {
			c.Next()
			return
		}

		// 获取标识符
		identifier := m.getIdentifier(c, rule.LimitType)

		// 检查是否允许请求
		allowed, remainingTokens, err := m.tokenBucket.AllowRequest(rule, identifier)

		if err != nil {
			// 如果Redis出错，允许请求通过（fail-open策略）
			fmt.Printf("速率限制中间件错误: %v\n", err)
			c.Next()
			return
		}

		// 设置响应头
		m.setRateLimitHeaders(c, rule, remainingTokens)

		if !allowed {
			m.handleRateLimitExceeded(c, rule, identifier)
			return
		}

		c.Next()
	}
}

// getIdentifier 获取速率限制标识符
func (m *RateLimitMiddleware) getIdentifier(c *gin.Context, limitType string) string {
	switch limitType {
	case "ip":
		// 获取客户端IP
		return m.getClientIP(c)
	case "user":
		// 获取用户ID
		return m.getUserID(c)
	case "global":
		// 全局标识符
		return "global"
	default:
		// 默认使用IP
		return m.getClientIP(c)
	}
}

// getClientIP 获取客户端IP
func (m *RateLimitMiddleware) getClientIP(c *gin.Context) string {
	// 尝试从X-Forwarded-For获取IP
	forwardedFor := c.GetHeader("X-Forwarded-For")
	if forwardedFor != "" {
		ips := strings.Split(forwardedFor, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// 尝试从X-Real-IP获取IP
	realIP := c.GetHeader("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// 使用远程地址
	return c.ClientIP()
}

// getUserID 获取用户ID
func (m *RateLimitMiddleware) getUserID(c *gin.Context) string {
	// 尝试从JWT token中获取用户ID
	userID, exists := c.Get("userID")
	if exists {
		if id, ok := userID.(string); ok {
			return id
		}
		if id, ok := userID.(uint); ok {
			return fmt.Sprintf("%d", id)
		}
	}

	// 如果无法获取用户ID，回退到IP
	return m.getClientIP(c)
}

// setRateLimitHeaders 设置速率限制响应头
func (m *RateLimitMiddleware) setRateLimitHeaders(c *gin.Context, rule *RateLimitRule, remainingTokens int64) {
	c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rule.Capacity))
	c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remainingTokens))
	c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", rule.WindowSeconds))
	c.Header("X-RateLimit-Policy", fmt.Sprintf("%s;w=%d", rule.LimitType, rule.WindowSeconds))
}

// handleRateLimitExceeded 处理速率限制超出
func (m *RateLimitMiddleware) handleRateLimitExceeded(c *gin.Context, rule *RateLimitRule, identifier string) {
	// 计算重试时间
	retryAfter := m.calculateRetryAfter(rule, identifier)

	c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))

	c.JSON(http.StatusTooManyRequests, gin.H{
		"error":   "请求过于频繁",
		"message": fmt.Sprintf("速率限制已超出，请 %d 秒后重试", retryAfter),
		"data": gin.H{
			"limit_type":     rule.LimitType,
			"retry_after":    retryAfter,
			"window_seconds": rule.WindowSeconds,
		},
	})

	c.Abort()
}

// calculateRetryAfter 计算重试时间
func (m *RateLimitMiddleware) calculateRetryAfter(rule *RateLimitRule, identifier string) int64 {
	// 获取当前令牌桶状态
	info, err := m.tokenBucket.GetBucketInfo(rule, identifier)
	if err != nil {
		// 如果获取失败，返回默认重试时间
		return rule.WindowSeconds
	}

	currentTokens, ok := info["tokens"].(int64)
	if !ok {
		return rule.WindowSeconds
	}

	// 计算需要等待的时间（秒）
	if currentTokens < 0 {
		currentTokens = 0
	}

	tokensNeeded := int64(1) // 需要1个令牌
	timeToWait := int64(float64(tokensNeeded) / rule.FillRate)

	return timeToWait
}

// GetRateLimitInfo 获取速率限制信息（用于调试和管理）
func (m *RateLimitMiddleware) GetRateLimitInfo(c *gin.Context, path string, method string, identifier string) (map[string]interface{}, error) {
	rule := m.config.GetRuleForRequest(path, method)

	info := make(map[string]interface{})
	info["rule"] = rule

	bucketInfo, err := m.tokenBucket.GetBucketInfo(rule, identifier)
	if err != nil {
		return nil, err
	}

	info["bucket"] = bucketInfo

	return info, nil
}

// ResetRateLimit 重置速率限制（用于管理）
func (m *RateLimitMiddleware) ResetRateLimit(c *gin.Context, path string, method string, identifier string) error {
	rule := m.config.GetRuleForRequest(path, method)
	return m.tokenBucket.ResetBucket(rule, identifier)
}

// UpdateRule 更新速率限制规则（用于动态配置）
func (m *RateLimitMiddleware) UpdateRule(ruleName string, newRule RateLimitRule) error {
	for i, rule := range m.config.Rules {
		if rule.Name == ruleName {
			m.config.Rules[i] = newRule
			return nil
		}
	}

	return fmt.Errorf("规则 '%s' 不存在", ruleName)
}

// AddRule 添加新的速率限制规则
func (m *RateLimitMiddleware) AddRule(newRule RateLimitRule) {
	m.config.Rules = append(m.config.Rules, newRule)
}

// RemoveRule 移除速率限制规则
func (m *RateLimitMiddleware) RemoveRule(ruleName string) error {
	for i, rule := range m.config.Rules {
		if rule.Name == ruleName {
			m.config.Rules = append(m.config.Rules[:i], m.config.Rules[i+1:]...)
			return nil
		}
	}

	return fmt.Errorf("规则 '%s' 不存在", ruleName)
}

// Enable 启用速率限制中间件
func (m *RateLimitMiddleware) Enable() {
	m.enabled = true
	fmt.Println("速率限制中间件已启用")
}

// Disable 禁用速率限制中间件
func (m *RateLimitMiddleware) Disable() {
	m.enabled = false
	fmt.Println("速率限制中间件已禁用")
}

// IsEnabled 检查中间件是否启用
func (m *RateLimitMiddleware) IsEnabled() bool {
	return m.enabled
}

// Toggle 切换中间件启用状态
func (m *RateLimitMiddleware) Toggle() bool {
	m.enabled = !m.enabled
	status := "启用"
	if !m.enabled {
		status = "禁用"
	}
	fmt.Printf("速率限制中间件已%s\n", status)
	return m.enabled
}
