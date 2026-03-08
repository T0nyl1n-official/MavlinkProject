package ErrorsMgr

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
	"time"
)

// ErrorDetail 错误详情结构
type ErrorDetail struct {
	Code        ErrorCode          `json:"code"`
	Message     string             `json:"message"`
	Description string             `json:"description"`
	Module      string             `json:"module"`
	Type        string             `json:"type"`
	HTTPStatus  int                `json:"http_status"`
	Timestamp   time.Time          `json:"timestamp"`
	StackTrace  []string           `json:"stack_trace,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	InnerError  *ErrorDetail       `json:"inner_error,omitempty"`
}

// ErrorManager 错误管理器
type ErrorManager struct {
	enableStackTrace bool
	enableLogging    bool
	logFile         string
}

// NewErrorManager 创建错误管理器
func NewErrorManager(enableStackTrace, enableLogging bool, logFile string) *ErrorManager {
	return &ErrorManager{
		enableStackTrace: enableStackTrace,
		enableLogging:    enableLogging,
		logFile:         logFile,
	}
}

// NewError 创建新的错误详情
func (em *ErrorManager) NewError(code ErrorCode, message string, context map[string]interface{}) *ErrorDetail {
	err := &ErrorDetail{
		Code:        code,
		Message:     message,
		Description: GetErrorDescription(code),
		Module:      GetModuleName(code),
		Type:        GetErrorTypeName(code),
		HTTPStatus:  GetHTTPStatusCode(code),
		Timestamp:   time.Now(),
		Context:     context,
	}
	
	if em.enableStackTrace {
		err.StackTrace = em.getStackTrace()
	}
	
	if em.enableLogging {
		em.logError(err)
	}
	
	return err
}

// WrapError 包装现有错误
func (em *ErrorManager) WrapError(code ErrorCode, message string, innerError error, context map[string]interface{}) *ErrorDetail {
	err := em.NewError(code, message, context)
	
	if innerError != nil {
		if detail, ok := innerError.(*ErrorDetail); ok {
			err.InnerError = detail
		} else {
			err.InnerError = &ErrorDetail{
				Code:        ErrInternalServer,
				Message:     innerError.Error(),
				Description: "内部错误",
				Module:      "通用模块",
				Type:        "内部错误",
				HTTPStatus:  500,
				Timestamp:   time.Now(),
			}
		}
	}
	
	return err
}

// CreateValidationError 创建验证错误
func (em *ErrorManager) CreateValidationError(field string, reason string) *ErrorDetail {
	context := map[string]interface{}{
		"field":  field,
		"reason": reason,
	}
	
	return em.NewError(ErrValidationFailed, fmt.Sprintf("字段 '%s' 验证失败: %s", field, reason), context)
}

// CreateAuthError 创建认证错误
func (em *ErrorManager) CreateAuthError(code ErrorCode, operation string) *ErrorDetail {
	context := map[string]interface{}{
		"operation": operation,
	}
	
	return em.NewError(code, fmt.Sprintf("认证操作 '%s' 失败", operation), context)
}

// CreateDatabaseError 创建数据库错误
func (em *ErrorManager) CreateDatabaseError(operation string, table string, err error) *ErrorDetail {
	context := map[string]interface{}{
		"operation": operation,
		"table":    table,
	}
	
	return em.WrapError(ErrDatabaseQuery, fmt.Sprintf("数据库操作 '%s' 失败", operation), err, context)
}

// CreateFileError 创建文件错误
func (em *ErrorManager) CreateFileError(code ErrorCode, operation string, filePath string, err error) *ErrorDetail {
	context := map[string]interface{}{
		"operation": operation,
		"file_path": filePath,
	}
	
	return em.WrapError(code, fmt.Sprintf("文件操作 '%s' 失败", operation), err, context)
}

// ToJSON 将错误详情转换为JSON
func (ed *ErrorDetail) ToJSON() string {
	jsonData, err := json.MarshalIndent(ed, "", "  ")
	if err != nil {
		return fmt.Sprintf(`{"error": "序列化错误详情失败: %v"}`, err)
	}
	return string(jsonData)
}

// ToSimpleJSON 转换为简化的JSON响应
func (ed *ErrorDetail) ToSimpleJSON() map[string]interface{} {
	return map[string]interface{}{
		"error": map[string]interface{}{
			"code":        FormatErrorCode(ed.Code),
			"message":     ed.Message,
			"description": ed.Description,
			"module":      ed.Module,
			"type":        ed.Type,
		},
		"timestamp": ed.Timestamp.Format(time.RFC3339),
	}
}

// Error 实现error接口
func (ed *ErrorDetail) Error() string {
	return fmt.Sprintf("[%s] %s: %s", FormatErrorCode(ed.Code), ed.Module, ed.Message)
}

// GetFormattedCode 获取格式化错误码
func (ed *ErrorDetail) GetFormattedCode() string {
	return FormatErrorCode(ed.Code)
}

// 获取堆栈跟踪
func (em *ErrorManager) getStackTrace() []string {
	stack := make([]string, 0)
	
	// 跳过前3个调用（getStackTrace, NewError, 调用者）
	for i := 3; ; i++ {
		pc, file, line, ok := runtime.Caller(i)
		if !ok {
			break
		}
		
		// 获取函数名
		fn := runtime.FuncForPC(pc)
		funcName := "unknown"
		if fn != nil {
			funcName = fn.Name()
		}
		
		// 简化文件路径
		parts := strings.Split(file, "/")
		if len(parts) > 3 {
			file = strings.Join(parts[len(parts)-3:], "/")
		}
		
		stack = append(stack, fmt.Sprintf("%s:%d %s", file, line, funcName))
	}
	
	return stack
}

// 记录错误到日志
func (em *ErrorManager) logError(err *ErrorDetail) {
	// 这里可以集成到项目的日志系统
	// 目前先简单输出到控制台
	logEntry := fmt.Sprintf("[ERROR] %s - %s\n", 
		err.Timestamp.Format("2006-01-02 15:04:05"), 
		err.Error())
	
	if len(err.StackTrace) > 0 {
		logEntry += "Stack Trace:\n"
		for _, frame := range err.StackTrace {
			logEntry += fmt.Sprintf("  %s\n", frame)
		}
	}
	
	fmt.Print(logEntry)
}

// 全局错误管理器实例
var globalErrorManager *ErrorManager

// InitGlobalErrorManager 初始化全局错误管理器
func InitGlobalErrorManager(enableStackTrace, enableLogging bool, logFile string) {
	globalErrorManager = NewErrorManager(enableStackTrace, enableLogging, logFile)
}

// GetGlobalErrorManager 获取全局错误管理器
func GetGlobalErrorManager() *ErrorManager {
	if globalErrorManager == nil {
		// 使用默认配置
		globalErrorManager = NewErrorManager(true, true, "")
	}
	return globalErrorManager
}

// GlobalNewError 全局创建错误
func GlobalNewError(code ErrorCode, message string, context map[string]interface{}) *ErrorDetail {
	return GetGlobalErrorManager().NewError(code, message, context)
}

// GlobalWrapError 全局包装错误
func GlobalWrapError(code ErrorCode, message string, innerError error, context map[string]interface{}) *ErrorDetail {
	return GetGlobalErrorManager().WrapError(code, message, innerError, context)
}

// GlobalCreateValidationError 全局创建验证错误
func GlobalCreateValidationError(field string, reason string) *ErrorDetail {
	return GetGlobalErrorManager().CreateValidationError(field, reason)
}

// GlobalCreateAuthError 全局创建认证错误
func GlobalCreateAuthError(code ErrorCode, operation string) *ErrorDetail {
	return GetGlobalErrorManager().CreateAuthError(code, operation)
}

// GlobalCreateDatabaseError 全局创建数据库错误
func GlobalCreateDatabaseError(operation string, table string, err error) *ErrorDetail {
	return GetGlobalErrorManager().CreateDatabaseError(operation, table, err)
}

// GlobalCreateFileError 全局创建文件错误
func GlobalCreateFileError(code ErrorCode, operation string, filePath string, err error) *ErrorDetail {
	return GetGlobalErrorManager().CreateFileError(code, operation, filePath, err)
}