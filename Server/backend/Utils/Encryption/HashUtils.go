package Encryption

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

var ErrDataTampered = errors.New("数据已被篡改")

// CalculateHash 计算数据的SHA-256哈希值
func CalculateHash(data string) string {
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

// VerifyHash 验证数据哈希是否匹配
func VerifyHash(data string, expectedHash string) bool {
	actualHash := CalculateHash(data)
	return actualHash == expectedHash
}

// CalculateNodeHash 计算节点数据的完整哈希（包含所有敏感字段）
func CalculateNodeHash(addressFrom, addressTo, data string) string {
	combinedData := fmt.Sprintf("%s|%s|%s", addressFrom, addressTo, data)
	return CalculateHash(combinedData)
}

// VerifyNodeHash 验证节点数据完整性
func VerifyNodeHash(addressFrom, addressTo, data, expectedHash string) error {
	actualHash := CalculateNodeHash(addressFrom, addressTo, data)
	if actualHash != expectedHash {
		return ErrDataTampered
	}
	return nil
}