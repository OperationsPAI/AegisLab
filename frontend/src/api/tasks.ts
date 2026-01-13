import { Configuration, TasksApi } from '@rcabench/client';
import axios, { type AxiosRequestConfig } from 'axios';

// Create configuration with dynamic token
const createTaskConfig = () => {
  const token = localStorage.getItem('access_token');

  return new Configuration({
    basePath: '/api/v2',
    accessToken: token ? `Bearer ${token}` : undefined,
    baseOptions: {
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    } as AxiosRequestConfig,
  });
};

// Create axios instance for manual API calls
const apiClient = axios.create({
  baseURL: '/api/v2',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor for auth
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Export the tasks API using generated SDK where available
export const taskApi = {
  // Get tasks list - using generated SDK
  getTasks: async (params?: {
    page?: number;
    size?: number;
    taskType?: number;
    state?: number;
    status?: number;
    traceId?: string;
    groupId?: string;
    projectId?: number;
  }) => {
    const tasksApi = new TasksApi(createTaskConfig());
    const response = await tasksApi.listTasks({
      page: params?.page,
      size: params?.size,
      taskType: params?.taskType,
      state: params?.state,
      status: params?.status,
      traceId: params?.traceId,
      groupId: params?.groupId,
      projectId: params?.projectId,
    });
    return response.data;
  },

  // Get task detail - using generated SDK
  getTask: async (taskId: string) => {
    const tasksApi = new TasksApi(createTaskConfig());
    const response = await tasksApi.getTaskById({ taskId });
    return response.data;
  },

  // Batch delete - manual endpoint (not in generated SDK)
  batchDelete: (ids: string[]) =>
    apiClient.post('/tasks/batch-delete', { ids }),

  // Get group stats - manual endpoint (not in generated SDK)
  getGroupStats: (groupId: string) =>
    apiClient.get(`/traces/group/stats`, { params: { groupId } }),
};

// SSE stream helper
export const createLogStream = (traceId: string) => {
  const token = localStorage.getItem('access_token');
  const url = `/api/v2/traces/${traceId}/stream`;

  return new EventSource(url, {
    headers: {
      Authorization: `Bearer ${token}`,
    },
  } as EventSourceInit);
};
