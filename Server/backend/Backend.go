package Backend

import (
	"log"

	gin "github.com/gin-gonic/gin"
)

type BackendServer struct {
	Router *gin.Engine
}

func NewBackendServer() *BackendServer {
	router := gin.Default()
	return &BackendServer{Router: router}
}

func RunBackendServer(server *BackendServer, port string) {
	server.Router.Run(port)
	log.Printf("Backend server started on port %s", port)
}

// 被整合的Backend创建方法
func StartBackend(port string) *BackendServer {
	backend := NewBackendServer()
	RunBackendServer(backend, port)
	return backend
}
