package Config

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// JWT_AUTH总配置
type Config struct {
	JWT    JWTConfig
	Server ServerConfig
}

// JWTConfig 定义JWT配置
type JWTConfig struct {
	SecretKey     []byte
	ExpireTime    time.Duration
	RefreshTime   time.Duration
	SigningMethod jwt.SigningMethod
	IdentityKey   string
}

// ServerConfig 定义服务器配置
type ServerConfig struct {
	Port string
}

// DefaultJWTConfig 默认JWT配置 (管理员设置)
var DefaultJWTConfig = JWTConfig{
	SecretKey:     []byte(os.Getenv("MavlinkProject_JWT_SECRET_KEY")),
	ExpireTime:    time.Hour,
	RefreshTime:   time.Hour * 24,
	SigningMethod: jwt.SigningMethodHS256,
	IdentityKey:   "user_id",
}

func Load() *Config {
	return &Config{
		JWT: JWTConfig{
			SecretKey:     DefaultJWTConfig.SecretKey,
			ExpireTime:    DefaultJWTConfig.ExpireTime,
			RefreshTime:   DefaultJWTConfig.RefreshTime,
			SigningMethod: DefaultJWTConfig.SigningMethod,
			IdentityKey:   DefaultJWTConfig.IdentityKey,
		},
		Server: ServerConfig{
			Port: os.Getenv("MavlinkProject_JWT_SERVER_PORT"),
		},
	}
}
