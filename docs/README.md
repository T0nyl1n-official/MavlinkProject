# MavlinkProject 无人机调度管理系统

## 项目概述

MavlinkProject 是一个基于 MAVLink 协议的无人机通信与调度管理系统。该系统提供完整的无人机控制、监控和调度功能，支持多无人机管理、充电站调度、进度链执行等核心功能。

## 技术栈

- **后端框架**: Gin (Go语言Web框架)
- **数据库**: MySQL (数据持久化)
- **缓存**: Redis (Token存储、验证码、错误日志)
- **无人机通信**: bluenviron/gomavlib (MAVLink协议)
- **认证**: JWT (JSON Web Token)

## 核心功能模块

### 1. 用户认证系统
- 用户注册与登录
- JWT Token 认证
- 角色权限管理 (普通用户/管理员)
- 邮箱验证码验证

### 2. MAVLink 通信模块
- **V1 API**: 原子化单功能接口，适合 AI 精细控制
  - 连接管理 (UDP/TCP/Serial)
  - 无人机控制 (起飞、降落、移动、返回)
  - 状态监控 (位置、电量、姿态)
  - 地面站配置
- **V2 API**: 组合式高级接口，适合人工编排
  - 一键起飞/降落
  - 传感器警报响应 (调度无人机拍照)
  - 充电站自动调度

### 3. 进度链系统 (Progress Chain)
- 最多 1000 步的链式任务执行
- 动态节点插入
- 任务状态跟踪
- 执行日志记录

### 4. 充电站管理
- 充电仓位管理
- 自动分配最优充电位置
- 无人机自动回充

### 5. 错误监控系统
- 分层错误分类 (Backend/Frontend/Agent/Drone/Sensor)
- Redis 分布式存储
- 错误日志记录

### 9. AI 模型集成系统
- **LSTM 时序模型**: 井下传感器时序数据异常检测
- **YOLOv8s 热源检测**: 无人机拍摄图片热源异常识别（支持温度等级判断）
- **实时告警推送**: WebSocket + SSE 双通道推送到前端 (www.deeppluse.dpdns.org)
- **标注图自动下载**: YOLO 返回的标注图自动保存到 output/yolo_Generated/
- **模型客户端**: 统一 ModelClient 管理 LSTM/YOLO 连接、重试、健康检查

## 项目目录结构

```
MavlinkProject/
├── Server/Backend/
│   ├── Backend.go                 # 后端服务入口
│   ├── Database/                  # 数据库配置
│   │   ├── Config/               # MySQL/Redis配置
│   │   ├── MysqlService.go       # MySQL服务
│   │   └── RedisService.go       # Redis服务
│   ├── Handler/                   # 业务处理器
│   │   ├── Mavlink/              # MAVLink处理器
│   │   ├── ProgressChain/        # 进度链处理器
│   │   ├── Users/                # 用户处理器
│   │   ├── WarningHandle/        # 错误处理
│   │   └── AI/                   # AI 处理器 (新增)
│   │       ├── analysis_service.go  # AI 分析服务 (传感器+无人机)
│   │       ├── alert_hub.go        # 告警中心 (WebSocket+SSE)
│   │       ├── ai_handler.go       # HTTP Handler
│   │       └── thermal_handler.go  # YOLOv8 热源检测 Handler
│   ├── Middles/                   # 中间件
│   │   ├── Jwt/                  # JWT认证
│   │   ├── ErrorMiddleHandle/    # 错误处理
│   │   ├── RateLimit/           # 限流
│   │   └── Security/            # 安全
│   ├── Routes/                    # 路由定义
│   │   ├── Mavlink/              # MAVLink路由
│   │   ├── User/                 # 用户路由
│   │   └── Routes.go             # 路由入口
│   ├── Shared/                    # 共享结构体
│   │   ├── Charge/              # 充电站
│   │   ├── Drones/              # 无人机
│   │   ├── LandNode/            # 地面站
│   │   ├── User/                # 用户
│   │   └── Warnings/            # 错误定义
│   ├── Models/                   # AI 数据类型与模型客户端 (新增)
│   │   ├── types.go             # AI 类型定义 (LSTM/YOLO/AlertJSON/Thermal)
│   │   └── model_client.go      # ModelClient (LSTM+YOLO API调用)
│   └── Utils/                     # 工具函数
│       ├── Encryption/           # 加密
│       └── Verification/         # 验证码
├── config/
│   └── Setting.yaml               # 含 ai.lstm / ai.yolo 配置段
├── output/
│   ├── thermal_photos/            # 无人机上传原始照片 (自动创建)
│   └── yolo_Generated/             # YOLO 标注图 (自动创建)
└── main.go                       # 程序入口
```

## 系统流程

### 用户认证流程
```
1. POST /users/register → 注册用户
2. POST /users/login → 登录返回 JWT Token
3. Header 添加 Authorization: Bearer <token>
4. 访问受保护接口
```

### 无人机控制流程 (V1)
```
1. 创建 Handler: POST /mavlink/v1/handler/create
2. 启动连接: POST /mavlink/v1/connection/start
3. 发送控制指令: POST /mavlink/v1/drone/takeoff
4. 获取状态: GET /mavlink/v1/drone/status
```

### 进度链执行流程
```
1. 创建链: POST /api/chain/create
2. 添加节点: POST /api/chain/:id/node/add
3. 启动执行: POST /api/chain/:id/start
4. 监控状态: GET /api/chain/:id
```

### AI 热源检测流程
```
Central/无人机 (回传照片)
    │
    ▼ POST /api/ai/drone/photo (multipart: photo + drone_id + lat + lon)
Backend (Go/Gin)
    │
    ├── 1. 保存原图到 output/thermal_photos/
    ├── 2. 调用 ModelClient.ThermalDetect()
    │       │
    │       ▼ POST {yolo_url}/api/v1/detect?conf=&iou=&imgsz= (multipart)
    │   YOLOv8 服务 (FastAPI, CUDA推理)
    │       │
    │       ▼ 返回 ThermalDetectResponse (detections + temperature.level)
    │
    ├── 3. 解析 temperature.level → Severity (HIGH Lv1→critical, HIGH Lv2→high)
    ├── 4. 生成 AlertJSON → Broadcast (WebSocket + SSE)
    ├── 5. 下载 image_annotated 到 output/yolo_Generated/
    └── 6. 返回 JSON (code=0, alert, raw_result, annotated_image_url)
            │
            ▼
Frontend (www.deeppluse.dpdns.org)
    │
    ├── WebSocket /api/ai/alerts/ws ← 实时接收 AlertJSON
    ├── SSE /api/ai/alerts/sse     ← 实时接收 AlertJSON
    └── 解析坐标+severity → 3D地图红色高亮警报
```

## 部署要求

- Go 1.25+
- MySQL 8.0+
- Redis 6.0+
- Python 3.x + FastAPI (YOLOv8s 算法服务，可选)
- NVIDIA GPU + CUDA (YOLO 推理加速，可选)

## 配置说明

### 环境变量
- `MavlinkMysqlDSN`: MySQL 连接字符串

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

## 安全特性

- **JWT Token 认证**
- **Token 存储于 Redis (支持登出即失效)**
- **密码 MD5 加密**
- **角色权限控制**
- **CORS 中间件**
- **XSS 防护**
- **请求限流**

## API 接口概览

### 核心 AI 接口

| 方法 | 路径 | 说明 | 认证 |
|------|------|------|------|
| POST | /api/ai/analyze/sensor | LSTM 传感器分析 | 无 |
| POST | /api/ai/analyze/drone | YOLO 图像分析(Base64) | 无 |
| POST | /api/ai/drone/photo | 无人机照片+热源检测 | 设备JWT |
| GET | /api/ai/alerts/ws | WebSocket 实时告警 | 用户JWT |
| GET | /api/ai/alerts/sse | SSE 实时告警 | 用户JWT |
| GET | /api/ai/alerts/history | 告警历史查询 | 用户JWT |
| GET | /api/ai/model/status | 模型状态 | 无 |
| GET | /api/ai/drone/photo/generated/:filename | 下载标注图 | 无 |

> 📌 详细接口文档请查看: [API.md](./API.md#十五ai-模型集成接口)
