package Shared

import (
	"crypto/md5"
	"fmt"

	gorm "gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"unique, not null"`
	Email    string `gorm:"unique, not null"`
	Password string `gorm:"not null"`

	IsAdmin  bool `gorm:"default:false"`
	IsOnline bool `gorm:"default:false"`
}

func (u *User) SetAdmin() {
	u.IsAdmin = true
}

func (u *User) DecAdmin() {
	u.IsAdmin = false
}

func (u *User) SetOnline() {
	u.IsOnline = true
}

func (u *User) SetOffline() {
	u.IsOnline = false
}

func (u *User) HidePassword() {
	u.Password = ""
}

func (u *User) CheckPassword(password string) bool {
	return u.Password == fmt.Sprintf("%x", md5.Sum([]byte(password)))
}

func (u *User) GetIsAdmin() bool {
	return u.IsAdmin
}

func (u *User) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"ID":       u.ID,
		"Username": u.Username,
		"Email":    u.Email,
		"IsAdmin":  u.IsAdmin,
		"IsOnline": u.IsOnline,
	}
}
