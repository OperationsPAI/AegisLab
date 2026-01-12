import apiClient from './client'
import type {
  Task,
  PaginationParams,
  PaginatedResponse,
  TaskType,
} from '@/types/api'

export const taskApi = {
  // Get tasks list
  getTasks: (params?: Partial<PaginationParams> & {
    task_type?: TaskType
    state?: number
    status?: number
    trace_id?: string
    group_id?: string
    project_id?: number
  }) => apiClient.get<PaginatedResponse<Task>>('/tasks', { params }),

  // Get task detail
  getTask: (taskId: string) => apiClient.get<Task>(`/tasks/${taskId}`),

  // Batch delete
  batchDelete: (ids: string[]) =>
    apiClient.post('/tasks/batch-delete', { ids }),

  // Get group stats
  getGroupStats: (groupId: string) =>
    apiClient.get(`/traces/group/stats`, { params: { group_id: groupId } }),
}

// SSE stream helper
export const createLogStream = (traceId: string) => {
  const token = localStorage.getItem('access_token')
  const url = `/api/v2/traces/${traceId}/stream`

  return new EventSource(url, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  } as EventSourceInit)
}
