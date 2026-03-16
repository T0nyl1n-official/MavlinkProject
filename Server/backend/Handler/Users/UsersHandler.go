package UsersHandler

import (
	"fmt"
	"strconv"

	gin "github.com/gin-gonic/gin"
	gorm "gorm.io/gorm"

	Server "MavlinkProject/Server"
	ErrorsMgr "MavlinkProject/Server/backend/Middles/ErrorMiddleHandle/ErrorsMgr"
	User "MavlinkProject/Server/backend/Shared/User"
)

type UserHandler struct {
	Mysql *gorm.DB
}

var Backend = Server.BackendServer

func (h *UserHandler) RegisterUser(c *gin.Context) {
	user := &User.User{
		Username: c.PostForm("username"),
		Email:    c.PostForm("email"),
		Password: c.PostForm("password"),
	}

	if err := c.ShouldBindJSON(&user); err != nil {
		validationErrors := []ErrorsMgr.ValidationError{
			{Field: "general", Message: "用户数据格式错误"},
		}
		ErrorsMgr.HandleValidationErrors(c, validationErrors)
		return
	}

	// validations
	ValidEmailErr := ErrorsMgr.ValidateEmail(user.Email)
	ValidPasswordErr := ErrorsMgr.ValidatePassword(user.Password)

	if ValidEmailErr != nil {
		ErrorsMgr.HandleError(c, ValidEmailErr)
		return
	}
	if ValidPasswordErr != nil {
		ErrorsMgr.HandleError(c, ValidPasswordErr)
		return
	}

	// 隐藏密码
	user.HidePassword()

	ErrorsMgr.CreateSuccessResponse(c, gin.H{
		"User_ID":  user.ID,
		"Username": user.Username,
		"Email":    user.Email,
		"message":  "用户注册成功",
	})
}

func (h *UserHandler) LoginUser(c *gin.Context) {
	user := &User.User{
		Email:    c.PostForm("email"),
		Password: c.PostForm("password"),
	}
	// (鲁棒性) 检查JSON绑定是否成功
	if err := c.ShouldBindJSON(&user); err != nil {
		validationErrors := []ErrorsMgr.ValidationError{
			{Field: "general", Message: "用户数据格式错误"},
		}
		ErrorsMgr.HandleValidationErrors(c, validationErrors)
		return
	}

	// validations
	ValidEmailErr := ErrorsMgr.ValidateEmail(user.Email)
	ValidPasswordErr := ErrorsMgr.ValidatePassword(user.Password)

	if ValidEmailErr != nil {
		ErrorsMgr.HandleError(c, ValidEmailErr)
		return
	}
	if ValidPasswordErr != nil {
		ErrorsMgr.HandleError(c, ValidPasswordErr)
		return
	}

	user.HidePassword()

	ErrorsMgr.CreateSuccessResponse(c, gin.H{
		"User_ID":  user.ID,
		"Username": user.Username,
		"Email":    user.Email,
		"message":  "用户登录成功",
	})
}

func (h *UserHandler) GetUserInfo(c *gin.Context) {
	user := &User.User{}

	userIDStr := c.Param("userID")
	if userIDStr == "" {
		err := c.ShouldBindJSON(&user)
		if err != nil {
			validationErrors := []ErrorsMgr.ValidationError{
				{Field: "general", Message: "用户数据格式错误"},
			}
			ErrorsMgr.HandleValidationErrors(c, validationErrors)
			return
		}
	}
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		errorDetail := ErrorsMgr.GlobalCreateValidationError("userID", "用户ID格式无效")
		ErrorsMgr.HandleError(c, errorDetail)
		return
	}

	if validIDErr := ErrorsMgr.ValidateRequired(userIDStr, "userID"); validIDErr != nil {
		ErrorsMgr.HandleError(c, validIDErr)
		return
	}

	err = h.Mysql.First(&user, uint(userID)).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errorDetail := ErrorsMgr.GlobalNewError(ErrorsMgr.ErrUserNotFound, "用户不存在", nil)
			ErrorsMgr.HandleError(c, errorDetail)
		} else {
			errorDetail := ErrorsMgr.GlobalCreateDatabaseError("查询用户", "users", err)
			ErrorsMgr.HandleError(c, errorDetail)
		}
		return
	}

	user.HidePassword()

	ErrorsMgr.CreateSuccessResponse(c, gin.H{
		"User_ID":  user.ID,
		"Username": user.Username,
		"Email":    user.Email,
	})
}

// 更新用户信息 ()
func (h *UserHandler) UpdateUserInfo(c *gin.Context) {
	userID := c.Param("id")

	// 参数验证
	if idErr := ErrorsMgr.ValidateRequired("user_id", userID); idErr != nil {
		ErrorsMgr.HandleError(c, idErr)
		return
	}

	var user User.User

	// 参数绑定验证
	if err := c.ShouldBindJSON(&user); err != nil {
		validationErrors := []ErrorsMgr.ValidationError{
			{Field: "general", Message: "用户数据格式错误"},
		}
		ErrorsMgr.HandleValidationErrors(c, validationErrors)
		return
	}

	// 检查用户是否存在
	err := h.Mysql.First(&user, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errorDetail := ErrorsMgr.GlobalNewError(ErrorsMgr.ErrUserNotFound, "用户不存在", nil)
			ErrorsMgr.HandleError(c, errorDetail)
		} else {
			errorDetail := ErrorsMgr.GlobalCreateDatabaseError("查询用户", "users", err)
			ErrorsMgr.HandleError(c, errorDetail)
		}
		return
	}

	// 更新用户
	err = h.Mysql.Save(&user).Error
	if err != nil {
		errorDetail := ErrorsMgr.GlobalCreateDatabaseError("更新用户", "users", err)
		ErrorsMgr.HandleError(c, errorDetail)
		return
	}

	// 隐藏密码
	user.HidePassword()

	// 返回成功响应
	ErrorsMgr.CreateSuccessResponse(c, gin.H{
		"user":    user,
		"message": "用户更新成功",
	})
}

// 删除用户 (软删除)
func (h *UserHandler) DeleteUser(c *gin.Context) {
	var user User.User

	userID := c.Param("id")

	// validate userID
	if idErr := ErrorsMgr.ValidateRequired("user_id", userID); idErr != nil {
		ErrorsMgr.HandleError(c, idErr)
		return
	}

	// 检查用户是否存在
	err := h.Mysql.First(&user, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errorDetail := ErrorsMgr.GlobalNewError(ErrorsMgr.ErrUserNotFound, "用户不存在", nil)
			ErrorsMgr.HandleError(c, errorDetail)
		} else {
			errorDetail := ErrorsMgr.GlobalCreateDatabaseError("查询用户", "users", err)
			ErrorsMgr.HandleError(c, errorDetail)
		}
		return
	}

	// 删除用户 (gorm软删除)
	err = h.Mysql.Delete(&user).Error
	if err != nil {
		errorDetail := ErrorsMgr.GlobalCreateDatabaseError("删除用户", "users", err)
		ErrorsMgr.HandleError(c, errorDetail)
		return
	}

	// 隐藏密码
	user.HidePassword()

	// 返回成功响应
	ErrorsMgr.CreateSuccessResponse(c, gin.H{
		"user":    user,
		"message": "用户删除成功",
	})
}

// 获取全部用户信息
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	var queryUsers []User.User
	var user User.User

	if err := c.ShouldBindJSON(&user); err != nil {
		validationErrors := []ErrorsMgr.ValidationError{
			{Field: "general", Message: "用户数据格式错误"},
		}
		ErrorsMgr.HandleValidationErrors(c, validationErrors)
		return
	}

	// 查询所有用户
	err := h.Mysql.Find(&queryUsers).Error
	if err != nil {
		errorDetail := ErrorsMgr.GlobalCreateDatabaseError("查询所有用户", "users", err)
		ErrorsMgr.HandleError(c, errorDetail)
		return
	}

	// 隐藏所有用户的密码 (非管理员)
	if !user.IsAdmin() {
		for _, qUser := range queryUsers {
			qUser.HidePassword()
		}
	}

	// 返回成功响应
	ErrorsMgr.CreateSuccessResponse(c, gin.H{
		"users": queryUsers,
	})
}

// 发送邮件验证码 (忘记密码/修改密码/首次创建 通用)
func (h *UserHandler) SendEmailVerification(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Type  string `json:"type" binding:"required,oneof=register login reset_password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := []ErrorsMgr.ValidationError{
			{Field: "email", Message: "邮箱格式无效或类型不支持"},
		}
		ErrorsMgr.HandleValidationErrors(c, validationErrors)
		return
	}

	verification := Server.BackendServer.Verification
	if err := verification.SendVerificationCode(req.Email, req.Type); err != nil {
		ErrorsMgr.HandleError(c, fmt.Errorf("发送验证码失败: %w", err))
		return
	}

	ErrorsMgr.CreateSuccessResponse(c, gin.H{
		"message": "验证码已发送到您的邮箱",
		"email":   req.Email,
	})
}

func (h *UserHandler) LogoutUser(c *gin.Context) {
	var user User.User

	userID := c.Param("id")

	// validate userID
	if idErr := ErrorsMgr.ValidateRequired("user_id", userID); idErr != nil {
		ErrorsMgr.HandleError(c, idErr)
		return
	}

	// 检查用户是否存在
	err := h.Mysql.First(&user, userID).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errorDetail := ErrorsMgr.GlobalNewError(ErrorsMgr.ErrUserNotFound, "用户不存在", nil)
			ErrorsMgr.HandleError(c, errorDetail)
		} else {
			errorDetail := ErrorsMgr.GlobalCreateDatabaseError("查询用户", "users", err)
			ErrorsMgr.HandleError(c, errorDetail)
		}
		return
	}
}