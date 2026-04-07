package FRPHelper

import (
	"encoding/json"
	"fmt"
	"log"
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

	timeout := time.Duration(frpCfg.Timeout) * time.Second
	readTimeout := time.Duration(frpCfg.ReadTimeout) * time.Second
	maxRetryAttempts := frpCfg.MaxRetryAttempts

	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if readTimeout <= 0 {
		readTimeout = 10 * time.Second
	}
	if maxRetryAttempts <= 0 {
		maxRetryAttempts = 3
	}

	servers := make([]CentralServerInfo, 0, len(cfgs))
	for _, cfg := range cfgs {
		servers = append(servers, CentralServerInfo{
			Name:             cfg.Name,
			Address:          cfg.Address,
			Port:             cfg.Port,
			Timeout:          timeout,
			ReadTimeout:      readTimeout,
			MaxRetryAttempts: maxRetryAttempts,
		})
	}

	return servers
}

func PushMessageToCentral(frpAddr string, timeout, readTimeout time.Duration, maxRetries int, msg *Board.BoardMessage) ([]byte, error) {
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		conn, err := net.DialTimeout("tcp", frpAddr, timeout)
		if err != nil {
			lastErr = fmt.Errorf("connection attempt %d failed: %v", attempt, err)
			log.Printf("[FRP] Connection attempt %d/%d failed: %v", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}
		defer conn.Close()

		data, err := json.Marshal(msg)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal message: %v", err)
		}

		_, err = conn.Write(data)
		if err != nil {
			lastErr = fmt.Errorf("send attempt %d failed: %v", attempt, err)
			log.Printf("[FRP] Send attempt %d/%d failed: %v", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		conn.SetReadDeadline(time.Now().Add(readTimeout))
		buffer := make([]byte, 4096)
		n, err := conn.Read(buffer)
		if err != nil {
			lastErr = fmt.Errorf("receive attempt %d failed: %v", attempt, err)
			log.Printf("[FRP] Receive attempt %d/%d failed: %v", attempt, maxRetries, err)
			time.Sleep(time.Second * 2)
			continue
		}

		log.Printf("[FRP] Message sent successfully to %s on attempt %d", frpAddr, attempt)
		return buffer[:n], nil
	}

	return nil, fmt.Errorf("failed to send message after %d attempts: %v", maxRetries, lastErr)
}
