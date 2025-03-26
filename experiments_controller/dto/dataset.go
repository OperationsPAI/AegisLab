package dto

import (
	"time"

	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/database"
)

type DatasetDeleteReq struct {
	Names []string `form:"names" binding:"required,min=1,dive,required,max=64"`
}

type DatasetDeleteResp struct {
	SuccessCount int64    `json:"success_count"`
	FailedNames  []string `json:"failed_names"`
}

type DatasetDownloadReq struct {
	GroupIDs []string `form:"group_ids" binding:"required"`
}

type DatasetItem struct {
	Name        string         `json:"name"`
	Param       InjectionParam `json:"param"`
	Preduration int            `json:"pre_duration"`
	StartTime   time.Time      `json:"start_time"`
	EndTime     time.Time      `json:"end_time"`
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
	Name string `form:"name" binding:"required,max=64"`
	Sort string `form:"sort" binding:"oneof=desc asc"`
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
	P90         *float64 `json:"P90"`
	P95         *float64 `json:"P95"`
	P99         *float64 `json:"P99"`
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
	DatasetItem
	DetectorResult   DetectorRecord    `json:"detector_result"`
	ExecutionResults []ExecutionRecord `json:"execution_results"`
}

func ConvertToDatasetItem(record database.FaultInjectionSchedule, param InjectionParam) DatasetItem {
	return DatasetItem{
		Name:        record.InjectionName,
		Param:       param,
		Preduration: record.PreDuration,
		StartTime:   record.StartTime,
		EndTime:     record.EndTime,
	}
}

var DatasetStatusMap = map[int]string{
	consts.DatasetInitial:       "initial",
	consts.DatasetInjectSuccess: "inject_success",
	consts.DatasetInjectFailed:  "inject_failed",
	consts.DatasetBuildSuccess:  "build_success",
	consts.DatasetBuildFailed:   "build_failed",
	consts.DatasetDeleted:       "deleted",
}
