import apiClient from './client'
import type {
  Project,
  PaginationParams,
  PaginatedResponse,
  Label,
} from '@/types/api'

export const projectApi = {
  // Get projects list
  getProjects: (params?: Partial<PaginationParams> & {
    is_public?: boolean
    status?: number
    label?: string
  }) => apiClient.get<PaginatedResponse<Project>>('/projects', { params }),

  // Get project detail
  getProject: (id: number) => apiClient.get<Project>(`/projects/${id}`),

  // Create project
  createProject: (data: {
    name: string
    description?: string
    is_public?: boolean
    labels?: Label[]
  }) => apiClient.post<Project>('/projects', data),

  // Update project
  updateProject: (id: number, data: Partial<Project>) =>
    apiClient.patch<Project>(`/projects/${id}`, data),

  // Delete project
  deleteProject: (id: number) => apiClient.delete(`/projects/${id}`),

  // Manage labels
  updateLabels: (id: number, labels: Label[]) =>
    apiClient.patch(`/projects/${id}/labels`, { labels }),
}
