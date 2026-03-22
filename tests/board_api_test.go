package tests

import (
    "bytes"
    "encoding/json"
    "fmt"
    "net"
    "net/http"
    "testing"
    "time"
)

// 发送UDP握手包以建立反向通信能力
func sendUDPHandshake(t *testing.T, address, port string) {
    serverAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%s", address, port))
    if err != nil {
        t.Fatalf("无法解析 UDP 地址: %v", err)
    }

    conn, err := net.DialUDP("udp", nil, serverAddr)
    if err != nil {
        t.Fatalf("无法连接 UDP 服务器: %v", err)
    }
    defer conn.Close()

    // 构造一个简单的 JSON 消息作为握手包
    msgPayload := map[string]interface{}{
        "message_id": "HANDSHAKE",
        "message_time": time.Now(),
        "from_id": "simulated_board",
        "from_type": "Drone",
    }
    msgBytes, _ := json.Marshal(msgPayload)

    _, err = conn.Write(msgBytes)
    if err != nil {
        t.Fatalf("无法发送 UDP 握手包: %v", err)
    }
    t.Log("已发送 UDP 握手包")
    time.Sleep(100 * time.Millisecond) // 等待服务器处理
}

func getAuthToken(t *testing.T) string {
    // 快速注册并登录获取 Token
    username := fmt.Sprintf("board_tester_%d", time.Now().Unix())
    email := fmt.Sprintf("%s@example.com", username)
    password := "TestPass@123"

    // 注册
    regBody, _ := json.Marshal(map[string]string{"username": username, "email": email, "password": password})
    http.Post(API_BASE_URL+"/users/register", "application/json", bytes.NewBuffer(regBody))

    // 登录
    loginBody, _ := json.Marshal(map[string]string{"email": email, "password": password})
    resp, err := http.Post(API_BASE_URL+"/users/login", "application/json", bytes.NewBuffer(loginBody))
    if err != nil {
        t.Fatalf("登录失败: %v", err)
    }
    defer resp.Body.Close()

    var res map[string]interface{}
    json.NewDecoder(resp.Body).Decode(&res)
    
    // 兼容不同的 Token 返回格式 (data.token 或 直接 token)
    if token, ok := res["token"].(string); ok {
        return token
    }
    if data, ok := res["data"].(map[string]interface{}); ok {
        if token, ok := data["token"].(string); ok {
            return token
        }
    }
    t.Fatal("无法获取 Token")
    return ""
}

func TestBoardLifecycle(t *testing.T) {
    token := getAuthToken(t)
    client := &http.Client{}
    boardID := fmt.Sprintf("test_board_%d", time.Now().Unix())

    // 1. 创建 Board Server (UDP)
    t.Log("步骤 1: 创建 Board 服务器(UDP)...")
    createPayload := map[string]string{
        "board_id":   boardID,
        "board_name": "UnitTestBoard",
        "board_type": "Drone",
        "connection": "UDP",
        "address":    "127.0.0.1",
        "port":       "14599", // 使用非常用端口避免冲突
    }
    body, _ := json.Marshal(createPayload)
    req, _ := http.NewRequest("POST", API_BASE_URL+"/api/board/create", bytes.NewBuffer(body))
    req.Header.Set("Authorization", "Bearer "+token)
    req.Header.Set("Content-Type", "application/json")
    
    resp, err := client.Do(req)
    if err != nil || resp.StatusCode != 200 {
        t.Fatalf("创建 Board 失败: %v, 状态码: %d", err, resp.StatusCode)
    }
    t.Log("Board 创建成功")

    // 2. 获取 Board 列表确认创建成功
    t.Log("步骤 2: 获取 Board 列表...")
    reqList, _ := http.NewRequest("GET", API_BASE_URL+"/api/board/list", nil)
    reqList.Header.Set("Authorization", "Bearer "+token)
    respList, err := client.Do(reqList)
    if err != nil {
        t.Fatalf("获取 Board 列表失败: %v", err)
    }
    // 检查状态码
    if respList.StatusCode != 200 {
        t.Errorf("获取 Board 列表返回非200状态码: %d", respList.StatusCode)
    }

    // 2. 发送 UDP 握手包 
    t.Log("Step 2a: 发送 UDP 握手包 (模拟板子连接)...")
    sendUDPHandshake(t, "127.0.0.1", "14599")

    // 3. 发送消息给 Board
    t.Log("步骤 3: 向 Board 发送消息...")
    msgPayload := map[string]interface{}{
        "to_id":   boardID,
        "command": "TEST_PING",
        "data":    map[string]string{"foo": "bar"},
    }
    msgBody, _ := json.Marshal(msgPayload)
    reqMsg, _ := http.NewRequest("POST", API_BASE_URL+"/api/board/send", bytes.NewBuffer(msgBody))
    reqMsg.Header.Set("Authorization", "Bearer "+token)
    reqMsg.Header.Set("Content-Type", "application/json")

    respMsg, err := client.Do(reqMsg)
    if err != nil || respMsg.StatusCode != 200 {
        t.Errorf("发送消息失败: %v, 状态码: %d", err, respMsg.StatusCode)
    } else {
        t.Log("消息发送成功")
    }
    
    // 5. 删除 Board Server
    t.Log("步骤 5: 删除 Board 服务器...")
    reqDel, _ := http.NewRequest("DELETE", API_BASE_URL+"/api/board/delete/"+boardID, nil)
    reqDel.Header.Set("Authorization", "Bearer "+token)
    
    respDel, err := client.Do(reqDel)
    if err != nil || respDel.StatusCode != 200 {
        t.Errorf("删除 Board 失败: %v, 状态码: %d", err, respDel.StatusCode)
    }
    t.Log("Board 删除成功")
}
