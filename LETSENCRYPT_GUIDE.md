# Let's Encrypt 配置指南

## 配置步骤

### 1. 修改配置文件
编辑 `config/Server_Config.yaml`，将以下占位符替换为你的实际信息：

```yaml
# 后端服务器配置
backend:
  # 后端 API 地址
  address: "localhost"
  port: "8080"
  
  # Let's Encrypt 配置
  lets_encrypt:
    email: "your-email@your-domain.com"  # 替换为你的邮箱
    domains:
      - "your-domain.com"               # 替换为你的域名
    webroot: "./webroot"
    use_staging: false
```

### 2. 确保域名解析正确
- 你的域名必须指向服务器的公网 IP
- 确保 80 和 443 端口对外开放

### 3. 创建 webroot 目录
```bash
mkdir -p webroot
```

### 4. 启动服务器
```bash
cd Server
./Server.exe
```

## 注意事项

### 开发环境
- 如果使用 `localhost` 或没有域名，系统会自动使用自签名证书
- 浏览器会显示"不安全"警告，这是正常的

### 生产环境
- 必须使用真实的域名
- Let's Encrypt 证书会被所有浏览器信任
- 证书每 90 天自动续期

### 测试环境
- 可以设置 `use_staging: true` 使用 Let's Encrypt 测试环境
- 测试环境证书不受信任，但不会消耗生产环境配额

## 证书存储位置
- Let's Encrypt 证书：`./letsencrypt/cert.pem` 和 `./letsencrypt/key.pem`
- 自签名证书：`cert.pem` 和 `key.pem`

## 故障排除

### 证书获取失败
1. 检查域名解析是否正确
2. 确保 80 端口可以访问
3. 检查防火墙设置

### HTTPS 连接问题
1. 检查证书文件是否存在
2. 验证证书权限
3. 检查 TLS 配置

### 浏览器警告
- 开发环境：使用 `https://localhost:8080` 访问
- 生产环境：确保证书配置正确