package terminal

import (
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	Backend "MavlinkProject/Server/Backend"
	User "MavlinkProject/Server/backend/Shared/User"
)

type TerminalManager struct {
	User    *User.User
	Command TerminalCMD
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
	
	default:
		return &TerminalResponse{
			Success: true,
			Message: TCS_map_User,
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

func (tm *TerminalManager) Server() *TerminalResponse {
	server := Backend.GetBackendServer()
	setting := server.SettingManager.GetSetting()
	uptime := time.Since(server.StartTime)

	switch tm.Command.Objects[0] {
	case TCO_config, TCO_empty:
		// server config and Running details implementation
		return &TerminalResponse{
			Success: true,
			Message: map[string]interface{}{
				"command": string(TCS_server),
				"server-status": map[string]interface{}{
					"uptime":         uptime.String(),
					"uptime_seconds": uptime.Seconds(),
					"start_time":     server.StartTime.Format(time.RFC3339),
				},
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
		// default restart time is 5 seconds, i will lock it there
		restartTime := 5
		// getting argument -time
		if tm.Command.Args["t"] != nil {
			if v, ok := tm.Command.Args["t"].(int); ok {
				restartTime = v
			}
		}

		go func() {
			// wait for the server to be ready
			log.Println("[Terminal] Server restart scheduled in", restartTime, "seconds...")
			time.Sleep(time.Duration(restartTime-5) * time.Second)

			itime := 5
			for itime > 0 {
				log.Println("[Terminal] Server restart scheduled in", itime, "seconds...")
				time.Sleep(time.Second)
				itime--
			}

			server := Backend.GetBackendServer()
			server.Restart()
		}()
		// and response, cool
		return &TerminalResponse{
			Success: true,
			Message: map[string]interface{}{
				"command":            string(TCS_server),
				"note":               "server restart has scheduled",
				"restart_in_seconds": restartTime,
			},
		}

	case TCO_shutdown:
		// server shutdown implementation, same like restart()
		// default shutdown time is 5 seconds
		shutdownTime := 5
		// getting argument -time
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
			server := Backend.GetBackendServer()
			server.Shutdown()
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

func (tm *TerminalManager) Backend() *TerminalResponse {
	server := Backend.GetBackendServer()
	setting := server.SettingManager.GetSetting()
	uptime := time.Since(server.StartTime)

	switch tm.Command.Objects[0] {
	case TCO_config, TCO_empty:
		return &TerminalResponse{
			Success: true,
			Message: map[string]interface{}{
				"command": string(TCS_backend),
				"status": map[string]interface{}{
					"uptime":         uptime.String(),
					"uptime_seconds": uptime.Seconds(),
					"start_time":     server.StartTime.Format(time.RFC3339),
				},
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

			server := Backend.GetBackendServer()
			server.Restart()
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

			server := Backend.GetBackendServer()
			server.Shutdown()
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
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_frontend),
			"note":    "frontend details",
			"args":    tm.Command.Args,
		},
	}
}

func (tm *TerminalManager) Database() *TerminalResponse {
	server := Backend.GetBackendServer()
	setting := server.SettingManager.GetSetting()

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
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_log),
			"note":    "backend logs",
			"args":    tm.Command.Args,
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
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_cache),
			"note":    "cache details",
			"args":    tm.Command.Args,
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
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command":  string(TCS_adduser),
			"username": tm.Command.Objects[0],
			"email":    tm.Command.Objects[1],
			"args":     tm.Command.Args,
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
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command":  string(TCS_deluser),
			"username": tm.Command.Objects[0],
			"args":     tm.Command.Args,
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
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_reboot),
			"note":    "restart server",
			"args":    tm.Command.Args,
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
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_shutdown),
			"note":    "shutdown server",
			"args":    tm.Command.Args,
		},
	}
}
