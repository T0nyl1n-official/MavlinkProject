# MavlinkProject 无人机调度管理系统

## 项目概述

MavlinkProject 是一个基于 MAVLink 协议的无人机通信与调度管理系统。该系统提供完整的无人机控制、监控和调度功能，支持多无人机管理、传感器响应、进度链执行等核心功能。

## 技术栈

- **后端框架**: Gin (Go语言Web框架)
- **数据库**: MySQL (数据持久化)
- **缓存**: Redis (Token存储、验证码、错误日志)
- **配置文件**: YAML (集中化配置管理)
- **通信**: TCP/UDP、FRP (远程端口转发)

## 核心功能模块

### 1. 用户认证系统
- 用户注册与登录
- JWT Token 认证
- 角色权限管理 (普通用户/管理员)

### 2. BoardConn 通信模块
- **TCP/UDP 服务器**: 接收来自 Board/Sensor 的消息
- **消息分发器 (MessageDispatcher)**: 策略模式路由消息
- **多 Handler 支持**

### 3. Handler 模块
#### Sensor 文件夹 (传感器相关处理)
- `SensorAlertHandler`: 处理传感器警报，自动生成任务链
- `AIAgentHandler`: AI Agent 扩展点 (预留接口)
- `SensorMessageHandler.go`: 任务链生成逻辑

#### Boards 文件夹 (Board 通信管理)
- `BoardConnection.go`: Board 连接管理
- `BoardHandler.go`: 飞控消息处理 (保留，暂不使用)
- `MessageDispatcher.go`: 消息分发器
- `MessageSender.go`: FRP 消息发送

### 4. FRP 多中心支持
- 支持配置多个 Central 服务器
- 自动重试机制 (可配置重试次数)
- 故障转移: 尝试所有 Central 都失败后才报错

### 5. 进度链系统 (Progress Chain)
- 链式任务执行
- 动态任务生成
- 状态跟踪

### 7. 集中配置管理
- 所有配置存储于 `Setting.yaml`
- 支持运行时配置更新
- 分模块管理 (Board、FRP、Server 等)

## 项目目录结构

```
MavlinkProject/
├── Server/Backend/
│   ├── Backend.go                 # 后端服务入口
│   ├── Config/                    # 配置管理
│   │   └── SettingManager.go      # 配置加载与内存管理
│   ├── Database/                  # 数据库配置
│   ├── Handler/                   # 业务处理器
│   │   ├── Boards/                # Board 通信处理器
│   │   │   ├── BoardConnection.go    # Board 连接管理
│   │   │   ├── BoardHandler.go        # 飞控消息处理 (保留)
│   │   │   ├── MessageDispatcher.go  # 消息分发器
│   │   │   ├── MessageSender.go      # FRP 消息发送
│   │   │   └── SensorBoard/          # 传感器处理 (从 Sensor 移入)
│   │   │       ├── SensorAlertHandler.go # 传感器警报处理
│   │   │       ├── SensorMessageHandler.go # 任务链生成
│   │   │       └── types.go           # 类型定义
│   │   ├── Mavlink/              # MAVLink处理器
│   │   └── ProgressChain/        # 进度链处理器
│   ├── Middles/                   # 中间件
│   ├── Routes/                    # 路由定义
│   ├── Shared/                    # 共享结构体
│   │   ├── Boards/               # Board 消息格式定义
│   │   └── FRPHelper/            # FRP 通信封装 (新建)
│   └── Utils/                     # 工具函数
├── config/
│   └── Setting.yaml              # 集中配置文件
├── tests/
│   └── OutputHistory/             # 测试输出历史
├── docs/
│   └── API.md                    # API 文档
└── README.md
```

## 系统流程

### Board 消息处理流程
```
Sensor/Board 设备
      │
      ▼ (TCP/UDP)
BoardConn 后端监听
      │
      ▼
isSensorMessage() 判断
      │
      ├─── YES ──→ MessageDispatcher ──→ SensorAlertHandler
      │                                    │
      │                                    ▼
      │                         GenerateChainAndSendToCentral
      │                                    │
      │                                    ▼ (FRP)
      │                              Central 服务器
      │
      └─── NO ──→ BoardConn 内部处理 (messageChan)
                      │
                      ▼
               BoardHandler (保留，暂不使用)
```

### 消息类型判断

| 消息类型 | FromType | Command/Attribute | 路由 |
|----------|----------|------------------|------|
| 传感器警报 | sensor/esp32/alarm | Warning/SensorAlert/Alert | → Dispatcher → SensorAlertHandler |
| 飞控消息 | board/drone/fc | Heartbeat/Status/Control | → BoardConn 内部处理 |

### 传感器警报处理流程
```
1. Sensor 检测异常
2. 通过 TCP 发送 SensorAlert 到 BoardConn
3. isSensorMessage() 识别为传感器消息
4. MessageDispatcher 路由到 SensorAlertHandler
5. SensorAlertHandler 处理:
   - 解析位置信息
   - 生成任务链 (TakeOff → GoTo → TakePhoto → Land)
   - 通过 FRPHelper 发送到 Central (树莓派)
6. Central 执行任务链
```

### 用户认证流程
```
1. POST /users/register → 注册用户
2. POST /users/login → 登录返回 JWT Token
3. Header 添加 Authorization: Bearer <token>
4. 访问受保护接口
```

## 配置说明

### Setting.yaml 结构

```yaml
server:
  port: "8080"
  mode: debug

board:
  listening:
    tcp_addr: "0.0.0.0"
    tcp_port: "8081"
    udp_addr: "0.0.0.0"
    udp_port: "8082"
  connection:
    timeout: 180
    max_retry_attempts: 3
    retry_delay: 5
    keepalive_interval: 10
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

### Redis 数据库分配

| DB | 用途 |
|----|------|
| 0 | 通用警告 |
| 1 | Backend 错误 |
| 2 | Frontend 错误 |
| 3 | Agent 错误 |
| 4 | Drone 错误 |
| 5 | Sensor 错误 |
| 13 | Token 存储 |
| 14 | 验证码 |

## 部署要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

## 安全特性

- **JWT Token 认证**
- **Token 存储于 Redis (支持登出即失效)**
- **密码加密**
- **角色权限控制**
- **CORS 中间件**
- **请求限流**
