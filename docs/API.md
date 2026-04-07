# MavlinkProject API 文档

## 基础信息

- **Base URL**: `http://localhost:8080`
- **认证方式**: `JWT Token (Bearer Token)`
- **Content-Type**: `application/json`

---

## 目录

1. [公共接口](#公共接口-无需认证)
2. [用户接口](#用户接口-需认证)
3. [管理员接口](#管理员接口-需-admin-角色)
4. [Board 通信接口](#board-板子通信接口-需认证)
5. [MAVLink V1 接口](#mavlink-v1-接口-需认证)
6. [MAVLink V2 接口](#mavlink-v2-接口-需认证)
7. [进度链接口](#进度链接口-需认证)
8. [消息处理流程](#消息处理流程)

---

## 公共接口 (无需认证)

### 1. 基础信息

#### GET /
获取服务基本信息

**响应示例**:
```json
{
  "status": "success",
  "message": "Hello world! - Welcome to The Mavlink Project!",
  "version": "Pre-Release 0.1.0"
}
```

---

### 2. 用户注册

#### POST /users/register
注册新用户

**请求体**:
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

---

### 3. 用户登录

#### POST /users/login
用户登录并获取 JWT Token

**请求体**:
```json
{
  "email": "test@example.com",
  "password": "password123"
}
```

**说明**: 登录成功后，请在后续请求的 Header 中添加:
```
Authorization: Bearer <token>
```

---

## 用户接口 (需认证)

### 4. 获取用户信息

#### GET /users/profile
获取当前登录用户信息

---

### 5. 更新用户信息

#### POST /users/update
更新用户信息

---

### 6. 删除用户

#### POST /users/delete
删除当前用户账户

---

### 7. 用户登出

#### POST /users/logout
用户登出 (使 Token 失效)

---

### 8. 发送邮箱验证码

#### POST /users/send-email-verification
发送邮箱验证码

---

## 管理员接口 (需 admin 角色)

### 9. 获取所有用户

#### GET /admin/all-profile
获取所有用户列表 (需 admin 权限)

---

## Board 板子通信接口 (需认证)

基础路径: `/api/board`

> ⚠️ 注意：Board API 使用 JWT 认证，用于前端/AI 发送指令给板子。
> 板子通过 TCP/UDP 直连后端，由 MessageDispatcher 自动路由消息。

### 23. 创建板子服务器

#### POST /api/board/create
创建并启动 Board 服务器

**请求体**:
```json
{
  "board_id": "drone_001",
  "board_name": "Drone Board",
  "board_type": "Drone",
  "connection": "TCP",
  "address": "0.0.0.0",
  "port": "14550"
}
```

---

### 24. 发送消息给板子

#### POST /api/board/send
后端发送消息给指定板子

**请求体**:
```json
{
  "to_id": "drone_001",
  "to_type": "Drone",
  "command": "TakePhoto",
  "attribute": "Command",
  "data": {
    "count": 10
  }
}
```

---

### 25. 获取板子列表

#### GET /api/board/list
获取所有已连接板子列表

---

### 26. 获取板子信息

#### GET /api/board/info/:boardID
获取指定板子详细信息

---

### 27. 停止板子服务器

#### POST /api/board/stop
停止 Board 服务器

---

## MAVLink V1 接口 (需认证)

基础路径: `/mavlink/v1`

### 10. Handler 管理

#### POST /mavlink/v1/handler/create
创建 MAVLink Handler

#### DELETE /mavlink/v1/handler/:id
删除 Handler

---

### 11. 连接管理

#### POST /mavlink/v1/connection/start
启动连接

#### POST /mavlink/v1/connection/stop
停止连接

---

### 12. 无人机控制

#### POST /mavlink/v1/drone/takeoff
无人机起飞

#### POST /mavlink/v1/drone/land
无人机降落

#### POST /mavlink/v1/drone/move
无人机移动

#### POST /mavlink/v1/drone/return
无人机返航

---

### 13. 状态监控

#### GET /mavlink/v1/drone/status
获取无人机状态

#### GET /mavlink/v1/drone/position
获取无人机位置

---

## MAVLink V2 接口 (需认证)

基础路径: `/mavlink/v2`

### 14. 高级控制

#### POST /mavlink/v2/takeoff
一键起飞

#### POST /mavlink/v2/land
一键降落

#### POST /mavlink/v2/move
移动到目标位置

---

### 15. 传感器警报

#### POST /mavlink/v2/sensor-alert
传感器警报响应 - 调度无人机前往指定位置拍照

**请求体**:
```json
{
  "sensor_id": "esp32_c3_001",
  "latitude": 37.7749,
  "longitude": -122.4194,
  "radius": 50,
  "photo_count": 10,
  "altitude": 100
}
```

**说明**: 此接口触发传感器警报处理流程:
1. 解析传感器位置信息
2. 生成任务链 (TakeOff → GoTo → TakePhoto → Land)
3. 通过 FRPHelper 发送到 Central 服务器执行

---

## 进度链接口 (需认证)

基础路径: `/api/chain`

### 16. 链管理

#### POST /api/chain/create
创建进度链

#### GET /api/chain/:id
获取进度链详情

#### GET /api/chain/list
列出所有进度链

---

### 17. 节点管理

#### POST /api/chain/:id/node/add
添加节点到链

#### DELETE /api/chain/:id/node/delete/:nodeId
从链中删除节点

---

### 18. 执行控制

#### POST /api/chain/:id/start
启动进度链执行

#### POST /api/chain/:id/stop
停止进度链

---

## 消息处理流程

### 整体架构

```
┌─────────────────────────────────────────────────────────────────┐
│                     设备消息入口                                  │
├─────────────────────────────────────────────────────────────────┤
│  Sensor/Board 设备                                                │
│       │                                                         │
│       ▼ (TCP/UDP)                                               │
│  BoardConn 后端监听 (0.0.0.0:8081 TCP / 0.0.0.0:8082 UDP)       │
│       │                                                         │
│       ▼                                                         │
│  isSensorMessage() 判断                                          │
│       │                                                         │
│       ├─── YES ──→ MessageDispatcher ──→ SensorAlertHandler     │
│       │                                    │                      │
│       │                                    ▼                      │
│       │                          GenerateChainAndSendToCentral   │
│       │                                    │                      │
│       │                                    ▼ (FRP)               │
│       │                              Central 服务器               │
│       │                                                         │
│       └─── NO ──→ BoardConn 内部处理 (messageChan)                │
│                       │                                          │
│                       ▼                                          │
│                BoardHandler (保留，暂不使用)                       │
└─────────────────────────────────────────────────────────────────┘
```

### 消息类型判断

| 消息类型 | FromType | Command/Attribute | Handler |
|----------|----------|-------------------|---------|
| 传感器警报 | sensor/esp32/alarm | Warning/SensorAlert/Alert | SensorAlertHandler |
| 飞控消息 | board/drone/fc | Heartbeat/Status/Control | BoardConn 内部处理 |

### SensorAlertHandler 处理流程

```
1. 接收 SensorAlert 消息
2. 解析 data 字段:
   - sensor_id: 传感器 ID
   - latitude/longitude: 目标位置
   - radius: 拍照半径
   - altitude: 飞行高度
   - photo_count: 拍照数量
3. 生成任务链:
   - TakeOff → GoTo → TakePhoto × N → Land
4. 调用 FRPHelper.PushMessageToCentral()
5. FRPHelper 遍历配置的 Central 服务器:
   - 每个 Central 重试 max_retry_attempts 次
   - 成功则返回，失败则尝试下一个
   - 全部失败才报错
```

### BoardMessage 结构

```json
{
  "message_id": "msg_1234567890",
  "message_time": "2026-04-07T12:00:00Z",
  "message": {
    "message_type": "Request",
    "attribute": "Warning",
    "connection": "TCP",
    "command": "SensorAlert",
    "data": {
      "sensor_id": "esp32_c3_001",
      "latitude": 37.7749,
      "longitude": -122.4194,
      "radius": 50,
      "altitude": 100,
      "photo_count": 10
    }
  },
  "from_id": "esp32_c3_001",
  "from_type": "sensor",
  "to_id": "server",
  "to_type": "server"
}
```

### FRP 多中心配置

```yaml
board:
  frp:
    timeout: 5
    read_timeout: 10
    max_retry_attempts: 3
    central_servers:
      - name: "central_1"
        address: "frp-put.com"
        port: 14465
      - name: "central_2"
        address: "frp-backup.com"
        port: 14466
```

**处理逻辑**:
1. 遍历所有配置的 Central 服务器
2. 每个 Central 最多重试 `max_retry_attempts` 次
3. 任一 Central 成功则返回
4. 全部失败才报错

### AIAgentHandler (预留接口)

```go
type AIAgentHandler struct {
    enabled       bool   // 是否启用
    analysisDepth string // 分析深度: "quick" / "full"
    RequireConfirm bool  // 执行前是否需要确认
}
```

**扩展点**:
- 启用后可拦截所有消息进行 AI 分析
- 可配置执行前需要人工/AI 确认
- 支持自定义分析深度

---

## 错误响应格式

所有错误响应遵循以下格式:

```json
{
  "success": false,
  "error": "错误信息",
  "code": 400
}
```

### 常见错误码

| 错误码 | 说明 |
|--------|------|
| 400 | 请求参数错误 |
| 401 | 未授权 (Token 无效或过期) |
| 403 | 权限不足 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |
| 503 | 服务不可用 (无可用 Central) |

---

## 注意事项

1. 所有受保护接口都需要在 Header 中携带 JWT Token
2. Token 默认存储在 Redis DB 13 中，登出后 Token 即失效
3. 管理员接口需要用户具有 `isadmin: true` 权限
4. MAVLink V1 接口为原子化操作，适合 AI 精细控制
5. MAVLink V2 接口为组合操作，适合快速任务执行
6. 进度链支持最多 1000 步任务执行
7. Board 消息通过 TCP/UDP 直连后端，由 MessageDispatcher 自动路由
8. Sensor 消息走 SensorAlertHandler 生成任务链
9. Board 消息暂由 BoardConn 内部处理 (BoardHandler 保留但未启用)
