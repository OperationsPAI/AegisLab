import type {
  Dataset,
  DatasetVersion,
  DatasetType,
  PaginationParams,
  PaginatedResponse,
  Label,
} from '@/types/api'

import apiClient from './client'

export const datasetApi = {
  // Get datasets list
  getDatasets: (params?: Partial<PaginationParams> & {
    type?: DatasetType
    is_public?: boolean
    status?: number
    label?: string
    project_id?: number
  }) => apiClient.get<PaginatedResponse<Dataset>>('/datasets', { params }),

  // Get dataset detail
  getDataset: (id: number) => apiClient.get<Dataset>(`/datasets/${id}`),

  // Create dataset
  createDataset: (data: {
    name: string
    type: DatasetType
    description?: string
    is_public?: boolean
    labels?: Label[]
  }) => apiClient.post<Dataset>('/datasets', data),

  // Update dataset
  updateDataset: (id: number, data: Partial<Dataset>) =>
    apiClient.patch<Dataset>(`/datasets/${id}`, data),

  // Delete dataset
  deleteDataset: (id: number) => apiClient.delete(`/datasets/${id}`),

  // Get versions
  getVersions: (datasetId: number) =>
    apiClient.get<DatasetVersion[]>(`/datasets/${datasetId}/versions`),

  // Create version
  createVersion: (datasetId: number, data: {
    version: string
    file_path: string
    checksum?: string
    size?: number
  }) => apiClient.post<DatasetVersion>(`/datasets/${datasetId}/versions`, data),

  // Update version
  updateVersion: (datasetId: number, versionId: number, data: Partial<DatasetVersion>) =>
    apiClient.patch<DatasetVersion>(`/datasets/${datasetId}/versions/${versionId}`, data),

  // Delete version
  deleteVersion: (datasetId: number, versionId: number) =>
    apiClient.delete(`/datasets/${datasetId}/versions/${versionId}`),

  // Upload dataset file
  uploadFile: (datasetId: number, file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return apiClient.post<DatasetVersion>(`/datasets/${datasetId}/upload`, formData, {
      headers: {
        'Content-Type': 'multipart/form-data',
      },
    })
  },

  // Batch delete
  batchDelete: (ids: number[]) =>
    apiClient.post('/datasets/batch-delete', { ids }),
}