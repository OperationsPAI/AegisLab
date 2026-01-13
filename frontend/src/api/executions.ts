import type { PaginationParams, PaginatedResponse, Execution } from '@/types/api';
import apiClient from './client';

export const executionApi = {
  // Get executions list - manual API call to match project types
  getExecutions: (
    params?: Partial<PaginationParams> & {
      state?: number;
      status?: number;
      labels?: string[];
      projectId?: number;
      datapackId?: string;
    }
  ) => apiClient.get<PaginatedResponse<Execution>>('/executions', { params }),

  // Get execution detail - manual API call
  getExecution: (id: number) => apiClient.get<Execution>(`/executions/${id}`),

  // Execute algorithm - manual API call
  executeAlgorithm: (data: {
    algorithmName: string;
    algorithmVersion: string;
    datapackId: string;
    labels?: Array<{ key: string; value: string }>;
  }) =>
    apiClient.post('/executions', {
      project_name: 'default',
      specs: [
        {
          algorithm: {
            name: data.algorithmName,
            version: data.algorithmVersion,
          },
          datapack: data.datapackId,
        },
      ],
      labels: data.labels,
    }),

  // Upload detector results - manual endpoint (not in generated SDK)
  uploadDetectorResults: (
    id: number,
    results: Array<Record<string, unknown>>
  ) => apiClient.post(`/executions/${id}/detector_results`, { results }),

  // Upload granularity results - manual endpoint (not in generated SDK)
  uploadGranularityResults: (id: number, results: Record<string, unknown>) =>
    apiClient.post(`/executions/${id}/granularity_results`, results),

  // Update labels - manual endpoint (not in generated SDK)
  updateLabels: (id: number, labels: Array<{ key: string; value: string }>) =>
    apiClient.patch(`/executions/${id}/labels`, { labels }),

  // Batch delete - manual endpoint (not in generated SDK)
  batchDelete: (ids: number[]) =>
    apiClient.post('/executions/batch-delete', { ids }),
};

export default apiClient;
