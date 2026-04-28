# MavlinkProject API 文档

---

## 一、基础信息

### 1.1 服务地址

| 环境 | 地址 | 说明 |
|------|------|------|
| 本地开发 | `http://localhost:8080` | HTTP 服务 |
| 本地开发 | `https://localhost:8080` | HTTPS 服务 |
| 生产环境 | `https://api.deeppluse.dpdns.org` | Cloudflare Tunnel |

### 1.2 认证方式

- **认证方式**: `JWT Token (Bearer Token)`
- **Header 格式**: `Authorization: Bearer <token>`
- **Token 存储**: Redis DB 13
- **Token 失效**: 用户登出后 Token 即失效

### 1.3 内容类型

- **Content-Type**: `application/json`
- **字符编码**: `UTF-8`

### 1.4 连接配置

| 服务 | 协议 | 地址 | 端口 |
|------|------|------|------|
| Board 监听 | TCP | 0.0.0.0 | 8081 |
| Board 监听 | UDP | 0.0.0.0 | 8082 |
| Backend 服务 | HTTPS | 0.0.0.0 | 8080 |
| Central 服务 | TCP | 0.0.0.0 | 8081 |

---

## 二、类型定义

### 2.1 通用类型

#### 2.1.1 分页参数

```go
type Pagination struct {
    Page    int `json:"page"`    // 页码，从 1 开始，默认 1
    Limit   int `json:"limit"`   // 每页数量，默认 10，最大 100
}
```

#### 2.1.2 时间范围

```go
type TimeRange struct {
    StartTime int64 `json:"start_time"` // 起始时间戳 (Unix)
    EndTime   int64 `json:"end_time"`   // 结束时间戳 (Unix)
}
```

#### 2.1.3 坐标

```go
type Coordinate struct {
    Latitude  float64 `json:"latitude"`   // 纬度 (-90 ~ 90)
    Longitude float64 `json:"longitude"` // 经度 (-180 ~ 180)
    Altitude  float64 `json:"altitude"`  // 高度 (米)
}
```

### 2.2 用户相关类型

#### 2.2.1 用户注册请求

```go
type RegisterRequest struct {
    Username string `json:"username"` // 用户名 (3-20字符)
    Email    string `json:"email"`    // 邮箱 (唯一)
    Password string `json:"password"` // 密码 (6-32字符)
}
```

#### 2.2.2 用户登录请求

```go
type LoginRequest struct {
    Email    string `json:"email"`    // 邮箱
    Password string `json:"password"` // 密码
}
```

#### 2.2.3 用户信息响应

```go
type UserProfile struct {
    UserID    string `json:"user_id"`
    Username  string `json:"username"`
    Email     string `json:"email"`
    IsAdmin   bool   `json:"is_admin"`
    IsBlocked bool   `json:"is_blocked"`
    CreatedAt int64  `json:"created_at"`
}
```

### 2.3 Board 相关类型

#### 2.3.1 Board 类型枚举

```go
type BoardType string

const (
    BoardTypeDrone   BoardType = "Drone"   // 无人机
    BoardTypeSensor  BoardType = "Sensor"  // 传感器
    BoardTypeCentral BoardType = "Central" // 中央服务器
)
```

#### 2.3.2 连接方式枚举

```go
type ConnectionType string

const (
    ConnectionTCP ConnectionType = "TCP"
    ConnectionUDP ConnectionType = "UDP"
    ConnectionHTTPS ConnectionType = "HTTPS"
)
```

#### 2.3.3 Board 信息

```go
type BoardInfo struct {
    BoardID     string         `json:"board_id"`
    BoardName   string         `json:"board_name"`
    BoardType   BoardType      `json:"board_type"`
    Connection  ConnectionType `json:"connection"`
    Address     string         `json:"address"`
    Port        string         `json:"port"`
    IsConnected bool          `json:"is_connected"`
    LastSeen    int64         `json:"last_seen"`
}
```

#### 2.3.4 Board 消息

```go
type BoardMessage struct {
    MessageID    string      `json:"message_id"`
    MessageTime int64       `json:"message_time"`
    Message     MessageData `json:"message"`
    FromID      string      `json:"from_id"`
    FromType    string      `json:"from_type"`
    ToID        string      `json:"to_id"`
    ToType      string      `json:"to_type"`
}

type MessageData struct {
    MessageType string                 `json:"message_type"` // Request/Response/Error
    Attribute   string                 `json:"attribute"`    // Warning/Info/Command
    Connection  string                 `json:"connection"`   // TCP/UDP
    Command     string                 `json:"command"`      // 具体命令
    Data        map[string]interface{} `json:"data"`         // 命令数据
}

#### 2.3.5 MessageAttribute 消息属性枚举

| Attribute 值 | 说明 | 使用场景 |
|-------------|------|---------|
| `Default` | 默认 | 通用消息传递 |
| `Status` | 状态 | 设备状态上报、心跳 |
| `Mission` | 任务 | 任务链相关消息 |
| `Control` | 控制 | 飞行控制指令 |
| `Command` | 命令 | 通用命令下发 |
| `Warning` | 警告 | 告警信息推送 |

#### 2.3.6 CommandType 命令类型枚举

| Command 值 | 说明 | 典型 Data 参数 |
|-----------|------|---------------|
| `TakeOff` | 起飞 | `{ "altitude": float }` |
| `Land` | 降落 | `{ "latitude": float, "longitude": float }` |
| `GoTo` | 飞往目标 | `{ "latitude": float, "longitude": float, "altitude": float }` |
| `SetSpeed` | 设置速度 | `{ "speed": float, "unit": "m/s" }` |
| `SetPosition` | 设置位置 | `{ "latitude": float, "longitude": float, "altitude": float }` |
| `TakePhoto` | 拍照 | `{ "latitude": float, "longitude": float }` |
| `SetConfig` | 设置配置 | `{ "key": "string", "value": any }` |
| `SetCamera` | 设置相机 | `{ "mode": "photo/video", "params": object }` |
| `Connect` | 连接 | `{ "target_id": "string" }` |
| `Disconnect` | 断开 | `{ "target_id": "string" }` |
| `GetConfig` | 获取配置 | `{ "key": "string" }` |
| `GetStatus` | 获取状态 | `{}` |
| `Status` | 状态响应 | 设备返回的状态数据 |
| `AutoReturn` | 自动返航 | `{ "home_latitude": float, "home_longitude": float }` |
| `StartRecord` | 开始录制 | `{ "camera_id": "string", "quality": "high/medium/low" }` |
| `StopRecord` | 停止录制 | `{ "camera_id": "string" }` |
| `Orbit` | 围绕某点环游 | `{ "center_lat": float, "center_lng": float, "altitude": float, "radius": float, "speed": float }` |
| `FourDirectionPhoto` | 四方位拍照 | `{ "latitude": float, "longitude": float, "altitude": float }` |
| `FourDirectionRecord` | 四方位录制 | `{ "latitude": float, "longitude": float, "altitude": float, "duration": int }` |
| `SetRPM` | 调整转速 | `{ "rpm": int, "motor_id": int (可选，默认所有) }` |

#### 2.3.7 MessageType 消息类型枚举

| MessageType 值 | 说明 |
|---------------|------|
| `Request` | 请求消息 |
| `Response` | 响应消息 |

#### 2.3.8 Connection 连接类型枚举

| Connection 值 | 说明 |
|---------------|------|
| `TCP` | TCP 连接 |
| `UDP` | UDP 连接 |
| `Serial` | 串口连接 |

### 2.4 传感器警报类型 (重点)

#### 2.4.1 传感器警报请求

```go
type SensorAlertRequest struct {
    SensorID   string  `json:"sensor_id"`    // 传感器ID (必填)
    SensorIP   string  `json:"sensor_ip"`    // 传感器IP地址
    SensorName string  `json:"sensor_name"`   // 传感器名称 (可选，未填则用SensorIP)
    AlertType  string  `json:"alert_type"`   // 警报类型 (必填)
    AlertMsg   string  `json:"alert_msg"`    // 预警消息
    Latitude   float64 `json:"latitude"`     // GPS纬度 (必填)
    Longitude  float64 `json:"longitude"`    // GPS经度 (必填)
    Timestamp  int64   `json:"timestamp"`    // 时间戳，默认当前时间
    Severity   string  `json:"severity"`     // 严重程度
}
```

#### 2.4.2 警报类型枚举

| AlertType | 说明 | 任务链 |
|------------|------|--------|
| `fire` | 火灾 | 起飞 → 飞往火源 → 区域侦察 |
| `Fire` / `FIRE` | 火灾 (大小写兼容) | 同上 |
| `rescue` | 搜救 | 起飞 → 飞往目标 → 网格搜索 → 降落 |
| `rescue` / `RESCUE` | 搜救 (大小写兼容) | 同上 |
| `missing` | 走失 | 同 rescue |
| `patrol` | 巡逻 | 起飞 → 飞往巡逻点 → 盘旋巡逻 → 返航 |
| `flood` | 洪灾 | 起飞 → 飞往洪区 → 洪情侦察 |
| 其他值 | 默认 | 起飞 → 飞往目标 → 返航 |

#### 2.4.3 严重程度枚举

```go
type SeverityLevel string

const (
    SeverityNone     SeverityLevel = "none"
    SeverityLow      SeverityLevel = "low"
    SeverityMedium   SeverityLevel = "medium"
    SeverityHigh     SeverityLevel = "high"
)
```

#### 2.4.4 无人机状态

```go
type DroneStatus struct {
    BoardID      string  `json:"board_id"`
    SystemID     uint8   `json:"system_id"`
    ComponentID  uint8   `json:"component_id"`
    IsIdle       bool    `json:"is_idle"`
    BatteryLevel float64 `json:"battery_level"`
    Latitude     float64 `json:"latitude"`
    Longitude    float64 `json:"longitude"`
    Altitude     float64 `json:"altitude"`
    LastUpdate   int64   `json:"last_update"`
}
```

### 2.5 任务链类型

#### 2.5.1 进度链

```go
type ProgressChain struct {
    ChainID       string     `json:"chain_id"`
    Tasks         []Task     `json:"tasks"`
    CurrentTask   int        `json:"current_task"`
    Status        ChainStatus `json:"status"`
    StartTime     int64      `json:"start_time"`
    EndTime       int64      `json:"end_time"`
    AssignedDrone string     `json:"assigned_drone"`
}
```

#### 2.5.2 任务

```go
type Task struct {
    TaskID     string                 `json:"task_id"`
    Command    string                 `json:"command"`
    Data       map[string]interface{} `json:"data"`
    Status     TaskStatus             `json:"status"`
    RetryCount int                    `json:"retry_count"`
    MaxRetries int                    `json:"max_retries"`
    Timeout    int                    `json:"timeout"` // 秒
    StartTime  int64                  `json:"start_time"`
    EndTime    int64                  `json:"end_time"`
}
```

#### 2.5.3 状态枚举

```go
// 任务链状态
type ChainStatus string

const (
    ChainStatusPending   ChainStatus = "pending"    // 待执行
    ChainStatusRunning   ChainStatus = "running"    // 执行中
    ChainStatusPaused    ChainStatus = "paused"     // 暂停
    ChainStatusCompleted ChainStatus = "completed"  // 已完成
    ChainStatusFailed    ChainStatus = "failed"     // 失败
)

// 任务状态
type TaskStatus string

const (
    TaskStatusPending   TaskStatus = "pending"
    TaskStatusRunning   TaskStatus = "running"
    TaskStatusCompleted TaskStatus = "completed"
    TaskStatusFailed    TaskStatus = "failed"
)
```

#### 2.5.4 任务命令枚举

| Command | 说明 | 必需参数 |
|---------|------|----------|
| `takeoff` | 起飞 | altitude (高度) |
| `land` | 降落 | latitude, longitude |
| `goto` / `goto_location` | 飞往目标 | latitude, longitude, altitude |
| `return_to_home` / `rtl` | 返航 | - |
| `survey` | 区域侦察 | latitude, longitude, radius, duration |
| `survey_grid` | 网格搜索 | latitude, longitude, width, height, altitude |
| `orbit` | 盘旋巡逻 | latitude, longitude, radius, duration |
| `take_photo` | 拍照 | - |
| `start_video` | 开始录像 | - |
| `stop_video` | 停止录像 | - |
| `set_mode` | 设置模式 | mode |

### 2.6 MAVLink 类型

#### 2.6.1 飞行模式

```go
type FlightMode string

const (
    FlightModeStabilize FlightMode = "STABILIZE"
    FlightModeAltHold   FlightMode = "ALT_HOLD"
    FlightModeLoiter    FlightMode = "LOITER"
    FlightModeRTL       FlightMode = "RTL"
    FlightModeAuto      FlightMode = "AUTO"
    FlightModeGuided    FlightMode = "GUIDED"
)
```

### 2.7 设备相关类型

#### 2.7.1 设备类型枚举

```go
type DeviceType string

const (
    DeviceTypeCentral  DeviceType = "central"  // 中央控制板
    DeviceTypeLandNode DeviceType = "landnode" // 节点板
    DeviceTypeSensor   DeviceType = "sensor"   // 传感器板
    DeviceTypeDrone    DeviceType = "drone"    // 无人机
)
```

#### 2.7.2 设备登录请求

```go
type DeviceLoginRequest struct {
    DeviceID   string `json:"device_id"`   // 设备ID (必填)
    DeviceKey  string `json:"device_key"`   // 设备密钥 (必填)
    DeviceType string `json:"device_type"` // 设备类型 (必填)
}
```

#### 2.7.3 设备信息响应

```go
type DeviceInfo struct {
    Device_ID   uint      `json:"Device_ID"`   // 数据库ID
    DeviceID    string    `json:"DeviceID"`    // 设备唯一ID
    DeviceName  string    `json:"DeviceName"`  // 设备名称
    DeviceType  string    `json:"DeviceType"`  // 设备类型
    IsOnline    bool      `json:"IsOnline"`    // 是否在线
    LastSeen    time.Time `json:"LastSeen"`    // 最后活跃时间
    Token       string    `json:"Token"`       // JWT Token
    ExpireTime  int       `json:"ExpireTime"`   // 有效期(秒)，默认86400
}
```

### 2.8 Terminal 终端类型

#### 2.8.1 Terminal 命令结构

```go
type TerminalCMD struct {
    CMD     string   // 主命令
    Objects []string // 对象列表
    Args    map[string]interface{} // 参数
}
```

#### 2.8.2 Terminal 响应结构

```go
type TerminalResponse struct {
    Success bool        // 是否成功
    Message interface{} // 响应消息
}
```

#### 2.8.3 Terminal 命令列表

| 命令 | 说明 | 权限 |
|------|------|------|
| `help` | 显示帮助信息 | 需要认证 |
| `whoami` | 显示当前用户 | 需要认证 |
| `ls` | 列出目录 | 需要认证 |
| `cd` | 切换目录 | 需要认证 |
| `pwd` | 显示当前目录 | 需要认证 |
| `clear` | 清屏 | 需要认证 |
| `server` | 服务器命令 | 需要认证 |
| `backend` | 后端命令 | 需要认证 |
| `database` | 数据库命令 | 需要认证 |
| `cache` | 缓存命令 | 需要认证 |
| `adduser` | 添加用户 | admin |
| `deluser` | 删除用户 | admin |
| `chmod` | 修改用户权限 | admin |
| `reboot` | 重启服务器 | admin |
| `shutdown` | 关闭服务器 | admin |

---

## 三、状态码规范

### 3.1 HTTP 状态码

| 状态码 | 说明 | 使用场景 |
|--------|------|----------|
| 200 | OK | 请求成功 |
| 201 | Created | 资源创建成功 |
| 202 | Accepted | 请求已接收，但未成功处理 (如无可用无人机) |
| 400 | Bad Request | 请求参数错误 |
| 401 | Unauthorized | 未授权 (Token 无效或过期) |
| 403 | Forbidden | 权限不足 |
| 404 | Not Found | 资源不存在 |
| 409 | Conflict | 资源冲突 |
| 429 | Too Many Requests | 请求过于频繁 |
| 500 | Internal Server Error | 服务器内部错误 |
| 503 | Service Unavailable | 服务不可用 |

### 3.2 业务错误码 (code)

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 通用错误 |
| 1001 | 用户不存在 |
| 1002 | 用户已存在 |
| 1003 | 密码错误 |
| 1004 | Token 无效 |
| 1005 | Token 已过期 |
| 2001 | 无人机不存在 |
| 2002 | 无人机不在线 |
| 2003 | 无人机忙碌中 |
| 2004 | 电量不足 |
| 3001 | 任务链不存在 |
| 3002 | 任务链执行失败 |
| 4001 | Board 不存在 |
| 4002 | Board 连接失败 |
| 5001 | 参数验证失败 |
| 5002 | 必填参数为空 |

### 3.3 任务链状态码

| 状态码 | 说明 |
|--------|------|
| pending | 待执行 |
| running | 执行中 |
| paused | 暂停 |
| completed | 已完成 |
| failed | 失败 |

---

## 四、目录

1. [公共接口 (无需认证)](#公共接口-无需认证)
   - [1.1 获取服务信息](#11-获取服务信息)
   - [1.2 用户注册](#12-用户注册)
   - [1.3 用户登录](#13-用户登录)
2. [设备认证接口 (无需认证)](#设备认证接口-无需认证)
   - [2.1 设备登录](#21-设备登录)
   - [2.2 设备登出](#22-设备登出)
3. [用户接口 (需认证)](#用户接口-需认证)
   - [3.1 获取用户信息](#31-获取用户信息)
   - [3.2 更新用户信息](#32-更新用户信息)
   - [3.3 删除用户账户](#33-删除用户账户)
   - [3.4 用户登出](#34-用户登出)
   - [3.5 发送邮箱验证码](#35-发送邮箱验证码)
4. [管理员接口 (需 admin 角色)](#管理员接口-需-admin-角色)
   - [4.1 获取所有用户](#41-获取所有用户)
5. [Board 板子通信接口 (需认证)](#board-板子通信接口-需认证)
   - [5.1 创建 Board 服务器](#51-创建-board-服务器)
   - [5.2 发送消息给 Board](#52-发送消息给-board)
   - [5.3 Board 发送消息 (设备认证)](#53-board-发送消息-设备认证)
   - [5.4 获取 Board 列表](#54-获取-board-列表)
   - [5.5 获取 Board 信息](#55-获取-board-信息)
   - [5.6 停止 Board 服务器](#56-停止-board-服务器)
6. [Terminal 终端接口 (需认证)](#terminal-终端接口-需认证)
   - [6.1 发送终端命令](#61-发送终端命令)
7. [MAVLink V1 接口 (需认证)](#mavlink-v1-接口-需认证)
   - [7.1 Handler 管理](#71-handler-管理)
   - [7.2 连接管理](#72-连接管理)
   - [7.3 无人机控制](#73-无人机控制)
   - [7.4 状态监控](#74-状态监控)
8. [MAVLink V2 接口 (需认证)](#mavlink-v2-接口-需认证)
   - [8.1 高级控制](#81-高级控制)
9. [传感器接口 (无需认证)](#传感器接口-无需认证)
   - [9.1 POST /api/sensor/message 传感器警报](#91-post-apisensormessage-传感器警报)
   - [9.2 GET /api/sensor/status 获取无人机状态](#92-get-apisensorstatus-获取无人机状态)
10. [视频流接口 (需认证)](#视频流接口-需认证)
    - [10.1 视频流类型定义](#101-视频流类型定义)
    - [10.2 Central 上传接口](#102-central-上传接口)
    - [10.3 前端获取接口](#103-前端获取接口)
    - [10.4 视频流架构说明](#104-视频流架构说明)
    - [10.5 前端集成示例](#105-前端集成示例)
    - [10.6 视频流错误码](#106-视频流错误码)
11. [进度链接口 (需认证)](#进度链接口-需认证)
    - [11.1 链管理](#111-链管理)
    - [11.2 节点管理](#112-节点管理)
    - [11.3 执行控制](#113-执行控制)
12. [消息处理流程](#消息处理流程)
13. [错误响应格式](#错误响应格式)
14. [注意事项](#注意事项)

---

## 五、接口详细说明

### 公共接口 (无需认证)

#### 1.1 获取服务信息

##### GET /

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

#### 1.2 用户注册

##### POST /users/register

注册新用户

**请求体**:
```json
{
  "username": "testuser",
  "email": "test@example.com",
  "password": "password123"
}
```

**参数说明**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| username | string | 是 | 用户名，3-20字符 |
| email | string | 是 | 邮箱地址，唯一 |
| password | string | 是 | 密码，6-32字符 |

**成功响应 (201)**:
```json
{
  "code": 0,
  "message": "User registered successfully",
  "user_id": "user_xxx"
}
```

**错误响应**:
```json
{
  "code": 1002,
  "message": "User already exists",
  "error": "email already registered"
}
```

---

#### 1.3 用户登录

##### POST /users/login

用户登录并获取 JWT Token

**请求体**:
```json
{
  "email": "test@example.com",
  "password": "password123"
}
```

**参数说明**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| email | string | 是 | 邮箱地址 |
| password | string | 是 | 密码 |

**成功响应 (200)**:
```json
{
  "code": 0,
  "message": "Login successful",
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "expire_time": 3600
}
```

**错误响应**:
```json
{
  "code": 1003,
  "message": "Invalid credentials"
}
```

**使用说明**: 登录成功后，请在后续请求的 Header 中添加:
```
Authorization: Bearer <token>
```

---

### 设备认证接口 (无需认证)

> ⚠️ 以下接口用于硬件设备（Central/LandNode/Sensor/Drone）认证，无需用户JWT Token
> ⚠️ 设备使用独立的认证系统，错误提示更友好，不会触发封禁

#### 2.1 设备登录

##### POST /device/login

硬件设备登录并获取 JWT Token

**请求体**:
```json
{
  "device_id": "central_001",
  "device_key": "your_device_key",
  "device_type": "central"
}
```

**参数说明**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| device_id | string | 是 | 设备唯一ID |
| device_key | string | 是 | 设备密钥 |
| device_type | string | 是 | 设备类型 (central/landnode/sensor/drone) |

**成功响应 (200)**:
```json
{
  "code": 0,
  "Device_ID": 1,
  "DeviceID": "central_001",
  "DeviceName": "Central Control",
  "DeviceType": "central",
  "Token": "eyJhbGciOiJIUzI1NiIs...",
  "ExpireTime": 86400
}
```

**错误响应 (401)** - 设备不存在或密钥错误:
```json
{
  "code": 1,
  "message": "设备不存在或密钥错误"
}
```

**使用说明**: 设备登录成功后，请在后续请求的 Header 中添加:
```
Authorization: Bearer <token>
X-Device-ID: <device_id>
X-Device-Type: <device_type>
```

---

#### 2.2 设备登出

##### POST /device/logout

设备登出，使 Token 失效

**请求头**:
```
Authorization: Bearer <token>
X-Device-ID: <device_id>
X-Device-Type: <device_type>
```

**成功响应 (200)**:
```json
{
  "code": 0,
  "message": "Logout successful"
}
```

---

### 用户接口 (需认证)

> ⚠️ 以下接口需要在 Header 中携带有效的 JWT Token

#### 3.1 获取用户信息

##### GET /users/profile

获取当前登录用户信息

**成功响应 (200)**:
```json
{
  "code": 0,
  "data": {
    "user_id": "user_xxx",
    "username": "testuser",
    "email": "test@example.com",
    "is_admin": false,
    "is_blocked": false,
    "created_at": 1712563200
  }
}
```

---

#### 2.2 更新用户信息

##### POST /users/update

更新用户信息

**请求体**:
```json
{
  "username": "newusername",
  "email": "newemail@example.com"
}
```

**成功响应 (200)**:
```json
{
  "code": 0,
  "message": "User updated successfully"
}
```

---

#### 2.3 删除用户账户

##### POST /users/delete

删除当前用户账户

**成功响应 (200)**:
```json
{
  "code": 0,
  "message": "User deleted successfully"
}
```

---

#### 2.4 用户登出

##### POST /users/logout

用户登出，使 Token 失效

**成功响应 (200)**:
```json
{
  "code": 0,
  "message": "Logout successful"
}
```

---

#### 2.5 发送邮箱验证码

##### POST /users/send-email-verification

发送邮箱验证码

**请求体**:
```json
{
  "email": "test@example.com"
}
```

**成功响应 (200)**:
```json
{
  "code": 0,
  "message": "Verification code sent"
}
```

---

### 管理员接口 (需 admin 角色)

> ⚠️ 以下接口需要用户具有 `is_admin: true` 权限

#### 3.1 获取所有用户

##### GET /admin/all-profile

获取所有用户列表

**请求参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| page | int | 否 | 页码，默认 1 |
| limit | int | 否 | 每页数量，默认 10 |

**成功响应 (200)**:
```json
{
  "code": 0,
  "data": [
    {
      "user_id": "user_xxx",
      "username": "testuser",
      "email": "test@example.com",
      "is_admin": true,
      "is_blocked": false
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 100
  }
}
```

---

### Board 板子通信接口 (需认证)

> ⚠️ 以下接口需要在 Header 中携带有效的 JWT Token
> ⚠️ 注意：Board API 用于前端/AI 发送指令给板子，板子通过 TCP/UDP 直连后端

基础路径: `/api/board`

#### 4.1 创建 Board 服务器

##### POST /api/board/create

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

**参数说明**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| board_id | string | 是 | Board ID，唯一标识 |
| board_name | string | 否 | Board 名称 |
| board_type | string | 是 | Board 类型 (Drone/Sensor/Central) |
| connection | string | 是 | 连接方式 (TCP/UDP) |
| address | string | 是 | 监听地址 |
| port | string | 是 | 监听端口 |

**成功响应 (201)**:
```json
{
  "code": 0,
  "message": "Board server created",
  "board_id": "drone_001"
}
```

---

#### 4.2 发送消息给 Board

##### POST /api/board/send

后端发送消息给指定 Board

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

**参数说明**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| to_id | string | 是 | 目标 Board ID |
| to_type | string | 是 | 目标 Board 类型 |
| command | string | 是 | 命令 |
| attribute | string | 是 | 属性 (Command/Query/Setting) |
| data | object | 否 | 命令数据 |

**成功响应 (200)**:
```json
{
  "code": 0,
  "message": "Message sent",
  "message_id": "msg_xxx"
}
```

---

#### 4.3 Board 发送消息 (设备认证)

##### POST /api/board/send-message

Board 设备发送消息到后端（需要设备认证）

**请求头**:
```
Authorization: Bearer <device_token>
X-Device-ID: <device_id>
X-Device-Type: <device_type>
User-Agent: AIAgent (非AIAgent才允许)
```

**请求体**:
```json
{
  "message_id": "msg_xxx",
  "message_time": 1713177600,
  "message": {
    "message_type": "Request",
    "attribute": "Warning",
    "connection": "TCP",
    "command": "StatusReport",
    "data": {
      "battery": 85.5,
      "latitude": 22.543123,
      "longitude": 114.052345
    }
  },
  "from_id": "central_001",
  "from_type": "Central",
  "to_id": "backend",
  "to_type": "Server"
}
```

**参数说明**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| message_id | string | 是 | 消息唯一ID |
| message_time | int64 | 是 | 消息时间戳 |
| message.message_type | string | 是 | 消息类型 (Request/Response/Error) |
| message.attribute | string | 是 | 属性 (Warning/Info/Command) |
| message.connection | string | 是 | 连接方式 (TCP/UDP) |
| message.command | string | 是 | 具体命令 |
| message.data | object | 否 | 命令数据 |
| from_id | string | 是 | 发送者ID |
| from_type | string | 是 | 发送者类型 |
| to_id | string | 是 | 接收者ID |
| to_type | string | 是 | 接收者类型 |

**成功响应 (200)**:
```json
{
  "success": true,
  "message": "Message received"
}
```

**错误响应 (401)** - 设备未认证:
```json
{
  "code": 1,
  "error": "Device authentication required",
  "message": "请先进行设备登录认证，获取设备令牌"
}
```

**错误响应 (401)** - 令牌失效:
```json
{
  "code": 1,
  "error": "Device token expired or invalid",
  "message": "设备令牌已失效，请重新登录获取令牌"
}
```

---

#### 4.5 获取 Board 列表

##### GET /api/board/list

获取所有已连接 Board 列表

**成功响应 (200)**:
```json
{
  "code": 0,
  "data": [
    {
      "board_id": "esp32_001",
      "board_name": "ESP32 Sensor",
      "board_type": "Sensor",
      "connection": "TCP",
      "address": "192.168.1.100",
      "port": "8081",
      "is_connected": true,
      "last_seen": 1713177600
    }
  ]
}
```

---

#### 4.4 获取 Board 信息

##### GET /api/board/info/:boardID

获取指定 Board 详细信息

**路径参数**:

| 参数 | 类型 | 说明 |
|------|------|------|
| boardID | string | Board ID |

**成功响应 (200)**:
```json
{
  "code": 0,
  "data": {
    "board_id": "esp32_001",
    "board_name": "ESP32 Sensor",
    "board_type": "Sensor",
    "is_connected": true,
    "last_seen": 1713177600
  }
}
```

---

#### 4.6 停止 Board 服务器

##### POST /api/board/stop

停止 Board 服务器

**请求体**:
```json
{
  "board_id": "drone_001"
}
```

**成功响应 (200)**:
```json
{
  "code": 0,
  "message": "Board server stopped"
}
```

---

### Terminal 终端接口 (需认证)

> ⚠️ 以下接口需要在 Header 中携带有效的 JWT Token
> ⚠️ Terminal 接口用于远程管理服务器，支持多种命令

基础路径: `/terminal`

#### 6.1 发送终端命令

##### POST /terminal/message

发送终端命令并获取响应

**请求体**:
```json
{
  "cmd": "help",
  "objects": [],
  "args": {}
}
```

**参数说明**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| cmd | string | 是 | 主命令 |
| objects | []string | 否 | 对象列表 |
| args | object | 否 | 参数字典 |

**Terminal 命令示例**:

```json
{
  "cmd": "server",
  "objects": ["config"],
  "args": {}
}
```

```json
{
  "cmd": "backend",
  "objects": ["status"],
  "args": {}
}
```

```json
{
  "cmd": "database",
  "objects": ["config"],
  "args": {}
}
```

```json
{
  "cmd": "cache",
  "objects": ["config"],
  "args": {}
}
```

```json
{
  "cmd": "log",
  "objects": ["server"],
  "args": {"type": "error", "level": 10}
}
```

```json
{
  "cmd": "adduser",
  "objects": ["newuser", "password123"],
  "args": {}
}
```

```json
{
  "cmd": "deluser",
  "objects": ["username"],
  "args": {}
}
```

```json
{
  "cmd": "chmod",
  "objects": ["username", "admin"],
  "args": {}
}
```

**成功响应 (200)**:
```json
{
  "success": true,
  "message": {
    "command": "server",
    "sub_command": "config",
    "result": {
      "mode": "release",
      "server_id": "mavlink_server_001",
      "version": "Pre-Release 0.1.0",
      "backend": {
        "host": "0.0.0.0",
        "port": "8080",
        "cors": true,
        "rate_lim": 100,
        "logger": true,
        "board": true
      }
    }
  }
}
```

**错误响应 (401)** - 未授权:
```json
{
  "success": false,
  "message": "User not authenticated"
}
```

---

### MAVLink V1 接口 (需认证)

> ⚠️ 以下接口需要在 Header 中携带有效的 JWT Token
> ⚠️ MAVLink V1 接口为原子化操作，适合 AI 精细控制

基础路径: `/mavlink/v1`

#### 7.1 Handler 管理

##### POST /mavlink/v1/handler/create

创建 MAVLink Handler

##### DELETE /mavlink/v1/handler/:id

删除 Handler

---

#### 7.2 连接管理

##### POST /mavlink/v1/connection/start

启动连接

##### POST /mavlink/v1/connection/stop

停止连接

---

#### 7.3 无人机控制

##### POST /mavlink/v1/drone/takeoff

无人机起飞

**请求体**:
```json
{
  "system_id": 1,
  "altitude": 10
}
```

##### POST /mavlink/v1/drone/land

无人机降落

##### POST /mavlink/v1/drone/move

无人机移动

**请求体**:
```json
{
  "latitude": 22.543123,
  "longitude": 114.052345,
  "altitude": 10
}
```

##### POST /mavlink/v1/drone/return

无人机返航

---

#### 7.4 状态监控

##### GET /mavlink/v1/drone/status

获取无人机状态

**成功响应 (200)**:
```json
{
  "code": 0,
  "data": {
    "system_id": 1,
    "battery": 85.5,
    "mode": "AUTO",
    "armed": true
  }
}
```

##### GET /mavlink/v1/drone/position

获取无人机位置

---

### MAVLink V2 接口 (需认证)

> ⚠️ 以下接口需要在 Header 中携带有效的 JWT Token
> ⚠️ MAVLink V2 接口为组合操作，适合快速任务执行

基础路径: `/mavlink/v2`

#### 8.1 高级控制

##### POST /mavlink/v2/takeoff

一键起飞

##### POST /mavlink/v2/land

一键降落

##### POST /mavlink/v2/move

移动到目标位置

---

### 传感器接口 (无需认证)

> ⚠️ 传感器接口为公共接口，无需认证
> ⚠️ 用于 ESP32 等传感器设备发送警报消息

基础路径: `/api/sensor`

#### 9.1 POST /api/sensor/message 传感器警报

调度无人机前往指定位置执行任务

**请求地址**:
```
POST https://api.deeppluse.dpdns.org/api/sensor/message
```

**请求体**:
```json
{
  "sensor_id": "esp32_001",
  "sensor_ip": "192.168.1.100",
  "sensor_name": "FireSensor-A1",
  "alert_type": "fire",
  "alert_msg": "检测到明火",
  "latitude": 22.543123,
  "longitude": 114.052345,
  "timestamp": 1713177600,
  "severity": "high"
}
```

**参数说明**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| sensor_id | string | 是 | - | 传感器ID (硬件自带) |
| sensor_ip | string | 否 | 客户端IP | 传感器IP地址 |
| sensor_name | string | 否 | sensor_ip | 传感器名称 |
| alert_type | string | 是 | - | 警报类型 |
| alert_msg | string | 否 | "" | 预警消息 |
| latitude | float | 是 | - | GPS纬度 |
| longitude | float | 是 | - | GPS经度 |
| timestamp | int64 | 否 | 当前时间 | 时间戳 |
| severity | string | 否 | "medium" | 严重程度 |

**成功响应 (200)** - 任务链提交成功:
```json
{
  "code": 0,
  "message": "Task chain created and submitted",
  "chain_id": "20260415203000_fire",
  "assigned_drone": "drone_001",
  "task_count": 3,
  "alert_type": "fire",
  "target": {
    "latitude": 22.543123,
    "longitude": 114.052345
  }
}
```

**未成功调度响应 (202)** - 请求已接收但未成功处理:
```json
{
  "code": 0,
  "message": "Alert received, logged - No available drones",
  "sensor_id": "esp32_001",
  "alert_type": "fire",
  "drones": 0
}
```

**错误响应 (400)** - 请求参数错误:
```json
{
  "code": 1,
  "message": "Invalid request body",
  "error": "Key: 'SensorAlertRequest.SensorName' ..."
}
```

**错误响应 (500)** - 服务器内部错误:
```json
{
  "code": 1,
  "message": "DroneSearch not available"
}
```

**警报类型对应任务链**:

| alert_type | 任务链 | 说明 |
|------------|--------|------|
| `fire` | 起飞 → 飞往火源 → 区域侦察 | 火灾响应 |
| `rescue` | 起飞 → 飞往目标 → 网格搜索 → 降落 | 搜救任务 |
| `missing` | 起飞 → 飞往目标 → 网格搜索 → 降落 | 走失搜救 |
| `patrol` | 起飞 → 飞往巡逻点 → 盘旋巡逻 → 返航 | 巡逻任务 |
| `flood` | 起飞 → 飞往洪区 → 洪情侦察 | 洪灾响应 |
| 其他 | 起飞 → 飞往目标 → 返航 | 默认任务 |

---

#### 9.2 GET /api/sensor/status 获取无人机状态

获取当前可用无人机状态

**请求地址**:
```
GET https://api.deeppluse.dpdns.org/api/sensor/status
```

**成功响应 (200)**:
```json
{
  "code": 0,
  "drones": [
    {
      "board_id": "drone_001",
      "system_id": 1,
      "is_idle": true,
      "battery_level": 85.5,
      "latitude": 22.543123,
      "longitude": 114.052345,
      "altitude": 0,
      "last_update": 1713177600
    }
  ]
}
```

---

### 视频流接口 (需认证)

> ⚠️ 以下接口用于实时视频流的上传和获取
> ⚠️ Central 上传接口使用设备JWT认证，前端获取接口使用用户JWT认证
> ⚠️ 服务地址: `https://api.deeppluse.dpdns.org`

基础路径: `/api/board/live` (Central上传) 和 `/api/backend/live` (前端获取)

#### 10.1 视频流类型定义

##### 10.1.1 视频流状态枚举

```go
type StreamStatus string

const (
    StreamStatus_Connected    StreamStatus = "connected"     // 已连接
    StreamStatus_Disconnected StreamStatus = "disconnected"  // 已断开
    StreamStatus_Buffering    StreamStatus = "buffering"     // 缓冲中
    StreamStatus_Error        StreamStatus = "error"         // 错误
)
```

##### 10.1.2 视频编码格式枚举

```go
type VideoCodec string

const (
    VideoCodec_H264  VideoCodec = "h264"   // H.264 编码
    VideoCodec_H265  VideoCodec = "h265"   // H.265 编码
    VideoCodec_MJPEG VideoCodec = "mjpeg"  // Motion JPEG
)
```

##### 10.1.3 音频编码格式枚举

```go
type AudioCodec string

const (
    AudioCodec_AAC AudioCodec = "aac"   // AAC 音频
    AudioCodec_PCM AudioCodec = "pcm"   // PCM 音频
)
```

##### 10.1.4 视频流信息

```go
type LiveStreamInfo struct {
    StreamID        string      `json:"stream_id"`        // 流唯一ID
    TaskCode        string      `json:"task_code"`         // 关联任务代码
    CentralID       string      `json:"central_id"`       // Central设备ID
    DroneID         string      `json:"drone_id"`          // 无人机ID
    StreamStatus    StreamStatus `json:"status"`            // 流状态
    VideoCodec      VideoCodec  `json:"video_codec"`       // 视频编码
    AudioCodec      AudioCodec  `json:"audio_codec"`       // 音频编码
    Resolution      string      `json:"resolution"`       // 分辨率 (如 "1920x1080")
    FPS             int         `json:"fps"`               // 帧率
    Bitrate         int64       `json:"bitrate"`          // 比特率 (bytes/s)
    Duration        int64       `json:"duration"`         // 持续时间 (秒)
    StartTime       time.Time   `json:"start_time"`       // 开始时间
    LastUpdateTime  time.Time   `json:"last_update_time"` // 最后更新时间
    ViewerCount     int         `json:"viewer_count"`     // 当前观看人数
}
```

##### 10.1.5 BoardMessage 格式上传 (LiveStreamRequest)

```go
type LiveStreamRequest struct {
    MessageID   string                 `json:"message_id"`
    MessageTime int64                  `json:"message_time"`
    Message     LiveStreamMessageData  `json:"message"`
    FromID      string                 `json:"from_id"`       // Central ID
    FromType    string                 `json:"from_type"`     // "Central"
    ToID        string                 `json:"to_id"`         // "backend"
    ToType      string                 `json:"to_type"`       // "Backend"
}

type LiveStreamMessageData struct {
    MessageType string                 `json:"message_type"` // Request
    Attribute   string                 `json:"attribute"`    // Mission
    Connection  string                 `json:"connection"`   // HTTPS
    Command     string                 `json:"command"`      // VideoStream
    Data        map[string]interface{} `json:"data"`         // 视频流参数
}
```

#### 10.2 Central 上传接口 (设备JWT认证)

##### 10.2.1 POST /api/board/live - BoardMessage格式上传

使用 BoardMessage 格式上传视频流（推荐方式）

**请求地址**:
```
POST https://api.deeppluse.dpdns.org/api/board/live
```

**请求头**:
```
Authorization: Bearer <device_token>
X-Device-ID: <device_id>
X-Device-Type: <device_type>
Content-Type: multipart/form-data
```

**请求体** (multipart/form-data):

| Part | 类型 | 必填 | 说明 |
|------|------|------|------|
| metadata | JSON字符串 | 是 | LiveStreamRequest 结构体 |
| stream_data | 二进制 | 是 | 视频流数据 |

**metadata 示例**:
```json
{
  "message_id": "msg_live_001",
  "message_time": 1714234567,
  "message": {
    "message_type": "Request",
    "attribute": "Mission",
    "connection": "HTTPS",
    "command": "VideoStream",
    "data": {
      "task_code": "TASK_20260427_001",
      "video_codec": "h264",
      "audio_codec": "aac",
      "resolution": "1920x1080",
      "fps": 30
    }
  },
  "from_id": "central_001",
  "from_type": "Central",
  "to_id": "backend_001",
  "to_type": "Backend"
}
```

**参数说明** (metadata.data):

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| task_code | string | 是 | - | 任务代码（与原任务链相同） |
| video_codec | string | 否 | "h264" | 视频编码 (h264/h265/mjpeg) |
| audio_codec | string | 否 | - | 音频编码 (aac/pcm) |
| resolution | string | 否 | "1920x1080" | 分辨率 |
| fps | int | 否 | 30 | 帧率 |
| drone_id | string | 否 | - | 无人机ID |

**成功响应 (200)**:
```json
{
  "success": true,
  "data": {
    "stream_id": "stream_abc123...",
    "task_code": "TASK_20260427_001",
    "bytes_received": 1048576,
    "message": "视频流接收成功"
  }
}
```

**错误响应 (400)** - 缺少必要参数:
```json
{
  "success": false,
  "error": "缺少必要参数: X-Task-Code 和 X-Central-ID 必填",
  "code": "MISSING_REQUIRED_PARAMS"
}
```

---

##### 10.2.2 POST /api/board/live/raw - Header元数据格式上传

使用 HTTP Header 传递元数据，二进制流作为 Body

**请求地址**:
```
POST https://api.deeppluse.dpdns.org/api/board/live/raw
```

**请求头**:
```
Authorization: Bearer <device_token>
X-Device-ID: <device_id>
X-Device-Type: <device_type>
X-Task-Code: <task_code>          (必填)
X-Central-ID: <central_id>        (必填)
X-Drone-ID: <drone_id>            (可选)
X-Video-Codec: <video_codec>      (可选，默认 h264)
X-Audio-Codec: <audio_codec>      (可选)
X-Resolution: <resolution>        (可选，默认 1920x1080)
X-FPS: <fps>                      (可选，默认 30)
Content-Type: application/octet-stream
```

**请求体**: 视频二进制数据流（H.264 NALU 单元）

**成功响应 (200)**:
```json
{
  "success": true,
  "data": {
    "stream_id": "stream_abc123...",
    "task_code": "TASK_20260427_001",
    "bytes_received": 1048576,
    "message": "视频流接收成功"
  }
}
```

---

##### 10.2.3 POST /api/board/live/rtmp/start - 启动RTMP转码监听

启动 FFmpeg RTMP 转码监听服务，将 RTMP 流转换为 H.264

**请求地址**:
```
POST https://api.deeppluse.dpdns.org/api/board/live/rtmp/start
```

**请求头**:
```
Authorization: Bearer <device_token>
X-Device-ID: <device_id>
X-Device-Type: <device_type>
X-Task-Code: <task_code>          (必填)
X-Central-ID: <central_id>       (必填)
X-RTMP-URL: <rtmp_url>            (必填，RTMP流地址)
X-Listen-Addr: <listen_addr>     (可选，默认 127.0.0.1:8554)
```

**参数说明**:

| Header | 必填 | 说明 |
|--------|------|------|
| X-Task-Code | 是 | 任务代码 |
| X-Central-ID | 是 | Central设备ID |
| X-RTMP-URL | 是 | RTMP流地址 (如 `rtmp://source:1935/live/stream`) |
| X-Listen-Addr | 否 | 监听地址，默认 `127.0.0.1:8554` |

**成功响应 (200)**:
```json
{
  "success": true,
  "message": "RTMP 监听服务已启动",
  "data": {
    "task_code": "TASK_20260427_001",
    "listen_addr": "127.0.0.1:8554",
    "rtmp_url": "rtmp://source:1935/live/stream",
    "ffmpeg_pid": 12345
  }
}
```

**错误响应 (409)** - 服务已在运行:
```json
{
  "success": false,
  "error": "RTMP 监听服务已在运行",
  "code": "ALREADY_RUNNING"
}
```

---

##### 10.2.4 POST /api/board/live/rtmp/stop - 停止RTMP转码监听

停止 FFmpeg RTMP 转码监听服务

**请求地址**:
```
POST https://api.deeppluse.dpdns.org/api/board/live/rtmp/stop
```

**成功响应 (200)**:
```json
{
  "success": true,
  "message": "RTMP 监听服务已停止"
}
```

---

##### 10.2.5 GET /api/board/live/rtmp/status - 获取RTMP转码状态

获取当前 RTMP 转码监听服务状态

**请求地址**:
```
GET https://api.deeppluse.dpdns.org/api/board/live/rtmp/status
```

**成功响应 (200)** - 运行中:
```json
{
  "success": true,
  "data": {
    "running": true,
    "listen_addr": "127.0.0.1:8554",
    "rtmp_url": "rtmp://source:1935/live/stream",
    "task_code": "TASK_20260427_001",
    "uptime_seconds": 3600,
    "bytes_processed": 104857600
  }
}
```

**成功响应 (200)** - 未运行:
```json
{
  "success": true,
  "data": {
    "running": false
  }
}
```

---

#### 10.3 前端获取接口 (用户JWT认证)

##### 10.3.1 GET /api/backend/live - 获取视频流

获取实时视频流（支持 MJPEG/RAW/FLV 格式）

**请求地址**:
```
GET https://api.deeppluse.dpdns.org/api/backend/live?stream_id=xxx&format=mjpeg
```

**Query参数**:

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| stream_id | string | 否* | - | 流ID（优先使用） |
| task_code | string | 否* | - | 任务代码（备选） |
| format | string | 否 | "raw" | 格式: raw/mjpeg/flv |

*stream_id 和 task_code 至少提供一个

**成功响应 (200)** - RAW格式:
```
Content-Type: video/h264
[原始H.264二进制数据]
```

**成功响应 (200)** - MJPEG格式:
```
Content-Type: multipart/x-mixed-replace

--frame
Content-Type: image/jpeg
[JPEG图像数据]
--frame
Content-Type: image/jpeg
[JPEG图像数据]
...
```

**错误响应 (404)** - 流不存在:
```json
{
  "success": false,
  "error": "视频流不存在",
  "code": "STREAM_NOT_FOUND"
}
```

---

##### 10.3.2 GET /api/backend/live/ws - WebSocket视频流

通过 WebSocket 获取实时视频流（低延迟推荐）

**请求地址**:
```
wss://api.deeppluse.dpdns.org/api/backend/live/ws?stream_id=xxx
```

**Query参数**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| stream_id | string | 否* | 流ID（优先使用） |
| task_code | string | 否* | 任务代码（备选） |

*stream_id 和 task_code 至少提供一个

**WebSocket消息格式**:

| 消息类型 | 说明 | 格式 |
|----------|------|------|
| `h264` | H.264视频帧 | `{"type":"h264","data":"<base64>","timestamp":1234567890}` |
| `mjpeg` | JPEG图像帧 | `{"type":"mjpeg","data":"<base64>","timestamp":1234567890}` |
| `status` | 流状态更新 | `{"type":"status","status":"connected","viewers":5}` |
| `error` | 错误信息 | `{"type":"error","error":"stream not found"}` |

**前端接收示例** (JavaScript):
```javascript
const ws = new WebSocket('wss://api.deeppluse.dpdns.org/api/backend/live/ws?stream_id=stream_xxx');

ws.onmessage = (event) => {
    const msg = JSON.parse(event.data);

    if (msg.type === 'h264') {
        // 解码并显示H.264帧
        const frameData = atob(msg.data);
        // ... 使用VideoDecoder或类似库渲染
    } else if (msg.type === 'mjpeg') {
        // 显示MJPEG帧
        const img = document.getElementById('video');
        img.src = 'data:image/jpeg;base64,' + msg.data;
    } else if (msg.type === 'status') {
        console.log('Stream status:', msg.status, 'Viewers:', msg.viewers);
    }
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};

ws.onclose = () => {
    console.log('WebSocket closed');
};
```

---

##### 10.3.3 GET /api/backend/live/list - 获取活跃流列表

获取当前所有活跃视频流列表

**请求地址**:
```
GET https://api.deeppluse.dpdns.org/api/backend/live/list
```

**成功响应 (200)**:
```json
{
  "success": true,
  "data": [
    {
      "stream_id": "stream_abc123",
      "task_code": "TASK_20260427_001",
      "central_id": "central_001",
      "drone_id": "drone_001",
      "status": "connected",
      "video_codec": "h264",
      "resolution": "1920x1080",
      "fps": 30,
      "viewer_count": 2,
      "start_time": "2026-04-27T10:00:00Z",
      "last_update_time": "2026-04-27T10:30:00Z"
    }
  ]
}
```

---

##### 10.3.4 GET /api/backend/live/info/:stream_id - 获取流详情

获取指定视频流的详细信息

**请求地址**:
```
GET https://api.deeppluse.dpdns.org/api/backend/live/info/:stream_id
```

**成功响应 (200)**:
```json
{
  "success": true,
  "data": {
    "stream_id": "stream_abc123",
    "task_code": "TASK_20260427_001",
    "central_id": "central_001",
    "drone_id": "drone_001",
    "status": "connected",
    "video_codec": "h264",
    "audio_codec": "aac",
    "resolution": "1920x1080",
    "fps": 30,
    "bitrate": 5000000,
    "duration": 1800,
    "viewer_count": 2,
    "start_time": "2026-04-27T10:00:00Z",
    "last_update_time": "2026-04-27T10:30:00Z"
  }
}
```

**错误响应 (404)**:
```json
{
  "success": false,
  "error": "视频流不存在",
  "code": "STREAM_NOT_FOUND"
}
```

---

##### 10.3.5 DELETE /api/backend/live/:stream_id - 停止视频流

停止指定的视频流

**请求地址**:
```
DELETE https://api.deeppluse.dpdns.org/api/backend/live/:stream_id
```

**成功响应 (200)**:
```json
{
  "success": true,
  "message": "视频流已停止",
  "data": {
    "stream_id": "stream_abc123",
    "duration": 1800,
    "total_bytes": 104857600
  }
}
```

---

#### 10.4 视频流架构说明

##### 架构方式一：Central 直接输出 H.264（推荐）

```
┌─────────────┐     POST /api/board/live      ┌─────────────┐
│   Central   │ ─────────────────────────────→│   Backend   │
│  (H.264)    │        设备JWT认证             │   (Go)      │
└─────────────┘                               └──────┬──────┘
                                                      │
                              GET /api/backend/live   │
                              GET /api/backend/live/ws │
                                      │               │
                                      ▼               ▼
                               ┌─────────────────────────┐
                               │       Frontend         │
                               │   (Vite + Vue/React)   │
                               └─────────────────────────┘
```

**特点**:
- 无需转码，延迟最低（1-2秒）
- Central 直接输出 H.264 NALU 单元
- 适合已经具备 H.264 编码能力的设备

##### 架构方式二：FFmpeg RTMP 转码

```
┌─────────────┐                        ┌─────────────────┐
│   Central   │──RTMP──→  FFmpeg ──H.264→│    Backend     │
│   (RTMP)    │         转码            │    (Go)        │
└─────────────┘                        └────────┬────────┘
                                                 │
                              GET /api/backend/live/ws
                                      │
                                      ▼
                               ┌─────────────────────────┐
                               │       Frontend         │
                               └─────────────────────────┘
```

**特点**:
- FFmpeg 监听 RTMP 流，实时转码为 H.264
- 延迟稍高（2-3秒）但兼容性更好
- 支持 RTMP 协议的视频源接入

---

#### 10.5 前端集成示例

##### React + Hooks 视频流组件

```jsx
import React, { useRef, useEffect, useState } from 'react';

function VideoStream({ streamId, format = 'mjpeg' }) {
    const imgRef = useRef(null);
    const [status, setStatus] = useState('connecting');
    const [error, setError] = useState(null);

    useEffect(() => {
        if (format === 'mjpeg') {
            imgRef.current.src = `https://api.deeppluse.dpdns.org/api/backend/live?stream_id=${streamId}&format=mjpeg`;
        } else if (format === 'ws') {
            const ws = new WebSocket(`wss://api.deeppluse.dpdns.org/api/backend/live/ws?stream_id=${streamId}`);

            ws.onopen = () => setStatus('connected');
            ws.onmessage = (event) => {
                const msg = JSON.parse(event.data);
                if (msg.type === 'mjpeg' && imgRef.current) {
                    imgRef.current.src = 'data:image/jpeg;base64,' + msg.data;
                } else if (msg.type === 'status') {
                    setStatus(msg.status);
                } else if (msg.type === 'error') {
                    setError(msg.error);
                }
            };
            ws.onerror = () => setError('WebSocket连接失败');
            ws.onclose = () => setStatus('disconnected');

            return () => ws.close();
        }
    }, [streamId, format]);

    if (error) return <div className="error">错误: {error}</div>;

    return (
        <div className="video-container">
            <img ref={imgRef} alt="Video Stream" style={{ width: '100%' }} />
            <div className="status">状态: {status}</div>
        </div>
    );
}

export default VideoStream;
```

##### Vue 3 视频流组件

```vue
<template>
  <div class="video-stream">
    <img v-if="format === 'mjpeg'" :src="mjpegUrl" alt="Video Stream" />
    <canvas v-else ref="canvasRef"></canvas>
    <div class="status">状态: {{ status }}</div>
    <div v-if="error" class="error">错误: {{ error }}</div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted } from 'vue';

const props = defineProps({
    streamId: { type: String, required: true },
    format: { type: String, default: 'mjpeg' }
});

const status = ref('connecting');
const error = ref(null);
const canvasRef = ref(null);
let ws = null;

const mjpegUrl = computed(() =>
    `https://api.deeppluse.dpdns.org/api/backend/live?stream_id=${props.streamId}&format=mjpeg`
);

onMounted(() => {
    if (props.format === 'ws') {
        ws = new WebSocket(`wss://api.deeppluse.dpdns.org/api/backend/live/ws?stream_id=${props.streamId}`);

        ws.onopen = () => status.value = 'connected';
        ws.onmessage = (event) => {
            const msg = JSON.parse(event.data);
            if (msg.type === 'status') status.value = msg.status;
            else if (msg.type === 'error') error.value = msg.error;
            else if (msg.type === 'mjpeg' && canvasRef.value) {
                const ctx = canvasRef.value.getContext('2d');
                const img = new Image();
                img.onload = () => {
                    canvasRef.value.width = img.width;
                    canvasRef.value.height = img.height;
                    ctx.drawImage(img, 0, 0);
                };
                img.src = 'data:image/jpeg;base64,' + msg.data;
            }
        };
        ws.onerror = () => error.value = 'WebSocket连接失败';
        ws.onclose = () => status.value = 'disconnected';
    }
});

onUnmounted(() => {
    if (ws) ws.close();
});
</script>
```

---

#### 10.6 视频流错误码

| 错误码 | 说明 |
|--------|------|
| `MISSING_REQUIRED_PARAMS` | 缺少必要参数 |
| `STREAM_NOT_FOUND` | 视频流不存在 |
| `ALREADY_RUNNING` | RTMP监听服务已在运行 |
| `FFMPEG_START_ERROR` | FFmpeg启动失败 |
| `LISTENER_START_ERROR` | TCP监听启动失败 |
| `METADATA_READ_ERROR` | 元数据读取失败 |
| `METADATA_PARSE_ERROR` | 元数据JSON解析失败 |
| `LIVE_ENDPOINT_NOT_FOUND` | 视频流接口不存在 |

---

## 十一、进度链接口 (需认证)

> ⚠️ 以下接口需要在 Header 中携带有效的 JWT Token

基础路径: `/api/chain`

#### 11.1 链管理

##### POST /api/chain/create

创建进度链

**请求体**:
```json
{
  "tasks": [
    {
      "command": "takeoff",
      "data": { "altitude": 30 },
      "timeout": 30
    },
    {
      "command": "goto",
      "data": { "latitude": 22.5, "longitude": 114.0, "altitude": 25 },
      "timeout": 60
    }
  ]
}
```

**成功响应 (201)**:
```json
{
  "code": 0,
  "chain_id": "chain_xxx",
  "status": "pending"
}
```

##### GET /api/chain/:id

获取进度链详情

##### GET /api/chain/list

列出所有进度链

---

#### 11.2 节点管理

##### POST /api/chain/:id/node/add

添加节点到链

##### DELETE /api/chain/:id/node/delete/:nodeId

从链中删除节点

---

#### 11.3 执行控制

##### POST /api/chain/:id/start

启动进度链执行

##### POST /api/chain/:id/stop

停止进度链

---

## 十二、消息处理流程

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
│       │                                    ▼ (HTTPS)             │
│       │                              Central 服务器               │
│       │                           (central.deeppluse.dpdns.org)  │
│       │                                                         │
│       └─── NO ──→ BoardConn 内部处理 (messageChan)                │
│                       │                                          │
│                       ▼                                          │
│                BoardHandler (保留，暂不使用)                       │
└─────────────────────────────────────────────────────────────────┘

### 后端向 Central 发送消息 (HTTPS)

当后端需要向 Central 服务器发送任务链或控制指令时，使用 HTTPS 协议：

```
后端 Backend
    │
    ▼ (HTTPS POST /central/message)
Central 服务器 (central.deeppluse.dpdns.org)
    │
    ▼
任务链执行 / 指令处理
```

**消息结构**:
```go
type BoardMessage struct {
    MessageID   string    `json:"message_id"`
    MessageTime time.Time `json:"message_time"`
    Message     Message   `json:"message"`
    FromID      string    `json:"from_id"`
    FromType    string    `json:"from_type"`
    ToID        string    `json:"to_id"`
    ToType      string    `json:"to_type"`
}
```

### 消息类型判断

| 消息类型 | FromType | Command/Attribute | Handler |
|----------|----------|-------------------|---------|
| 传感器警报 | sensor/esp32/alarm | Warning/SensorAlert/Alert | SensorAlertHandler |
| 飞控消息 | board/drone/fc | Heartbeat/Status/Control | BoardConn 内部处理 |

---

## 十三、错误响应格式

### 通用错误响应

```json
{
  "success": false,
  "error": "错误信息",
  "code": 400
}
```

### 带详细信息的错误响应

```json
{
  "code": 1,
  "message": "Invalid request body",
  "error": "具体错误详情"
}
```

---

## 十四、注意事项

1. 所有受保护接口都需要在 Header 中携带 JWT Token
2. Token 默认存储在 Redis DB 13 中，登出后 Token 即失效
3. 管理员接口需要用户具有 `is_admin: true` 权限
4. MAVLink V1 接口为原子化操作，适合 AI 精细控制
5. MAVLink V2 接口为组合操作，适合快速任务执行
6. 进度链支持最多 1000 步任务执行
7. Board 消息通过 TCP/UDP 直连后端，由 MessageDispatcher 自动路由
8. Sensor 消息走 SensorAlertHandler 生成任务链
9. Board 消息暂由 BoardConn 内部处理 (BoardHandler 保留但未启用)
10. 传感器接口为公共接口，无需认证，适合 ESP32 等设备直接调用
11. 后端向 Central 发送消息使用 HTTPS 协议，地址: `https://central.deeppluse.dpdns.org/central/message`
12. 设备认证使用独立系统，错误提示友好，不会触发封禁

---

*文档版本: 3.1*
*最后更新: 2026-04-15*
