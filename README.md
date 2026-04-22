# 无人机平台前端项目

## 项目简介

这是一个基于 Vue 3 + TypeScript + Vite 开发的无人机平台前端项目，包含任务链管理、板子管理、MAVLink 控制、实时监控等功能。

## 技术栈

- **前端框架**：Vue 3 + TypeScript + Vite
- **状态管理**：Pinia
- **路由管理**：Vue Router
- **HTTP 客户端**：Axios
- **UI 组件库**：Element Plus
- **CSS 预处理器**：SCSS

## 项目结构

```
drone-frontend/
├── public/           # 静态资源
├── src/
│   ├── api/          # API 接口封装
│   ├── components/   # 组件
│   ├── router/       # 路由配置
│   ├── stores/       # Pinia 状态管理
│   ├── styles/       # 全局样式
│   ├── types/        # TypeScript 类型定义
│   ├── utils/        # 工具函数
│   ├── views/        # 页面组件
│   └── main.ts       # 应用入口
├── .env.development  # 开发环境配置
├── .env.production   # 生产环境配置
├── index.html        # HTML 模板
├── package.json      # 项目配置
├── tsconfig.json     # TypeScript 配置
├── vite.config.ts    # Vite 配置
└── README.md         # 项目说明
```

## 快速开始

### 安装依赖

```bash
npm install
```

### 开发模式运行

```bash
npm run dev
```

### 构建生产版本

```bash
npm run build
```

### 预览生产构建

```bash
npm run preview
```

## 功能说明

### 1. 认证系统
- 登录/注册功能
- JWT Token 认证
- 路由守卫

### 2. 任务链管理
- 创建/编辑/删除任务链
- 添加任务节点（起飞、移动、拍照、降落、悬停等）
- 节点参数配置
- 任务链执行控制

### 3. 板子管理
- 板子列表展示
- 创建/删除板子
- 向板子发送指令
- 板子连接状态监控

### 4. MAVLink 控制
- V1 控制：起飞、降落、移动、返航
- V2 控制：一键起飞、一键降落
- 参数配置

### 5. 实时监控
- 任务链状态轮询
- 错误日志展示
- 日志管理

### 6. 终端功能
- 全局终端（按 / 键呼出）
- 支持多种命令
- 命令执行结果展示

### 7. 设置管理
- 连接设置
- 日志设置
- 安全设置

## 联调说明

项目支持两种模式：

### Mock 模式
- 默认模式，不依赖后端服务
- 所有功能可独立演示

### 联调模式
- 在设置页面开启「真实后端联调模式」
- 需后端服务正常运行
- 调用真实后端 API

## 代理配置

在 `vite.config.ts` 中配置了代理，可根据实际后端地址进行修改：

```typescript
proxy: {
  '/users': {
    target: 'https://api.deeppluse.dpdns.org',
    changeOrigin: true,
    secure: false
  },
  '/api': {
    target: 'https://api.deeppluse.dpdns.org',
    changeOrigin: true,
    secure: false
  },
  '/terminal': {
    target: 'https://api.deeppluse.dpdns.org',
    changeOrigin: true,
    secure: false
  }
}
```

## 注意事项

1. 开发环境下使用 Mock 数据，生产环境下需要后端服务支持
2. 登录/注册功能需要后端服务正常运行
3. 部分功能可能需要根据实际硬件设备进行调整