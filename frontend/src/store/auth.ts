import type { LoginResp, UserInfo as User } from '@rcabench/client';
import { create } from 'zustand';

import { authApi } from '@/api/auth';

interface AuthState {
  user: User | null;
  accessToken: string | null;
  refreshToken: string | null;
  isAuthenticated: boolean;
  loading: boolean;

  // Actions
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshAccessToken: () => Promise<void>;
  loadUser: () => Promise<void>;
  setUser: (user: User | null) => void;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: localStorage.getItem('access_token'),
  refreshToken: localStorage.getItem('refresh_token'),
  isAuthenticated: !!localStorage.getItem('access_token'),
  loading: false,

  login: async (username: string, password: string) => {
    set({ loading: true });
    try {
      // console.log('Attempting login...')
      const response = await authApi.login({ username, password });
      // console.log('Login response:', response)
      // The response structure needs to be checked
      const token = (response as LoginResp)?.token;
      const user = (response as LoginResp)?.user;

      // Backend returns 'token' instead of 'access_token'
      // Store the same token as both access and refresh token for now
      if (token) {
        localStorage.setItem('access_token', token);
        localStorage.setItem('refresh_token', token);
      }

      // console.log('Setting auth state with token:', token)
      set({
        user,
        accessToken: token,
        refreshToken: token,
        isAuthenticated: true,
        loading: false,
      });
      // console.log('Login successful, isAuthenticated set to true')
    } catch (error) {
      // console.error('Login failed:', error)
      set({ loading: false });
      throw error;
    }
  },

  logout: async () => {
    try {
      await authApi.logout();
    } catch (error) {
      // console.error('Logout error:', error)
    } finally {
      localStorage.removeItem('access_token');
      localStorage.removeItem('refresh_token');

      set({
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
      });
    }
  },

  refreshAccessToken: async () => {
    const { refreshToken } = get();
    if (!refreshToken) {
      throw new Error('No refresh token available');
    }

    try {
      // Use the actual refresh endpoint instead of login
      const response = await authApi.refreshToken(refreshToken);
      const token = response?.token;

      // Backend returns single 'token' for refresh
      localStorage.setItem('access_token', token);
      localStorage.setItem('refresh_token', token);

      set({
        accessToken: token,
        refreshToken: token,
      });
    } catch (error) {
      get().logout();
      throw error;
    }
  },

  loadUser: async () => {
    const { accessToken } = get();
    if (!accessToken) return;

    set({ loading: true });
    try {
      const response = await authApi.getProfile();
      set({
        user: response,
        loading: false,
      });
    } catch (error) {
      set({ loading: false });
      get().logout();
    }
  },

  setUser: (user: User | null) => {
    set({ user });
  },
}));
