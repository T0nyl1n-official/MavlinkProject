package Verification

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gopkg.in/gomail.v2"
)

// 定义全局context
var ctx = context.Background()

// VerificationConfig 验证码配置结构体
type VerificationConfig struct {
	// Redis配置
	RedisAddr     string
	RedisPassword string
	RedisDB       int

	// 邮箱配置
	SMTPHost     string
	SMTPPort     int
	SMTPUsername string
	SMTPPassword string
	FromEmail    string
	FromName     string

	// 验证码配置
	CodeLength   int
	Expiration   time.Duration // 验证码过期时间
	MaxAttempts  int           // 最大尝试次数
	CoolDownTime time.Duration // 重复发送冷却时间
}

// Redis中存储的验证码结构
type verificationCode struct {
	Code      string    `json:"code"`
	Email     string    `json:"email"`
	Type      string    `json:"type"`
	Attempts  int       `json:"attempts"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// 请求结构体
type SendCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Type  string `json:"type" binding:"required,oneof=register login reset_password"`
}

type VerifyCodeRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required"`
	Type  string `json:"type" binding:"required,oneof=register login reset_password"`
}

// 响应结构体
type Response struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// VerificationManager 验证管理器
type VerificationManager struct {
	redisClient        *redis.Client
	VerificationConfig *VerificationConfig
}

func (cfg *VerificationConfig) Default() {
	cfg.RedisAddr = getEnvOrDefault("SMTP_REDIS_ADDR", "localhost:6379")
	cfg.RedisPassword = getEnvOrDefault("SMTP_REDIS_PASSWORD", "")
	cfg.RedisDB = 15
	cfg.SMTPHost = getEnvOrDefault("SMTP_HOST", "smtp.qq.com")
	cfg.SMTPPort = 587
	cfg.CodeLength = 5
	cfg.Expiration = 5 * 60 * time.Minute
	cfg.MaxAttempts = 20
	cfg.CoolDownTime = 60 * time.Second

	cfg.SMTPUsername = getEnvOrDefault("SMTP_USERNAME", "")
	cfg.SMTPPassword = getEnvOrDefault("SMTP_PASSWORD", "")

	cfg.FromEmail = getEnvOrDefault("SMTP_FROM_EMAIL", "")
	cfg.FromName = getEnvOrDefault("SMTP_FROM_NAME", "MavlinkProject")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return defaultValue
}

// 验证管理器
func NewVerificationManager(cfg *VerificationConfig) (*VerificationManager, error) {
	// 设置 验证码长度, 作废时间, 最大尝试次数, 冷却时间 参数
	if cfg.CodeLength == 0 {
		cfg.CodeLength = 6
	}
	if cfg.Expiration == 0 {
		cfg.Expiration = 5 * time.Minute
	}
	if cfg.MaxAttempts == 0 {
		cfg.MaxAttempts = 5
	}
	if cfg.CoolDownTime == 0 {
		cfg.CoolDownTime = 1 * time.Minute
	}

	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// 测试Redis连接
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("verification: failed to connect to Redis: %v", err)
	}

	return &VerificationManager{
		redisClient:        redisClient,
		VerificationConfig: cfg,
	}, nil
}

// ==================== 核心功能方法 ====================

// SendVerificationCode 发送验证码（函数形式）
func (sv *VerificationManager) SendVerificationCode(email, codeType string) error {
	// 检查发送频率
	if err := sv.checkCoolDown(email, codeType); err != nil {
		return err
	}

	// 生成验证码
	code, err := sv.generateCode()
	if err != nil {
		fmt.Println("Verification: 生成验证码失败: ", err)
		return fmt.Errorf("verification: 生成验证码失败: %v", err)
	}

	// 发送邮件
	if err := sv.sendCodeEmail(email, code, codeType); err != nil {
		fmt.Println("Verification: 发送邮件失败: ", err)
		return fmt.Errorf("verification: 发送邮件失败: %v", err)
	}

	// 存储验证码
	if err := sv.storeCode(email, code, codeType); err != nil {
		fmt.Println("Verification: 存储验证码失败: ", err)
		return fmt.Errorf("verification: 存储验证码失败: %v", err)
	}

	return nil
}

// VerifyVerificationCode 验证验证码（函数形式）
func (sv *VerificationManager) VerifyVerificationCode(email, code, codeType string) (bool, error) {
	// 获取验证码
	storedCode, err := sv.getCode(email, codeType)
	if err != nil {
		return false, errors.New("验证码不存在或已过期")
	}

	// 检查是否过期
	if time.Now().After(storedCode.ExpiresAt) {
		sv.deleteCode(email, codeType)
		return false, errors.New("验证码已过期")
	}

	// 检查尝试次数
	if storedCode.Attempts >= sv.VerificationConfig.MaxAttempts {
		sv.deleteCode(email, codeType)
		return false, errors.New("验证码尝试次数过多，请重新获取")
	}

	// 验证验证码
	if storedCode.Code != code {
		sv.incrementAttempts(email, codeType)
		return false, errors.New("验证码不正确")
	}

	// 验证成功，删除验证码
	sv.deleteCode(email, codeType)
	return true, nil
}

// ==================== HTTP API 接口 ====================
// SendCodeHandler 发送验证码API接口
func (sv *VerificationManager) SendCodeHandler(c *gin.Context) {
	var req SendCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 发送验证码
	if err := sv.SendVerificationCode(req.Email, req.Type); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Success: false,
			Message: "发送验证码失败",
			Error:   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "验证码发送成功",
	})
}

// VerifyCodeHandler 验证验证码API接口
func (sv *VerificationManager) VerifyCodeHandler(c *gin.Context) {
	var req VerifyCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "请求参数错误",
			Error:   err.Error(),
		})
		return
	}

	// 验证验证码
	isValid, err := sv.VerifyVerificationCode(req.Email, req.Code, req.Type)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	if isValid {
		c.JSON(http.StatusOK, Response{
			Success: true,
			Message: "验证成功",
		})
	} else {
		c.JSON(http.StatusBadRequest, Response{
			Success: false,
			Message: "验证失败",
		})
	}
}

// ==================== 内部工具方法 ====================

// 生成随机验证码
func (sv *VerificationManager) generateCode() (string, error) {
	const digits = "0123456789"
	code := make([]byte, sv.VerificationConfig.CodeLength)

	for i := range code {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		code[i] = digits[num.Int64()]
	}

	return string(code), nil
}

// 发送邮件
func (sv *VerificationManager) sendCodeEmail(email, code, codeType string) error {
	subject := sv.getSubjectByType(codeType)
	htmlBody := sv.generateEmailHTML(code, codeType)

	codeMessage := gomail.NewMessage()
	codeMessage.SetHeader("From", codeMessage.FormatAddress(sv.VerificationConfig.FromEmail, sv.VerificationConfig.FromName))
	codeMessage.SetHeader("To", email)
	codeMessage.SetHeader("Subject", subject)
	codeMessage.SetBody("text/html", htmlBody)

	d := gomail.NewDialer(
		sv.VerificationConfig.SMTPHost,
		sv.VerificationConfig.SMTPPort,
		sv.VerificationConfig.SMTPUsername,
		sv.VerificationConfig.SMTPPassword,
	)

	return d.DialAndSend(codeMessage)
}

// 存储验证码到Redis
func (sv *VerificationManager) storeCode(email, code, codeType string) error {
	verificationCode := &verificationCode{
		Code:      code,
		Email:     strings.ToLower(email),
		Type:      codeType,
		Attempts:  0,
		ExpiresAt: time.Now().Add(sv.VerificationConfig.Expiration),
		CreatedAt: time.Now(),
	}

	data, err := json.Marshal(verificationCode)
	if err != nil {
		return err
	}

	key := sv.getRedisKey(email, codeType)
	return sv.redisClient.Set(ctx, key, data, sv.VerificationConfig.Expiration).Err()
}

// 从Redis获取验证码
func (sv *VerificationManager) getCode(email, codeType string) (*verificationCode, error) {
	key := sv.getRedisKey(email, codeType)
	data, err := sv.redisClient.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var code verificationCode
	err = json.Unmarshal(data, &code)
	return &code, err
}

// 删除验证码
func (sv *VerificationManager) deleteCode(email, codeType string) error {
	key := sv.getRedisKey(email, codeType)
	return sv.redisClient.Del(ctx, key).Err()
}

// 增加尝试次数
func (sv *VerificationManager) incrementAttempts(email, codeType string) error {
	code, err := sv.getCode(email, codeType)
	if err != nil {
		return err
	}

	code.Attempts++
	data, err := json.Marshal(code)
	if err != nil {
		return err
	}

	key := sv.getRedisKey(email, codeType)
	ttl := sv.redisClient.TTL(ctx, key).Val()
	return sv.redisClient.Set(ctx, key, data, ttl).Err()
}

// 检查发送频率
func (sv *VerificationManager) checkCoolDown(email, codeType string) error {
	code, err := sv.getCode(email, codeType)
	if err != nil {
		return nil // 没有记录，可以发送
	}

	if time.Since(code.CreatedAt) < sv.VerificationConfig.CoolDownTime {
		return errors.New("发送频率过高，请稍后再试")
	}

	return nil
}

// 获取Redis键名
func (sv *VerificationManager) getRedisKey(email, codeType string) string {
	return fmt.Sprintf("verification:%s:%s", codeType, strings.ToLower(email))
}

// 根据类型获取邮件主题
func (sv *VerificationManager) getSubjectByType(codeType string) string {
	switch codeType {
	case "register":
		return "注册验证码"
	case "login":
		return "登录验证码"
	case "reset_password":
		return "重置密码验证码"
	default:
		return "验证码"
	}
}

// 生成邮件HTML内容
func (sv *VerificationManager) generateEmailHTML(code, codeType string) string {
	actionText := sv.getActionTextByType(codeType)

	return fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8">
    <style>
        .container { max-width: 600px; margin: 0 auto; padding: 30px; font-family: Arial, sans-serif; border: 1px solid #e0e0e0; border-radius: 10px; }
        .header { text-align: center; color: #333; margin-bottom: 30px; }
        .code { font-size: 32px; font-weight: bold; color: #1890ff; padding: 20px; background: #f5f5f5; text-align: center; margin: 30px 0; border-radius: 5px; letter-spacing: 5px; }
        .warning { color: #ff4d4f; font-size: 14px; margin-top: 20px; padding: 10px; background: #fff2f0; border-radius: 5px; }
        .footer { margin-top: 30px; padding-top: 20px; border-top: 1px solid #e0e0e0; color: #666; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Project G.I.T <strong>邮箱验证</strong></h1>
        </div>
        <p>您好!</p>
        <p>您正在进行%s操作, 验证码为: </p>
        <div class="code">%s</div>
        <p>验证码有效期%.0f分钟，请及时使用。</p>
        <div class="warning">
            <strong>安全提示：</strong>请勿将此验证码透露给他人！如非本人操作，请忽略此邮件。
        </div>
        <div class="footer">
            <p>此为系统邮件，请勿回复</p>
            <p>© 2026 - %d MavlinkProject Verification By Tonyl1n - All Rights Reserved </p>
        </div>
    </div>
</body>
</html>
    `, actionText, code, sv.VerificationConfig.Expiration.Minutes(), time.Now().Year())
}

// 根据类型获取操作文本
func (sv *VerificationManager) getActionTextByType(codeType string) string {
	switch codeType {
	case "register":
		return "账号注册"
	case "login":
		return "登录"
	case "reset_password":
		return "重置密码"
	default:
		return "验证"
	}
}

// Verification by TonyL1n
// Using in MavlinkProject
