import { Configuration, AuthenticationApi } from '@rcabench/client'
import { message } from 'antd'
import axios, { type AxiosRequestConfig } from 'axios'

// Create configuration with dynamic token
const createAuthConfig = () => {
  const token = localStorage.getItem('access_token')

  return new Configuration({
    basePath: '/api/v2',
    accessToken: token ? `Bearer ${token}` : undefined,
    baseOptions: {
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    } as AxiosRequestConfig,
  })
}

// Create axios instance for manual API calls
const apiClient = axios.create({
  baseURL: '/api/v2',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Request interceptor for auth
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

// Response interceptor for token refresh
apiClient.interceptors.response.use(
  (response) => {
    return response
  },
  async (error) => {
    const originalRequest = error.config as { _retry?: boolean; headers?: Record<string, string> } // TODO: Fix this type properly

    // Handle 401 Unauthorized - refresh token
    if (error.response?.status === 401 && !originalRequest._retry) {
      originalRequest._retry = true

      try {
        const refreshToken = localStorage.getItem('refresh_token')
        if (refreshToken) {
          const response = await apiClient.post('/auth/refresh', {
            token: refreshToken,
          })

          // Backend returns single 'token' field
          const { token } = response.data
          localStorage.setItem('access_token', token)
          localStorage.setItem('refresh_token', token)

          if (originalRequest.headers) {
            originalRequest.headers.Authorization = `Bearer ${token}`
          }
          return apiClient(originalRequest)
        }
      } catch (refreshError) {
        // Refresh failed, redirect to login
        localStorage.removeItem('access_token')
        localStorage.removeItem('refresh_token')
        window.location.href = '/login'
        return Promise.reject(refreshError)
      }
    }

    // Handle other errors
    const errorMessage =
      (error.response?.data as { message?: string })?.message ||
      error.message ||
      '请求失败'

    message.error(errorMessage)
    return Promise.reject(error)
  }
)

// Export the auth API using generated SDK where available
export const authApi = {
  // Login - using generated SDK
  login: async (data: { username: string; password: string }) => {
    const authApi = new AuthenticationApi(createAuthConfig())
    const response = await authApi.login({ request: data })
    return response.data
  },

  // Register - using generated SDK
  register: async (data: { username: string; password: string; email: string }) => {
    const authApi = new AuthenticationApi(createAuthConfig())
    const response = await authApi.registerUser({ request: data })
    return response.data
  },

  // Logout - manual endpoint (not in generated SDK)
  logout: () => apiClient.post('/auth/logout'),

  // Get profile - manual endpoint (not in generated SDK)
  getProfile: async () => {
    const response = await apiClient.get('/auth/profile')
    return response.data
  },

  // Change password - manual endpoint (not in generated SDK)
  changePassword: (data: { old_password: string; new_password: string }) =>
    apiClient.post('/auth/change-password', data),

  // Refresh token - manual endpoint (not in generated SDK)
  refreshToken: async (token: string) => {
    const response = await apiClient.post<{ token: string }>('/auth/refresh', { token })
    return response.data
  },
}

export default apiClient