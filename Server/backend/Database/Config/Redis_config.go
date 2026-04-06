package DBConfig

import (
	"fmt"

	Conf "MavlinkProject/Server/backend/Config"
)

type RedisDB_allocate int

const (
	GeneralWarning RedisDB_allocate = iota
	Backend
	Frontend
	Agent
	Drone
	Sensor
	Token        = 14
	Verification = 15
)

type RedisClientConfig struct {
	Network    string
	Addr       string
	ClientName string
	Password   string
	DB         RedisDB_allocate
}

func (cfg *RedisClientConfig) RedisConfig_Default(db RedisDB_allocate) *RedisClientConfig {
	setting := Conf.GetSetting()
	redisCfg := setting.Redis

	addr := fmt.Sprintf("%s:%s", redisCfg.Host, redisCfg.Port)

	return &RedisClientConfig{
		Network:    "tcp",
		Addr:       addr,
		ClientName: "MavlinkProject",
		Password:   redisCfg.Password,
		DB:         db,
	}
}
