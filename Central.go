package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/acme/autocert"

	api "MavlinkProject_Board/app/api"
	config "MavlinkProject_Board/app/config"
	backend "MavlinkProject_Board/app/services/backend"
	mavlink "MavlinkProject_Board/app/services/mavlink"
	Core "MavlinkProject_Board/Core"
	Distribute "MavlinkProject_Board/Distribute"
)

func main() {
	configPath := "config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	if err := config.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := Core.LoadConfig(configPath); err != nil {
		log.Printf("Warning: Failed to load Core config: %v", err)
	}

	mode := "api"
	if m, ok := os.LookupEnv("CENTRAL_MODE"); ok {
		mode = m
	}
	if len(os.Args) > 2 && (os.Args[2] == "--tcp" || os.Args[2] == "--central") {
		mode = "tcp"
	}

	switch mode {
	case "tcp":
		startTCPMode()
	default:
		startAPIMode()
	}
}

func startAPIMode() {
	log.Println("[Central] Starting in HTTP API mode...")

	backend.InitBackendClient()

	if err := mavlink.InitMavlinkCommander(
		config.AppConfig.Mavlink.SerialPort,
		config.AppConfig.Mavlink.SerialBaud,
		config.AppConfig.Mavlink.TargetSystem,
	); err != nil {
		log.Printf("Warning: Failed to initialize MAVLink: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.Logger())

	api.SetupRoutes(router)

	addr := fmt.Sprintf("0.0.0.0:%s", config.AppConfig.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}

	if config.AppConfig.Server.Domain != "" {
		certManager := autocert.Manager{
			Prompt:     autocert.AcceptTOS,
			HostPolicy: autocert.HostWhitelist(config.AppConfig.Server.Domain),
			Cache:      autocert.DirCache("letsencrypt-cache"),
			Email:      config.AppConfig.Server.Email,
		}

		tlsConfig := &tls.Config{
			GetCertificate: certManager.GetCertificate,
		}

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

		tlsListener, err := tls.Listen("tcp", addr, tlsConfig)
		if err != nil {
			log.Fatalf("Failed to create TLS listener: %v", err)
		}

		log.Printf("[Central] HTTPS Server started on %s with Let's Encrypt", addr)
		if err := server.Serve(tlsListener); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	} else if config.AppConfig.Server.CertFile != "" && config.AppConfig.Server.KeyFile != "" {
		log.Printf("[Central] HTTPS Server started on %s with manual certificate", addr)
		if err := server.ListenAndServeTLS(config.AppConfig.Server.CertFile, config.AppConfig.Server.KeyFile); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		log.Printf("[Central] HTTP Server started on %s (no TLS)", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}
}

func startTCPMode() {
	log.Println("[Central] Starting in TCP Server mode (Central调度系统)...")

	cfg := Core.GetConfig()
	port := cfg.Central.Port
	if port == "" {
		port = "8081"
	}
	address := cfg.Central.Address

	central := Core.NewCentralServer(address, port)

	timeoutSec := cfg.Central.Drone.StatusCheckTimeout
	if timeoutSec == 0 {
		timeoutSec = 10
	}
	central.DroneSearch().UpdateConfig(Distribute.DroneConfig{
		MinBatteryLevel:    cfg.Central.Drone.MinBatteryLevel,
		MaxDroneDistance:   cfg.Central.Drone.MaxDroneDistance,
		StatusCheckTimeout: time.Duration(timeoutSec) * time.Second,
	})

	if err := central.Start(); err != nil {
		log.Fatalf("Failed to start CentralServer: %v", err)
	}

	if cfg.Pixhawk.Enabled && cfg.Pixhawk.SerialPort != "" {
		Core.StartLocalPixhawk(central, cfg.Pixhawk.SerialPort, cfg.Pixhawk.SerialBaud)
	}

	log.Printf("[Central] TCP 调度系统已启动, 监听地址 %s:%s", address, port)
	log.Printf("[Central] 等待接收ProgressChain任务链...")

	central.WaitForShutdown()

	log.Printf("[Central] TCP 调度系统正在关闭...")
	central.Stop()
	log.Printf("[Central] TCP 调度系统已关闭")
}
