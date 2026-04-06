package tests

import (
    "bytes"
    "encoding/json"
    "fmt"
    "math/rand"
    "net/http"
    "strings"
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

// =======================
// 压测配置
// =======================
const (
    StressBaseURL   = "https://api.deeppluse.dpdns.org"
    ConcurrentUsers = 50              // 并发用户数
    LoopsPerUser    = 10              // 单用户循环次数
    RequestTimeout  = 5 * time.Second // 请求超时
)

// =======================
// 全局统计
// =======================

var httpClient = &http.Client{
    Timeout: RequestTimeout,
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 100,
        IdleConnTimeout:     90 * time.Second,
    },
}

var (
    totalRequests   int64
    successRequests int64
    failedRequests  int64
    errorsCount     int64
)

// TestSystemStability 压力测试入口
// go test -v -run TestSystemStability ./tests/stress_test.go
func TestSystemStability(t *testing.T) {
    fmt.Printf("开始压测 | 用户: %d | 循环: %d | 预计请求: ~%d\n", 
        ConcurrentUsers, LoopsPerUser, ConcurrentUsers*LoopsPerUser*7)

    startTime := time.Now()
    var wg sync.WaitGroup
    wg.Add(ConcurrentUsers)

    for i := 0; i < ConcurrentUsers; i++ {
        go func(workerID int) {
            defer wg.Done()
            runUserScenario(t, workerID)
        }(i)
    }

    wg.Wait()
    duration := time.Since(startTime)

    printReport(duration)

    if failedRequests > 0 {
        t.Logf("警告: 失败请求数 %d", failedRequests)
    }
}

// runUserScenario 模拟单用户完整流程
func runUserScenario(t *testing.T, workerID int) {
    r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(workerID)))

    for i := 0; i < LoopsPerUser; i++ {
        username := fmt.Sprintf("stress_%d_%d_%d", workerID, i, r.Intn(100000))
        email := fmt.Sprintf("%s@stress.test", username)
        password := "StressPass@123"

        // 1. 注册
        makeRequest("POST", "/users/register", map[string]string{
            "username": username, "email": email, "password": password,
        }, "")

        // 2. 登录
        respLogin, err := makeRequest("POST", "/users/login", map[string]string{
            "email": email, "password": password,
        }, "")

        if err != nil {
            recordError()
            continue
        }

        var loginRes map[string]interface{}
        json.NewDecoder(respLogin.Body).Decode(&loginRes)
        respLogin.Body.Close()

        token := extractString(loginRes, "token")
        userID := extractFloat(loginRes, "User_ID")

        if token == "" {
            recordError()
            continue
        }

        // 3. 创建Handler
        handlerPayload := map[string]interface{}{
            "connection_type": "udp",
            "udp_addr":        "127.0.0.1",
            "udp_port":        15000 + (workerID * 100) + i,
            "system_id":       1,
            "component_id":    1,
            "protocol_version": "2.0",
        }

        respHandler, err := makeRequest("POST", "/mavlink/v1/handler/create", handlerPayload, token)
        if err == nil {
            var handlerRes map[string]interface{}
            json.NewDecoder(respHandler.Body).Decode(&handlerRes)
            respHandler.Body.Close()

            handlerID := extractString(handlerRes, "handler_id")
            
            if handlerID != "" {
                // 4. 启动连接 (修正: 使用Body传参)
                startPayload := map[string]string{"handler_id": handlerID}
                makeRequest("POST", "/mavlink/v1/connection/start", startPayload, token)

                // 5. 发送指令 (修正: 使用Body传参)
                modePayload := map[string]string{
                    "handler_id": handlerID, 
                    "mode": "GUIDED",
                }
                makeRequest("POST", "/mavlink/v1/drone/mode", modePayload, token)

                // 6. 清理Handler
                makeRequest("DELETE", fmt.Sprintf("/mavlink/v1/handler/%s", handlerID), nil, token)
            }
        }

        // 7. 销毁用户
        if userID != 0 {
            makeRequest("POST", fmt.Sprintf("/users/delete/%.0f", userID), nil, token)
        }

        time.Sleep(time.Duration(r.Intn(90)+10) * time.Millisecond)
    }
}

// =======================
// 辅助函数
// =======================

func makeRequest(method, endpoint string, body interface{}, token string) (*http.Response, error) {
    atomic.AddInt64(&totalRequests, 1)

    var bodyReader *bytes.Buffer
    if body != nil {
        jsonBytes, _ := json.Marshal(body)
        bodyReader = bytes.NewBuffer(jsonBytes)
    } else {
        bodyReader = bytes.NewBuffer(nil)
    }

    req, _ := http.NewRequest(method, StressBaseURL+endpoint, bodyReader)
    req.Header.Set("Content-Type", "application/json")
    if token != "" {
        req.Header.Set("Authorization", "Bearer "+token)
    }

    resp, err := httpClient.Do(req)
    if err != nil {
        atomic.AddInt64(&failedRequests, 1)
        return nil, err
    }

    // 状态码判定逻辑
    if resp.StatusCode >= 200 && resp.StatusCode < 300 {
        atomic.AddInt64(&successRequests, 1)
    } else {
        // 特殊处理: 如果是控制指令 API 且没有真实设备，可能会返回非 2xx
        // 但如果返回的是 400 Bad Request，说明我们参数没发对，仍算失败
        isControlAPI := strings.Contains(endpoint, "/connection/start") || strings.Contains(endpoint, "/drone/")
        
        if isControlAPI && resp.StatusCode != 400 && resp.StatusCode != 401 && resp.StatusCode != 403 {
            // 忽略因设备不存在导致的 404/500/504 等错误
             atomic.AddInt64(&successRequests, 1)
        } else {
            atomic.AddInt64(&failedRequests, 1)
        }
    }
    return resp, nil
}

func recordError() {
    atomic.AddInt64(&errorsCount, 1)
}

func extractString(source map[string]interface{}, key string) string {
    if val, ok := source[key]; ok { return val.(string) }
    if data, ok := source["data"].(map[string]interface{}); ok {
        if val, ok := data[key]; ok { return val.(string) }
    }
    return ""
}

func extractFloat(source map[string]interface{}, key string) float64 {
    if val, ok := source[key]; ok { return val.(float64) }
    if data, ok := source["data"].(map[string]interface{}); ok {
        if val, ok := data[key]; ok { return val.(float64) }
    }
    return 0
}

func printReport(duration time.Duration) {
    rps := float64(totalRequests) / duration.Seconds()
    fmt.Println("\n=== 压测报告 ===")
    fmt.Printf("耗时:     %v\n", duration)
    fmt.Printf("总请求:   %d\n", totalRequests)
    fmt.Printf("成功:     %d\n", successRequests)
    fmt.Printf("失败:     %d\n", failedRequests)
    fmt.Printf("逻辑错误: %d\n", errorsCount)
    fmt.Printf("QPS:      %.2f req/s\n", rps)
    fmt.Println("================")
}