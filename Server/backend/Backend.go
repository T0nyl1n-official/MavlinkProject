package Backend

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	Conf "MavlinkProject/Server/backend/Config"
	DBService "MavlinkProject/Server/backend/Database"
	DBConfig "MavlinkProject/Server/backend/Database/Config"
	UsersHandler "MavlinkProject/Server/backend/Handler/Users"
	Middleware "MavlinkProject/Server/backend/Middles"
	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
	Listening "MavlinkProject/Server/backend/Middles/Listening"
	Routes "MavlinkProject/Server/backend/Routes"
	FRPHelper "MavlinkProject/Server/backend/Utils/FRPHelper"
	Verification "MavlinkProject/Server/backend/Utils/Verification"
	WarningHandler "MavlinkProject/Server/backend/Utils/WarningHandle"
)

var (
	settingPath         = "config/Setting.yaml"
	backendServer       *BackendServer
	GlobalBackendServer *BackendServer
)

func GetBackendServer() *BackendServer {
	if backendServer != nil {
		GlobalBackendServer = backendServer
	}
	return backendServer
}

func GetGlobalBackendServer() *BackendServer {
	return GlobalBackendServer
}

func GetCentralClient() *FRPHelper.CentralHTTPClient {
	return FRPHelper.GetCentralClient()
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
	StartTime         time.Time
}

func (bs *BackendServer) New() {
	router := gin.Default()

	bs.SettingManager = Conf.GetSettingManager()
	err := bs.SettingManager.LoadSetting(settingPath)
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

var httpServer *http.Server

type HTTPSConfig struct {
	Enabled  bool   `yaml:"enabled"`
	Port     string `yaml:"port"`
	CertFile string `yaml:"cert_file"`
	KeyFile  string `yaml:"key_file"`
}

func (bs *BackendServer) Run(port string, httpsConfig HTTPSConfig) {
	bs.StartTime = time.Now()
	Routes.InitAllRoutes(bs.Router, bs.JWTManager, bs.TokenManager, bs.Mysql, bs.SettingManager, bs)

	if httpsConfig.Enabled {
		addr := "0.0.0.0:" + httpsConfig.Port
		log.Printf("启动 HTTPS 服务器: %s", addr)

		httpServer = &http.Server{
			Addr:    addr,
			Handler: bs.Router,
		}

		go func() {
			if err := httpServer.ListenAndServeTLS(httpsConfig.CertFile, httpsConfig.KeyFile); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTPS 服务器启动失败: %v", err)
			}
		}()

		log.Printf("Backend server started on port %s (HTTPS)", httpsConfig.Port)
	} else {
		addr := "0.0.0.0:" + port
		log.Printf("启动 HTTP 服务器: %s", addr)

		httpServer = &http.Server{
			Addr:    addr,
			Handler: bs.Router,
		}

		go func() {
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("HTTP 服务器启动失败: %v", err)
			}
		}()

		log.Printf("Backend server started on port %s (HTTP)", port)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
}

func (bs *BackendServer) Shutdown() error {
	log.Printf("[BackendServer] Shutting down server...")
	Listening.StopBoardListener()
	if httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("[BackendServer] Server shutdown error: %v", err)
			return err
		}
	}
	log.Printf("[BackendServer] Server shutdown complete")
	return nil
}

func (bs *BackendServer) Restart() {
	bs.Shutdown()
	time.Sleep(2 * time.Second)
	execPath, _ := os.Executable()
	syscall.Exec(execPath, os.Args, os.Environ())
}

func (bs *BackendServer) Start(addr, port string, httpsConfig HTTPSConfig) *BackendServer {
	// run Redis-Server-Cli and Cloudflare Tunnel in background process
	go func() {
		redisDefaultEXE := exec.Command("redis-server.exe")
		if err := redisDefaultEXE.Start(); err != nil {
			log.Printf("[MavlinkProject/Backend] Failed to start redis: %v", err)
		}
	}()

	go func() {
		tunnelCMD := exec.Command("cloudflared", "tunnel", "run")
		outs, err := tunnelCMD.Output()
		if err != nil {
			log.Printf("[MavlinkProject/Backend] Failed to start tunnel: %v, output: %v", err, outs)
		}
		log.Printf("[MavlinkProject/Backend] Cloudflare Tunnel started")
	}()

	// start BackendServer-GIN
	backendServer = bs
	bs.New()

	// 初始化 CentralBoard HTTP 客户端
	FRPHelper.InitCentralClient()
	log.Printf("[Backend] CentralBoard HTTP client initialized")

	bs.Run(port, httpsConfig)
	log.Printf("Backend server starting...")
	return bs
}

func (bs *BackendServer) onBoardConfigChange(newSetting *Conf.Setting) error {
	log.Printf("[Setting] Board config changed, restarting board listener...")
	Listening.StopBoardListener()
	Listening.StartBoardListener()
	log.Printf("[Setting] Board listener restarted")
	return nil
}
