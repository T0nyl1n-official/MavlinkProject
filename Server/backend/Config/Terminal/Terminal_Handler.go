package terminal

import (
	"os"
	"path/filepath"
	"strings"

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
	if tm.User.IsAdmin {
		return &TerminalResponse{
			Success: true,
			Message: TCS_map_Admin,
		}
	}
	return &TerminalResponse{
		Success: true,
		Message: TCS_map_User,
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
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_server),
			"note":    "server details",
			"args":    tm.Command.Args,
		},
	}
}

func (tm *TerminalManager) Backend() *TerminalResponse {
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_backend),
			"note":    "backend details",
			"args":    tm.Command.Args,
		},
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
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			"command": string(TCS_database),
			"note":    "database details",
			"objects": tm.Command.Objects,
			"args":    tm.Command.Args,
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
