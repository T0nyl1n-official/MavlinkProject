# MavlinkProject API 文档

## 基础信息

- **Base URL**: `http://localhost:8080`
- **认证方式**: `JWT Token (Bearer Token)`
- **Content-Type**: `application/json`

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

**响应示例**:
```json
{
  "success": true,
  "User_ID": 1,
  "Username": "testuser",
  "Email": "test@example.com",
  "message": "用户注册成功"
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

**响应示例**:
```json
{
  "success": true,
  "User_ID": 1,
  "Username": "testuser",
  "Email": "test@example.com",
  "Role": "user",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "用户登录成功"
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

**Header**: `Authorization: Bearer <token>`

**响应示例**:
```json
{
  "success": true,
  "user": {...}
}
```

---

### 5. 更新用户信息

#### POST /users/update
更新用户信息

**Header**: `Authorization: Bearer <token>`

**请求体**:
```json
{
  "username": "newusername"
}
```

---

### 6. 删除用户

#### POST /users/delete
删除当前用户账户

**Header**: `Authorization: Bearer <token>`

---

### 7. 用户登出

#### POST /users/logout
用户登出 (使 Token 失效)

**Header**: `Authorization: Bearer <token>`

**响应示例**:
```json
{
  "success": true,
  "message": "用户已成功登出"
}
```

---

### 8. 发送邮箱验证码

#### POST /users/send-email-verification
发送邮箱验证码

**Header**: `Authorization: Bearer <token>`

**请求体**:
```json
{
  "email": "test@example.com",
  "type": "register"
}
```

**type 选项**: `register`, `login`, `reset_password`

---

## 管理员接口 (需 admin 角色)

### 9. 获取所有用户

#### GET /admin/all-profile
获取所有用户列表 (需 admin 权限)

**Header**: 
```
Authorization: Bearer <token>
```

**请求体**:
```json
{
  "isadmin": true
}
```

---

## MAVLink V1 接口 (需认证)

基础路径: `/mavlink/v1`

**Header**: `Authorization: Bearer <token>`

### 10. Handler 管理

#### POST /mavlink/v1/handler/create
创建 MAVLink Handler

**请求体**:
```json
{
  "connection_type": "udp",
  "udp_addr": "0.0.0.0",
  "udp_port": 14550,
  "system_id": 255,
  "component_id": 1,
  "protocol_version": "v2"
}
```

**connection_type 选项**: `udp`, `tcp`, `serial`

---

#### DELETE /mavlink/v1/handler/:id
删除 Handler

**参数**: `id` - Handler ID

---

#### GET /mavlink/v1/handler/:id
获取 Handler 状态

**参数**: `id` - Handler ID

**响应示例**:
```json
{
  "success": true,
  "handler_id": "12345",
  "connected": true,
  "status": "flying"
}
```

---

### 11. 连接管理

#### POST /mavlink/v1/connection/start
启动连接

**请求体**:
```json
{
  "handler_id": "12345"
}
```

---

#### POST /mavlink/v1/connection/stop
停止连接

**请求体**:
```json
{
  "handler_id": "12345"
}
```

---

#### POST /mavlink/v1/connection/restart
重启连接

**请求体**:
```json
{
  "handler_id": "12345"
}
```

---

### 12. 无人机控制

#### POST /mavlink/v1/drone/takeoff
无人机起飞

**请求体**:
```json
{
  "handler_id": "12345",
  "altitude": 10
}
```

---

#### POST /mavlink/v1/drone/land
无人机降落

**请求体**:
```json
{
  "handler_id": "12345"
}
```

---

#### POST /mavlink/v1/drone/move
无人机移动

**请求体**:
```json
{
  "handler_id": "12345",
  "lat": 40.7128,
  "lng": -74.0060,
  "alt": 10
}
```

---

#### POST /mavlink/v1/drone/return
无人机返航

**请求体**:
```json
{
  "handler_id": "12345"
}
```

---

#### POST /mavlink/v1/drone/mode
设置无人机模式

**请求体**:
```json
{
  "handler_id": "12345",
  "mode": "GUIDED"
}
```

---

### 13. 状态监控

#### GET /mavlink/v1/drone/status
获取无人机状态

**参数**: `handler_id`

**响应示例**:
```json
{
  "success": true,
  "handler_id": "12345",
  "status": "flying"
}
```

---

#### GET /mavlink/v1/drone/position
获取无人机位置

**参数**: `handler_id`

**响应示例**:
```json
{
  "success": true,
  "handler_id": "12345",
  "position": {
    "latitude": 40.7128,
    "longitude": -74.0060,
    "altitude": 10
  }
}
```

---

#### GET /mavlink/v1/drone/attitude
获取无人机姿态

**参数**: `handler_id`

---

#### GET /mavlink/v1/drone/battery
获取无人机电量

**参数**: `handler_id`

---

### 14. 地面站

#### POST /mavlink/v1/ground-station/set
设置地面站

**请求体**:
```json
{
  "handler_id": "12345",
  "system_id": 1,
  "component_id": 1
}
```

---

#### GET /mavlink/v1/ground-station
获取地面站信息

**参数**: `handler_id`

---

### 15. 高级功能

#### POST /mavlink/v1/stream/request
请求数据流

**请求体**:
```json
{
  "handler_id": "12345",
  "stream_id": "GPS_RAW_INT",
  "rate": 10
}
```

---

#### POST /mavlink/v1/heartbeat/send
发送心跳

**请求体**:
```json
{
  "handler_id": "12345"
}
```

---

## MAVLink V2 接口 (需认证)

基础路径: `/mavlink/v2`

**Header**: `Authorization: Bearer <token>`

### 16. 高级控制

#### POST /mavlink/v2/takeoff
一键起飞

**请求体**:
```json
{
  "handler_id": "12345",
  "altitude": 20
}
```

---

#### POST /mavlink/v2/land
一键降落

**请求体**:
```json
{
  "handler_id": "12345"
}
```

---

#### POST /mavlink/v2/move
移动到目标位置

**请求体**:
```json
{
  "handler_id": "12345",
  "lat": 40.7128,
  "lng": -74.0060,
  "alt": 15,
  "speed": 5.0
}
```

---

#### POST /mavlink/v2/return
一键返航

**请求体**:
```json
{
  "handler_id": "12345"
}
```

---

#### POST /mavlink/v2/mode
设置飞行模式

**请求体**:
```json
{
  "handler_id": "12345",
  "mode": "AUTO"
}
```

---

### 17. 状态查询

#### GET /mavlink/v2/status
获取无人机完整状态

**参数**: `handler_id`

---

#### GET /mavlink/v2/position
获取位置信息

**参数**: `handler_id`

---

#### GET /mavlink/v2/battery
获取电池信息

**参数**: `handler_id`

---

### 18. 地面站管理

#### POST /mavlink/v2/ground-station
配置地面站

**请求体**:
```json
{
  "handler_id": "12345",
  "config": {
    "system_id": 1,
    "component_id": 1
  }
}
```

---

### 19. 智能功能

#### POST /mavlink/v2/sensor-alert
传感器警报响应 - 调度无人机前往指定位置拍照

**请求体**:
```json
{
  "handler_id": "12345",
  "lat": 40.7128,
  "lng": -74.0060,
  "alt": 20,
  "photo_count": 10
}
```

---

#### POST /mavlink/v2/return-charge
无人机自动返回充电

**请求体**:
```json
{
  "handler_id": "12345"
}
```

---

## 进度链接口 (需认证)

基础路径: `/api/chain`

**Header**: `Authorization: Bearer <token>`

### 20. 链管理

#### POST /api/chain/create
创建进度链

**请求体**:
```json
{
  "name": "My Chain"
}
```

**响应示例**:
```json
{
  "success": true,
  "chain_id": "chain-123",
  "chain_name": "My Chain",
  "message": "Chain created successfully"
}
```

---

#### DELETE /api/chain/:id
删除进度链

**参数**: `id` - Chain ID

---

#### GET /api/chain/:id
获取进度链详情

**参数**: `id` - Chain ID

---

#### GET /api/chain/list
列出所有进度链

---

### 21. 节点管理

#### POST /api/chain/:id/node/add
添加节点到链

**请求体**:
```json
{
  "node_type": "takeoff",
  "handler_config": {
    "connection_type": "udp",
    "udp_port": 14550
  },
  "params": {
    "altitude": 10
  }
}
```

**node_type 选项**: `takeoff`, `land`, `move`, `return`, `wait`, `custom`

---

#### DELETE /api/chain/:id/node/delete/:nodeId
从链中删除节点

---

### 22. 执行控制

#### POST /api/chain/:id/start
启动进度链执行

---

#### POST /api/chain/:id/reset
重置进度链

---

#### POST /api/chain/:id/pause
暂停进度链

---

#### POST /api/chain/:id/stop
停止进度链

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

---

## 认证流程示例

### 1. 登录获取 Token
```bash
curl -X POST http://localhost:8080/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

### 2. 使用 Token 访问受保护接口
```bash
curl -X GET http://localhost:8080/mavlink/v1/drone/status?handler_id=12345 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

---

## 注意事项

1. 所有受保护接口都需要在 Header 中携带 JWT Token
2. Token 默认存储在 Redis DB 13 中，登出后 Token 即失效
3. 管理员接口需要用户具有 `isadmin: true` 权限
4. MAVLink V1 接口为原子化操作，适合精细控制
5. MAVLink V2 接口为组合操作，适合快速任务执行
6. 进度链支持最多 1000 步任务执行

---

## Board 板子通信接口 (需认证)

基础路径: `/api/board`

**Header**: `Authorization: Bearer <token>`

> ⚠️ 注意：Board API 需要 JWT 认证，用于前端/AI 发送指令给板子。

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

**connection 选项**: `TCP`, `UDP`

**响应示例**:
```json
{
  "success": true,
  "message": "Board server created successfully",
  "handler_id": "drone_001",
  "board": {
    "board_id": "drone_001",
    "board_name": "Drone Board"
  }
}
```

---

### 24. 启动板子服务器

#### POST /api/board/start
启动已创建的 Board 服务器

**请求体**:
```json
{
  "board_id": "drone_001",
  "connection": "TCP",
  "address": "0.0.0.0",
  "port": "14550"
}
```

---

### 25. 停止板子服务器

#### POST /api/board/stop
停止 Board 服务器

**请求体**:
```json
{
  "board_id": "drone_001"
}
```

---

### 26. 删除板子服务器

#### DELETE /api/board/delete/:boardID
删除 Board 服务器

**参数**: `boardID` - Board ID

---

### 27. 发送消息给板子

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

**command 选项**: `TakePhoto`, `TakeOff`, `Land`, `GoTo`, `SetSpeed`, `SetPosition`, `SetConfig`, `SetCamera`, `GetStatus`, `GetConfig`, `Connect`, `Disconnect`

**attribute 选项**: `Default`, `Status`, `Mission`, `Control`, `Command`, `Warning`

---

### 28. 板子间消息转发

#### POST /api/board/forward
手动转发消息从一个板子到另一个板子

**请求体**:
```json
{
  "from_board_id": "drone_001",
  "to_board_id": "drone_002",
  "command": "RelayCommand",
  "attribute": "Control",
  "data": {
    "message": "hello"
  }
}
```

**说明**: 
- `forwardMessage` 是**手动调用**的转发方式
- 消息从后端手动指定源板子和目标板子
- 适用于 AI/前端需要精确控制转发路径的场景

---

### 29. 启用自动转发

#### POST /api/board/auto-forward
启用消息自动转发功能

**说明**:
- 启用后，后端会自动将收到的消息根据 `ToID` 转发到对应板子
- 板子发送消息时只需设置 `ToID` 为目标板子 ID
- 适用于板子自主通信场景，减少后端干预

**响应示例**:
```json
{
  "success": true,
  "message": "Auto-forward enabled"
}
```

---

### 30. 获取板子列表

#### GET /api/board/list
获取所有已连接板子列表

**响应示例**:
```json
{
  "success": true,
  "boards": [
    {
      "board_id": "drone_001",
      "board_ip": "0.0.0.0",
      "board_port": "14550",
      "board_status": "TCP",
      "is_connected": true
    }
  ]
}
```

---

### 31. 获取板子信息

#### GET /api/board/info/:boardID
获取指定板子详细信息

**参数**: `boardID` - Board ID

**响应示例**:
```json
{
  "success": true,
  "board": {
    "board_id": "drone_001",
    "board_ip": "0.0.0.0",
    "board_port": "14550",
    "board_status": "TCP",
    "is_connected": true
  }
}
```

---

## 板子通信流程说明

### 通信模式

#### 1. 后端 → 板子
```
前端/AI → 后端 → /api/board/send → 板子
```

#### 2. 板子 → 后端
```
板子 → TCP/UDP 监听端口 → BoardListener → BoardClassifier → 处理/存储
```

#### 3. 板子 → 板子 (自动转发)
```
板子A → 后端 → 自动根据ToID转发 → 板子B
(需先调用 /api/board/auto-forward 启用)
```

#### 4. 板子 → 板子 (手动转发)
```
前端/AI → 后端 → /api/board/forward → 板子B
```

### BoardMessage 结构

```json
{
  "message_id": "20260321153045-abc123",
  "message_time": "2026-03-21T15:30:45Z",
  "message": {
    "message_type": "Request",
    "attribute": "Command",
    "connection": "TCP",
    "command": "TakePhoto",
    "data": {"count": 10}
  },
  "from_id": "drone_001",
  "from_type": "Drone",
  "to_id": "drone_002",
  "to_type": "Drone"
}
```

### 消息分类处理

BoardClassifier 会根据消息类型自动处理：

| 消息类型 | 处理动作 | 说明 |
|----------|----------|------|
| Heartbeat | Ignore | 正常心跳 |
| Status | Log | 状态更新记录 |
| Mission | DispatchNext/Reschedule | 任务完成/失败处理 |
| Control | Log | 控制指令记录 |
| Command | Response | 命令执行响应 |
| Warning/Error | ReportError | 错误上报 Agent |
