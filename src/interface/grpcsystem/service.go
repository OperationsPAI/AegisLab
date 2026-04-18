package grpcsysteminterface

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"aegis/consts"
	"aegis/dto"
	systemmodule "aegis/module/system"
	systemmetricmodule "aegis/module/systemmetric"
	systemv1 "aegis/proto/system/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

const systemServiceName = "system-service"

type systemReader interface {
	GetHealth(context.Context) (*systemmodule.HealthCheckResp, error)
	GetMetrics(context.Context) (*systemmodule.MonitoringMetricsResp, error)
	GetSystemInfo(context.Context) (*systemmodule.SystemInfo, error)
	ListNamespaceLocks(context.Context) (*systemmodule.ListNamespaceLockResp, error)
	ListQueuedTasks(context.Context) (*systemmodule.QueuedTasksResp, error)
	GetAuditLog(context.Context, int) (*systemmodule.AuditLogDetailResp, error)
	ListAuditLogs(context.Context, *systemmodule.ListAuditLogReq) (*dto.ListResp[systemmodule.AuditLogResp], error)
	GetConfig(context.Context, int) (*systemmodule.ConfigDetailResp, error)
	ListConfigs(context.Context, *systemmodule.ListConfigReq) (*dto.ListResp[systemmodule.ConfigResp], error)
}

type metricsReader interface {
	GetSystemMetrics(context.Context) (*systemmetricmodule.SystemMetricsResp, error)
	GetSystemMetricsHistory(context.Context) (*systemmetricmodule.SystemMetricsHistoryResp, error)
}

type systemServer struct {
	systemv1.UnimplementedSystemServiceServer
	system  systemReader
	metrics metricsReader
}

func newSystemServer(system *systemmodule.Service, metrics *systemmetricmodule.Service) *systemServer {
	return &systemServer{
		system:  system,
		metrics: metrics,
	}
}

func (s *systemServer) Ping(context.Context, *systemv1.PingRequest) (*systemv1.PingResponse, error) {
	return &systemv1.PingResponse{
		Service:       systemServiceName,
		AppId:         consts.AppID,
		Status:        "ok",
		TimestampUnix: time.Now().Unix(),
	}, nil
}

func (s *systemServer) GetHealth(ctx context.Context, _ *systemv1.PingRequest) (*systemv1.ResourceItemResponse, error) {
	resp, err := s.system.GetHealth(ctx)
	if err != nil {
		return nil, mapSystemError(err)
	}
	return encodeItemResponse(resp)
}

func (s *systemServer) GetMetrics(ctx context.Context, _ *systemv1.PingRequest) (*systemv1.ResourceItemResponse, error) {
	resp, err := s.system.GetMetrics(ctx)
	if err != nil {
		return nil, mapSystemError(err)
	}
	return encodeItemResponse(resp)
}

func (s *systemServer) GetSystemInfo(ctx context.Context, _ *systemv1.PingRequest) (*systemv1.ResourceItemResponse, error) {
	resp, err := s.system.GetSystemInfo(ctx)
	if err != nil {
		return nil, mapSystemError(err)
	}
	return encodeItemResponse(resp)
}

func (s *systemServer) ListConfigs(ctx context.Context, req *systemv1.ListConfigsRequest) (*systemv1.ResourceListResponse, error) {
	query, err := decodeQuery[systemmodule.ListConfigReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.system.ListConfigs(ctx, query)
	if err != nil {
		return nil, mapSystemError(err)
	}
	return encodeListResponse(resp)
}

func (s *systemServer) GetConfig(ctx context.Context, req *systemv1.GetResourceRequest) (*systemv1.ResourceItemResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	resp, err := s.system.GetConfig(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapSystemError(err)
	}
	return encodeItemResponse(resp)
}

func (s *systemServer) ListAuditLogs(ctx context.Context, req *systemv1.ListAuditLogsRequest) (*systemv1.ResourceListResponse, error) {
	query, err := decodeQuery[systemmodule.ListAuditLogReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.system.ListAuditLogs(ctx, query)
	if err != nil {
		return nil, mapSystemError(err)
	}
	return encodeListResponse(resp)
}

func (s *systemServer) GetAuditLog(ctx context.Context, req *systemv1.GetResourceRequest) (*systemv1.ResourceItemResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	resp, err := s.system.GetAuditLog(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapSystemError(err)
	}
	return encodeItemResponse(resp)
}

func (s *systemServer) ListNamespaceLocks(ctx context.Context, _ *systemv1.PingRequest) (*systemv1.ResourceItemResponse, error) {
	resp, err := s.system.ListNamespaceLocks(ctx)
	if err != nil {
		return nil, mapSystemError(err)
	}
	return encodeItemResponse(resp)
}

func (s *systemServer) ListQueuedTasks(ctx context.Context, _ *systemv1.PingRequest) (*systemv1.ResourceItemResponse, error) {
	resp, err := s.system.ListQueuedTasks(ctx)
	if err != nil {
		return nil, mapSystemError(err)
	}
	return encodeItemResponse(resp)
}

func (s *systemServer) GetSystemMetrics(ctx context.Context, _ *systemv1.PingRequest) (*systemv1.ResourceItemResponse, error) {
	resp, err := s.metrics.GetSystemMetrics(ctx)
	if err != nil {
		return nil, mapSystemError(err)
	}
	return encodeItemResponse(resp)
}

func (s *systemServer) GetSystemMetricsHistory(ctx context.Context, _ *systemv1.PingRequest) (*systemv1.ResourceItemResponse, error) {
	resp, err := s.metrics.GetSystemMetricsHistory(ctx)
	if err != nil {
		return nil, mapSystemError(err)
	}
	return encodeItemResponse(resp)
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

func encodeItemResponse(value any) (*systemv1.ResourceItemResponse, error) {
	item, err := toStruct(value)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &systemv1.ResourceItemResponse{Data: item}, nil
}

func encodeListResponse(value any) (*systemv1.ResourceListResponse, error) {
	item, err := toStruct(value)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &systemv1.ResourceListResponse{Data: item}, nil
}

func toStruct(value any) (*structpb.Struct, error) {
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

func mapSystemError(err error) error {
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
