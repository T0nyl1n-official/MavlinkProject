package terminal

import (
	User "MavlinkProject/Server/backend/Shared/User"
)

type TerminalManager struct {
	User *User.User
	Command TerminalCMD
}

func (tm *TerminalManager) Handle() *TerminalResponse {
	switch tm.Command.Command {
	case TCS_help:
		return tm.Help()
	case TCS_whoami:
		return tm.Whoami()
	}
	return &TerminalResponse{
		Success: false,
		Message: map[string]interface{}{
			"error": "Unknown command",
		},
	}
}

func (tm *TerminalManager) Help() *TerminalResponse {
	return &TerminalResponse{
		Success: true,
		Message: map[string]interface{}{
			
		},
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
			"username": tm.User.Username,
			"Permission": permission,
		},
	}
}
