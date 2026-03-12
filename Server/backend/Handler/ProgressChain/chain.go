package ProgressChain

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

type ChainManager struct {
	mu     sync.RWMutex
	chains map[string]*Chain
	config *ChainConfig
}

var (
	defaultChainManager *ChainManager
	chainManagerOnce    sync.Once
)

func GetChainManager() *ChainManager {
	chainManagerOnce.Do(func() {
		defaultChainManager = &ChainManager{
			chains: make(map[string]*Chain),
			config: NewChainConfig(),
		}
	})
	return defaultChainManager
}

// GenerateChainID 生成唯一的链ID
func GenerateChainID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// GenerateNodeID 生成唯一的节点ID
func GenerateNodeID() string {
	b := make([]byte, 4)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// 创建链
func NewChain(name string) *Chain {
	chain := &Chain{
		ID:           GenerateChainID(),
		Name:         name,
		Nodes:        make([]*Node, 0),
		CurrentNode:  nil,
		CurrentIndex: -1,
		Status:       "created",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	return chain
}

// 带入链配置创建链
func NewChainWithConfig(name string, config *ChainConfig) *Chain {
	chain := NewChain(name)
	if config != nil {
		chain.Status = fmt.Sprintf("max_nodes:%d,timeout:%v,auto_continue:%v", config.MaxNodes, config.Timeout, config.AutoContinue)
	}
	return chain
}

// 增加ProgressChain链中节点 *Node
func (c *Chain) AddNode(nodeType NodeType, handlerConfig *HandlerConfig, params map[string]interface{}) *Node {
	node := &Node{
		ID:            GenerateNodeID(),
		Type:          nodeType,
		Status:        NodeStatusWaiting,
		HandlerConfig: handlerConfig,
		Params:        params,
	}

	if c.Head == nil {
		c.Head = node
		c.Tail = node
	} else {
		c.Tail.Next = node
		node.Prev = c.Tail
		c.Tail = node
	}

	c.Nodes = append(c.Nodes, node)
	c.UpdatedAt = time.Now()

	return node
}

// 在ProgressChain链中当前运作节点后插入新节点
func (c *Chain) InsertNodeAfter(currentNodeID string, newNodeType NodeType, handlerConfig *HandlerConfig, params map[string]interface{}) (*Node, error) {
	currentNode := c.FindNode(currentNodeID)
	if currentNode == nil {
		return nil, fmt.Errorf("node with id %s not found", currentNodeID)
	}

	newNode := &Node{
		ID:            GenerateNodeID(),
		Type:          newNodeType,
		Status:        NodeStatusWaiting,
		HandlerConfig: handlerConfig,
		Params:        params,
	}

	// 防止最后一个任务后面再加任务
	if currentNode.Next != nil {
		currentNode.Next.Prev = newNode
		newNode.Next = currentNode.Next
	}

	newNode.Prev = currentNode
	currentNode.Next = newNode

	if currentNode == c.Tail {
		c.Tail = newNode
	}

	c.rebuildNodesArray()
	c.UpdatedAt = time.Now()

	return newNode, nil
}

// 在ProgressChain链中当前运作节点前插入新节点
func (c *Chain) InsertNodeBefore(currentNodeID string, newNodeType NodeType, handlerConfig *HandlerConfig, params map[string]interface{}) (*Node, error) {
	currentNode := c.FindNode(currentNodeID)
	if currentNode == nil {
		return nil, fmt.Errorf("node with id %s not found", currentNodeID)
	}

	newNode := &Node{
		ID:            GenerateNodeID(),
		Type:          newNodeType,
		Status:        NodeStatusWaiting,
		HandlerConfig: handlerConfig,
		Params:        params,
	}

	// 防止不是第一个节点前插入
	if currentNode.Prev != nil {
		currentNode.Prev.Next = newNode
		newNode.Prev = currentNode.Prev
	}

	newNode.Next = currentNode
	currentNode.Prev = newNode

	if currentNode == c.Head {
		c.Head = newNode
	}

	c.rebuildNodesArray()
	c.UpdatedAt = time.Now()

	return newNode, nil
}

// 更新ProgressChain链中节点的MavlinkHandler事件配置
func (c *Chain) UpdateNodeHandlerConfig(nodeID string, handlerConfig *HandlerConfig) error {
	node := c.FindNode(nodeID)
	if node == nil {
		return fmt.Errorf("node with id %s not found", nodeID)
	}

	node.HandlerConfig = handlerConfig

	c.propagateHandlerConfig(node)
	c.UpdatedAt = time.Now()

	return nil
}

// 从当前节点传播MavlinkHandler事件配置到后续节点 (UpdateNodeHandlerConfig 子方法)
func (c *Chain) propagateHandlerConfig(fromNode *Node) {
	current := fromNode.Next
	for current != nil {
		if current.Status == NodeStatusWaiting {
			current.HandlerConfig = fromNode.HandlerConfig
		}
		current = current.Next
	}
}

// 从ProgressChain链中移除某个节点(以nodeID获取节点位置, 遍历 O(n) )
func (c *Chain) RemoveNode(nodeID string) error {
	node := c.FindNode(nodeID)
	if node == nil {
		return fmt.Errorf("node with id %s not found", nodeID)
	}

	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		c.Head = node.Next
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		c.Tail = node.Prev
	}

	c.rebuildNodesArray()
	c.UpdatedAt = time.Now()

	return nil
}

// 从ProgressChain链中根据nodeID查找节点(遍历 O(n) )
func (c *Chain) FindNode(nodeID string) *Node {
	current := c.Head
	for current != nil {
		if current.ID == nodeID {
			return current
		}
		current = current.Next
	}
	return nil
}

// 从ProgressChain链中重建节点数组
func (c *Chain) rebuildNodesArray() {
	c.Nodes = make([]*Node, 0)
	current := c.Head
	for current != nil {
		c.Nodes = append(c.Nodes, current)
		current = current.Next
	}
}

func (c *Chain) GetCurrentNode() *Node {
	return c.CurrentNode
}

func (c *Chain) SetCurrentNode(node *Node) {
	c.CurrentNode = node
	c.UpdatedAt = time.Now()
}

func (c *Chain) GetNextNode() *Node {
	if c.CurrentNode == nil {
		return c.Head
	}
	return c.CurrentNode.Next
}

func (c *Chain) MoveToNextNode() bool {
	if c.CurrentNode == nil {
		if c.Head != nil {
			c.CurrentNode = c.Head
			c.CurrentIndex = 0
			c.UpdatedAt = time.Now()
			return true
		}
		return false
	}

	if c.CurrentNode.Next != nil {
		c.CurrentNode = c.CurrentNode.Next
		c.CurrentIndex++
		c.UpdatedAt = time.Now()
		return true
	}
	return false
}

func (c *Chain) GetNodeCount() int {
	return len(c.Nodes)
}

func (c *Chain) GetWaitingNodeCount() int {
	count := 0
	current := c.Head
	for current != nil {
		if current.Status == NodeStatusWaiting {
			count++
		}
		current = current.Next
	}
	return count
}

func (c *Chain) GetFinishedNodeCount() int {
	count := 0
	current := c.Head
	for current != nil {
		if current.Status == NodeStatusFinished {
			count++
		}
		current = current.Next
	}
	return count
}

func (c *Chain) IsComplete() bool {
	return c.GetWaitingNodeCount() == 0 && c.CurrentNode != nil && c.CurrentNode.Next == nil
}

func (c *Chain) Start() {
	c.Status = "running"
	c.UpdatedAt = time.Now()
}

func (c *Chain) Pause() {
	c.Status = "paused"
	c.UpdatedAt = time.Now()
}

func (c *Chain) Stop() {
	c.Status = "stopped"
	c.UpdatedAt = time.Now()
}

func (c *Chain) Reset() {
	c.CurrentNode = nil
	c.CurrentIndex = -1
	c.Status = "created"
	current := c.Head
	for current != nil {
		current.Status = NodeStatusWaiting
		current.Result = ""
		current.Error = ""
		current.StartedAt = nil
		current.FinishedAt = nil
		current = current.Next
	}
	c.UpdatedAt = time.Now()
}

// ===== ProgressChain链 CRUD =====

func (cm *ChainManager) CreateChain(name string) *Chain {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	chain := NewChain(name)
	cm.chains[chain.ID] = chain

	return chain
}

func (cm *ChainManager) GetChain(chainID string) *Chain {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.chains[chainID]
}

func (cm *ChainManager) DeleteChain(chainID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if _, ok := cm.chains[chainID]; !ok {
		return fmt.Errorf("chain with id %s not found", chainID)
	}

	delete(cm.chains, chainID)
	return nil
}

func (cm *ChainManager) GetAllChains() []*Chain {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	result := make([]*Chain, 0, len(cm.chains))
	for _, chain := range cm.chains {
		result = append(result, chain)
	}
	return result
}

func (cm *ChainManager) GetChainCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.chains)
}

func (c *Chain) ToJSON() (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// 供存储的美化json输出
func (c *Chain) ToPrettyJSON() (string, error) {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// 保存链日志json
func (c *Chain) SaveToFile(dir string) error {
	jsonData, err := c.ToPrettyJSON()
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("%s/%s_%s.json", dir, c.Name, c.ID)
	return os.WriteFile(filename, []byte(jsonData), 0644)
}
