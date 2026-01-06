package dto

import (
	"fmt"

	"aegis/consts"
)

var ValidFirstTaskTypes = map[consts.TaskType]struct{}{
	consts.TaskTypeBuildContainer:  {},
	consts.TaskTypeRestartPedestal: {},
	consts.TaskTypeBuildDatapack:   {},
	consts.TaskTypeRunAlgorithm:    {},
}

type AnalyzeTracesReq struct {
	FirstTaskType *consts.TaskType `form:"first_task_type" binding:"omitempty"`

	TimeRangeQuery
}

func (req *AnalyzeTracesReq) Validate() error {
	if req.FirstTaskType != nil {
		if _, exists := ValidFirstTaskTypes[*req.FirstTaskType]; !exists {
			return fmt.Errorf("invalid event name: %d", req.FirstTaskType)
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
