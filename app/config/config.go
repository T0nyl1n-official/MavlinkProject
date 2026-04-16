package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port     string `yaml:"port"`
		Domain   string `yaml:"domain"`
		Email    string `yaml:"email"`
		CertFile string `yaml:"cert_file"`
		KeyFile  string `yaml:"key_file"`
	} `yaml:"server"`
	Mavlink struct {
		SerialPort   string `yaml:"serial_port"`
		SerialBaud   int    `yaml:"serial_baud"`
		TargetSystem uint8  `yaml:"target_system"`
	} `yaml:"mavlink"`
	Backend struct {
		Address   string `yaml:"address"`
		Token     string `yaml:"token"`
		DeviceID  string `yaml:"device_id"`
		DeviceType string `yaml:"device_type"`
	} `yaml:"backend"`
	Drone struct {
		Search struct {
			MinBatteryLevel    float64 `yaml:"min_battery_level"`
			MaxDroneDistance   float64 `yaml:"max_drone_distance"`
			StatusCheckTimeout int     `yaml:"status_check_timeout"` // 秒
			StatusUpdateInterval int    `yaml:"status_update_interval"` // 秒
			MessageChanSize    int     `yaml:"message_chan_size"`
			ScoreWeight        float64 `yaml:"score_weight"`
		} `yaml:"search"`
	} `yaml:"drone"`
}

var AppConfig *Config

func LoadConfig(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	// 设置默认值
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8084"
	}
	if cfg.Mavlink.SerialBaud == 0 {
		cfg.Mavlink.SerialBaud = 115200
	}
	if cfg.Mavlink.TargetSystem == 0 {
		cfg.Mavlink.TargetSystem = 1
	}
	if cfg.Backend.Address == "" {
		cfg.Backend.Address = "https://api.deeppluse.dpdns.org"
	}
	if cfg.Backend.DeviceID == "" {
		cfg.Backend.DeviceID = "central_001"
	}
	if cfg.Backend.DeviceType == "" {
		cfg.Backend.DeviceType = "central"
	}

	AppConfig = &cfg
	return nil
}
