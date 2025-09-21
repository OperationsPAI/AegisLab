package dto

import (
	"fmt"
	"time"

	"aegis/config"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
)

// GroundTruthReq represents ground truth request
type GroundTruthReq struct {
	Datasets []string `json:"datasets" binding:"required"`
}

// GroundTruthResp represents ground truth response
type GroundTruthResp map[string]chaos.Groundtruth

// AlgorithmDatasetPair represents algorithm and dataset pair
type AlgorithmDatasetPair struct {
	Algorithm string
	Dataset   string
}

// RawDataReq represents raw data request
type RawDataReq struct {
	Pairs        []AlgorithmDatasetPair `json:"pairs" binding:"omitempty"`
	ExecutionIDs []int                  `json:"execution_ids" binding:"omitempty"`

	TimeRangeQuery
}

func (req *RawDataReq) HasPairsMode() bool {
	return len(req.Pairs) > 0
}

func (req *RawDataReq) HasExecutionMode() bool {
	return len(req.ExecutionIDs) > 0
}

func (req *RawDataReq) Validate() error {
	modeCount := 0
	if req.HasPairsMode() {
		modeCount++
	}
	if req.HasExecutionMode() {
		modeCount++
	}

	if modeCount == 0 {
		return fmt.Errorf("one of the following must be provided: pairs, (algorithms and datasets), or execution_ids")
	}

	if modeCount > 1 {
		return fmt.Errorf("only one query mode can be used at a time: pairs, (algorithms and datasets), or execution_ids")
	}

	if req.HasPairsMode() {
		for i, pair := range req.Pairs {
			if pair.Algorithm == "" {
				return fmt.Errorf("algorithm cannot be empty in pair at index %d", i)
			}

			if pair.Algorithm == config.GetString("algo.detector") {
				return fmt.Errorf("algorithm '%s' is reserved and cannot be used in pairs", config.GetString("algo.detector"))
			}

			if pair.Dataset == "" {
				return fmt.Errorf("dataset cannot be empty in pair at index %d", i)
			}
		}
	}

	if req.HasExecutionMode() {
		for i, id := range req.ExecutionIDs {
			if id <= 0 {
				return fmt.Errorf("execution ID must be greater than 0 at index %d", i)
			}
		}
	}

	return req.TimeRangeQuery.Validate()
}

// RawDataItem represents raw data item
type RawDataItem struct {
	Algorithm   string              `json:"algorithm"`
	Dataset     string              `json:"dataset"`
	ExecutionID int                 `json:"execution_id,omitempty"`
	Groundtruth chaos.Groundtruth   `json:"groundtruth"`
	Entries     []GranularityRecord `json:"entries,omitempty"`
}

// RawDataResp represents raw data response
type RawDataResp []RawDataItem

// Execution represents execution data for evaluation
type Execution struct {
	Dataset            DatasetItem         `json:"dataset"`
	GranularityRecords []GranularityRecord `json:"granularity_records"`
}

// Conclusion represents evaluation conclusion
type Conclusion struct {
	Level  string  `json:"level"`  // For example service level
	Metric string  `json:"metric"` // For example topk
	Rate   float64 `json:"rate"`
}

// EvaluateMetric represents evaluation metric function type
type EvaluateMetric func([]Execution) ([]Conclusion, error)

type AlgorithmDatasetReq struct {
	Algorithm      string `json:"algorithm" binding:"required"`
	Dataset        string `json:"dataset" binding:"required"`
	DatasetVersion string `json:"dataset_version,omitempty" form:"dataset_version"` // Dataset version (optional, defaults to "v1.0")
	Tag            string `json:"tag,omitempty" form:"tag"`
}

// DatasetEvaluationBatchReq represents request for algorithm evaluation on datasets
type DatasetEvaluationBatchReq struct {
	Items []AlgorithmDatasetReq `json:"items" binding:"required"`
}

func (req *DatasetEvaluationBatchReq) Validate() error {
	if len(req.Items) == 0 {
		return fmt.Errorf("at least one algorithm-dataset pair is required")
	}

	for i, item := range req.Items {
		if item.Algorithm == "" {
			return fmt.Errorf("algorithm cannot be empty in item at index %d", i)
		}

		if item.Dataset == "" {
			return fmt.Errorf("dataset cannot be empty in item at index %d", i)
		}

		if item.DatasetVersion == "" {
			item.DatasetVersion = "v1.0"
		}
	}

	return nil
}

// DatapackEvaluationItem represents evaluation item for a single datapack
type DatapackEvaluationItem struct {
	DatapackName      string              `json:"datapack_name"`      // Datapack name (from FaultInjectionSchedule)
	ExecutionID       int                 `json:"execution_id"`       // Execution ID
	ExecutionDuration float64             `json:"execution_duration"` // Execution duration in seconds
	Groundtruth       chaos.Groundtruth   `json:"groundtruth"`        // Ground truth for this datapack
	Predictions       []GranularityRecord `json:"predictions"`        // Algorithm predictions
	ExecutedAt        time.Time           `json:"executed_at"`        // Execution time
}

// AlgorithmDatasetResp represents response for algorithm evaluation on a dataset
type AlgorithmDatasetResp struct {
	Algorithm      string                   `json:"algorithm"`       // Algorithm name
	Dataset        string                   `json:"dataset"`         // Dataset name
	DatasetVersion string                   `json:"dataset_version"` // Dataset version
	TotalCount     int                      `json:"total_count"`     // Total number of datapacks in dataset
	ExecutedCount  int                      `json:"executed_count"`  // Number of successfully executed datapacks
	Items          []DatapackEvaluationItem `json:"items"`           // Evaluation items for each datapack
}

// DatasetEvaluationBatchResp represents response for algorithm evaluation on datasets
type DatasetEvaluationBatchResp []AlgorithmDatasetResp

// AlgorithmDatapackReq represents request for algorithm evaluation on a single datapack
type AlgorithmDatapackReq struct {
	Algorithm string `json:"algorithm" binding:"required"`
	Datapack  string `json:"datapack" binding:"required"`
	Tag       string `json:"tag,omitempty" form:"tag"` // Tag filter for filtering execution results
}

// DatapackEvaluationBatchReq represents request for algorithm evaluation on datapacks
type DatapackEvaluationBatchReq struct {
	Items []AlgorithmDatapackReq `json:"items" binding:"required"`
}

func (req *DatapackEvaluationBatchReq) Validate() error {
	if len(req.Items) == 0 {
		return fmt.Errorf("at least one algorithm-datapack pair is required")
	}

	for i, item := range req.Items {
		if item.Algorithm == "" {
			return fmt.Errorf("algorithm cannot be empty in item at index %d", i)
		}

		if item.Datapack == "" {
			return fmt.Errorf("datapack cannot be empty in item at index %d", i)
		}
	}

	return nil
}

// AlgorithmDatapackResp represents response for algorithm evaluation on a single datapack
type AlgorithmDatapackResp struct {
	Algorithm         string              `json:"algorithm"`          // Algorithm name
	Datapack          string              `json:"datapack"`           // Datapack name
	ExecutionID       int                 `json:"execution_id"`       // Execution ID (0 if no execution found)
	ExecutionDuration float64             `json:"execution_duration"` // Execution duration in seconds
	Groundtruth       chaos.Groundtruth   `json:"groundtruth"`        // Ground truth for this datapack
	Predictions       []GranularityRecord `json:"predictions"`        // Algorithm predictions
	ExecutedAt        time.Time           `json:"executed_at"`        // Execution time
	Found             bool                `json:"found"`              // Whether execution result was found
}

// DatapackEvaluationBatchResp represents response for algorithm evaluation on datapacks
type DatapackEvaluationBatchResp []AlgorithmDatapackResp

// DetectorRecord represents detector analysis result
type DetectorRecord struct {
	SpanName            string   `json:"span_name"`
	Issues              string   `json:"issue"`
	AbnormalAvgDuration *float64 `json:"abnormal_avg_duration" swaggertype:"number" example:"0.5"`
	NormalAvgDuration   *float64 `json:"normal_avg_duration" swaggertype:"number" example:"0.3"`
	AbnormalSuccRate    *float64 `json:"abnormal_succ_rate" swaggertype:"number" example:"0.8"`
	NormalSuccRate      *float64 `json:"normal_succ_rate" swaggertype:"number" example:"0.95"`
	AbnormalP90         *float64 `json:"abnormal_p90" swaggertype:"number" example:"1.2"`
	NormalP90           *float64 `json:"normal_p90" swaggertype:"number" example:"0.8"`
	AbnormalP95         *float64 `json:"abnormal_p95" swaggertype:"number" example:"1.5"`
	NormalP95           *float64 `json:"normal_p95" swaggertype:"number" example:"1.0"`
	AbnormalP99         *float64 `json:"abnormal_p99" swaggertype:"number" example:"2.0"`
	NormalP99           *float64 `json:"normal_p99" swaggertype:"number" example:"1.3"`
}

// DatapackDetectorReq represents request for detector results on datapacks
type DatapackDetectorReq struct {
	Datapacks []string `json:"datapacks" binding:"required,min=1"`
	Tag       string   `json:"tag,omitempty" form:"tag"` // Tag filter for filtering execution results
}

// DatapackDetectorItem represents detector results for a single datapack
type DatapackDetectorItem struct {
	Datapack    string           `json:"datapack"`     // Datapack name (from FaultInjectionSchedule)
	ExecutionID int              `json:"execution_id"` // Execution ID (0 if no execution found)
	Found       bool             `json:"found"`        // Whether detector result was found
	ExecutedAt  string           `json:"executed_at"`  // Execution time
	Results     []DetectorRecord `json:"results"`      // Detector analysis results
}

// DatapackDetectorResp represents response for detector results on datapacks
type DatapackDetectorResp struct {
	TotalCount    int                    `json:"total_count"`     // Total number of requested datapacks
	FoundCount    int                    `json:"found_count"`     // Number of datapacks with detector results
	NotFoundCount int                    `json:"not_found_count"` // Number of datapacks without detector results
	Items         []DatapackDetectorItem `json:"items"`           // Detector results for each datapack
}
