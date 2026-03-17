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

	isAdmin  bool `gorm:"default:false"`
	isOnline bool `gorm:"default:false"`
}

func (u *User) SetAdmin(isAdmin bool) {
	u.isAdmin = true
}

func (u *User) DecAdmin(username string) {
	u.isAdmin = false
}

func (u *User) IsOnline() bool {
	return u.isOnline
}

func (u *User) SetOnline() {
	u.isOnline = true
}

func (u *User) SetOffline() {
	u.isOnline = false
}

func (u *User) HidePassword() {
	u.Password = ""
}

func (u *User) CheckPassword(password string) bool {
	return u.Password == fmt.Sprintf("%x", md5.Sum([]byte(password)))
}

func (u *User) ToJSON() map[string]interface{} {
	return map[string]interface{}{
		"ID":       u.ID,
		"Username": u.Username,
		"Email":    u.Email,
		"IsAdmin":  u.isAdmin,
		"IsOnline": u.isOnline,
	}
}
