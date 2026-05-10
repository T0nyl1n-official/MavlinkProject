package Config

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
)

// ==================== AIConfig YAML 解析测试 ====================

func TestAIConfigYAMLParsing(t *testing.T) {
	yamlData := `
ai:
  lstm:
    enabled: true
    url: "http://localhost:5000"
    timeout: 30
  yolo:
    enabled: true
    url: "http://localhost:5001"
    timeout: 60
`
	var setting Setting
	err := yaml.Unmarshal([]byte(yamlData), &setting)
	if err != nil {
		t.Fatalf("AIConfig YAML 解析失败: %v", err)
	}

	if !setting.AI.LSTM.Enabled {
		t.Error("LSTM 应该启用")
	}
	if setting.AI.LSTM.URL != "http://localhost:5000" {
		t.Errorf("LSTM URL 不匹配: 期望 http://localhost:5000, 实际 %s", setting.AI.LSTM.URL)
	}
	if setting.AI.LSTM.Timeout != 30 {
		t.Errorf("LSTM Timeout 不匹配: 期望 30, 实际 %d", setting.AI.LSTM.Timeout)
	}

	if !setting.AI.YOLO.Enabled {
		t.Error("YOLO 应该启用")
	}
	if setting.AI.YOLO.URL != "http://localhost:5001" {
		t.Errorf("YOLO URL 不匹配: 期望 http://localhost:5001, 实际 %s", setting.AI.YOLO.URL)
	}
	if setting.AI.YOLO.Timeout != 60 {
		t.Errorf("YOLO Timeout 不匹配: 期望 60, 实际 %d", setting.AI.YOLO.Timeout)
	}
}

func TestAIConfigYAMLDisabled(t *testing.T) {
	yamlData := `
ai:
  lstm:
    enabled: false
    url: ""
    timeout: 0
  yolo:
    enabled: false
    url: ""
    timeout: 0
`
	var setting Setting
	err := yaml.Unmarshal([]byte(yamlData), &setting)
	if err != nil {
		t.Fatalf("AIConfig YAML 解析失败: %v", err)
	}

	if setting.AI.LSTM.Enabled {
		t.Error("LSTM 不应该启用")
	}
	if setting.AI.YOLO.Enabled {
		t.Error("YOLO 不应该启用")
	}
}

func TestAIConfigYAMLEmpty(t *testing.T) {
	yamlData := `
ai:
  lstm:
  yolo:
`
	var setting Setting
	err := yaml.Unmarshal([]byte(yamlData), &setting)
	if err != nil {
		t.Fatalf("空 AIConfig YAML 解析失败: %v", err)
	}

	if setting.AI.LSTM.Enabled {
		t.Error("空 LSTM 配置不应该启用")
	}
	if setting.AI.YOLO.Enabled {
		t.Error("空 YOLO 配置不应该启用")
	}
}

func TestAIConfigYAMLMissing(t *testing.T) {
	yamlData := `
database:
  mysql:
    host: "localhost"
`
	var setting Setting
	err := yaml.Unmarshal([]byte(yamlData), &setting)
	if err != nil {
		t.Fatalf("缺少 AI 配置的 YAML 解析失败: %v", err)
	}

	// 缺少 ai 配置时，应该使用零值
	if setting.AI.LSTM.Enabled {
		t.Error("缺少 AI 配置时 LSTM 不应该启用")
	}
	if setting.AI.YOLO.Enabled {
		t.Error("缺少 AI 配置时 YOLO 不应该启用")
	}
}

// ==================== LSTMModelConfig 测试 ====================

func TestLSTMModelConfigYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		expected LSTMModelConfig
	}{
		{
			name: "完整配置",
			yamlData: `
enabled: true
url: "http://lstm-server:5000"
timeout: 45
`,
			expected: LSTMModelConfig{Enabled: true, URL: "http://lstm-server:5000", Timeout: 45},
		},
		{
			name: "禁用配置",
			yamlData: `
enabled: false
url: ""
timeout: 0
`,
			expected: LSTMModelConfig{Enabled: false, URL: "", Timeout: 0},
		},
		{
			name: "只有 URL",
			yamlData: `
url: "http://lstm:5000"
`,
			expected: LSTMModelConfig{Enabled: false, URL: "http://lstm:5000", Timeout: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config LSTMModelConfig
			err := yaml.Unmarshal([]byte(tt.yamlData), &config)
			if err != nil {
				t.Fatalf("LSTMModelConfig YAML 解析失败: %v", err)
			}
			if config.Enabled != tt.expected.Enabled {
				t.Errorf("Enabled 不匹配: 期望 %v, 实际 %v", tt.expected.Enabled, config.Enabled)
			}
			if config.URL != tt.expected.URL {
				t.Errorf("URL 不匹配: 期望 %s, 实际 %s", tt.expected.URL, config.URL)
			}
			if config.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout 不匹配: 期望 %d, 实际 %d", tt.expected.Timeout, config.Timeout)
			}
		})
	}
}

// ==================== YOLOModelConfig 测试 ====================

func TestYOLOModelConfigYAML(t *testing.T) {
	tests := []struct {
		name     string
		yamlData string
		expected YOLOModelConfig
	}{
		{
			name: "完整配置",
			yamlData: `
enabled: true
url: "http://yolo-server:5001"
timeout: 90
`,
			expected: YOLOModelConfig{Enabled: true, URL: "http://yolo-server:5001", Timeout: 90},
		},
		{
			name: "禁用配置",
			yamlData: `
enabled: false
url: ""
timeout: 0
`,
			expected: YOLOModelConfig{Enabled: false, URL: "", Timeout: 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var config YOLOModelConfig
			err := yaml.Unmarshal([]byte(tt.yamlData), &config)
			if err != nil {
				t.Fatalf("YOLOModelConfig YAML 解析失败: %v", err)
			}
			if config.Enabled != tt.expected.Enabled {
				t.Errorf("Enabled 不匹配: 期望 %v, 实际 %v", tt.expected.Enabled, config.Enabled)
			}
			if config.URL != tt.expected.URL {
				t.Errorf("URL 不匹配: 期望 %s, 实际 %s", tt.expected.URL, config.URL)
			}
			if config.Timeout != tt.expected.Timeout {
				t.Errorf("Timeout 不匹配: 期望 %d, 实际 %d", tt.expected.Timeout, config.Timeout)
			}
		})
	}
}

// ==================== Setting 完整 YAML 解析测试 ====================

func TestSettingFullYAMLParsing(t *testing.T) {
	yamlData := `
database:
  mysql:
    host: "localhost"
    port: "3306"
    user: "root"
    password: "testpass"
    database: "testdb"
    charset: "utf8mb4"

redis:
  host: "localhost"
  port: "6379"

jwt:
  secret_key: "test_secret"

ai:
  lstm:
    enabled: true
    url: "http://lstm:5000"
    timeout: 30
  yolo:
    enabled: true
    url: "http://yolo:5001"
    timeout: 60
`
	var setting Setting
	err := yaml.Unmarshal([]byte(yamlData), &setting)
	if err != nil {
		t.Fatalf("Setting YAML 解析失败: %v", err)
	}

	// 验证 AI 配置
	if !setting.AI.LSTM.Enabled {
		t.Error("LSTM 应该启用")
	}
	if setting.AI.LSTM.URL != "http://lstm:5000" {
		t.Errorf("LSTM URL 不匹配: 期望 http://lstm:5000, 实际 %s", setting.AI.LSTM.URL)
	}
	if !setting.AI.YOLO.Enabled {
		t.Error("YOLO 应该启用")
	}
	if setting.AI.YOLO.URL != "http://yolo:5001" {
		t.Errorf("YOLO URL 不匹配: 期望 http://yolo:5001, 实际 %s", setting.AI.YOLO.URL)
	}

	// 验证其他配置不受影响
	if setting.Database.MySQL.Host != "localhost" {
		t.Errorf("MySQL Host 不匹配: 期望 localhost, 实际 %s", setting.Database.MySQL.Host)
	}
}

// ==================== SettingManager 文件加载测试 ====================

func TestSettingManagerLoadWithAIConfig(t *testing.T) {
	// 创建临时 YAML 文件
	yamlContent := `
ai:
  lstm:
    enabled: true
    url: "http://lstm-test:5000"
    timeout: 30
  yolo:
    enabled: false
    url: ""
    timeout: 0

logger:
  level: "INFO"
  monitor_window: 15

board:
  connection:
    timeout: 180
    max_retry_attempts: 3
    retry_delay: 5
    keepalive_interval: 10
`
	tmpFile, err := os.CreateTemp("", "setting_test_*.yaml")
	if err != nil {
		t.Fatalf("创建临时文件失败: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(yamlContent)
	if err != nil {
		t.Fatalf("写入临时文件失败: %v", err)
	}
	tmpFile.Close()

	sm := &SettingManager{
		setting:           &Setting{},
		onChangeCallbacks: make(map[string]func(*Setting) error),
	}

	err = sm.LoadSetting(tmpFile.Name())
	if err != nil {
		t.Fatalf("LoadSetting 失败: %v", err)
	}

	setting := sm.GetSetting()
	if !setting.AI.LSTM.Enabled {
		t.Error("LSTM 应该启用")
	}
	if setting.AI.LSTM.URL != "http://lstm-test:5000" {
		t.Errorf("LSTM URL 不匹配: 期望 http://lstm-test:5000, 实际 %s", setting.AI.LSTM.URL)
	}
	if setting.AI.YOLO.Enabled {
		t.Error("YOLO 不应该启用")
	}
}

func TestSettingManagerLoadNonexistentFile(t *testing.T) {
	sm := &SettingManager{
		setting:           &Setting{},
		onChangeCallbacks: make(map[string]func(*Setting) error),
	}

	// 加载不存在的文件不应该返回错误，而是使用默认值
	err := sm.LoadSetting("/nonexistent/path/setting.yaml")
	if err != nil {
		t.Fatalf("加载不存在的文件不应该返回错误: %v", err)
	}

	setting := sm.GetSetting()
	// AI 配置应该为零值
	if setting.AI.LSTM.Enabled {
		t.Error("加载不存在的文件后 LSTM 不应该启用")
	}
}

// ==================== AIConfig 结构体测试 ====================

func TestAIConfigStructFields(t *testing.T) {
	config := AIConfig{
		LSTM: LSTMModelConfig{
			Enabled: true,
			URL:     "http://lstm:5000",
			Timeout: 30,
		},
		YOLO: YOLOModelConfig{
			Enabled: true,
			URL:     "http://yolo:5001",
			Timeout: 60,
		},
	}

	if !config.LSTM.Enabled {
		t.Error("LSTM.Enabled 应该为 true")
	}
	if config.LSTM.URL != "http://lstm:5000" {
		t.Errorf("LSTM.URL 不匹配: 期望 http://lstm:5000, 实际 %s", config.LSTM.URL)
	}
	if config.LSTM.Timeout != 30 {
		t.Errorf("LSTM.Timeout 不匹配: 期望 30, 实际 %d", config.LSTM.Timeout)
	}

	if !config.YOLO.Enabled {
		t.Error("YOLO.Enabled 应该为 true")
	}
	if config.YOLO.URL != "http://yolo:5001" {
		t.Errorf("YOLO.URL 不匹配: 期望 http://yolo:5001, 实际 %s", config.YOLO.URL)
	}
	if config.YOLO.Timeout != 60 {
		t.Errorf("YOLO.Timeout 不匹配: 期望 60, 实际 %d", config.YOLO.Timeout)
	}
}

func TestAIConfigZeroValues(t *testing.T) {
	var config AIConfig

	if config.LSTM.Enabled {
		t.Error("零值 AIConfig 的 LSTM.Enabled 应该为 false")
	}
	if config.LSTM.URL != "" {
		t.Errorf("零值 AIConfig 的 LSTM.URL 应该为空, 实际 %s", config.LSTM.URL)
	}
	if config.LSTM.Timeout != 0 {
		t.Errorf("零值 AIConfig 的 LSTM.Timeout 应该为 0, 实际 %d", config.LSTM.Timeout)
	}

	if config.YOLO.Enabled {
		t.Error("零值 AIConfig 的 YOLO.Enabled 应该为 false")
	}
	if config.YOLO.URL != "" {
		t.Errorf("零值 AIConfig 的 YOLO.URL 应该为空, 实际 %s", config.YOLO.URL)
	}
	if config.YOLO.Timeout != 0 {
		t.Errorf("零值 AIConfig 的 YOLO.Timeout 应该为 0, 实际 %d", config.YOLO.Timeout)
	}
}
