# MavlinkProject 技术文档

## 1. 设计重点难点与解决方案

### 1.1 导入循环问题的解决

**问题描述**：在开发过程中，我们遇到了 Terminal 模块与 Backend 模块之间的构建问题，导致无法直接干预系统操作。

**解决方案**：

- **接口抽象**：创建了 `BackendOperations` 接口，定义了重启和关闭服务器的方法
- **依赖注入**：通过接口注入的方式，将 Backend 服务器实例传递给 Terminal 模块
- **解耦设计**：消除了 Terminal 模块对 Backend 模块的直接依赖，使用接口进行通信

**代码实现**：
```go
// TerminalRoutes.go
type BackendOperations interface {
    Restart()
    Shutdown() error
}

// Backend.go
func (bs *BackendServer) Run(port string, httpsConfig HTTPSConfig) {
    Routes.InitAllRoutes(bs.Router, bs.JWTManager, bs.TokenManager, bs.Mysql, bs.SettingManager, bs)
}
```

### 1.2 设备认证系统的设计

**问题描述**：需要为不同类型的硬件设备（Central、LandNode、Sensor、Drone）设计独立的认证系统，同时确保安全性和友好的错误提示。

**解决方案**：

- **独立认证系统**：为设备创建了专用的登录/登出接口和 JWT 中间件
- **设备状态管理**：实现了设备在线/离线状态的跟踪
- **友好错误提示**：为设备认证失败提供清晰的错误信息，不触发封禁
- **Token 管理**：使用 Redis 存储设备 Token，支持 Token 失效机制

**技术亮点**：
- 设备认证与用户认证分离，提高安全性
- 专用的设备 JWT 中间件，提供设备特定的认证逻辑
- 设备状态实时更新，便于监控和管理

### 1.3 多中心通信的实现

**问题描述**：需要实现与多个 Central 服务器的可靠通信，确保任务链能够**及时送达，准确送达**。

**解决方案**：

- **HTTPS 通信**：实现了基于 HTTPS 的 Central 通信客户端
- **自动重试机制**：通信失败时自动重试，确保消息可靠送达
- **故障转移**：当一个 Central 不可用时，自动尝试其他 Central 服务器
- **超时控制**：为每个请求设置合理的超时时间，避免阻塞

**技术实现**：
- 创建了 `CentralHTTPClient` 结构体，封装了 HTTPS 通信逻辑
- 实现了 `PushMessageToCentralHTTP` 方法，支持重试和故障转移
- 使用 `tls.Config{InsecureSkipVerify: true}` 处理自签名证书

### 1.4 传感器警报处理和任务链生成

**问题描述**：需要根据传感器警报自动生成合适的任务链，并选择最佳无人机执行任务，需要内置调度算法来精准完成对应问题。

**解决方案**：

- **警报类型识别**：根据警报类型（fire、rescue、patrol 等）生成不同的任务链
- **无人机选择**：基于无人机状态、电量和距离选择最佳执行任务的无人机
- **任务链动态生成**：根据警报类型和位置动态生成任务链
- **并行处理**：使用并行方式检查无人机状态，提高响应速度

**技术亮点**：
- 采用策略模式处理不同类型的警报
- 实现了基于多因素的无人机选择算法
- 任务链生成逻辑与传感器处理解耦，便于扩展

### 1.5 性能优化措施

**问题描述**：系统需要处理大量设备连接和传感器警报，要求**高性能、低延迟、主打高性价比**。

**解决方案**：
- **并发处理**：使用 Go 的 goroutine 处理并发任务
- **连接池**：实现了设备连接池，减少连接建立的开销
- **缓存优化**：使用 Redis 缓存热点数据，减少数据库查询
- **消息队列**：使用 channel 作为消息队列，实现异步处理

**技术实现**：
- 使用 `sync.WaitGroup` 协调并发任务
- 实现了 Redis 缓存层，缓存设备状态和配置信息
- 使用 buffered channel 处理消息队列，提高吞吐量

### 1.6 安全性优化

**问题描述**：系统需要处理敏感的控制指令和认证信息，**系统安全性和用户的安全对于我们是至关重要的**。

**解决方案**：

- **HTTPS 加密**：所有外部通信采用 HTTPS 加密
- **JWT 安全**：使用安全的 JWT 签名算法
- **密码加密**：用户密码采用安全的哈希算法存储
- **权限控制**：实现了基于角色的权限控制

**技术实现**：
- 使用 `golang-jwt/jwt/v4` 库实现 JWT 认证
- 密码存储使用 bcrypt 哈希算法
- 实现了基于角色的中间件，控制接口访问权限

## 2. 开源组件说明

### 2.1 核心框架

| 组件名称 | 版本 | 用途 | 来源 |
|----------|------|------|------|
| Gin | v1.12.0 | Web 框架 | github.com/gin-gonic/gin |
| GORM | v1.6.0 | ORM 框架 | github.com/jinzhu/gorm |
| MySQL 驱动 | v1.8.1 | 数据库驱动 | github.com/go-sql-driver/mysql |
| Redis 客户端 | v9 | 缓存客户端 | github.com/redis/go-redis/v9 |
| MAVLink 库 | v3.3.0 | MAVLink 协议实现 | github.com/bluenviron/gomavlib/v3 |
| JWT 库 | v4.5.2 | JWT 认证 | github.com/golang-jwt/jwt/v4 |
| YAML 库 | v3.0.1 | 配置文件解析 | gopkg.in/yaml.v3 |

### 2.2 辅助库

| 组件名称 | 版本 | 用途 | 来源 |
|----------|------|------|------|
| validator | v10.30.1 | 数据验证 | github.com/go-playground/validator/v10 |
| cors | v1.7.6 | CORS 中间件 | github.com/gin-contrib/cors |
| sonic | v1.15.0 | JSON 解析 | github.com/bytedance/sonic |
| go-toml | v2.2.4 | TOML 解析 | github.com/pelletier/go-toml/v2 |
| go-json | v0.10.5 | JSON 解析 | github.com/goccy/go-json |
| backoff | v5.0.3 | 重试机制 | github.com/cenkalti/backoff/v5 |

### 2.3 测试库

| 组件名称 | 版本 | 用途 | 来源 |
|----------|------|------|------|
| ginkgo | v2.12.0 | BDD 测试框架 | github.com/bsm/ginkgo/v2 |
| gomega | v1.27.10 | 测试断言库 | github.com/bsm/gomega |
| go-spew | v1.1.2 | 测试输出 | github.com/davecgh/go-spew |
| go-difflib | v1.0.1 | 测试差异比较 | github.com/pmezard/go-difflib |

### 2.4 安全库

| 组件名称 | 版本 | 用途 | 来源 |
|----------|------|------|------|
| go-jose | v4.1.3 | JWT 实现 | github.com/go-jose/go-jose/v4 |
| lego | v4.33.0 | Let's Encrypt 客户端 | github.com/go-acme/lego/v4 |
| edwards25519 | v1.1.0 | 加密算法 | filippo.io/edwards25519 |

## 3. 技术架构总结

MavlinkProject 采用了现代化的 Go 语言架构，具有以下特点：

1. **模块化设计**：系统分为用户认证、设备认证、Board 通信、传感器处理、FRP 通信等模块，每个模块职责明确，便于维护和扩展。

2. **接口抽象**：通过接口抽象解决了模块间的依赖问题，提高了系统的灵活性和可测试性。

3. **并发处理**：充分利用 Go 语言的并发特性，提高系统的处理能力和响应速度。

4. **安全可靠**：采用 HTTPS 加密、JWT 认证、密码哈希等安全措施，确保系统的安全性。

5. **高性能**：通过连接池、缓存优化、并行处理等技术，提高系统的性能和响应速度。

6. **可扩展性**：支持多 Central 服务器配置，便于系统的横向扩展。

7. **可靠性**：实现了自动重试、故障转移等机制，提高系统的可靠性和可用性。

通过这些技术措施，MavlinkProject 实现了一个功能完整、安全可靠、性能优异的无人机调度管理系统，能够满足各种复杂场景的需求。