package grpcorchestratorinterface

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/dto"
	redisinfra "aegis/infra/redis"
	executionmodule "aegis/module/execution"
	groupmodule "aegis/module/group"
	injectionmodule "aegis/module/injection"
	metricmodule "aegis/module/metric"
	notificationmodule "aegis/module/notification"
	taskmodule "aegis/module/task"
	tracemodule "aegis/module/trace"
	orchestratorv1 "aegis/proto/orchestrator/v1"
	"aegis/service/consumer"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

const orchestratorServiceName = "orchestrator-service"

type executionSubmitter interface {
	SubmitAlgorithmExecution(context.Context, *executionmodule.SubmitExecutionReq, string, int) (*executionmodule.SubmitExecutionResp, error)
	CreateExecutionRecord(context.Context, *executionmodule.RuntimeCreateExecutionReq) (int, error)
	UpdateExecutionState(context.Context, *executionmodule.RuntimeUpdateExecutionStateReq) error
	GetExecution(context.Context, int) (*executionmodule.ExecutionDetailResp, error)
	ListEvaluationExecutionsByDatapack(context.Context, *executionmodule.EvaluationExecutionsByDatapackReq) ([]executionmodule.EvaluationExecutionItem, error)
	ListEvaluationExecutionsByDataset(context.Context, *executionmodule.EvaluationExecutionsByDatasetReq) ([]executionmodule.EvaluationExecutionItem, error)
}

type injectionSubmitter interface {
	SubmitFaultInjection(context.Context, *injectionmodule.SubmitInjectionReq, string, int, *int) (*injectionmodule.SubmitInjectionResp, error)
	SubmitDatapackBuilding(context.Context, *injectionmodule.SubmitDatapackBuildingReq, string, int, *int) (*injectionmodule.SubmitDatapackBuildingResp, error)
	CreateInjectionRecord(context.Context, *injectionmodule.RuntimeCreateInjectionReq) (*dto.InjectionItem, error)
	UpdateInjectionState(context.Context, *injectionmodule.RuntimeUpdateInjectionStateReq) error
	UpdateInjectionTimestamps(context.Context, *injectionmodule.RuntimeUpdateInjectionTimestampReq) (*dto.InjectionItem, error)
}

type metricsReader interface {
	GetInjectionMetrics(context.Context, *metricmodule.GetMetricsReq) (*metricmodule.InjectionMetrics, error)
	GetExecutionMetrics(context.Context, *metricmodule.GetMetricsReq) (*metricmodule.ExecutionMetrics, error)
}

type taskReader interface {
	GetDetail(context.Context, string) (*taskmodule.TaskDetailResp, error)
	PollLogs(context.Context, string, time.Time) (*taskmodule.TaskLogPollResp, error)
	List(context.Context, *taskmodule.ListTaskReq) (*dto.ListResp[taskmodule.TaskResp], error)
}

type traceReader interface {
	GetTrace(context.Context, string) (*tracemodule.TraceDetailResp, error)
	ListTraces(context.Context, *tracemodule.ListTraceReq) (*dto.ListResp[tracemodule.TraceResp], error)
	GetTraceStreamAlgorithms(context.Context, string) ([]dto.ContainerVersionItem, error)
	ReadTraceStreamMessages(context.Context, string, string, int64, time.Duration) ([]redis.XStream, error)
}

type groupReader interface {
	GetGroupStats(context.Context, *groupmodule.GetGroupStatsReq) (*groupmodule.GroupStats, error)
	GetGroupTraceCount(string) (int64, error)
	ReadGroupStreamMessages(context.Context, string, string, int64, time.Duration) ([]redis.XStream, error)
}

type notificationReader interface {
	ReadStreamMessages(context.Context, string, string, int64, time.Duration) ([]redis.XStream, error)
}

type taskController interface {
	CancelTask(context.Context, string) error
	RetryTask(context.Context, string) (string, error)
	ListDeadLetterTasks(context.Context, int64) ([]QueuedTaskResp, error)
}

type taskQueueController struct {
	redis *redisinfra.Gateway
}

type QueuedTaskResp struct {
	TaskID      string `json:"task_id"`
	Type        string `json:"type"`
	Queue       string `json:"queue"`
	TraceID     string `json:"trace_id"`
	GroupID     string `json:"group_id"`
	ProjectID   int    `json:"project_id"`
	UserID      int    `json:"user_id"`
	Immediate   bool   `json:"immediate"`
	ExecuteTime int64  `json:"execute_time"`
	RestartNum  int    `json:"restart_num"`
	State       string `json:"state"`
}

func newTaskQueueController(redis *redisinfra.Gateway) taskController {
	return &taskQueueController{redis: redis}
}

func (c *taskQueueController) CancelTask(_ context.Context, taskID string) error {
	return consumer.CancelTask(c.redis, taskID)
}

func (c *taskQueueController) RetryTask(ctx context.Context, taskID string) (string, error) {
	queue, _, task, err := c.findTask(ctx, taskID)
	if err != nil {
		return "", err
	}

	switch queue {
	case redisinfra.ReadyQueueKey:
		return queue, nil
	case redisinfra.DelayedQueueKey, redisinfra.DeadLetterKey:
		if ok := c.redis.RemoveFromZSet(ctx, queue, taskID); !ok {
			return "", fmt.Errorf("%w: task %s not found in %s", consts.ErrNotFound, taskID, queue)
		}
	default:
		return "", fmt.Errorf("%w: unsupported queue %s", consts.ErrBadRequest, queue)
	}

	if err := c.redis.DeleteTaskIndex(ctx, taskID); err != nil {
		return "", fmt.Errorf("delete task index: %w", err)
	}

	task.State = consts.TaskPending
	data, err := json.Marshal(task)
	if err != nil {
		return "", err
	}
	if task.ExecuteTime > time.Now().Unix() && !task.Immediate {
		if err := c.redis.SubmitDelayedTask(ctx, data, task.TaskID, task.ExecuteTime); err != nil {
			return "", err
		}
		return redisinfra.DelayedQueueKey, nil
	}

	task.Immediate = true
	task.ExecuteTime = time.Now().Unix()
	data, err = json.Marshal(task)
	if err != nil {
		return "", err
	}
	if err := c.redis.SubmitImmediateTask(ctx, data, task.TaskID); err != nil {
		return "", err
	}
	return redisinfra.ReadyQueueKey, nil
}

func (c *taskQueueController) ListDeadLetterTasks(ctx context.Context, limit int64) ([]QueuedTaskResp, error) {
	if limit <= 0 {
		limit = 100
	}
	items, err := c.redis.ListDeadLetterTasks(ctx, limit)
	if err != nil {
		return nil, err
	}
	return decodeQueuedTasks(items, redisinfra.DeadLetterKey)
}

func (c *taskQueueController) findTask(ctx context.Context, taskID string) (string, string, *dto.UnifiedTask, error) {
	if taskID == "" {
		return "", "", nil, fmt.Errorf("%w: task_id is required", consts.ErrBadRequest)
	}

	if queue, err := c.redis.GetTaskQueue(ctx, taskID); err == nil && queue != "" {
		if taskData, task, ok := c.findTaskInQueue(ctx, queue, taskID); ok {
			return queue, taskData, task, nil
		}
	}

	for _, queue := range []string{redisinfra.ReadyQueueKey, redisinfra.DelayedQueueKey, redisinfra.DeadLetterKey} {
		if taskData, task, ok := c.findTaskInQueue(ctx, queue, taskID); ok {
			return queue, taskData, task, nil
		}
	}

	return "", "", nil, fmt.Errorf("%w: task %s not found", consts.ErrNotFound, taskID)
}

func (c *taskQueueController) findTaskInQueue(ctx context.Context, queue, taskID string) (string, *dto.UnifiedTask, bool) {
	var items []string
	var err error

	switch queue {
	case redisinfra.ReadyQueueKey:
		items, err = c.redis.ListReadyTasks(ctx)
	case redisinfra.DelayedQueueKey:
		items, err = c.redis.ListDelayedTasks(ctx, 1000)
	case redisinfra.DeadLetterKey:
		items, err = c.redis.ListDeadLetterTasks(ctx, 1000)
	default:
		return "", nil, false
	}
	if err != nil {
		return "", nil, false
	}

	for _, item := range items {
		var task dto.UnifiedTask
		if json.Unmarshal([]byte(item), &task) == nil && task.TaskID == taskID {
			return item, &task, true
		}
	}
	return "", nil, false
}

func decodeQueuedTasks(items []string, queue string) ([]QueuedTaskResp, error) {
	result := make([]QueuedTaskResp, 0, len(items))
	for _, item := range items {
		var task dto.UnifiedTask
		if err := json.Unmarshal([]byte(item), &task); err != nil {
			return nil, err
		}
		result = append(result, QueuedTaskResp{
			TaskID:      task.TaskID,
			Type:        consts.GetTaskTypeName(task.Type),
			Queue:       queue,
			TraceID:     task.TraceID,
			GroupID:     task.GroupID,
			ProjectID:   task.ProjectID,
			UserID:      task.UserID,
			Immediate:   task.Immediate,
			ExecuteTime: task.ExecuteTime,
			RestartNum:  task.ReStartNum,
			State:       consts.GetTaskStateName(task.State),
		})
	}
	return result, nil
}

type orchestratorServer struct {
	orchestratorv1.UnimplementedOrchestratorServiceServer
	execution executionSubmitter
	injection injectionSubmitter
	metrics   metricsReader
	projects  projectStatisticsReader
	tasks     taskController
	taskRead  taskReader
	traceRead traceReader
	groupRead groupReader
	notify    notificationReader
}

func newOrchestratorServer(
	execution *executionmodule.Service,
	injection *injectionmodule.Service,
	metrics *metricmodule.Service,
	projects projectStatisticsReader,
	tasks taskController,
	taskRead *taskmodule.Service,
	traceRead *tracemodule.Service,
	groupRead *groupmodule.Service,
	notify *notificationmodule.Service,
) *orchestratorServer {
	return &orchestratorServer{
		execution: execution,
		injection: injection,
		metrics:   metrics,
		projects:  projects,
		tasks:     tasks,
		taskRead:  taskRead,
		traceRead: traceRead,
		groupRead: groupRead,
		notify:    notify,
	}
}

func (s *orchestratorServer) Ping(context.Context, *orchestratorv1.PingRequest) (*orchestratorv1.PingResponse, error) {
	return &orchestratorv1.PingResponse{
		Service:       orchestratorServiceName,
		AppId:         consts.AppID,
		Status:        "ok",
		TimestampUnix: time.Now().Unix(),
	}, nil
}

func (s *orchestratorServer) SubmitExecution(ctx context.Context, req *orchestratorv1.SubmitExecutionRequest) (*orchestratorv1.SubmitExecutionResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	body, err := decodeBody[executionmodule.SubmitExecutionReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.execution.SubmitAlgorithmExecution(ctx, body, resolveGroupID(req.GetGroupId()), int(req.GetUserId()))
	if err != nil {
		return nil, mapOrchestratorError(err)
	}

	items := make([]*orchestratorv1.SubmittedExecutionItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		pbItem := &orchestratorv1.SubmittedExecutionItem{
			Index:              int64(item.Index),
			TraceId:            item.TraceID,
			TaskId:             item.TaskID,
			AlgorithmId:        int64(item.AlgorithmID),
			AlgorithmVersionId: int64(item.AlgorithmVersionID),
		}
		if item.DatapackID != nil {
			pbItem.HasDatapackId = true
			pbItem.DatapackId = int64(*item.DatapackID)
		}
		if item.DatasetID != nil {
			pbItem.HasDatasetId = true
			pbItem.DatasetId = int64(*item.DatasetID)
		}
		items = append(items, pbItem)
	}

	return &orchestratorv1.SubmitExecutionResponse{
		GroupId: resp.GroupID,
		Items:   items,
	}, nil
}

func (s *orchestratorServer) SubmitFaultInjection(ctx context.Context, req *orchestratorv1.SubmitFaultInjectionRequest) (*orchestratorv1.SubmitFaultInjectionResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	body, err := decodeBody[injectionmodule.SubmitInjectionReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.injection.SubmitFaultInjection(ctx, body, resolveGroupID(req.GetGroupId()), int(req.GetUserId()), optionalID(req.GetProjectId()))
	if err != nil {
		return nil, mapOrchestratorError(err)
	}

	items := make([]*orchestratorv1.SubmittedInjectionItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, &orchestratorv1.SubmittedInjectionItem{
			Index:   int64(item.Index),
			TraceId: item.TraceID,
			TaskId:  item.TaskID,
		})
	}

	result := &orchestratorv1.SubmitFaultInjectionResponse{
		GroupId:       resp.GroupID,
		Items:         items,
		OriginalCount: int64(resp.OriginalCount),
	}
	if resp.Warnings != nil {
		result.Warnings = &orchestratorv1.InjectionWarnings{
			DuplicateServicesInBatch: resp.Warnings.DuplicateServicesInBatch,
			DuplicateBatchesInRequest: intsToInt64s(
				resp.Warnings.DuplicateBatchesInRequest,
			),
			BatchesExistInDatabase: intsToInt64s(resp.Warnings.BatchesExistInDatabase),
		}
	}
	return result, nil
}

func (s *orchestratorServer) SubmitDatapackBuilding(ctx context.Context, req *orchestratorv1.SubmitDatapackBuildingRequest) (*orchestratorv1.SubmitDatapackBuildingResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	body, err := decodeBody[injectionmodule.SubmitDatapackBuildingReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.injection.SubmitDatapackBuilding(ctx, body, resolveGroupID(req.GetGroupId()), int(req.GetUserId()), optionalID(req.GetProjectId()))
	if err != nil {
		return nil, mapOrchestratorError(err)
	}

	items := make([]*orchestratorv1.SubmittedBuildingItem, 0, len(resp.Items))
	for _, item := range resp.Items {
		items = append(items, &orchestratorv1.SubmittedBuildingItem{
			Index:   int64(item.Index),
			TraceId: item.TraceID,
			TaskId:  item.TaskID,
		})
	}

	return &orchestratorv1.SubmitDatapackBuildingResponse{
		GroupId: resp.GroupID,
		Items:   items,
	}, nil
}

func (s *orchestratorServer) CreateExecution(ctx context.Context, req *orchestratorv1.MutationRequest) (*orchestratorv1.StructResponse, error) {
	body, err := decodeBody[executionmodule.RuntimeCreateExecutionReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	executionID, err := s.execution.CreateExecutionRecord(ctx, body)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(map[string]any{"execution_id": executionID})
}

func (s *orchestratorServer) CreateInjection(ctx context.Context, req *orchestratorv1.MutationRequest) (*orchestratorv1.StructResponse, error) {
	body, err := decodeBody[injectionmodule.RuntimeCreateInjectionReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.injection.CreateInjectionRecord(ctx, body)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(resp)
}

func (s *orchestratorServer) UpdateExecutionState(ctx context.Context, req *orchestratorv1.MutationRequest) (*orchestratorv1.StructResponse, error) {
	body, err := decodeBody[executionmodule.RuntimeUpdateExecutionStateReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.execution.UpdateExecutionState(ctx, body); err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(map[string]any{"updated": true})
}

func (s *orchestratorServer) UpdateInjectionState(ctx context.Context, req *orchestratorv1.MutationRequest) (*orchestratorv1.StructResponse, error) {
	body, err := decodeBody[injectionmodule.RuntimeUpdateInjectionStateReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.injection.UpdateInjectionState(ctx, body); err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(map[string]any{"updated": true})
}

func (s *orchestratorServer) UpdateInjectionTimestamps(ctx context.Context, req *orchestratorv1.MutationRequest) (*orchestratorv1.StructResponse, error) {
	body, err := decodeBody[injectionmodule.RuntimeUpdateInjectionTimestampReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.injection.UpdateInjectionTimestamps(ctx, body)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(resp)
}

func (s *orchestratorServer) CancelTask(ctx context.Context, req *orchestratorv1.CancelTaskRequest) (*orchestratorv1.CancelTaskResponse, error) {
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}

	if err := s.tasks.CancelTask(ctx, req.GetTaskId()); err != nil {
		return nil, mapOrchestratorError(err)
	}
	return &orchestratorv1.CancelTaskResponse{Cancelled: true}, nil
}

func (s *orchestratorServer) GetExecution(ctx context.Context, req *orchestratorv1.GetExecutionRequest) (*orchestratorv1.StructResponse, error) {
	if req.GetExecutionId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "execution_id is required")
	}
	resp, err := s.execution.GetExecution(ctx, int(req.GetExecutionId()))
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(resp)
}

func (s *orchestratorServer) GetInjectionMetrics(ctx context.Context, req *orchestratorv1.MutationRequest) (*orchestratorv1.StructResponse, error) {
	body, err := decodeBody[metricmodule.GetMetricsReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if s.metrics == nil {
		return nil, status.Error(codes.FailedPrecondition, "metrics service is not configured")
	}
	resp, err := s.metrics.GetInjectionMetrics(ctx, body)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(resp)
}

func (s *orchestratorServer) GetExecutionMetrics(ctx context.Context, req *orchestratorv1.MutationRequest) (*orchestratorv1.StructResponse, error) {
	body, err := decodeBody[metricmodule.GetMetricsReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if s.metrics == nil {
		return nil, status.Error(codes.FailedPrecondition, "metrics service is not configured")
	}
	resp, err := s.metrics.GetExecutionMetrics(ctx, body)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(resp)
}

func (s *orchestratorServer) ListProjectStatistics(ctx context.Context, req *orchestratorv1.ListProjectStatisticsRequest) (*orchestratorv1.StructResponse, error) {
	projectIDs := int64sToInts(req.GetProjectIds())
	for _, projectID := range projectIDs {
		if projectID <= 0 {
			return nil, status.Error(codes.InvalidArgument, "project_ids must be greater than 0")
		}
	}
	resp, err := s.projects.ListProjectStatistics(projectIDs)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(resp)
}

func (s *orchestratorServer) ListEvaluationExecutionsByDatapack(ctx context.Context, req *orchestratorv1.MutationRequest) (*orchestratorv1.StructResponse, error) {
	body, err := decodeBody[executionmodule.EvaluationExecutionsByDatapackReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.execution.ListEvaluationExecutionsByDatapack(ctx, body)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(map[string]any{"items": resp})
}

func (s *orchestratorServer) ListEvaluationExecutionsByDataset(ctx context.Context, req *orchestratorv1.MutationRequest) (*orchestratorv1.StructResponse, error) {
	body, err := decodeBody[executionmodule.EvaluationExecutionsByDatasetReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.execution.ListEvaluationExecutionsByDataset(ctx, body)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(map[string]any{"items": resp})
}

func (s *orchestratorServer) GetTask(ctx context.Context, req *orchestratorv1.GetTaskRequest) (*orchestratorv1.StructResponse, error) {
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}
	resp, err := s.taskRead.GetDetail(ctx, req.GetTaskId())
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(resp)
}

func (s *orchestratorServer) PollTaskLogs(ctx context.Context, req *orchestratorv1.PollTaskLogsRequest) (*orchestratorv1.StructResponse, error) {
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}
	after := time.Time{}
	if req.GetAfterUnixNano() > 0 {
		after = time.Unix(0, req.GetAfterUnixNano())
	}
	resp, err := s.taskRead.PollLogs(ctx, req.GetTaskId(), after)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(resp)
}

func (s *orchestratorServer) ListTasks(ctx context.Context, req *orchestratorv1.ListTasksRequest) (*orchestratorv1.StructResponse, error) {
	query, err := decodeQuery[taskmodule.ListTaskReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.taskRead.List(ctx, query)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(resp)
}

func (s *orchestratorServer) GetTrace(ctx context.Context, req *orchestratorv1.GetTraceRequest) (*orchestratorv1.StructResponse, error) {
	if req.GetTraceId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trace_id is required")
	}
	resp, err := s.traceRead.GetTrace(ctx, req.GetTraceId())
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(resp)
}

func (s *orchestratorServer) ListTraces(ctx context.Context, req *orchestratorv1.ListTracesRequest) (*orchestratorv1.StructResponse, error) {
	query, err := decodeQuery[tracemodule.ListTraceReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	resp, err := s.traceRead.ListTraces(ctx, query)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(resp)
}

func (s *orchestratorServer) GetGroupStats(ctx context.Context, req *orchestratorv1.GetGroupStatsRequest) (*orchestratorv1.StructResponse, error) {
	if req.GetGroupId() == "" {
		return nil, status.Error(codes.InvalidArgument, "group_id is required")
	}
	resp, err := s.groupRead.GetGroupStats(ctx, &groupmodule.GetGroupStatsReq{GroupID: req.GetGroupId()})
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(resp)
}

func (s *orchestratorServer) GetTraceStreamState(ctx context.Context, req *orchestratorv1.GetTraceStreamStateRequest) (*orchestratorv1.StructResponse, error) {
	if req.GetTraceId() == "" {
		return nil, status.Error(codes.InvalidArgument, "trace_id is required")
	}
	algorithms, err := s.traceRead.GetTraceStreamAlgorithms(ctx, req.GetTraceId())
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(traceStreamStateResp{Algorithms: algorithms})
}

func (s *orchestratorServer) ReadTraceStreamMessages(ctx context.Context, req *orchestratorv1.ReadStreamMessagesRequest) (*orchestratorv1.StructResponse, error) {
	if req.GetStreamKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "stream_key is required")
	}
	resp, err := s.traceRead.ReadTraceStreamMessages(ctx, req.GetStreamKey(), req.GetLastId(), req.GetCount(), time.Duration(req.GetBlockMillis())*time.Millisecond)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStreamMessages(resp)
}

func (s *orchestratorServer) GetGroupStreamState(ctx context.Context, req *orchestratorv1.GetGroupStreamStateRequest) (*orchestratorv1.StructResponse, error) {
	if req.GetGroupId() == "" {
		return nil, status.Error(codes.InvalidArgument, "group_id is required")
	}
	totalTraces, err := s.groupRead.GetGroupTraceCount(req.GetGroupId())
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(groupStreamStateResp{TotalTraces: int(totalTraces)})
}

func (s *orchestratorServer) ReadGroupStreamMessages(ctx context.Context, req *orchestratorv1.ReadStreamMessagesRequest) (*orchestratorv1.StructResponse, error) {
	if req.GetStreamKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "stream_key is required")
	}
	resp, err := s.groupRead.ReadGroupStreamMessages(ctx, req.GetStreamKey(), req.GetLastId(), req.GetCount(), time.Duration(req.GetBlockMillis())*time.Millisecond)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStreamMessages(resp)
}

func (s *orchestratorServer) ReadNotificationStreamMessages(ctx context.Context, req *orchestratorv1.ReadStreamMessagesRequest) (*orchestratorv1.StructResponse, error) {
	if req.GetStreamKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "stream_key is required")
	}
	if s.notify == nil {
		return nil, status.Error(codes.FailedPrecondition, "notification service is not configured")
	}
	resp, err := s.notify.ReadStreamMessages(ctx, req.GetStreamKey(), req.GetLastId(), req.GetCount(), time.Duration(req.GetBlockMillis())*time.Millisecond)
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStreamMessages(resp)
}

func (s *orchestratorServer) ListDeadLetterTasks(ctx context.Context, req *orchestratorv1.ListDeadLetterTasksRequest) (*orchestratorv1.StructResponse, error) {
	resp, err := s.tasks.ListDeadLetterTasks(ctx, req.GetLimit())
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return encodeStruct(map[string]any{"items": resp})
}

func (s *orchestratorServer) RetryTask(ctx context.Context, req *orchestratorv1.RetryTaskRequest) (*orchestratorv1.RetryTaskResponse, error) {
	if req.GetTaskId() == "" {
		return nil, status.Error(codes.InvalidArgument, "task_id is required")
	}
	queue, err := s.tasks.RetryTask(ctx, req.GetTaskId())
	if err != nil {
		return nil, mapOrchestratorError(err)
	}
	return &orchestratorv1.RetryTaskResponse{Accepted: true, Queue: queue}, nil
}

func resolveGroupID(groupID string) string {
	if groupID != "" {
		return groupID
	}
	return uuid.NewString()
}

func optionalID(value int64) *int {
	if value <= 0 {
		return nil
	}
	id := int(value)
	return &id
}

func intsToInt64s(items []int) []int64 {
	if len(items) == 0 {
		return nil
	}
	result := make([]int64, 0, len(items))
	for _, item := range items {
		result = append(result, int64(item))
	}
	return result
}

func int64sToInts(items []int64) []int {
	if len(items) == 0 {
		return nil
	}
	result := make([]int, 0, len(items))
	for _, item := range items {
		result = append(result, int(item))
	}
	return result
}

func decodeBody[T any](body *structpb.Struct) (*T, error) {
	if body == nil {
		return nil, errors.New("body is required")
	}

	data, err := json.Marshal(body.AsMap())
	if err != nil {
		return nil, err
	}

	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func decodeQuery[T any](query *structpb.Struct) (*T, error) {
	var result T
	if query == nil {
		return &result, nil
	}

	data, err := json.Marshal(query.AsMap())
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func encodeStruct(value any) (*orchestratorv1.StructResponse, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	payload := map[string]any{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	item, err := structpb.NewStruct(payload)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &orchestratorv1.StructResponse{Data: item}, nil
}

type traceStreamStateResp struct {
	Algorithms []dto.ContainerVersionItem `json:"algorithms"`
}

type groupStreamStateResp struct {
	TotalTraces int `json:"total_traces"`
}

type streamBatchResp struct {
	Messages []streamMessageResp `json:"messages"`
}

type streamMessageResp struct {
	ID     string         `json:"id"`
	Values map[string]any `json:"values"`
}

func encodeStreamMessages(streams []redis.XStream) (*orchestratorv1.StructResponse, error) {
	messages := []streamMessageResp{}
	if len(streams) > 0 {
		messages = make([]streamMessageResp, 0, len(streams[0].Messages))
		for _, item := range streams[0].Messages {
			messages = append(messages, streamMessageResp{
				ID:     item.ID,
				Values: item.Values,
			})
		}
	}
	return encodeStruct(streamBatchResp{Messages: messages})
}

func mapOrchestratorError(err error) error {
	switch {
	case errors.Is(err, consts.ErrAuthenticationFailed):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, consts.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, consts.ErrBadRequest):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, consts.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, consts.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case err != nil:
		return status.Error(codes.Internal, err.Error())
	default:
		return nil
	}
}
