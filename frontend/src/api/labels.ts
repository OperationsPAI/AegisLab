/**
 * 标签 API
 * 使用 @rcabench/client SDK
 */
import {
  LabelsApi,
  type BatchDeleteLabelReq,
  type CreateLabelReq,
  type LabelDetailResp,
  type LabelResp,
  type ListLabelReq,
  type UpdateLabelReq,
} from '@rcabench/client';

import { createApiConfig } from './config';

export const labelApi = {
  /**
   * 获取标签列表
   */
  getLabels: async (params?: ListLabelReq) => {
    const api = new LabelsApi(createApiConfig());
    const response = await api.listLabels(params);
    return response.data.data;
  },

  /**
   * 获取标签详情
   */
  getLabel: async (labelId: number): Promise<LabelDetailResp | undefined> => {
    const api = new LabelsApi(createApiConfig());
    const response = await api.getLabelById({ labelId });
    return response.data.data;
  },

  /**
   * 创建标签
   */
  createLabel: async (data: CreateLabelReq): Promise<LabelResp | undefined> => {
    const api = new LabelsApi(createApiConfig());
    const response = await api.createLabel({ request: data });
    return response.data.data;
  },

  /**
   * 更新标签
   */
  updateLabel: async (labelId: number, data: UpdateLabelReq): Promise<LabelResp | undefined> => {
    const api = new LabelsApi(createApiConfig());
    const response = await api.updateLabel({ labelId, request: data });
    return response.data.data;
  },

  /**
   * 删除标签
   */
  deleteLabel: async (labelId: number) => {
    const api = new LabelsApi(createApiConfig());
    await api.deleteLabel({ labelId });
  },

  /**
   * 批量删除标签
   */
  batchDeleteLabels: async (data: BatchDeleteLabelReq) => {
    const api = new LabelsApi(createApiConfig());
    await api.batchDeleteLabels({ request: data });
  },
};
