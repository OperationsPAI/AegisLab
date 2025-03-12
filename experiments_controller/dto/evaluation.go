package dto

import "github.com/CUHK-SE-Group/rcabench/executor"

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
	Algorithm   string                `json:"algorithm"`
	Executions  []executor.Execution  `json:"executions"`
	Conclusions []executor.Conclusion `json:"conclusions"`
}
