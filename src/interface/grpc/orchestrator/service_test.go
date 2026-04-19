package grpcorchestrator

import (
	"context"
	"errors"
	"testing"
	"time"

	"aegis/consts"
	"aegis/dto"
	execution "aegis/module/execution"
	group "aegis/module/group"
	injection "aegis/module/injection"
	metric "aegis/module/metric"
	task "aegis/module/task"
	trace "aegis/module/trace"
	orchestratorv1 "aegis/proto/orchestrator/v1"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type executionSubmitterStub struct {
	resp            *execution.SubmitExecutionResp
	id              int
	item            *execution.ExecutionDetailResp
	evaluationItems []execution.EvaluationExecutionItem
	err             error
}

func (s executionSubmitterStub) SubmitAlgorithmExecution(_ context.Context, req *execution.SubmitExecutionReq, groupID string, userID int) (*execution.SubmitExecutionResp, error) {
	if req.ProjectName == "" || groupID == "" || userID <= 0 {
		return nil, errors.New("unexpected request")
	}
	return s.resp, s.err
}

func (s executionSubmitterStub) CreateExecutionRecord(_ context.Context, req *execution.RuntimeCreateExecutionReq) (int, error) {
	if req.TaskID == "" || req.AlgorithmVersionID <= 0 || req.DatapackID <= 0 {
		return 0, errors.New("unexpected runtime execution request")
	}
	return s.id, s.err
}

func (s executionSubmitterStub) UpdateExecutionState(_ context.Context, req *execution.RuntimeUpdateExecutionStateReq) error {
	if req.ExecutionID <= 0 {
		return errors.New("unexpected execution state request")
	}
	return s.err
}

func (s executionSubmitterStub) GetExecution(_ context.Context, executionID int) (*execution.ExecutionDetailResp, error) {
	if executionID <= 0 {
		return nil, errors.New("missing execution id")
	}
	return s.item, s.err
}

func (s executionSubmitterStub) ListEvaluationExecutionsByDatapack(_ context.Context, req *execution.EvaluationExecutionsByDatapackReq) ([]execution.EvaluationExecutionItem, error) {
	if req.AlgorithmVersionID <= 0 || req.DatapackName == "" {
		return nil, errors.New("unexpected datapack evaluation query")
	}
	return s.evaluationItems, s.err
}

func (s executionSubmitterStub) ListEvaluationExecutionsByDataset(_ context.Context, req *execution.EvaluationExecutionsByDatasetReq) ([]execution.EvaluationExecutionItem, error) {
	if req.AlgorithmVersionID <= 0 || req.DatasetVersionID <= 0 {
		return nil, errors.New("unexpected dataset evaluation query")
	}
	return s.evaluationItems, s.err
}

type injectionSubmitterStub struct {
	injectionResp *injection.SubmitInjectionResp
	buildResp     *injection.SubmitDatapackBuildingResp
	item          *dto.InjectionItem
	err           error
}

func (s injectionSubmitterStub) SubmitFaultInjection(_ context.Context, req *injection.SubmitInjectionReq, groupID string, userID int, projectID *int) (*injection.SubmitInjectionResp, error) {
	if req.Pedestal == nil || req.Benchmark == nil || groupID == "" || userID <= 0 {
		return nil, errors.New("unexpected injection request")
	}
	if projectID == nil || *projectID != 9 {
		return nil, errors.New("unexpected project id")
	}
	return s.injectionResp, s.err
}

func (s injectionSubmitterStub) SubmitDatapackBuilding(_ context.Context, req *injection.SubmitDatapackBuildingReq, groupID string, userID int, projectID *int) (*injection.SubmitDatapackBuildingResp, error) {
	if len(req.Specs) == 0 || groupID == "" || userID <= 0 {
		return nil, errors.New("unexpected datapack request")
	}
	if projectID == nil || *projectID != 5 {
		return nil, errors.New("unexpected project id")
	}
	return s.buildResp, s.err
}

func (s injectionSubmitterStub) CreateInjectionRecord(_ context.Context, req *injection.RuntimeCreateInjectionReq) (*dto.InjectionItem, error) {
	if req.Name == "" || req.TaskID == "" {
		return nil, errors.New("unexpected runtime injection request")
	}
	return s.item, s.err
}

type metricsReaderStub struct {
	injection *metric.InjectionMetrics
	execution *metric.ExecutionMetrics
	err       error
}

func (s metricsReaderStub) GetInjectionMetrics(_ context.Context, req *metric.GetMetricsReq) (*metric.InjectionMetrics, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return s.injection, s.err
}

func (s metricsReaderStub) GetExecutionMetrics(_ context.Context, req *metric.GetMetricsReq) (*metric.ExecutionMetrics, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return s.execution, s.err
}

func (s injectionSubmitterStub) UpdateInjectionState(_ context.Context, req *injection.RuntimeUpdateInjectionStateReq) error {
	if req.Name == "" {
		return errors.New("unexpected injection state request")
	}
	return s.err
}

func (s injectionSubmitterStub) UpdateInjectionTimestamps(_ context.Context, req *injection.RuntimeUpdateInjectionTimestampReq) (*dto.InjectionItem, error) {
	if req.Name == "" {
		return nil, errors.New("unexpected injection timestamp request")
	}
	return s.item, s.err
}

type taskControllerStub struct {
	taskID   string
	queue    string
	listResp []QueuedTaskResp
	err      error
}

func (s taskControllerStub) CancelTask(_ context.Context, taskID string) error {
	if taskID == "" {
		return errors.New("missing task id")
	}
	if s.taskID != "" && s.taskID != taskID {
		return errors.New("unexpected task id")
	}
	return s.err
}

func (s taskControllerStub) RetryTask(_ context.Context, taskID string) (string, error) {
	if taskID == "" {
		return "", errors.New("missing task id")
	}
	if s.taskID != "" && s.taskID != taskID {
		return "", errors.New("unexpected task id")
	}
	return s.queue, s.err
}

func (s taskControllerStub) ListDeadLetterTasks(_ context.Context, limit int64) ([]QueuedTaskResp, error) {
	if limit == 0 {
		return s.listResp, s.err
	}
	return s.listResp, s.err
}

type taskReaderStub struct {
	detail *task.TaskDetailResp
	list   *dto.ListResp[task.TaskResp]
	err    error
}

func (s taskReaderStub) GetDetail(_ context.Context, taskID string) (*task.TaskDetailResp, error) {
	if taskID == "" {
		return nil, errors.New("missing task id")
	}
	return s.detail, s.err
}

func (s taskReaderStub) PollLogs(_ context.Context, taskID string, _ time.Time) (*task.TaskLogPollResp, error) {
	if taskID == "" {
		return nil, errors.New("missing task id")
	}
	return &task.TaskLogPollResp{
		Logs:      []dto.LogEntry{{TaskID: taskID, Line: "hello"}},
		Terminal:  false,
		State:     consts.GetTaskStateName(consts.TaskPending),
		CreatedAt: time.Unix(1710000000, 0),
	}, s.err
}

func (s taskReaderStub) List(_ context.Context, req *task.ListTaskReq) (*dto.ListResp[task.TaskResp], error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return s.list, s.err
}

type traceReaderStub struct {
	detail     *trace.TraceDetailResp
	list       *dto.ListResp[trace.TraceResp]
	algorithms []dto.ContainerVersionItem
	messages   []redis.XStream
	err        error
}

func (s traceReaderStub) GetTrace(_ context.Context, traceID string) (*trace.TraceDetailResp, error) {
	if traceID == "" {
		return nil, errors.New("missing trace id")
	}
	return s.detail, s.err
}

func (s traceReaderStub) ListTraces(_ context.Context, req *trace.ListTraceReq) (*dto.ListResp[trace.TraceResp], error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return s.list, s.err
}

func (s traceReaderStub) GetTraceStreamAlgorithms(_ context.Context, traceID string) ([]dto.ContainerVersionItem, error) {
	if traceID == "" {
		return nil, errors.New("missing trace id")
	}
	return s.algorithms, s.err
}

func (s traceReaderStub) ReadTraceStreamMessages(_ context.Context, streamKey, _ string, _ int64, _ time.Duration) ([]redis.XStream, error) {
	if streamKey == "" {
		return nil, errors.New("missing stream key")
	}
	return s.messages, s.err
}

type groupReaderStub struct {
	stats    *group.GroupStats
	count    int64
	messages []redis.XStream
	err      error
}

func (s groupReaderStub) GetGroupStats(_ context.Context, req *group.GetGroupStatsReq) (*group.GroupStats, error) {
	if req == nil || req.GroupID == "" {
		return nil, errors.New("missing group id")
	}
	return s.stats, s.err
}

func (s groupReaderStub) GetGroupTraceCount(groupID string) (int64, error) {
	if groupID == "" {
		return 0, errors.New("missing group id")
	}
	if s.count == 0 {
		return 1, s.err
	}
	return s.count, s.err
}

func (s groupReaderStub) ReadGroupStreamMessages(_ context.Context, streamKey, _ string, _ int64, _ time.Duration) ([]redis.XStream, error) {
	if streamKey == "" {
		return nil, errors.New("missing stream key")
	}
	return s.messages, s.err
}

type notificationReaderStub struct {
	messages []redis.XStream
	err      error
}

func (s notificationReaderStub) ReadStreamMessages(_ context.Context, streamKey, _ string, _ int64, _ time.Duration) ([]redis.XStream, error) {
	if streamKey == "" {
		return nil, errors.New("missing stream key")
	}
	return s.messages, s.err
}

func TestOrchestratorServerSubmitExecution(t *testing.T) {
	server := &orchestratorServer{
		execution: executionSubmitterStub{resp: &execution.SubmitExecutionResp{
			GroupID: "group-1",
			Items: []execution.SubmitExecutionItem{{
				Index:              0,
				TraceID:            "trace-1",
				TaskID:             "task-1",
				AlgorithmID:        11,
				AlgorithmVersionID: 12,
			}},
		}},
		injection: injectionSubmitterStub{},
		metrics:   metricsReaderStub{},
		tasks:     taskControllerStub{},
		taskRead:  taskReaderStub{},
		traceRead: traceReaderStub{},
		groupRead: groupReaderStub{},
		notify:    notificationReaderStub{},
	}

	body, err := structpb.NewStruct(map[string]any{
		"project_name": "demo",
		"specs": []any{
			map[string]any{
				"algorithm": map[string]any{
					"name":    "algo",
					"version": "1.0.0",
				},
				"datapack": "dp-1",
			},
		},
	})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	resp, err := server.SubmitExecution(context.Background(), &orchestratorv1.SubmitExecutionRequest{
		GroupId: "group-1",
		UserId:  7,
		Body:    body,
	})
	if err != nil {
		t.Fatalf("SubmitExecution() error = %v", err)
	}
	if resp.GroupId != "group-1" || len(resp.Items) != 1 || resp.Items[0].TaskId != "task-1" {
		t.Fatalf("SubmitExecution() unexpected response: %+v", resp)
	}
}

func TestOrchestratorServerSubmitFaultInjection(t *testing.T) {
	server := &orchestratorServer{
		execution: executionSubmitterStub{},
		injection: injectionSubmitterStub{injectionResp: &injection.SubmitInjectionResp{
			GroupID:       "group-2",
			OriginalCount: 1,
			Items: []injection.SubmitInjectionItem{{
				Index:   0,
				TraceID: "trace-2",
				TaskID:  "task-2",
			}},
			Warnings: &injection.InjectionWarnings{
				DuplicateServicesInBatch: []string{"svc-a"},
			},
		}},
		metrics:   metricsReaderStub{},
		tasks:     taskControllerStub{},
		taskRead:  taskReaderStub{},
		traceRead: traceReaderStub{},
		groupRead: groupReaderStub{},
		notify:    notificationReaderStub{},
	}

	body, err := structpb.NewStruct(map[string]any{
		"project_name": "demo",
		"pedestal": map[string]any{
			"name":    "pedestal",
			"version": "1.0.0",
		},
		"benchmark": map[string]any{
			"name":    "bench",
			"version": "1.0.0",
		},
		"interval":     10,
		"pre_duration": 5,
		"specs": []any{
			[]any{
				map[string]any{
					"type": "PodChaos",
					"name": "fault-a",
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	resp, err := server.SubmitFaultInjection(context.Background(), &orchestratorv1.SubmitFaultInjectionRequest{
		GroupId:   "group-2",
		UserId:    8,
		ProjectId: 9,
		Body:      body,
	})
	if err != nil {
		t.Fatalf("SubmitFaultInjection() error = %v", err)
	}
	if resp.GroupId != "group-2" || len(resp.Items) != 1 || resp.Warnings == nil {
		t.Fatalf("SubmitFaultInjection() unexpected response: %+v", resp)
	}
}

func TestOrchestratorServerRuntimeMutations(t *testing.T) {
	injectionItem := &dto.InjectionItem{ID: 33, Name: "dp-1"}
	server := &orchestratorServer{
		execution: executionSubmitterStub{
			id: 22,
			item: &execution.ExecutionDetailResp{
				ExecutionResp: execution.ExecutionResp{ID: 22},
			},
			evaluationItems: []execution.EvaluationExecutionItem{{
				Datapack: "dp-1",
				ExecutionRef: execution.ExecutionRef{
					ExecutionID: 22,
				},
			}},
		},
		injection: injectionSubmitterStub{item: injectionItem},
		metrics: metricsReaderStub{
			injection: &metric.InjectionMetrics{TotalCount: 3},
			execution: &metric.ExecutionMetrics{TotalCount: 4},
		},
		tasks:     taskControllerStub{},
		taskRead:  taskReaderStub{},
		traceRead: traceReaderStub{},
		groupRead: groupReaderStub{},
		notify:    notificationReaderStub{},
	}

	body, err := structpb.NewStruct(map[string]any{
		"task_id":              "task-1",
		"algorithm_version_id": 10,
		"datapack_id":          11,
	})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}
	resp, err := server.CreateExecution(context.Background(), &orchestratorv1.MutationRequest{Body: body})
	if err != nil {
		t.Fatalf("CreateExecution() error = %v", err)
	}
	if resp.GetData().AsMap()["execution_id"] != float64(22) {
		t.Fatalf("CreateExecution() unexpected response: %+v", resp.GetData().AsMap())
	}

	injectionBody, err := structpb.NewStruct(map[string]any{
		"name":               "dp-1",
		"task_id":            "task-2",
		"display_config":     "{}",
		"engine_config":      "[]",
		"groundtruth_source": "auto",
		"pre_duration":       5,
		"state":              consts.GetDatapackStateName(consts.DatapackInitial),
	})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}
	created, err := server.CreateInjection(context.Background(), &orchestratorv1.MutationRequest{Body: injectionBody})
	if err != nil {
		t.Fatalf("CreateInjection() error = %v", err)
	}
	if created.GetData().AsMap()["name"] != "dp-1" {
		t.Fatalf("CreateInjection() unexpected response: %+v", created.GetData().AsMap())
	}

	updateExecBody, _ := structpb.NewStruct(map[string]any{"execution_id": 22, "state": consts.GetExecutionStateName(consts.ExecutionSuccess)})
	if _, err := server.UpdateExecutionState(context.Background(), &orchestratorv1.MutationRequest{Body: updateExecBody}); err != nil {
		t.Fatalf("UpdateExecutionState() error = %v", err)
	}

	updateInjectionBody, _ := structpb.NewStruct(map[string]any{"name": "dp-1", "state": consts.GetDatapackStateName(consts.DatapackInjectSuccess)})
	if _, err := server.UpdateInjectionState(context.Background(), &orchestratorv1.MutationRequest{Body: updateInjectionBody}); err != nil {
		t.Fatalf("UpdateInjectionState() error = %v", err)
	}

	updateTimestampBody, _ := structpb.NewStruct(map[string]any{
		"name":       "dp-1",
		"start_time": time.Now().Format(time.RFC3339Nano),
		"end_time":   time.Now().Add(time.Minute).Format(time.RFC3339Nano),
	})
	if _, err := server.UpdateInjectionTimestamps(context.Background(), &orchestratorv1.MutationRequest{Body: updateTimestampBody}); err != nil {
		t.Fatalf("UpdateInjectionTimestamps() error = %v", err)
	}

	got, err := server.GetExecution(context.Background(), &orchestratorv1.GetExecutionRequest{ExecutionId: 22})
	if err != nil {
		t.Fatalf("GetExecution() error = %v", err)
	}
	if got.GetData().AsMap()["id"] != float64(22) {
		t.Fatalf("GetExecution() unexpected response: %+v", got.GetData().AsMap())
	}

	metricQuery, _ := structpb.NewStruct(map[string]any{})
	injectionMetricsResp, err := server.GetInjectionMetrics(context.Background(), &orchestratorv1.MutationRequest{Body: metricQuery})
	if err != nil {
		t.Fatalf("GetInjectionMetrics() error = %v", err)
	}
	if injectionMetricsResp.GetData().AsMap()["total_count"] != float64(3) {
		t.Fatalf("GetInjectionMetrics() unexpected response: %+v", injectionMetricsResp.GetData().AsMap())
	}

	executionMetricsResp, err := server.GetExecutionMetrics(context.Background(), &orchestratorv1.MutationRequest{Body: metricQuery})
	if err != nil {
		t.Fatalf("GetExecutionMetrics() error = %v", err)
	}
	if executionMetricsResp.GetData().AsMap()["total_count"] != float64(4) {
		t.Fatalf("GetExecutionMetrics() unexpected response: %+v", executionMetricsResp.GetData().AsMap())
	}

	datapackQuery, _ := structpb.NewStruct(map[string]any{
		"algorithm_version_id": 11,
		"datapack_name":        "dp-1",
	})
	datapackResp, err := server.ListEvaluationExecutionsByDatapack(context.Background(), &orchestratorv1.MutationRequest{Body: datapackQuery})
	if err != nil {
		t.Fatalf("ListEvaluationExecutionsByDatapack() error = %v", err)
	}
	datapackItems, ok := datapackResp.GetData().AsMap()["items"].([]any)
	if !ok || len(datapackItems) != 1 {
		t.Fatalf("ListEvaluationExecutionsByDatapack() unexpected response: %+v", datapackResp.GetData().AsMap())
	}

	datasetQuery, _ := structpb.NewStruct(map[string]any{
		"algorithm_version_id": 11,
		"dataset_version_id":   7,
	})
	datasetResp, err := server.ListEvaluationExecutionsByDataset(context.Background(), &orchestratorv1.MutationRequest{Body: datasetQuery})
	if err != nil {
		t.Fatalf("ListEvaluationExecutionsByDataset() error = %v", err)
	}
	datasetItems, ok := datasetResp.GetData().AsMap()["items"].([]any)
	if !ok || len(datasetItems) != 1 {
		t.Fatalf("ListEvaluationExecutionsByDataset() unexpected response: %+v", datasetResp.GetData().AsMap())
	}
}

func TestOrchestratorServerCancelTaskNotFound(t *testing.T) {
	server := &orchestratorServer{
		execution: executionSubmitterStub{},
		injection: injectionSubmitterStub{},
		metrics:   metricsReaderStub{},
		tasks:     taskControllerStub{taskID: "task-404", err: consts.ErrNotFound},
		taskRead:  taskReaderStub{},
		traceRead: traceReaderStub{},
		groupRead: groupReaderStub{},
		notify:    notificationReaderStub{},
	}

	_, err := server.CancelTask(context.Background(), &orchestratorv1.CancelTaskRequest{TaskId: "task-404"})
	if err == nil {
		t.Fatal("CancelTask() error = nil, want error")
	}
	if status.Code(err) != codes.NotFound {
		t.Fatalf("CancelTask() code = %s, want %s", status.Code(err), codes.NotFound)
	}
}

func TestOrchestratorServerListDeadLetterTasks(t *testing.T) {
	server := &orchestratorServer{
		execution: executionSubmitterStub{},
		injection: injectionSubmitterStub{},
		metrics:   metricsReaderStub{},
		tasks: taskControllerStub{listResp: []QueuedTaskResp{{
			TaskID: "task-dead",
			Queue:  "task:dead",
			Type:   consts.GetTaskTypeName(consts.TaskTypeRunAlgorithm),
		}}},
		taskRead:  taskReaderStub{},
		traceRead: traceReaderStub{},
		groupRead: groupReaderStub{},
		notify:    notificationReaderStub{},
	}

	resp, err := server.ListDeadLetterTasks(context.Background(), &orchestratorv1.ListDeadLetterTasksRequest{Limit: 10})
	if err != nil {
		t.Fatalf("ListDeadLetterTasks() error = %v", err)
	}
	items, ok := resp.GetData().AsMap()["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("ListDeadLetterTasks() unexpected response: %+v", resp.GetData().AsMap())
	}
}

func TestOrchestratorServerRetryTask(t *testing.T) {
	server := &orchestratorServer{
		execution: executionSubmitterStub{},
		injection: injectionSubmitterStub{},
		metrics:   metricsReaderStub{},
		tasks:     taskControllerStub{taskID: "task-dead", queue: "task:ready"},
		taskRead:  taskReaderStub{},
		traceRead: traceReaderStub{},
		groupRead: groupReaderStub{},
		notify:    notificationReaderStub{},
	}

	resp, err := server.RetryTask(context.Background(), &orchestratorv1.RetryTaskRequest{TaskId: "task-dead"})
	if err != nil {
		t.Fatalf("RetryTask() error = %v", err)
	}
	if !resp.GetAccepted() || resp.GetQueue() != "task:ready" {
		t.Fatalf("RetryTask() unexpected response: %+v", resp)
	}
}

func TestOrchestratorServerGetTask(t *testing.T) {
	server := &orchestratorServer{
		execution: executionSubmitterStub{},
		injection: injectionSubmitterStub{},
		metrics:   metricsReaderStub{},
		tasks:     taskControllerStub{},
		taskRead: taskReaderStub{detail: &task.TaskDetailResp{
			TaskResp: task.TaskResp{
				ID:    "task-1",
				Type:  consts.GetTaskTypeName(consts.TaskTypeRunAlgorithm),
				State: consts.GetTaskStateName(consts.TaskPending),
			},
		}},
		traceRead: traceReaderStub{},
		groupRead: groupReaderStub{},
		notify:    notificationReaderStub{},
	}

	resp, err := server.GetTask(context.Background(), &orchestratorv1.GetTaskRequest{TaskId: "task-1"})
	if err != nil {
		t.Fatalf("GetTask() error = %v", err)
	}
	if resp.GetData().AsMap()["id"] != "task-1" {
		t.Fatalf("GetTask() unexpected response: %+v", resp.GetData().AsMap())
	}
}

func TestOrchestratorServerPollTaskLogs(t *testing.T) {
	server := &orchestratorServer{
		execution: executionSubmitterStub{},
		injection: injectionSubmitterStub{},
		metrics:   metricsReaderStub{},
		tasks:     taskControllerStub{},
		taskRead:  taskReaderStub{},
		traceRead: traceReaderStub{},
		groupRead: groupReaderStub{},
		notify:    notificationReaderStub{},
	}

	resp, err := server.PollTaskLogs(context.Background(), &orchestratorv1.PollTaskLogsRequest{TaskId: "task-1"})
	if err != nil {
		t.Fatalf("PollTaskLogs() error = %v", err)
	}
	items, ok := resp.GetData().AsMap()["logs"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("PollTaskLogs() unexpected response: %+v", resp.GetData().AsMap())
	}
}

func TestOrchestratorServerListTasks(t *testing.T) {
	server := &orchestratorServer{
		execution: executionSubmitterStub{},
		injection: injectionSubmitterStub{},
		metrics:   metricsReaderStub{},
		tasks:     taskControllerStub{},
		taskRead: taskReaderStub{list: &dto.ListResp[task.TaskResp]{
			Items: []task.TaskResp{{ID: "task-1"}},
			Pagination: &dto.PaginationInfo{
				Page: 1, Size: 20, Total: 1, TotalPages: 1,
			},
		}},
		traceRead: traceReaderStub{},
		groupRead: groupReaderStub{},
		notify:    notificationReaderStub{},
	}

	query, err := structpb.NewStruct(map[string]any{"page": 1, "size": 20})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	resp, err := server.ListTasks(context.Background(), &orchestratorv1.ListTasksRequest{Query: query})
	if err != nil {
		t.Fatalf("ListTasks() error = %v", err)
	}
	items, ok := resp.GetData().AsMap()["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("ListTasks() unexpected response: %+v", resp.GetData().AsMap())
	}
}

func TestOrchestratorServerGetTrace(t *testing.T) {
	server := &orchestratorServer{
		execution: executionSubmitterStub{},
		injection: injectionSubmitterStub{},
		metrics:   metricsReaderStub{},
		tasks:     taskControllerStub{},
		taskRead:  taskReaderStub{},
		traceRead: traceReaderStub{detail: &trace.TraceDetailResp{
			TraceResp: trace.TraceResp{
				ID:        "trace-1",
				Type:      "full_pipeline",
				GroupID:   "group-1",
				State:     consts.GetTraceStateName(consts.TracePending),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
				StartTime: time.Now(),
			},
		}},
		groupRead: groupReaderStub{},
		notify:    notificationReaderStub{},
	}

	resp, err := server.GetTrace(context.Background(), &orchestratorv1.GetTraceRequest{TraceId: "trace-1"})
	if err != nil {
		t.Fatalf("GetTrace() error = %v", err)
	}
	if resp.GetData().AsMap()["id"] != "trace-1" {
		t.Fatalf("GetTrace() unexpected response: %+v", resp.GetData().AsMap())
	}
}

func TestOrchestratorServerListTraces(t *testing.T) {
	server := &orchestratorServer{
		execution: executionSubmitterStub{},
		injection: injectionSubmitterStub{},
		metrics:   metricsReaderStub{},
		tasks:     taskControllerStub{},
		taskRead:  taskReaderStub{},
		traceRead: traceReaderStub{list: &dto.ListResp[trace.TraceResp]{
			Items: []trace.TraceResp{{ID: "trace-1"}},
			Pagination: &dto.PaginationInfo{
				Page: 1, Size: 20, Total: 1, TotalPages: 1,
			},
		}},
		groupRead: groupReaderStub{},
		notify:    notificationReaderStub{},
	}

	query, err := structpb.NewStruct(map[string]any{"page": 1, "size": 20})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	resp, err := server.ListTraces(context.Background(), &orchestratorv1.ListTracesRequest{Query: query})
	if err != nil {
		t.Fatalf("ListTraces() error = %v", err)
	}
	items, ok := resp.GetData().AsMap()["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("ListTraces() unexpected response: %+v", resp.GetData().AsMap())
	}
}

func TestOrchestratorServerGetGroupStats(t *testing.T) {
	server := &orchestratorServer{
		execution: executionSubmitterStub{},
		injection: injectionSubmitterStub{},
		metrics:   metricsReaderStub{},
		tasks:     taskControllerStub{},
		taskRead:  taskReaderStub{},
		traceRead: traceReaderStub{},
		groupRead: groupReaderStub{stats: &group.GroupStats{
			TotalTraces: 3,
			AvgDuration: 4.5,
		}},
		notify: notificationReaderStub{},
	}

	resp, err := server.GetGroupStats(context.Background(), &orchestratorv1.GetGroupStatsRequest{
		GroupId: "d7a4ed4b-1c91-4cdb-8af8-5520fa8d0ce0",
	})
	if err != nil {
		t.Fatalf("GetGroupStats() error = %v", err)
	}
	if resp.GetData().AsMap()["total_traces"] != float64(3) {
		t.Fatalf("GetGroupStats() unexpected response: %+v", resp.GetData().AsMap())
	}
	if resp.GetData().AsMap()["avg_duration"] != 4.5 {
		t.Fatalf("GetGroupStats() unexpected response: %+v", resp.GetData().AsMap())
	}
}

func TestOrchestratorServerTraceAndGroupStreamRPCs(t *testing.T) {
	traceMessages := []redis.XStream{{
		Stream: "trace:trace-1:log",
		Messages: []redis.XMessage{{
			ID: "1-0",
			Values: map[string]any{
				"type": "info",
			},
		}},
	}}
	groupMessages := []redis.XStream{{
		Stream: "group:group-1:log",
		Messages: []redis.XMessage{{
			ID: "2-0",
			Values: map[string]any{
				"trace_id": "trace-1",
			},
		}},
	}}

	server := &orchestratorServer{
		execution: executionSubmitterStub{},
		injection: injectionSubmitterStub{},
		metrics:   metricsReaderStub{},
		tasks:     taskControllerStub{},
		taskRead:  taskReaderStub{},
		traceRead: traceReaderStub{
			algorithms: []dto.ContainerVersionItem{{ContainerName: "algo-a"}},
			messages:   traceMessages,
		},
		groupRead: groupReaderStub{
			count:    3,
			messages: groupMessages,
		},
		notify: notificationReaderStub{},
	}

	traceState, err := server.GetTraceStreamState(context.Background(), &orchestratorv1.GetTraceStreamStateRequest{TraceId: "trace-1"})
	if err != nil {
		t.Fatalf("GetTraceStreamState() error = %v", err)
	}
	algorithms, ok := traceState.GetData().AsMap()["algorithms"].([]any)
	if !ok || len(algorithms) != 1 {
		t.Fatalf("GetTraceStreamState() unexpected response: %+v", traceState.GetData().AsMap())
	}

	traceResp, err := server.ReadTraceStreamMessages(context.Background(), &orchestratorv1.ReadStreamMessagesRequest{
		StreamKey: "trace:trace-1:log",
		LastId:    "0",
		Count:     10,
	})
	if err != nil {
		t.Fatalf("ReadTraceStreamMessages() error = %v", err)
	}
	traceItems, ok := traceResp.GetData().AsMap()["messages"].([]any)
	if !ok || len(traceItems) != 1 {
		t.Fatalf("ReadTraceStreamMessages() unexpected response: %+v", traceResp.GetData().AsMap())
	}

	groupState, err := server.GetGroupStreamState(context.Background(), &orchestratorv1.GetGroupStreamStateRequest{GroupId: "group-1"})
	if err != nil {
		t.Fatalf("GetGroupStreamState() error = %v", err)
	}
	if groupState.GetData().AsMap()["total_traces"] != float64(3) {
		t.Fatalf("GetGroupStreamState() unexpected response: %+v", groupState.GetData().AsMap())
	}

	groupResp, err := server.ReadGroupStreamMessages(context.Background(), &orchestratorv1.ReadStreamMessagesRequest{
		StreamKey: "group:group-1:log",
		LastId:    "0",
		Count:     10,
	})
	if err != nil {
		t.Fatalf("ReadGroupStreamMessages() error = %v", err)
	}
	groupItems, ok := groupResp.GetData().AsMap()["messages"].([]any)
	if !ok || len(groupItems) != 1 {
		t.Fatalf("ReadGroupStreamMessages() unexpected response: %+v", groupResp.GetData().AsMap())
	}
}

func TestOrchestratorServerReadNotificationStreamMessages(t *testing.T) {
	server := &orchestratorServer{
		execution: executionSubmitterStub{},
		injection: injectionSubmitterStub{},
		metrics:   metricsReaderStub{},
		tasks:     taskControllerStub{},
		taskRead:  taskReaderStub{},
		traceRead: traceReaderStub{},
		groupRead: groupReaderStub{},
		notify: notificationReaderStub{messages: []redis.XStream{{
			Stream: consts.NotificationStreamKey,
			Messages: []redis.XMessage{{
				ID: "3-0",
				Values: map[string]any{
					"type": "execution",
				},
			}},
		}}},
	}

	resp, err := server.ReadNotificationStreamMessages(context.Background(), &orchestratorv1.ReadStreamMessagesRequest{
		StreamKey: consts.NotificationStreamKey,
		LastId:    "0",
		Count:     10,
	})
	if err != nil {
		t.Fatalf("ReadNotificationStreamMessages() error = %v", err)
	}
	items, ok := resp.GetData().AsMap()["messages"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("ReadNotificationStreamMessages() unexpected response: %+v", resp.GetData().AsMap())
	}
}
