import {
  AuthenticationApi,
  type LoginReq,
  type LoginResp,
  type RegisterReq,
  type UserInfo,
  type UserDetailResp,
} from '@rcabench/client';

import { apiClient, createApiConfig } from './config';

export const authApi = {
  login: async (data: LoginReq): Promise<LoginResp | undefined> => {
    const api = new AuthenticationApi(createApiConfig());
    const response = await api.login({ request: data });
    return response.data.data;
  },

  register: async (data: RegisterReq): Promise<UserInfo | undefined> => {
    const api = new AuthenticationApi(createApiConfig());
    const response = await api.registerUser({ request: data });
    return response.data.data;
  },

  logout: () => apiClient.post('/auth/logout'),

  /**
   * 获取当前用户详细信息
   * 后端返回 UserDetailResp（包含 UserResp 基础信息 + 角色和权限）
   */
  getProfile: async (): Promise<UserDetailResp> => {
    const response = await apiClient.get<{ data: UserDetailResp }>(
      '/auth/profile'
    );
    return response.data.data;
  },


  changePassword: (data: { old_password: string; new_password: string }) =>
    apiClient.post('/auth/change-password', data),


  refreshToken: async (token: string): Promise<{ token: string }> => {
    const response = await apiClient.post<{ token: string }>('/auth/refresh', {
      token,
    });
    return response.data;
  },
};
