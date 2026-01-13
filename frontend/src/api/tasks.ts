/**
 * 任务 API
 * 使用 @rcabench/client SDK，手工实现缺失的端点
 */
import {
  TasksApi,
  TracesApi,
  type GroupStats,
  type ListTaskResp,
  type ListTasksTaskType,
  type StatusType,
  type TaskDetailResp,
  type TaskState,
} from '@rcabench/client';
import { apiClient, createApiConfig } from './config';

export const taskApi = {
  // ==================== SDK 方法 ====================

  /**
   * 获取任务列表 - 使用 SDK
   */
  getTasks: async (params?: {
    page?: number;
    size?: number;
    taskType?: ListTasksTaskType;
    immediate?: boolean;
    traceId?: string;
    groupId?: string;
    projectId?: number;
    state?: TaskState;
    status?: StatusType;
  }): Promise<ListTaskResp | undefined> => {
    const api = new TasksApi(createApiConfig());
    const response = await api.listTasks({
      page: params?.page,
      size: params?.size,
      taskType: params?.taskType,
      immediate: params?.immediate,
      traceId: params?.traceId,
      groupId: params?.groupId,
      projectId: params?.projectId,
      state: params?.state,
      status: params?.status,
    });
    return response.data.data;
  },

  /**
   * 获取任务详情 - 使用 SDK
   */
  getTask: async (taskId: string): Promise<TaskDetailResp | undefined> => {
    const api = new TasksApi(createApiConfig());
    const response = await api.getTaskById({ taskId });
    return response.data.data;
  },

  /**
   * 获取追踪组统计 - 使用 SDK
   */
  getGroupStats: async (groupId: string): Promise<GroupStats | undefined> => {
    const api = new TracesApi(createApiConfig());
    const response = await api.getGroupStats({ groupId });
    return response.data.data;
  },

  // ==================== 手工实现 (SDK 缺失) ====================

  /**
   * 批量删除任务 - 手工实现 (SDK 缺失)
   */
  batchDelete: (ids: string[]) =>
    apiClient.post('/tasks/batch-delete', { ids }),
};

/**
 * 创建实时日志流 (SSE)
 * 注意：EventSource 不支持自定义 headers，
 * 如果需要认证，后端需要支持通过 query params 传递 token
 */
export const createLogStream = (traceId: string): EventSource => {
  const token = localStorage.getItem('access_token');
  // SSE 连接通常需要通过 URL 参数传递 token
  const url = `/api/v2/traces/${traceId}/stream${token ? `?token=${encodeURIComponent(token)}` : ''}`;

  return new EventSource(url);
};
