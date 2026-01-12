import apiClient from './client'
import type {
  Execution,
  PaginationParams,
  PaginatedResponse,
  Label,
  GranularityResults,
  DetectorResult,
} from '@/types/api'

export const executionApi = {
  // Get executions list
  getExecutions: (params?: Partial<PaginationParams> & {
    state?: number
    status?: number
    label?: string
    project_id?: number
    datapack_id?: string
  }) => apiClient.get<PaginatedResponse<Execution>>('/executions', { params }),

  // Get execution detail
  getExecution: (id: number) => apiClient.get<Execution>(`/executions/${id}`),

  // Execute algorithm
  executeAlgorithm: (data: {
    algorithm_name: string
    algorithm_version: string
    datapack_id: string
    labels?: Label[]
  }) => apiClient.post<Execution>('/executions/execute', data),

  // Upload detector results
  uploadDetectorResults: (id: number, results: DetectorResult[]) =>
    apiClient.post(`/executions/${id}/detector_results`, { results }),

  // Upload granularity results
  uploadGranularityResults: (id: number, results: GranularityResults) =>
    apiClient.post(`/executions/${id}/granularity_results`, results),

  // Update labels
  updateLabels: (id: number, labels: Label[]) =>
    apiClient.patch(`/executions/${id}/labels`, { labels }),

  // Batch delete
  batchDelete: (ids: number[]) =>
    apiClient.post('/executions/batch-delete', { ids }),
}
