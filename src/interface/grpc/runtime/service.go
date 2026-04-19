package grpcruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"aegis/consts"
	"aegis/dto"
	buildkit "aegis/infra/buildkit"
	helm "aegis/infra/helm"
	k8s "aegis/infra/k8s"
	redis "aegis/infra/redis"
	task "aegis/module/task"
	runtimev1 "aegis/proto/runtime/v1"
	"aegis/service/consumer"

	"go.uber.org/fx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"gorm.io/gorm"
)

type runtimeServerParams struct {
	fx.In

	DB             *gorm.DB
	RedisGateway   *redis.Gateway
	K8sGateway     *k8s.Gateway
	BuildKit       *buildkit.Gateway
	Helm           *helm.Gateway
	RestartLimiter *consumer.TokenBucketRateLimiter `name:"restart_limiter"`
	BuildLimiter   *consumer.TokenBucketRateLimiter `name:"build_limiter"`
	AlgoLimiter    *consumer.TokenBucketRateLimiter `name:"algo_limiter"`
}

type runtimeServer struct {
	runtimev1.UnimplementedRuntimeServiceServer
	snapshots *consumer.RuntimeSnapshotService
	redis     *redis.Gateway
}

func newRuntimeServer(params runtimeServerParams) *runtimeServer {
	return &runtimeServer{
		snapshots: consumer.NewRuntimeSnapshotService(
			params.DB,
			params.RedisGateway,
			params.K8sGateway,
			params.BuildKit,
			params.Helm,
			params.RestartLimiter,
			params.BuildLimiter,
			params.AlgoLimiter,
		),
		redis: params.RedisGateway,
	}
}

func (s *runtimeServer) Ping(ctx context.Context, _ *runtimev1.PingRequest) (*runtimev1.PingResponse, error) {
	status := s.snapshots.RuntimeStatus(ctx)
	return &runtimev1.PingResponse{
		Service:       status.ServiceName,
		AppId:         status.AppID,
		Status:        "ok",
		TimestampUnix: time.Now().Unix(),
	}, nil
}

func (s *runtimeServer) GetRuntimeStatus(ctx context.Context, _ *runtimev1.RuntimeStatusRequest) (*runtimev1.RuntimeStatusResponse, error) {
	status := s.snapshots.RuntimeStatus(ctx)
	return &runtimev1.RuntimeStatusResponse{
		Service:           status.ServiceName,
		Mode:              status.Mode,
		AppId:             status.AppID,
		StartedAtUnix:     status.StartedAt.Unix(),
		UptimeSeconds:     status.UptimeSeconds,
		DbAvailable:       status.DB.Available,
		DbHealthy:         status.DB.Healthy,
		DbError:           status.DB.Error,
		RedisAvailable:    status.Redis.Available,
		RedisHealthy:      status.Redis.Healthy,
		RedisError:        status.Redis.Error,
		K8SAvailable:      status.K8s.Available,
		K8SHealthy:        status.K8s.Healthy,
		K8SError:          status.K8s.Error,
		BuildkitAvailable: status.BuildKit.Available,
		BuildkitHealthy:   status.BuildKit.Healthy,
		BuildkitError:     status.BuildKit.Error,
		HelmAvailable:     status.Helm.Available,
		HelmHealthy:       status.Helm.Healthy,
		HelmError:         status.Helm.Error,
	}, nil
}

func (s *runtimeServer) GetQueueStatus(ctx context.Context, _ *runtimev1.QueueStatusRequest) (*runtimev1.QueueStatusResponse, error) {
	stats, err := s.snapshots.QueueStatus(ctx)
	if err != nil {
		return nil, err
	}
	return &runtimev1.QueueStatusResponse{
		ReadyCount:       stats.ReadyCount,
		DelayedCount:     stats.DelayedCount,
		DeadCount:        stats.DeadCount,
		IndexedCount:     stats.IndexedCount,
		ConcurrencyCount: stats.ConcurrencyCount,
	}, nil
}

func (s *runtimeServer) GetLimiterStatus(ctx context.Context, _ *runtimev1.LimiterStatusRequest) (*runtimev1.LimiterStatusResponse, error) {
	snapshots := s.snapshots.LimiterStatus(ctx)
	items := make([]*runtimev1.LimiterStatus, 0, len(snapshots))
	for _, snapshot := range snapshots {
		item := &runtimev1.LimiterStatus{
			ServiceName:        snapshot.ServiceName,
			BucketKey:          snapshot.BucketKey,
			MaxTokens:          int64(snapshot.MaxTokens),
			WaitTimeoutSeconds: int64(snapshot.WaitTimeout.Seconds()),
			InUseTokens:        snapshot.InUseTokens,
		}
		if snapshot.InUseTokensLoadErr != nil {
			item.Error = snapshot.InUseTokensLoadErr.Error()
		}
		items = append(items, item)
	}
	return &runtimev1.LimiterStatusResponse{Items: items}, nil
}

func (s *runtimeServer) GetNamespaceLocks(ctx context.Context, _ *runtimev1.PingRequest) (*runtimev1.StructResponse, error) {
	items, err := listNamespaceLocks(ctx, s.redis)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return encodeStruct(items)
}

func (s *runtimeServer) GetQueuedTasks(ctx context.Context, _ *runtimev1.PingRequest) (*runtimev1.StructResponse, error) {
	items, err := listQueuedTasks(ctx, s.redis)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return encodeStruct(items)
}

func listNamespaceLocks(ctx context.Context, redis *redis.Gateway) (map[string]map[string]any, error) {
	namespaces, err := redis.SetMembers(ctx, consts.NamespacesKey)
	if err != nil {
		return nil, err
	}

	items := make(map[string]map[string]any, len(namespaces))
	for _, namespace := range namespaces {
		nsKey := fmt.Sprintf(consts.NamespaceKeyPattern, namespace)
		values, err := redis.HashGetAll(ctx, nsKey)
		if err != nil {
			return nil, err
		}
		entry := make(map[string]any, len(values))
		for key, value := range values {
			entry[key] = value
		}
		items[namespace] = entry
	}
	return items, nil
}

func listQueuedTasks(ctx context.Context, redis *redis.Gateway) (map[string]any, error) {
	readyItems, err := redis.ListReadyTasks(ctx)
	if err != nil {
		return nil, err
	}
	delayedItems, err := redis.ListDelayedTasks(ctx, 1000)
	if err != nil {
		return nil, err
	}

	readyTasks, err := decodeQueuedTasks(readyItems)
	if err != nil {
		return nil, err
	}
	delayedTasks, err := decodeQueuedTasks(delayedItems)
	if err != nil {
		return nil, err
	}

	return map[string]any{
		"ready_tasks":   readyTasks,
		"delayed_tasks": delayedTasks,
	}, nil
}

func decodeQueuedTasks(items []string) ([]task.TaskResp, error) {
	result := make([]task.TaskResp, 0, len(items))
	for _, item := range items {
		var queuedTask dto.UnifiedTask
		if err := json.Unmarshal([]byte(item), &queuedTask); err != nil {
			return nil, err
		}
		result = append(result, task.TaskResp{
			ID:          queuedTask.TaskID,
			Type:        consts.GetTaskTypeName(queuedTask.Type),
			Immediate:   queuedTask.Immediate,
			ExecuteTime: queuedTask.ExecuteTime,
			CronExpr:    queuedTask.CronExpr,
			TraceID:     queuedTask.TraceID,
			GroupID:     queuedTask.GroupID,
			State:       consts.GetTaskStateName(queuedTask.State),
			Status:      consts.GetStatusTypeName(consts.CommonEnabled),
			ProjectID:   queuedTask.ProjectID,
		})
	}
	return result, nil
}

func encodeStruct(value any) (*runtimev1.StructResponse, error) {
	data, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	payload := map[string]any{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	item, err := structpb.NewStruct(payload)
	if err != nil {
		return nil, err
	}
	return &runtimev1.StructResponse{Data: item}, nil
}
