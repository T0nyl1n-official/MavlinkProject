package ProgressChain

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	Mavlink "MavlinkProject/Server/backend/Handler/Mavlink"
)
// ChainExecutor 执行链中的节点操作
type ChainExecutor struct {
	mu      sync.RWMutex
	chain   *Chain
	manager *ChainManager
}

func NewChainExecutor(chain *Chain, manager *ChainManager) *ChainExecutor {
	return &ChainExecutor{
		chain:   chain,
		manager: manager,
	}
}

func (ce *ChainExecutor) ExecuteNode(node *Node, ctx *gin.Context) error {
	now := time.Now()
	node.Status = NodeStatusRunning
	node.StartedAt = &now

	var err error
	var result string

	switch node.Type {
	case NodeTypeCreateHandler:
		err = ce.executeHandlerOperation(node, "create")
	case NodeTypeDeleteHandler:
		err = ce.executeHandlerOperation(node, "delete")
	case NodeTypeUpdateHandler:
		err = ce.executeHandlerOperation(node, "update")
	case NodeTypeConnectionStart:
		err = ce.executeConnection(node, "start")
	case NodeTypeConnectionStop:
		err = ce.executeConnection(node, "stop")
	case NodeTypeConnectionRestart:
		err = ce.executeConnection(node, "restart")
	case NodeTypeDroneStatus:
		result, err = ce.executeStatus(node)
	case NodeTypeDroneTakeoff:
		err = ce.executeDroneTakeoff(node)
	case NodeTypeDroneLand:
		err = ce.executeDroneLand(node)
	case NodeTypeDroneMove:
		err = ce.executeDroneMove(node)
	case NodeTypeDroneReturn:
		err = ce.executeDroneReturn(node)
	case NodeTypeDroneMode:
		err = ce.executeDroneMode(node)
	case NodeTypeGroundStationSet:
		err = ce.executeGroundStationSet(node)
	case NodeTypeTaskVerification:
		result, err = ce.executeTaskVerification(node)
	case NodeTypeStreamRequest:
		err = ce.executeStreamRequest(node)
	case NodeTypeHeartbeatSend:
		err = ce.executeHeartbeatSend(node)
	default:
		err = fmt.Errorf("unknown node type: %s", node.Type)
	}

	finishTime := time.Now()
	node.FinishedAt = &finishTime

	if err != nil {
		node.Status = NodeStatusError
		node.Error = err.Error()
	} else {
		node.Status = NodeStatusFinished
		node.Result = result
	}

	return err
}

func (ce *ChainExecutor) getHandlerFromNode(node *Node) (*Mavlink.MAVLinkHandlerV1, error) {
	if node.HandlerConfig == nil {
		return nil, fmt.Errorf("node %s has no handler config", node.ID)
	}

	config := Mavlink.MAVLinkConfigV1{
		ConnectionType:  Mavlink.ConnectionType(node.HandlerConfig.ConnectionType),
		SerialPort:      node.HandlerConfig.SerialPort,
		SerialBaud:      node.HandlerConfig.SerialBaud,
		UDPAddr:         node.HandlerConfig.UDPAddr,
		UDPPort:         node.HandlerConfig.UDPPort,
		TCPAddr:         node.HandlerConfig.TCPAddr,
		TCPPort:         node.HandlerConfig.TCPPort,
		SystemID:        node.HandlerConfig.SystemID,
		ComponentID:     node.HandlerConfig.ComponentID,
		ProtocolVersion: Mavlink.ProtocolVersion(node.HandlerConfig.ProtocolVersion),
		HeartbeatRate:   node.HandlerConfig.HeartbeatRate,
	}

	return Mavlink.NewMAVLinkHandlerV1(config), nil
}

func (ce *ChainExecutor) executeHandlerOperation(node *Node, operation string) error {
	switch operation {
	case "create":
		return nil
	case "delete":
		return nil
	case "update":
		return nil
	}
	return nil
}

func (ce *ChainExecutor) executeConnection(node *Node, action string) error {
	handler, err := ce.getHandlerFromNode(node)
	if err != nil {
		return err
	}

	switch action {
	case "start":
		return handler.Start()
	case "stop":
		return handler.Stop()
	case "restart":
		return handler.Restart()
	}
	return nil
}

func (ce *ChainExecutor) executeStatus(node *Node) (string, error) {
	handler, err := ce.getHandlerFromNode(node)
	if err != nil {
		return "", err
	}

	status := handler.GetDroneStatus()
	position := handler.GetDronePosition()
	attitude := handler.GetDroneAttitude()
	battery := handler.GetDroneBattery()

	return fmt.Sprintf("status: %+v, position: %+v, attitude: %+v, battery: %+v", status, position, attitude, battery), nil
}

func (ce *ChainExecutor) executeDroneTakeoff(node *Node) error {
	handler, err := ce.getHandlerFromNode(node)
	if err != nil {
		return err
	}

	altitude, ok := node.Params["altitude"].(float64)
	if !ok {
		return fmt.Errorf("invalid altitude parameter")
	}

	return handler.SendTakeoff(float32(altitude))
}

func (ce *ChainExecutor) executeDroneLand(node *Node) error {
	handler, err := ce.getHandlerFromNode(node)
	if err != nil {
		return err
	}

	lat, _ := node.Params["latitude"].(float64)
	lon, _ := node.Params["longitude"].(float64)
	alt, _ := node.Params["altitude"].(float64)

	return handler.SendLand(lat, lon, alt)
}

func (ce *ChainExecutor) executeDroneMove(node *Node) error {
	handler, err := ce.getHandlerFromNode(node)
	if err != nil {
		return err
	}

	lat, latOk := node.Params["latitude"].(float64)
	lon, lonOk := node.Params["longitude"].(float64)
	alt, altOk := node.Params["altitude"].(float64)

	if !latOk || !lonOk || !altOk {
		return fmt.Errorf("invalid position parameters")
	}

	speed := float32(5.0)
	if s, ok := node.Params["speed"].(float64); ok {
		speed = float32(s)
	}

	return handler.SendMoveToPosition(lat, lon, alt, speed)
}

func (ce *ChainExecutor) executeDroneReturn(node *Node) error {
	handler, err := ce.getHandlerFromNode(node)
	if err != nil {
		return err
	}

	return handler.SendReturnToLaunch()
}

func (ce *ChainExecutor) executeDroneMode(node *Node) error {
	handler, err := ce.getHandlerFromNode(node)
	if err != nil {
		return err
	}

	modeStr, ok := node.Params["mode"].(string)
	if !ok {
		return fmt.Errorf("invalid mode parameter")
	}

	return handler.SetFlightMode(Mavlink.FlightMode(modeStr))
}

func (ce *ChainExecutor) executeGroundStationSet(node *Node) error {
	handler, err := ce.getHandlerFromNode(node)
	if err != nil {
		return err
	}

	name, _ := node.Params["name"].(string)
	id, _ := node.Params["id"].(string)
	lat, _ := node.Params["latitude"].(float64)
	lon, _ := node.Params["longitude"].(float64)
	alt, _ := node.Params["altitude"].(float64)

	handler.SetGroundStation(name, id, lat, lon, alt)
	return nil
}

func (ce *ChainExecutor) executeTaskVerification(node *Node) (string, error) {
	result, ok := node.Params["result"].(string)
	if !ok {
		return "Task verification: no specific result required", nil
	}
	return fmt.Sprintf("Task verified: %s", result), nil
}

func (ce *ChainExecutor) executeStreamRequest(node *Node) error {
	handler, err := ce.getHandlerFromNode(node)
	if err != nil {
		return err
	}

	msgID, ok := node.Params["message_id"].(int)
	rate, rateOk := node.Params["rate"].(int)

	if !ok || !rateOk {
		return fmt.Errorf("invalid stream parameters")
	}

	return handler.RequestMessageStream(msgID, rate)
}

func (ce *ChainExecutor) executeHeartbeatSend(node *Node) error {
	handler, err := ce.getHandlerFromNode(node)
	if err != nil {
		return err
	}

	return handler.SendHeartbeat()
}

func (ce *ChainExecutor) ExecuteAll(ctx *gin.Context) error {
	ce.chain.Start()

	current := ce.chain.Head
	for current != nil {
		ce.chain.SetCurrentNode(current)

		if err := ce.ExecuteNode(current, ctx); err != nil {
			ce.chain.Status = "error"
			ce.chain.UpdatedAt = time.Now()
			ce.saveChainLog()
			return err
		}

		current = current.Next
	}

	ce.chain.Status = "completed"
	ce.chain.UpdatedAt = time.Now()
	ce.saveChainLog()
	return nil
}

func (ce *ChainExecutor) saveChainLog() {
	dir := ensureChainLogDir()
	if err := ce.chain.SaveToFile(dir); err != nil {
		fmt.Printf("Failed to save chain log: %v\n", err)
	}
}

func (ce *ChainExecutor) GetChainJSON() (string, error) {
	return ce.chain.ToPrettyJSON()
}

type ChainContextKey string

const ChainContextKeyVal ChainContextKey = "progress_chain"

func GetChainFromContext(c *gin.Context) *Chain {
	if val, exists := c.Get(string(ChainContextKeyVal)); exists {
		if chain, ok := val.(*Chain); ok {
			return chain
		}
	}
	return nil
}

func SetChainToContext(c *gin.Context, chain *Chain) {
	c.Set(string(ChainContextKeyVal), chain)
}

// 链日志记录
const (
	DefaultChainLogDir = "logs/chains"
)

func ensureChainLogDir() string {
	dir := DefaultChainLogDir
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		os.MkdirAll(dir, 0755)
	}
	return dir
}

