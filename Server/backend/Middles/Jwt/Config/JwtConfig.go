package Config

import (
	"time"

	"github.com/golang-jwt/jwt/v4"

	Conf "MavlinkProject/Server/backend/Config"
)

type JWTConfig struct {
	SecretKey     []byte
	ExpireTime    time.Duration
	RefreshTime   time.Duration
	SigningMethod jwt.SigningMethod
	IdentityKey   string
}

func LoadJWTConfig() *JWTConfig {
	setting := Conf.GetSetting()
	jwtCfg := setting.JWT

	return &JWTConfig{
		SecretKey:     []byte(jwtCfg.SecretKey),
		ExpireTime:    time.Duration(jwtCfg.ExpireTime) * time.Second,
		RefreshTime:   time.Duration(jwtCfg.RefreshTime) * time.Second,
		SigningMethod: jwt.SigningMethodHS256,
		IdentityKey:   jwtCfg.IdentityKey,
	}
}
