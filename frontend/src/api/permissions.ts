/**
 * 权限 API
 * 使用 @rcabench/client SDK
 */
import {
  PermissionsApi,
  type CreatePermissionReq,
  type ListPermissionReq,
  type PermissionDetailResp,
  type PermissionResp,
  type UpdatePermissionReq,
} from '@rcabench/client';

import { createApiConfig } from './config';

export const permissionApi = {
  /**
   * 获取权限列表
   */
  getPermissions: async (params?: ListPermissionReq) => {
    const api = new PermissionsApi(createApiConfig());
    const response = await api.listPermissions(params);
    return response.data.data;
  },

  /**
   * 获取权限详情
   */
  getPermission: async (id: number): Promise<PermissionDetailResp | undefined> => {
    const api = new PermissionsApi(createApiConfig());
    const response = await api.getPermissionById({ id });
    return response.data.data;
  },

  /**
   * 创建权限
   */
  createPermission: async (data: CreatePermissionReq): Promise<PermissionResp | undefined> => {
    const api = new PermissionsApi(createApiConfig());
    const response = await api.createPermission({ request: data });
    return response.data.data;
  },

  /**
   * 更新权限
   */
  updatePermission: async (id: number, data: UpdatePermissionReq): Promise<PermissionResp | undefined> => {
    const api = new PermissionsApi(createApiConfig());
    const response = await api.updatePermission({ id, request: data });
    return response.data.data;
  },

  /**
   * 删除权限
   */
  deletePermission: async (id: number) => {
    const api = new PermissionsApi(createApiConfig());
    await api.deletePermission({ id });
  },

  /**
   * 获取权限关联的角色列表
   */
  getPermissionRoles: async (permissionId: number) => {
    const api = new PermissionsApi(createApiConfig());
    const response = await api.listRolesFromPermission({ permissionId });
    return response.data.data;
  },
};
