/**
 * Datapack API
 * Datapacks are fault injections that have been built and are ready for algorithm execution
 */
import type {
  DatapackState,
  InjectionDetailResp,
  ListInjectionResp,
} from '@rcabench/client';

import { apiClient } from './config';

export const datapackApi = {
  /**
   * Get datapacks list with filtering
   */
  getDatapacks: async (params?: {
    page?: number;
    size?: number;
    state?: DatapackState;
    benchmark?: string;
    fault_type?: string;
  }): Promise<ListInjectionResp> => {
    const response = await apiClient.get('/datapacks', { params });
    return response.data.data;
  },

  /**
   * Get datapack details by ID
   */
  getDatapack: async (id: number): Promise<InjectionDetailResp> => {
    const response = await apiClient.get(`/datapacks/${id}`);
    return response.data.data;
  },

  /**
   * Get datapack metadata (trace count, size, etc.)
   */
  getDatapackMetadata: async (id: number): Promise<InjectionDetailResp> => {
    const response = await apiClient.get(`/datapacks/${id}/metadata`);
    return response.data.data;
  },
};
