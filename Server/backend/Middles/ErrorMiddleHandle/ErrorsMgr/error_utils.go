package ErrorsMgr

import (
	"fmt"
	"net/http"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// --错误响应--
// ErrorResponse 错误响应结构
type ErrorResponse struct {
	Success   bool        `json:"success"`
	Error     *ErrorInfo  `json:"err_info,omitempty"` // ErrorInfo 错误信息
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

// ErrorInfo 错误信息
type ErrorInfo struct {
	Code        string `json:"code"`
	Message     string `json:"message"`
	Description string `json:"description"`
	Module      string `json:"module"`
	Type        string `json:"type"`
}

// --验证--
// ValidationErrorResponse 验证错误响应
type ValidationErrorResponse struct {
	Success     bool              `json:"success"`
	Error       *ErrorInfo        `json:"error"`
	Validations []ValidationError `json:"validations,omitempty"`
	Timestamp   string            `json:"timestamp"`
}

// ValidationError 验证错误详情
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ErrorHandler 错误处理中间件
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			for _, ginErr := range c.Errors {
				err := ginErr.Err

				// 如果是自定义错误详情
				if detail, ok := err.(*ErrorDetail); ok {
					errorResponse := createErrorResponse(detail, c)
					c.JSON(detail.HTTPStatus, errorResponse)
					return
				}

				// 处理其他类型的错误
				handleGenericError(c, err)
				return
			}
		}

		// 处理HTTP状态码错误
		if c.Writer.Status() >= 400 {
			handleHTTPStatusError(c)
		}
	}
}

// HandleError 处理错误并返回响应
func HandleError(c *gin.Context, err error) {
	if detail, ok := err.(*ErrorDetail); ok {
		errorResponse := createErrorResponse(detail, c)
		c.JSON(detail.HTTPStatus, errorResponse)
		return
	}

	handleGenericError(c, err)
}

// HandleValidationErrors 处理验证错误
func HandleValidationErrors(c *gin.Context, validationErrors []ValidationError) {
	errorDetail := GlobalNewError(
		ErrValidationFailed,
		"请求参数验证失败",
		map[string]interface{}{
			"error_count": len(validationErrors),
			"errors":      validationErrors,
		},
	)

	response := ValidationErrorResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:        errorDetail.GetFormattedCode(),
			Message:     errorDetail.Message,
			Description: errorDetail.Description,
			Module:      errorDetail.Module,
			Type:        errorDetail.Type,
		},
		Validations: validationErrors,
		Timestamp:   errorDetail.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusBadRequest, response)
}

// CreateSuccessResponse 创建成功响应
func CreateSuccessResponse(c *gin.Context, data interface{}) {
	response := ErrorResponse{
		Success:   true,
		Data:      data,
		Timestamp: time.Now().Format("2006-01-02T15:04:05Z07:00"),
	}

	c.JSON(http.StatusOK, response)
}

// ValidateEmail 验证邮箱格式
func ValidateEmail(email string) *ErrorDetail {
	if email == "" {
		return GlobalCreateValidationError("email", "邮箱不能为空")
	}

	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
	re := regexp.MustCompile(emailRegex)
	if !re.MatchString(email) {
		return GlobalCreateValidationError("email", "邮箱格式无效")
	}

	return nil
}

// ValidatePhone 验证手机号格式
func ValidatePhone(phone string) *ErrorDetail {
	if phone == "" {
		return GlobalCreateValidationError("phone", "手机号不能为空")
	}

	phoneRegex := `^1[3-9]\d{9}$`
	re := regexp.MustCompile(phoneRegex)
	if !re.MatchString(phone) {
		return GlobalCreateValidationError("phone", "手机号格式无效")
	}

	return nil
}

// ValidatePassword 验证密码强度
func ValidatePassword(password string) *ErrorDetail {
	if len(password) < 8 {
		return GlobalCreateValidationError("password", "密码长度至少8位")
	}

	if !containsUpperAndLower(password) {
		return GlobalCreateValidationError("password", "密码必须包含大小写字母")
	}

	if !containsDigit(password) {
		return GlobalCreateValidationError("password", "密码必须包含数字")
	}

	return nil
}

// ValidateRequired 验证必填字段
func ValidateRequired(fieldName string, value interface{}) *ErrorDetail {
	if value == nil {
		return GlobalCreateValidationError(fieldName, "字段不能为空")
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return GlobalCreateValidationError(fieldName, "字段不能为空")
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		if strings.TrimSpace(v.String()) == "" {
			return GlobalCreateValidationError(fieldName, "字段不能为空")
		}
	case reflect.Slice, reflect.Array, reflect.Map:
		if v.Len() == 0 {
			return GlobalCreateValidationError(fieldName, "字段不能为空")
		}
	}

	return nil
}

// ValidateLength 验证字段长度
func ValidateLength(fieldName string, value string, min, max int) *ErrorDetail {
	length := len(value)

	if min > 0 && length < min {
		return GlobalCreateValidationError(fieldName, fmt.Sprintf("长度不能少于%d个字符", min))
	}

	if max > 0 && length > max {
		return GlobalCreateValidationError(fieldName, fmt.Sprintf("长度不能超过%d个字符", max))
	}

	return nil
}

// ValidateRange 验证数值范围
func ValidateRange(fieldName string, value, min, max int) *ErrorDetail {
	if value < min {
		return GlobalCreateValidationError(fieldName, fmt.Sprintf("值不能小于%d", min))
	}

	if value > max {
		return GlobalCreateValidationError(fieldName, fmt.Sprintf("值不能大于%d", max))
	}

	return nil
}

// 创建错误响应
func createErrorResponse(detail *ErrorDetail, c *gin.Context) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Error: &ErrorInfo{
			Code:        detail.GetFormattedCode(),
			Message:     detail.Message,
			Description: detail.Description,
			Module:      detail.Module,
			Type:        detail.Type,
		},
		Timestamp: detail.Timestamp.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// 处理通用错误
func handleGenericError(c *gin.Context, err error) {
	errorDetail := GlobalWrapError(
		ErrInternalServer,
		"服务器内部错误",
		err,
		nil,
	)

	response := createErrorResponse(errorDetail, c)
	c.JSON(http.StatusInternalServerError, response)
}

// 处理HTTP状态码错误
func handleHTTPStatusError(c *gin.Context) {
	status := c.Writer.Status()
	var code ErrorCode
	var message string

	switch status {
	case http.StatusBadRequest:
		code = ErrInvalidParameter
		message = "请求参数错误"
	case http.StatusUnauthorized:
		code = ErrAuthInvalidToken
		message = "未授权访问"
	case http.StatusForbidden:
		code = ErrAuthPermissionDenied
		message = "访问被拒绝"
	case http.StatusNotFound:
		code = ErrUserNotFound
		message = "资源不存在"
	case http.StatusConflict:
		code = ErrChainDuplicate
		message = "资源冲突"
	case http.StatusTooManyRequests:
		code = ErrTooManyRequests
		message = "请求过于频繁"
	default:
		code = ErrInternalServer
		message = "服务器错误"
	}

	errorDetail := GlobalNewError(code, message, nil)
	response := createErrorResponse(errorDetail, c)
	c.JSON(status, response)
}

// 辅助函数：检查是否包含大小写字母
func containsUpperAndLower(s string) bool {
	hasUpper := false
	hasLower := false

	for _, char := range s {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
		} else if char >= 'a' && char <= 'z' {
			hasLower = true
		}

		if hasUpper && hasLower {
			return true
		}
	}

	return false
}

// 辅助函数：检查是否包含数字
func containsDigit(s string) bool {
	for _, char := range s {
		if char >= '0' && char <= '9' {
			return true
		}
	}
	return false
}

// 辅助函数：检查是否包含特殊字符
func containsSpecialChar(s string) bool {
	specialChars := "!@#$%^&*()_+-=[]{}|;:,.<>?"
	for _, char := range s {
		if strings.ContainsRune(specialChars, char) {
			return true
		}
	}
	return false
}

// GetRequestID 获取请求ID（可以从上下文中获取或生成）
func GetRequestID(c *gin.Context) string {
	if requestID := c.GetHeader("X-Request-ID"); requestID != "" {
		return requestID
	}

	// 如果没有请求ID，可以生成一个简单的
	return fmt.Sprintf("req-%d", time.Now().UnixNano())
}
