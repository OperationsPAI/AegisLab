package dto

import (
	"aegis/consts"
	"encoding/json"
	"reflect"

	"github.com/sirupsen/logrus"
)

type StreamEvent struct {
	TimeStamp int              `json:"timestamp,omitempty" swaggerignore:"true"`
	TaskID    string           `json:"task_id"`
	TaskType  consts.TaskType  `json:"task_type"`
	FileName  string           `json:"file_name" swaggerignore:"true"`
	FnName    string           `json:"function_name" swaggerignore:"true"`
	Line      int              `json:"line" swaggerignore:"true"`
	EventName consts.EventType `json:"event_name"`
	Payload   any              `json:"payload,omitempty" swaggertype:"object"`
}

type DatasetOptions struct {
	Dataset string `json:"dataset"`
}

type ExecutionOptions struct {
	Algorithm   AlgorithmItem `json:"algorithm"`
	Dataset     string        `json:"dataset"`
	ExecutionID int           `json:"execution_id"`
	Timestamp   string        `json:"timestamp"`
}

type InfoPayloadTemplate struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}

type JobMessage struct {
	JobName   string `json:"job_name"`
	Namespace string `json:"namespace"`
	LogFile   string `json:"log_file,omitempty"`
}

var PayloadTypeRegistry = map[consts.EventType]reflect.Type{
	// Algorithm execution events
	consts.EventAlgoRunSucceed: reflect.TypeOf(ExecutionOptions{}),
	consts.EventAlgoRunFailed:  reflect.TypeOf(ExecutionOptions{}),

	// Dataset Build events
	consts.EventDatapackBuildSucceed: reflect.TypeOf(DatasetOptions{}),
	consts.EventDatapackBuildFailed:  reflect.TypeOf(DatasetOptions{}),

	// Task status events
	consts.EventTaskStatusUpdate: reflect.TypeOf(InfoPayloadTemplate{}),

	// K8s Job events
	consts.EventJobSucceed: reflect.TypeOf(JobMessage{}),
	consts.EventJobFailed:  reflect.TypeOf(JobMessage{}),
}

func (s *StreamEvent) ToRedisStream() map[string]any {
	payload, err := json.Marshal(s.Payload)
	if err != nil {
		logrus.Errorf("Failed to marshal payload: %v", err)
		return nil
	}

	return map[string]any{
		consts.RdbEventTaskID:   s.TaskID,
		consts.RdbEventTaskType: string(s.TaskType),
		consts.RdbEventFileName: s.FileName,
		consts.RdbEventFn:       s.FnName,
		consts.RdbEventLine:     s.Line,
		consts.RdbEventName:     string(s.EventName),
		consts.RdbEventPayload:  payload,
	}
}

func (s *StreamEvent) ToSSE() (string, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
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
