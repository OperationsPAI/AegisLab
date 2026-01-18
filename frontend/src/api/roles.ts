/**
 * 角色 API
 * 使用 @rcabench/client SDK
 */
import {
  RolesApi,
  type AssignRolePermissionReq,
  type CreateRoleReq,
  type ListRoleReq,
  type RemoveRolePermissionReq,
  type RoleDetailResp,
  type RoleResp,
  type UpdateRoleReq,
} from '@rcabench/client';

import { createApiConfig } from './config';

export const roleApi = {
  /**
   * 获取角色列表
   */
  getRoles: async (params?: ListRoleReq) => {
    const api = new RolesApi(createApiConfig());
    const response = await api.listRoles(params);
    return response.data.data;
  },

  /**
   * 获取角色详情
   */
  getRole: async (id: number): Promise<RoleDetailResp | undefined> => {
    const api = new RolesApi(createApiConfig());
    const response = await api.getRoleById({ id });
    return response.data.data;
  },

  /**
   * 创建角色
   */
  createRole: async (data: CreateRoleReq): Promise<RoleResp | undefined> => {
    const api = new RolesApi(createApiConfig());
    const response = await api.createRole({ request: data });
    return response.data.data;
  },

  /**
   * 更新角色
   */
  updateRole: async (id: number, data: UpdateRoleReq): Promise<RoleResp | undefined> => {
    const api = new RolesApi(createApiConfig());
    const response = await api.updateRole({ id, request: data });
    return response.data.data;
  },

  /**
   * 删除角色
   */
  deleteRole: async (id: number) => {
    const api = new RolesApi(createApiConfig());
    await api.deleteRole({ id });
  },

  /**
   * 为角色分配权限
   */
  assignPermissions: async (roleId: number, data: AssignRolePermissionReq) => {
    const api = new RolesApi(createApiConfig());
    await api.grantPermissionsToRole({ roleId, request: data });
  },

  /**
   * 从角色移除权限
   */
  removePermissions: async (roleId: number, data: RemoveRolePermissionReq) => {
    const api = new RolesApi(createApiConfig());
    await api.revokePermissionsFromRole({ roleId, request: data });
  },
};
