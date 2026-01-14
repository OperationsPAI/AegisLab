/**
 * 注入 API
 * 使用 @rcabench/client SDK，手工实现缺失的端点
 */
import {
  InjectionsApi,
  type BatchManageInjectionLabelReq,
  type DatapackState,
  type GetInjectionMetadataSystem,
  type InjectionDetailResp,
  type InjectionNoIssuesResp,
  type InjectionWithIssuesResp,
  type LabelItem,
  type ListInjectionResp,
  type ListInjectionsType,
  type ManageInjectionLabelReq,
  type SubmitDatapackBuildingReq,
  type SubmitInjectionReq,
} from '@rcabench/client';

import { apiClient, createApiConfig } from './config';

export const injectionApi = {
  // ==================== SDK 方法 ====================

  /**
   * 获取注入列表 - 使用 SDK
   */
  getInjections: async (params?: {
    page?: number;
    size?: number;
    fault_type?: string;
    benchmark?: string;
    state?: DatapackState;
    status?: number;
    labels?: string[];
  }): Promise<ListInjectionResp | undefined> => {
    const api = new InjectionsApi(createApiConfig());
    const response = await api.listInjections({
      page: params?.page,
      size: params?.size,
      type: params?.fault_type as ListInjectionsType | undefined,
      benchmark: params?.benchmark,
      state: params?.state,
      status: params?.status,
      labels: params?.labels,
    });
    return response.data.data;
  },

  /**
   * 获取注入详情 - 使用 SDK
   */
  getInjection: async (id: number): Promise<InjectionDetailResp | undefined> => {
    const api = new InjectionsApi(createApiConfig());
    const response = await api.getInjectionById({ id });
    return response.data.data;
  },

  /**
   * 提交注入 - 使用 SDK
   */
  submitInjection: async (data: SubmitInjectionReq) => {
    const api = new InjectionsApi(createApiConfig());
    const response = await api.injectFault({ body: data });
    return response.data.data;
  },

  /**
   * 构建数据包 - 使用 SDK
   */
  buildDatapack: async (data: SubmitDatapackBuildingReq) => {
    const api = new InjectionsApi(createApiConfig());
    const response = await api.buildDatapack({ body: data });
    return response.data.data;
  },

  /**
   * 获取故障元数据 - 使用 SDK
   */
  getFaultMetadata: async (system: GetInjectionMetadataSystem) => {
    const api = new InjectionsApi(createApiConfig());
    const response = await api.getInjectionMetadata({ system });
    return response.data.data;
  },

  /**
   * 获取失败的注入 (无问题) - 使用 SDK
   */
  getNoIssues: async (params?: {
    labels?: string[];
    lookback?: string;
    customStartTime?: string;
    customEndTime?: string;
  }): Promise<InjectionNoIssuesResp[] | undefined> => {
    const api = new InjectionsApi(createApiConfig());
    const response = await api.listFailedInjections({
      labels: params?.labels,
      lookback: params?.lookback,
      customStartTime: params?.customStartTime,
      customEndTime: params?.customEndTime,
    });
    return response.data.data;
  },

  /**
   * 获取成功的注入 (有问题) - 使用 SDK
   */
  getWithIssues: async (params?: {
    labels?: string[];
    lookback?: string;
    customStartTime?: string;
    customEndTime?: string;
  }): Promise<InjectionWithIssuesResp[] | undefined> => {
    const api = new InjectionsApi(createApiConfig());
    const response = await api.listSuccessfulInjections({
      labels: params?.labels,
      lookback: params?.lookback,
      customStartTime: params?.customStartTime,
      customEndTime: params?.customEndTime,
    });
    return response.data.data;
  },

  /**
   * 管理注入标签 - 使用 SDK
   */
  manageLabels: async (
    id: number,
    manage: ManageInjectionLabelReq
  ) => {
    const api = new InjectionsApi(createApiConfig());
    const response = await api.manageInjectionLabels({ id, manage });
    return response.data.data;
  },

  /**
   * 批量管理注入标签 - 使用 SDK
   */
  batchManageLabels: async (data: {
    injection_ids: number[];
    add_labels?: LabelItem[];
    remove_labels?: string[];
  }) => {
    const api = new InjectionsApi(createApiConfig());
    // Convert to SDK expected format: items array with InjectionLabelOperation objects
    // Note: SDK expects remove_labels as LabelItem[] (key only), convert from string[]
    const batchManage: BatchManageInjectionLabelReq = {
      items: data.injection_ids.map((id) => ({
        injection_id: id,
        add_labels: data.add_labels,
        remove_labels: data.remove_labels?.map((key) => ({ key })),
      })),
    };
    const response = await api.batchManageInjectionLabels({ batchManage });
    return response.data.data;
  },

  // ==================== 手工实现 (SDK 缺失) ====================

  /**
   * 更新标签 (替换所有) - 手工实现 (SDK 使用增量修改)
   */
  updateLabels: (id: number, labels: Array<{ key: string; value: string }>) =>
    apiClient.patch(`/injections/${id}/labels`, { labels }),

  /**
   * 批量删除注入 - 手工实现 (SDK 缺失)
   */
  batchDelete: (ids: number[]) =>
    apiClient.post('/injections/batch-delete', { ids }),

  /**
   * 创建注入 (可视化创建) - 手工实现 (SDK 缺失)
   */
  createInjection: async (data: {
    project_id: number;
    name: string;
    description?: string;
    container_config: {
      pedestal_container_id: number;
      benchmark_container_id: number;
      algorithm_container_ids: number[];
    };
    fault_matrix: Array<
      Array<{
        id: number;
        name: string;
        type: string;
        category?: string;
        parameters?: unknown[];
      }>
    >;
    experiment_params: {
      duration: number;
      interval: number;
      parallel: boolean;
    };
    tags?: string[];
  }) => {
    const response = await apiClient.post('/injections', data);
    return response.data;
  },
};
