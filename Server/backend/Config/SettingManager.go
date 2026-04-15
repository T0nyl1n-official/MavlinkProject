package Config

import (
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

	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	sm.setting = &Setting{}
	if err := yaml.Unmarshal(data, sm.setting); err != nil {
		return err
	}

	sm.setDefaults()
	return nil
}

// 提高鲁棒性, 这些if值会导致系统错误
func (sm *SettingManager) setDefaults() {
	if sm.setting.Logger.MonitorWindow <= 0 {
		sm.setting.Logger.MonitorWindow = 15
	}
	if sm.setting.Board.Connection.Timeout <= 0 {
		sm.setting.Board.Connection.Timeout = 180
	}
	if sm.setting.Board.Connection.MaxRetryAttempts <= 0 {
		sm.setting.Board.Connection.MaxRetryAttempts = 3
	}
	if sm.setting.Board.Connection.RetryDelay <= 0 {
		sm.setting.Board.Connection.RetryDelay = 5
	}
	if sm.setting.Board.Connection.KeepaliveInterval <= 0 {
		sm.setting.Board.Connection.KeepaliveInterval = 10
	}
}

// 设置重加载
func (sm *SettingManager) ReloadSetting() error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.configPath == "" {
		return nil
	}

	data, err := os.ReadFile(sm.configPath)
	if err != nil {
		return err
	}

	newSetting := &Setting{}
	if err := yaml.Unmarshal(data, newSetting); err != nil {
		return err
	}

	sm.setDefaults()

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

	sm.setDefaults()

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
