package Config

/*
	SettingManager: 负责对于Config文件夹内的setting文件进行加载, 重加载, 并提供设置的读取接口
	LoadSetting: 加载setting文件
	ReloadSetting: 重加载setting文件
	GetSetting: 获取setting
*/

import (
	"log"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

type SettingManager struct {
	mu                sync.RWMutex
	setting           *Setting
	configPath        string
	onChangeCallbacks map[string]func(*Setting) error
}

var (
	settingManager     *SettingManager
	settingManagerOnce sync.Once
)

func GetSettingManager() *SettingManager {
	settingManagerOnce.Do(func() {
		settingManager = &SettingManager{
			setting:           &Setting{},
			onChangeCallbacks: make(map[string]func(*Setting) error),
		}
	})
	return settingManager
}

func (sm *SettingManager) GetSetting() *Setting {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.setting
}

func (sm *SettingManager) LoadSetting(path string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.configPath = path

	sm.setting = &Setting{}

	sm.setting.loadFromEnv()

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[SettingManager] 警告: 无法读取配置文件 %s: %v, 使用环境变量和默认值", path, err)
		sm.setting.setDefaults()
		return nil
	}

	yamlSetting := &Setting{}
	if err := yaml.Unmarshal(data, yamlSetting); err != nil {
		return err
	}

	sm.setting.mergeFromYAML(yamlSetting)

	sm.setting.setDefaults()
	return nil
}

func (s *Setting) loadFromEnv() {
	log.Println("[SettingManager] 优先从环境变量加载配置...")

	if val := os.Getenv("MavlinkProject_backend_database_mysql_host"); val != "" {
		s.Database.MySQL.Host = val
		log.Printf("[SettingManager] MySQL Host 从环境变量加载: %s", val)
	}
	if val := os.Getenv("MavlinkProject_backend_database_mysql_port"); val != "" {
		s.Database.MySQL.Port = val
		log.Printf("[SettingManager] MySQL Port 从环境变量加载: %s", val)
	}
	if val := os.Getenv("MavlinkProject_backend_database_mysql_user"); val != "" {
		s.Database.MySQL.User = val
		log.Printf("[SettingManager] MySQL User 从环境变量加载: %s", val)
	}
	if val := os.Getenv("MavlinkProject_backend_database_mysql_password"); val != "" {
		s.Database.MySQL.Password = val
		log.Printf("[SettingManager] MySQL Password 从环境变量加载 (长度: %d)", len(val))
	}
	if val := os.Getenv("MavlinkProject_backend_database_mysql_database"); val != "" {
		s.Database.MySQL.Database = val
		log.Printf("[SettingManager] MySQL Database 从环境变量加载: %s", val)
	}
	if val := os.Getenv("MavlinkProject_backend_database_mysql_charset"); val != "" {
		s.Database.MySQL.Charset = val
		log.Printf("[SettingManager] MySQL Charset 从环境变量加载: %s", val)
	}

	if val := os.Getenv("MavlinkProject_backend_redis_host"); val != "" {
		s.Redis.Host = val
	}
	if val := os.Getenv("MavlinkProject_backend_redis_port"); val != "" {
		s.Redis.Port = val
	}
	if val := os.Getenv("MavlinkProject_backend_redis_password"); val != "" {
		s.Redis.Password = val
	}

	if val := os.Getenv("MavlinkProject_backend_jwt_secret_key"); val != "" {
		s.JWT.SecretKey = val
	}
}

func (s *Setting) mergeFromYAML(yamlSetting *Setting) {
	log.Println("[SettingManager] 从YAML文件补充缺失的配置...")

	if s.Database.MySQL.Host == "" && yamlSetting.Database.MySQL.Host != "" {
		s.Database.MySQL.Host = yamlSetting.Database.MySQL.Host
		log.Printf("[SettingManager] MySQL Host 从YAML加载: %s", yamlSetting.Database.MySQL.Host)
	}
	if s.Database.MySQL.Port == "" && yamlSetting.Database.MySQL.Port != "" {
		s.Database.MySQL.Port = yamlSetting.Database.MySQL.Port
		log.Printf("[SettingManager] MySQL Port 从YAML加载: %s", yamlSetting.Database.MySQL.Port)
	}
	if s.Database.MySQL.User == "" && yamlSetting.Database.MySQL.User != "" {
		s.Database.MySQL.User = yamlSetting.Database.MySQL.User
		log.Printf("[SettingManager] MySQL User 从YAML加载: %s", yamlSetting.Database.MySQL.User)
	}
	if s.Database.MySQL.Password == "" && yamlSetting.Database.MySQL.Password != "" {
		s.Database.MySQL.Password = yamlSetting.Database.MySQL.Password
		log.Printf("[SettingManager] MySQL Password 从YAML加载 (长度: %d)", len(yamlSetting.Database.MySQL.Password))
	}
	if s.Database.MySQL.Database == "" && yamlSetting.Database.MySQL.Database != "" {
		s.Database.MySQL.Database = yamlSetting.Database.MySQL.Database
		log.Printf("[SettingManager] MySQL Database 从YAML加载: %s", yamlSetting.Database.MySQL.Database)
	}
	if s.Database.MySQL.Charset == "" && yamlSetting.Database.MySQL.Charset != "" {
		s.Database.MySQL.Charset = yamlSetting.Database.MySQL.Charset
		log.Printf("[SettingManager] MySQL Charset 从YAML加载: %s", yamlSetting.Database.MySQL.Charset)
	}

	if s.Redis.Host == "" && yamlSetting.Redis.Host != "" {
		s.Redis.Host = yamlSetting.Redis.Host
	}
	if s.Redis.Port == "" && yamlSetting.Redis.Port != "" {
		s.Redis.Port = yamlSetting.Redis.Port
	}
	if s.Redis.Password == "" && yamlSetting.Redis.Password != "" {
		s.Redis.Password = yamlSetting.Redis.Password
	}

	if s.JWT.SecretKey == "" && yamlSetting.JWT.SecretKey != "" {
		s.JWT.SecretKey = yamlSetting.JWT.SecretKey
	}

	s.Database.MySQL.MaxIdleConns = yamlSetting.Database.MySQL.MaxIdleConns
	s.Database.MySQL.MaxOpenConns = yamlSetting.Database.MySQL.MaxOpenConns
	s.Database.MySQL.ConnMaxLifetime = yamlSetting.Database.MySQL.ConnMaxLifetime

	s.Logger = yamlSetting.Logger
	s.CORS = yamlSetting.CORS
	s.RateLimit = yamlSetting.RateLimit
	s.Board = yamlSetting.Board

	log.Printf("[SettingManager] 最终配置: Host=%s, User=%s, Password长度=%d, Database=%s",
		s.Database.MySQL.Host,
		s.Database.MySQL.User,
		len(s.Database.MySQL.Password),
		s.Database.MySQL.Database)
}

// 提高鲁棒性, 这些if值会导致系统错误
func (s *Setting) setDefaults() {
	if s.Logger.MonitorWindow <= 0 {
		s.Logger.MonitorWindow = 15
	}
	if s.Board.Connection.Timeout <= 0 {
		s.Board.Connection.Timeout = 180
	}
	if s.Board.Connection.MaxRetryAttempts <= 0 {
		s.Board.Connection.MaxRetryAttempts = 3
	}
	if s.Board.Connection.RetryDelay <= 0 {
		s.Board.Connection.RetryDelay = 5
	}
	if s.Board.Connection.KeepaliveInterval <= 0 {
		s.Board.Connection.KeepaliveInterval = 10
	}
}

// 设置重加载 从文件加载设置, 优先使用环境变量, 其次使用默认值, 最后使用文件中的值
func (sm *SettingManager) ReloadSetting() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.configPath == "" {
		return nil
	}

	newSetting := &Setting{}
	newSetting.loadFromEnv()

	data, err := os.ReadFile(sm.configPath)
	if err != nil {
		log.Printf("[SettingManager] 警告: 无法读取配置文件 %s: %v, 使用环境变量和默认值", sm.configPath, err)
		sm.setting = newSetting
		sm.setting.setDefaults()
		return nil
	}

	yamlSetting := &Setting{}
	if err := yaml.Unmarshal(data, yamlSetting); err != nil {
		return err
	}

	newSetting.mergeFromYAML(yamlSetting)
	newSetting.setDefaults()

	oldSetting := sm.setting
	sm.setting = newSetting

	sm.notifyCallbacks(oldSetting, newSetting)

	return nil
}

// 更新设置
func (sm *SettingManager) UpdateSetting(newSetting *Setting) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	oldSetting := sm.setting
	sm.setting = newSetting

	sm.setting.setDefaults()

	if sm.configPath != "" {
		if err := sm.saveToFile(); err != nil {
			sm.setting = oldSetting
			return err
		}
	}

	sm.notifyCallbacks(oldSetting, newSetting)

	return nil
}

// 保存设置到文件
func (sm *SettingManager) saveToFile() error {
	data, err := yaml.Marshal(sm.setting)
	if err != nil {
		return err
	}
	return os.WriteFile(sm.configPath, data, 0644)
}

func (sm *SettingManager) notifyCallbacks(oldSetting, newSetting *Setting) {
	for _, callback := range sm.onChangeCallbacks {
		if err := callback(newSetting); err != nil {
			continue
		}
	}
}

// 注册设置变化回调
func (sm *SettingManager) RegisterChangeCallback(name string, callback func(*Setting) error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.onChangeCallbacks[name] = callback
}

func (sm *SettingManager) UnregisterChangeCallback(name string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	delete(sm.onChangeCallbacks, name)
}

func GetSetting() *Setting {
	return GetSettingManager().GetSetting()
}

func LoadSetting(path string) error {
	return GetSettingManager().LoadSetting(path)
}

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
	DroneSearch   DroneSearchConfig   `yaml:"drone_search"`
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
	FRP        BoardFRPConfig        `yaml:"frp"`
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

type CentralServerConfig struct {
	Name    string `yaml:"name"`
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

type BoardFRPConfig struct {
	Timeout          int                   `yaml:"timeout"`
	ReadTimeout      int                   `yaml:"read_timeout"`
	MaxRetryAttempts int                   `yaml:"max_retry_attempts"`
	CentralServers   []CentralServerConfig `yaml:"central_servers"`
}

type DroneSearchConfig struct {
	Timeout             int     `yaml:"timeout"`
	Retry               int     `yaml:"retry"`
	Batch               int     `yaml:"batch"`
	MinBattery          float64 `yaml:"min_battery"`
	MaxDistance         float64 `yaml:"max_distance"`
	StatusCheckInterval int     `yaml:"status_check_interval"`
}
