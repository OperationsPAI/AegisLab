import { create } from 'zustand'
import type { User } from '@/types/api'
import { authApi } from '@/api/auth'

interface AuthState {
  user: User | null
  accessToken: string | null
  refreshToken: string | null
  isAuthenticated: boolean
  loading: boolean

  // Actions
  login: (username: string, password: string) => Promise<void>
  logout: () => Promise<void>
  refreshAccessToken: () => Promise<void>
  loadUser: () => Promise<void>
  setUser: (user: User | null) => void
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  accessToken: localStorage.getItem('access_token'),
  refreshToken: localStorage.getItem('refresh_token'),
  isAuthenticated: !!localStorage.getItem('access_token'),
  loading: false,

  login: async (username: string, password: string) => {
    set({ loading: true })
    try {
      const response = await authApi.login({ username, password })
      const { access_token, refresh_token, user } = response.data

      localStorage.setItem('access_token', access_token)
      localStorage.setItem('refresh_token', refresh_token)

      set({
        user,
        accessToken: access_token,
        refreshToken: refresh_token,
        isAuthenticated: true,
        loading: false,
      })
    } catch (error) {
      set({ loading: false })
      throw error
    }
  },

  logout: async () => {
    try {
      await authApi.logout()
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      localStorage.removeItem('access_token')
      localStorage.removeItem('refresh_token')

      set({
        user: null,
        accessToken: null,
        refreshToken: null,
        isAuthenticated: false,
      })
    }
  },

  refreshAccessToken: async () => {
    const { refreshToken } = get()
    if (!refreshToken) {
      throw new Error('No refresh token available')
    }

    try {
      const response = await authApi.login({
        username: '',
        password: '',
      })
      const { access_token, refresh_token } = response.data

      localStorage.setItem('access_token', access_token)
      localStorage.setItem('refresh_token', refresh_token)

      set({
        accessToken: access_token,
        refreshToken: refresh_token,
      })
    } catch (error) {
      get().logout()
      throw error
    }
  },

  loadUser: async () => {
    const { accessToken } = get()
    if (!accessToken) return

    set({ loading: true })
    try {
      const response = await authApi.getProfile()
      set({
        user: response.data,
        loading: false,
      })
    } catch (error) {
      set({ loading: false })
      get().logout()
    }
  },

  setUser: (user: User | null) => {
    set({ user })
  },
}))
