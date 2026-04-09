package FRPHelper

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	Conf "MavlinkProject/Server/backend/Config"
	Board "MavlinkProject/Server/backend/Shared/Boards"
)

type CentralServerInfo struct {
	Name             string
	Address          string
	Port             int
	Timeout          time.Duration
	ReadTimeout      time.Duration
	MaxRetryAttempts int
}

func GetFRPCentrals() []CentralServerInfo {
	setting := Conf.GetSetting()
	cfgs := setting.Board.FRP.CentralServers
	frpCfg := setting.Board.FRP

	if len(cfgs) == 0 {
		return []CentralServerInfo{}
	}

	result := make([]CentralServerInfo, 0, len(cfgs))
	for _, cfg := range cfgs {
		result = append(result, CentralServerInfo{
			Name:             cfg.Name,
			Address:          cfg.Address,
			Port:             cfg.Port,
			Timeout:          time.Duration(frpCfg.Timeout) * time.Second,
			ReadTimeout:      time.Duration(frpCfg.ReadTimeout) * time.Second,
			MaxRetryAttempts: frpCfg.MaxRetryAttempts,
		})
	}
	return result
}

func PushMessageToCentral(frpAddress string, timeout time.Duration, maxRetryAttempts int, message *Board.BoardMessage) (*Board.BoardMessage, error) {
	// 总会重试一次 
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if maxRetryAttempts <= 0 {
		maxRetryAttempts = 1
	}

	var lastErr error

	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		conn, err := net.DialTimeout("tcp", frpAddress, timeout)
		if err != nil {
			lastErr = err
			continue
		}

		data, err := json.Marshal(message)
		if err != nil {
			conn.Close()
			lastErr = err
			continue
		}

		_, err = conn.Write(data)
		if err != nil {
			conn.Close()
			lastErr = err
			continue
		}

		conn.SetReadDeadline(time.Now().Add(timeout))
		buffer := make([]byte, 4096)
		n, err := conn.Read(buffer)
		conn.Close()
		if err != nil {
			lastErr = err
			continue
		}

		if n > 0 {
			var response Board.BoardMessage
			if err := json.Unmarshal(buffer[:n], &response); err == nil {
				return &response, nil
			}
		}
	}

	return nil, fmt.Errorf("PushMessageToCentral failed after %d attempts: %v", maxRetryAttempts, lastErr)
}
