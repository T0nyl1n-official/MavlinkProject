package Database

import (
	"context"
	"fmt"

	redis "github.com/redis/go-redis/v9"

	Config "MavlinkProject/Server/backend/Database/Config"
	Verification "MavlinkProject/Server/backend/Utils/Verification"
)

// InitRedis 初始化Redis客户端; 一个redisClient管理一个db
func InitRedis(config *Config.RedisClientConfig) (*redis.Client, *Verification.VerificationManager) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       int(config.DB),
	})

	if config.DB == 15 {
		verificationConfig := Verification.VerificationConfig{}
		verificationConfig.Default()

		verification, veriErr := Verification.NewVerificationManager(&verificationConfig)
		if veriErr != nil {

			return nil, nil
		}

		// 测试连接
		_, err := client.Ping(context.Background()).Result()
		if err != nil {
			return nil, nil
		}

		fmt.Println("DB - RedisClient: Redis初始化成功")
		return client, verification
	}
	return client, nil
}
