/**
 * 认证 API
 * 使用 @rcabench/client SDK，手工实现缺失的端点
 */
import {
  AuthenticationApi,
  type LoginReq,
  type LoginResp,
  type RegisterReq,
  type UserInfo,
} from '@rcabench/client';
import { apiClient, createApiConfig } from './config';

export const authApi = {
  // ==================== SDK 方法 ====================

  /**
   * 登录 - 使用 SDK
   */
  login: async (data: LoginReq): Promise<LoginResp | undefined> => {
    const api = new AuthenticationApi(createApiConfig());
    const response = await api.login({ request: data });
    return response.data.data;
  },

  /**
   * 注册 - 使用 SDK
   */
  register: async (data: RegisterReq): Promise<UserInfo | undefined> => {
    const api = new AuthenticationApi(createApiConfig());
    const response = await api.registerUser({ request: data });
    return response.data.data;
  },

  // ==================== 手工实现 (SDK 缺失) ====================

  /**
   * 登出 - 手工实现 (SDK 缺失)
   */
  logout: () => apiClient.post('/auth/logout'),

  /**
   * 获取用户信息 - 手工实现 (SDK 缺失)
   */
  getProfile: async (): Promise<UserInfo> => {
    const response = await apiClient.get<{ data: UserInfo }>('/auth/profile');
    return response.data.data;
  },

  /**
   * 修改密码 - 手工实现 (SDK 缺失)
   */
  changePassword: (data: { old_password: string; new_password: string }) =>
    apiClient.post('/auth/change-password', data),

  /**
   * 刷新 token - 手工实现 (SDK 缺失)
   */
  refreshToken: async (token: string): Promise<{ token: string }> => {
    const response = await apiClient.post<{ token: string }>('/auth/refresh', {
      token,
    });
    return response.data;
  },
};
