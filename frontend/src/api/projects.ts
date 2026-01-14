/**
 * 项目 API
 * 使用 @rcabench/client SDK，手工实现缺失的端点
 */
import {
  ProjectsApi,
  type CreateProjectReq,
  type LabelItem,
  type ListProjectResp,
  type ProjectDetailResp,
  type ProjectResp,
  type StatusType,
} from '@rcabench/client';

import { apiClient, createApiConfig } from './config';

export const projectApi = {
  // ==================== SDK 方法 ====================

  /**
   * 获取项目列表 - 使用 SDK
   */
  getProjects: async (params?: {
    page?: number;
    size?: number;
    isPublic?: boolean;
    status?: StatusType;
  }): Promise<ListProjectResp | undefined> => {
    const api = new ProjectsApi(createApiConfig());
    const response = await api.listProjects({
      page: params?.page,
      size: params?.size,
      isPublic: params?.isPublic,
      status: params?.status,
    });
    return response.data.data;
  },

  /**
   * 获取项目详情 - 使用 SDK
   */
  getProject: async (id: number): Promise<ProjectDetailResp | undefined> => {
    const api = new ProjectsApi(createApiConfig());
    const response = await api.getProjectById({ projectId: id });
    return response.data.data;
  },

  /**
   * 创建项目 - 使用 SDK
   */
  createProject: async (data: {
    name: string;
    description?: string;
    is_public?: boolean;
  }): Promise<ProjectResp | undefined> => {
    const api = new ProjectsApi(createApiConfig());
    const request: CreateProjectReq = {
      name: data.name,
      description: data.description,
      is_public: data.is_public ?? false,
    };
    const response = await api.createProject({ request });
    return response.data.data;
  },

  // ==================== 手工实现 (SDK 缺失) ====================

  /**
   * 更新项目 - 手工实现 (SDK 缺失)
   */
  updateProject: (
    id: number,
    data: {
      name?: string;
      description?: string;
      is_public?: boolean;
      labels?: LabelItem[];
    }
  ) => apiClient.patch<{ data: ProjectDetailResp }>(`/projects/${id}`, data),

  /**
   * 删除项目 - 手工实现 (SDK 缺失)
   */
  deleteProject: (id: number) => apiClient.delete(`/projects/${id}`),

  /**
   * 管理标签 - 手工实现 (SDK 缺失)
   */
  updateLabels: (id: number, labels: Array<{ key: string; value: string }>) =>
    apiClient.patch(`/projects/${id}/labels`, { labels }),
};
