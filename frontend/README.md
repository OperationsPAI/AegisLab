# RCABench Frontend

RCABench (AegisLab) 微服务根因分析基准测试平台的前端应用。

## 技术栈

- **框架**: React 18 + TypeScript
- **构建工具**: Vite 5
- **UI 组件库**: Ant Design 5
- **状态管理**: Zustand
- **数据获取**: TanStack Query (React Query)
- **HTTP 客户端**: Axios
- **路由**: React Router v6
- **图表库**: ECharts
- **代码编辑器**: Monaco Editor
- **样式**: CSS + Ant Design 主题定制

## 设计风格

采用**学术研究风格**:
- 中性色调,以深蓝色 (#2563eb) 为主色调
- 清晰的信息层次和数据展示
- 强调图表和分析结果的可读性
- 简洁专业的界面设计

## 项目结构

```
frontend/
├── src/
│   ├── api/                  # API 客户端
│   │   ├── client.ts        # Axios 配置和拦截器
│   │   ├── auth.ts          # 认证相关 API
│   │   ├── projects.ts      # 项目管理 API
│   │   ├── containers.ts    # 容器管理 API
│   │   ├── injections.ts    # 故障注入 API
│   │   ├── executions.ts    # 算法执行 API
│   │   ├── tasks.ts         # 任务管理 API
│   │   └── evaluations.ts   # 评估 API
│   ├── assets/              # 静态资源
│   ├── components/          # React 组件
│   │   ├── common/          # 通用组件
│   │   └── layout/          # 布局组件
│   │       └── MainLayout.tsx
│   ├── hooks/               # 自定义 Hooks
│   ├── pages/               # 页面组件
│   │   ├── auth/            # 认证页面
│   │   │   └── Login.tsx
│   │   ├── dashboard/       # 仪表盘
│   │   │   └── Dashboard.tsx
│   │   ├── projects/        # 项目管理
│   │   │   └── ProjectList.tsx
│   │   ├── containers/      # 容器管理
│   │   │   └── ContainerList.tsx
│   │   ├── datasets/        # 数据集管理
│   │   ├── injections/      # 故障注入
│   │   │   ├── InjectionList.tsx
│   │   │   └── InjectionCreate.tsx
│   │   ├── executions/      # 算法执行
│   │   │   └── ExecutionList.tsx
│   │   ├── evaluations/     # 评估
│   │   ├── tasks/           # 任务监控
│   │   ├── system/          # 系统管理
│   │   └── settings/        # 设置
│   ├── store/               # 状态管理
│   │   └── auth.ts         # 认证状态
│   ├── types/               # TypeScript 类型定义
│   │   └── api.ts          # API 类型定义
│   ├── utils/               # 工具函数
│   │   └── theme.ts        # 主题配置
│   ├── App.tsx             # 根组件
│   ├── main.tsx            # 应用入口
│   └── index.css           # 全局样式
├── index.html              # HTML 模板
├── package.json            # 依赖配置
├── tsconfig.json           # TypeScript 配置
├── vite.config.ts          # Vite 配置
└── README.md               # 项目说明
```

## 快速开始

### 前置要求

- Node.js >= 18
- npm >= 9

### 安装依赖

```bash
cd frontend
npm install
```

### 开发模式

```bash
npm run dev
```

应用将在 http://localhost:3000 启动,并代理 API 请求到 http://localhost:8082

### 构建生产版本

```bash
npm run build
```

构建产物将生成在 `dist` 目录。

### 预览生产构建

```bash
npm run preview
```

### 代码检查

```bash
npm run lint
```

### 类型检查

```bash
npm run type-check
```

## 核心功能

### 已实现

✅ **认证系统**
- 登录页面
- JWT Token 认证
- Token 自动刷新
- 认证状态管理

✅ **主布局**
- 固定式 Header 和 Sidebar
- 响应式设计
- 用户信息下拉菜单

✅ **仪表盘**
- 关键指标展示(项目、实验、任务、执行)
- 任务状态分布饼图
- 最近活动列表

✅ **项目管理**
- 项目列表(分页、搜索、筛选)
- 创建/编辑/删除项目
- 标签管理

✅ **容器管理**
- 容器列表(分页、搜索、类型筛选)
- 支持 Pedestal/Benchmark/Algorithm 三种类型
- 版本管理

### 待实现

⏳ **故障注入**
- 可视化故障编排器(批次 + 并行节点)
- 故障配置表单(动态表单根据故障类型)
- 注入详情页(实时状态、日志流)

⏳ **算法执行**
- 执行列表和详情
- 结果可视化(服务拓扑图、分层结果)
- 准确率分析(Top-K、混淆矩阵)

⏳ **评估功能**
- Datapack 评估
- Dataset 评估
- 算法对比(雷达图、柱状图)

⏳ **任务监控**
- 任务列表和详情
- 实时日志流(SSE)
- 任务依赖树可视化

⏳ **系统管理**
- 用户管理
- 角色管理
- 权限管理
- 标签管理

## API 代理配置

开发模式下,Vite 会将 `/api` 请求代理到后端服务器:

```typescript
// vite.config.ts
server: {
  port: 3000,
  proxy: {
    '/api': {
      target: 'http://localhost:8082',
      changeOrigin: true,
    },
  },
}
```

## 环境变量

创建 `.env.local` 文件来配置环境变量:

```bash
# API Base URL (默认使用代理)
VITE_API_BASE_URL=/api/v2

# 其他配置...
```

## 主题定制

主题配置在 `src/utils/theme.ts` 和 `src/main.tsx` 中:

```typescript
// 主色调
colorPrimary: '#2563eb'  // 深蓝色
colorSuccess: '#10b981'  // 绿色
colorWarning: '#f59e0b'  // 琥珀色
colorError: '#ef4444'    // 红色
colorInfo: '#06b6d4'     // 青色
```

## 状态管理

使用 Zustand 进行轻量级状态管理,主要用于认证状态:

```typescript
// 使用示例
import { useAuthStore } from '@/store/auth'

const { user, login, logout, isAuthenticated } = useAuthStore()
```

## 数据获取

使用 TanStack Query (React Query) 进行服务端状态管理:

```typescript
// 使用示例
import { useQuery } from '@tanstack/react-query'
import { projectApi } from '@/api/projects'

const { data, isLoading } = useQuery({
  queryKey: ['projects', { page: 1, size: 10 }],
  queryFn: () => projectApi.getProjects({ page: 1, size: 10 }),
})
```

## 路由结构

```
/                     # 重定向到 /dashboard
/login                # 登录页
/dashboard            # 仪表盘
/projects             # 项目列表
/containers           # 容器列表
/datasets             # 数据集列表
/injections           # 故障注入列表
/injections/create    # 创建故障注入
/executions           # 算法执行列表
/evaluations          # 评估页面
/tasks                # 任务监控
/system               # 系统管理
/settings             # 个人设置
```

## 下一步开发计划

1. **实现故障注入可视化编排器** (核心功能)
   - 拖拽式批次管理
   - 并行故障节点配置
   - 实时预览和验证

2. **实现算法执行结果可视化**
   - 服务拓扑图(使用 D3.js 或 Cytoscape)
   - 分层结果展示
   - 准确率分析图表

3. **实现 SSE 实时日志流**
   - EventSource 集成
   - 自动重连机制
   - 日志级别筛选

4. **完善通用组件库**
   - LabelSelector (标签选择器)
   - ContainerSelector (容器选择器)
   - LogStream (日志流组件)
   - TaskStatusBadge (任务状态徽章)

## 贡献指南

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

## License

MIT License
