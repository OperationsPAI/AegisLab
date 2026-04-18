package orchestratorclient

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	"aegis/httpx"
	executionmodule "aegis/module/execution"
	groupmodule "aegis/module/group"
	injectionmodule "aegis/module/injection"
	metricmodule "aegis/module/metric"
	taskmodule "aegis/module/task"
	tracemodule "aegis/module/trace"
	orchestratorv1 "aegis/proto/orchestrator/v1"

	"github.com/redis/go-redis/v9"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type Client struct {
	target string
	conn   *grpc.ClientConn
	rpc    orchestratorv1.OrchestratorServiceClient
}

func NewClient(lc fx.Lifecycle) (*Client, error) {
	target := config.GetString("clients.orchestrator.target")
	if target == "" {
		target = config.GetString("orchestrator.grpc.target")
	}
	if target == "" {
		return &Client{}, nil
	}

	conn, err := grpc.NewClient(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(httpx.UnaryClientRequestIDInterceptor()),
	)
	if err != nil {
		return nil, fmt.Errorf("create orchestrator grpc client: %w", err)
	}

	client := &Client{
		target: target,
		conn:   conn,
		rpc:    orchestratorv1.NewOrchestratorServiceClient(conn),
	}

	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			return conn.Close()
		},
	})

	return client, nil
}

func (c *Client) Enabled() bool {
	return c != nil && c.rpc != nil
}

func (c *Client) SubmitExecution(ctx context.Context, req *executionmodule.SubmitExecutionReq, groupID string, userID int) (*executionmodule.SubmitExecutionResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode submit execution request: %w", err)
	}

	resp, err := c.rpc.SubmitExecution(ctx, &orchestratorv1.SubmitExecutionRequest{
		GroupId: groupID,
		UserId:  int64(userID),
		Body:    body,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}

	items := make([]executionmodule.SubmitExecutionItem, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		mapped := executionmodule.SubmitExecutionItem{
			Index:              int(item.GetIndex()),
			TraceID:            item.GetTraceId(),
			TaskID:             item.GetTaskId(),
			AlgorithmID:        int(item.GetAlgorithmId()),
			AlgorithmVersionID: int(item.GetAlgorithmVersionId()),
		}
		if item.GetHasDatapackId() {
			value := int(item.GetDatapackId())
			mapped.DatapackID = &value
		}
		if item.GetHasDatasetId() {
			value := int(item.GetDatasetId())
			mapped.DatasetID = &value
		}
		items = append(items, mapped)
	}

	return &executionmodule.SubmitExecutionResp{
		GroupID: resp.GetGroupId(),
		Items:   items,
	}, nil
}

func (c *Client) SubmitFaultInjection(ctx context.Context, req *injectionmodule.SubmitInjectionReq, groupID string, userID int, projectID *int) (*injectionmodule.SubmitInjectionResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode submit fault injection request: %w", err)
	}

	pbReq := &orchestratorv1.SubmitFaultInjectionRequest{
		GroupId: groupID,
		UserId:  int64(userID),
		Body:    body,
	}
	if projectID != nil {
		pbReq.ProjectId = int64(*projectID)
	}

	resp, err := c.rpc.SubmitFaultInjection(ctx, pbReq)
	if err != nil {
		return nil, mapRPCError(err)
	}

	items := make([]injectionmodule.SubmitInjectionItem, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		items = append(items, injectionmodule.SubmitInjectionItem{
			Index:   int(item.GetIndex()),
			TraceID: item.GetTraceId(),
			TaskID:  item.GetTaskId(),
		})
	}

	result := &injectionmodule.SubmitInjectionResp{
		GroupID:       resp.GetGroupId(),
		Items:         items,
		OriginalCount: int(resp.GetOriginalCount()),
	}
	if warnings := resp.GetWarnings(); warnings != nil {
		result.Warnings = &injectionmodule.InjectionWarnings{
			DuplicateServicesInBatch:  warnings.GetDuplicateServicesInBatch(),
			DuplicateBatchesInRequest: int64sToInts(warnings.GetDuplicateBatchesInRequest()),
			BatchesExistInDatabase:    int64sToInts(warnings.GetBatchesExistInDatabase()),
		}
	}
	return result, nil
}

func (c *Client) SubmitDatapackBuilding(ctx context.Context, req *injectionmodule.SubmitDatapackBuildingReq, groupID string, userID int, projectID *int) (*injectionmodule.SubmitDatapackBuildingResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode submit datapack building request: %w", err)
	}

	pbReq := &orchestratorv1.SubmitDatapackBuildingRequest{
		GroupId: groupID,
		UserId:  int64(userID),
		Body:    body,
	}
	if projectID != nil {
		pbReq.ProjectId = int64(*projectID)
	}

	resp, err := c.rpc.SubmitDatapackBuilding(ctx, pbReq)
	if err != nil {
		return nil, mapRPCError(err)
	}

	items := make([]injectionmodule.SubmitBuildingItem, 0, len(resp.GetItems()))
	for _, item := range resp.GetItems() {
		items = append(items, injectionmodule.SubmitBuildingItem{
			Index:   int(item.GetIndex()),
			TraceID: item.GetTraceId(),
			TaskID:  item.GetTaskId(),
		})
	}

	return &injectionmodule.SubmitDatapackBuildingResp{
		GroupID: resp.GetGroupId(),
		Items:   items,
	}, nil
}

func (c *Client) CreateExecution(ctx context.Context, req *executionmodule.RuntimeCreateExecutionReq) (int, error) {
	if !c.Enabled() {
		return 0, fmt.Errorf("orchestrator grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return 0, fmt.Errorf("encode create execution request: %w", err)
	}
	resp, err := c.rpc.CreateExecution(ctx, &orchestratorv1.MutationRequest{Body: body})
	if err != nil {
		return 0, mapRPCError(err)
	}
	data := resp.GetData().AsMap()
	executionID, ok := data["execution_id"].(float64)
	if !ok {
		return 0, fmt.Errorf("orchestrator payload missing execution_id")
	}
	return int(executionID), nil
}

func (c *Client) CreateInjection(ctx context.Context, req *injectionmodule.RuntimeCreateInjectionReq) (*dto.InjectionItem, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode create injection request: %w", err)
	}
	resp, err := c.rpc.CreateInjection(ctx, &orchestratorv1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.InjectionItem](resp.GetData())
}

func (c *Client) UpdateExecutionState(ctx context.Context, req *executionmodule.RuntimeUpdateExecutionStateReq) error {
	if !c.Enabled() {
		return fmt.Errorf("orchestrator grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return fmt.Errorf("encode update execution state request: %w", err)
	}
	_, err = c.rpc.UpdateExecutionState(ctx, &orchestratorv1.MutationRequest{Body: body})
	if err != nil {
		return mapRPCError(err)
	}
	return nil
}

func (c *Client) UpdateInjectionState(ctx context.Context, req *injectionmodule.RuntimeUpdateInjectionStateReq) error {
	if !c.Enabled() {
		return fmt.Errorf("orchestrator grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return fmt.Errorf("encode update injection state request: %w", err)
	}
	_, err = c.rpc.UpdateInjectionState(ctx, &orchestratorv1.MutationRequest{Body: body})
	if err != nil {
		return mapRPCError(err)
	}
	return nil
}

func (c *Client) UpdateInjectionTimestamps(ctx context.Context, req *injectionmodule.RuntimeUpdateInjectionTimestampReq) (*dto.InjectionItem, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode update injection timestamps request: %w", err)
	}
	resp, err := c.rpc.UpdateInjectionTimestamps(ctx, &orchestratorv1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.InjectionItem](resp.GetData())
}

func (c *Client) GetExecution(ctx context.Context, executionID int) (*executionmodule.ExecutionDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	resp, err := c.rpc.GetExecution(ctx, &orchestratorv1.GetExecutionRequest{ExecutionId: int64(executionID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[executionmodule.ExecutionDetailResp](resp.GetData())
}

func (c *Client) GetInjectionMetrics(ctx context.Context, req *metricmodule.GetMetricsReq) (*metricmodule.InjectionMetrics, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode injection metrics request: %w", err)
	}
	resp, err := c.rpc.GetInjectionMetrics(ctx, &orchestratorv1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[metricmodule.InjectionMetrics](resp.GetData())
}

func (c *Client) GetExecutionMetrics(ctx context.Context, req *metricmodule.GetMetricsReq) (*metricmodule.ExecutionMetrics, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode execution metrics request: %w", err)
	}
	resp, err := c.rpc.GetExecutionMetrics(ctx, &orchestratorv1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[metricmodule.ExecutionMetrics](resp.GetData())
}

func (c *Client) ListProjectStatistics(ctx context.Context, projectIDs []int) (map[int]*dto.ProjectStatistics, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	resp, err := c.rpc.ListProjectStatistics(ctx, &orchestratorv1.ListProjectStatisticsRequest{
		ProjectIds: intsToInt64s(projectIDs),
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeProjectStatisticsMap(resp.GetData())
}

func (c *Client) ListEvaluationExecutionsByDatapack(ctx context.Context, req *executionmodule.EvaluationExecutionsByDatapackReq) ([]executionmodule.EvaluationExecutionItem, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode evaluation datapack query: %w", err)
	}
	resp, err := c.rpc.ListEvaluationExecutionsByDatapack(ctx, &orchestratorv1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStructItems[executionmodule.EvaluationExecutionItem](resp.GetData())
}

func (c *Client) ListEvaluationExecutionsByDataset(ctx context.Context, req *executionmodule.EvaluationExecutionsByDatasetReq) ([]executionmodule.EvaluationExecutionItem, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode evaluation dataset query: %w", err)
	}
	resp, err := c.rpc.ListEvaluationExecutionsByDataset(ctx, &orchestratorv1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStructItems[executionmodule.EvaluationExecutionItem](resp.GetData())
}

func (c *Client) GetTask(ctx context.Context, taskID string) (*taskmodule.TaskDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	resp, err := c.rpc.GetTask(ctx, &orchestratorv1.GetTaskRequest{TaskId: taskID})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[taskmodule.TaskDetailResp](resp.GetData())
}

func (c *Client) PollTaskLogs(ctx context.Context, taskID string, after time.Time) (*taskmodule.TaskLogPollResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	req := &orchestratorv1.PollTaskLogsRequest{TaskId: taskID}
	if !after.IsZero() {
		req.AfterUnixNano = after.UnixNano()
	}
	resp, err := c.rpc.PollTaskLogs(ctx, req)
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[taskmodule.TaskLogPollResp](resp.GetData())
}

func (c *Client) ListTasks(ctx context.Context, req *taskmodule.ListTaskReq) (*dto.ListResp[taskmodule.TaskResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode task list request: %w", err)
	}
	resp, err := c.rpc.ListTasks(ctx, &orchestratorv1.ListTasksRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[taskmodule.TaskResp]](resp.GetData())
}

func (c *Client) GetTrace(ctx context.Context, traceID string) (*tracemodule.TraceDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	resp, err := c.rpc.GetTrace(ctx, &orchestratorv1.GetTraceRequest{TraceId: traceID})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[tracemodule.TraceDetailResp](resp.GetData())
}

func (c *Client) ListTraces(ctx context.Context, req *tracemodule.ListTraceReq) (*dto.ListResp[tracemodule.TraceResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode trace list request: %w", err)
	}
	resp, err := c.rpc.ListTraces(ctx, &orchestratorv1.ListTracesRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[tracemodule.TraceResp]](resp.GetData())
}

func (c *Client) GetGroupStats(ctx context.Context, groupID string) (*groupmodule.GroupStats, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	resp, err := c.rpc.GetGroupStats(ctx, &orchestratorv1.GetGroupStatsRequest{GroupId: groupID})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[groupmodule.GroupStats](resp.GetData())
}

func (c *Client) GetTraceStreamAlgorithms(ctx context.Context, traceID string) ([]dto.ContainerVersionItem, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	resp, err := c.rpc.GetTraceStreamState(ctx, &orchestratorv1.GetTraceStreamStateRequest{TraceId: traceID})
	if err != nil {
		return nil, mapRPCError(err)
	}
	state, err := decodeStruct[traceStreamStateResp](resp.GetData())
	if err != nil {
		return nil, err
	}
	return state.Algorithms, nil
}

func (c *Client) ReadTraceStreamMessages(ctx context.Context, streamKey, lastID string, count int64, block time.Duration) ([]redis.XStream, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	resp, err := c.rpc.ReadTraceStreamMessages(ctx, &orchestratorv1.ReadStreamMessagesRequest{
		StreamKey:   streamKey,
		LastId:      lastID,
		Count:       count,
		BlockMillis: block.Milliseconds(),
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStreamMessages(resp.GetData(), streamKey)
}

func (c *Client) GetGroupTraceCount(ctx context.Context, groupID string) (int, error) {
	if !c.Enabled() {
		return 0, fmt.Errorf("orchestrator grpc client is not configured")
	}
	resp, err := c.rpc.GetGroupStreamState(ctx, &orchestratorv1.GetGroupStreamStateRequest{GroupId: groupID})
	if err != nil {
		return 0, mapRPCError(err)
	}
	state, err := decodeStruct[groupStreamStateResp](resp.GetData())
	if err != nil {
		return 0, err
	}
	return state.TotalTraces, nil
}

func (c *Client) ReadGroupStreamMessages(ctx context.Context, streamKey, lastID string, count int64, block time.Duration) ([]redis.XStream, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	resp, err := c.rpc.ReadGroupStreamMessages(ctx, &orchestratorv1.ReadStreamMessagesRequest{
		StreamKey:   streamKey,
		LastId:      lastID,
		Count:       count,
		BlockMillis: block.Milliseconds(),
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStreamMessages(resp.GetData(), streamKey)
}

func (c *Client) ReadNotificationStreamMessages(ctx context.Context, lastID string, count int64, block time.Duration) ([]redis.XStream, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("orchestrator grpc client is not configured")
	}
	resp, err := c.rpc.ReadNotificationStreamMessages(ctx, &orchestratorv1.ReadStreamMessagesRequest{
		StreamKey:   consts.NotificationStreamKey,
		LastId:      lastID,
		Count:       count,
		BlockMillis: block.Milliseconds(),
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStreamMessages(resp.GetData(), consts.NotificationStreamKey)
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

func toStructPB(value any) (*structpb.Struct, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}

	return structpb.NewStruct(payload)
}

func decodeStruct[T any](payload *structpb.Struct) (*T, error) {
	if payload == nil {
		return nil, fmt.Errorf("orchestrator payload is nil")
	}
	data, err := json.Marshal(payload.AsMap())
	if err != nil {
		return nil, err
	}
	var result T
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func decodeStructItems[T any](payload *structpb.Struct) ([]T, error) {
	type listEnvelope[T any] struct {
		Items []T `json:"items"`
	}

	result, err := decodeStruct[listEnvelope[T]](payload)
	if err != nil {
		return nil, err
	}
	return result.Items, nil
}

func decodeStreamMessages(payload *structpb.Struct, streamKey string) ([]redis.XStream, error) {
	result, err := decodeStruct[streamBatchResp](payload)
	if err != nil {
		return nil, err
	}
	messages := make([]redis.XMessage, 0, len(result.Messages))
	for _, item := range result.Messages {
		messages = append(messages, redis.XMessage{
			ID:     item.ID,
			Values: item.Values,
		})
	}
	return []redis.XStream{{
		Stream:   streamKey,
		Messages: messages,
	}}, nil
}

func decodeProjectStatisticsMap(payload *structpb.Struct) (map[int]*dto.ProjectStatistics, error) {
	if payload == nil {
		return map[int]*dto.ProjectStatistics{}, nil
	}
	data, err := json.Marshal(payload.AsMap())
	if err != nil {
		return nil, err
	}
	raw := map[string]dto.ProjectStatistics{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	result := make(map[int]*dto.ProjectStatistics, len(raw))
	for key, value := range raw {
		var projectID int
		if _, err := fmt.Sscanf(key, "%d", &projectID); err != nil {
			return nil, fmt.Errorf("invalid project statistics key %q: %w", key, err)
		}
		stats := value
		result[projectID] = &stats
	}
	return result, nil
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

func mapRPCError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.Unauthenticated:
		return fmt.Errorf("%w: %s", consts.ErrAuthenticationFailed, st.Message())
	case codes.PermissionDenied:
		return fmt.Errorf("%w: %s", consts.ErrPermissionDenied, st.Message())
	case codes.InvalidArgument:
		return fmt.Errorf("%w: %s", consts.ErrBadRequest, st.Message())
	case codes.NotFound:
		return fmt.Errorf("%w: %s", consts.ErrNotFound, st.Message())
	case codes.AlreadyExists:
		return fmt.Errorf("%w: %s", consts.ErrAlreadyExists, st.Message())
	default:
		return fmt.Errorf("orchestrator rpc failed: %w", err)
	}
}
