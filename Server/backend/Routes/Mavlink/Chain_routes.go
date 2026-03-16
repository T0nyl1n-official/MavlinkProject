package MavlinkRoute

import (
	"net/http"

	gin "github.com/gin-gonic/gin"

	Chain "MavlinkProject/Server/backend/Handler/ProgressChain"
	JwtMiddleware "MavlinkProject/Server/backend/Middles"
	Jwt "MavlinkProject/Server/backend/Middles/Jwt"
	jwtUtils "MavlinkProject/Server/backend/Middles/Jwt/Claims-Manager"
)

func SetupChainRoutes(router *gin.Engine, jwtManager *jwtUtils.JWTManager, tokenManager *Jwt.RedisTokenManager) {
	chainGroup := router.Group("/api/chain")

	chainGroup.Use(JwtMiddleware.JwtAuthMiddleWareWithRedis(jwtManager, tokenManager, nil))

	{
		chainGroup.POST("/create", createChain)
		chainGroup.DELETE("/:id", deleteChain)
		chainGroup.GET("/:id", getChain)
		chainGroup.GET("/list", listChains)

		chainGroup.POST("/:id/node/add", addNodeToChain)
		chainGroup.POST("/:id/node/delete/:nodeId", deleteNodeFromChain)

		chainGroup.POST("/:id/start", startChain)
		chainGroup.POST("/:id/reset", resetChain)
		chainGroup.POST("/:id/pause", pauseChain)
		chainGroup.POST("/:id/stop", stopChain)
	}
}

func createChain(c *gin.Context) {
	var req struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Name = "Unnamed Chain"
	}

	manager := Chain.GetChainManager()
	chain := manager.CreateChain(req.Name)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"chain_id":   chain.ID,
		"chain_name": chain.Name,
		"message":    "Chain created successfully",
	})
}

func deleteChain(c *gin.Context) {
	chainID := c.Param("id")
	manager := Chain.GetChainManager()
	err := manager.DeleteChain(chainID)

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Chain deleted successfully",
	})
}

func getChain(c *gin.Context) {
	chainID := c.Param("id")
	manager := Chain.GetChainManager()
	chain := manager.GetChain(chainID)

	if chain == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Chain not found",
		})
		return
	}

	jsonStr, _ := chain.ToJSON()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"chain":   jsonStr,
	})
}

func listChains(c *gin.Context) {
	manager := Chain.GetChainManager()
	chains := manager.GetAllChains()

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"chain_count": len(chains),
		"chains":      chains,
	})
}

func addNodeToChain(c *gin.Context) {
	chainID := c.Param("id")

	var req struct {
		NodeType      string                 `json:"node_type"`
		HandlerConfig *Chain.HandlerConfig   `json:"handler_config"`
		Params        map[string]interface{} `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	manager := Chain.GetChainManager()
	chain := manager.GetChain(chainID)
	if chain == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Chain not found",
		})
		return
	}

	nodeType := Chain.NodeType(req.NodeType)
	node := chain.AddNode(nodeType, req.HandlerConfig, req.Params)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"node_id": node.ID,
		"message": "Node added successfully",
	})
}

func deleteNodeFromChain(c *gin.Context) {
	chainID := c.Param("id")
	nodeID := c.Param("nodeId")

	manager := Chain.GetChainManager()
	chain := manager.GetChain(chainID)
	if chain == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Chain not found",
		})
		return
	}

	err := chain.RemoveNode(nodeID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Node deleted successfully",
	})
}

func startChain(c *gin.Context) {
	chainID := c.Param("id")

	manager := Chain.GetChainManager()
	chain := manager.GetChain(chainID)
	if chain == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Chain not found",
		})
		return
	}

	chain.Start()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Chain started successfully",
	})
}

func resetChain(c *gin.Context) {
	chainID := c.Param("id")

	manager := Chain.GetChainManager()
	chain := manager.GetChain(chainID)
	if chain == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Chain not found",
		})
		return
	}

	chain.Reset()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Chain reset successfully",
	})
}

func pauseChain(c *gin.Context) {
	chainID := c.Param("id")

	manager := Chain.GetChainManager()
	chain := manager.GetChain(chainID)
	if chain == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Chain not found",
		})
		return
	}

	chain.Pause()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Chain paused successfully",
	})
}

func stopChain(c *gin.Context) {
	chainID := c.Param("id")

	manager := Chain.GetChainManager()
	chain := manager.GetChain(chainID)
	if chain == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Chain not found",
		})
		return
	}

	chain.Stop()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Chain stopped successfully",
	})
}
