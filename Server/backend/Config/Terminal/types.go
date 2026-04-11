package terminal

type (
	Terminal_Command_String string
	Terminal_Command_Args   string
	Terminal_Command_Object string
)

// Command String Standard
const (
	TCS_help   Terminal_Command_String = "help"
	TCS_whoami Terminal_Command_String = "whoami"
	TCS_ls     Terminal_Command_String = "ls"
	TCS_cd     Terminal_Command_String = "cd"
	TCS_mod    Terminal_Command_String = "mod"

	TCS_adduser Terminal_Command_String = "adduser"
	TCS_deluser Terminal_Command_String = "deluser"
	TCS_chmod   Terminal_Command_String = "chmod"

	TCS_server   Terminal_Command_String = "server"
	TCS_backend  Terminal_Command_String = "backend"
	TCS_frontend Terminal_Command_String = "frontend"
	TCS_database Terminal_Command_String = "database"
	TCS_mavlink  Terminal_Command_String = "mavlink"
	TCS_log      Terminal_Command_String = "log"
	TCS_board    Terminal_Command_String = "board"
	TCS_drone    Terminal_Command_String = "drone"
	TCS_sensor   Terminal_Command_String = "sensor"

	TCS_auto     Terminal_Command_String = "auto"
	TCS_shutdown Terminal_Command_String = "shutdown"
	TCS_show     Terminal_Command_String = "show"
	TCS_reboot   Terminal_Command_String = "reboot"
	TCS_cache    Terminal_Command_String = "cache"
)

const (
	TCA_f           Terminal_Command_Args = "f"
	TCA_force       Terminal_Command_Args = "force"
	TCA_r           Terminal_Command_Args = "r"
	TCA_recursive   Terminal_Command_Args = "recursive"
	TCA_i           Terminal_Command_Args = "i"
	TCA_interactive Terminal_Command_Args = "interactive"
	TCA_d           Terminal_Command_Args = "d"
	TCA_directory   Terminal_Command_Args = "directory"
	TCA_w           Terminal_Command_Args = "w"

	TCA_x Terminal_Command_Args = "x"

	TCA_a    Terminal_Command_Args = "a"
	TCA_all  Terminal_Command_Args = "all"
	TCA_s    Terminal_Command_Args = "s"
	TCA_self Terminal_Command_Args = "self"

	TCA_auto Terminal_Command_Args = "auto"
)

const (
	TCO_create Terminal_Command_Object = "create"
	TCO_del    Terminal_Command_Object = "del"
	TCO_delete Terminal_Command_Object = "delete"
	TCO_update Terminal_Command_Object = "update"
	TCO_alter  Terminal_Command_Object = "alter"

	TCO_user     Terminal_Command_Object = "user"
	TCO_path     Terminal_Command_Object = "path"
	TCO_mode     Terminal_Command_Object = "mode"
	TCO_group    Terminal_Command_Object = "group"
	TCO_owner    Terminal_Command_Object = "owner"
	TCO_config   Terminal_Command_Object = "config"
	TCO_restart  Terminal_Command_Object = "restart"
	TCO_password Terminal_Command_Object = "password"
)

type TerminalCMD struct {
	Command Terminal_Command_String   `json:"command"`
	Objects []Terminal_Command_Object `json:"objects"`
	Args    []Terminal_Command_Args   `json:"args"`
}

type TerminalResponse struct {
	Success bool `json:"success"`
	Message map[string]interface{} `json:"message"`
}
