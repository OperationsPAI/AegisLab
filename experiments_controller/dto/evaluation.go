package dto

import (
	"github.com/CUHK-SE-Group/rcabench/database"
)

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
	Dataset            database.FaultInjectionSchedule `json:"dataset"`
	DetectorResult     database.Detector               `json:"detector_result"`
	ExecutionRecord    database.ExecutionResult        `json:"execution_record"`
	GranularityResults []database.GranularityResult    `json:"granularity_results"`
}

type Conclusion struct {
	Level  string  `json:"level"`  // 例如 service level
	Metric string  `json:"metric"` // 例如 topk
	Rate   float64 `json:"rate"`
}

type EvaluateMetric func([]Execution) ([]*Conclusion, error)
