package dto

import (
	"fmt"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
)

type GroundTruthReq struct {
	Datasets []string `json:"datasets" binding:"required"`
}

type GroundTruthResp map[string]chaos.Groundtruth

type AlgorithmDatasetPair struct {
	Algorithm string
	Dataset   string
}

type RawDataReq struct {
	Pairs        []AlgorithmDatasetPair `json:"pairs" binding:"omitempty"`
	Algorithms   []string               `json:"algorithms" binding:"omitempty"`
	Datasets     []string               `json:"datasets" binding:"omitempty"`
	ExecutionIDs []int                  `json:"execution_ids" binding:"omitempty"`
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

func (req *RawDataReq) HasPairsMode() bool {
	return len(req.Pairs) > 0
}

func (req *RawDataReq) HasCartesianMode() bool {
	return len(req.Algorithms) > 0 && len(req.Datasets) > 0
}

func (req *RawDataReq) HasExecutionMode() bool {
	return len(req.ExecutionIDs) > 0
}

func (req *RawDataReq) Validate() error {
	modeCount := 0
	if req.HasPairsMode() {
		modeCount++
	}
	if req.HasCartesianMode() {
		modeCount++
	}
	if req.HasExecutionMode() {
		modeCount++
	}

	if modeCount == 0 {
		return fmt.Errorf("One of the following must be provided: pairs, (algorithms and datasets), or execution_ids")
	}

	if modeCount > 1 {
		return fmt.Errorf("Only one query mode can be used at a time: pairs, (algorithms and datasets), or execution_ids")
	}

	// 验证pairs模式下的数据
	if req.HasPairsMode() {
		for i, pair := range req.Pairs {
			if pair.Algorithm == "" {
				return fmt.Errorf("Algorithm cannot be empty in pair at index %d", i)
			}
			if pair.Dataset == "" {
				return fmt.Errorf("Dataset cannot be empty in pair at index %d", i)
			}
		}
	}

	return nil
}

type RawDataItem struct {
	Algorithm   string              `json:"algorithm"`
	Dataset     string              `json:"dataset"`
	ExecutionID int                 `json:"execution_id"`
	Entries     []GranularityRecord `json:"entries"`
	Groundtruth chaos.Groundtruth   `json:"groundtruth"`
}

type RawDataResp []RawDataItem

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

type SuccessfulExecutionItem struct {
	ID        int       `json:"id"`         // 执行ID
	Algorithm string    `json:"algorithm"`  // 算法名称
	Dataset   string    `json:"dataset"`    // 数据集名称
	CreatedAt time.Time `json:"created_at"` // 创建时间
}

type SuccessfulExecutionsResp []SuccessfulExecutionItem

type SuccessfulExecutionsReq struct {
	StartTime *time.Time `json:"start_time" form:"start_time"`
	EndTime   *time.Time `json:"end_time" form:"end_time"`
	Limit     *int       `json:"limit" form:"limit"`
	Offset    *int       `json:"offset" form:"offset"`
}
