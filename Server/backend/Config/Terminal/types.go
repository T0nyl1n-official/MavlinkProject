package terminal

type (
	Terminal_Command_String string
	Terminal_Command_Args   string
	Terminal_Command_Object string
)

// Command String Standard
const (
	TCS_HelpBegin Terminal_Command_String = ""

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
	TCA_t           Terminal_Command_Args = "t"
	TCA_time        Terminal_Command_Args = "time"
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
	TCO_empty Terminal_Command_Object = ""

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
	TCO_shutdown Terminal_Command_Object = "shutdown"
	TCO_abort    Terminal_Command_Object = "abort"
	TCO_password Terminal_Command_Object = "password"
	TCO_cache    Terminal_Command_Object = "cache"

	TCO_server   Terminal_Command_Object = "server"
	TCO_backend  Terminal_Command_Object = "backend"
	TCO_frontend Terminal_Command_Object = "frontend"
	TCO_database Terminal_Command_Object = "database"
	TCO_mavlink  Terminal_Command_Object = "mavlink"
	TCO_log      Terminal_Command_Object = "log"
	TCO_board    Terminal_Command_Object = "board"
	TCO_drone    Terminal_Command_Object = "drone"
	TCO_sensor   Terminal_Command_Object = "sensor"
	TCO_landnode Terminal_Command_Object = "landnode"
	TCO_conn     Terminal_Command_Object = "connection"
	TCO_agent    Terminal_Command_Object = "agent"

	TCO_mysql Terminal_Command_Object = "mysql"
	TCO_redis Terminal_Command_Object = "redis"
)

type TerminalCMD struct {
	Command Terminal_Command_String   `json:"command"`
	Objects []Terminal_Command_Object `json:"objects"`
	Args    map[string]interface{}    `json:"args"`
	Path    string                    `json:"path"`
}

type TerminalResponse struct {
	Success bool        `json:"success"`
	Message interface{} `json:"message"`
}

var TCS_map_User = map[Terminal_Command_String]map[string]interface{}{
	TCS_HelpBegin: {
		"-": "Welcome to MavlinkProject Terminal!\nThere are commands now available:",
	},
	TCS_help: {
		"help":    "Show all permitted commands and details",
		"format":  "help ([command]/[pages])",
		"example": "help ls",
	},
	TCS_whoami: {
		"whoami":  "Show the current username and permission",
		"format":  "(no arguments)",
		"example": "whoami",
	},
	TCS_ls: {
		"ls":      "Show the current directory",
		"format":  "ls ([path])",
		"example": "ls Backend/Config",
	},
	TCS_cd: {
		"cd":      "Change directory",
		"format":  "cd [path]",
		"example": "cd Backend/Config",
	},
	TCS_mod: {
		"mod":     "Show the mod directory",
		"format":  "mod ([object]) [operation]  ([value]) [args]",
		"example": "mod show",
	},
	TCS_show: {
		"show":    "Show the object's details",
		"format":  "show [object] [command] [args]",
		"example": "show Server",
	},
	TCS_server: {
		"server":  "Show the server details",
		"format":  "server [command] [args]",
		"example": "server restart",
	},
	TCS_backend: {
		"backend": "Show the backend details",
		"format":  "backend [command] [args]",
		"example": "backend config",
	},
	TCS_frontend: {
		"frontend": "Show the frontend details",
		"format":   "frontend [command] [args]",
		"example":  "frontend config",
	},
	TCS_database: {
		"database": "Show the database details",
		"format":   "database [object] [command] [args]",
		"example":  "database redis add {test:testMessage} to 15 -f",
	},
	TCS_mavlink: {
		"mavlink": "Show the mavlink details",
		"format":  "mavlink [command] [args]",
	},
	TCS_log: {
		"log":     "Show Backend logs",
		"format1": "log [level] [args]",
		"format2": "log [beginTime] [endTime] [level] [args]",
	},
	TCS_board: {
		"board":  "Show the board details",
		"format": "board [object] [command] [args]",
	},
	TCS_drone: {
		"drone":  "Show the drone details",
		"format": "drone [object] [command] [args]",
	},
	TCS_sensor: {
		"sensor": "Show the sensor details",
		"format": "sensor [object] [command] [args]",
	},
}

var TCS_map_Admin = map[Terminal_Command_String]map[string]interface{}{
	TCS_HelpBegin: {
		"-": "Welcome to MavlinkProject Terminal!\nThere are commands now available:",
	},
	TCS_help: {
		"help":   "Show all permitted commands and details",
		"format": "help [pages]",
	},
	TCS_whoami: {
		"whoami": "Show the current username and permission",
		"format": "(no arguments)",
	},
	TCS_ls: {
		"ls":     "Show the current directory",
		"format": "ls [path]",
	},
	TCS_cd: {
		"cd":     "Change directory",
		"format": "cd [path]",
	},
	TCS_mod: {
		"mod":    "Show the mod directory",
		"format": "mod [CRUD-args] [object] ([value]) [args]",
	},
	TCS_show: {
		"show":   "Show the object's details",
		"format": "show [object] [command] [args]",
	},
	TCS_server: {
		"server": "Show the server details",
		"format": "server [command] [args]",
	},
	TCS_backend: {
		"backend": "Show the backend details",
		"format":  "backend [command] [args]",
	},
	TCS_frontend: {
		"frontend": "Show the frontend details",
		"format":   "frontend [command] [args]",
	},
	TCS_database: {
		"database": "Show the database details",
		"format":   "database [object] [command] [args]",
	},
	TCS_mavlink: {
		"mavlink": "Show the mavlink details",
		"format":  "mavlink [command] [args]",
	},
	TCS_log: {
		"log":     "Show Backend logs",
		"format1": "log [level] [args]",
		"format2": "log [beginTime] [endTime] [level] [args]",
	},
	TCS_board: {
		"board":  "Show the board details",
		"format": "board [object] [command] [args]",
	},
	TCS_drone: {
		"drone":  "Show the drone details",
		"format": "drone [object] [command] [args]",
	},
	TCS_sensor: {
		"sensor": "Show the sensor details",
		"format": "sensor [object] [command] [args]",
	},

	TCS_cache: {
		"cache":  "Show the cache details",
		"format": "cache [command] [args]",
	},
	TCS_adduser: {
		"adduser": "Add user with specified config",
		"format":  "adduser [newuser] [args]",
		"example": "adduser steve",
	},
	TCS_deluser: {
		"deluser": "Delete user with specified config",
		"format":  "deluser [user] [args]",
		"example": "deluser steve -f",
	},
	TCS_auto: {
		"auto":    "Using AI agent handle the Server automatically",
		"format":  "auto [AI-Agent object] [args]",
		"example": "auto Deepseekv3.5-turbo-16k",
	},
	TCS_reboot: {
		"reboot":  "Restart the server",
		"format":  "reboot [command] [args]",
		"example": "reboot -t 5",
	},
	TCS_shutdown: {
		"shutdown": "Shutdown the server",
		"format":   "shutdown ([object]) ([args])",
		"notes":    "Default object is Server, default times is 5, default args is empty",
		"example":  "shutdown -t 10",
	},
}

// Terminal_Help_Details 定义帮助命令的详细信息
var (
	THD_help = map[string]interface{}{
		"help":     "Show all permitted commands and details",
		"format1":  "help [pages]",
		"example1": "help 1",
		"format2":  "help [command]",
		"example2": "help mod",
	}

	THD_mod = map[string]interface{}{
		"mod":     "Show the mod directory",
		"format":  "mod ([object]) [operation]  ([value]) [args]",
		"example": "mod show",
	}

	THD_show = map[string]interface{}{
		"show":    "Show the object's details",
		"format":  "show [object] [command] [args]",
		"example": "show Server",
	}

	THD_server = map[string]interface{}{
		"server":   "Server management commands",
		"format1":  "server config",
		"example1": "server config",
		"format2":  "server restart [-t seconds]",
		"example2": "server restart -t 10",
		"format3":  "server shutdown [-t seconds]",
		"example3": "server shutdown -t 5",
	}

	THD_backend = map[string]interface{}{
		"backend":  "Backend service management",
		"format1":  "backend config",
		"example1": "backend config",
		"format2":  "backend restart [-t seconds]",
		"example2": "backend restart -t 10",
		"format3":  "backend shutdown [-t seconds]",
		"example3": "backend shutdown -t 5",
	}

	THD_frontend = map[string]interface{}{
		"frontend": "Frontend service management",
		"format1":  "frontend config",
		"example1": "frontend config",
		"format2":  "frontend restart [-t seconds]",
		"example2": "frontend restart -t 10",
		"format3":  "frontend shutdown [-t seconds]",
		"example3": "frontend shutdown -t 5",
	}

	THD_database = map[string]interface{}{
		"database": "Database configuration management",
		"format1":  "database",
		"example1": "database",
		"format2":  "database mysql",
		"example2": "database mysql",
		"format3":  "database redis",
		"example3": "database redis",
	}

	THD_mavlink = map[string]interface{}{
		"mavlink": "Mavlink protocol configuration",
		"format":  "mavlink config",
		"example": "mavlink config",
	}

	THD_board = map[string]interface{}{
		"board":    "Board connection management",
		"format1":  "board config",
		"example1": "board config",
		"format2":  "board restart",
		"example2": "board restart",
	}

	THD_log = map[string]interface{}{
		"log":      "Show Backend logs",
		"format1":  "log ([object]) ([type]/[level])",
		"example1": "log Backend error",
		"format2":  "log [beginTime] [endTime] ([type]/[level])",
		"example2": "log 2005-12-29-21:00:00 2005-12-29-22:00:00 200",
	}

	THD_drone = map[string]interface{}{
		"drone":    "Drone management",
		"format1":  "drone",
		"example1": "drone",
		"format2":  "drone [drone_id]",
		"example2": "drone drone_001",
	}

	THD_sensor = map[string]interface{}{
		"sensor":  "Sensor management",
		"format":  "sensor",
		"example": "sensor",
	}

	THD_cache = map[string]interface{}{
		"cache":   "Cache management",
		"format":  "cache config",
		"example": "cache config",
	}

	THD_adduser = map[string]interface{}{
		"adduser": "Add new user (admin only)",
		"format":  "adduser [username] [email]",
		"example": "adduser admin admin@example.com",
	}

	THD_deluser = map[string]interface{}{
		"deluser": "Delete user (admin only)",
		"format":  "deluser [username]",
		"example": "deluser testuser",
	}

	THD_auto = map[string]interface{}{
		"auto":    "AI agent auto management (admin only)",
		"format":  "auto [agent_name] [args]",
		"example": "auto mission_planner -enable",
	}

	THD_reboot = map[string]interface{}{
		"reboot":  "System reboot (admin only)",
		"format":  "reboot",
		"example": "reboot",
	}

	THD_shutdown = map[string]interface{}{
		"shutdown": "System shutdown (admin only)",
		"format":   "shutdown",
		"example":  "shutdown",
	}
)
