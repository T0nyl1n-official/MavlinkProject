# 深地哨兵前端项目

## 项目简介

深地哨兵是一个用于地下空间监测和控制的无人机系统前端项目，基于 Vue 3 + TypeScript + Element Plus 开发。

## 技术栈

- **框架**: Vue 3 + TypeScript
- **UI 库**: Element Plus
- **状态管理**: Pinia
- **路由**: Vue Router
- **HTTP 客户端**: Axios
- **构建工具**: Vite

## 项目结构

```
src/
├── api/          # API 相关代码
├── components/   # 公共组件
├── composables/  # 组合式 API
├── router/       # 路由配置
├── stores/       # 状态管理
├── utils/        # 工具函数
├── views/        # 页面组件
├── App.vue       # 根组件
└── main.ts       # 入口文件
```

## 核心功能

### 1. 仪表盘
- 实时数据统计
- 任务链状态监控
- 板子连接状态

### 2. 任务链管理
- 创建和编辑任务链
- 任务链执行控制（启动、暂停、停止）
- 任务链详情查看

### 3. 板子管理
- 板子列表展示
- 板子连接状态监控
- 板子配置管理

### 4. 全局终端
- UDP 命令发送
- 命令历史记录
- 快捷命令按钮

### 5. 错误日志
- 错误日志展示
- 日志清空功能
- 日志导出

## 工具函数

### 错误处理
- `errorHandler.ts` - API 错误处理工具

### 表单验证
- `formValidator.ts` - 表单验证工具

### 数据导出
- `exportUtils.ts` - 数据导出工具，支持 JSON 和 CSV 格式

### 性能优化
- `performance.ts` - 性能优化工具，包括防抖、节流、懒加载等

### 主题管理
- `themeManager.ts` - 主题管理工具，支持深色/浅色主题切换

### WebSocket
- `useWebSocket.ts` - WebSocket 连接管理工具

## 组件

### 布局组件
- `layout/Sidebar.vue` - 侧边栏组件
- `layout/Header.vue` - 头部组件

### 功能组件
- `ErrorBoundary.vue` - 全局错误边界组件
- `Skeleton.vue` - 骨架屏组件
- `GlobalTerminal.vue` - 全局终端组件
- `ErrorLog.vue` - 错误日志组件

## 开发指南

### 安装依赖

```bash
npm install
```

### 开发模式

```bash
npm run dev
```

### 构建生产版本

```bash
npm run build
```

### 代码检查

```bash
npm run lint
```

## 浏览器兼容性

- Chrome (最新版本)
- Firefox (最新版本)
- Safari (最新版本)
- Edge (最新版本)

## 注意事项

1. 后端服务暂时不可用，前端使用模拟数据
2. WebSocket 连接需要后端支持
3. 部分功能需要在实际环境中测试

## 许可证

MIT
