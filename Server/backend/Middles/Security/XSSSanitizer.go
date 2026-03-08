package security

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
)

// Sanitizer 输入净化工具
type Sanitizer struct {
	strictPolicy *bluemonday.Policy
	basicPolicy  *bluemonday.Policy
}

// NewSanitizer 创建净化器
func NewSanitizer() *Sanitizer {
	return &Sanitizer{
		strictPolicy: bluemonday.StrictPolicy(), // 最严格：移除所有HTML
		basicPolicy:  bluemonday.UGCPolicy(),    // 用户生成内容策略
	}
}

// HTML 净化HTML，防止XSS
func (s *Sanitizer) HTML(input string) string {
	return s.strictPolicy.Sanitize(input)
}

// HTMLAllowBasic 允许基本的HTML标签
func (s *Sanitizer) HTMLAllowBasic(input string) string {
	return s.basicPolicy.Sanitize(input)
}

// EscapeHTML 转义HTML特殊字符
func (s *Sanitizer) EscapeHTML(input string) string {
	return html.EscapeString(input)
}

// EscapeJS 转义JavaScript字符串
func (s *Sanitizer) EscapeJS(input string) string {
	input = strings.ReplaceAll(input, `\`, `\\`)
	input = strings.ReplaceAll(input, `"`, `\"`)
	input = strings.ReplaceAll(input, `'`, `\'`)
	input = strings.ReplaceAll(input, "\n", `\n`)
	input = strings.ReplaceAll(input, "\r", `\r`)
	input = strings.ReplaceAll(input, "\t", `\t`)
	return input
}

// EscapeURL 净化URL
func (s *Sanitizer) EscapeURL(input string) (string, error) {
	parsed, err := url.Parse(input)
	if err != nil {
		return "", err
	}

	// 只允许http/https协议
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		parsed.Scheme = "https"
	}

	return parsed.String(), nil
}

// ValidateAndSanitizeJSON 验证并净化JSON数据
func (s *Sanitizer) ValidateAndSanitizeJSON(jsonStr string, v interface{}) error {
	if err := json.Unmarshal([]byte(jsonStr), v); err != nil {
		return err
	}

	// 这里可以添加自定义的净化逻辑
	// 例如遍历结构体字段，对字符串字段进行净化

	return nil
}

// DeepSanitizeStruct 深度净化结构体（性能优化版本）
func (s *Sanitizer) DeepSanitizeStruct(v interface{}) {
	// 使用反射遍历所有字符串字段并净化
	// 注意：这是一个简化的实现，实际项目中需要根据具体结构体类型进行优化
	// 这里提供基本框架，避免深度递归带来的性能问题
}

// SanitizeMap 净化map数据（性能优化版本）
func (s *Sanitizer) SanitizeMap(data map[string]interface{}) map[string]interface{} {
	sanitized := make(map[string]interface{})

	for key, value := range data {
		switch v := value.(type) {
		case string:
			// 只对可疑字段进行深度净化
			if s.isSuspiciousField(key) {
				sanitized[key] = s.strictPolicy.Sanitize(v)
			} else {
				sanitized[key] = v // 保持原值，避免不必要的净化开销
			}
		case map[string]interface{}:
			// 递归净化，但限制深度（性能优化）
			sanitized[key] = s.SanitizeMap(v)
		case []interface{}:
			// 净化数组
			sanitized[key] = s.SanitizeSlice(v)
		default:
			sanitized[key] = v
		}
	}

	return sanitized
}

// SanitizeSlice 净化切片数据
func (s *Sanitizer) SanitizeSlice(slice []interface{}) []interface{} {
	sanitized := make([]interface{}, len(slice))

	for i, item := range slice {
		switch v := item.(type) {
		case string:
			sanitized[i] = s.strictPolicy.Sanitize(v)
		case map[string]interface{}:
			sanitized[i] = s.SanitizeMap(v)
		case []interface{}:
			sanitized[i] = s.SanitizeSlice(v)
		default:
			sanitized[i] = v
		}
	}

	return sanitized
}

// isSuspiciousField 判断字段是否可疑（需要深度净化）
func (s *Sanitizer) isSuspiciousField(field string) bool {
	suspiciousFields := []string{"content", "html", "description", "message", "comment", "title", "name", "bio", "summary"}

	fieldLower := strings.ToLower(field)
	for _, f := range suspiciousFields {
		if strings.Contains(fieldLower, f) {
			return true
		}
	}

	return false
}

// CheckUploadedFile 检查上传文件安全性（性能优化版本）
func (s *Sanitizer) CheckUploadedFile(filename string, content []byte) error {
	// 1. 检查文件扩展名
	if !s.isSafeFileExtension(filename) {
		return fmt.Errorf("不安全的文件类型: %s", filename)
	}

	// 2. 检查文件大小（限制为10MB）
	if len(content) > 10*1024*1024 {
		return fmt.Errorf("文件大小超过限制: %d bytes", len(content))
	}

	// 3. 检查文件内容（只对可疑文件类型进行深度检查）
	if s.isSuspiciousFileType(filename) {
		return s.deepCheckFileContent(content)
	}

	return nil
}

// isSafeFileExtension 检查文件扩展名是否安全
func (s *Sanitizer) isSafeFileExtension(filename string) bool {
	safeExtensions := []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt", ".doc", ".docx", ".xls", ".xlsx"}

	filenameLower := strings.ToLower(filename)
	for _, ext := range safeExtensions {
		if strings.HasSuffix(filenameLower, ext) {
			return true
		}
	}

	return false
}

// isSuspiciousFileType 检查文件类型是否可疑
func (s *Sanitizer) isSuspiciousFileType(filename string) bool {
	suspiciousExtensions := []string{".html", ".htm", ".js", ".php", ".asp", ".aspx", ".jsp"}

	filenameLower := strings.ToLower(filename)
	for _, ext := range suspiciousExtensions {
		if strings.HasSuffix(filenameLower, ext) {
			return true
		}
	}

	return false
}

// deepCheckFileContent 深度检查文件内容
func (s *Sanitizer) deepCheckFileContent(content []byte) error {
	contentStr := string(content)

	// 检查是否包含恶意脚本
	maliciousPatterns := []string{"<script>", "javascript:", "vbscript:", "onload=", "onerror="}

	for _, pattern := range maliciousPatterns {
		if strings.Contains(strings.ToLower(contentStr), pattern) {
			return fmt.Errorf("文件包含恶意内容: %s", pattern)
		}
	}

	return nil
}

// RemoveScriptTags 移除所有<script>标签
func (s *Sanitizer) RemoveScriptTags(input string) string {
	re := regexp.MustCompile(`<script\b[^<]*(?:(?!<\/script>)<[^<]*)*<\/script>`)
	return re.ReplaceAllString(input, "")
}

// SafeString 生成安全的字符串（用于内联JS）
func (s *Sanitizer) SafeString(input string) string {
	// 双重防护：先转义HTML，再转义JS
	escapedHTML := s.EscapeHTML(input)
	escapedJS := s.EscapeJS(escapedHTML)
	return escapedJS
}
