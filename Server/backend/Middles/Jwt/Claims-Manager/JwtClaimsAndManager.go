package Claims

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"

	Config "MavlinkProject/Server/backend/Middles/Jwt/Config"
)

// JWTClaims JWT声明
type JWTClaims struct {
	jwt.RegisteredClaims
	UserID    uint   `json:"user_id"`
	UserEmail string `json:"email"`
	Username  string `json:"username"`
	Role      string `json:"role"`
}

// JWTManager JWT管理器
type JWTManager struct {
	Config Config.JWTConfig
}

// GenerateToken 生成JWT Token
func (m *JWTManager) GenerateToken(userID uint, username, role string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,

		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.Config.ExpireTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "MavlinkProject",
			Subject:   "authentication",
		},
	}

	token := jwt.NewWithClaims(m.Config.SigningMethod, claims)
	tokenString, err := token.SignedString(m.Config.SecretKey)
	if err != nil {
		return "", fmt.Errorf("生成Token失败: %v", err)
	}

	return tokenString, nil
}

// GenerateRefreshToken 生成Refresh Token
func (m *JWTManager) GenerateRefreshToken(userID uint, username, role string) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.Config.RefreshTime)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "MavlinkProject",
			Subject:   "refresh",
		},
	}

	token := jwt.NewWithClaims(m.Config.SigningMethod, claims)
	tokenString, err := token.SignedString(m.Config.SecretKey)
	if err != nil {
		return "", fmt.Errorf("生成Refresh Token失败: %v", err)
	}

	return tokenString, nil
}

// ParseToken 解析并验证Token
func (m *JWTManager) ParseToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("不支持的签名方法: %v", token.Header["alg"])
		}
		return m.Config.SecretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("解析Token失败: %v", err)
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的Token")
}

// ValidateToken 验证Token是否有效
func (m *JWTManager) ValidateToken(tokenString string) (bool, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return false, err
	}

	// 检查Token是否过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return false, errors.New("Token已过期")
	}

	return true, nil
}

// RefreshToken 刷新Token
func (m *JWTManager) RefreshToken(refreshToken string) (string, string, error) {
	// 验证Refresh Token
	claims, err := m.ParseToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("refresh Token验证失败: %v", err)
	}

	// 检查Refresh Token是否过期
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return "", "", errors.New("refresh Token已过期")
	}

	// 生成新的Access Token
	newAccessToken, err := m.GenerateToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		return "", "", fmt.Errorf("生成新Access Token失败: %v", err)
	}

	// 生成新的Refresh Token
	newRefreshToken, err := m.GenerateRefreshToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		return "", "", fmt.Errorf("生成新Refresh Token失败: %v", err)
	}

	return newAccessToken, newRefreshToken, nil
}

// GetUserInfoFromToken 从Token中获取用户信息
func (m *JWTManager) GetUserInfoFromToken(tokenString string) (uint, string, string, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return 0, "", "", err
	}

	return claims.UserID, claims.Username, claims.Role, nil
}

// IsTokenExpired 检查Token是否过期
func (m *JWTManager) IsTokenExpired(tokenString string) (bool, error) {
	claims, err := m.ParseToken(tokenString)
	if err != nil {
		return false, err
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return true, nil
	}

	return false, nil
}

// NewJWTManager 创建新的JWT管理器
func NewJWTManager(config Config.JWTConfig) *JWTManager {
	return &JWTManager{
		Config: config,
	}
}
