package MiddleWare

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	FirstBanDuration       = 2 * time.Minute
	MaxBanDuration         = 10 * 365 * 24 * time.Hour
	BanishedFolder         = "./Banished"
	BanishedUpdateInterval = 15 * time.Minute
	HighFreqThreshold      = 250
	HighFreqTimeWindow     = 1 * time.Second
)

type BanishRecord struct {
	IP         string    `json:"ip"`
	BanEndTime time.Time `json:"ban_end_time"`
	BanCount   int       `json:"ban_count"`
	Reason     string    `json:"reason"`
}

var (
	banishMap       map[string]*BanishRecord
	banishMutex     sync.RWMutex
	highFreqTracker map[string][]time.Time
	highFreqMutex   sync.RWMutex
)

func init() {
	banishMap = make(map[string]*BanishRecord)
	highFreqTracker = make(map[string][]time.Time)

	if err := os.MkdirAll(BanishedFolder, 0755); err != nil {
		log.Printf("创建Banished文件夹失败: %v", err)
	}

	go cleanupExpiredBans()
	go periodicBanishedUpdate()
}

func calculateBanDuration(banCount int) time.Duration {
	duration := FirstBanDuration
	for i := 1; i < banCount; i++ {
		duration *= 2
	}
	if duration > MaxBanDuration {
		duration = MaxBanDuration
	}
	return duration
}

func formatBanDuration(d time.Duration) string {
	hours := int(d.Hours())
	minutes := int(d.Minutes()) % 60
	seconds := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
}

func IsIPBanned(ip string) bool {
	banishMutex.RLock()
	defer banishMutex.RUnlock()

	record, exists := banishMap[ip]
	if !exists {
		return false
	}

	if time.Now().After(record.BanEndTime) {
		return false
	}
	return true
}

func GetBanInfo(ip string) (banned bool, remainingTime time.Duration) {
	banishMutex.RLock()
	defer banishMutex.RUnlock()

	record, exists := banishMap[ip]
	if !exists || time.Now().After(record.BanEndTime) {
		return false, 0
	}

	remainingTime = time.Until(record.BanEndTime)
	return true, remainingTime
}

func BanIP(ip string, reason string) {
	banishMutex.Lock()
	defer banishMutex.Unlock()

	record, exists := banishMap[ip]
	if !exists {
		record = &BanishRecord{
			IP:       ip,
			BanCount: 1,
			Reason:   reason,
		}
	} else {
		record.BanCount++
		record.Reason = reason
	}

	duration := calculateBanDuration(record.BanCount)
	record.BanEndTime = time.Now().Add(duration)
	banishMap[ip] = record

	log.Printf("🛑 IP[%s] 已被封禁，时长: %s (第%d次违规)", ip, formatBanDuration(duration), record.BanCount)
	saveBanishedToFile()
}

func RecordHighFreqAccess(ip string) bool {
	highFreqMutex.Lock()
	defer highFreqMutex.Unlock()

	now := time.Now()
	accessTimes := highFreqTracker[ip]

	var validAccesses []time.Time
	for _, t := range accessTimes {
		if now.Sub(t) < HighFreqTimeWindow {
			validAccesses = append(validAccesses, t)
		}
	}

	validAccesses = append(validAccesses, now)
	highFreqTracker[ip] = validAccesses

	if len(validAccesses) >= HighFreqThreshold {
		BanIP(ip, "高频访问")
		delete(highFreqTracker, ip)
		return true
	}

	return false
}

func cleanupExpiredBans() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		banishMutex.Lock()
		changed := false
		for ip, record := range banishMap {
			if time.Now().After(record.BanEndTime) {
				delete(banishMap, ip)
				changed = true
			}
		}
		banishMutex.Unlock()

		if changed {
			saveBanishedToFile()
		}
	}
}

func periodicBanishedUpdate() {
	ticker := time.NewTicker(BanishedUpdateInterval)
	defer ticker.Stop()

	for range ticker.C {
		banishMutex.RLock()
		activeBans := make(map[string]*BanishRecord)
		for ip, record := range banishMap {
			if time.Now().Before(record.BanEndTime) {
				activeBans[ip] = record
			}
		}
		banishMutex.RUnlock()

		if len(activeBans) > 0 {
			saveBanishedToFile()
		} else {
			clearBanishedFile()
		}
	}
}

func saveBanishedToFile() {
	banishMutex.RLock()
	defer banishMutex.RUnlock()

	activeBans := make(map[string]*BanishRecord)
	for ip, record := range banishMap {
		if time.Now().Before(record.BanEndTime) {
			activeBans[ip] = record
		}
	}

	if len(activeBans) == 0 {
		clearBanishedFile()
		return
	}

	filePath := filepath.Join(BanishedFolder, "BanishedMap.json")
	data, err := json.MarshalIndent(activeBans, "", "  ")
	if err != nil {
		log.Printf("序列化BanishedMap失败: %v", err)
		return
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		log.Printf("保存BanishedMap失败: %v", err)
	}
}

func clearBanishedFile() {
	filePath := filepath.Join(BanishedFolder, "BanishedMap.json")
	if _, err := os.Stat(filePath); err == nil {
		if err := os.Remove(filePath); err != nil {
			log.Printf("删除BanishedMap文件失败: %v", err)
		}
	}
}

func LoadBanishedFromFile() {
	filePath := filepath.Join(BanishedFolder, "BanishedMap.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return
	}

	var records map[string]*BanishRecord
	if err := json.Unmarshal(data, &records); err != nil {
		log.Printf("加载BanishedMap失败: %v", err)
		return
	}

	banishMutex.Lock()
	defer banishMutex.Unlock()

	for ip, record := range records {
		if time.Now().Before(record.BanEndTime) {
			banishMap[ip] = record
		}
	}
	log.Printf("已加载 %d 条封禁记录", len(banishMap))
}

func GetBanishedContext() string {
	banishMutex.RLock()
	defer banishMutex.RUnlock()

	activeBans := make(map[string]interface{})
	for ip, record := range banishMap {
		if time.Now().Before(record.BanEndTime) {
			activeBans[ip] = map[string]interface{}{
				"ban_end_time": record.BanEndTime.Format(time.RFC3339),
				"ban_count":    record.BanCount,
				"reason":       record.Reason,
				"banned_time":  formatBanDuration(time.Until(record.BanEndTime)),
			}
		}
	}

	data, _ := json.Marshal(activeBans)
	return string(data)
}

func BanishCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := getRealClientIP(c)

		if IsIPBanned(ip) {
			_, remaining := GetBanInfo(ip)
			c.JSON(403, gin.H{
				"message": fmt.Sprintf("🛑 You have banned by Backend Website for %s 🈲", formatBanDuration(remaining)),
				"banned":  true,
			})
			c.Abort()
			return
		}

		banishedData := GetBanishedContext()
		c.Set("Banished", banishedData)

		c.Next()
	}
}

func RecordSuspiciousAccess(ip string) {
	BanIP(ip, "可疑访问")
}
