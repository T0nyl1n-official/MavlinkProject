# Central Board 中央控制板

## 项目简介

Central Board 是 MavlinkProject 的中央控制板，负责：
- 接收和处理来自后端的任务链
- 将任务转译为 MAVLink 命令并发送给飞控
- 与后端 API 进行安全的 HTTPS 通信
- 提供 RESTful API 接口
- 管理和调度无人机资源

## 技术栈

| 技术 | 版本 | 用途 |
|------|------|------|
| Go | 1.24.0 | 主开发语言 |
| Gin | 1.9.1 | HTTP/HTTPS 框架 |
| gomavlib | v3.3.0 | MAVLink 通信 |
| go-resty | v2.10.0 | HTTP 客户端 |
| yaml.v3 | 3.0.1 | 配置管理 |
| acme/autocert | 最新 | Let's Encrypt 证书 |

## 项目结构

```
CentralBoard/
├── app/                 # 应用代码
│   ├── api/             # API 接口
│   │   ├── handlers/    # 请求处理器
│   │   └── routes.go    # 路由注册
│   ├── services/        # 业务服务
│   │   ├── backend/     # 后端通信
│   │   ├── mavlink/     # MAVLink 服务
│   │   └── task/        # 任务链管理
│   ├── config/          # 配置管理
│   └── utils/           # 工具函数
├── MavlinkCommand/      # MAVLink 命令模块
├── Distribute/          # 无人机搜索模块
├── Handler/             # 处理器模块
├── Shared/              # 共享数据结构
├── config.yaml          # 主配置文件
├── Central.go           # 主入口
├── CentralServer.exe    # 编译产物
├── docs/                # 文档
│   ├── README.md        # 项目说明
│   ├── API.md           # API 接口文档
│   └── CONFIG.md        # 配置说明
└── tests/               # 测试程序
    ├── integration/     # 集成测试
    └── unit/            # 单元测试
```

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 配置文件

复制并编辑配置文件：

```bash
cp config.yaml config.yaml.local
# 编辑配置文件
```

关键配置项：
- `server.domain`: 域名（用于 Let's Encrypt 证书）
- `server.port`: 服务器端口
- `mavlink.serial_port`: 飞控串口
- `backend.token`: 后端认证令牌
- `drone.search.*`: 无人机搜索相关阈值

### 3. 编译

```bash
go build -o CentralServer.exe .
```

### 4. 运行

```bash
# 使用默认配置
./CentralServer.exe

# 使用自定义配置
./CentralServer.exe config.yaml.local
```

## 功能特性

### 🔒 HTTPS 服务
- 支持 Let's Encrypt 自动证书
- 支持手动证书配置
- 安全的 TLS 通信

### 📡 后端通信
- 与 `https://api.deeppluse.dpdns.org` 通信
- 严格按照 API 文档规范
- 设备认证（JWT Token）

### 🎯 任务链处理
- 解析和执行任务链
- 任务状态跟踪
- 任务重试机制

### 🚁 MAVLink 集成
- 支持常见 MAVLink 命令：
  - `takeoff` - 起飞
  - `land` - 降落
  - `goto` - 飞往目标
  - `return_to_home` - 返航
  - `survey` - 区域侦察
  - `orbit` - 盘旋巡逻
  - 等其他命令

### 🔍 无人机搜索
- 基于电池电量和可用性的无人机选择
- 动态阈值配置（最小电量、最大距离等）
- 任务分配和状态管理

### ⚙️ 灵活配置
- 所有阈值和参数均可通过 config.yaml 配置
- 支持无人机搜索参数自定义：
  - 最小电池电量阈值
  - 最大无人机距离阈值
  - 状态检查超时时间
  - 状态更新间隔
  - 消息通道大小
  - 评分权重

### 📊 API 接口
- `/health` - 健康检查
- `/api/board/message` - 接收任务链
- `/api/board/status` - 获取板卡状态
- `/api/chain/*` - 任务链管理

## 配置详解

### 无人机搜索配置 (drone.search)

所有参数都在 `config.yaml` 中的 `drone.search` 部分配置：

```yaml
drone:
  search:
    # 最小电池电量阈值（百分比）
    # 低于此电量的无人机将被排除
    min_battery_level: 20.0

    # 最大无人机距离阈值（米）
    # 超过此距离的无人机将被排除
    max_drone_distance: 1000.0

    # 状态检查超时时间（秒）
    # 超过此时间未更新的无人机将被标记为不可用
    status_check_timeout: 5

    # 状态更新间隔（秒）
    # 定期检查无人机状态的间隔
    status_update_interval: 2

    # 消息通道大小
    # 无人机消息通道的缓冲区大小
    message_chan_size: 1000

    # 评分权重
    # 计算最佳无人机时的电池电量权重
    score_weight: 10.0
```

## 测试

### 运行单元测试

```bash
go test -v ./tests/unit/
```

### 运行集成测试

```bash
go test -v ./tests/integration/
```

## 部署

### 生产环境

1. 配置域名 `central.deeppluse.dpdns.org` 解析到服务器
2. 确保 80 端口开放（用于 Let's Encrypt 验证）
3. 确保 8084 端口开放（用于 HTTPS 服务）
4. 配置 `config.yaml` 中的 `server.domain` 和 `server.email`
5. 启动服务：
   ```bash
   nohup ./CentralServer.exe > central.log 2>&1 &
   ```

### 开发环境

```bash
# 不使用 TLS 运行
# 在 config.yaml 中设置 server.domain=""
./CentralServer.exe
```

## 监控与日志

- 服务启动日志：控制台输出
- 运行日志：标准输出/错误
- 建议使用 systemd 或 supervisor 管理服务

## 故障排查

### 常见问题

1. **证书申请失败**：
   - 检查域名解析是否正确
   - 确保 80 端口可访问
   - 查看 Let's Encrypt 错误日志

2. **MAVLink 连接失败**：
   - 检查串口配置
   - 确保飞控电源开启
   - 查看 MAVLink 初始化日志

3. **后端通信失败**：
   - 检查网络连接
   - 验证认证令牌
   - 查看后端 API 状态

4. **无人机搜索问题**：
   - 检查 `drone.search` 配置是否正确
   - 确保无人机状态正常上报
   - 查看无人机超时设置是否合理

## 版本历史

### v1.1.0
- ✅ 无人机搜索模块参数配置化
- ✅ 所有阈值和参数可通过 config.yaml 配置
- ✅ 文档更新

### v1.0.0
- ✅ HTTPS 服务配置
- ✅ 路由分离
- ✅ 后端 API 集成
- ✅ 任务链处理
- ✅ MAVLink 集成
- ✅ 测试程序
- ✅ 文档生成

## 贡献

欢迎提交 Issue 和 Pull Request！

## 许可证

MIT License
