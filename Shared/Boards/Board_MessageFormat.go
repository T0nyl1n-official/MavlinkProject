package boards

import (
	"time"
)

type MessageAttribute string

const (
	MessageAttribute_Default MessageAttribute = "Default"
	MessageAttribute_Status  MessageAttribute = "Status"
	MessageAttribute_Mission MessageAttribute = "Mission"
	MessageAttribute_Control MessageAttribute = "Control"
	MessageAttribute_Command MessageAttribute = "Command"
	MessageAttribute_Warning MessageAttribute = "Warning"
)

type CommandType string

const (
	Command_TakeOff     CommandType = "TakeOff"
	Command_Land        CommandType = "Land"
	Command_GoTo        CommandType = "GoTo"
	Command_SetSpeed    CommandType = "SetSpeed"
	Command_SetPosition CommandType = "SetPosition"
	Command_TakePhoto   CommandType = "TakePhoto"
	Command_SetConfig   CommandType = "SetConfig"
	Command_SetCamera   CommandType = "SetCamera"

	Command_Connect    CommandType = "Connect"
	Command_Disconnect CommandType = "Disconnect"

	Command_GetConfig      CommandType = "GetConfig"
	Command_GetStatus      CommandType = "GetStatus"
	Command_DetailResponse CommandType = "Status"
)

const (
	Message_Request  string = "Request"
	Message_Response string = "Response"

	Connection_TCP    string = "TCP"
	Connection_UDP    string = "UDP"
	Connection_Serial string = "Serial"
)

// 针对于Message的封装式传递结构体
type BoardMessage struct {
	MessageID   string    `json:"message_id"`
	MessageTime time.Time `json:"message_time"`
	Message     Message   `json:"message"`

	FromID   string `json:"from_id"`
	FromType string `json:"from_type"`

	ToID   string `json:"to_id"`
	ToType string `json:"to_type"`
}

// Message 各元件之间传递信息的行为规范和数据格式
type Message struct {
	MessageType string           `json:"message_type"`
	Attribute   MessageAttribute `json:"message_attribute"`
	Connection  string           `json:"connection"`

	Command string                 `json:"command"`
	Data    map[string]interface{} `json:"data"`
}
/*
	Message 属性详解:
	MessageType: 消息类型，用于标识消息的类型，例如 "Request" 或 "Response"。
	Attribute: 消息属性，用于标识消息的属性，例如 "Default"、"Status"、"Mission" 等。
	Command: 命令，用于执行的操作的大致类型，例如 "TakeOff"、"Land" 等。
	Data: 数据，用于存储具体的操作数据，例如 "Position"、"Speed" 等。
*/
