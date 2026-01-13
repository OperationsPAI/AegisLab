import { Configuration, ProjectsApi } from '@rcabench/client';
import axios, { type AxiosRequestConfig } from 'axios';

// Create configuration with dynamic token
const createProjectConfig = () => {
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

// Export the projects API using generated SDK where available
export const projectApi = {
  // Get projects list - using generated SDK
  getProjects: async (params?: {
    page?: number;
    size?: number;
    isPublic?: boolean;
    status?: number;
  }) => {
    const projectsApi = new ProjectsApi(createProjectConfig());
    const response = await projectsApi.listProjects({
      page: params?.page,
      size: params?.size,
      isPublic: params?.isPublic,
      status: params?.status,
    });
    return response.data;
  },

  // Get project detail - using generated SDK
  getProject: async (id: number) => {
    const projectsApi = new ProjectsApi(createProjectConfig());
    const response = await projectsApi.getProjectById({ projectId: id });
    return response.data;
  },

  // Create project - using generated SDK
  createProject: async (data: {
    name: string;
    description?: string;
    isPublic?: boolean;
    labels?: Array<{ key: string; value: string }>;
  }) => {
    const projectsApi = new ProjectsApi(createProjectConfig());
    const response = await projectsApi.createProject({
      request: {
        name: data.name,
        description: data.description,
        is_public: data.isPublic || false,
      },
    });
    return response.data;
  },

  // Update project - manual endpoint (not in generated SDK)
  updateProject: (id: number, data: Record<string, unknown>) =>
    apiClient.patch(`/projects/${id}`, data),

  // Delete project - manual endpoint (not in generated SDK)
  deleteProject: (id: number) => apiClient.delete(`/projects/${id}`),

  // Manage labels - manual endpoint (not in generated SDK)
  updateLabels: (id: number, labels: Array<{ key: string; value: string }>) =>
    apiClient.patch(`/projects/${id}/labels`, { labels }),
};

export default apiClient;
