import apiClient from './client'
import type {
  LoginReq,
  LoginRes,
  User,
} from '@/types/api'

export const authApi = {
  // Login
  login: (data: LoginReq) =>
    apiClient.post<LoginRes>('/auth/login', data),

  // Register
  register: (data: { username: string; password: string; email: string }) =>
    apiClient.post<LoginRes>('/auth/register', data),

  // Logout
  logout: () => apiClient.post('/auth/logout'),

  // Get profile
  getProfile: () => apiClient.get<User>('/auth/profile'),

  // Change password
  changePassword: (data: { old_password: string; new_password: string }) =>
    apiClient.post('/auth/change-password', data),
}
