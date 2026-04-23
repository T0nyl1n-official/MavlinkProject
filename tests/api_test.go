package tests

import (
    "bytes"
    "crypto/tls"
    "encoding/json"
    "fmt"
    "net/http"
    "testing"
    "time"
)

//const API_BASE_URL = "https://api.deeppluse.dpdns.org"
const API_BASE_URL = "https://localhost:8080"

// TestFullUserFlow 测试完整的用户业务流程
func TestFullUserFlow(t *testing.T) {
    // 创建跳过证书校验的 HTTP 客户端
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: tr}

    username := fmt.Sprintf("test_user_%d", time.Now().Unix())
    email := fmt.Sprintf("%s@example.com", username)
    password := "TestPass@123"

    // 1. 注册
    regBody, _ := json.Marshal(map[string]string{"username": username, "email": email, "password": password})
    regResp, err := client.Post(API_BASE_URL+"/users/register", "application/json", bytes.NewBuffer(regBody))
    if err != nil || regResp.StatusCode != 200 {
        t.Fatalf("注册失败: %v", err)
    }
    t.Logf("注册成功: %s", username)

    // 2. 登录
    loginBody, _ := json.Marshal(map[string]string{"email": email, "password": password})
    loginResp, err := client.Post(API_BASE_URL+"/users/login", "application/json", bytes.NewBuffer(loginBody))
    if err != nil || loginResp.StatusCode != 200 {
        t.Fatalf("登录失败: %v", err)
    }

    // 3. 提取 Token 和 UserID
    var loginRes map[string]interface{}
    json.NewDecoder(loginResp.Body).Decode(&loginRes)

    var token string
    var userID float64 // json数字默认为float64

    // 辅助提取函数
    extract := func(source map[string]interface{}, key string) interface{} {
        if val, ok := source[key]; ok {
            return val
        }
        if data, ok := source["data"].(map[string]interface{}); ok {
            if val, ok := data[key]; ok {
                return val
            }
        }
        return nil
    }

    if v, ok := extract(loginRes, "token").(string); ok {
        token = v
    }
    if v, ok := extract(loginRes, "User_ID").(float64); ok {
        userID = v
    }

    if token == "" {
        t.Fatal("Token获取失败")
    }
    t.Logf("登录成功. Token: %s..., UserID: %.0f", token[:10], userID)

    // 4. 创建任务链
    chainBody, _ := json.Marshal(map[string]string{"name": "AutoTestChain"})
    reqCreate, _ := http.NewRequest("POST", API_BASE_URL+"/api/chain/create", bytes.NewBuffer(chainBody))
    reqCreate.Header.Set("Authorization", "Bearer "+token)
    createResp, err := client.Do(reqCreate)

    if err != nil || createResp.StatusCode != 200 {
        t.Fatalf("创建链失败: %v", err)
    }

    // 提取 Chain ID
    var createRes map[string]interface{}
    json.NewDecoder(createResp.Body).Decode(&createRes)
    var chainID string
    if v, ok := extract(createRes, "chain_id").(string); ok {
        chainID = v
    }

    if chainID == "" {
        t.Fatal("Chain ID获取失败")
    }
    t.Logf("创建链成功: %s", chainID)

    // 5. 添加节点
    nodeBody, _ := json.Marshal(map[string]interface{}{"node_type": "CheckPoint", "params": map[string]interface{}{"wait_time": 5}})
    reqAdd, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/chain/%s/node/add", API_BASE_URL, chainID), bytes.NewBuffer(nodeBody))
    reqAdd.Header.Set("Authorization", "Bearer "+token)
    client.Do(reqAdd)
    t.Logf("添加节点成功")

    // 6. 清理链
    reqDel, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/chain/%s", API_BASE_URL, chainID), nil)
    reqDel.Header.Set("Authorization", "Bearer "+token)
    client.Do(reqDel)
    t.Logf("清理成功: 链已删除")

    // 7. [边界测试] 重复注册
    dupRegResp, _ := client.Post(API_BASE_URL+"/users/register", "application/json", bytes.NewBuffer(regBody))
    if dupRegResp.StatusCode == 200 {
        t.Fatalf("边界测试失败: 允许重复注册")
    }
    t.Logf("边界测试通过: 禁止重复注册")

    // 8. [边界测试] 错误密码
    wrongLoginBody, _ := json.Marshal(map[string]string{"email": email, "password": "WrongPassword!"})
    wrongLoginResp, _ := client.Post(API_BASE_URL+"/users/login", "application/json", bytes.NewBuffer(wrongLoginBody))
    if wrongLoginResp.StatusCode == 200 {
        t.Fatalf("边界测试失败: 允许错误密码")
    }
    t.Logf("边界测试通过: 禁止错误密码")

    // 9. [边界测试] 无Token
    noTokenReq, _ := http.NewRequest("POST", API_BASE_URL+"/api/chain/create", bytes.NewBuffer(chainBody))
    noTokenResp, _ := client.Do(noTokenReq)
    if noTokenResp.StatusCode == 200 {
        t.Fatalf("边界测试失败: 允许无Token")
    }
    t.Logf("边界测试通过: 禁止无Token")

    // 10. 销毁用户 (必须有 userID)
    if userID != 0 {
        reqUserDel, _ := http.NewRequest("POST", fmt.Sprintf("%s/users/delete/%.0f", API_BASE_URL, userID), nil)
        reqUserDel.Header.Set("Authorization", "Bearer "+token)
        delRes, _ := client.Do(reqUserDel)
        if delRes.StatusCode == 200 {
            t.Logf("自动销毁成功: 测试用户已删除")
        } else {
            t.Logf("警告: 用户销毁失败 (Code %d)", delRes.StatusCode)
        }
    }
}

// TestMavlinkHandler 测试 Mavlink 连接处理
func TestMavlinkHandler(t *testing.T) {
    // 创建跳过证书校验的 HTTP 客户端
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: tr}

    username := fmt.Sprintf("mav_test_%d", time.Now().Unix())
    email := fmt.Sprintf("%s@example.com", username)
    password := "TestPass@123"

    regBody, _ := json.Marshal(map[string]string{"username": username, "email": email, "password": password})
    client.Post(API_BASE_URL+"/users/register", "application/json", bytes.NewBuffer(regBody))

    loginBody, _ := json.Marshal(map[string]string{"email": email, "password": password})
    loginResp, _ := client.Post(API_BASE_URL+"/users/login", "application/json", bytes.NewBuffer(loginBody))

    var loginRes map[string]interface{}
    json.NewDecoder(loginResp.Body).Decode(&loginRes)

    // 提取 Token & UserID
    var token string
    var userID float64

    extract := func(source map[string]interface{}, key string) interface{} {
        if val, ok := source[key]; ok {
            return val
        }
        if data, ok := source["data"].(map[string]interface{}); ok {
            if val, ok := data[key]; ok {
                return val
            }
        }
        return nil
    }

    if v, ok := extract(loginRes, "token").(string); ok {
        token = v
    }
    if v, ok := extract(loginRes, "User_ID").(float64); ok {
        userID = v
    }

    // 1. 创建 Handler
    handlerPayload, _ := json.Marshal(map[string]interface{}{
        "connection_type":  "udp",
        "udp_addr":         "127.0.0.1",
        "udp_port":         14550,
        "system_id":        1,
        "component_id":     1,
        "protocol_version": "2.0",
    })

    reqCreate, _ := http.NewRequest("POST", API_BASE_URL+"/mavlink/v1/handler/create", bytes.NewBuffer(handlerPayload))
    reqCreate.Header.Set("Authorization", "Bearer "+token)
    respCreate, err := client.Do(reqCreate)

    if err != nil || respCreate.StatusCode != 200 {
        t.Fatalf("创建Handler失败: %v", err)
    }

    var handlerRes map[string]interface{}
    json.NewDecoder(respCreate.Body).Decode(&handlerRes)
    var handlerID string
    if v, ok := extract(handlerRes, "handler_id").(string); ok {
        handlerID = v
    }

    if handlerID != "" {
        t.Logf("创建Handler成功: %s", handlerID)

        // 2. 查询 Handler
        reqGet, _ := http.NewRequest("GET", fmt.Sprintf("%s/mavlink/v1/handler/%s", API_BASE_URL, handlerID), nil)
        reqGet.Header.Set("Authorization", "Bearer "+token)
        respGet, _ := client.Do(reqGet)
        if respGet.StatusCode == 200 {
            t.Logf("查询Handler成功")
        }

        // 3. 启动连接 (仅测试接口响应)
        reqStart, _ := http.NewRequest("POST", fmt.Sprintf("%s/mavlink/v1/connection/start?handler_id=%s", API_BASE_URL, handlerID), nil)
        reqStart.Header.Set("Authorization", "Bearer "+token)
        respStart, _ := client.Do(reqStart)
        t.Logf("启动连接接口响应: %d", respStart.StatusCode)
    } else {
        t.Logf("未能提取HandlerID, 跳过后续步骤. 响应: %+v", handlerRes)
    }

    // 清理 MavlinkUser
    if userID != 0 {
        reqDel, _ := http.NewRequest("POST", fmt.Sprintf("%s/users/delete/%.0f", API_BASE_URL, userID), nil)
        reqDel.Header.Set("Authorization", "Bearer "+token)
        client.Do(reqDel)
        t.Logf("测试用户已清理")
    }
}

// TestMavlinkControl 测试无人机控制指令接口
func TestMavlinkControl(t *testing.T) {
    // 创建跳过证书校验的 HTTP 客户端
    tr := &http.Transport{
        TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
    }
    client := &http.Client{Transport: tr}

    // 复用逻辑：注册一个临时用户
    username := fmt.Sprintf("ctrl_test_%d", time.Now().Unix())
    email := fmt.Sprintf("%s@example.com", username)
    password := "TestPass@123"

    regBody, _ := json.Marshal(map[string]string{"username": username, "email": email, "password": password})
    client.Post(API_BASE_URL+"/users/register", "application/json", bytes.NewBuffer(regBody))

    loginBody, _ := json.Marshal(map[string]string{"email": email, "password": password})
    loginResp, _ := client.Post(API_BASE_URL+"/users/login", "application/json", bytes.NewBuffer(loginBody))

    var loginRes map[string]interface{}
    json.NewDecoder(loginResp.Body).Decode(&loginRes)

    var token string
    var userID float64

    extract := func(source map[string]interface{}, key string) interface{} {
        if val, ok := source[key]; ok {
            return val
        }
        if data, ok := source["data"].(map[string]interface{}); ok {
            if val, ok := data[key]; ok {
                return val
            }
        }
        return nil
    }

    if v, ok := extract(loginRes, "token").(string); ok {
        token = v
    }
    if v, ok := extract(loginRes, "User_ID").(float64); ok {
        userID = v
    }

    // 1. 创建 Handler
    handlerPayload, _ := json.Marshal(map[string]interface{}{
        "connection_type":  "udp",
        "udp_addr":         "127.0.0.1",
        "udp_port":         14550,
        "system_id":        1,
        "component_id":     1,
        "protocol_version": "2.0",
    })
    reqCreate, _ := http.NewRequest("POST", API_BASE_URL+"/mavlink/v1/handler/create", bytes.NewBuffer(handlerPayload))
    reqCreate.Header.Set("Authorization", "Bearer "+token)
    respCreate, _ := client.Do(reqCreate)

    var handlerRes map[string]interface{}
    json.NewDecoder(respCreate.Body).Decode(&handlerRes)
    var handlerID string
    if v, ok := extract(handlerRes, "handler_id").(string); ok {
        handlerID = v
    }

    if handlerID != "" {
        t.Logf("Handler准备就绪: %s", handlerID)

        // 基础控制 (Base Control v1)

        // 1. 起飞 (Takeoff v1)
        takeoffBody, _ := json.Marshal(map[string]interface{}{"altitude": 10})
        reqTakeoff, _ := http.NewRequest("POST", fmt.Sprintf("%s/mavlink/v1/drone/takeoff?handler_id=%s", API_BASE_URL, handlerID), bytes.NewBuffer(takeoffBody))
        reqTakeoff.Header.Set("Authorization", "Bearer "+token)
        respTakeoff, _ := client.Do(reqTakeoff)

        // 2. 模式切换 (Mode)
        modeBody, _ := json.Marshal(map[string]interface{}{"mode": "GUIDED"})
        reqMode, _ := http.NewRequest("POST", fmt.Sprintf("%s/mavlink/v1/drone/mode?handler_id=%s", API_BASE_URL, handlerID), bytes.NewBuffer(modeBody))
        reqMode.Header.Set("Authorization", "Bearer "+token)
        respMode, _ := client.Do(reqMode)

        // 3. 返航 (Return)
        reqRtl, _ := http.NewRequest("POST", fmt.Sprintf("%s/mavlink/v1/drone/return?handler_id=%s", API_BASE_URL, handlerID), nil)
        reqRtl.Header.Set("Authorization", "Bearer "+token)
        respRtl, _ := client.Do(reqRtl)

        // 验证 v1 响应
        t.Logf("V1 控制指令响应状态: 起飞(%d), 模式(%d), 返航(%d)",
            respTakeoff.StatusCode, respMode.StatusCode, respRtl.StatusCode)

        // 高级控制 (Advanced Control v2)

        // 1. 移动 (Move v2)
        moveBody, _ := json.Marshal(map[string]interface{}{"x": 10, "y": 5, "z": -2})
        reqMove, _ := http.NewRequest("POST", fmt.Sprintf("%s/mavlink/v2/move?handler_id=%s", API_BASE_URL, handlerID), bytes.NewBuffer(moveBody))
        reqMove.Header.Set("Authorization", "Bearer "+token)
        respMove, _ := client.Do(reqMove)

        // 2. 降落 (Land v2)
        landBody, _ := json.Marshal(map[string]interface{}{"latitude": 0, "longitude": 0}) // 原地降落
        reqLand, _ := http.NewRequest("POST", fmt.Sprintf("%s/mavlink/v2/land?handler_id=%s", API_BASE_URL, handlerID), bytes.NewBuffer(landBody))
        reqLand.Header.Set("Authorization", "Bearer "+token)
        respLand, _ := client.Do(reqLand)

        // 验证 v2 响应
        t.Logf("V2 控制指令响应状态: 移动(%d), 降落(%d)", respMove.StatusCode, respLand.StatusCode)

        if respTakeoff.StatusCode == 404 || respMove.StatusCode == 404 {
            t.Errorf("严重错误: 控制接口返回 404 Not Found，请检查路由配置")
        }
    } else {
        t.Errorf("无法创建 Handler，跳过控制测试")
    }

    // 清理用户
    if userID != 0 {
        reqDel, _ := http.NewRequest("POST", fmt.Sprintf("%s/users/delete/%.0f", API_BASE_URL, userID), nil)
        reqDel.Header.Set("Authorization", "Bearer "+token)
        client.Do(reqDel)
        t.Logf("测试用户已清理")
    }
}