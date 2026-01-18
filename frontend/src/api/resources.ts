/**
 * 资源 API
 * 使用 @rcabench/client SDK
 */
import {
  ResourcesApi,
  type ListResourceReq,
  type ResourceResp,
} from '@rcabench/client';

import { createApiConfig } from './config';

export const resourceApi = {
  /**
   * 获取资源列表
   */
  getResources: async (params?: ListResourceReq) => {
    const api = new ResourcesApi(createApiConfig());
    const response = await api.listResources(params);
    return response.data.data;
  },

  /**
   * 获取资源详情
   */
  getResource: async (id: number) => {
    const api = new ResourcesApi(createApiConfig());
    const response = await api.getResourceById({ id });
    return response.data.data;
  },

  /**
   * 获取资源的权限列表
   */
  getResourcePermissions: async (id: number) => {
    const api = new ResourcesApi(createApiConfig());
    const response = await api.listResourcePermissions({ id });
    return response.data.data;
  },
};
