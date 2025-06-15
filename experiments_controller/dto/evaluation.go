package dto

type RawDataReq struct {
	Algorithms []string `form:"algorithms" bind:"required"`
	Datasets   []string `form:"datasets" bind:"required"`
}

type AlgorithmDatasetPair struct {
	Algorithm string
	Dataset   string
}

func (r *RawDataReq) CartesianProduct() []AlgorithmDatasetPair {
	var result []AlgorithmDatasetPair
	for _, algorithm := range r.Algorithms {
		for _, dataset := range r.Datasets {
			result = append(result, AlgorithmDatasetPair{
				Algorithm: algorithm,
				Dataset:   dataset,
			})
		}
	}

	return result
}

type RawDataItem struct {
	Algorithm   string              `json:"algorithm"`
	Dataset     string              `json:"dataset"`
	Entries     []GranularityRecord `json:"entries"`
	GroundTruth string              `json:"ground_truth"`
}

type EvaluationListReq struct {
	ExecutionIDs []int    `form:"execution_ids"`
	Algoritms    []string `form:"algorithms"`
	Levels       []string `form:"levels"`
	Metrics      []string `form:"metrics"`
	Rank         *int     `form:"rank"`
}

type EvaluationListResp struct {
	Results []EvaluationItem `json:"results"`
}

type EvaluationItem struct {
	Algorithm   string       `json:"algorithm"`
	Executions  []Execution  `json:"executions"`
	Conclusions []Conclusion `json:"conclusions"`
}

type Execution struct {
	Dataset            DatasetItem         `json:"dataset"`
	GranularityRecords []GranularityRecord `json:"granularity_records"`
}

type Conclusion struct {
	Level  string  `json:"level"`  // 例如 service level
	Metric string  `json:"metric"` // 例如 topk
	Rate   float64 `json:"rate"`
}

type EvaluateMetric func([]Execution) ([]Conclusion, error)
