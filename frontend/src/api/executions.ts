/**
 * 执行 API
 * 使用 @rcabench/client SDK，手工实现缺失的端点
 */
import {
  ExecutionsApi,
  type ContainerSpec,
  type ExecutionDetailResp,
  type ExecutionSpec,
  type ExecutionState,
  type GranularityResultItem,
  type DetectorResultItem,
  type LabelItem,
  type ListExecutionResp,
  type StatusType,
  type SubmitExecutionReq,
  type UploadDetectorResultReq,
  type UploadGranularityResultReq,
} from '@rcabench/client';

import { apiClient, createApiConfig } from './config';

export const executionApi = {
  // ==================== SDK 方法 ====================

  /**
   * 获取执行列表 - 使用 SDK
   */
  getExecutions: async (params?: {
    page?: number;
    size?: number;
    state?: ExecutionState;
    status?: StatusType;
    labels?: string[];
  }): Promise<ListExecutionResp> => {
    const api = new ExecutionsApi(createApiConfig());
    const response = await api.listExecutions({
      page: params?.page,
      size: params?.size,
      state: params?.state,
      status: params?.status,
      labels: params?.labels,
    });
    return response.data.data!;
  },

  /**
   * 获取执行详情 - 使用 SDK
   */
  getExecution: async (id: number): Promise<ExecutionDetailResp> => {
    const api = new ExecutionsApi(createApiConfig());
    const response = await api.getExecutionById({ id });
    return response.data.data!;
  },

  /**
   * 执行算法 - 使用 SDK
   */
  executeAlgorithm: async (data: {
    algorithmName: string;
    algorithmVersion: string;
    datapackId: string;
    labels?: Array<{ key: string; value: string }>;
  }) => {
    const api = new ExecutionsApi(createApiConfig());
    const algorithm: ContainerSpec = {
      name: data.algorithmName,
      version: data.algorithmVersion,
    };
    const spec: ExecutionSpec = {
      algorithm,
      datapack: data.datapackId,
    };
    const request: SubmitExecutionReq = {
      project_name: 'default',
      specs: [spec],
      labels: data.labels as LabelItem[],
    };
    const response = await api.runAlgorithm({ request });
    return response.data.data;
  },

  /**
   * 上传检测结果 - 使用 SDK
   */
  uploadDetectorResults: async (
    id: number,
    data: { duration: number; results: DetectorResultItem[] }
  ) => {
    const api = new ExecutionsApi(createApiConfig());
    const request: UploadDetectorResultReq = {
      duration: data.duration,
      results: data.results,
    };
    const response = await api.uploadDetectionResults({
      executionId: id,
      request,
    });
    return response.data.data;
  },

  /**
   * 上传定位结果 - 使用 SDK
   */
  uploadGranularityResults: async (
    id: number,
    data: { duration: number; results: GranularityResultItem[] }
  ) => {
    const api = new ExecutionsApi(createApiConfig());
    const request: UploadGranularityResultReq = {
      duration: data.duration,
      results: data.results,
    };
    const response = await api.uploadLocalizationResults({
      executionId: id,
      request,
    });
    return response.data.data;
  },

  /**
   * 获取执行标签列表 - 使用 SDK
   */
  getExecutionLabels: async (): Promise<LabelItem[] | undefined> => {
    const api = new ExecutionsApi(createApiConfig());
    const response = await api.listExecutionLabels();
    return response.data.data;
  },

  // ==================== 手工实现 (SDK 缺失) ====================

  /**
   * 更新执行标签 - 手工实现 (SDK 缺失)
   */
  updateLabels: (id: number, labels: Array<{ key: string; value: string }>) =>
    apiClient.patch(`/executions/${id}/labels`, { labels }),

  /**
   * 批量删除执行 - 手工实现 (SDK 缺失)
   */
  batchDelete: (ids: number[]) =>
    apiClient.post('/executions/batch-delete', { ids }),
};
