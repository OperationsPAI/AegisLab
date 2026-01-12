import apiClient from './client'
import type {
  Injection,
  SubmitInjectionReq,
  PaginationParams,
  PaginatedResponse,
  Label,
} from '@/types/api'

export const injectionApi = {
  // Get injections list
  getInjections: (params?: Partial<PaginationParams> & {
    lookback?: string
    fault_type?: string
    state?: number
    label?: string
    project_id?: number
  }) => apiClient.get<PaginatedResponse<Injection>>('/injections', { params }),

  // Get injection detail
  getInjection: (id: number) => apiClient.get<Injection>(`/injections/${id}`),

  // Submit injection
  submitInjection: (data: SubmitInjectionReq) =>
    apiClient.post<Injection>('/injections/inject', data),

  // Build datapack
  buildDatapack: (data: {
    benchmark: { name: string; version: string; namespace: string }
    datapack_id?: string
    dataset_id?: number
    dataset_version?: string
    pre_duration?: number
  }) => apiClient.post('/injections/build', data),

  // Get fault metadata
  getFaultMetadata: (params: { system: string }) =>
    apiClient.get('/injections/metadata', { params }),

  // Update labels
  updateLabels: (id: number, labels: Label[]) =>
    apiClient.patch(`/injections/${id}/labels`, { labels }),

  // Batch delete
  batchDelete: (ids: number[]) =>
    apiClient.post('/injections/batch-delete', { ids }),

  // Analysis
  getNoIssues: (params?: Partial<PaginationParams>) =>
    apiClient.get<PaginatedResponse<Injection>>('/injections/analysis/no-issues', { params }),

  getWithIssues: (params?: Partial<PaginationParams>) =>
    apiClient.get<PaginatedResponse<Injection>>('/injections/analysis/with-issues', { params }),
}
