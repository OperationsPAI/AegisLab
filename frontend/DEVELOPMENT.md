# RCABench 前端开发指南

## 项目概述

RCABench 前端是一个基于 React + TypeScript 的现代化 Web 应用,采用学术研究风格设计,为微服务根因分析提供可视化界面。

## 技术亮点

### 1. 类型安全

所有 API 调用都有完整的 TypeScript 类型定义,确保前后端数据结构一致:

```typescript
// types/api.ts 定义了所有数据类型
export interface Injection {
  id: number
  name: string
  state: InjectionState
  // ...
}

// API 调用自动推断返回类型
const { data } = useQuery({
  queryKey: ['injections'],
  queryFn: () => injectionApi.getInjections(),
})
// data 类型: PaginatedResponse<Injection>
```

### 2. 自动 Token 刷新

Axios 拦截器自动处理 Token 过期和刷新:

```typescript
// api/client.ts
if (error.response?.status === 401 && !originalRequest._retry) {
  // 自动刷新 token 并重试请求
}
```

### 3. 乐观更新

使用 TanStack Query 的 mutation 实现删除等操作的乐观更新:

```typescript
const deleteMutation = useMutation({
  mutationFn: (id: number) => projectApi.deleteProject(id),
  onSuccess: () => {
    // 自动使缓存失效并重新获取
    queryClient.invalidateQueries({ queryKey: ['projects'] })
  },
})
```

### 4. 响应式设计

所有页面使用 Ant Design 的 Grid 系统实现响应式布局:

```typescript
<Row gutter={[16, 16]}>
  <Col xs={24} sm={12} lg={6}>
    {/* 小屏全宽,中屏半宽,大屏1/4宽 */}
  </Col>
</Row>
```

## 代码风格

### 组件组织

```typescript
// 1. Imports
import { useState } from 'react'
import { Button, Table } from 'antd'
import { useQuery } from '@tanstack/react-query'
import { projectApi } from '@/api/projects'
import type { Project } from '@/types/api'

// 2. Types (如果需要)
interface Props {
  // ...
}

// 3. Component
const ProjectList = () => {
  // 3.1 State
  const [page, setPage] = useState(1)

  // 3.2 Queries/Mutations
  const { data, isLoading } = useQuery({
    queryKey: ['projects', { page }],
    queryFn: () => projectApi.getProjects({ page }),
  })

  // 3.3 Handlers
  const handlePageChange = (newPage: number) => {
    setPage(newPage)
  }

  // 3.4 Render
  return (
    <div>
      {/* JSX */}
    </div>
  )
}

// 4. Export
export default ProjectList
```

### 命名约定

- **组件**: PascalCase (`ProjectList`, `MainLayout`)
- **文件**: 与组件同名 (`ProjectList.tsx`)
- **变量**: camelCase (`isLoading`, `handleClick`)
- **常量**: UPPER_SNAKE_CASE (`API_BASE_URL`)
- **类型**: PascalCase (`Project`, `InjectionState`)

### 注释

```typescript
// 使用 JSDoc 注释公共 API
/**
 * 创建故障注入任务
 * @param data 注入配置
 * @returns Promise<Injection>
 */
export const submitInjection = (data: SubmitInjectionReq) => {
  // ...
}

// 使用单行注释解释复杂逻辑
// 计算任务状态分布
const taskStateData = tasks?.data.data.reduce(...)
```

## 调试技巧

### 1. React Query Devtools

安装开发工具查看缓存状态:

```bash
npm install @tanstack/react-query-devtools -D
```

```typescript
// main.tsx
import { ReactQueryDevtools } from '@tanstack/react-query-devtools'

<QueryClientProvider client={queryClient}>
  <App />
  <ReactQueryDevtools initialIsOpen={false} />
</QueryClientProvider>
```

### 2. 网络请求调试

在浏览器开发者工具的 Network 标签查看 API 请求:
- Headers: 查看 Authorization token
- Response: 查看返回数据结构
- Timing: 分析请求性能

### 3. Zustand Devtools

安装 Redux DevTools 扩展可以调试 Zustand 状态:

```typescript
import { create } from 'zustand'
import { devtools } from 'zustand/middleware'

export const useAuthStore = create(
  devtools((set) => ({
    // ...
  }), { name: 'AuthStore' })
)
```

## 性能优化

### 1. 代码分割

使用 React.lazy 进行路由级别的代码分割:

```typescript
const Dashboard = lazy(() => import('@/pages/dashboard/Dashboard'))
const ProjectList = lazy(() => import('@/pages/projects/ProjectList'))
```

### 2. 图片优化

使用 Vite 的资源处理功能:

```typescript
import logo from '@/assets/logo.png?url'  // URL
import logoInline from '@/assets/logo.png?inline'  // Base64
```

### 3. 查询缓存

合理设置 TanStack Query 的缓存时间:

```typescript
const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 5 * 60 * 1000, // 5 分钟
      cacheTime: 10 * 60 * 1000, // 10 分钟
    },
  },
})
```

## 常见问题

### Q: 如何添加新的 API 端点?

1. 在 `src/types/api.ts` 定义类型
2. 在 `src/api/{module}.ts` 添加 API 函数
3. 在组件中使用 `useQuery` 或 `useMutation` 调用

### Q: 如何添加新的路由?

1. 在 `src/pages/{module}/` 创建组件
2. 在 `src/App.tsx` 添加路由配置
3. 在 `src/components/layout/MainLayout.tsx` 添加菜单项

### Q: 如何自定义主题?

修改 `src/main.tsx` 中的 ConfigProvider 配置:

```typescript
<ConfigProvider
  theme={{
    token: {
      colorPrimary: '#your-color',
      // ...
    },
  }}
>
```

### Q: 如何处理 CORS 错误?

开发模式使用 Vite 代理,生产环境需要后端配置 CORS 头:

```go
// Go backend
c.Header("Access-Control-Allow-Origin", "https://your-frontend-domain.com")
```

## 推荐 VS Code 扩展

- ESLint
- Prettier
- TypeScript Vue Plugin (Volar)
- Tailwind CSS IntelliSense
- Auto Rename Tag
- Path Intellisense

## 资源链接

- [React 文档](https://react.dev/)
- [TypeScript 文档](https://www.typescriptlang.org/docs/)
- [Ant Design 文档](https://ant.design/)
- [TanStack Query 文档](https://tanstack.com/query/latest)
- [Zustand 文档](https://zustand-demo.pmnd.rs/)
- [Vite 文档](https://vitejs.dev/)
