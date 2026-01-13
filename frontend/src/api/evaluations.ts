import {
  Configuration,
  type EvaluateDatapackSpec,
  type EvaluateDatasetSpec,
  EvaluationsApi,
} from '@rcabench/client';
import axios, { type AxiosRequestConfig } from 'axios';

// Create configuration with dynamic token
const createEvaluationConfig = () => {
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

// Export the evaluations API using generated SDK
export const evaluationApi = {
  // Evaluate datapacks - using generated SDK
  evaluateDatapacks: async (specs: Array<Record<string, unknown>>) => {
    const evaluationsApi = new EvaluationsApi(createEvaluationConfig());
    const response = await evaluationsApi.evaluateAlgorithmOnDatapacks({
      request: {
        specs: specs as unknown as EvaluateDatapackSpec[],
      },
    });
    return response;
  },

  // Evaluate datasets - using generated SDK
  evaluateDatasets: async (specs: Array<Record<string, unknown>>) => {
    const evaluationsApi = new EvaluationsApi(createEvaluationConfig());
    const response = await evaluationsApi.evaluateAlgorithmOnDatasets({
      request: {
        specs: specs as unknown as EvaluateDatasetSpec[],
      },
    });
    return response;
  },
};

export default apiClient;
