package dto

import (
	"rcabench/consts"
)

var ValidTaskEventMap = map[consts.TaskType][]consts.EventType{
	consts.TaskTypeBuildDataset: {
		consts.EventDatapackBuildSucceed,
	},
	consts.TaskTypeCollectResult: {
		consts.EventDatapackResultCollection,
		consts.EventDatapackNoAnomaly,
		consts.EventDatapackNoDetectorData,
	},
	consts.TaskTypeFaultInjection: {
		consts.EventFaultInjectionStarted,
		consts.EventFaultInjectionCompleted,
		consts.EventFaultInjectionFailed,
	},
	consts.TaskTypeRunAlgorithm: {
		consts.EventAlgoRunSucceed,
	},
	consts.TaskTypeRestartService: {
		consts.EventNoNamespaceAvailable,
		consts.EventRestartServiceStarted,
		consts.EventRestartServiceCompleted,
		consts.EventRestartServiceFailed,
	},
}

type TraceReq struct {
	TraceID string `uri:"trace_id" binding:"required"`
}

type TraceStreamReq struct {
	LastID string `bind:"last_event_id"`
}

type GetCompletedMapReq struct {
	TimeRangeQuery
}

type GetCompletedMapResp struct {
	AnomalyTraces   []string `json:"has_anomaly"` // List of trace IDs with detected anomalies
	NoAnomalyTraces []string `json:"no_anomaly"`  // List of trace IDs without anomalies
}

func (req *GetCompletedMapReq) Validate() error {
	return req.TimeRangeQuery.Validate()
}
