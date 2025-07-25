package dto

import (
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/consts"
)

var ValidFirstTaskTypes = map[consts.TaskType]struct{}{
	consts.TaskTypeBuildDataset:   {},
	consts.TaskTypeBuildImage:     {},
	consts.TaskTypeRestartService: {},
	consts.TaskTypeRunAlgorithm:   {},
}

type AnalyzeTracesReq struct {
	FirstTaskType string `form:"first_task_type" binding:"omitempty"`

	TimeRangeQuery
}

func (req *AnalyzeTracesReq) Validate() error {
	if req.FirstTaskType != "" {
		if _, exists := ValidFirstTaskTypes[consts.TaskType(req.FirstTaskType)]; !exists {
			return fmt.Errorf("Invalid event name: %s", req.FirstTaskType)
		}
	}

	return req.TimeRangeQuery.Validate()
}

func (req *AnalyzeTracesReq) Convert() (*TraceAnalyzeFilterOptions, error) {
	opts, err := req.TimeRangeQuery.Convert()
	if err != nil {
		return nil, err
	}

	return &TraceAnalyzeFilterOptions{
		FirstTaskType:     consts.TaskType(req.FirstTaskType),
		TimeFilterOptions: *opts,
	}, nil
}
