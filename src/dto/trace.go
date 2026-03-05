package dto

import (
	"aegis/consts"
	"aegis/database"
	"aegis/utils"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type TraceStreamEvent struct {
	TimeStamp int              `json:"timestamp"`
	TaskID    string           `json:"task_id"`
	TaskType  consts.TaskType  `json:"task_type"`
	FileName  string           `json:"file_name" swaggerignore:"true"`
	FnName    string           `json:"function_name" swaggerignore:"true"`
	Line      int              `json:"line" swaggerignore:"true"`
	EventName consts.EventType `json:"event_name"`
	Payload   any              `json:"payload,omitempty" swaggertype:"object"`
}

func (s *TraceStreamEvent) ToRedisStream() map[string]any {
	payload, err := json.Marshal(s.Payload)
	if err != nil {
		return nil
	}

	return map[string]any{
		consts.RdbEventTaskID:   s.TaskID,
		consts.RdbEventTaskType: consts.GetTaskTypeName(s.TaskType),
		consts.RdbEventFileName: s.FileName,
		consts.RdbEventFn:       s.FnName,
		consts.RdbEventLine:     s.Line,
		consts.RdbEventName:     string(s.EventName),
		consts.RdbEventPayload:  payload,
	}
}

func (s *TraceStreamEvent) ToSSE() (string, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}

type DatapackInfo struct {
	Datapack *InjectionItem `json:"datapack"`
	JobName  string         `json:"job_name"`
}

type DatapackResult struct {
	Datapack string `json:"datapack"`
	JobName  string `json:"job_name"`
}

type ExecutionInfo struct {
	Algorithm   *ContainerVersionItem `json:"algorithm"`
	Datapack    *InjectionItem        `json:"datapack"`
	ExecutionID int                   `json:"execution_id"`
	JobName     string                `json:"job_name"`
}

type ExecutionResult struct {
	Algorithm string `json:"algorithm"`
	JobName   string `json:"job_name"`
}

type InfoPayloadTemplate struct {
	State string `json:"task_state"`
	Msg   string `json:"msg"`
}

type JobMessage struct {
	JobName   string `json:"job_name"`
	Namespace string `json:"namespace"`
	LogFile   string `json:"log_file,omitempty"`
}

type TraceQuery struct {
	TraceID       string          `json:"trace_id"`
	FirstTaskType consts.TaskType `json:"first_task_type"`
	StartTime     time.Time       `json:"start_time"`
	EndTime       time.Time       `json:"end_time"`
}

type GetTraceStreamReq struct {
	LastID string `form:"last_id" binding:"omitempty"`
}

func (req *GetTraceStreamReq) Validate() error {
	if req.LastID == "" {
		req.LastID = "0"
	}

	if req.LastID == "0" {
		return nil
	}

	if strings.Count(req.LastID, "-") != 1 {
		return fmt.Errorf("invalid last_id format: must be '0' or a valid stream ID (e.g., 1678886400000-0)")
	}

	return nil
}

// GetGroupStatsReq represents the request to get group stats
type GetGroupStatsReq struct {
	GroupID string `form:"group_id" binding:"required"` // Group ID to query
}

func (req *GetGroupStatsReq) Validate() error {
	if !utils.IsValidUUID(req.GroupID) {
		return fmt.Errorf("invalid group_id: must be a valid UUID")
	}
	return nil
}

// TraceStatsItem represents the stat of a trace
type TraceStatsItem struct {
	TraceID           string             `json:"trace_id"`
	Type              string             `json:"type"`
	State             string             `json:"state"`
	StartTime         time.Time          `json:"start_time"`
	EndTime           *time.Time         `json:"end_time,omitempty"`
	CurrentEvent      string             `json:"current_event"`
	CurrentTask       string             `json:"current_task"`
	TaskTypeDurations map[string]float64 `json:"task_type_durations,omitempty" swaggertype:"object"` // Average durations per task type in seconds
}

func NewTraceStats(trace *database.Trace) *TraceStatsItem {
	detail := &TraceStatsItem{
		TraceID:      trace.ID,
		Type:         consts.GetTraceTypeName(trace.Type),
		State:        consts.GetTraceStateName(trace.State),
		StartTime:    trace.StartTime,
		EndTime:      trace.EndTime,
		CurrentEvent: trace.LastEvent.String(),
	}

	if len(trace.Tasks) > 0 {
		detail.CurrentTask = trace.Tasks[0].ID

		taskTypeMap := make(map[string][]database.Task)
		for _, task := range trace.Tasks {
			if task.State == consts.TaskCompleted || task.State == consts.TaskError {
				taskTypeName := consts.GetTaskTypeName(task.Type)
				if _, exists := taskTypeMap[taskTypeName]; !exists {
					taskTypeMap[taskTypeName] = []database.Task{}
				}
				taskTypeMap[taskTypeName] = append(taskTypeMap[taskTypeName], task)
			}
		}

		detail.TaskTypeDurations = make(map[string]float64)
		for taskTypeName, tasks := range taskTypeMap {
			totalDuration := 0.0
			for _, task := range tasks {
				duration := task.UpdatedAt.Sub(task.CreatedAt).Seconds()
				totalDuration += duration
			}
			detail.TaskTypeDurations[taskTypeName] = totalDuration / float64(len(tasks))
		}
	}

	return detail
}

// GroupStats represents the response for group stats
type GroupStats struct {
	TotalTraces   int                         `json:"total_traces"`
	AvgDuration   float64                     `json:"avg_duration"`
	MinDuration   float64                     `json:"min_duration"`
	MaxDuration   float64                     `json:"max_duration"`
	TraceStateMap map[string][]TraceStatsItem `json:"trace_state_map"`
}

func NewDefaultGroupStats() *GroupStats {
	return &GroupStats{
		TotalTraces: 0,
		AvgDuration: 0.0,
		MinDuration: 0.0,
		MaxDuration: 0.0,
	}
}

// ===================== Trace CRUD DTOs =====================

// TraceResp represents the response for a trace in list views
type TraceResp struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	LastEvent   string     `json:"last_event"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	GroupID     string     `json:"group_id"`
	ProjectID   int        `json:"project_id,omitempty"`
	ProjectName string     `json:"project_name,omitempty"`
	LeafNum     int        `json:"leaf_num"`
	State       string     `json:"state"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func NewTraceResp(trace *database.Trace) *TraceResp {
	resp := &TraceResp{
		ID:        trace.ID,
		Type:      consts.GetTraceTypeName(trace.Type),
		LastEvent: trace.LastEvent.String(),
		StartTime: trace.StartTime,
		EndTime:   trace.EndTime,
		GroupID:   trace.GroupID,
		ProjectID: trace.ProjectID,
		LeafNum:   trace.LeafNum,
		State:     consts.GetTraceStateName(trace.State),
		Status:    consts.GetStatusTypeName(trace.Status),
		CreatedAt: trace.CreatedAt,
		UpdatedAt: trace.UpdatedAt,
	}
	if trace.Project != nil {
		resp.ProjectName = trace.Project.Name
	}
	return resp
}

// TraceDetailResp represents the detailed response for a single trace
type TraceDetailResp struct {
	TraceResp

	Tasks []TaskResp `json:"tasks"`
}

func NewTraceDetailResp(trace *database.Trace) *TraceDetailResp {
	resp := &TraceDetailResp{
		TraceResp: *NewTraceResp(trace),
		Tasks:     make([]TaskResp, 0, len(trace.Tasks)),
	}
	for i := range trace.Tasks {
		resp.Tasks = append(resp.Tasks, *NewTaskResp(&trace.Tasks[i]))
	}
	return resp
}

// ListTraceFilters represents the filters for listing traces
type ListTraceFilters struct {
	TraceType *consts.TraceType
	GroupID   string
	ProjectID int
	State     *consts.TraceState
	Status    *consts.StatusType
}

// ListTraceReq represents the request to list traces
type ListTraceReq struct {
	PaginationReq
	TraceType *consts.TraceType  `form:"trace_type" binding:"omitempty"`
	GroupID   string             `form:"group_id" binding:"omitempty"`
	ProjectID int                `form:"project_id" binding:"omitempty"`
	State     *consts.TraceState `form:"state" binding:"omitempty"`
	Status    *consts.StatusType `form:"status" binding:"omitempty"`
}

func (req *ListTraceReq) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}
	if req.TraceType != nil {
		if _, exists := consts.ValidTraceTypes[*req.TraceType]; !exists {
			return fmt.Errorf("invalid trace type: %d", *req.TraceType)
		}
	}
	if err := validateUUID(req.GroupID); err != nil {
		return err
	}
	if req.ProjectID < 0 {
		return fmt.Errorf("invalid project ID: %d", req.ProjectID)
	}
	if req.State != nil {
		if _, exists := consts.ValidTraceStates[*req.State]; !exists {
			return fmt.Errorf("invalid trace state: %d", *req.State)
		}
	}
	return validateStatusField(req.Status, true)
}

func (req *ListTraceReq) ToFilterOptions() *ListTraceFilters {
	return &ListTraceFilters{
		TraceType: req.TraceType,
		GroupID:   req.GroupID,
		ProjectID: req.ProjectID,
		State:     req.State,
		Status:    req.Status,
	}
}
