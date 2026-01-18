import { SystemApi } from '@rcabench/client';
import { createApiConfig } from './config';

const systemApi = new SystemApi(createApiConfig());

export const getSystemMetrics = () => systemApi.getSystemMetrics();

export const getSystemMetricsHistory = () => systemApi.getSystemMetricsHistory();

export default {
  getSystemMetrics,
  getSystemMetricsHistory,
};
