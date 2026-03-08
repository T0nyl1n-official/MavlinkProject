package RateLimit

// RateLimitRule 速率限制规则
type RateLimitRule struct {
	// 规则名称
	Name string `json:"name"`

	// 接口路径模式（支持通配符）
	PathPattern string `json:"path_pattern"`

	// HTTP方法（GET, POST, PUT, DELETE等，空表示所有方法）
	Method string `json:"method"`

	// 限制类型："ip" - IP级别, "user" - 用户级别, "global" - 全局级别
	LimitType string `json:"limit_type"`

	// 令牌桶容量
	Capacity int64 `json:"capacity"`

	// 令牌填充速率（每秒填充的令牌数）
	FillRate float64 `json:"fill_rate"`

	// 限制时间窗口（秒）
	WindowSeconds int64 `json:"window_seconds"`

	// 是否启用 (默认: true)
	Enabled bool `json:"enabled"`

	// 优先级（数字越小优先级越高）
	Priority int `json:"priority"`
}

// RateLimitConfig 速率限制配置
type RateLimitConfig struct {
	// 默认规则（当没有匹配的规则时使用）
	DefaultRule RateLimitRule `json:"default_rule"`
	// 自定义规则列表
	Rules []RateLimitRule `json:"rules"`
	// Redis前缀
	RedisPrefix string `json:"redis_prefix"`
	// 是否启用全局速率限制
	EnableGlobalLimit bool `json:"enable_global_limit"`
	// 全局限制规则
	GlobalRule RateLimitRule `json:"global_rule"`
}

// DefaultRateLimitConfig 默认速率限制配置
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RedisPrefix:       "rate_limit:",
		EnableGlobalLimit: true,
		DefaultRule: RateLimitRule{
			Name:          "default",
			PathPattern:   "*",
			Method:        "",
			LimitType:     "ip",
			Capacity:      10000,
			FillRate:      400.0,
			WindowSeconds: 60,
			Enabled:       true,
			Priority:      100,
		},

		GlobalRule: RateLimitRule{
			Name:          "global",
			PathPattern:   "*",
			Method:        "",
			LimitType:     "global",
			Capacity:      20000,
			FillRate:      500.0,
			WindowSeconds: 60,
			Enabled:       true,
			Priority:      1,
		},

		Rules: []RateLimitRule{
			{
				Name:          "auth_login",
				PathPattern:   "/user/login",
				Method:        "POST",
				LimitType:     "ip",
				Capacity:      10,
				FillRate:      1.0,
				WindowSeconds: 60,
				Enabled:       true,
				Priority:      10,
			},
			{
				Name:          "auth_register",
				PathPattern:   "/user/register",
				Method:        "POST",
				LimitType:     "ip",
				Capacity:      10,
				FillRate:      1,
				WindowSeconds: 600,
				Enabled:       true,
				Priority:      10,
			},
			{
				Name:          "chain_operations",
				PathPattern:   "/chain/*",
				Method:        "",
				LimitType:     "user",
				Capacity:      300,
				FillRate:      5.0,
				WindowSeconds: 60,
				Enabled:       true,
				Priority:      30,
			},
			{
				Name:          "avatar_upload",
				PathPattern:   "/user/avatars/*",
				Method:        "POST",
				LimitType:     "user",
				Capacity:      20,
				FillRate:      1.0,
				WindowSeconds: 120,
				Enabled:       true,
				Priority:      20,
			},
		},
	}
}

// GetRuleForRequest 根据请求信息获取匹配的速率限制规则
func (config *RateLimitConfig) GetRuleForRequest(path string, method string) *RateLimitRule {
	// 按优先级排序规则
	rules := make([]RateLimitRule, len(config.Rules))
	copy(rules, config.Rules)

	// 添加全局规则
	if config.EnableGlobalLimit {
		rules = append(rules, config.GlobalRule)
	}

	// 按优先级排序（数字越小优先级越高）
	for i := 0; i < len(rules)-1; i++ {
		for j := i + 1; j < len(rules); j++ {
			if rules[i].Priority > rules[j].Priority {
				rules[i], rules[j] = rules[j], rules[i]
			}
		}
	}

	// 查找匹配的规则
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		// 检查方法匹配
		if rule.Method != "" && rule.Method != method {
			continue
		}

		// 检查路径匹配
		if rule.PathPattern == "*" || rule.PathPattern == path {
			return &rule
		}

		// 支持简单的通配符匹配
		if rule.PathPattern[len(rule.PathPattern)-1] == '*' {
			prefix := rule.PathPattern[:len(rule.PathPattern)-1]
			if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
				return &rule
			}
		}
	}

	// 返回默认规则
	return &config.DefaultRule
}

// GetBucketKey 获取令牌桶的Redis键
func (config *RateLimitConfig) GetBucketKey(rule *RateLimitRule, identifier string) string {
	key := config.RedisPrefix + rule.LimitType + ":" + rule.Name
	if rule.LimitType != "global" {
		key += ":" + identifier
	}
	return key
}
