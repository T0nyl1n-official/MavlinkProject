package RateLimit

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// RateLimitAdmin 速率限制API访问
type RateLimitAdmin struct {
	middleware *RateLimitMiddleware
}

// NewRateLimitAdmin 创建速率限制管理API
func NewRateLimitAdmin(middleware *RateLimitMiddleware) *RateLimitAdmin {
	return &RateLimitAdmin{
		middleware: middleware,
	}
}

// GetRateLimitInfoRequest 获取速率限制信息请求
type GetRateLimitInfoRequest struct {
	Path       string `json:"path" form:"path"`
	Method     string `json:"method" form:"method"`
	Identifier string `json:"identifier" form:"identifier"`
}

// GetRateLimitInfo 获取速率限制信息
func (admin *RateLimitAdmin) GetRateLimitInfo(c *gin.Context) {
	var req GetRateLimitInfoRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "请求参数错误",
		})
		return
	}

	// 如果没有提供标识符，使用当前请求的标识符
	if req.Identifier == "" {
		req.Identifier = admin.middleware.getIdentifier(c, "ip")
	}

	info, err := admin.middleware.GetRateLimitInfo(c, req.Path, req.Method, req.Identifier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "获取速率限制信息失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"message": "获取速率限制信息成功",
		"data":    info,
	})
}

// ResetRateLimitRequest 重置速率限制请求
type ResetRateLimitRequest struct {
	Path       string `json:"path" binding:"required"`
	Method     string `json:"method" binding:"required"`
	Identifier string `json:"identifier" binding:"required"`
}

// ResetRateLimit 重置速率限制
func (admin *RateLimitAdmin) ResetRateLimit(c *gin.Context) {
	var req ResetRateLimitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "请求参数错误",
		})
		return
	}

	err := admin.middleware.ResetRateLimit(c, req.Path, req.Method, req.Identifier)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   err.Error(),
			"message": "重置速率限制失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"message": "重置速率限制成功",
	})
}

// UpdateRuleRequest 更新规则请求
type UpdateRuleRequest struct {
	RuleName string        `json:"rule_name" binding:"required"`
	NewRule  RateLimitRule `json:"new_rule" binding:"required"`
}

// UpdateRule 更新速率限制规则
func (admin *RateLimitAdmin) UpdateRule(c *gin.Context) {
	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "请求参数错误",
		})
		return
	}

	err := admin.middleware.UpdateRule(req.RuleName, req.NewRule)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   err.Error(),
			"message": "更新规则失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"message": "更新规则成功",
	})
}

// AddRuleRequest 添加规则请求
type AddRuleRequest struct {
	NewRule RateLimitRule `json:"new_rule" binding:"required"`
}

// AddRule 添加新的速率限制规则
func (admin *RateLimitAdmin) AddRule(c *gin.Context) {
	var req AddRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "请求参数错误",
		})
		return
	}

	admin.middleware.AddRule(req.NewRule)

	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"message": "添加规则成功",
	})
}

// RemoveRuleRequest 移除规则请求
type RemoveRuleRequest struct {
	RuleName string `json:"rule_name" binding:"required"`
}

// RemoveRule 移除速率限制规则
func (admin *RateLimitAdmin) RemoveRule(c *gin.Context) {
	var req RemoveRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "请求参数错误",
		})
		return
	}

	err := admin.middleware.RemoveRule(req.RuleName)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   err.Error(),
			"message": "移除规则失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"message": "移除规则成功",
	})
}

// GetConfig 获取当前配置
func (admin *RateLimitAdmin) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"message": "获取配置成功",
		"data":    admin.middleware.config,
	})
}

// EnableMiddlewareRequest 启用中间件请求
type EnableMiddlewareRequest struct {
	Enable bool `json:"enable" binding:"required"`
}

// ToggleMiddleware 切换中间件启用状态
func (admin *RateLimitAdmin) ToggleMiddleware(c *gin.Context) {
	var req EnableMiddlewareRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   err.Error(),
			"message": "请求参数错误",
		})
		return
	}

	if req.Enable {
		admin.middleware.Enable()
	} else {
		admin.middleware.Disable()
	}

	status := "启用"
	if !req.Enable {
		status = "禁用"
	}

	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"message": fmt.Sprintf("速率限制中间件已%s", status),
		"data": gin.H{
			"enabled": req.Enable,
		},
	})
}

// GetMiddlewareStatus 获取中间件状态
func (admin *RateLimitAdmin) GetMiddlewareStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"error":   nil,
		"message": "获取中间件状态成功",
		"data": gin.H{
			"enabled": admin.middleware.IsEnabled(),
		},
	})
}

// RegisterAdminRoutes 注册管理路由
func (admin *RateLimitAdmin) RegisterAdminRoutes(router *gin.RouterGroup) {
	// 需要管理员权限的路由组
	adminGroup := router.Group("/admin/rate-limit")
	{
		adminGroup.GET("/info", admin.GetRateLimitInfo)
		adminGroup.POST("/reset", admin.ResetRateLimit)
		adminGroup.PUT("/rule", admin.UpdateRule)
		adminGroup.POST("/rule", admin.AddRule)
		adminGroup.DELETE("/rule", admin.RemoveRule)
		adminGroup.GET("/config", admin.GetConfig)

		// 中间件开关控制
		adminGroup.POST("/toggle", admin.ToggleMiddleware)
		adminGroup.GET("/status", admin.GetMiddlewareStatus)
	}
}
