package FRPHelper

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	Conf "MavlinkProject/Server/backend/Config"
	Board "MavlinkProject/Server/backend/Shared/Boards"
)

type CentralHTTPClient struct {
	client           *http.Client
	maxRetryAttempts int
	centralURL       string
}

var centralClient *CentralHTTPClient

func InitCentralClient() {
	setting := Conf.GetSetting()
	cfgs := setting.Board.FRP.CentralServers
	if len(cfgs) == 0 {
		centralClient = NewCentralHTTPClient("https://central.deeppluse.dpdns.org:8084/central/message", 10*time.Second, 3)
		return
	}
	cfg := cfgs[0]
	centralURL := fmt.Sprintf("https://%s/central/message", cfg.Address)
	frpCfg := setting.Board.FRP
	centralClient = NewCentralHTTPClient(centralURL, time.Duration(frpCfg.Timeout)*time.Second, frpCfg.MaxRetryAttempts)
	log.Printf("[FRPHelper] CentralBoard client initialized: %s", centralURL)
}

func GetCentralClient() *CentralHTTPClient {
	return centralClient
}

type DroneInfo struct {
	BoardID      string  `json:"board_id"`
	BatteryLevel float64 `json:"battery_level"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Altitude     float64 `json:"altitude"`
	IsIdle       bool    `json:"is_idle"`
}

func (c *CentralHTTPClient) GetAvailableDrones() ([]DroneInfo, error) {
	if c == nil || c.centralURL == "" {
		return nil, fmt.Errorf("central client not initialized")
	}

	statusURL := fmt.Sprintf("https://%s/api/drones/available", strings.TrimPrefix(c.centralURL, "https://"))
	statusURL = strings.TrimSuffix(statusURL, "/central/message")

	req, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Code   int         `json:"code"`
		Drones []DroneInfo `json:"drones"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return []DroneInfo{}, nil
	}

	return result.Drones, nil
}

type CentralServerHTTPInfo struct {
	Name             string
	URL              string
	Timeout          time.Duration
	MaxRetryAttempts int
}

func GetCentralHTTPClients() []CentralHTTPClient {
	setting := Conf.GetSetting()
	cfgs := setting.Board.FRP.CentralServers
	frpCfg := setting.Board.FRP

	clients := make([]CentralHTTPClient, 0, len(cfgs))
	for _, cfg := range cfgs {
		url := fmt.Sprintf("https://%s/central/message", cfg.Address)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		}

		clients = append(clients, CentralHTTPClient{
			client: &http.Client{
				Transport: tr,
				Timeout:   time.Duration(frpCfg.Timeout) * time.Second,
			},
			maxRetryAttempts: frpCfg.MaxRetryAttempts,
			centralURL:       url,
		})
	}
	return clients
}

func NewCentralHTTPClient(centralURL string, timeout time.Duration, maxRetryAttempts int) *CentralHTTPClient {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	if maxRetryAttempts <= 0 {
		maxRetryAttempts = 1
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &CentralHTTPClient{
		client: &http.Client{
			Transport: tr,
			Timeout:   timeout,
		},
		maxRetryAttempts: maxRetryAttempts,
		centralURL:       centralURL,
	}
}

func (c *CentralHTTPClient) SendMessage(message *Board.BoardMessage) (*Board.BoardMessage, error) {
	if c.centralURL == "" {
		return nil, fmt.Errorf("central URL is empty")
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}

	var lastErr error

	for attempt := 1; attempt <= c.maxRetryAttempts; attempt++ {
		req, err := http.NewRequest("POST", c.centralURL, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "MavlinkBackend/0.2.0")

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %v", err)
			continue
		}

		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %v", err)
			resp.Body.Close()
			continue
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			lastErr = fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
			continue
		}

		var response Board.BoardMessage
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %v", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("SendMessageToCentral failed after %d attempts: %v", c.maxRetryAttempts, lastErr)
}

func SendMessageToCentralHTTP(centralAddress string, message *Board.BoardMessage) (*Board.BoardMessage, error) {
	centralURL := fmt.Sprintf("https://%s/central/message", centralAddress)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   10 * time.Second,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}

	req, err := http.NewRequest("POST", centralURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "MavlinkBackend/0.2.0")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send message: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("central returned status %d: %s", resp.StatusCode, string(body))
	}

	var response Board.BoardMessage
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &response, nil
}

func PushMessageToCentralHTTP(centralAddress string, timeout time.Duration, maxRetryAttempts int, message *Board.BoardMessage) (*Board.BoardMessage, error) {
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	if maxRetryAttempts <= 0 {
		maxRetryAttempts = 1
	}

	centralURL := fmt.Sprintf("https://%s/central/message", centralAddress)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	client := &http.Client{
		Transport: tr,
		Timeout:   timeout,
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal message: %v", err)
	}

	var lastErr error

	for attempt := 1; attempt <= maxRetryAttempts; attempt++ {
		req, err := http.NewRequest("POST", centralURL, bytes.NewBuffer(jsonData))
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %v", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "MavlinkBackend/1.0")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("request failed: %v", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response: %v", err)
			continue
		}

		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
			lastErr = fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
			continue
		}

		var response Board.BoardMessage
		if err := json.Unmarshal(body, &response); err != nil {
			lastErr = fmt.Errorf("failed to unmarshal response: %v", err)
			continue
		}

		return &response, nil
	}

	return nil, fmt.Errorf("PushMessageToCentralHTTP failed after %d attempts: %v", maxRetryAttempts, lastErr)
}

func GetCentralHTTPAddresses() []string {
	setting := Conf.GetSetting()
	cfgs := setting.Board.FRP.CentralServers

	addresses := make([]string, 0, len(cfgs))
	for _, cfg := range cfgs {
		addresses = append(addresses, cfg.Address)
	}
	return addresses
}
