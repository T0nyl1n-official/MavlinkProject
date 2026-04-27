package terminal

import (
	"bufio"
	"crypto/md5"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"

	Conf "MavlinkProject/Server/backend/Config"
	MiddleWare "MavlinkProject/Server/backend/Middles"
	User "MavlinkProject/Server/backend/Shared/User"
)

type TerminalManager struct {
	User           *User.User
	Command        TerminalCMD
	DB             *gorm.DB
	SettingManager *Conf.SettingManager
}

func (tm *TerminalManager) Handle() *TerminalResponse {
	switch tm.Command.Command {
	case TCS_help:
		return tm.Help()
	case TCS_whoami:
		return tm.Whoami()
	case TCS_ls:
		return tm.Ls()
	case TCS_cd:
		return tm.Cd()
	case TCS_mod:
		return tm.Mod()
	case TCS_show:
		return tm.Show()
	case TCS_server:
		return tm.Server()
	case TCS_backend:
		return tm.Backend()
	case TCS_frontend:
		return tm.Frontend()
	case TCS_database:
		return tm.Database()
	case TCS_mavlink:
		return tm.Mavlink()
	case TCS_log:
		return tm.Log()
	case TCS_board:
		return tm.Board()
	case TCS_drone:
		return tm.Drone()
	case TCS_sensor:
		return tm.Sensor()
	case TCS_cache:
		return tm.Cache()
	case TCS_adduser:
		return tm.Adduser()
	case TCS_deluser:
		return tm.Deluser()
	case TCS_auto:
		return tm.Auto()
	case TCS_reboot:
		return tm.Reboot()
	case TCS_shutdown:
		return tm.Shutdown()
	}
	return &TerminalResponse{
		Success: false,
		Message: map[string]interface{}{
			"error":   "Unknown command",
			"message": "Use command \"help\" for more information",
		},
	}
}

func (tm *TerminalManager) Help() *TerminalResponse {
	switch tm.Command.Objects[0] {
	case TCO_empty:
		if tm.User.IsAdmin {
			return &TerminalResponse{
				Success: true,
				Message: TCS_map_Admin,
			}
		} else {
			return &TerminalResponse{
				Success: true,
				Message: TCS_map_User,
			}
		}
	case "help":
		return &TerminalResponse{Success: true, Message: THD_help}
	case "mod":
		return &TerminalResponse{Success: true, Message: THD_mod}
	case "show":
		return &TerminalResponse{Success: true, Message: THD_show}
	case "server":
		return &TerminalResponse{Success: true, Message: THD_server}
	case "backend":
		return &TerminalResponse{Success: true, Message: THD_backend}
	case "database":
		return &TerminalResponse{Success: true, Message: THD_database}
	case "mavlink":
		return &TerminalResponse{Success: true, Message: THD_mavlink}
	case "board":
		return &TerminalResponse{Success: true, Message: THD_board}
	case "log":
		return &TerminalResponse{Success: true, Message: THD_log}
	case "drone":
		return &TerminalResponse{Success: true, Message: THD_drone}
	case "sensor":
		return &TerminalResponse{Success: true, Message: THD_sensor}
	case "cache":
		return &TerminalResponse{Success: true, Message: THD_cache}
	case "adduser":
		return &TerminalResponse{Success: true, Message: THD_adduser}
	case "deluser":
		return &TerminalResponse{Success: true, Message: THD_deluser}
	case "auto":
		return &TerminalResponse{Success: true, Message: THD_auto}
	case "reboot":
		return &TerminalResponse{Success: true, Message: THD_reboot}
	case "shutdown":
		return &TerminalResponse{Success: true, Message: THD_shutdown}
	case "frontend":
		return &TerminalResponse{Success: true, Message: THD_frontend}
	default:
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error":   "Unknown command",
				"message": "Use command \"help\" for more information",
			},
		}
	}

}

func (tm *TerminalManager) Whoami() *TerminalResponse {
	permission := "Unknown_ERROR"
	if tm.User.IsAdmin {
		permission = "Admin"
	} else {
		permission = "User"
	}
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"username":   tm.User.Username,
			"permission": permission,
		},
	}
}

func (tm *TerminalManager) Ls() *TerminalResponse {
	targetPath := tm.Command.Path
	if targetPath == "" {
		targetPath = "."
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"command": "/" + string(TCS_ls) + " :",
				"error":   "invalid path",
			},
		}
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"command": "/" + string(TCS_ls) + " :",
				"error":   "cannot read directory: " + err.Error(),
			},
		}
	}

	var objects string
	if len(entries) == 0 {
		objects = "(none)"
	} else {
		var names []string
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		objects = strings.Join(names, "\t")
	}

	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": "/" + string(TCS_ls) + " :",
			"path":    absPath,
			"objects": objects,
		},
	}
}

func (tm *TerminalManager) Cd() *TerminalResponse {
	if _, err := os.Stat(tm.Command.Path); err != nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"command": string(TCS_cd),
				"error":   "invalid path",
				"message": "Please check the path exists.\nUse command \"help ls\" for more information",
			},
		}
	}
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_cd),
			"note":    "cd implementation",
			"path":    tm.Command.Path,
		},
	}
}

func (tm *TerminalManager) Mod() *TerminalResponse {
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_mod),
			"note":    "mod directory management",
			"objects": tm.Command.Objects,
			"args":    tm.Command.Args,
		},
	}
}

func (tm *TerminalManager) Show() *TerminalResponse {
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_show),
			"note":    "show object details",
			"objects": tm.Command.Objects,
			"args":    tm.Command.Args,
		},
	}
}

// Server: 服务器配置
func (tm *TerminalManager) Server() *TerminalResponse {
	if tm.SettingManager == nil {
		return &TerminalResponse{Success: false, Message: "SettingManager not initialized"}
	}
	setting := tm.SettingManager.GetSetting()

	switch tm.Command.Objects[0] {
	case TCO_config, TCO_empty:
		return &TerminalResponse{
			Success: true,
			Message: map[string]interface{}{
				"command": string(TCS_server),
				"server-config": map[string]interface{}{
					"database": setting.Database,
					"redis":    setting.Redis,
					"jwt":      setting.JWT,
					"cors":     setting.CORS,
					"rate_lim": setting.RateLimit,
					"logger":   setting.Logger,
					"board":    setting.Board,
				},
			},
		}
	case TCO_restart:
		restartTime := 5
		if tm.Command.Args["t"] != nil {
			if v, ok := tm.Command.Args["t"].(int); ok {
				restartTime = v
			}
		}

		// 重启服务器 - 新线程
		go func() {
			log.Println("[Terminal] Server restart scheduled in", restartTime, "seconds...")
			time.Sleep(time.Duration(restartTime-5) * time.Second)

			itime := 5
			for itime > 0 {
				log.Println("[Terminal] Server restart scheduled in", itime, "seconds...")
				time.Sleep(time.Second)
				itime--
			}
		}()

		return &TerminalResponse{
			Success: true,
			Message: map[string]interface{}{
				"command":            string(TCS_server),
				"note":               "server restart has scheduled",
				"restart_in_seconds": restartTime,
			},
		}

	case TCO_shutdown:
		shutdownTime := 5
		if tm.Command.Args["t"] != nil {
			if v, ok := tm.Command.Args["t"].(int); ok {
				shutdownTime = v
			}
		}
		go func() {
			log.Println("[Terminal] Server shutdown scheduled in", shutdownTime, "seconds...")
			time.Sleep(time.Duration(shutdownTime-5) * time.Second)

			itime := 5
			for itime > 0 {
				log.Println("[Terminal] Server shutdown scheduled in", itime, "seconds...")
				time.Sleep(time.Second)
				itime--
			}
		}()

		return &TerminalResponse{
			Success: true,
			Message: map[string]interface{}{
				"command":             string(TCS_server),
				"note":                "server shutdown has scheduled",
				"shutdown_in_seconds": shutdownTime,
			},
		}

	default:
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error":   "Unknown command syntax",
				"message": "Use command \"help server\" for more information",
			},
		}
	}
}

// Backend: 后端配置
func (tm *TerminalManager) Backend() *TerminalResponse {
	if tm.SettingManager == nil {
		return &TerminalResponse{Success: false, Message: "SettingManager not initialized"}
	}
	setting := tm.SettingManager.GetSetting()

	switch tm.Command.Objects[0] {
	case TCO_config, TCO_empty:
		return &TerminalResponse{
			Success: true,
			Message: map[string]interface{}{
				"command": string(TCS_backend),
				"config": map[string]interface{}{
					"database": setting.Database,
					"redis":    setting.Redis,
					"jwt":      setting.JWT,
					"cors":     setting.CORS,
					"rate_lim": setting.RateLimit,
					"logger":   setting.Logger,
					"board":    setting.Board,
				},
			},
		}
	case TCO_restart:
		restartTime := 5
		if tm.Command.Args["t"] != nil {
			if v, ok := tm.Command.Args["t"].(int); ok {
				restartTime = v
			}
		}

		go func() {
			log.Println("[Terminal] Backend restart scheduled in", restartTime, "seconds...")
			time.Sleep(time.Duration(restartTime-5) * time.Second)

			itime := 5
			for itime > 0 {
				log.Println("[Terminal] Backend restart in", itime, "seconds...")
				time.Sleep(time.Second)
				itime--
			}
		}()

		return &TerminalResponse{
			Success: true,
			Message: map[string]interface{}{
				"command":            string(TCS_backend),
				"note":               "backend restart scheduled",
				"restart_in_seconds": restartTime,
			},
		}
	case TCO_shutdown:
		shutdownTime := 5
		if tm.Command.Args["t"] != nil {
			if v, ok := tm.Command.Args["t"].(int); ok {
				shutdownTime = v
			}
		}

		go func() {
			log.Println("[Terminal] Backend shutdown scheduled in", shutdownTime, "seconds...")
			time.Sleep(time.Duration(shutdownTime-5) * time.Second)

			itime := 5
			for itime > 0 {
				log.Println("[Terminal] Backend shutdown in", itime, "seconds...")
				time.Sleep(time.Second)
				itime--
			}
		}()

		return &TerminalResponse{
			Success: true,
			Message: map[string]interface{}{
				"command":             string(TCS_backend),
				"note":                "backend shutdown scheduled",
				"shutdown_in_seconds": shutdownTime,
			},
		}

	default:
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error":   "Unknown command syntax",
				"message": "Use command \"help backend\" for more information",
			},
		}
	}
}

func (tm *TerminalManager) Frontend() *TerminalResponse {
	return &TerminalResponse{
		Success: false,
		Message: map[string]interface{}{
			"error":   "frontend not implemented",
			"message": "Fatal: frontend config request do not transport to Backend",
		},
	}
}

// Database: 数据库配置
func (tm *TerminalManager) Database() *TerminalResponse {
	if tm.SettingManager == nil {
		return &TerminalResponse{Success: false, Message: "SettingManager not initialized"}
	}
	setting := tm.SettingManager.GetSetting()

	objectsLen := len(tm.Command.Objects)

	switch {
	case objectsLen == 0 || (objectsLen == 1 && (tm.Command.Objects[0] == TCO_empty || tm.Command.Objects[0] == TCO_config)):
		return &TerminalResponse{
			Success: true,
			Message: map[string]interface{}{
				"command": string(TCS_database),
				"mysql": map[string]interface{}{
					"host":     setting.Database.MySQL.Host,
					"port":     setting.Database.MySQL.Port,
					"user":     setting.Database.MySQL.User,
					"database": setting.Database.MySQL.Database,
					"charset":  setting.Database.MySQL.Charset,
				},
				"redis": map[string]interface{}{
					"host": setting.Redis.Host,
					"port": setting.Redis.Port,
				},
			},
		}
	case objectsLen >= 1 && tm.Command.Objects[0] == "mysql":
		if objectsLen == 1 || (objectsLen == 2 && (tm.Command.Objects[1] == TCO_empty || tm.Command.Objects[1] == TCO_config)) {
			return &TerminalResponse{
				Success: true,
				Message: map[string]interface{}{
					"command": string(TCS_database),
					"mysql": map[string]interface{}{
						"host":     setting.Database.MySQL.Host,
						"port":     setting.Database.MySQL.Port,
						"user":     setting.Database.MySQL.User,
						"database": setting.Database.MySQL.Database,
						"charset":  setting.Database.MySQL.Charset,
					},
				},
			}
		}
	case objectsLen >= 1 && tm.Command.Objects[0] == "redis":
		if objectsLen == 1 || (objectsLen == 2 && (tm.Command.Objects[1] == TCO_empty || tm.Command.Objects[1] == TCO_config)) {
			return &TerminalResponse{
				Success: true,
				Message: map[string]interface{}{
					"command": string(TCS_database),
					"redis": map[string]interface{}{
						"host": setting.Redis.Host,
						"port": setting.Redis.Port,
					},
				},
			}
		}
	}

	return &TerminalResponse{
		Success: false,
		Message: map[string]interface{}{
			"command": string(TCS_database),
			"error":   "Unknown syntax",
			"message": "Use \"database mysql \" as example or use \"help database\" to get more information",
		},
	}
}


func (tm *TerminalManager) Mavlink() *TerminalResponse {
	// unfinished
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_mavlink),
			"note":    "mavlink details",
			"args":    tm.Command.Args,
		},
	}
}

func (tm *TerminalManager) Log() *TerminalResponse {
	object := "server"
	logType := "any"
	maxLines := 10

	if len(tm.Command.Objects) > 0 && tm.Command.Objects[0] != TCO_empty {
		object = string(tm.Command.Objects[0])
	}

	if len(tm.Command.Objects) > 1 && tm.Command.Objects[1] != TCO_empty {
		logType = string(tm.Command.Objects[1])
	}

	if len(tm.Command.Objects) > 2 {
		if level, err := strconv.Atoi(string(tm.Command.Objects[2])); err == nil {
			maxLines = level
		}
	}

	logDir := MiddleWare.LogDirFunc()

	var logFiles []string
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error":   "Cannot read log directory",
				"details": err.Error(),
			},
		}
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "log_") && strings.HasSuffix(entry.Name(), ".log") {
			logFiles = append(logFiles, entry.Name())
		}
	}

	if len(logFiles) == 0 {
		return &TerminalResponse{
			Success: true,
			Message: map[string]interface{}{
				"object":  object,
				"type":    logType,
				"count":   0,
				"logs":    []string{},
				"message": "No log files found",
			},
		}
	}

	sortFiles := false
	for _, f := range logFiles {
		if strings.Contains(f, object) || object == "server" {
			sortFiles = true
			break
		}
	}

	var sortedFiles []string
	if sortFiles {
		for i := len(logFiles) - 1; i >= 0; i-- {
			sortedFiles = append(sortedFiles, logFiles[i])
		}
	} else {
		sortedFiles = logFiles
	}

	var results []string
	logTypeLower := strings.ToLower(logType)

	for _, fileName := range sortedFiles {
		if len(results) >= maxLines {
			break
		}

		logPath := filepath.Join(logDir, fileName)
		file, err := os.Open(logPath)
		if err != nil {
			continue
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()

			if logTypeLower != "any" && !strings.Contains(strings.ToLower(line), logTypeLower) {
				continue
			}

			results = append(results, line)

			if len(results) >= maxLines {
				break
			}
		}
		file.Close()
	}

	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"object": object,
			"type":   logType,
			"count":  len(results),
			"logs":   results,
		},
	}
}

func (tm *TerminalManager) Board() *TerminalResponse {
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_board),
			"note":    "board details",
			"objects": tm.Command.Objects,
			"args":    tm.Command.Args,
		},
	}
}

func (tm *TerminalManager) Drone() *TerminalResponse {
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_drone),
			"note":    "drone details",
			"objects": tm.Command.Objects,
			"args":    tm.Command.Args,
		},
	}
}

func (tm *TerminalManager) Sensor() *TerminalResponse {
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_sensor),
			"note":    "sensor details",
			"objects": tm.Command.Objects,
			"args":    tm.Command.Args,
		},
	}
}

func (tm *TerminalManager) Cache() *TerminalResponse {
	if tm.SettingManager == nil {
		return &TerminalResponse{Success: false, Message: "SettingManager not initialized"}
	}
	setting := tm.SettingManager.GetSetting()

	object := ""
	if len(tm.Command.Objects) > 0 && tm.Command.Objects[0] != TCO_empty {
		object = string(tm.Command.Objects[0])
	}

	if object == "" || object == "config" {
		return &TerminalResponse{
			Success: true,
			Message: map[string]interface{}{
				"command": string(TCS_cache),
				"object":  "config",
				"redis": map[string]interface{}{
					"host": setting.Redis.Host,
					"port": setting.Redis.Port,
				},
			},
		}
	}

	return &TerminalResponse{
		Success: false,
		Message: map[string]interface{}{
			"command": string(TCS_cache),
			"object":  object,
			"message": "Use 'cache config' to view cache configuration",
		},
	}
}

func (tm *TerminalManager) Adduser() *TerminalResponse {
	if !tm.User.IsAdmin {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "permission denied: admin required",
			},
		}
	}

	if len(tm.Command.Objects) < 2 {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error":   "invalid arguments",
				"usage":   "adduser [username] [email] [password]",
				"example": "adduser admin admin@example.com password123",
			},
		}
	}

	username := string(tm.Command.Objects[0])
	email := string(tm.Command.Objects[1])
	password := "default_password"
	if len(tm.Command.Objects) >= 3 {
		password = string(tm.Command.Objects[2])
	}

	if tm.DB == nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "database not available",
			},
		}
	}

	var existingUser User.User
	err := tm.DB.Where("username = ? OR email = ?", username, email).First(&existingUser).Error
	if err == nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "user already exists",
			},
		}
	}

	newUser := User.User{
		Username: username,
		Email:    email,
		Password: fmt.Sprintf("%x", md5.Sum([]byte(password))),
		IsAdmin:  false,
		IsOnline: false,
	}

	err = tm.DB.Create(&newUser).Error
	if err != nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error":   "failed to create user",
				"details": err.Error(),
			},
		}
	}

	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command":  string(TCS_adduser),
			"username": username,
			"email":    email,
			"user_id":  newUser.ID,
			"message":  "user created successfully",
		},
	}
}

func (tm *TerminalManager) Deluser() *TerminalResponse {
	if !tm.User.IsAdmin {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "permission denied: admin required",
			},
		}
	}

	if len(tm.Command.Objects) < 1 {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error":   "invalid arguments",
				"usage":   "deluser [username]",
				"example": "deluser testuser",
			},
		}
	}

	username := string(tm.Command.Objects[0])

	if tm.DB == nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "database not available",
			},
		}
	}

	var user User.User
	err := tm.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "user not found",
			},
		}
	}

	if user.ID == tm.User.ID {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "cannot delete yourself",
			},
		}
	}

	err = tm.DB.Delete(&user).Error
	if err != nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error":   "failed to delete user",
				"details": err.Error(),
			},
		}
	}

	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command":  string(TCS_deluser),
			"username": username,
			"user_id":  user.ID,
			"message":  "user deleted successfully",
		},
	}
}

func (tm *TerminalManager) Chmod() *TerminalResponse {
	if !tm.User.IsAdmin {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "permission denied: admin required",
			},
		}
	}

	if len(tm.Command.Objects) < 2 {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error":    "invalid arguments",
				"usage":    "chmod [username] [property] [value]",
				"example":  "chmod testuser admin true",
				"example2": "chmod testuser password newpass123",
			},
		}
	}

	username := string(tm.Command.Objects[0])
	property := string(tm.Command.Objects[1])
	value := ""
	if len(tm.Command.Objects) >= 3 {
		value = string(tm.Command.Objects[2])
	}

	if tm.DB == nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "database not available",
			},
		}
	}

	var user User.User
	err := tm.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "user not found",
			},
		}
	}

	switch property {
	case "admin":
		if value == "true" {
			user.IsAdmin = true
		} else if value == "false" {
			user.IsAdmin = false
		} else {
			return &TerminalResponse{
				Success: false,
				Message: map[string]interface{}{
					"error":   "invalid value for admin, use true or false",
					"example": "chmod username admin true",
				},
			}
		}
	case "password":
		if value == "" {
			return &TerminalResponse{
				Success: false,
				Message: map[string]interface{}{
					"error":   "password cannot be empty",
					"example": "chmod username password newpassword",
				},
			}
		}
		user.Password = fmt.Sprintf("%x", md5.Sum([]byte(value)))
	case "email":
		if value == "" {
			return &TerminalResponse{
				Success: false,
				Message: map[string]interface{}{
					"error":   "email cannot be empty",
					"example": "chmod username email newemail@example.com",
				},
			}
		}
		user.Email = value
	default:
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error":    "unknown property",
				"property": "available properties: admin, password, email",
			},
		}
	}

	err = tm.DB.Save(&user).Error
	if err != nil {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error":   "failed to update user",
				"details": err.Error(),
			},
		}
	}

	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command":  string(TCS_chmod),
			"username": username,
			"property": property,
			"value":    value,
			"message":  "user updated successfully",
		},
	}
}

func (tm *TerminalManager) Auto() *TerminalResponse {
	if !tm.User.IsAdmin {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "permission denied: admin required",
			},
		}
	}
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_auto),
			"note":    "AI agent auto management",
			"args":    tm.Command.Args,
		},
	}
}

func (tm *TerminalManager) Reboot() *TerminalResponse {
	if !tm.User.IsAdmin {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "permission denied: admin required",
			},
		}
	}

	restartTime := 5
	if tm.Command.Args["t"] != nil {
		if v, ok := tm.Command.Args["t"].(int); ok {
			restartTime = v
		}
	}

	go func() {
		log.Println("[Terminal] System reboot scheduled in", restartTime, "seconds...")
		time.Sleep(time.Duration(restartTime-5) * time.Second)

		itime := 5
		for itime > 0 {
			log.Println("[Terminal] System reboot scheduled in", itime, "seconds...")
			time.Sleep(time.Second)
			itime--
		}
	}()

	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command":           string(TCS_reboot),
			"note":              "system reboot has scheduled",
			"reboot_in_seconds": restartTime,
		},
	}
}

func (tm *TerminalManager) Shutdown() *TerminalResponse {
	if !tm.User.IsAdmin {
		return &TerminalResponse{
			Success: false,
			Message: map[string]interface{}{
				"error": "permission denied: admin required",
			},
		}
	}

	shutdownTime := 5
	if tm.Command.Args["t"] != nil {
		if v, ok := tm.Command.Args["t"].(int); ok {
			shutdownTime = v
		}
	}

	go func() {
		log.Println("[Terminal] System shutdown scheduled in", shutdownTime, "seconds...")
		time.Sleep(time.Duration(shutdownTime-5) * time.Second)

		itime := 5
		for itime > 0 {
			log.Println("[Terminal] System shutdown scheduled in", itime, "seconds...")
			time.Sleep(time.Second)
			itime--
		}
	}()

	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command":             string(TCS_shutdown),
			"note":                "system shutdown has scheduled",
			"shutdown_in_seconds": shutdownTime,
		},
	}
}
