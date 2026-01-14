/**
 * 用户管理 API
 * 使用 @rcabench/client SDK
 */
import {
  UsersApi,
  type CreateUserReq,
  type UpdateUserReq,
  type UserResp,
  type UserDetailResp,
  type ListUserResp,
  type StatusType,
} from '@rcabench/client';

import { createApiConfig } from './config';

export const usersApi = {
  // ==================== SDK 方法 ====================

  /**
   * 获取用户列表 - 使用 SDK
   */
  getUsers: async (params?: {
    page?: number;
    size?: number;
    username?: string;
    email?: string;
    isActive?: boolean;
    status?: StatusType;
  }): Promise<ListUserResp | undefined> => {
    const api = new UsersApi(createApiConfig());
    const response = await api.listUsers({
      page: params?.page,
      size: params?.size,
      username: params?.username,
      email: params?.email,
      isActive: params?.isActive,
      status: params?.status,
    });
    return response.data.data;
  },

  /**
   * 获取用户详情 - 使用 SDK
   */
  getUserDetail: async (id: number): Promise<UserDetailResp | undefined> => {
    const api = new UsersApi(createApiConfig());
    const response = await api.getUserById({ id });
    return response.data.data;
  },

  /**
   * 创建用户 - 使用 SDK
   */
  createUser: async (data: CreateUserReq): Promise<UserResp | undefined> => {
    const api = new UsersApi(createApiConfig());
    const response = await api.createUser({ request: data });
    return response.data.data;
  },

  /**
   * 更新用户 - 使用 SDK
   */
  updateUser: async (
    id: number,
    data: UpdateUserReq
  ): Promise<UserResp | undefined> => {
    const api = new UsersApi(createApiConfig());
    const response = await api.updateUser({ id, request: data });
    return response.data.data;
  },

  /**
   * 删除用户 - 使用 SDK
   */
  deleteUser: async (id: number): Promise<void> => {
    const api = new UsersApi(createApiConfig());
    await api.deleteUser({ id });
  },
};

// 重新导出类型以便其他文件使用
export type { UserResp, UserDetailResp, ListUserResp, CreateUserReq, UpdateUserReq };
