package Models

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type ModelClient struct {
	lstmBaseURL  string
	yoloBaseURL  string
	httpClient   *http.Client
	maxRetry     int
	retryDelay   time.Duration
	mu           sync.RWMutex
	lstmEnabled  bool
	yoloEnabled  bool
}

var (
	globalClient *ModelClient
	clientOnce   sync.Once
)

func InitModelClient(lstmURL, yoloURL string) {
	clientOnce.Do(func() {
		globalClient = &ModelClient{
			lstmBaseURL: lstmURL,
			yoloBaseURL: yoloURL,
			httpClient: &http.Client{
				Timeout: 30 * time.Second,
			},
			maxRetry:    3,
			retryDelay:  500 * time.Millisecond,
			lstmEnabled: lstmURL != "",
			yoloEnabled: yoloURL != "",
		}
		log.Printf("[ModelClient] 初始化完成: LSTM=%s (enabled=%v), YOLO=%s (enabled=%v)",
			lstmURL, globalClient.lstmEnabled, yoloURL, globalClient.yoloEnabled)
	})
}

func GetModelClient() *ModelClient {
	if globalClient == nil {
		InitModelClient("", "")
	}
	return globalClient
}

func (mc *ModelClient) IsLSTMEnabled() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.lstmEnabled
}

func (mc *ModelClient) IsYOLOEnabled() bool {
	mc.mu.RLock()
	defer mc.mu.RUnlock()
	return mc.yoloEnabled
}

func (mc *ModelClient) SetLSTMURL(url string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.lstmBaseURL = url
	mc.lstmEnabled = url != ""
	log.Printf("[ModelClient] LSTM URL 更新: %s (enabled=%v)", url, mc.lstmEnabled)
}

func (mc *ModelClient) SetYOLOURL(url string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.yoloBaseURL = url
	mc.yoloEnabled = url != ""
	log.Printf("[ModelClient] YOLO URL 更新: %s (enabled=%v)", url, mc.yoloEnabled)
}

func (mc *ModelClient) AnalyzeSensorData(req LSTMRequest) (*LSTMResponse, error) {
	if !mc.IsLSTMEnabled() {
		return &LSTMResponse{
			IsAnomaly:   false,
			AnomalyType: AnomalyUnknown,
			Confidence:  0,
		}, nil
	}

	url := mc.lstmBaseURL + "/predict"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("LSTM request marshal failed: %v", err)
	}

	var resp *http.Response
	for attempt := 0; attempt < mc.maxRetry; attempt++ {
		httpReq, _ := http.NewRequest("POST", url, bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err = mc.httpClient.Do(httpReq)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if attempt < mc.maxRetry-1 {
			time.Sleep(mc.retryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("LSTM API call failed after %d retries: %v", mc.maxRetry, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("LSTM API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var lstmResp LSTMResponse
	if err := json.NewDecoder(resp.Body).Decode(&lstmResp); err != nil {
		return nil, fmt.Errorf("LSTM response decode failed: %v", err)
	}

	log.Printf("[ModelClient] LSTM 分析完成: sensor=%s, anomaly=%v, score=%.4f, type=%s",
		req.SensorID, lstmResp.IsAnomaly, lstmResp.AnomalyScore, lstmResp.AnomalyType)

	return &lstmResp, nil
}

func (mc *ModelClient) AnalyzeImage(req YOLORequest) (*YOLOResponse, error) {
	if !mc.IsYOLOEnabled() {
		return &YOLOResponse{
			HasAnomaly:  false,
			AnomalyType: AnomalyUnknown,
			Severity:    SeverityInfo,
		}, nil
	}

	url := mc.yoloBaseURL + "/detect"

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("YOLO request marshal failed: %v", err)
	}

	var resp *http.Response
	for attempt := 0; attempt < mc.maxRetry; attempt++ {
		httpReq, _ := http.NewRequest("POST", url, bytes.NewReader(body))
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err = mc.httpClient.Do(httpReq)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if attempt < mc.maxRetry-1 {
			time.Sleep(mc.retryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("YOLO API call failed after %d retries: %v", mc.maxRetry, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YOLO API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var yoloResp YOLOResponse
	if err := json.NewDecoder(resp.Body).Decode(&yoloResp); err != nil {
		return nil, fmt.Errorf("YOLO response decode failed: %v", err)
	}

	log.Printf("[ModelClient] YOLO 分析完成: source=%s, anomaly=%v, type=%s, severity=%s, detections=%d",
		req.Source, yoloResp.HasAnomaly, yoloResp.AnomalyType, yoloResp.Severity, len(yoloResp.Detections))

	return &yoloResp, nil
}

func (mc *ModelClient) AnalyzeImageFile(filePath string, source string, metadata map[string]string) (*YOLOResponse, error) {
	if !mc.IsYOLOEnabled() {
		return &YOLOResponse{
			HasAnomaly:  false,
			AnomalyType: AnomalyUnknown,
			Severity:    SeverityInfo,
		}, nil
	}

	url := mc.yoloBaseURL + "/detect/file"

	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file failed: %v", err)
	}
	defer file.Close()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("image", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("create form file failed: %v", err)
	}

	if _, err := io.Copy(part, file); err != nil {
		return nil, fmt.Errorf("copy file data failed: %v", err)
	}

	writer.WriteField("source", source)
	if metadata != nil {
		for k, v := range metadata {
			writer.WriteField("meta_"+k, v)
		}
	}

	writer.Close()

	var resp *http.Response
	for attempt := 0; attempt < mc.maxRetry; attempt++ {
		httpReq, _ := http.NewRequest("POST", url, &buf)
		httpReq.Header.Set("Content-Type", writer.FormDataContentType())

		resp, err = mc.httpClient.Do(httpReq)
		if err == nil && resp.StatusCode == 200 {
			break
		}
		if resp != nil {
			resp.Body.Close()
		}
		if attempt < mc.maxRetry-1 {
			time.Sleep(mc.retryDelay)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("YOLO file API call failed after %d retries: %v", mc.maxRetry, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("YOLO file API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var yoloResp YOLOResponse
	if err := json.NewDecoder(resp.Body).Decode(&yoloResp); err != nil {
		return nil, fmt.Errorf("YOLO response decode failed: %v", err)
	}

	log.Printf("[ModelClient] YOLO 文件分析完成: file=%s, anomaly=%v, type=%s, severity=%s",
		filepath.Base(filePath), yoloResp.HasAnomaly, yoloResp.AnomalyType, yoloResp.Severity)

	return &yoloResp, nil
}

func (mc *ModelClient) HealthCheck() map[string]interface{} {
	status := map[string]interface{}{
		"lstm_enabled": mc.IsLSTMEnabled(),
		"yolo_enabled": mc.IsYOLOEnabled(),
	}

	if mc.IsLSTMEnabled() {
		url := mc.lstmBaseURL + "/health"
		resp, err := mc.httpClient.Get(url)
		if err != nil || (resp != nil && resp.StatusCode != 200) {
			status["lstm_status"] = "unhealthy"
		} else {
			status["lstm_status"] = "healthy"
		}
		if resp != nil {
			resp.Body.Close()
		}
	} else {
		status["lstm_status"] = "disabled"
	}

	if mc.IsYOLOEnabled() {
		url := mc.yoloBaseURL + "/health"
		resp, err := mc.httpClient.Get(url)
		if err != nil || (resp != nil && resp.StatusCode != 200) {
			status["yolo_status"] = "unhealthy"
		} else {
			status["yolo_status"] = "healthy"
		}
		if resp != nil {
			resp.Body.Close()
		}
	} else {
		status["yolo_status"] = "disabled"
	}

	return status
}
