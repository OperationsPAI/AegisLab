/**
 * 集中的 API 配置
 * 所有 API 模块共享此配置，避免代码重复
 */
import { Configuration } from '@rcabench/client';
import { message } from 'antd';
import axios, { type AxiosRequestConfig } from 'axios';


export const createApiConfig = (): Configuration => {
  const token = localStorage.getItem('access_token');

  return new Configuration({
    basePath: '',
    // SDK 使用 apiKey 设置 Authorization header (需带 Bearer 前缀)
    apiKey: token ? `Bearer ${token}` : undefined,
    baseOptions: {
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    } as AxiosRequestConfig,
  });
};

/**
 * 共享的 Axios 实例
 * 用于 SDK 中缺失的手工 API 调用
 */
export const apiClient = axios.create({
  baseURL: '/api/v2',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// 请求拦截器：添加 Authorization header
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// 响应拦截器：处理 401 和错误消息
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config as {
      _retry?: boolean;
      headers?: Record<string, string>;
    };

    // 处理 401 - 尝试刷新 token
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
          const response = await apiClient.post('/auth/refresh', {
            token: refreshToken,
          });

          const { token } = response.data;
          localStorage.setItem('access_token', token);
          localStorage.setItem('refresh_token', token);

          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${token}`;
          }
          return apiClient(originalRequest);
        }
      } catch (refreshError) {
        // 刷新失败，重定向到登录页
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    // 处理其他错误
    const errorMessage =
      (error.response?.data as { message?: string })?.message ||
      error.message ||
      '请求失败';

    message.error(errorMessage);
    return Promise.reject(error);
  }
);

export default apiClient;
