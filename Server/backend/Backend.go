package Backend

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	WarningHandler "MavlinkProject/Server/backend/Utils/WarningHandle"
	Conf "MavlinkProject/Server/backend/Config"
	DBService "MavlinkProject/Server/backend/Database"
	DBConfig "MavlinkProject/Server/backend/Database/Config"
	UsersHandler "MavlinkProject/Server/backend/Handler/Users"
	Middleware "MavlinkProject/Server/backend/Middles"
	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
	Listening "MavlinkProject/Server/backend/Middles/Listening"
	Routes "MavlinkProject/Server/backend/Routes"
	Verification "MavlinkProject/Server/backend/Utils/Verification"
)

func (bs *BackendServer) onBoardConfigChange(newSetting *Conf.Setting) error {
	log.Printf("[Setting] Board config changed, restarting board listener...")
	Listening.StopBoardListener()
	Listening.StartBoardListener()
	log.Printf("[Setting] Board listener restarted")
	return nil
}

type BackendServer struct {
	Router            *gin.Engine
	Mysql             *gorm.DB
	RedisClient       *[]redis.Client
	VerificationRedis *redis.Client
	JWTManager        *jwtUtils.JWTManager
	TokenManager      *Jwt.RedisTokenManager
	Verification      Verification.VerificationManager
	SettingManager    *Conf.SettingManager
}

func (bs *BackendServer) New() {
	router := gin.Default()

	bs.SettingManager = Conf.GetSettingManager()
	err := bs.SettingManager.LoadSetting("config/Setting.yaml")
	if err != nil {
		log.Fatalf("MavlinkProject - Backend : 加载配置文件失败 : %v", err)
	}

	bs.SettingManager.RegisterChangeCallback("board", bs.onBoardConfigChange)

	redisClients := make([]redis.Client, 0)
	verification := Verification.VerificationManager{}
	redisDB := []DBConfig.RedisDB_allocate{
		DBConfig.GeneralWarning,
		DBConfig.Backend,
		DBConfig.Token,
		DBConfig.Verification,
	}

	mysqlDB, err := DBService.InitMysql()
	if err != nil {
		log.Fatalf("MavlinkProject - Backend : 初始化Mysql失败 : %v", err)
	}

	for _, db := range redisDB {
		redisConfig := &DBConfig.RedisClientConfig{}
		redisConfig = redisConfig.RedisConfig_Default(db)
		redisClient, _ := DBService.InitRedis(redisConfig)
		if redisClient == nil {
			log.Fatalf("MavlinkProject - Backend : 初始化Redis失败: DB=%d", db)
		}
		redisClients = append(redisClients, *redisClient)

	}

	tokenRedis := redisClients[len(redisClients)-2]
	verificationRedis := redisClients[len(redisClients)-1]

	jwtManager := Middleware.NewDefaultJWTManager()
	tokenManager := Jwt.NewRedisTokenManager(&tokenRedis)

	router.Use(
		Middleware.BanishCheck(),
		Listening.ListeningErrorMiddleWare(),
		Listening.BoardListenerMiddleware(),
		Middleware.Logger(mysqlDB),
	)

	bs.Router = router
	bs.Mysql = mysqlDB
	bs.RedisClient = &redisClients
	bs.VerificationRedis = &verificationRedis
	bs.JWTManager = jwtManager
	bs.TokenManager = tokenManager
	bs.Verification = verification

	UsersHandler.SetVerification(verification)
	UsersHandler.SetJWTManager(jwtManager)

	WarningHandler.SetRedisClients(&redisClients)

	Listening.StartBoardListener()
	log.Printf("[BackendServer] Board listener service started")
}

func (bs *BackendServer) Run(port string) {

	Routes.InitAllRoutes(bs.Router, bs.JWTManager, bs.TokenManager, bs.Mysql, bs.SettingManager)

	addr := "0.0.0.0:" + port
	log.Printf("启动 HTTP 服务器: %s", addr)

	err := bs.Router.RunTLS("0.0.0.0:8080", "cert.pem", "key.pem")
	if err != nil {
		log.Printf("HTTP 服务器启动失败: %v", err)
	} else {
		log.Printf("Backend server started on port %s (HTTP)", port)
	}
}

func (bs *BackendServer) Start(addr, port string) *BackendServer {
	bs.New()
	bs.Run(port)
	log.Printf("Backend server starting on port %s (HTTP)", port)
	return bs
}
