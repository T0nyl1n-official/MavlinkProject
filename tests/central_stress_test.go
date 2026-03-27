package tests

import (
    "encoding/json"
    "fmt"
    "net"
    "os"
    "strconv"
    "sync"
    "sync/atomic"
    "testing"
    "time"
)

const (
    defaultStressWorkers  = 5
    defaultStressRounds   = 20
    defaultStressTimeout  = 20 * time.Second
    defaultReadBufferSize = 4096
)

// TestCentralStressLoad
// 目标: 压测 Central TCP 接入能力，不依赖无人机在线状态。
// 判定: 只要 Central 返回 JSON 且 status=received 即视为本次发送成功。
func TestCentralStressLoad(t *testing.T) {
    addr := getEnvOrDefault("CENTRAL_STRESS_ADDR", CentralServerAddress)
    workers := getEnvIntOrDefault("CENTRAL_STRESS_WORKERS", defaultStressWorkers)
    rounds := getEnvIntOrDefault("CENTRAL_STRESS_ROUNDS", defaultStressRounds)
    timeout := getEnvDurationMsOrDefault("CENTRAL_STRESS_TIMEOUT_MS", defaultStressTimeout)

    t.Logf("Central 压测开始 | addr=%s workers=%d rounds=%d timeout=%s", addr, workers, rounds, timeout)

    start := time.Now()
    total := int64(workers * rounds)
    var success int64
    var failed int64

    errCh := make(chan string, workers*rounds)
    var wg sync.WaitGroup
    wg.Add(workers)

    for w := 0; w < workers; w++ {
        workerID := w
        go func() {
            defer wg.Done()

            for i := 0; i < rounds; i++ {
                chain := buildStressChain(workerID, i)

                if err := sendChainToCentralWithAddr(addr, timeout, chain); err != nil {
                    atomic.AddInt64(&failed, 1)
                    errCh <- fmt.Sprintf("worker=%d round=%d chain=%s err=%v", workerID, i, chain.ChainID, err)
                    continue
                }

                atomic.AddInt64(&success, 1)
            }
        }()
    }

    wg.Wait()
    close(errCh)

    duration := time.Since(start)
    qps := float64(total) / duration.Seconds()

    t.Logf("Central 压测结束 | total=%d success=%d failed=%d duration=%s qps=%.2f",
        total, success, failed, duration, qps)

    // 打印前 20 条失败样本，避免日志刷爆
    printed := 0
    for e := range errCh {
        if printed >= 20 {
            break
        }
        t.Logf("失败样本: %s", e)
        printed++
    }

    if failed > 0 {
        t.Fatalf("压测失败: failed=%d success=%d total=%d", failed, success, total)
    }
}

func sendChainToCentralWithAddr(addr string, timeout time.Duration, chain ProgressChain) error {
    conn, err := net.DialTimeout("tcp", addr, timeout)
    if err != nil {
        return fmt.Errorf("dial失败: %w", err)
    }
    defer conn.Close()

    _ = conn.SetDeadline(time.Now().Add(timeout))

    boardMsg := BoardMessage{
        MessageID:   fmt.Sprintf("stress_msg_%d", time.Now().UnixNano()),
        MessageTime: time.Now(),
        FromID:      "stress_client",
        FromType:    "stress",
        ToID:        "central",
        ToType:      "server",
        Message: MessageData{
            MessageType: "Request",
            Attribute:   "Default",
            Connection:  "tcp",
            Command:     "schedule_chain",
            Data: map[string]interface{}{
                "progress_chain": chain,
            },
        },
    }

    reqBytes, err := json.Marshal(boardMsg)
    if err != nil {
        return fmt.Errorf("marshal请求失败: %w", err)
    }

    if _, err = conn.Write(reqBytes); err != nil {
        return fmt.Errorf("write失败: %w", err)
    }

    buf := make([]byte, defaultReadBufferSize)
    n, err := conn.Read(buf)
    if err != nil {
        return fmt.Errorf("read失败: %w", err)
    }

    var resp map[string]interface{}
    if err = json.Unmarshal(buf[:n], &resp); err != nil {
        return fmt.Errorf("响应非JSON: %w, raw=%s", err, string(buf[:n]))
    }

    status, _ := resp["status"].(string)
    if status != "received" {
        return fmt.Errorf("响应状态异常: status=%v resp=%v", resp["status"], resp)
    }

    return nil
}

func buildStressChain(workerID, round int) ProgressChain {
    chainID := fmt.Sprintf("stress_chain_w%03d_r%03d_%d", workerID, round, time.Now().UnixNano())

    return ProgressChain{
        ChainID: chainID,
        Tasks: []Task{
            {
                TaskID:  "task_0",
                Command: "TakeOff",
                Data: map[string]interface{}{
                    "altitude": 10.0,
                },
                Status: "pending",
            },
            {
                TaskID:  "task_1",
                Command: "Land",
                Data:    map[string]interface{}{},
                Status:  "pending",
            },
        },
        CurrentTask: 0,
        Status:      "pending",
    }
}

func getEnvOrDefault(key, def string) string {
    v := os.Getenv(key)
    if v == "" {
        return def
    }
    return v
}

func getEnvIntOrDefault(key string, def int) int {
    v := os.Getenv(key)
    if v == "" {
        return def
    }
    n, err := strconv.Atoi(v)
    if err != nil || n <= 0 {
        return def
    }
    return n
}

func getEnvDurationMsOrDefault(key string, def time.Duration) time.Duration {
    v := os.Getenv(key)
    if v == "" {
        return def
    }
    ms, err := strconv.Atoi(v)
    if err != nil || ms <= 0 {
        return def
    }
    return time.Duration(ms) * time.Millisecond
}