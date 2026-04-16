# 配置说明

## 配置文件

Central Board 使用 YAML 格式的配置文件，默认文件名为 `config.yaml`。

## 配置结构

```yaml
server:
  port: "8084"
  domain: "central.deeppluse.dpdns.org"
  email: "your-email@example.com"
  cert_file: ""
  key_file: ""

mavlink:
  serial_port: "/dev/ttyUSB0"
  serial_baud: 115200
  target_system: 1

backend:
  address: "https://api.deeppluse.dpdns.org"
  token: "your-device-token"
  device_id: "central_001"
  device_type: "central"
```

## 详细配置项

### 1. server 配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `port` | string | "8084" | 服务器监听端口 |
| `domain` | string | "" | 域名（用于 Let's Encrypt 证书） |
| `email` | string | "" | Let's Encrypt 通知邮箱 |
| `cert_file` | string | "" | 手动证书路径（PEM 格式） |
| `key_file` | string | "" | 手动密钥路径（PEM 格式） |

### 2. mavlink 配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `serial_port` | string | "/dev/ttyUSB0" | 飞控串口路径 |
| `serial_baud` | int | 115200 | 串口波特率 |
| `target_system` | uint8 | 1 | 目标飞控系统 ID |

### 3. backend 配置

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `address` | string | "https://api.deeppluse.dpdns.org" | 后端 API 地址 |
| `token` | string | "" | 设备认证 Token |
| `device_id` | string | "central_001" | 设备 ID |
| `device_type` | string | "central" | 设备类型 |

## 配置模式

### 模式 1: Let's Encrypt 自动证书

```yaml
server:
  domain: "central.deeppluse.dpdns.org"
  email: "your-email@example.com"
  # 不需要设置 cert_file 和 key_file
```

**要求**:
- 域名已解析到服务器 IP
- 80 端口可访问（用于 ACME 验证）
- 443/8084 端口可访问（用于 HTTPS）

### 模式 2: 手动证书

```yaml
server:
  cert_file: "/path/to/cert.pem"
  key_file: "/path/to/key.pem"
  # 不需要设置 domain
```

**要求**:
- 证书文件必须是 PEM 格式
- 证书必须包含完整的证书链
- 密钥文件必须是未加密的

### 模式 3: 无 TLS（仅开发环境）

```yaml
server:
  # 不设置 domain、cert_file、key_file
```

**警告**: 仅用于开发测试，生产环境必须使用 TLS

## 环境变量

配置文件中的所有选项都可以通过环境变量覆盖：

| 环境变量 | 对应配置项 | 示例 |
|----------|------------|------|
| `CENTRAL_SERVER_PORT` | `server.port` | `8084` |
| `CENTRAL_SERVER_DOMAIN` | `server.domain` | `central.deeppluse.dpdns.org` |
| `CENTRAL_SERVER_EMAIL` | `server.email` | `your-email@example.com` |
| `CENTRAL_SERVER_CERT_FILE` | `server.cert_file` | `/path/to/cert.pem` |
| `CENTRAL_SERVER_KEY_FILE` | `server.key_file` | `/path/to/key.pem` |
| `CENTRAL_MAVLINK_SERIAL_PORT` | `mavlink.serial_port` | `/dev/ttyUSB0` |
| `CENTRAL_MAVLINK_SERIAL_BAUD` | `mavlink.serial_baud` | `115200` |
| `CENTRAL_MAVLINK_TARGET_SYSTEM` | `mavlink.target_system` | `1` |
| `CENTRAL_BACKEND_ADDRESS` | `backend.address` | `https://api.deeppluse.dpdns.org` |
| `CENTRAL_BACKEND_TOKEN` | `backend.token` | `your-device-token` |
| `CENTRAL_BACKEND_DEVICE_ID` | `backend.device_id` | `central_001` |
| `CENTRAL_BACKEND_DEVICE_TYPE` | `backend.device_type` | `central` |

## 配置验证

启动时会自动验证配置的有效性：

- 端口必须是有效的数字
- 域名必须是有效的 FQDN
- 证书文件必须存在且有效
- 串口配置必须合理

## 示例配置

### 生产环境（Let's Encrypt）

```yaml
server:
  port: "8084"
  domain: "central.deeppluse.dpdns.org"
  email: "admin@deeppluse.dpdns.org"

mavlink:
  serial_port: "/dev/ttyUSB0"
  serial_baud: 115200
  target_system: 1

backend:
  address: "https://api.deeppluse.dpdns.org"
  token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  device_id: "central_001"
  device_type: "central"
```

### 开发环境（无 TLS）

```yaml
server:
  port: "8084"

mavlink:
  serial_port: "COM3"  # Windows 串口
  serial_baud: 115200
  target_system: 1

backend:
  address: "https://api.deeppluse.dpdns.org"
  token: "test-token"
  device_id: "central_dev"
  device_type: "central"
```

### 测试环境（手动证书）

```yaml
server:
  port: "8084"
  cert_file: "./certs/cert.pem"
  key_file: "./certs/key.pem"

mavlink:
  serial_port: "/dev/ttyUSB0"
  serial_baud: 115200
  target_system: 1

backend:
  address: "https://api.deeppluse.dpdns.org"
  token: "test-token"
  device_id: "central_test"
  device_type: "central"
```

## 配置热重载

服务启动后不支持配置热重载，修改配置后需要重启服务。

## 故障排查

### 常见配置问题

1. **证书申请失败**:
   - 检查域名解析是否正确
   - 确保 80 端口可访问
   - 查看 Let's Encrypt 错误日志

2. **MAVLink 连接失败**:
   - 检查串口路径是否正确
   - 确保飞控电源开启
   - 验证波特率设置

3. **后端通信失败**:
   - 检查网络连接
   - 验证认证令牌
   - 确认后端服务状态

4. **端口占用**:
   - 检查是否有其他服务占用 8084 端口
   - 使用 `netstat -ano | findstr :8084` 查看

## 最佳实践

1. **生产环境**:
   - 使用 Let's Encrypt 自动证书
   - 定期更新配置文件中的令牌
   - 备份配置文件

2. **开发环境**:
   - 使用独立的配置文件
   - 禁用 TLS 加速开发
   - 使用测试令牌

3. **安全性**:
   - 不要在配置文件中硬编码敏感信息
   - 使用环境变量存储令牌
   - 定期轮换认证令牌
