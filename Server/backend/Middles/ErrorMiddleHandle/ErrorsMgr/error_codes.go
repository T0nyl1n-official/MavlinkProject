package ErrorsMgr

import (
	"fmt"
	"strconv"
)

type ErrorCode string

// 错误码定义 - 采用模块化编码方式
// 格式: XXYYZZ
// XX: 模块代码 (01-99)
// YY: 错误类型 (01-99)
// ZZ: 具体错误 (01-99)

// 模块代码定义
const (
	ModuleCommon     = "00" // 通用模块
	ModuleAuth       = "01" // 认证模块
	ModuleUser       = "02" // 用户模块
	ModuleChain      = "03" // 链模块
	ModuleChat       = "04" // 聊天模块
	ModuleMedia      = "05" // 媒体模块
	ModuleRateLimit  = "06" // 速率限制模块
	ModuleDatabase   = "07" // 数据库模块
	ModuleValidation = "08" // 验证模块
	ModuleFile       = "09" // 文件模块
	ModuleCode       = "10" // 生成码模块

	ModuleTerminal = "11" // 终端模块

	ModuleUnknown = "99" // 未知模块
)

// 错误类型定义
const (
	ErrorTypeValidation = "01" // 验证错误
	ErrorTypeAuth       = "02" // 认证错误
	ErrorTypePermission = "03" // 权限错误
	ErrorTypeNotFound   = "04" // 未找到错误
	ErrorTypeConflict   = "05" // 冲突错误
	ErrorTypeDatabase   = "06" // 数据库错误
	ErrorTypeExternal   = "07" // 外部服务错误
	ErrorTypeInternal   = "08" // 内部错误
	ErrorTypeRateLimit  = "09" // 速率限制错误
	ErrorTypeFile       = "10" // 文件错误

	ErrorTypeUnknown = "99" // 未知错误类型
)

// 具体错误内容码 - 在 errorManager.golang 内部署

// 通用错误码 (00XXXX)
const (
	// 通用验证错误 (0001XX)
	ErrValidationFailed      ErrorCode = "000101" // 验证失败
	ErrInvalidParameter      ErrorCode = "000102" // 参数无效
	ErrMissingParameter      ErrorCode = "000103" // 参数缺失
	ErrParameterTypeMismatch ErrorCode = "000104" // 参数类型不匹配

	// 通用系统错误 (0002XX)
	ErrInternalServer     ErrorCode = "000201" // 内部服务器错误
	ErrServiceUnavailable ErrorCode = "000202" // 服务不可用
	ErrTimeout            ErrorCode = "000203" // 请求超时
	ErrNotImplemented     ErrorCode = "000204" // 功能未实现

	// 通用业务错误 (0003XX)
	ErrOperationFailed   ErrorCode = "000301" // 操作失败
	ErrResourceExhausted ErrorCode = "000302" // 资源耗尽
	ErrTooManyRequests   ErrorCode = "000303" // 请求过多

	// 通用权限错误 (0004XX)
	ErrUnauthorized ErrorCode = "000401" // 未授权访问
)

// 认证模块错误码 (01XXXX)
const (
	// 认证验证错误 (0101XX)
	ErrAuthInvalidToken       ErrorCode = "010101" // Token无效
	ErrAuthTokenExpired       ErrorCode = "010102" // Token过期
	ErrAuthInvalidCredentials ErrorCode = "010103" // 凭证无效
	ErrAuthUserNotFound       ErrorCode = "010104" // 用户不存在

	// 认证权限错误 (0102XX)
	ErrAuthPermissionDenied ErrorCode = "010201" // 权限不足
	ErrAuthRoleRequired     ErrorCode = "010202" // 需要特定角色
	ErrAuthSessionExpired   ErrorCode = "010203" // 会话过期

	// 认证业务错误 (0103XX)
	ErrAuthLoginFailed    ErrorCode = "010301" // 登录失败
	ErrAuthRegisterFailed ErrorCode = "010302" // 注册失败
	ErrAuthLogoutFailed   ErrorCode = "010303" // 登出失败
)

// 用户模块错误码 (02XXXX)
const (
	// 用户验证错误 (0201XX)
	ErrUserInvalidEmail ErrorCode = "020101" // 邮箱格式无效
	ErrUserInvalidPhone ErrorCode = "020102" // 手机号格式无效
	ErrUserPasswordWeak ErrorCode = "020103" // 密码强度不足
	ErrUserEmailExists  ErrorCode = "020104" // 邮箱已存在
	ErrUserPhoneExists  ErrorCode = "020105" // 手机号已存在

	// 用户权限错误 (0202XX)
	ErrUserNotAuthorized ErrorCode = "020201" // 用户未授权
	ErrUserBanned        ErrorCode = "020202" // 用户被封禁
	ErrUserInactive      ErrorCode = "020203" // 用户未激活

	// 用户业务错误 (0203XX)
	ErrUserCreateFailed ErrorCode = "020301" // 用户创建失败
	ErrUserUpdateFailed ErrorCode = "020302" // 用户更新失败
	ErrUserDeleteFailed ErrorCode = "020303" // 用户删除失败
	ErrUserNotFound     ErrorCode = "020304" // 用户不存在
)

// 链模块错误码 (03XXXX)
const (
	// 链验证错误 (0301XX)
	ErrChainInvalidID   ErrorCode = "030101" // 链ID无效
	ErrChainInvalidData ErrorCode = "030102" // 链数据无效
	ErrChainDuplicate   ErrorCode = "030103" // 链重复
	ErrDecryptionFailed ErrorCode = "030104" // 解密失败

	// 链权限错误 (0302XX)
	ErrChainAccessDenied ErrorCode = "030201" // 链访问被拒绝
	ErrChainLocked       ErrorCode = "030202" // 链被锁定
	ErrChainNotAvailable ErrorCode = "030203" // 链不可用

	// 链业务错误 (0303XX)
	ErrChainCreateFailed ErrorCode = "030301" // 链创建失败
	ErrChainUpdateFailed ErrorCode = "030302" // 链更新失败
	ErrChainDeleteFailed ErrorCode = "030303" // 链删除失败
	ErrChainNotFound     ErrorCode = "030304" // 链不存在
)

// 聊天模块错误码 (04XXXX)
const (
	// 聊天验证错误 (0401XX)
	ErrChatInvalidMessage     ErrorCode = "040101" // 消息无效
	ErrChatInvalidRecipient   ErrorCode = "040102" // 收件人无效
	ErrChatInvalidMessageType ErrorCode = "040103" // 消息类型无效

	// 聊天权限错误 (0402XX)
	ErrChatSendDenied ErrorCode = "040201" // 发送消息被拒绝
	ErrChatReadDenied ErrorCode = "040202" // 读取消息被拒绝

	// 聊天业务错误 (0403XX)
	ErrChatSendFailed           ErrorCode = "040301" // 发送消息失败
	ErrChatMessageNotFound      ErrorCode = "040302" // 消息不存在
	ErrChatRoomFull             ErrorCode = "040303" // 聊天室已满
	ErrChatConversationNotFound ErrorCode = "040304" // 对话不存在
)

// 媒体模块错误码 (05XXXX)
const (
	// 媒体验证错误 (0501XX)
	ErrMediaInvalidFile     ErrorCode = "050101" // 文件无效
	ErrMediaUnsupportedType ErrorCode = "050102" // 不支持的文件类型
	ErrMediaFileTooLarge    ErrorCode = "050103" // 文件过大
	ErrMediaFileNotFound    ErrorCode = "050104" // 文件不存在

	// 媒体权限错误 (0502XX)
	ErrMediaUploadDenied ErrorCode = "050201" // 上传被拒绝
	ErrMediaAccessDenied ErrorCode = "050202" // 访问被拒绝

	// 媒体业务错误 (0503XX)
	ErrMediaUploadFailed  ErrorCode = "050301" // 上传失败
	ErrMediaProcessFailed ErrorCode = "050302" // 处理失败
	ErrMediaNotFound      ErrorCode = "050303" // 媒体不存在
)

// 速率限制模块错误码 (06XXXX)
const (
	ErrRateLimitExceeded ErrorCode = "060101" // 速率限制超出
	ErrRateLimitDisabled ErrorCode = "060102" // 速率限制已禁用
)

// 数据库模块错误码 (07XXXX)
const (
	ErrDatabaseConnection  ErrorCode = "070101" // 数据库连接失败
	ErrDatabaseQuery       ErrorCode = "070102" // 数据库查询失败
	ErrDatabaseTransaction ErrorCode = "070103" // 数据库事务失败
)

// 验证模块错误码 (08XXXX)
const (
	ErrVerificationCodeInvalid ErrorCode = "080101" // 验证码无效
	ErrVerificationCodeExpired ErrorCode = "080102" // 验证码过期
	ErrVerificationSendFailed  ErrorCode = "080103" // 验证码发送失败
)

// 文件模块错误码 (09XXXX)
const (
	ErrFileInvalid      ErrorCode = "090101" // 文件生成无效
	ErrFilePathInvalid  ErrorCode = "090102" // 文件路径无效
	ErrFileNotFound     ErrorCode = "090103" // 文件不存在
	ErrFileReadFailed   ErrorCode = "090104" // 文件读取失败
	ErrFileWriteFailed  ErrorCode = "090105" // 文件写入失败
	ErrFilePermission   ErrorCode = "090106" // 文件权限错误
	ErrFileSizeExceeded ErrorCode = "090107" // 文件大小超出限制
	ErrFileCorrupted    ErrorCode = "090108" // 文件损坏
)

// 生成码模块错误码 (10XXXX)
const (
	ErrCodeGenerationFailed       ErrorCode = "100101" // 生成码失败
	ErrCodeInvalid                ErrorCode = "100102" // 生成码无效
	ErrCodeExpired                ErrorCode = "100103" // 生成码过期
	ErrCodeUsed                   ErrorCode = "100104" // 生成码已使用
	ErrCodeDuplicate              ErrorCode = "100105" // 生成码重复
	ErrCodeGenerationLimit        ErrorCode = "100106" // 生成码超出限制
	ErrCodeGenerationRateExceeded ErrorCode = "100107" // 生成码超出速率限制

)

// 终端模块错误码 (11XXXX)
const (
	// 终端验证错误 (1101XX)
	ErrTerminalInvalidSession ErrorCode = "110101" // 会话无效
	ErrTerminalInvalidCommand ErrorCode = "110102" // 命令无效

	// 终端权限错误 (1102XX)
	ErrTerminalAccessDenied ErrorCode = "110201" // 终端访问被拒绝

	// 终端业务错误 (1103XX)
	ErrTerminalSessionNotFound ErrorCode = "110301" // 会话不存在或无权访问
	ErrTerminalSessionClosed   ErrorCode = "110302" // 会话已关闭
	ErrTerminalCommandFailed   ErrorCode = "110303" // 命令执行失败
)

// GetModuleName 获取模块名称
func GetModuleName(code ErrorCode) string {
	if len(code) < 2 {
		return "未知模块"
	}
	module := string(code)[:2]
	switch module {
	case ModuleCommon:
		return "通用模块"
	case ModuleAuth:
		return "认证模块"
	case ModuleUser:
		return "用户模块"
	case ModuleChain:
		return "链模块"
	case ModuleChat:
		return "聊天模块"
	case ModuleMedia:
		return "媒体模块"
	case ModuleRateLimit:
		return "速率限制模块"
	case ModuleDatabase:
		return "数据库模块"
	case ModuleValidation:
		return "验证模块"
	case ModuleFile:
		return "文件模块"
	case ModuleCode:
		return "生成码模块"
	case ModuleTerminal:
		return "终端模块"
	default:
		return "未知模块"
	}
}

// GetErrorTypeName 获取错误类型名称
func GetErrorTypeName(code ErrorCode) string {
	if len(code) < 4 {
		return "未知错误类型"
	}
	typeCode := string(code)[2:4]
	switch typeCode {
	case ErrorTypeValidation:
		return "验证错误"
	case ErrorTypeAuth:
		return "认证错误"
	case ErrorTypePermission:
		return "权限错误"
	case ErrorTypeNotFound:
		return "未找到错误"
	case ErrorTypeConflict:
		return "冲突错误"
	case ErrorTypeDatabase:
		return "数据库错误"
	case ErrorTypeExternal:
		return "外部服务错误"
	case ErrorTypeInternal:
		return "内部错误"
	case ErrorTypeRateLimit:
		return "速率限制错误"
	case ErrorTypeFile:
		return "文件错误"
	default:
		return "未知错误类型"
	}
}

// GetErrorDescription 获取错误描述
func GetErrorDescription(code ErrorCode) string {
	switch code {
	// 通用错误
	case ErrValidationFailed:
		return "请求参数验证失败"
	case ErrInvalidParameter:
		return "参数格式无效"
	case ErrMissingParameter:
		return "缺少必要参数"
	case ErrParameterTypeMismatch:
		return "参数类型不匹配"
	case ErrInternalServer:
		return "服务器内部错误"
	case ErrServiceUnavailable:
		return "服务暂时不可用"
	case ErrTimeout:
		return "请求超时"
	case ErrNotImplemented:
		return "功能尚未实现"
	case ErrOperationFailed:
		return "操作执行失败"
	case ErrResourceExhausted:
		return "系统资源耗尽"
	case ErrTooManyRequests:
		return "请求过于频繁"

	// 认证错误
	case ErrAuthInvalidToken:
		return "认证令牌无效"
	case ErrAuthTokenExpired:
		return "认证令牌已过期"
	case ErrAuthInvalidCredentials:
		return "用户名或密码错误"
	case ErrAuthUserNotFound:
		return "用户不存在"
	case ErrAuthPermissionDenied:
		return "权限不足"
	case ErrAuthRoleRequired:
		return "需要特定角色权限"
	case ErrAuthSessionExpired:
		return "会话已过期"
	case ErrAuthLoginFailed:
		return "登录失败"
	case ErrAuthRegisterFailed:
		return "注册失败"
	case ErrAuthLogoutFailed:
		return "登出失败"

	// 用户错误
	case ErrUserInvalidEmail:
		return "邮箱格式无效"
	case ErrUserInvalidPhone:
		return "手机号格式无效"
	case ErrUserPasswordWeak:
		return "密码强度不足"
	case ErrUserEmailExists:
		return "邮箱已被注册"
	case ErrUserPhoneExists:
		return "手机号已被注册"
	case ErrUserNotAuthorized:
		return "用户未授权"
	case ErrUserBanned:
		return "用户已被封禁"
	case ErrUserInactive:
		return "用户账户未激活"
	case ErrUserCreateFailed:
		return "用户创建失败"
	case ErrUserUpdateFailed:
		return "用户更新失败"
	case ErrUserDeleteFailed:
		return "用户删除失败"
	case ErrUserNotFound:
		return "用户不存在"

	// 链错误
	case ErrChainInvalidID:
		return "链ID格式无效"
	case ErrChainInvalidData:
		return "链数据格式无效"
	case ErrChainDuplicate:
		return "链已存在"
	case ErrDecryptionFailed:
		return "解密失败"
	case ErrChainAccessDenied:
		return "无权访问该链"
	case ErrChainLocked:
		return "链已被锁定"
	case ErrChainNotAvailable:
		return "链不可用"
	case ErrChainCreateFailed:
		return "链创建失败"
	case ErrChainUpdateFailed:
		return "链更新失败"
	case ErrChainDeleteFailed:
		return "链删除失败"
	case ErrChainNotFound:
		return "链不存在"

	// 聊天错误
	case ErrChatInvalidMessage:
		return "聊天消息格式无效"
	case ErrChatInvalidRecipient:
		return "聊天接收人格式无效"
	case ErrChatSendDenied:
		return "发送消息被拒绝"
	case ErrChatReadDenied:
		return "读取消息被拒绝"
	case ErrChatSendFailed:
		return "发送消息失败"
	case ErrChatMessageNotFound:
		return "消息不存在"
	case ErrChatRoomFull:
		return "聊天房间已满"

	// 媒体传输错误
	case ErrMediaInvalidFile:
		return "媒体文件格式无效"
	case ErrMediaUnsupportedType:
		return "媒体形式不支持"
	case ErrMediaFileTooLarge:
		return "媒体文件过大"
	case ErrMediaFileNotFound:
		return "媒体文件不存在"
	case ErrMediaUploadDenied:
		return "上传被拒绝"
	case ErrMediaAccessDenied:
		return "访问被拒绝"
	case ErrMediaUploadFailed:
		return "上传失败"
	case ErrMediaProcessFailed:
		return "媒体处理失败"
	case ErrMediaNotFound:
		return "媒体不存在"

	default:
		return "未知错误"
	}
}

// GetHTTPStatusCode 获取对应的HTTP状态码
func GetHTTPStatusCode(code ErrorCode) int {
	if len(code) < 4 {
		return 500
	}
	typeCode := string(code)[2:4]

	switch typeCode {
	case ErrorTypeValidation, ErrorTypeAuth:
		return 400 // Bad Request
	case ErrorTypePermission:
		return 403 // Forbidden
	case ErrorTypeNotFound:
		return 404 // Not Found
	case ErrorTypeConflict:
		return 409 // Conflict
	case ErrorTypeRateLimit:
		return 429 // Too Many Requests
	case ErrorTypeDatabase, ErrorTypeExternal, ErrorTypeInternal:
		return 500 // Internal Server Error
	default:
		return 500 // Internal Server Error
	}
}

// FormatErrorCode 格式化错误码为字符串
func FormatErrorCode(code ErrorCode) string {
	// 尝试将错误码转换为数字，如果失败则直接返回
	if _, err := strconv.Atoi(string(code)); err == nil {
		return fmt.Sprintf("ERR%s", code)
	}
	return fmt.Sprintf("ERR%s", code)
}
