package WarningHandler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	Server "MavlinkProject/Server"
	Warning_Agent "MavlinkProject/Server/backend/Shared/Warnings/Agent"
	Warning_Backend "MavlinkProject/Server/backend/Shared/Warnings/Backend"
	Warning_Drone "MavlinkProject/Server/backend/Shared/Warnings/Drone"
	Warning_Frontend "MavlinkProject/Server/backend/Shared/Warnings/Frontend"
	Warning_General "MavlinkProject/Server/backend/Shared/Warnings/General"
	Warning_Sensor "MavlinkProject/Server/backend/Shared/Warnings/Sensor"

	"github.com/redis/go-redis/v9"
)

var BackendServer = Server.BackendServer

/*
================================================================================
错误分配器 - WarningHandler
================================================================================

本模块提供错误分类和分配功能，根据传入的错误信息自动识别错误类型，
并将错误包装为特定的告警结构体存储到对应的 Redis Database 中。

使用方式:
---------
1. 直接传入错误信息:
   WarningHandler.DistributeError("database connection failed", "mysql")

2. 传入带上下文的错误:
   errorDetail := WarningHandler.ErrorDetail{
       Error:       err,
       ErrorType:   WarningHandler.ErrorTypeDatabase,
       Source:      "MysqlService",
       Details:     "connection timeout",
   }
   WarningHandler.Distribute(errorDetail)

3. 传入特定类型的错误:
   WarningHandler.HandleBackendError("database connection failed", "MysqlService")
   WarningHandler.HandleDroneError("battery low", "drone-001", "handler-001")

Redis DB 分配:
--------------
- DB0 (GeneralWarning): 通用告警/未知错误/未分类错误
- DB1 (Backend): 后端相关错误 (API/数据库/认证/配置等)
- DB2 (Frontend): 前端相关错误 (渲染/状态/组件等)
- DB3 (Agent): AI/代理相关错误 (执行/规划/超时等)
- DB4 (Drone): 无人机相关错误 (连接/电池/GPS/通信等)
- DB5 (Sensor): 传感器相关错误 (温度/湿度/气体等)
================================================================================
*/

type ErrorType string

const (
	ErrorTypeUnknown    ErrorType = "unknown"
	ErrorTypeDatabase   ErrorType = "database"
	ErrorTypeAPI        ErrorType = "api"
	ErrorTypeAuth       ErrorType = "auth"
	ErrorTypeRoute      ErrorType = "route"
	ErrorTypeConfig     ErrorType = "config"
	ErrorTypeConnection ErrorType = "connection"
	ErrorTypeTimeout    ErrorType = "timeout"
	ErrorTypeMavlink    ErrorType = "mavlink"
	ErrorTypeDrone      ErrorType = "drone"
	ErrorTypeSensor     ErrorType = "sensor"
	ErrorTypeAgent      ErrorType = "agent"
	ErrorTypeFrontend   ErrorType = "frontend"
	ErrorTypeBackend    ErrorType = "backend"
)

type ErrorDetail struct {
	Error      error
	ErrorType  ErrorType
	Source     string
	Details    string
	DroneID    string
	HandlerID  string
	SensorID   string
	SensorType string
	AgentID    string
	TaskID     string
	Component  string
	Value      string
	Threshold  string
}

var errorTypeMapping = map[ErrorType]DBIndex{
	ErrorTypeDatabase:   DBBackend,
	ErrorTypeAPI:        DBBackend,
	ErrorTypeAuth:       DBBackend,
	ErrorTypeRoute:      DBBackend,
	ErrorTypeConfig:     DBBackend,
	ErrorTypeConnection: DBBackend,
	ErrorTypeTimeout:    DBBackend,
	ErrorTypeMavlink:    DBDrone,
	ErrorTypeDrone:      DBDrone,
	ErrorTypeSensor:     DBSensor,
	ErrorTypeAgent:      DBAgent,
	ErrorTypeFrontend:   DBFrontend,
	ErrorTypeBackend:    DBBackend,
	ErrorTypeUnknown:    DBGeneralWarning,
}

type DBIndex int

const (
	DBGeneralWarning DBIndex = 0
	DBBackend        DBIndex = 1
	DBFrontend       DBIndex = 2
	DBAgent          DBIndex = 3
	DBDrone          DBIndex = 4
	DBSensor         DBIndex = 5
)

func GetRedisClient(dbIndex DBIndex) *redis.Client {
	redisClients := BackendServer.RedisClient
	if redisClients == nil {
		return nil
	}
	if int(dbIndex) >= len(*redisClients) {
		return nil
	}
	client := (*redisClients)[dbIndex]
	return &client
}

func ClassifyError(err error) ErrorType {
	if err == nil {
		return ErrorTypeUnknown
	}

	errStr := strings.ToLower(err.Error())

	if strings.Contains(errStr, "database") ||
		strings.Contains(errStr, "mysql") ||
		strings.Contains(errStr, "sql") ||
		strings.Contains(errStr, "gorm") ||
		strings.Contains(errStr, "redis") {
		return ErrorTypeDatabase
	}

	if strings.Contains(errStr, "api") ||
		strings.Contains(errStr, "http") ||
		strings.Contains(errStr, "request") ||
		strings.Contains(errStr, "response") {
		return ErrorTypeAPI
	}

	if strings.Contains(errStr, "auth") ||
		strings.Contains(errStr, "jwt") ||
		strings.Contains(errStr, "token") ||
		strings.Contains(errStr, "permission") ||
		strings.Contains(errStr, "unauthorized") {
		return ErrorTypeAuth
	}

	if strings.Contains(errStr, "route") ||
		strings.Contains(errStr, "router") ||
		strings.Contains(errStr, "not found") ||
		strings.Contains(errStr, "404") {
		return ErrorTypeRoute
	}

	if strings.Contains(errStr, "config") ||
		strings.Contains(errStr, "configuration") ||
		strings.Contains(errStr, "setting") {
		return ErrorTypeConfig
	}

	if strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "connect") ||
		strings.Contains(errStr, "dial") ||
		strings.Contains(errStr, "tcp") ||
		strings.Contains(errStr, "udp") {
		return ErrorTypeConnection
	}

	if strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "deadline") ||
		strings.Contains(errStr, "timed out") {
		return ErrorTypeTimeout
	}

	if strings.Contains(errStr, "mavlink") ||
		strings.Contains(errStr, "mavlink") ||
		strings.Contains(errStr, "heartbeat") ||
		strings.Contains(errStr, "dialect") {
		return ErrorTypeMavlink
	}

	if strings.Contains(errStr, "drone") ||
		strings.Contains(errStr, "battery") ||
		strings.Contains(errStr, "gps") ||
		strings.Contains(errStr, "motor") ||
		strings.Contains(errStr, "flight") {
		return ErrorTypeDrone
	}

	if strings.Contains(errStr, "sensor") ||
		strings.Contains(errStr, "temperature") ||
		strings.Contains(errStr, "humidity") ||
		strings.Contains(errStr, "pressure") ||
		strings.Contains(errStr, "gas") {
		return ErrorTypeSensor
	}

	if strings.Contains(errStr, "agent") ||
		strings.Contains(errStr, "ai") ||
		strings.Contains(errStr, "planning") ||
		strings.Contains(errStr, "execution") {
		return ErrorTypeAgent
	}

	if strings.Contains(errStr, "frontend") ||
		strings.Contains(errStr, "render") ||
		strings.Contains(errStr, "component") ||
		strings.Contains(errStr, "state") {
		return ErrorTypeFrontend
	}

	return ErrorTypeUnknown
}

func DistributeError(errorMsg string, source string) ErrorType {
	err := fmt.Errorf(errorMsg)
	detail := ErrorDetail{
		Error:     err,
		ErrorType: ClassifyError(err),
		Source:    source,
		Details:   errorMsg,
	}
	return Distribute(detail)
}

func Distribute(detail ErrorDetail) ErrorType {
	var dbIndex DBIndex

	if detail.ErrorType != "" {
		dbIndex = errorTypeMapping[detail.ErrorType]
	} else if detail.Error != nil {
		errorType := ClassifyError(detail.Error)
		detail.ErrorType = errorType
		dbIndex = errorTypeMapping[errorType]
	} else {
		dbIndex = DBGeneralWarning
	}

	storeToRedis(detail, dbIndex)
	return detail.ErrorType
}

func storeToRedis(detail ErrorDetail, dbIndex DBIndex) {
	client := GetRedisClient(dbIndex)
	if client == nil {
		return
	}

	ctx := context.Background()
	warningID := time.Now().UnixNano()
	timestamp := time.Now().Unix()

	var warningData map[string]interface{}

	switch dbIndex {
	case DBGeneralWarning:
		warning := Warning_General.Warning_General{
			WarningID:      warningID,
			WarningType:    string(detail.ErrorType),
			WarningContent: getErrorMessage(detail),
		}
		warningData = map[string]interface{}{
			"warning_id":      warning.WarningID,
			"warning_type":    warning.WarningType,
			"warning_content": warning.WarningContent,
		}

	case DBBackend:
		warning := Warning_Backend.Warning_Backend{
			WarningID:      warningID,
			WarningType:    string(detail.ErrorType),
			WarningContent: getErrorMessage(detail),
			Timestamp:      timestamp,
			Source:         detail.Source,
			Details:        detail.Details,
		}
		warningData = map[string]interface{}{
			"warning_id":      warning.WarningID,
			"warning_type":    warning.WarningType,
			"warning_content": warning.WarningContent,
			"timestamp":       warning.Timestamp,
			"source":          warning.Source,
			"details":         warning.Details,
		}

	case DBFrontend:
		warning := Warning_Frontend.Warning_Frontend{
			WarningID:      warningID,
			WarningType:    string(detail.ErrorType),
			WarningContent: getErrorMessage(detail),
			Timestamp:      timestamp,
			Source:         detail.Source,
			Component:      detail.Component,
			Details:        detail.Details,
		}
		warningData = map[string]interface{}{
			"warning_id":      warning.WarningID,
			"warning_type":    warning.WarningType,
			"warning_content": warning.WarningContent,
			"timestamp":       warning.Timestamp,
			"source":          warning.Source,
			"component":       warning.Component,
			"details":         warning.Details,
		}

	case DBAgent:
		warning := Warning_Agent.Warning_Agent{
			WarningID:      warningID,
			WarningType:    string(detail.ErrorType),
			WarningContent: getErrorMessage(detail),
			Timestamp:      timestamp,
			AgentID:        detail.AgentID,
			TaskID:         detail.TaskID,
			Details:        detail.Details,
		}
		warningData = map[string]interface{}{
			"warning_id":      warning.WarningID,
			"warning_type":    warning.WarningType,
			"warning_content": warning.WarningContent,
			"timestamp":       warning.Timestamp,
			"agent_id":        warning.AgentID,
			"task_id":         warning.TaskID,
			"details":         warning.Details,
		}

	case DBDrone:
		warning := Warning_Drone.Warning_Drone{
			WarningID:      warningID,
			WarningType:    string(detail.ErrorType),
			WarningContent: getErrorMessage(detail),
			Timestamp:      timestamp,
			DroneID:        detail.DroneID,
			HandlerID:      detail.HandlerID,
			Details:        detail.Details,
		}
		warningData = map[string]interface{}{
			"warning_id":      warning.WarningID,
			"warning_type":    warning.WarningType,
			"warning_content": warning.WarningContent,
			"timestamp":       warning.Timestamp,
			"drone_id":        warning.DroneID,
			"handler_id":      warning.HandlerID,
			"details":         warning.Details,
		}

	case DBSensor:
		warning := Warning_Sensor.Warning_Sensor{
			WarningID:      warningID,
			WarningType:    string(detail.ErrorType),
			WarningContent: getErrorMessage(detail),
			Timestamp:      timestamp,
			SensorID:       detail.SensorID,
			SensorType:     detail.SensorType,
			Value:          detail.Value,
			Threshold:      detail.Threshold,
			Details:        detail.Details,
		}
		warningData = map[string]interface{}{
			"warning_id":      warning.WarningID,
			"warning_type":    warning.WarningType,
			"warning_content": warning.WarningContent,
			"timestamp":       warning.Timestamp,
			"sensor_id":       warning.SensorID,
			"sensor_type":     warning.SensorType,
			"value":           warning.Value,
			"threshold":       warning.Threshold,
			"details":         warning.Details,
		}
	}

	jsonData, err := json.Marshal(warningData)
	if err != nil {
		return
	}

	key := fmt.Sprintf("warning:%d", warningID)
	client.Set(ctx, key, jsonData, 7*24*time.Hour)
	client.LPush(ctx, "warnings:list", key)
}

func getErrorMessage(detail ErrorDetail) string {
	if detail.Error != nil {
		return detail.Error.Error()
	}
	if detail.Details != "" {
		return detail.Details
	}
	return "unknown error"
}

func HandleBackendError(errorMsg string, source string, details string) {
	detail := ErrorDetail{
		ErrorType: ErrorTypeBackend,
		Source:    source,
		Details:   details,
		Error:     fmt.Errorf(errorMsg),
	}
	Distribute(detail)
}

func HandleDroneError(errorMsg string, droneID string, handlerID string, details string) {
	detail := ErrorDetail{
		ErrorType: ErrorTypeDrone,
		DroneID:   droneID,
		HandlerID: handlerID,
		Details:   details,
		Error:     fmt.Errorf(errorMsg),
	}
	Distribute(detail)
}

func HandleSensorError(errorMsg string, sensorID string, sensorType string, value string, threshold string, details string) {
	detail := ErrorDetail{
		ErrorType:  ErrorTypeSensor,
		SensorID:   sensorID,
		SensorType: sensorType,
		Value:      value,
		Threshold:  threshold,
		Details:    details,
		Error:      fmt.Errorf(errorMsg),
	}
	Distribute(detail)
}

func HandleAgentError(errorMsg string, agentID string, taskID string, details string) {
	detail := ErrorDetail{
		ErrorType: ErrorTypeAgent,
		AgentID:   agentID,
		TaskID:    taskID,
		Details:   details,
		Error:     fmt.Errorf(errorMsg),
	}
	Distribute(detail)
}

func HandleFrontendError(errorMsg string, source string, component string, details string) {
	detail := ErrorDetail{
		ErrorType: ErrorTypeFrontend,
		Source:    source,
		Component: component,
		Details:   details,
		Error:     fmt.Errorf(errorMsg),
	}
	Distribute(detail)
}
