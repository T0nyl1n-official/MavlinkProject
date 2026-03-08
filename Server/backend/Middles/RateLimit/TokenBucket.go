package RateLimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenBucket 令牌桶结构
type TokenBucket struct {
	redisClient *redis.Client
	config      *RateLimitConfig
}

// NewTokenBucket 创建令牌桶实例
func NewTokenBucket(redisClient *redis.Client, config *RateLimitConfig) *TokenBucket {
	return &TokenBucket{
		redisClient: redisClient,
		config:      config,
	}
}

// AllowRequest 检查是否允许请求
func (tb *TokenBucket) AllowRequest(rule *RateLimitRule, identifier string) (bool, int64, error) {
	ctx := context.Background()
	bucketKey := tb.config.GetBucketKey(rule, identifier)

	// 获取当前时间戳
	now := time.Now().Unix()

	// 使用Lua脚本实现原子操作
	luaScript := `
local bucket_key = KEYS[1]
local now = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])
local fill_rate = tonumber(ARGV[3])
local window_seconds = tonumber(ARGV[4])

-- 获取当前桶状态
local bucket_data = redis.call('HMGET', bucket_key, 'tokens', 'last_refill_time')
local current_tokens = tonumber(bucket_data[1]) or capacity
local last_refill_time = tonumber(bucket_data[2]) or now

-- 计算应该填充的令牌数
local time_passed = now - last_refill_time
if time_passed > 0 then
    local tokens_to_add = math.floor(time_passed * fill_rate)
    current_tokens = math.min(capacity, current_tokens + tokens_to_add)
    last_refill_time = now
end

-- 检查是否有足够的令牌
if current_tokens >= 1 then
    -- 消耗一个令牌
    current_tokens = current_tokens - 1
    
    -- 更新桶状态
    redis.call('HMSET', bucket_key, 
        'tokens', current_tokens,
        'last_refill_time', last_refill_time
    )
    
    -- 设置过期时间
    redis.call('EXPIRE', bucket_key, window_seconds * 2)
    
    return {1, current_tokens}
else
    return {0, current_tokens}
end
`

	result, err := tb.redisClient.Eval(ctx, luaScript, []string{bucketKey}, 
		now, rule.Capacity, rule.FillRate, rule.WindowSeconds).Result()
	
	if err != nil {
		return false, 0, fmt.Errorf("执行令牌桶脚本失败: %v", err)
	}

	// 解析结果
	results, ok := result.([]interface{})
	if !ok || len(results) != 2 {
		return false, 0, fmt.Errorf("无效的令牌桶结果")
	}

	allowed := results[0].(int64) == 1
	remainingTokens := results[1].(int64)

	return allowed, remainingTokens, nil
}

// GetRemainingTokens 获取剩余令牌数
func (tb *TokenBucket) GetRemainingTokens(rule *RateLimitRule, identifier string) (int64, error) {
	ctx := context.Background()
	bucketKey := tb.config.GetBucketKey(rule, identifier)

	// 获取当前桶状态
	bucketData, err := tb.redisClient.HGetAll(ctx, bucketKey).Result()
	if err != nil {
		return 0, fmt.Errorf("获取令牌桶状态失败: %v", err)
	}

	if len(bucketData) == 0 {
		return rule.Capacity, nil
	}

	// 计算当前令牌数
	currentTokens, ok := bucketData["tokens"]
	if !ok {
		return rule.Capacity, nil
	}

	tokens, err := tb.parseTokenCount(currentTokens)
	if err != nil {
		return 0, err
	}

	// 如果需要，重新计算令牌数
	lastRefillTimeStr, ok := bucketData["last_refill_time"]
	if ok {
		lastRefillTime, err := tb.parseTimestamp(lastRefillTimeStr)
		if err == nil {
			now := time.Now().Unix()
			timePassed := now - lastRefillTime
			
			if timePassed > 0 {
				tokensToAdd := int64(float64(timePassed) * rule.FillRate)
				tokens = min(rule.Capacity, tokens+int64(tokensToAdd))
			}
		}
	}

	return tokens, nil
}

// ResetBucket 重置令牌桶
func (tb *TokenBucket) ResetBucket(rule *RateLimitRule, identifier string) error {
	ctx := context.Background()
	bucketKey := tb.config.GetBucketKey(rule, identifier)

	_, err := tb.redisClient.Del(ctx, bucketKey).Result()
	if err != nil {
		return fmt.Errorf("重置令牌桶失败: %v", err)
	}

	return nil
}

// GetBucketInfo 获取令牌桶信息
func (tb *TokenBucket) GetBucketInfo(rule *RateLimitRule, identifier string) (map[string]interface{}, error) {
	ctx := context.Background()
	bucketKey := tb.config.GetBucketKey(rule, identifier)

	info := make(map[string]interface{})
	
	// 获取桶数据
	bucketData, err := tb.redisClient.HGetAll(ctx, bucketKey).Result()
	if err != nil {
		return nil, fmt.Errorf("获取令牌桶信息失败: %v", err)
	}

	if len(bucketData) == 0 {
		info["tokens"] = rule.Capacity
		info["last_refill_time"] = time.Now().Unix()
		info["is_new"] = true
	} else {
		info["tokens"], _ = tb.parseTokenCount(bucketData["tokens"])
		info["last_refill_time"], _ = tb.parseTimestamp(bucketData["last_refill_time"])
		info["is_new"] = false
	}

	info["capacity"] = rule.Capacity
	info["fill_rate"] = rule.FillRate
	info["window_seconds"] = rule.WindowSeconds
	info["bucket_key"] = bucketKey

	return info, nil
}

// 辅助函数
func (tb *TokenBucket) parseTokenCount(tokenStr string) (int64, error) {
	if tokenStr == "" {
		return 0, nil
	}
	
	var tokens int64
	_, err := fmt.Sscanf(tokenStr, "%d", &tokens)
	return tokens, err
}

func (tb *TokenBucket) parseTimestamp(timestampStr string) (int64, error) {
	if timestampStr == "" {
		return 0, nil
	}
	
	var timestamp int64
	_, err := fmt.Sscanf(timestampStr, "%d", &timestamp)
	return timestamp, err
}

func min(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}