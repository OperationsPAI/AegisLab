# Migrating Frontend to Use Generated TypeScript SDK

This guide explains how to migrate the frontend from using manual API clients to the generated TypeScript SDK.

## Step 1: Install the SDK

```bash
cd frontend
npm install @rcabench/client
```

## Step 2: Create SDK Configuration

Create a new file `frontend/src/api/sdk-config.ts`:

```typescript
import { Configuration } from '@rcabench/client'

// Create a singleton configuration instance
let config: Configuration | null = null

export const getSdkConfig = (): Configuration => {
  if (!config) {
    config = new Configuration({
      basePath: '/api/v2',
      accessToken: localStorage.getItem('access_token') || undefined,
    })
  }

  // Update token if it changes
  const token = localStorage.getItem('access_token')
  if (token !== config.accessToken) {
    config.accessToken = token || undefined
  }

  return config
}

// Export a function to clear config (useful for logout)
export const clearSdkConfig = () => {
  config = null
}
```

## Step 3: Create API Instance Manager

Create `frontend/src/api/sdk-manager.ts`:

```typescript
import {
  AuthApi,
  ProjectsApi,
  TasksApi,
  ExecutionsApi,
  InjectionsApi,
  ContainersApi,
  EvaluationsApi,
} from '@rcabench/client'
import { getSdkConfig } from './sdk-config'

// Create API instances with proper configuration
export const getAuthApi = () => new AuthApi(getSdkConfig())
export const getProjectsApi = () => new ProjectsApi(getSdkConfig())
export const getTasksApi = () => new TasksApi(getSdkConfig())
export const getExecutionsApi = () => new ExecutionsApi(getSdkConfig())
export const getInjectionsApi = () => new InjectionsApi(getSdkConfig())
export const getContainersApi = () => new ContainersApi(getSdkConfig())
export const getEvaluationsApi = () => new EvaluationsApi(getSdkConfig())
```

## Step 4: Update API Files

### Update `frontend/src/api/auth.ts`

```typescript
import { getAuthApi } from './sdk-manager'
import type { LoginRequest, RefreshRequest } from '@rcabench/client'

export const authApi = {
  login: async (username: string, password: string) => {
    const authApi = getAuthApi()
    const response = await authApi.login({
      username,
      password,
    } as LoginRequest)
    return response.data
  },

  refresh: async (token: string) => {
    const authApi = getAuthApi()
    const response = await authApi.refresh({
      token,
    } as RefreshRequest)
    return response.data
  },
}
```

### Update `frontend/src/api/projects.ts`

```typescript
import { getProjectsApi } from './sdk-manager'
import type {
  Project,
  Label,
  ProjectsGetProjectsRequest,
} from '@rcabench/client'

export const projectApi = {
  getProjects: async (params?: Partial<ProjectsGetProjectsRequest>) => {
    const api = getProjectsApi()
    const response = await api.getProjects(params)
    return response.data
  },

  getProject: async (id: number) => {
    const api = getProjectsApi()
    const response = await api.getProject(id)
    return response.data
  },

  createProject: async (data: {
    name: string
    description?: string
    is_public?: boolean
    labels?: Label[]
  }) => {
    const api = getProjectsApi()
    const response = await api.createProject(data)
    return response.data
  },

  updateProject: async (id: number, data: Partial<Project>) => {
    const api = getProjectsApi()
    const response = await api.updateProject(id, data)
    return response.data
  },

  deleteProject: async (id: number) => {
    const api = getProjectsApi()
    await api.deleteProject(id)
  },

  updateLabels: async (id: number, labels: Label[]) => {
    const api = getProjectsApi()
    await api.updateProjectLabels(id, { labels })
  },
}
```

## Step 5: Update Components

### Update Project List Component

In `frontend/src/pages/projects/ProjectList.tsx`:

```typescript
// Before
import { projectApi } from '@/api/projects'
import type { Project } from '@/types/api'

// After
import { projectApi } from '@/api/projects'
import type { Project } from '@rcabench/client'
```

### Update Login Component

In `frontend/src/pages/auth/Login.tsx`:

```typescript
// Update the login function
const handleLogin = async (values: LoginForm) => {
  setLoading(true)
  try {
    const data = await authApi.login(values.username, values.password)

    // Store tokens
    localStorage.setItem('access_token', data.token)
    localStorage.setItem('refresh_token', data.token)

    // Clear SDK config to force refresh
    clearSdkConfig()

    message.success('登录成功')
    navigate('/')
  } catch (error) {
    message.error('登录失败，请检查用户名和密码')
  } finally {
    setLoading(false)
  }
}
```

## Step 6: Update Type Imports

Replace all imports from `@/types/api` with imports from `@rcabench/client`:

```typescript
// Before
import type { Project, Task, Execution } from '@/types/api'

// After
import type { Project, Task, Execution } from '@rcabench/client'
```

## Step 7: Handle SSE Events

For Server-Sent Events (like execution logs), you might need to handle them differently:

```typescript
import { getExecutionsApi } from './sdk-manager'

export const executionApi = {
  // For SSE events, you might need to use the native EventSource
  // or axios directly if the generated SDK doesn't support SSE well
  getExecutionLogs: (executionId: number, onMessage: (data: any) => void) => {
    const token = localStorage.getItem('access_token')
    const eventSource = new EventSource(
      `/api/v2/executions/${executionId}/logs`,
      {
        headers: token ? { Authorization: `Bearer ${token}` } : {}
      }
    )

    eventSource.onmessage = (event) => {
      onMessage(JSON.parse(event.data))
    }

    return eventSource
  },
}
```

## Step 8: Update Error Handling

The SDK might have different error handling:

```typescript
try {
  const projects = await projectApi.getProjects()
} catch (error) {
  // The SDK error structure might be different
  if (error.response?.status === 404) {
    message.error('项目不存在')
  } else if (error.response?.status === 401) {
    // Handle unauthorized
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    clearSdkConfig()
    window.location.href = '/login'
  } else {
    message.error(error.message || '请求失败')
  }
}
```

## Step 9: Clean Up

Once everything is working:

1. Delete the old manual API files in `frontend/src/api/` (except `client.ts` if still needed for SSE)
2. Delete the manual type definitions in `frontend/src/types/api.ts`
3. Update all imports to use the new SDK-based API files
4. Test all functionality thoroughly

## Benefits of Using Generated SDK

1. **Type Safety**: All API calls are fully typed
2. **Consistency**: API calls follow the same pattern
3. **Auto-update**: When backend API changes, just regenerate the SDK
4. **Documentation**: The SDK includes all API documentation
5. **Less Code**: No need to maintain manual API clients

## Troubleshooting

### Common Issues

1. **Missing Types**: If some types are missing, check if the backend API has proper Swagger annotations
2. **Authentication Issues**: Ensure the SDK configuration is updated when tokens change
3. **SSE Not Working**: The generated SDK might not handle SSE well; use native EventSource for streaming

### Regenerating SDK

When the backend API changes:

```bash
# From project root
make generate-typescript-sdk SDK_VERSION=1.0.1

# Then in frontend
npm update @rcabench/client
```