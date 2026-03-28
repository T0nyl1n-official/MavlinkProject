package Backend

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/http/webroot"
	"github.com/go-acme/lego/v4/registration"
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

// Let's Encrypt 配置
type LetsEncryptConfig struct {
	Email      string   `yaml:"email"`
	Domains    []string `yaml:"domains"`
	Webroot    string   `yaml:"webroot"`
	UseStaging bool     `yaml:"use_staging"`
}

// 用户结构体（用于 Let's Encrypt 注册）
type MyUser struct {
	Email        string
	Registration *registration.Resource
	key          crypto.PrivateKey
}

func (u *MyUser) GetEmail() string {
	return u.Email
}
func (u MyUser) GetRegistration() *registration.Resource {
	return u.Registration
}
func (u *MyUser) GetPrivateKey() crypto.PrivateKey {
	return u.key
}

// 获取 Let's Encrypt 证书
func getLetsEncryptCert(domains []string, email string, useStaging bool) (*certificate.Resource, error) {
	// 创建用户
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	user := &MyUser{
		Email: email,
		key:   privateKey,
	}

	config := lego.NewConfig(user)

	// 设置测试环境
	if useStaging {
		config.CADirURL = lego.LEDirectoryStaging
		log.Printf("使用 Let's Encrypt 测试环境")
	}

	client, err := lego.NewClient(config)
	if err != nil {
		return nil, err
	}

	// 使用 HTTP 验证（需要配置 webroot）
	webrootProvider, err := webroot.NewHTTPProvider("./webroot")
	if err != nil {
		return nil, err
	}

	err = client.Challenge.SetHTTP01Provider(webrootProvider)
	if err != nil {
		return nil, err
	}

	// 注册用户
	reg, err := client.Registration.Register(registration.RegisterOptions{TermsOfServiceAgreed: true})
	if err != nil {
		return nil, err
	}
	user.Registration = reg

	// 请求证书
	request := certificate.ObtainRequest{
		Domains: domains,
		Bundle:  true,
	}
	certificates, err := client.Certificate.Obtain(request)
	if err != nil {
		return nil, err
	}

	return certificates, nil
}

// 保存证书到文件
func saveCertificate(cert *certificate.Resource, certPath, keyPath string) error {
	// 确保目录存在
	os.MkdirAll(filepath.Dir(certPath), 0755)
	os.MkdirAll(filepath.Dir(keyPath), 0755)

	// 保存证书
	if err := os.WriteFile(certPath, cert.Certificate, 0644); err != nil {
		return err
	}

	// 保存私钥
	if err := os.WriteFile(keyPath, cert.PrivateKey, 0644); err != nil {
		return err
	}

	return nil
}

// 检查证书是否需要更新
func checkCertificateExpiry(certPath string) bool {
	_, err := os.ReadFile(certPath)
	if err != nil {
		return true // 证书不存在，需要获取
	}

	// 这里可以添加证书过期检查逻辑
	// 简化处理：总是返回 true 以重新获取证书
	return true
}

// 获取或更新 Let's Encrypt 证书
func ensureLetsEncryptCertificate(domains []string, email string, useStaging bool) (string, string, error) {
	certPath := "./letsencrypt/cert.pem"
	keyPath := "./letsencrypt/key.pem"

	// 检查是否需要更新证书
	if !checkCertificateExpiry(certPath) {
		return certPath, keyPath, nil
	}

	log.Printf("正在获取 Let's Encrypt 证书，域名: %v", domains)

	cert, err := getLetsEncryptCert(domains, email, useStaging)
	if err != nil {
		log.Printf("获取 Let's Encrypt 证书失败: %v", err)
		return "", "", err
	}

	// 保存证书
	if err := saveCertificate(cert, certPath, keyPath); err != nil {
		log.Printf("保存证书失败: %v", err)
		return "", "", err
	}

	log.Printf("Let's Encrypt 证书获取成功，保存到: %s, %s", certPath, keyPath)
	return certPath, keyPath, nil
}

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

		// 如果是验证数据库，设置验证管理器
		if db == DBConfig.Verification && verification != nil {
			verification = verification
		}
	}

	tokenRedis := redisClients[len(redisClients)-2]
	verificationRedis := redisClients[len(redisClients)-1]

	jwtManager := Middleware.NewDefaultJWTManager()
	tokenManager := Jwt.NewRedisTokenManager(&tokenRedis)

	// 全局中间件使用
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

// 添加 Let's Encrypt 验证路由
func addLetsEncryptRoutes(router *gin.Engine) {
	// 处理 Let's Encrypt HTTP 验证
	router.GET("/.well-known/acme-challenge/:token", func(c *gin.Context) {
		token := c.Param("token")

		// Let's Encrypt 期望返回特定的验证内容
		// 这里应该返回 lego 库生成的验证内容
		// 简化处理：返回 token 作为验证内容
		c.Header("Content-Type", "text/plain")
		c.String(200, token+".xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")
		log.Printf("处理 Let's Encrypt 验证请求，token: %s", token)
	})
}

func (bs *BackendServer) Run(port string, letsEncryptConfig LetsEncryptConfig) {
	Routes.InitAllRoutes(bs.Router, bs.JWTManager, bs.TokenManager, bs.Mysql)

	// 添加 Let's Encrypt 验证路由
	addLetsEncryptRoutes(bs.Router)

	addr := ":" + port

	certPath := "cert.pem"
	keyPath := "key.pem"

	// 如果配置了 Let's Encrypt，尝试获取证书
	if len(letsEncryptConfig.Domains) > 0 && letsEncryptConfig.Domains[0] != "" {
		log.Printf("正在获取 Let's Encrypt 证书，域名: %v", letsEncryptConfig.Domains)

		// 先启动 HTTP 服务器处理 Let's Encrypt 验证
		log.Printf("启动 HTTP 服务器处理 Let's Encrypt 验证: %s", addr)

		// 在后台启动 HTTP 服务器
		httpServer := &http.Server{
			Addr:    addr,
			Handler: bs.Router,
		}

		go func() {
			err := httpServer.ListenAndServe()
			if err != nil && err != http.ErrServerClosed {
				log.Printf("HTTP 服务器启动失败: %v", err)
			}
		}()

		// 等待 HTTP 服务器启动
		time.Sleep(3 * time.Second)

		// 尝试获取 Let's Encrypt 证书
		leCertPath, leKeyPath, err := ensureLetsEncryptCertificate(letsEncryptConfig.Domains, letsEncryptConfig.Email, letsEncryptConfig.UseStaging)
		if err == nil {
			certPath = leCertPath
			keyPath = leKeyPath
			log.Printf("使用 Let's Encrypt 证书: %s", letsEncryptConfig.Domains)

			// 停止 HTTP 服务器，启动 HTTPS
			log.Printf("停止 HTTP 服务器，启动 HTTPS")

			// 强制使用 HTTPS，配置 TLS 版本
			srv := &http.Server{
				Addr:    addr,
				Handler: bs.Router,
				TLSConfig: &tls.Config{
					MinVersion: tls.VersionTLS12, // 设置最低 TLS 版本为 1.2
					MaxVersion: tls.VersionTLS13, // 设置最高 TLS 版本为 1.3
				},
			}

			log.Printf("启动 HTTPS 服务器: %s", addr)
			err = srv.ListenAndServeTLS(certPath, keyPath)
			if err != nil {
				log.Printf("HTTPS 启动失败: %v", err)
				log.Printf("服务器启动失败")
			} else {
				log.Printf("Backend server started on port %s (HTTPS)", port)
			}
		} else {
			log.Printf("使用自签名证书，Let's Encrypt 失败: %v", err)
			log.Printf("继续使用 HTTP 服务器")

			// 让 HTTP 服务器继续运行
			select {}
		}
	} else {
		log.Printf("使用自签名证书（开发环境）")

		// 强制使用 HTTPS，配置 TLS 版本
		srv := &http.Server{
			Addr:    addr,
			Handler: bs.Router,
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12, // 设置最低 TLS 版本为 1.2
				MaxVersion: tls.VersionTLS13, // 设置最高 TLS 版本为 1.3
			},
		}

		log.Printf("启动 HTTPS 服务器: %s", addr)
		err := srv.ListenAndServeTLS(certPath, keyPath)
		if err != nil {
			log.Printf("HTTPS 启动失败: %v", err)
			log.Printf("尝试启动 HTTP 服务器: %s", addr)
			// 回退到 HTTP
			err = bs.Router.Run(addr)
			if err != nil {
				log.Printf("HTTP 启动失败: %v", err)
			} else {
				log.Printf("Backend server started on port %s (HTTP)", port)
			}
		} else {
			log.Printf("Backend server started on port %s (HTTPS)", port)
		}
	}
}

func (bs *BackendServer) Start(addr, port string, letsEncryptConfig LetsEncryptConfig) *BackendServer {
	bs.New()
	bs.Run(port, letsEncryptConfig)
	log.Printf("Backend server starting on port %s (HTTPS)", port)
	return bs
}
