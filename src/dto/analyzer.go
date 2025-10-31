package dto

import (
	"fmt"

	"aegis/consts"
)

var ValidFirstTaskTypes = map[consts.TaskType]struct{}{
	consts.TaskTypeBuildDataset:   {},
	consts.TaskTypeBuildContainer: {},
	consts.TaskTypeRestartService: {},
	consts.TaskTypeRunAlgorithm:   {},
}

type AnalyzeInjectionsReq struct {
	InjectionFilterOptions
	TimeRangeQuery
}

func (req *AnalyzeInjectionsReq) ToListInjectionsReq() *ListInjectionsReq {
	return &ListInjectionsReq{
		InjectionFilterOptions: req.InjectionFilterOptions,
		TimeRangeQuery:         req.TimeRangeQuery,
		ListOptionsQuery: ListOptionsQuery{
			SortField: "created_at",
			SortOrder: "desc",
			Limit:     0,
		},
		PaginationQuery: PaginationQuery{
			PageNum:  0,
			PageSize: 0,
		},
	}
}

func (req *AnalyzeInjectionsReq) Validate() error {
	if err := req.InjectionFilterOptions.Validate(); err != nil {
		return err
	}

	return req.TimeRangeQuery.Validate()
}

type AnalyzeTracesReq struct {
	FirstTaskType string `form:"first_task_type" binding:"omitempty"`

	TimeRangeQuery
}

func (req *AnalyzeTracesReq) Validate() error {
	if req.FirstTaskType != "" {
		if _, exists := ValidFirstTaskTypes[consts.TaskType(req.FirstTaskType)]; !exists {
			return fmt.Errorf("invalid event name: %s", req.FirstTaskType)
		}
	}

	return req.TimeRangeQuery.Validate()
}

type PairStats struct {
	Name      string
	InDegree  int
	OutDegree int
}

type ServiceCoverageItem struct {
	Num        int
	NotCovered []string
	Coverage   float64
}

type AttributeCoverageItem struct {
	Num      int
	Coverage float64
}

type InjectionDiversity struct {
	FaultDistribution   map[string]int                              `json:"fault_distribution"`
	ServiceDistribution map[string]int                              `json:"service_distribution"`
	PairDistribution    []PairStats                                 `json:"pair_distribution"`
	ServiceCoverages    map[string]ServiceCoverageItem              `json:"fault_service_coverages"`
	AttributeCoverages  map[string]map[string]AttributeCoverageItem `json:"attribute_coverages"`
}

type InjectionStats struct {
	Diversity InjectionDiversity `json:"diversity"`
}

type AnalyzeInjectionsResp struct {
	Efficiency string                    `json:"efficiency"`
	Stats      map[string]InjectionStats `json:"stats"`
}

type TraceStats struct {
	Total       int     `json:"total"`
	AvgDuration float64 `json:"avg_duration"`
	MinDuration float64 `json:"min_duration"`
	MaxDuration float64 `json:"max_duration"`

	EndCountMap          map[consts.TaskType]map[string]int     `json:"end_count_map"`
	TraceStatusTimeMap   map[string]map[consts.TaskType]float64 `json:"trace_status_time_map"`
	TraceCompletedList   []string                               `json:"trace_completed_list"`
	FaultInjectionTraces []string                               `json:"fault_injection_traces"`
	TraceErrors          any                                    `json:"trace_errors"`
}
