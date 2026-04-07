package boardHandler

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	Board "MavlinkProject/Server/backend/Shared/Boards"
)

// PushMessageToCentral 会作为 TCP 客户端，通过 FRP 建立临时连接并向树莓派推送消息
func PushMessageToCentral(frpAddress string, message *Board.BoardMessage) (*Board.BoardMessage, error) {
	// 1. 作为客户端主动连接树莓派在FRP暴露出的地址
	conn, err := net.DialTimeout("tcp", frpAddress, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("连接中央服务器失败: %v", err)
	}
	defer conn.Close()

	// 2. 序列化数据
	data, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("序列化消息失败: %v", err)
	}

	// 3. 发送数据
	if _, err := conn.Write(data); err != nil {
		return nil, fmt.Errorf("发送消息失败: %v", err)
	}

	// 4. 等待树莓派回执响应
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buffer := make([]byte, 4096)
	n, err := conn.Read(buffer)
	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, fmt.Errorf("读取响应超时: %v", err)
		}
		return nil, fmt.Errorf("读取响应失败: %v", err)
	}

	var response Board.BoardMessage
	if err := json.Unmarshal(buffer[:n], &response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %v", err)
	}

	return &response, nil
}
