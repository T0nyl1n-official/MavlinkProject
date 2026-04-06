package Config

import (
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type Setting struct {
	Database      DatabaseConfig      `yaml:"database"`
	Redis         RedisConfig         `yaml:"redis"`
	JWT           JWTConfig           `yaml:"jwt"`
	CORS          CORSConfig          `yaml:"cors"`
	RateLimit     RateLimitConfig     `yaml:"rate_limit"`
	Logger        LoggerConfig        `yaml:"logger"`
	Resources     ResourcesConfig     `yaml:"resources"`
	Verification  VerificationConfig  `yaml:"verification"`
	ErrorListener ErrorListenerConfig `yaml:"error_listener"`
	Board         BoardConfig         `yaml:"board"`
}

type DatabaseConfig struct {
	MySQL MySQLConfig `yaml:"mysql"`
}

type MySQLConfig struct {
	Host            string `yaml:"host"`
	Port            string `yaml:"port"`
	User            string `yaml:"user"`
	Password        string `yaml:"password"`
	Database        string `yaml:"database"`
	Charset         string `yaml:"charset"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
}

type RedisConfig struct {
	Host         string             `yaml:"host"`
	Port         string             `yaml:"port"`
	Password     string             `yaml:"password"`
	DBAllocation DBAllocationConfig `yaml:"db_allocation"`
}

type DBAllocationConfig struct {
	GeneralWarning int `yaml:"general_warning"`
	Backend        int `yaml:"backend"`
	Frontend       int `yaml:"frontend"`
	Agent          int `yaml:"agent"`
	Drone          int `yaml:"drone"`
	Sensor         int `yaml:"sensor"`
	Token          int `yaml:"token"`
	Verification   int `yaml:"verification"`
}

type JWTConfig struct {
	SecretKey   string `yaml:"secret_key"`
	ExpireTime  int    `yaml:"expire_time"`
	RefreshTime int    `yaml:"refresh_time"`
	IdentityKey string `yaml:"identity_key"`
}

type CORSConfig struct {
	AllowOrigins     []string `yaml:"allow_origins"`
	AllowMethods     []string `yaml:"allow_methods"`
	AllowHeaders     []string `yaml:"allow_headers"`
	ExposeHeaders    []string `yaml:"expose_headers"`
	AllowCredentials bool     `yaml:"allow_credentials"`
	MaxAge           int      `yaml:"max_age"`
}

type RateLimitConfig struct {
	Enabled           bool              `yaml:"enabled"`
	RedisPrefix       string            `yaml:"redis_prefix"`
	EnableGlobalLimit bool              `yaml:"enable_global_limit"`
	DefaultRule       DefaultRuleConfig `yaml:"default_rule"`
}

type DefaultRuleConfig struct {
	LimitType     string  `yaml:"limit_type"`
	Capacity      int64   `yaml:"capacity"`
	FillRate      float64 `yaml:"fill_rate"`
	WindowSeconds int64   `yaml:"window_seconds"`
}

type LoggerConfig struct {
	Level                string  `yaml:"level"`
	LogDir               string  `yaml:"log_dir"`
	MaxFileSize          int     `yaml:"max_file_size"`
	AccessThreshold      int     `yaml:"access_threshold"`
	URLErrorThreshold    int     `yaml:"url_error_threshold"`
	SlowRequestThreshold float64 `yaml:"slow_request_threshold"`
	MonitorWindow        int     `yaml:"monitor_window"`
}

type ResourcesConfig struct {
	StaticDir string `yaml:"static_dir"`
}

type VerificationConfig struct {
	CodeLength          int `yaml:"code_length"`
	CodeExpireTime      int `yaml:"code_expire_time"`
	MaxRequestPerMinute int `yaml:"max_request_per_minute"`
}

type ErrorListenerConfig struct {
	EnablePanicRecovery bool     `yaml:"enable_panic_recovery"`
	EnableErrorLogging  bool     `yaml:"enable_error_logging"`
	EnableWarningPush   bool     `yaml:"enable_warning_push"`
	Sources             []string `yaml:"sources"`
}

type BoardConfig struct {
	TCP        BoardTCPConfig        `yaml:"tcp"`
	UDP        BoardUDPConfig        `yaml:"udp"`
	Connection BoardConnectionConfig `yaml:"connection"`
}

type BoardTCPConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Address       string `yaml:"address"`
	Port          string `yaml:"port"`
	MaxBufferSize int    `yaml:"max_buffer_size"`
}

type BoardUDPConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Address       string `yaml:"address"`
	Port          string `yaml:"port"`
	MaxBufferSize int    `yaml:"max_buffer_size"`
}

type BoardConnectionConfig struct {
	Timeout           int `yaml:"timeout"`
	MaxRetryAttempts  int `yaml:"max_retry_attempts"`
	RetryDelay        int `yaml:"retry_delay"`
	KeepaliveInterval int `yaml:"keepalive_interval"`
}

var (
	setting     *Setting
	settingOnce sync.Once
)

func LoadSetting(path string) (*Setting, error) {
	var err error
	settingOnce.Do(func() {
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			err = readErr
			return
		}
		setting = &Setting{}
		err = yaml.Unmarshal(data, setting)
		if err != nil {
			return
		}
		setDefaults()
	})
	return setting, err
}

func setDefaults() {
	if setting.Logger.MonitorWindow <= 0 {
		setting.Logger.MonitorWindow = 15
	}
}

func GetSetting() *Setting {
	return setting
}
