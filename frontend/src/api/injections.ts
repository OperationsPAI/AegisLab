import {
  Configuration,
  InjectionsApi,
  type ListInjectionsType,
  type SubmitInjectionReq,
} from '@rcabench/client';
import axios, { type AxiosRequestConfig } from 'axios';

import type { GenericResponse } from '../types/api';

// Create configuration with dynamic token
const createInjectionConfig = () => {
  const token = localStorage.getItem('access_token');

  return new Configuration({
    basePath: '/api/v2',
    accessToken: token ? `Bearer ${token}` : undefined,
    baseOptions: {
      timeout: 30000,
      headers: {
        'Content-Type': 'application/json',
      },
    } as AxiosRequestConfig,
  });
};

// Create axios instance for manual API calls
const apiClient = axios.create({
  baseURL: '/api/v2',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor for auth
apiClient.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('access_token');
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

// Export the injections API using generated SDK where available
export const injectionApi = {
  // Get injections list - using generated SDK
  getInjections: async (params?: {
    page?: number;
    size?: number;
    lookback?: string;
    fault_type?: string;
    state?: number;
    label?: string;
    project_id?: number;
  }) => {
    const injectionsApi = new InjectionsApi(createInjectionConfig());
    const response = await injectionsApi.listInjections({
      page: params?.page,
      size: params?.size,
      type: params?.fault_type as unknown as ListInjectionsType,
      state: params?.state,
      labels: params?.label ? [params.label] : undefined,
    });
    return response.data;
  },

  // Get injection detail - using generated SDK
  getInjection: async (id: number) => {
    const injectionsApi = new InjectionsApi(createInjectionConfig());
    const response = await injectionsApi.getInjectionById({ id });
    return response.data;
  },

  // Submit injection - using generated SDK
  submitInjection: async (data: Record<string, unknown>) => {
    const injectionsApi = new InjectionsApi(createInjectionConfig());
    const response = await injectionsApi.injectFault({
      body: data as unknown as SubmitInjectionReq,
    });
    return response.data;
  },

  // Build datapack - using generated SDK
  buildDatapack: async (data: {
    benchmark: { name: string; version: string; namespace: string };
    datapack_id?: string;
    dataset_id?: number;
    dataset_version?: string;
    pre_duration?: number;
  }) => {
    const injectionsApi = new InjectionsApi(createInjectionConfig());
    const response = await injectionsApi.buildDatapack({
      body: {
        project_name: 'default',
        specs: [
          {
            benchmark: data.benchmark,
          },
        ],
      },
    });
    return response.data;
  },

  // Get fault metadata - using generated SDK
  getFaultMetadata: async (params: { system: string }) => {
    const injectionsApi = new InjectionsApi(createInjectionConfig());
    const response = await injectionsApi.getInjectionMetadata({
      system: params.system as any,
    });
    return response;
  },

  // Update labels - manual endpoint (not in generated SDK)
  updateLabels: (id: number, labels: Array<{ key: string; value: string }>) =>
    apiClient.patch(`/injections/${id}/labels`, { labels }),

  // Batch delete - manual endpoint (not in generated SDK)
  batchDelete: (ids: number[]) =>
    apiClient.post<GenericResponse<null>>('/injections/batch-delete', { ids }),

  // Analysis - using generated SDK
  getNoIssues: async () => {
    const injectionsApi = new InjectionsApi(createInjectionConfig());
    const response = await injectionsApi.listFailedInjections();
    return response.data;
  },

  // Analysis - using generated SDK
  getWithIssues: async () => {
    const injectionsApi = new InjectionsApi(createInjectionConfig());
    const response = await injectionsApi.listSuccessfulInjections();
    return response.data;
  },

  // Create injection - manual endpoint for visual injection creation
  createInjection: async (data: {
    project_id: number;
    name: string;
    description?: string;
    container_config: {
      pedestal_container_id: number;
      benchmark_container_id: number;
      algorithm_container_ids: number[];
    };
    fault_matrix: Array<
      Array<{
        id: number;
        name: string;
        type: string;
        category?: string;
        parameters?: unknown[];
      }>
    >;
    experiment_params: {
      duration: number;
      interval: number;
      parallel: boolean;
    };
    tags?: string[];
  }) => {
    const response = await apiClient.post('/injections', data);
    return response.data;
  },

  // Get fault types - manual endpoint
  getFaultTypes: async () => {
    const response = await apiClient.get('/injections/fault-types');
    return response.data;
  },
};

export default apiClient;
