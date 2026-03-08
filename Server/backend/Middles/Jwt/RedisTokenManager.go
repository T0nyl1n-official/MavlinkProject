package Jwt

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// TokenInfo 存储在Redis中的Token信息
type TokenInfo struct {
	Token      string    `json:"token"`
	UserID     uint      `json:"user_id"`
	Username   string    `json:"username"`
	Role       string    `json:"role"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
	IsActive   bool      `json:"is_active"`
}

// RedisTokenManager Redis Token管理器
type RedisTokenManager struct {
	redisClient *redis.Client
	ctx         context.Context
}

// NewRedisTokenManager 创建新的Redis Token管理器
func NewRedisTokenManager(redisClient *redis.Client) *RedisTokenManager {
	return &RedisTokenManager{
		redisClient: redisClient,
		ctx:         context.Background(),
	}
}

// StoreToken 存储Token到Redis
func (rtm *RedisTokenManager) StoreToken(token string, userID uint, username, role string, expiresAt time.Time) error {
	tokenInfo := TokenInfo{
		Token:      token,
		UserID:     userID,
		Username:   username,
		Role:       role,
		ExpiresAt:  expiresAt,
		CreatedAt:  time.Now(),
		IsActive:   true,
	}

	data, err := json.Marshal(tokenInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal token info: %v", err)
	}

	// 设置过期时间
	expiration := time.Until(expiresAt)
	if expiration <= 0 {
		expiration = time.Hour // 默认1小时
	}

	key := rtm.getTokenKey(token)
	err = rtm.redisClient.Set(rtm.ctx, key, data, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to store token in Redis: %v", err)
	}

	// 同时存储用户ID到Token的映射
	userTokenKey := rtm.getUserTokenKey(userID)
	err = rtm.redisClient.Set(rtm.ctx, userTokenKey, token, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to store user-token mapping: %v", err)
	}

	return nil
}

// GetToken 从Redis获取Token信息
func (rtm *RedisTokenManager) GetToken(token string) (*TokenInfo, error) {
	key := rtm.getTokenKey(token)
	data, err := rtm.redisClient.Get(rtm.ctx, key).Bytes()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("token not found")
		}
		return nil, fmt.Errorf("failed to get token from Redis: %v", err)
	}

	var tokenInfo TokenInfo
	err = json.Unmarshal(data, &tokenInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal token info: %v", err)
	}

	return &tokenInfo, nil
}

// ValidateToken 验证Token是否有效
func (rtm *RedisTokenManager) ValidateToken(token string) (bool, error) {
	tokenInfo, err := rtm.GetToken(token)
	if err != nil {
		return false, err
	}

	// 检查Token是否过期
	if time.Now().After(tokenInfo.ExpiresAt) {
		return false, fmt.Errorf("token expired")
	}

	// 检查Token是否被禁用
	if !tokenInfo.IsActive {
		return false, fmt.Errorf("token is not active")
	}

	return true, nil
}

// RevokeToken 撤销Token（软删除）
func (rtm *RedisTokenManager) RevokeToken(token string) error {
	tokenInfo, err := rtm.GetToken(token)
	if err != nil {
		return err
	}

	tokenInfo.IsActive = false
	data, err := json.Marshal(tokenInfo)
	if err != nil {
		return fmt.Errorf("failed to marshal updated token info: %v", err)
	}

	key := rtm.getTokenKey(token)
	// 设置较短的过期时间，让Redis自动清理
	err = rtm.redisClient.Set(rtm.ctx, key, data, 5*time.Minute).Err()
	if err != nil {
		return fmt.Errorf("failed to update token in Redis: %v", err)
	}

	return nil
}

// RevokeAllUserTokens 撤销用户的所有Token
func (rtm *RedisTokenManager) RevokeAllUserTokens(userID uint) error {
	// 获取用户的当前Token
	userTokenKey := rtm.getUserTokenKey(userID)
	token, err := rtm.redisClient.Get(rtm.ctx, userTokenKey).Result()
	if err != nil && err != redis.Nil {
		return fmt.Errorf("failed to get user token: %v", err)
	}

	// 如果用户有Token，则撤销它
	if token != "" {
		err = rtm.RevokeToken(token)
		if err != nil {
			return fmt.Errorf("failed to revoke user token: %v", err)
		}
	}

	// 删除用户Token映射
	err = rtm.redisClient.Del(rtm.ctx, userTokenKey).Err()
	if err != nil {
		return fmt.Errorf("failed to delete user token mapping: %v", err)
	}

	return nil
}

// GetUserByToken 通过Token获取用户信息
func (rtm *RedisTokenManager) GetUserByToken(token string) (uint, string, string, error) {
	tokenInfo, err := rtm.GetToken(token)
	if err != nil {
		return 0, "", "", err
	}

	return tokenInfo.UserID, tokenInfo.Username, tokenInfo.Role, nil
}

// CleanExpiredTokens 清理过期的Token
func (rtm *RedisTokenManager) CleanExpiredTokens() error {
	// Redis会自动清理过期的key，这里可以添加额外的清理逻辑
	return nil
}

// getTokenKey 获取Token的Redis键名
func (rtm *RedisTokenManager) getTokenKey(token string) string {
	return fmt.Sprintf("jwt:token:%s", token)
}

// getUserTokenKey 获取用户Token映射的Redis键名
func (rtm *RedisTokenManager) getUserTokenKey(userID uint) string {
	return fmt.Sprintf("jwt:user:%d:token", userID)
}