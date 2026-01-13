/**
 * 容器 API
 * 使用 @rcabench/client SDK，手工实现缺失的端点
 */
import {
  ContainersApi,
  type ContainerDetailResp,
  type ContainerResp,
  type ContainerType,
  type ContainerVersionResp,
  type CreateContainerReq,
  type CreateContainerVersionReq,
  type LabelItem,
  type ListContainerResp,
  type ListContainerVersionResp,
  type StatusType,
} from '@rcabench/client';
import { apiClient, createApiConfig } from './config';

export const containerApi = {
  // ==================== SDK 方法 ====================

  /**
   * 获取容器列表 - 使用 SDK
   */
  getContainers: async (params?: {
    page?: number;
    size?: number;
    type?: ContainerType;
    isPublic?: boolean;
    status?: StatusType;
  }): Promise<ListContainerResp> => {
    const api = new ContainersApi(createApiConfig());
    const response = await api.listContainers({
      page: params?.page,
      size: params?.size,
      type: params?.type,
      isPublic: params?.isPublic,
      status: params?.status,
    });
    return response.data.data!;
  },

  /**
   * 获取容器详情 - 使用 SDK
   */
  getContainer: async (id: number): Promise<ContainerDetailResp> => {
    const api = new ContainersApi(createApiConfig());
    const response = await api.getContainerById({ containerId: id });
    return response.data.data!;
  },

  /**
   * 创建容器 - 使用 SDK
   */
  createContainer: async (data: {
    name: string;
    type: ContainerType;
    readme?: string;
    is_public?: boolean;
  }): Promise<ContainerResp | undefined> => {
    const api = new ContainersApi(createApiConfig());
    const request: CreateContainerReq = {
      name: data.name,
      type: data.type,
      readme: data.readme,
      is_public: data.is_public,
    };
    const response = await api.createContainer({ request });
    return response.data.data;
  },

  /**
   * 获取容器版本列表 - 使用 SDK
   */
  getVersions: async (
    containerId: number,
    params?: { page?: number; size?: number; status?: StatusType }
  ): Promise<ListContainerVersionResp> => {
    const api = new ContainersApi(createApiConfig());
    const response = await api.listContainerVersions({
      containerId,
      page: params?.page,
      size: params?.size,
      status: params?.status,
    });
    return response.data.data!;
  },

  /**
   * 创建容器版本 - 使用 SDK
   * 注意：SDK 使用 image_ref 格式，如 "registry/repository:tag"
   */
  createVersion: async (
    containerId: number,
    data: {
      name: string;
      image_ref: string;
      command?: string;
    }
  ): Promise<ContainerVersionResp | undefined> => {
    const api = new ContainersApi(createApiConfig());
    const request: CreateContainerVersionReq = {
      name: data.name,
      image_ref: data.image_ref,
      command: data.command,
    };
    const response = await api.createContainerVersion({
      containerId,
      request,
    });
    return response.data.data;
  },

  // ==================== 手工实现 (SDK 缺失) ====================

  /**
   * 更新容器 - 手工实现 (SDK 缺失)
   */
  updateContainer: (
    id: number,
    data: {
      name?: string;
      type?: ContainerType;
      readme?: string;
      is_public?: boolean;
      labels?: LabelItem[];
    }
  ) => apiClient.patch<{ data: ContainerDetailResp }>(`/containers/${id}`, data),

  /**
   * 删除容器 - 手工实现 (SDK 缺失)
   */
  deleteContainer: (id: number) => apiClient.delete(`/containers/${id}`),

  /**
   * 更新容器版本 - 手工实现 (SDK 缺失)
   */
  updateVersion: (
    containerId: number,
    versionId: number,
    data: {
      name?: string;
      image_ref?: string;
      command?: string;
    }
  ) =>
    apiClient.patch<{ data: ContainerVersionResp }>(
      `/containers/${containerId}/versions/${versionId}`,
      data
    ),

  /**
   * 删除容器版本 - 手工实现 (SDK 缺失)
   */
  deleteVersion: (containerId: number, versionId: number) =>
    apiClient.delete(`/containers/${containerId}/versions/${versionId}`),
};
