package Encryption

import (
	"encoding/base64"
	"errors"
	"os"
)

// KeyManager 密钥管理器
type KeyManager struct {
	encryptionKey []byte
}

// NewKeyManager 创建密钥管理器
func NewKeyManager() (*KeyManager, error) {
	// 从环境变量获取密钥
	key := os.Getenv("GUIYU_CHAIN_ENCRYPTION_KEY")
	// if key == "" {
	// 	// 生成一个默认密钥用于开发环境
	// 	return &KeyManager{encryptionKey: []byte("00000000000000000000000000000000")}, nil
	// }
	// 验证密钥长度和解码
	decodedKey, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return nil, errors.New("encryption - KeyManager: 密钥不是有效的Base64编码")
	}
	if len(decodedKey) != 32 {
		return nil, errors.New("encryption - KeyManager: 解码后密钥长度必须为32字节")
	}
	return &KeyManager{encryptionKey: decodedKey}, nil
}

// GetEncryptionKey 获取加密密钥
func (km *KeyManager) GetEncryptionKey() []byte {
	return km.encryptionKey
}

// RotateKey 密钥轮换（需要重新加密所有数据）
func (km *KeyManager) RotateKey(newKey string) error {
	if len(newKey) != 44 { // base64编码的32字节密钥长度
		return errors.New("新密钥格式不正确")
	}
	// 对新密钥进行Base64解码
	decodedKey, err := base64.StdEncoding.DecodeString(newKey)
	if err != nil {
		return errors.New("新密钥不是有效的Base64编码")
	}
	km.encryptionKey = decodedKey
	return nil
}
