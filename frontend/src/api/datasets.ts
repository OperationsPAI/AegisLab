/**
 * 数据集 API
 * 使用 @rcabench/client SDK，手工实现缺失的端点
 */
import {
  DatasetsApi,
  type CreateDatasetReq,
  type CreateDatasetVersionReq,
  type DatasetDetailResp,
  type DatasetVersionResp,
  type LabelItem,
  type ListDatasetResp,
  type ListDatasetVersionResp,
  type StatusType,
} from '@rcabench/client';
import { apiClient, createApiConfig } from './config';

// Dataset type for API calls
type DatasetTypeParam = 'Trace' | 'Log' | 'Metric';

export const datasetApi = {
  // ==================== SDK 方法 ====================

  /**
   * 获取数据集列表 - 使用 SDK
   */
  getDatasets: async (params?: {
    page?: number;
    size?: number;
    type?: DatasetTypeParam;
    is_public?: boolean;
    status?: StatusType;
  }): Promise<ListDatasetResp> => {
    const api = new DatasetsApi(createApiConfig());
    const response = await api.listDatasets({
      page: params?.page,
      size: params?.size,
      type: params?.type,
      isPublic: params?.is_public,
      status: params?.status,
    });
    return response.data.data!;
  },

  /**
   * 获取数据集详情 - 使用 SDK
   */
  getDataset: async (id: number): Promise<DatasetDetailResp> => {
    const api = new DatasetsApi(createApiConfig());
    const response = await api.getDatasetById({ datasetId: id });
    return response.data.data!;
  },

  /**
   * 创建数据集 - 使用 SDK
   */
  createDataset: async (data: {
    name: string;
    type: DatasetTypeParam;
    description?: string;
    is_public?: boolean;
  }) => {
    const api = new DatasetsApi(createApiConfig());
    const request: CreateDatasetReq = {
      name: data.name,
      type: data.type,
      description: data.description,
      is_public: data.is_public,
    };
    const response = await api.createDataset({ request });
    return response.data.data;
  },

  /**
   * 获取数据集版本列表 - 使用 SDK
   */
  getVersions: async (
    datasetId: number,
    params?: { page?: number; size?: number; status?: StatusType }
  ): Promise<ListDatasetVersionResp> => {
    const api = new DatasetsApi(createApiConfig());
    const response = await api.listDatasetVersions({
      datasetId,
      page: params?.page,
      size: params?.size,
      status: params?.status,
    });
    return response.data.data!;
  },

  /**
   * 创建数据集版本 - 使用 SDK
   */
  createVersion: async (
    datasetId: number,
    data: { name: string; datapacks?: string[] }
  ): Promise<DatasetVersionResp | undefined> => {
    const api = new DatasetsApi(createApiConfig());
    const request: CreateDatasetVersionReq = {
      name: data.name,
      datapacks: data.datapacks,
    };
    const response = await api.createDatasetVersion({
      datasetId,
      request,
    });
    return response.data.data;
  },

  // ==================== 手工实现 (SDK 缺失) ====================

  /**
   * 更新数据集 - 手工实现 (SDK 缺失)
   */
  updateDataset: (
    id: number,
    data: {
      name?: string;
      type?: DatasetTypeParam;
      description?: string;
      is_public?: boolean;
      labels?: LabelItem[];
    }
  ) => apiClient.patch<{ data: DatasetDetailResp }>(`/datasets/${id}`, data),

  /**
   * 删除数据集 - 手工实现 (SDK 缺失)
   */
  deleteDataset: (id: number) => apiClient.delete(`/datasets/${id}`),

  /**
   * 更新数据集版本 - 手工实现 (SDK 缺失)
   */
  updateVersion: (
    datasetId: number,
    versionId: number,
    data: { name?: string; datapacks?: string[] }
  ) =>
    apiClient.patch<{ data: DatasetVersionResp }>(
      `/datasets/${datasetId}/versions/${versionId}`,
      data
    ),

  /**
   * 删除数据集版本 - 手工实现 (SDK 缺失)
   */
  deleteVersion: (datasetId: number, versionId: number) =>
    apiClient.delete(`/datasets/${datasetId}/versions/${versionId}`),

  /**
   * 上传数据集文件 - 手工实现 (SDK 缺失)
   */
  uploadFile: (datasetId: number, file: File) => {
    const formData = new FormData();
    formData.append('file', file);
    return apiClient.post<{ data: DatasetVersionResp }>(
      `/datasets/${datasetId}/upload`,
      formData,
      {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      }
    );
  },

  /**
   * 批量删除数据集 - 手工实现 (SDK 缺失)
   */
  batchDelete: (ids: number[]) =>
    apiClient.post('/datasets/batch-delete', { ids }),
};
