import { Configuration, ExecutionsApi } from '@rcabench/client'
import axios, { type AxiosRequestConfig } from 'axios'

// Create configuration with dynamic token
const createExecutionConfig = () => {
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

// Export the executions API using generated SDK where available
export const executionApi = {
  // Get executions list - using generated SDK
  getExecutions: async (params?: {
    page?: number
    size?: number
    state?: number
    status?: number
    labels?: string[]
    projectId?: number
    datapackId?: string
  }) => {
    const executionsApi = new ExecutionsApi(createExecutionConfig())
    const response = await executionsApi.listExecutions({
      page: params?.page,
      size: params?.size,
      state: params?.state,
      status: params?.status,
      labels: params?.labels,
    })
    return response
  },

  // Get execution detail - using generated SDK
  getExecution: async (id: number) => {
    const executionsApi = new ExecutionsApi(createExecutionConfig())
    const response = await executionsApi.getExecutionById({ id })
    return response.data
  },

  // Execute algorithm - using generated SDK
  executeAlgorithm: async (data: {
    algorithmName: string
    algorithmVersion: string
    datapackId: string
    labels?: Array<{ key: string; value: string }>
  }) => {
    const executionsApi = new ExecutionsApi(createExecutionConfig())
    const response = await executionsApi.runAlgorithm({
      request: {
        project_name: 'default',
        specs: [{
          algorithm: {
            name: data.algorithmName,
            version: data.algorithmVersion,
          },
          datapack: data.datapackId,
        }],
        labels: data.labels,
      },
    })
    return response.data
  },

  // Upload detector results - manual endpoint (not in generated SDK)
  uploadDetectorResults: (id: number, results: Array<Record<string, unknown>>) =>
    apiClient.post(`/executions/${id}/detector_results`, { results }),

  // Upload granularity results - manual endpoint (not in generated SDK)
  uploadGranularityResults: (id: number, results: Record<string, unknown>) =>
    apiClient.post(`/executions/${id}/granularity_results`, results),

  // Update labels - manual endpoint (not in generated SDK)
  updateLabels: (id: number, labels: Array<{ key: string; value: string }>) =>
    apiClient.patch(`/executions/${id}/labels`, { labels }),

  // Batch delete - manual endpoint (not in generated SDK)
  batchDelete: (ids: number[]) =>
    apiClient.post('/executions/batch-delete', { ids }),
}

export default apiClient