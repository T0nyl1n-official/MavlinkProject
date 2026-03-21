package BoardsRoutes

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	boardHandler "MavlinkProject/Server/Backend/Handler/Boards"
	Board "MavlinkProject/Server/Backend/Shared/Boards"

	JwtMiddleware "MavlinkProject/Server/backend/Middles"
	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
)

type BoardMessageRequest struct {
	ToID      string                 `json:"to_id"`
	ToType    string                 `json:"to_type"`
	Command   string                 `json:"command"`
	Data      map[string]interface{} `json:"data"`
	Attribute string                 `json:"attribute"`
}

type BoardCreateRequest struct {
	BoardID    string `json:"board_id"`
	BoardName  string `json:"board_name"`
	BoardType  string `json:"board_type"`
	Connection string `json:"connection"`
	Address    string `json:"address"`
	Port       string `json:"port"`
}

type BoardResponse struct {
	Success   bool          `json:"success"`
	Message   string        `json:"message,omitempty"`
	Error     string        `json:"error,omitempty"`
	HandlerID string        `json:"handler_id,omitempty"`
	Board     *Board.Board  `json:"board,omitempty"`
	Boards    []Board.Board `json:"boards,omitempty"`
}

func SetupBoardRoutes(router *gin.Engine, jwtManager *jwtUtils.JWTManager, tokenManager *Jwt.RedisTokenManager) {
	board := router.Group("/api/board")
	board.Use(JwtMiddleware.JwtAuthMiddleWareWithRedis(jwtManager, tokenManager, nil))
	{
		board.POST("/create", createBoardServer)
		board.POST("/start", startBoardServer)
		board.POST("/stop", stopBoardServer)
		board.POST("/send", sendMessageToBoard)
		board.POST("/forward", forwardMessage)
		board.GET("/list", listBoards)
		board.GET("/info/:boardID", getBoardInfo)
		board.POST("/auto-forward", enableAutoForward)
		board.DELETE("/delete/:boardID", deleteBoardServer)
	}
}

func createBoardServer(c *gin.Context) {
	var req BoardCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, BoardResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	manager := boardHandler.GetBoardManager()

	board := Board.Board{
		BoardID:   req.BoardID,
		BoardName: req.BoardName,
		BoardType: Board.BoardType(req.BoardType),
		BoardIP:   req.Address,
		BoardPort: req.Port,
	}

	var err error
	if req.Connection == "TCP" {
		err = manager.StartTCPServer(req.BoardID, req.Address, req.Port)
	} else if req.Connection == "UDP" {
		err = manager.StartUDPServer(req.BoardID, req.Address, req.Port)
	} else {
		c.JSON(http.StatusBadRequest, BoardResponse{
			Success: false,
			Error:   "Invalid connection type. Use TCP or UDP",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, BoardResponse{
			Success: false,
			Error:   "Failed to create board server: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BoardResponse{
		Success:   true,
		Message:   "Board server created successfully",
		HandlerID: req.BoardID,
		Board:     &board,
	})
}

func startBoardServer(c *gin.Context) {
	var req BoardCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, BoardResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	manager := boardHandler.GetBoardManager()
	var err error

	if req.Connection == "TCP" {
		err = manager.StartTCPServer(req.BoardID, req.Address, req.Port)
	} else if req.Connection == "UDP" {
		err = manager.StartUDPServer(req.BoardID, req.Address, req.Port)
	} else {
		c.JSON(http.StatusBadRequest, BoardResponse{
			Success: false,
			Error:   "Invalid connection type",
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, BoardResponse{
			Success: false,
			Error:   "Failed to start board server: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BoardResponse{
		Success: true,
		Message: "Board server started successfully",
	})
}

func stopBoardServer(c *gin.Context) {
	type StopRequest struct {
		BoardID string `json:"board_id"`
	}

	var req StopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, BoardResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	manager := boardHandler.GetBoardManager()
	if err := manager.StopBoard(req.BoardID); err != nil {
		c.JSON(http.StatusInternalServerError, BoardResponse{
			Success: false,
			Error:   "Failed to stop board server: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BoardResponse{
		Success: true,
		Message: "Board server stopped successfully",
	})
}

func sendMessageToBoard(c *gin.Context) {
	var req BoardMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, BoardResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	manager := boardHandler.GetBoardManager()

	msg := &Board.BoardMessage{
		MessageID:   generateMessageID(),
		MessageTime: time.Now(),
		Message: Board.Message{
			MessageType: "Request",
			Attribute:   Board.MessageAttribute(req.Attribute),
			Command:     req.Command,
			Data:        req.Data,
		},
		FromID:   "backend",
		FromType: "Control",
		ToID:     req.ToID,
		ToType:   req.ToType,
	}

	if err := manager.SendMessageToBoard(req.ToID, msg); err != nil {
		c.JSON(http.StatusInternalServerError, BoardResponse{
			Success: false,
			Error:   "Failed to send message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BoardResponse{
		Success: true,
		Message: "Message sent successfully",
	})
}

func forwardMessage(c *gin.Context) {
	type ForwardRequest struct {
		FromBoardID string                 `json:"from_board_id"`
		ToBoardID   string                 `json:"to_board_id"`
		Command     string                 `json:"command"`
		Data        map[string]interface{} `json:"data"`
		Attribute   string                 `json:"attribute"`
	}

	var req ForwardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, BoardResponse{
			Success: false,
			Error:   "Invalid request body: " + err.Error(),
		})
		return
	}

	manager := boardHandler.GetBoardManager()

	msg := &Board.BoardMessage{
		MessageID:   generateMessageID(),
		MessageTime: time.Now(),
		Message: Board.Message{
			MessageType: "Request",
			Attribute:   Board.MessageAttribute(req.Attribute),
			Command:     req.Command,
			Data:        req.Data,
		},
	}

	if err := manager.ForwardMessageToBoard(req.FromBoardID, req.ToBoardID, msg); err != nil {
		c.JSON(http.StatusInternalServerError, BoardResponse{
			Success: false,
			Error:   "Failed to forward message: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BoardResponse{
		Success: true,
		Message: "Message forwarded successfully",
	})
}

func listBoards(c *gin.Context) {
	manager := boardHandler.GetBoardManager()
	boards := manager.GetAllBoards()

	var boardList []Board.Board
	for _, bs := range boards {
		boardList = append(boardList, Board.Board{
			BoardID:     bs.BoardID,
			BoardIP:     bs.Addr,
			BoardPort:   bs.Port,
			BoardStatus: bs.Connection,
			IsConnected: bs.Connected,
		})
	}

	c.JSON(http.StatusOK, BoardResponse{
		Success: true,
		Boards:  boardList,
	})
}

func getBoardInfo(c *gin.Context) {
	boardID := c.Param("boardID")
	if boardID == "" {
		c.JSON(http.StatusBadRequest, BoardResponse{
			Success: false,
			Error:   "Board ID is required",
		})
		return
	}

	manager := boardHandler.GetBoardManager()
	connection, addr, port, connected := manager.GetBoardConnectionInfo(boardID)

	if !connected {
		c.JSON(http.StatusNotFound, BoardResponse{
			Success: false,
			Error:   "Board not found",
		})
		return
	}

	board := Board.Board{
		BoardID:     boardID,
		BoardIP:     addr,
		BoardPort:   port,
		BoardStatus: connection,
		IsConnected: true,
	}

	c.JSON(http.StatusOK, BoardResponse{
		Success: true,
		Board:   &board,
	})
}

func enableAutoForward(c *gin.Context) {
	manager := boardHandler.GetBoardManager()
	manager.EnableAutoForward()

	c.JSON(http.StatusOK, BoardResponse{
		Success: true,
		Message: "Auto-forward enabled",
	})
}

func deleteBoardServer(c *gin.Context) {
	boardID := c.Param("boardID")
	if boardID == "" {
		c.JSON(http.StatusBadRequest, BoardResponse{
			Success: false,
			Error:   "Board ID is required",
		})
		return
	}

	manager := boardHandler.GetBoardManager()
	if err := manager.StopBoard(boardID); err != nil {
		c.JSON(http.StatusInternalServerError, BoardResponse{
			Success: false,
			Error:   "Failed to delete board server: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, BoardResponse{
		Success: true,
		Message: "Board server deleted successfully",
	})
}

func generateMessageID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
		time.Sleep(time.Nanosecond)
	}
	return string(b)
}
