package dto

import (
	"time"

	"github.com/CUHK-SE-Group/rcabench/database"
	"github.com/CUHK-SE-Group/rcabench/executor"
)

type DatasetDeleteReq struct {
	IDs []int `form:"ids" binding:"required"`
}

type DatasetDownloadReq struct {
	GroupIDs []string `form:"group_ids" binding:"required"`
}

type DatasetItem struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type DatasetListReq struct {
	PaginationReq
}

type DatasetPayload struct {
	Benchmark   string     `json:"benchmark"`
	DatasetName string     `json:"dataset"`
	Namespace   string     `json:"namespace"`
	PreDuration int        `json:"pre_duration"`
	Service     string     `json:"service"`
	StartTime   *time.Time `json:"start_time,omitempty"`
	EndTime     *time.Time `json:"end_time,omitempty"`
}

type QueryDatasetReq struct {
	Name string `form:"dataset" binding:"required"`
	Sort string `form:"sort"`
}

type InjectionParam struct {
	Duration  int            `json:"duration"`
	FaultType string         `json:"fault_type"`
	Namespace string         `json:"namespace"`
	Pod       string         `json:"pod"`
	Spec      map[string]int `json:"spec"`
}

type DetectorRecord struct {
	SpanName    string   `json:"span_name"`
	Issues      string   `json:"issue"`
	AvgDuration *float64 `json:"avg_duration"`
	SuccRate    *float64 `json:"succ_rate"`
	P90         *float64 `json:"p90"`
	P95         *float64 `json:"p95"`
	P99         *float64 `json:"p99"`
}

type ExecutionRecord struct {
	Algorithm          string              `json:"algorithm"`
	GranularityResults []GranularityRecord `json:"granularity_results"`
}

type GranularityRecord struct {
	Level      string  `json:"level"`
	Result     string  `json:"result"`
	Rank       int     `json:"rank"`
	Confidence float64 `json:"confidence"`
}

type QueryDatasetResp struct {
	Param            InjectionParam    `json:"param"`
	StartTime        time.Time         `json:"start_time"`
	EndTime          time.Time         `json:"end_time"`
	DetectorResult   DetectorRecord    `json:"detector_result"`
	ExecutionResults []ExecutionRecord `json:"execution_results"`
}

func ConvertToDatasetItem(f *database.FaultInjectionSchedule) *DatasetItem {
	return &DatasetItem{
		ID:   f.ID,
		Name: f.InjectionName,
	}
}

var DatasetStatusMap = map[int]string{
	executor.DatasetInitial: "initial",
	executor.DatasetSuccess: "success",
	executor.DatasetFailed:  "failed",
	executor.DatesetDeleted: "deleted",
}
