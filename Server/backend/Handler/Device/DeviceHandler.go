package Device

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	ErrorsMgr "MavlinkProject/Server/backend/Middles/ErrorMiddleHandle/ErrorsMgr"
	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
	Device "MavlinkProject/Server/backend/Shared/Device"
)

type DeviceHandler struct {
	Mysql      *gorm.DB
	JWTManager *jwtUtils.JWTManager
}

func (h *DeviceHandler) LoginDevice(c *gin.Context) {
	var req struct {
		DeviceID   string `json:"device_id" binding:"required"`
		DeviceKey  string `json:"device_key" binding:"required"`
		DeviceType string `json:"device_type" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		validationErrors := []ErrorsMgr.ValidationError{
			{Field: "general", Message: "设备数据格式错误"},
		}
		ErrorsMgr.HandleValidationErrors(c, validationErrors)
		return
	}

	device := &Device.Device{}
	err := h.Mysql.Where("device_id = ? AND device_type = ?", req.DeviceID, req.DeviceType).First(&device).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			ErrorsMgr.HandleError(c, fmt.Errorf("设备不存在或密钥错误"))
		} else {
			ErrorsMgr.HandleError(c, fmt.Errorf("数据库错误: %v", err))
		}
		return
	}

	if !device.CheckKey(req.DeviceKey) {
		log.Printf("[DeviceAuth] Invalid key for device: %s", req.DeviceID)
		ErrorsMgr.HandleError(c, fmt.Errorf("设备不存在或密钥错误"))
		return
	}

	device.SetOnline()
	if err := h.Mysql.Save(&device).Error; err != nil {
		ErrorsMgr.HandleError(c, fmt.Errorf("数据库错误: %v", err))
		return
	}

	token := ""
	if globalDeviceJWTManager != nil {
		var err error
		token, err = globalDeviceJWTManager.GenerateToken(device.ID, device.DeviceID, string(device.DeviceType))
		if err != nil {
			ErrorsMgr.HandleError(c, fmt.Errorf("生成Token失败: %v", err))
			return
		}

		if globalDeviceRedisTokenManager != nil {
			err := globalDeviceRedisTokenManager.StoreToken(token, device.ID, device.DeviceID, string(device.DeviceType), time.Now().Add(time.Hour*24))
			if err != nil {
				ErrorsMgr.HandleError(c, fmt.Errorf("保存Token失败: %v", err))
				return
			}
		}
	}

	device.HideKey()

	ErrorsMgr.CreateSuccessResponse(c, gin.H{
		"Device_ID":  device.ID,
		"DeviceID":   device.DeviceID,
		"DeviceName": device.DeviceName,
		"DeviceType": device.DeviceType,
		"Token":      token,
		"ExpireTime": 86400,
	})
}

func (h *DeviceHandler) LogoutDevice(c *gin.Context) {
	deviceID := c.GetHeader("X-Device-ID")
	deviceType := c.GetHeader("X-Device-Type")

	if deviceID == "" || deviceType == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Device ID and Type are required",
		})
		return
	}

	device := &Device.Device{}
	err := h.Mysql.Where("device_id = ? AND device_type = ?", deviceID, deviceType).First(&device).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    1,
				"message": "Device not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "Database error: " + err.Error(),
		})
		return
	}

	device.SetOffline()
	if err := h.Mysql.Save(&device).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "Database error: " + err.Error(),
		})
		return
	}

	token := c.GetHeader("Authorization")
	if token != "" && globalDeviceRedisTokenManager != nil {
		globalDeviceRedisTokenManager.RevokeToken(token)
	}

	ErrorsMgr.CreateSuccessResponse(c, gin.H{
		"message": "Logout successful",
	})
}

var globalDeviceJWTManager *jwtUtils.JWTManager
var globalDeviceRedisTokenManager *Jwt.RedisTokenManager

func SetDeviceJWTManager(jwtMgr *jwtUtils.JWTManager) {
	globalDeviceJWTManager = jwtMgr
}

func SetDeviceRedisTokenManager(redisMgr *Jwt.RedisTokenManager) {
	globalDeviceRedisTokenManager = redisMgr
}
