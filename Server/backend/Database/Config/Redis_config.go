package DBConfig

type RedisDB_allocate int

/*
	RedisDB_allocate 是 Redis 数据库的DB分配
	对应的DB编号根据const常量设置参见

	GeneralWarning 是通用警告/未知警告/未分配警告&日志
	Backend 是后端警告
	Frontend 是前端警告
	Agent 是AI/代理警告/日志
	Drone 是无人机警告/日志
	Sensor 是传感器警告/日志
	Verification 是验证数据库 (非错误)
*/
const (
	GeneralWarning RedisDB_allocate = iota
	Backend
	Frontend
	Agent
	Drone
	Sensor
	Verification = 15
)

type RedisClientConfig struct {
	Network string
	Addr    string
	ClientName string
	Password string
	DB       RedisDB_allocate
}

func (cfg *RedisClientConfig) RedisConfig_Default(db RedisDB_allocate) *RedisClientConfig {
	return &RedisClientConfig{
		Network: "tcp",
		Addr:    "localhost:6379",
		ClientName: "MavlinkProject",
		Password: "",
		DB:       db,
	}
}

