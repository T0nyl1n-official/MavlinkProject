package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"

	api "MavlinkProject_Board/app/api"
	config "MavlinkProject_Board/app/config"
	backend "MavlinkProject_Board/app/services/backend"
	mavlink "MavlinkProject_Board/app/services/mavlink"
)

func main() {
	// 加载配置
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	if err := config.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化后端客户端
	backend.InitBackendClient()

	// 初始化 MAVLink
	if err := mavlink.InitMavlinkCommander(
		config.AppConfig.Mavlink.SerialPort,
		config.AppConfig.Mavlink.SerialBaud,
		config.AppConfig.Mavlink.TargetSystem,
	); err != nil {
		log.Printf("Warning: Failed to initialize MAVLink: %v", err)
	}

	// 设置 Gin 模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	// 注册路由
	api.SetupRoutes(router)

	// 配置 HTTP 服务器
	addr := fmt.Sprintf("0.0.0.0:%s", config.AppConfig.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	// 启动服务器
	if config.AppConfig.Server.Domain != "" {
		// 使用 Let's Encrypt
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(config.AppConfig.Server.Domain),
			Cache:      autocert.DirCache("letsencrypt-cache"),
			Email:      config.AppConfig.Server.Email,
		}

		tlsConfig := &tls.Config{
			GetCertificate: certManager.GetCertificate,
		}

		// 启动 HTTP 服务器用于 ACME 验证
		go func() {
			httpServer := &http.Server{
				Addr:    ":80",
				Handler: certManager.HTTPHandler(nil),
			}
			log.Printf("[Central] ACME HTTP server started on :80 for certificate verification")
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("[Central] ACME HTTP server error: %v", err)
			}
		}()

		// 启动 HTTPS 服务器
		tlsListener, err := tls.Listen("tcp", addr, tlsConfig)
		if err != nil {
			log.Fatalf("Failed to create TLS listener: %v", err)
		}

		log.Printf("[Central] HTTPS Server started on %s with Let's Encrypt", addr)
		if err := server.Serve(tlsListener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	} else if config.AppConfig.Server.CertFile != "" && config.AppConfig.Server.KeyFile != "" {
		// 使用手动证书
		log.Printf("[Central] HTTPS Server started on %s with manual certificate", addr)
		if err := server.ListenAndServeTLS(config.AppConfig.Server.CertFile, config.AppConfig.Server.KeyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		// 不使用 TLS
		log.Printf("[Central] HTTP Server started on %s (no TLS)", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}
}
