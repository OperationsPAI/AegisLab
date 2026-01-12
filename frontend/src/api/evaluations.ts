import apiClient from './client'
import type {
  DatapackEvaluationSpec,
  DatapackEvaluationResult,
} from '@/types/api'

export const evaluationApi = {
  // Evaluate datapacks
  evaluateDatapacks: (specs: DatapackEvaluationSpec[]) =>
    apiClient.post<DatapackEvaluationResult[]>('/evaluations/datapacks', { specs }),

  // Evaluate datasets
  evaluateDatasets: (specs: DatapackEvaluationSpec[]) =>
    apiClient.post<DatapackEvaluationResult[]>('/evaluations/datasets', { specs }),
}
