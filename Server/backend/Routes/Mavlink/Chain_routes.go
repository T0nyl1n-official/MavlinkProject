package MavlinkRoute

import (
	"net/http"

	gin "github.com/gin-gonic/gin"

	Chain "MavlinkProject/Server/backend/Handler/ProgressChain"
)

func SetupChainRoutes(router *gin.Engine) {
	chainGroup := router.Group("/api/chain")
	{
		chainGroup.POST("/create", createChain)
		chainGroup.DELETE("/:id", deleteChain)
		chainGroup.GET("/:id", getChain)
		chainGroup.GET("/list", listChains)

		chainGroup.POST("/:id/node/add", addNodeToChain)
		chainGroup.POST("/:id/node/insert-after", insertNodeAfter)
		chainGroup.POST("/:id/node/insert-before", insertNodeBefore)
		chainGroup.PUT("/:id/node/:nodeId/update-config", updateNodeConfig)
		chainGroup.DELETE("/:id/node/:nodeId", deleteNodeFromChain)

		chainGroup.POST("/:id/execute", executeChain)
		chainGroup.POST("/:id/execute-step", executeNextStep)
		chainGroup.POST("/:id/reset", resetChain)
		chainGroup.POST("/:id/pause", pauseChain)
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

	if err := manager.DeleteChain(chainID); err != nil {
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

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"chain":          chain,
		"node_count":     chain.GetNodeCount(),
		"waiting_count":  chain.GetWaitingNodeCount(),
		"finished_count": chain.GetFinishedNodeCount(),
	})
}

func listChains(c *gin.Context) {
	manager := Chain.GetChainManager()
	chains := manager.GetAllChains()

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"chains":      chains,
		"total_count": len(chains),
	})
}

func addNodeToChain(c *gin.Context) {
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

	var req struct {
		Type          string                       `json:"type" binding:"required"`
		HandlerConfig *Chain.HandlerConfig `json:"handler_config"`
		Params        map[string]interface{}       `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	nodeType := Chain.NodeType(req.Type)
	node := chain.AddNode(nodeType, req.HandlerConfig, req.Params)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"node_id":    node.ID,
		"node_type":  node.Type,
		"node_count": chain.GetNodeCount(),
	})
}

func insertNodeAfter(c *gin.Context) {
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

	var req struct {
		AfterNodeID   string                       `json:"after_node_id" binding:"required"`
		Type          string                       `json:"type" binding:"required"`
		HandlerConfig *Chain.HandlerConfig `json:"handler_config"`
		Params        map[string]interface{}       `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	nodeType := Chain.NodeType(req.Type)
	node, err := chain.InsertNodeAfter(req.AfterNodeID, nodeType, req.HandlerConfig, req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"node_id":    node.ID,
		"node_type":  node.Type,
		"node_count": chain.GetNodeCount(),
	})
}

func insertNodeBefore(c *gin.Context) {
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

	var req struct {
		BeforeNodeID  string                       `json:"before_node_id" binding:"required"`
		Type          string                       `json:"type" binding:"required"`
		HandlerConfig *Chain.HandlerConfig `json:"handler_config"`
		Params        map[string]interface{}       `json:"params"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	nodeType := Chain.NodeType(req.Type)
	node, err := chain.InsertNodeBefore(req.BeforeNodeID, nodeType, req.HandlerConfig, req.Params)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"node_id":    node.ID,
		"node_type":  node.Type,
		"node_count": chain.GetNodeCount(),
	})
}

func updateNodeConfig(c *gin.Context) {
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

	var req struct {
		HandlerConfig *Chain.HandlerConfig `json:"handler_config" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	if err := chain.UpdateNodeHandlerConfig(nodeID, req.HandlerConfig); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Node config updated and propagated to subsequent nodes",
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

	if err := chain.RemoveNode(nodeID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"message":    "Node deleted",
		"node_count": chain.GetNodeCount(),
	})
}

func executeChain(c *gin.Context) {
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

	executor := Chain.NewChainExecutor(chain, manager)

	Chain.SetChainToContext(c, chain)

	if err := executor.ExecuteAll(c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success":    false,
			"error":      err.Error(),
			"chain":      chain,
			"node_count": chain.GetNodeCount(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":        true,
		"message":        "Chain executed successfully",
		"chain":          chain,
		"finished_count": chain.GetFinishedNodeCount(),
		"node_count":     chain.GetNodeCount(),
	})
}

func executeNextStep(c *gin.Context) {
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

	executor := Chain.NewChainExecutor(chain, manager)
	nextNode := chain.GetNextNode()

	if nextNode == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success":    false,
			"error":      "No more nodes to execute",
			"node_count": chain.GetNodeCount(),
		})
		return
	}

	Chain.SetChainToContext(c, chain)

	if err := executor.ExecuteNode(nextNode, c); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   err.Error(),
			"node":    nextNode,
		})
		return
	}

	chain.MoveToNextNode()

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"message":       "Node executed",
		"node":          nextNode,
		"current_index": chain.CurrentIndex,
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
		"success":    true,
		"message":    "Chain reset",
		"node_count": chain.GetNodeCount(),
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
		"status":  chain.Status,
	})
}
