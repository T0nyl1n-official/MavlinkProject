package Server

import (
	Backend "MavlinkProject/Server/backend"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Backend struct {
		Port string `yaml:"port"`
	} `yaml:"backend"`
	Frontend struct {
		Port string `yaml:"port"`
	} `yaml:"frontend"`
}

var BackendServer Backend.BackendServer

func Server_start() {
	// 配置获取
	configPath := "config/Server_Config.yaml"
	cfg := getConfig(configPath)

	// 后端开启
	BackendServer.Start(cfg.Backend.Port)
}

func getConfig(configPath string) Config {
	config := Config{}
	data, err := os.ReadFile(configPath)
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("Failed to load config file: %v", err)
	}
	return config
}
