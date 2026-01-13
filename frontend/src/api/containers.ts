import { Configuration, ContainersApi } from '@rcabench/client'
import axios, { type AxiosRequestConfig } from 'axios'

// Create configuration with dynamic token
const createContainerConfig = () => {
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

// Export the containers API using generated SDK where available
export const containerApi = {
  // Get containers list - using generated SDK
  getContainers: async (params?: {
    page?: number
    size?: number
    type?: number
    isPublic?: boolean
    status?: number
    projectId?: number
  }) => {
    const containersApi = new ContainersApi(createContainerConfig())
    const response = await containersApi.listContainers({
      page: params?.page,
      size: params?.size,
      type: params?.type,
      isPublic: params?.isPublic,
      status: params?.status,
    })
    return response.data
  },

  // Get container detail - using generated SDK
  getContainer: async (id: number) => {
    const containersApi = new ContainersApi(createContainerConfig())
    const response = await containersApi.getContainerById({ containerId: id })
    return response.data
  },

  // Create container - using generated SDK
  createContainer: async (data: {
    name: string
    type: number
    readme?: string
    isPublic?: boolean
    labels?: Array<{ key: string; value: string }>
  }) => {
    const containersApi = new ContainersApi(createContainerConfig())
    const response = await containersApi.createContainer({
      request: data,
    })
    return response.data
  },

  // Update container - manual endpoint (not in generated SDK)
  updateContainer: (id: number, data: Record<string, unknown>) =>
    apiClient.patch(`/containers/${id}`, data),

  // Delete container - manual endpoint (not in generated SDK)
  deleteContainer: (id: number) =>
    apiClient.delete(`/containers/${id}`),

  // Get versions - manual endpoint (not in generated SDK)
  getVersions: (containerId: number) =>
    apiClient.get(`/containers/${containerId}/versions`),

  // Create version - manual endpoint (not in generated SDK)
  createVersion: (containerId: number, data: {
    version: string
    registry: string
    repository: string
    tag: string
    command?: string
  }) =>
    apiClient.post(`/containers/${containerId}/versions`, data),

  // Update version - manual endpoint (not in generated SDK)
  updateVersion: (containerId: number, versionId: number, data: Record<string, unknown>) =>
    apiClient.patch(`/containers/${containerId}/versions/${versionId}`, data),

  // Delete version - manual endpoint (not in generated SDK)
  deleteVersion: (containerId: number, versionId: number) =>
    apiClient.delete(`/containers/${containerId}/versions/${versionId}`),

  // Build container - using generated SDK
  buildContainer: async (data: {
    containerId: number
    config: Record<string, unknown>
    dockerfile: string
  }) => {
    const containersApi = new ContainersApi(createContainerConfig())
    const response = await containersApi.buildContainerImage({
      request: {
        github_repository: (data.config.github_repository as string) || '',
        image_name: (data.config.image_name as string) || '',
        build_options: {
          dockerfile_path: data.dockerfile,
          ...(data.config.build_options as Record<string, unknown>),
        },
      },
    })
    return response.data
  },
}

export default apiClient