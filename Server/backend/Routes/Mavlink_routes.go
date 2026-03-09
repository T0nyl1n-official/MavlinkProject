package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	Mavlink "MavlinkProject/Server/backend/Handler/Mavlink"
)

func SetupMavlinkRoutes(router *gin.Engine) {
	mavlinkGroup := router.Group("/api/mavlink")
	{
		// 链管理CRUD接口
		mavlinkGroup.POST("/chains", createChain)
		mavlinkGroup.GET("/chains/current", getCurrentChain)
		mavlinkGroup.GET("/chains/:id", getChain)
		mavlinkGroup.PUT("/chains/current", switchChain)
		mavlinkGroup.POST("/chains/new", createNewChain)
		mavlinkGroup.DELETE("/chains/:id", deleteChain)
		mavlinkGroup.GET("/chains", listChains)
		mavlinkGroup.POST("/chains/records", addRecord)
		mavlinkGroup.GET("/chains/:id/records", getChainRecords)
		mavlinkGroup.GET("/chains/current/records", getCurrentChainRecords)
	}
}

func createChain(c *gin.Context) {
	manager := Mavlink.GetChainManager()
	chainID := manager.CreateChain()

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"chain_id": chainID,
	})
}

func getCurrentChain(c *gin.Context) {
	manager := Mavlink.GetChainManager()
	chain := manager.GetCurrentChain()

	if chain == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "no active chain",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"chain":   chain.GetInfo(),
	})
}

func getChain(c *gin.Context) {
	chainID := c.Param("id")
	manager := Mavlink.GetChainManager()
	chain := manager.GetChain(chainID)

	if chain == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "chain not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"chain":   chain.GetInfo(),
		"records": chain.GetRecords(),
	})
}

func switchChain(c *gin.Context) {
	var req struct {
		ChainID string `json:"chain_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	manager := Mavlink.GetChainManager()
	if err := manager.SwitchChain(req.ChainID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"chain_id": req.ChainID,
	})
}

func createNewChain(c *gin.Context) {
	manager := Mavlink.GetChainManager()
	chainID := manager.CreateNewChainAndSwitch()

	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"chain_id": chainID,
	})
}

func deleteChain(c *gin.Context) {
	chainID := c.Param("id")
	manager := Mavlink.GetChainManager()

	if err := manager.DeleteChain(chainID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

func listChains(c *gin.Context) {
	manager := Mavlink.GetChainManager()
	chains := manager.GetAllChains()

	result := make([]map[string]interface{}, 0, len(chains))
	for _, chain := range chains {
		result = append(result, chain.GetInfo())
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"chains":      result,
		"current_id":  manager.GetCurrentChainID(),
		"total_count": len(chains),
	})
}

func addRecord(c *gin.Context) {
	var req struct {
		HandlerID string `json:"handler_id" binding:"required"`
		Route     string `json:"route" binding:"required"`
		Params    string `json:"params"`
		Result    string `json:"result"`
		Success   bool   `json:"success"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	manager := Mavlink.GetChainManager()

	chain := manager.GetCurrentChain()
	if chain == nil {
		chainID := manager.CreateChain()
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"error":        "no active chain, created new one",
			"new_chain_id": chainID,
		})
		return
	}

	if chain.IsFull() {
		newChainID := manager.CreateNewChainAndSwitch()
		c.JSON(http.StatusBadRequest, gin.H{
			"success":      false,
			"error":        "chain is full, created new chain",
			"new_chain_id": newChainID,
		})
		return
	}

	err := manager.AddRecordToCurrentChain(req.HandlerID, req.Route, req.Params, req.Result, req.Success)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"record_done": true,
		"chain_info":  chain.GetInfo(),
	})
}

func getRecords(c *gin.Context) {
	chainID := c.Query("chain_id")
	manager := Mavlink.GetChainManager()

	var chain *Mavlink.DispatchChain
	if chainID != "" {
		chain = manager.GetChain(chainID)
	} else {
		chain = manager.GetCurrentChain()
	}

	if chain == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "chain not found",
		})
		return
	}

	records := chain.GetRecords()

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"chain_id":     chain.ChainID,
		"chain_status": chain.Status,
		"total_count":  len(records),
		"records":      records,
	})
}
