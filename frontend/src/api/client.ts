/**
 * Unified API client configuration for @rcabench/client SDK
 * This module provides a shared axios instance and configuration factory
 * for all SDK API classes.
 */
import { Configuration } from '@rcabench/client';
import { message } from 'antd';
import axios, { type AxiosInstance, type AxiosRequestConfig } from 'axios';

// Base configuration
const BASE_PATH = '/api/v2';
const TIMEOUT = 30000;

// Create shared axios instance
export const axiosInstance: AxiosInstance = axios.create({
  baseURL: BASE_PATH,
  timeout: TIMEOUT,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor for auth
axiosInstance.interceptors.request.use(
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

// Response interceptor for token refresh
axiosInstance.interceptors.response.use(
  (response) => {
    return response;
  },
  async (error) => {
    const originalRequest = error.config as AxiosRequestConfig & {
      _retry?: boolean;
    };

    // Handle 401 Unauthorized - refresh token
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true;

      try {
        const refreshToken = localStorage.getItem('refresh_token');
        if (refreshToken) {
          const response = await axios.post(`${BASE_PATH}/auth/refresh`, {
            token: refreshToken,
          });

          // Backend returns single 'token' field
          const { token } = response.data;
          localStorage.setItem('access_token', token);
          localStorage.setItem('refresh_token', token);

          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${token}`;
          }
          return axiosInstance(originalRequest);
        }
      } catch (refreshError) {
        // Refresh failed, redirect to login
        localStorage.removeItem('access_token');
        localStorage.removeItem('refresh_token');
        window.location.href = '/login';
        return Promise.reject(refreshError);
      }
    }

    // Handle other errors
    const errorMessage =
      (error.response?.data as { message?: string })?.message ||
      error.message ||
      '请求失败';

    message.error(errorMessage);
    return Promise.reject(error);
  }
);

/**
 * Create SDK Configuration with current auth token
 * Use this function to create Configuration for SDK API classes
 */
export const createApiConfig = (): Configuration => {
  const token = localStorage.getItem('access_token');

  return new Configuration({
    basePath: BASE_PATH,
    accessToken: token ? `Bearer ${token}` : undefined,
    baseOptions: {
      timeout: TIMEOUT,
      headers: {
        'Content-Type': 'application/json',
      },
    } as AxiosRequestConfig,
  });
};

export default axiosInstance;
