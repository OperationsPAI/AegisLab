import apiClient from './client'
import type {
  Container,
  ContainerVersion,
  ContainerType,
  PaginationParams,
  PaginatedResponse,
  Label,
} from '@/types/api'

export const containerApi = {
  // Get containers list
  getContainers: (params?: Partial<PaginationParams> & {
    type?: ContainerType
    is_public?: boolean
    status?: number
    label?: string
    project_id?: number
  }) => apiClient.get<PaginatedResponse<Container>>('/containers', { params }),

  // Get container detail
  getContainer: (id: number) => apiClient.get<Container>(`/containers/${id}`),

  // Create container
  createContainer: (data: {
    name: string
    type: ContainerType
    readme?: string
    is_public?: boolean
    labels?: Label[]
  }) => apiClient.post<Container>('/containers', data),

  // Update container
  updateContainer: (id: number, data: Partial<Container>) =>
    apiClient.patch<Container>(`/containers/${id}`, data),

  // Delete container
  deleteContainer: (id: number) => apiClient.delete(`/containers/${id}`),

  // Get versions
  getVersions: (containerId: number) =>
    apiClient.get<ContainerVersion[]>(`/containers/${containerId}/versions`),

  // Create version
  createVersion: (containerId: number, data: {
    version: string
    registry: string
    repository: string
    tag: string
    command?: string
  }) => apiClient.post<ContainerVersion>(`/containers/${containerId}/versions`, data),

  // Update version
  updateVersion: (containerId: number, versionId: number, data: Partial<ContainerVersion>) =>
    apiClient.patch<ContainerVersion>(`/containers/${containerId}/versions/${versionId}`, data),

  // Delete version
  deleteVersion: (containerId: number, versionId: number) =>
    apiClient.delete(`/containers/${containerId}/versions/${versionId}`),

  // Build container
  buildContainer: (data: {
    container_id: number
    config: unknown
    dockerfile: string
  }) => apiClient.post('/containers/build', data),
}
