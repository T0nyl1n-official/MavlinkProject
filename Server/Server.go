package Server

import (
	Backend "MavlinkProject/Server/backend"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Backend struct {
		Address string `yaml:"address"`
		Port    string `yaml:"port"`
	} `yaml:"backend"`
}

var BackendServer Backend.BackendServer

func Server_start() {
	configPath := "config/Server_Config.yaml"
	cfg := getConfig(configPath)

	BackendServer.Start(cfg.Backend.Address, cfg.Backend.Port)
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