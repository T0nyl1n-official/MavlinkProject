# Central Board API 文档

## 1. 服务信息

### 1.1 服务地址

| 环境 | 地址 | 端口 |
|------|------|------|
| 本地开发 | `http://localhost:8084` | 8084 |
| 生产环境 | `https://central.deeppluse.dpdns.org` | 8084 |

### 1.2 认证方式

- **认证方式**: 无需认证（内部服务）
- **Content-Type**: `application/json`

## 2. 接口列表

| 方法 | 路径 | 描述 | 状态码 |
|------|------|------|--------|
| GET | `/health` | 健康检查 | 200 |
| POST | `/api/board/message` | 接收板卡消息 | 200/400 |
| GET | `/api/board/status` | 获取板卡状态 | 200 |
| POST | `/api/chain/create` | 创建任务链 | 201/400 |
| GET | `/api/chain/list` | 列出任务链 | 200 |
| GET | `/api/chain/:id` | 获取任务链详情 | 200/404 |

## 3. 详细接口说明

### 3.1 健康检查

#### GET /health

**功能**: 检查服务健康状态

**响应示例**:
```json
{
  "status": "ok",
  "timestamp": 1713177600
}
```

### 3.2 接收板卡消息

#### POST /api/board/message

**功能**: 接收来自后端的板卡消息

**请求体**:
```json
{
  "message_id": "msg_001",
  "message_time": 1713177600,
  "message": {
    "message_type": "Request",
    "attribute": "Command",
    "connection": "HTTPS",
    "command": "StatusReport",
    "data": {
      "battery": 85.5,
      "latitude": 22.543123,
      "longitude": 114.052345
    }
  },
  "from_id": "backend",
  "from_type": "Server",
  "to_id": "central_001",
  "to_type": "central"
}
```

**成功响应 (200)**:
```json
{
  "success": true,
  "message": "Message received"
}
```

**错误响应 (400)**:
```json
{
  "code": 1,
  "message": "Invalid request",
  "error": "..."
}
```

### 3.3 获取板卡状态

#### GET /api/board/status

**功能**: 获取板卡和无人机状态

**成功响应 (200)**:
```json
{
  "code": 0,
  "data": [
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

### 3.4 创建任务链

#### POST /api/chain/create

**功能**: 创建并启动任务链

**请求体**:
```json
{
  "tasks": [
    {
      "command": "takeoff",
      "data": { "altitude": 10 },
      "timeout": 30
    },
    {
      "command": "goto",
      "data": { "latitude": 22.543123, "longitude": 114.052345, "altitude": 15 },
      "timeout": 60
    },
    {
      "command": "land",
      "data": { "latitude": 22.543123, "longitude": 114.052345, "altitude": 0 },
      "timeout": 30
    }
  ]
}
```

**成功响应 (201)**:
```json
{
  "code": 0,
  "chain_id": "chain_1713177600",
  "status": "pending"
}
```

**错误响应 (400)**:
```json
{
  "code": 1,
  "message": "Invalid request",
  "error": "..."
}
```

### 3.5 列出任务链

#### GET /api/chain/list

**功能**: 列出所有任务链

**成功响应 (200)**:
```json
{
  "code": 0,
  "data": [
    {
      "chain_id": "chain_1713177600",
      "tasks": [
        {
          "task_id": "task_chain_1713177600_0",
          "command": "takeoff",
          "status": "completed",
          "start_time": 1713177600,
          "end_time": 1713177630
        }
      ],
      "current_task": 1,
      "status": "running",
      "start_time": 1713177600,
      "end_time": 0,
      "assigned_drone": ""
    }
  ]
}
```

### 3.6 获取任务链详情

#### GET /api/chain/:id

**功能**: 获取指定任务链的详细信息

**路径参数**:
- `id`: 任务链 ID

**成功响应 (200)**:
```json
{
  "code": 0,
  "data": {
    "chain_id": "chain_1713177600",
    "tasks": [
      {
        "task_id": "task_chain_1713177600_0",
        "command": "takeoff",
        "data": { "altitude": 10 },
        "status": "completed",
        "retry_count": 0,
        "max_retries": 3,
        "timeout": 30,
        "start_time": 1713177600,
        "end_time": 1713177630
      },
      {
        "task_id": "task_chain_1713177600_1",
        "command": "goto",
        "data": { "latitude": 22.543123, "longitude": 114.052345, "altitude": 15 },
        "status": "running",
        "retry_count": 0,
        "max_retries": 3,
        "timeout": 60,
        "start_time": 1713177630,
        "end_time": 0
      }
    ],
    "current_task": 1,
    "status": "running",
    "start_time": 1713177600,
    "end_time": 0,
    "assigned_drone": ""
  }
}
```

**错误响应 (404)**:
```json
{
  "code": 1,
  "message": "Task chain not found",
  "error": "..."
}
```

## 4. 任务命令参考

| 命令 | 描述 | 必需参数 |
|------|------|----------|
| `takeoff` | 起飞 | `altitude` (高度) |
| `land` | 降落 | `latitude`, `longitude` |
| `goto` / `goto_location` | 飞往目标 | `latitude`, `longitude`, `altitude` |
| `return_to_home` / `rtl` | 返航 | - |
| `survey` | 区域侦察 | `latitude`, `longitude`, `radius`, `duration` |
| `survey_grid` | 网格搜索 | `latitude`, `longitude`, `width`, `height`, `altitude` |
| `orbit` | 盘旋巡逻 | `latitude`, `longitude`, `radius`, `duration` |
| `take_photo` | 拍照 | - |
| `start_video` | 开始录像 | - |
| `stop_video` | 停止录像 | - |
| `set_mode` | 设置模式 | `mode` |

## 5. 状态码

### 5.1 HTTP 状态码

| 状态码 | 说明 |
|--------|------|
| 200 | OK - 请求成功 |
| 201 | Created - 资源创建成功 |
| 400 | Bad Request - 请求参数错误 |
| 404 | Not Found - 资源不存在 |
| 500 | Internal Server Error - 服务器内部错误 |

### 5.2 业务错误码

| 错误码 | 说明 |
|--------|------|
| 0 | 成功 |
| 1 | 通用错误 |
| 1001 | 任务链不存在 |
| 1002 | 任务链执行失败 |
| 2001 | MAVLink 连接失败 |
| 3001 | 后端通信失败 |

## 6. 与后端通信

### 6.1 后端 API 地址

- **地址**: `https://api.deeppluse.dpdns.org`
- **核心接口**: `POST /api/board/send-message`
- **认证**: JWT Token (Bearer Token)

### 6.2 消息格式

```json
{
  "message_id": "msg_001",
  "message_time": 1713177600,
  "message": {
    "message_type": "Request",
    "attribute": "Command",
    "connection": "HTTPS",
    "command": "StatusReport",
    "data": {...}
  },
  "from_id": "central_001",
  "from_type": "central",
  "to_id": "backend",
  "to_type": "Server"
}
```

## 7. 无人机搜索配置参数

### 7.1 配置位置

所有参数在 `config.yaml` 的 `drone.search` 部分配置：

```yaml
drone:
  search:
    min_battery_level: 20.0      # 最小电池电量阈值（%）
    max_drone_distance: 1000.0   # 最大无人机距离阈值（米）
    status_check_timeout: 5      # 状态检查超时时间（秒）
    status_update_interval: 2    # 状态更新间隔（秒）
    message_chan_size: 1000      # 消息通道大小
    score_weight: 10.0           # 评分权重
```

### 7.2 参数说明

| 参数 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `min_battery_level` | float64 | 20.0 | 最小电池电量百分比，低于此值的无人机将被排除 |
| `max_drone_distance` | float64 | 1000.0 | 最大无人机距离（米），超过此距离的无人机将被排除 |
| `status_check_timeout` | int | 5 | 状态检查超时时间（秒），超过此时间未更新状态的无人机将被标记为不可用 |
| `status_update_interval` | int | 2 | 状态更新检查间隔（秒） |
| `message_chan_size` | int | 1000 | 无人机消息通道的缓冲区大小 |
| `score_weight` | float64 | 10.0 | 计算最佳无人机时的电池电量权重系数 |

### 7.3 无人机选择算法

最佳无人机选择算法：

1. 筛选出 `is_idle == true` 的无人机
2. 排除电池电量低于 `min_battery_level` 的无人机
3. 排除距离超过 `max_drone_distance` 的无人机
4. 计算得分：`score = battery_level * score_weight`
5. 选择得分最高的无人机

## 8. 示例请求

### 8.1 创建任务链

```bash
curl -X POST http://localhost:8084/api/chain/create \
  -H "Content-Type: application/json" \
  -d '{
    "tasks": [
      {
        "command": "takeoff",
        "data": { "altitude": 10 },
        "timeout": 30
      },
      {
        "command": "land",
        "data": { "latitude": 22.543123, "longitude": 114.052345, "altitude": 0 },
        "timeout": 30
      }
    ]
  }'
```

### 8.2 获取任务链状态

```bash
curl http://localhost:8084/api/chain/list
```

### 8.3 获取板卡状态

```bash
curl http://localhost:8084/api/board/status
```

### 8.4 健康检查

```bash
curl http://localhost:8084/health
```

## 9. 注意事项

1. **安全**: 生产环境必须使用 HTTPS
2. **认证**: 与后端通信需要有效的 JWT Token
3. **性能**: 任务链处理是异步的，不会阻塞 API 响应
4. **错误处理**: 详细的错误信息会在响应中返回
5. **日志**: 所有操作都会记录详细日志
6. **配置**: 无人机搜索参数可通过 config.yaml 灵活配置，无需修改代码

## 10. 版本信息

- **API 版本**: v1.1.0
- **最后更新**: 2026-04-16
