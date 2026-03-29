package Backend

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	WarningHandler "MavlinkProject/Server/Backend/Utils/WarningHandle"
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
		redisClient, verification := DBService.InitRedis(redisConfig)
		if redisClient == nil {
			log.Fatalf("MavlinkProject - Backend : 初始化Redis失败: DB=%d", db)
		}
		redisClients = append(redisClients, *redisClient)

		if db == DBConfig.Verification && verification != nil {
			verification = verification
		}
	}

	tokenRedis := redisClients[len(redisClients)-2]
	verificationRedis := redisClients[len(redisClients)-1]

	jwtManager := Middleware.NewDefaultJWTManager()
	tokenManager := Jwt.NewRedisTokenManager(&tokenRedis)

	router.Use(Listening.ListeningErrorMiddleWare(),
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
	Routes.InitAllRoutes(bs.Router, bs.JWTManager, bs.TokenManager, bs.Mysql)

	certPath := "cert.pem"
	keyPath := "key.pem"

	certExists := false
	if _, err := os.Stat(certPath); err == nil {
		if _, err := os.Stat(keyPath); err == nil {
			certExists = true
		}
	}

	if certExists {
		log.Printf("检测到证书文件，同时启动 HTTP 和 HTTPS 服务器")

		httpAddr := ":8080"
		httpsAddr := ":443"

		go func() {
			log.Printf("启动 HTTP 服务器: %s", httpAddr)
			err := bs.Router.Run(httpAddr)
			if err != nil {
				log.Printf("HTTP 服务器启动失败: %v", err)
			}
		}()

		srv := &http.Server{
			Addr: httpsAddr,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				MaxVersion: tls.VersionTLS13,
			},
		}

		log.Printf("启动 HTTPS 服务器: %s", httpsAddr)
		err := srv.ListenAndServeTLS(certPath, keyPath)
		if err != nil {
			log.Printf("HTTPS 启动失败: %v", err)
		} else {
			log.Printf("Backend server started on port %s (HTTP) and %s (HTTPS)", httpAddr, httpsAddr)
		}
	} else {
		httpAddr := ":" + port
		log.Printf("未检测到证书文件，启动 HTTP 服务器: %s", httpAddr)
		err := bs.Router.Run(httpAddr)
		if err != nil {
			log.Printf("HTTP 启动失败: %v", err)
		} else {
			log.Printf("Backend server started on port %s (HTTP)", port)
		}
	}
}

func (bs *BackendServer) Start(addr, port string) *BackendServer {
	bs.New()
	bs.Run(port)
	log.Printf("Backend server starting on port %s", port)
	return bs
}
