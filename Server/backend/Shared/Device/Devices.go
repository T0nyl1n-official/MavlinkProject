package Shared

import (
	"crypto/md5"
	"fmt"
	"time"

	gorm "gorm.io/gorm"
)

type DeviceType string

const (
	DeviceTypeCentral  DeviceType = "central"
	DeviceTypeLandNode DeviceType = "landnode"
	DeviceTypeSensor   DeviceType = "sensor"
	DeviceTypeDrone   DeviceType = "drone"
)

type Device struct {
	gorm.Model
	DeviceID   string     `gorm:"unique, not null"`
	DeviceName string     `gorm:"not null"`
	DeviceType DeviceType `gorm:"not null"`
	DeviceKey  string     `gorm:"not null"`
	IsOnline   bool       `gorm:"default:false"`
	LastSeen   time.Time  `gorm:"default:null"`
}

func (d *Device) CheckKey(key string) bool {
	return d.DeviceKey == fmt.Sprintf("%x", md5.Sum([]byte(key)))
}

func (d *Device) SetOnline() {
	d.IsOnline = true
	d.LastSeen = time.Now()
}

func (d *Device) SetOffline() {
	d.IsOnline = false
}

func (d *Device) HideKey() {
	d.DeviceKey = ""
}

func (d *Device) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"ID":         d.ID,
		"DeviceID":   d.DeviceID,
		"DeviceName": d.DeviceName,
		"DeviceType": d.DeviceType,
		"IsOnline":   d.IsOnline,
		"LastSeen":   d.LastSeen,
	}
}
