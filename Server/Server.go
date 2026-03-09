package Server

import (
	Backend "MavlinkProject/Server/backend"
)

var BackendServer Backend.BackendServer



func Server_start() {
	// 后端开启
	BackendServer.Start(":8080")
}