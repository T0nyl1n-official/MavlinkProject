package Backend

import (
	"log"

	gin "github.com/gin-gonic/gin"
	redis "github.com/redis/go-redis/v9"
	gorm "gorm.io/gorm"

	DBService "MavlinkProject/Server/backend/Database"
	DBConfig "MavlinkProject/Server/backend/Database/Config"
	UsersHandler "MavlinkProject/Server/backend/Handler/Users"
	JwtMiddleware "MavlinkProject/Server/backend/Middles"
	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
	Routes "MavlinkProject/Server/backend/Routes"
	Verification "MavlinkProject/Server/backend/Utils/Verification"
)

type BackendServer struct {
	Router            *gin.Engine
	Mysql             *gorm.DB
	RedisClient       *[]redis.Client
	VerificationRedis *redis.Client
	JWTManager        *jwtUtils.JWTManager
	TokenManager      *Jwt.RedisTokenManager
	Verification      Verification.VerificationManager
}

func (bs *BackendServer) New() {
	router := gin.Default()
	redisClients := make([]redis.Client, 0)
	verification := Verification.VerificationManager{}
	redisDB := []DBConfig.RedisDB_allocate{
		DBConfig.GeneralWarning,
		DBConfig.Backend,
		DBConfig.Frontend,
		DBConfig.Agent,
		DBConfig.Drone,
		DBConfig.Sensor,
		DBConfig.Token,
		DBConfig.Verification,
	}

	for i := range redisDB {
		config := DBConfig.RedisClientConfig{
			DB: redisDB[i],
		}
		config.RedisConfig_Default(redisDB[i])

		client, veri := DBService.InitRedis(&config)
		if veri != nil {
			verification = *veri
		}
		redisClients = append(redisClients, *client)
	}

	mysqlDB, err := DBService.InitMysql()
	if err != nil {
		log.Fatalf("MavlinkProject - Backend : 初始化Mysql失败 : %v", err)
	}

	tokenRedis := redisClients[len(redisClients)-2]
	verificationRedis := redisClients[len(redisClients)-1]

	jwtManager := JwtMiddleware.NewDefaultJWTManager()
	tokenManager := Jwt.NewRedisTokenManager(&tokenRedis)

	bs.Router = router
	bs.Mysql = mysqlDB
	bs.RedisClient = &redisClients
	bs.VerificationRedis = &verificationRedis
	bs.JWTManager = jwtManager
	bs.TokenManager = tokenManager
	bs.Verification = verification

	UsersHandler.SetVerification(verification)
	UsersHandler.SetJWTManager(jwtManager)
}

func (bs *BackendServer) Run(port string) {
	Routes.InitAllRoutes(bs.Router, bs.JWTManager, bs.TokenManager, bs.Mysql)

	bs.Router.Run(port)
	log.Printf("Backend server started on port %s", port)
}

// 被整合的Backend创建方法
func (bs *BackendServer) Start(port string) *BackendServer {
	bs.New()
	bs.Run(port)
	return bs
}
