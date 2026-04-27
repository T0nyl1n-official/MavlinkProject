# MavlinkProject 无人机调度管理系统

## 项目概述

MavlinkProject 是一个基于 MAVLink 协议的无人机通信与调度管理系统。该系统提供完整的无人机控制、监控和调度功能，支持多无人机管理、传感器响应、进度链执行、实时视频流等核心功能。

## 技术栈

- **后端框架**: Gin (Go语言Web框架)
- **数据库**: MySQL (数据持久化)
- **缓存**: Redis (Token存储、验证码、错误日志)
- **配置文件**: YAML (集中化配置管理) + 环境变量覆盖
- **通信**: TCP/UDP、HTTPS、FRP (远程端口转发)

## 核心功能模块

### 1. 用户认证系统
- 用户注册与登录
- JWT Token 认证
- 角色权限管理 (普通用户/管理员)

### 2. 设备认证系统
- 硬件设备独立登录 (Central/LandNode/Sensor/Drone)
- 专用JWT Token认证
- 设备状态管理 (在线/离线)
- 友好的错误提示 (不触发封禁)

### 3. BoardConn 通信模块
- **TCP/UDP 服务器**: 接收来自 Board/Sensor 的消息
- **消息分发器 (MessageDispatcher)**: 策略模式路由消息
- **多 Handler 支持**

### 4. Handler 模块
#### Sensor 文件夹 (传感器相关处理)
- `SensorAlertHandler`: 处理传感器警报，自动生成任务链
- `AIAgentHandler`: AI Agent 扩展点 (预留接口)
- `SensorMessageHandler.go`: 任务链生成逻辑

#### Boards 文件夹 (Board 通信管理)
- `BoardConnection.go`: Board 连接管理
- `BoardHandler.go`: 飞控消息处理 (保留，暂不使用)
- `MessageDispatcher.go`: 消息分发器
- `MessageSender.go`: FRP 消息发送
- `LiveStreamHandler.go`: 实时视频流处理

### 5. FRP 多中心支持
- 支持配置多个 Central 服务器
- 自动重试机制 (可配置重试次数)
- 故障转移: 尝试所有 Central 都失败后才报错
- **HTTPS 支持**: 后端可通过 HTTPS 向 Central 发送消息

### 6. 进度链系统 (Progress Chain)
- 链式任务执行
- 动态任务生成
- 状态跟踪

### 7. 实时视频流模块 (LiveStream)
- **Board/Live 接口**: Central 上传视频流
- **Backend/Live 接口**: 前端获取实时视频
- **多协议支持**: MJPEG、WebSocket、RAW、FLV
- **任务链关联**: 视频流与任务链绑定
- **FFmpeg 转码支持**: 支持 RTMP 流实时转码为 H.264

#### 7.1 视频流架构

##### 方式一：Central 直接输出 H.264（推荐，最简低延迟）
```
Central (H.264 NALU) → POST /api/board/live → Backend → WebSocket/MJPEG → 前端
```
- Central 直接输出 H.264 编码视频流
- 无需额外转码，延迟最低（1-2秒）
- 适合已经具备 H.264 编码能力的设备

##### 方式二：FFmpeg RTMP 转码（支持 RTMP 源）
```
Central (RTMP) → FFmpeg 转换 → Backend (H.264) → WebSocket/MJPEG → 前端
```
- FFmpeg 监听 RTMP 流，实时转码为 H.264
- 支持 RTMP 协议的视频源接入
- 延迟稍高（2-3秒）但兼容性更好

#### 7.2 视频流数据类型

| 类型 | 说明 | 使用场景 |
|------|------|----------|
| `h264` | H.264 视频编码 | 主流推荐格式，低带宽占用 |
| `h265` | H.265 视频编码 | 更高压缩率，适合高分辨率 |
| `mjpeg` | Motion JPEG | 简单实现，兼容性最好 |

#### 7.3 视频流状态

| 状态 | 说明 |
|------|------|
| `connected` | 已连接，正在传输 |
| `disconnected` | 已断开 |
| `buffering` | 缓冲中 |
| `error` | 错误状态 |

### 8. 集中配置管理
- 所有配置存储于 `Setting.yaml`
- 支持运行时配置更新
- **环境变量覆盖**: 优先读取环境变量，环境变量为空时读取 YAML
- 分模块管理 (Board、FRP、Server 等)

## 项目目录结构

```
MavlinkProject/
├── Server/Backend/
│   ├── Backend.go                 # 后端服务入口
│   ├── BackendAccessor.go        # 后端访问器 (解耦)
│   ├── Config/                    # 配置管理
│   │   └── SettingManager.go      # 配置加载与内存管理
│   ├── Database/                  # 数据库配置
│   ├── Handler/                   # 业务处理器
│   │   ├── Boards/               # Board 通信处理器
│   │   │   ├── BoardConnection.go    # Board 连接管理
│   │   │   ├── BoardHandler.go        # 飞控消息处理 (保留)
│   │   │   ├── MessageDispatcher.go  # 消息分发器
│   │   │   ├── MessageSender.go      # FRP 消息发送
│   │   │   ├── LiveStreamHandler.go   # 实时视频流处理
│   │   │   └── SensorBoard/          # 传感器处理
│   │   │       ├── SensorAlertHandler.go # 传感器警报处理
│   │   │       ├── SensorMessageHandler.go # 任务链生成
│   │   │       └── types.go           # 类型定义
│   │   ├── Device/               # 设备认证处理
│   │   ├── Mavlink/              # MAVLink处理器
│   │   ├── Sensor/               # 传感器处理
│   │   ├── Users/                # 用户处理
│   │   └── ProgressChain/        # 进度链处理器
│   ├── Middles/                   # 中间件
│   ├── Routes/                    # 路由定义
│   │   ├── Boards/              # Board 路由
│   │   │   ├── BoardMessageRoute.go
│   │   │   └── LiveStreamRoutes.go   # 视频流路由
│   │   ├── Device/              # 设备路由
│   │   ├── Sensor/              # 传感器路由
│   │   ├── Terminal/            # 终端路由
│   │   └── User/                # 用户路由
│   ├── Shared/                    # 共享结构体
│   │   ├── Boards/             # Board 消息格式定义
│   │   │   ├── Board_MessageFormat.go
│   │   │   └── LiveStream_Types.go    # 视频流类型
│   │   ├── Device/              # 设备数据模型
│   │   └── FRPHelper/           # FRP 通信封装
│   │       ├── FRPHelper.go           # FRP TCP 通信封装
│   │       └── CentralHTTPClient.go   # Central HTTPS 客户端
│   └── Utils/                     # 工具函数
├── config/
│   └── Setting.yaml              # 集中配置文件
├── tests/
│   └── OutputHistory/             # 测试输出历史
├── docs/
│   ├── API.md                    # API 文档
│   ├── requirements.md           # 需求文档
│   ├── tech-doc.md              # 技术文档
│   └── Frontend_LiveStream_Guide.md  # 前端视频流指南
└── README.md
```

## 系统流程

### Board 消息处理流程
```
Sensor/Board 设备
      │
      ▼ (TCP/UDP)
BoardConn 后端监听 (0.0.0.0:8081 TCP / 0.0.0.0:8082 UDP)
      │
      ▼
isSensorMessage() 判断
      │
      ├─── YES ──→ MessageDispatcher ──→ SensorAlertHandler
      │                                    │
      │                                    ▼
      │                          GenerateChainAndSendToCentral
      │                                    │
      │                                    ▼ (HTTPS)
      │                              Central 服务器
      │                           (central.deeppluse.dpdns.org)
      │
      └─── NO ──→ BoardConn 内部处理 (messageChan)
                      │
                      ▼
               BoardHandler (保留，暂不使用)
```

### 实时视频流流程
```
Central (无人机)
    │
    │ POST /api/board/live (BoardMessage + 视频二进制)
    ▼
Backend (Go/Gin)
    │
    ├── 缓冲视频流
    └── 转发
        │
        ├── GET /api/backend/live (MJPEG/WebSocket)
        ▼
Frontend (Vite + Vue/React)
    │
    ▼
<video> 或 <canvas> 实时展示
```

### 消息类型判断

| 消息类型 | FromType | Command/Attribute | 路由 |
|----------|----------|------------------|------|
| 传感器警报 | sensor/esp32/alarm | Warning/SensorAlert/Alert | → Dispatcher → SensorAlertHandler |
| 飞控消息 | board/drone/fc | Heartbeat/Status/Control | → BoardConn 内部处理 |
| 视频流 | central | VideoStream | → LiveStreamHandler |

### 传感器警报处理流程
```
1. Sensor 检测异常
2. 通过 TCP/UDP 发送 SensorAlert 到 BoardConn
3. isSensorMessage() 识别为传感器消息
4. MessageDispatcher 路由到 SensorAlertHandler
5. SensorAlertHandler 处理:
   - 解析位置信息
   - 生成任务链 (TakeOff → GoTo → TakePhoto → Land)
   - 通过 CentralHTTPClient 发送到 Central
6. Central 执行任务链
```

### 用户认证流程
```
1. POST /users/register → 注册用户
2. POST /users/login → 登录返回 JWT Token
3. Header 添加 Authorization: Bearer <token>
4. 访问受保护接口
```

### 设备认证流程
```
1. POST /device/login → 设备登录获取 Token
2. Header 添加 X-Device-ID 和 X-Device-Type
3. 访问设备专属接口
```

## 环境变量配置

系统优先读取环境变量，环境变量为空时使用 YAML 配置或默认值。

### MySQL 数据库配置
| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `MavlinkProject_backend_database_mysql_host` | MySQL 主机 | localhost |
| `MavlinkProject_backend_database_mysql_port` | MySQL 端口 | 3306 |
| `MavlinkProject_backend_database_mysql_user` | 用户名 | root |
| `MavlinkProject_backend_database_mysql_password` | 密码 | (空) |
| `MavlinkProject_backend_database_mysql_database` | 数据库名 | mavlinkproject |
| `MavlinkProject_backend_database_mysql_charset` | 字符集 | utf8mb4 |

### Redis 配置
| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `MavlinkProject_backend_redis_host` | Redis 主机 | localhost |
| `MavlinkProject_backend_redis_port` | Redis 端口 | 6379 |
| `MavlinkProject_backend_redis_password` | 密码 | (空) |

### JWT 配置
| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `MavlinkProject_backend_jwt_secret_key` | JWT 密钥 | MavlinkBackendMadeByTonyl1n |

### SMTP 邮件配置
| 环境变量 | 说明 | 默认值 |
|----------|------|--------|
| `SMTP_HOST` | SMTP 服务器 | smtp.qq.com |
| `SMTP_USERNAME` | 用户名 | (空) |
| `SMTP_PASSWORD` | 密码/授权码 | (空) |
| `SMTP_FROM_EMAIL` | 发件人邮箱 | (空) |
| `SMTP_FROM_NAME` | 发件人名称 | MavlinkProject |

## Redis 数据库分配

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

## API 接口概览

### 用户接口
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /users/register | 用户注册 | 无 |
| POST | /users/login | 用户登录 | 无 |
| POST | /users/logout | 用户登出 | JWT |
| PUT | /users/profile | 更新用户信息 | JWT |
| PUT | /users/password | 修改密码 | JWT |

### 设备接口
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /device/login | 设备登录 | 无 |
| POST | /device/logout | 设备登出 | 设备JWT |
| GET | /device/status | 设备状态 | 设备JWT |

### 传感器接口
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/sensor/message | 接收传感器警报 | 无 |
| GET | /api/sensor/status | 传感器状态 | 无 |

### 视频流接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/board/live | Central 上传视频流 (BoardMessage格式) | 设备JWT |
| POST | /api/board/live/raw | Central 上传视频流 (Header元数据) | 设备JWT |
| POST | /api/board/live/rtmp/start | 启动RTMP转码监听 | 设备JWT |
| POST | /api/board/live/rtmp/stop | 停止RTMP转码监听 | 设备JWT |
| GET | /api/board/live/rtmp/status | 获取RTMP转码状态 | 设备JWT |
| POST | /api/board/live/ffmpeg | FFmpeg直接转码 | 设备JWT |
| GET | /api/backend/live | 前端获取视频流 (MJPEG/RAW/FLV) | JWT |
| GET | /api/backend/live/ws | 前端WebSocket视频流 | JWT |
| GET | /api/backend/live/list | 获取活跃流列表 | JWT |
| GET | /api/backend/live/info/:stream_id | 获取指定流详情 | JWT |
| DELETE | /api/backend/live/:stream_id | 停止指定视频流 | JWT |

### 终端接口
| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /terminal/message | 终端命令 | JWT |

## 部署要求

- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

## 安全特性

- **JWT Token 认证**
- **Token 存储于 Redis (支持登出即失效)**
- **密码加密 (MD5)**
- **角色权限控制**
- **CORS 中间件**
- **请求限流**
- **设备独立认证系统**
- **HTTPS 加密通信**

## 许可证

本项目仅供学习和研究使用。
