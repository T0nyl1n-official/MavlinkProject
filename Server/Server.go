package Server

import (
	Backend "MavlinkProject/Server/backend"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Backend struct {
		Address     string `yaml:"address"`
		Port        string `yaml:"port"`
		LetsEncrypt struct {
			Email      string   `yaml:"email"`
			Domains    []string `yaml:"domains"`
			Webroot    string   `yaml:"webroot"`
			UseStaging bool     `yaml:"use_staging"`
		} `yaml:"lets_encrypt"`
	} `yaml:"backend"`
}

var BackendServer Backend.BackendServer

func Server_start() {
	configPath := "config/Server_Config.yaml"
	cfg := getConfig(configPath)

	BackendServer.Start(cfg.Backend.Address, cfg.Backend.Port, cfg.Backend.LetsEncrypt)
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
