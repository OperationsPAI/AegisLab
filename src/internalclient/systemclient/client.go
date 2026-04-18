package systemclient

import (
	"context"
	"encoding/json"
	"fmt"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	"aegis/httpx"
	systemmodule "aegis/module/system"
	systemmetricmodule "aegis/module/systemmetric"
	systemv1 "aegis/proto/system/v1"

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
	rpc    systemv1.SystemServiceClient
}

func NewClient(lc fx.Lifecycle) (*Client, error) {
	target := config.GetString("clients.system.target")
	if target == "" {
		target = config.GetString("system.grpc.target")
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
		return nil, fmt.Errorf("create system grpc client: %w", err)
	}

	client := &Client{
		target: target,
		conn:   conn,
		rpc:    systemv1.NewSystemServiceClient(conn),
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

func (c *Client) GetHealth(ctx context.Context) (*systemmodule.HealthCheckResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("system grpc client is not configured")
	}
	resp, err := c.rpc.GetHealth(ctx, &systemv1.PingRequest{})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[systemmodule.HealthCheckResp](resp.GetData())
}

func (c *Client) GetMetrics(ctx context.Context) (*systemmodule.MonitoringMetricsResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("system grpc client is not configured")
	}
	resp, err := c.rpc.GetMetrics(ctx, &systemv1.PingRequest{})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[systemmodule.MonitoringMetricsResp](resp.GetData())
}

func (c *Client) GetSystemInfo(ctx context.Context) (*systemmodule.SystemInfo, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("system grpc client is not configured")
	}
	resp, err := c.rpc.GetSystemInfo(ctx, &systemv1.PingRequest{})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[systemmodule.SystemInfo](resp.GetData())
}

func (c *Client) ListConfigs(ctx context.Context, req *systemmodule.ListConfigReq) (*dto.ListResp[systemmodule.ConfigResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("system grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode system config list request: %w", err)
	}
	resp, err := c.rpc.ListConfigs(ctx, &systemv1.ListConfigsRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[systemmodule.ConfigResp]](resp.GetData())
}

func (c *Client) GetConfig(ctx context.Context, configID int) (*systemmodule.ConfigDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("system grpc client is not configured")
	}
	resp, err := c.rpc.GetConfig(ctx, &systemv1.GetResourceRequest{Id: int64(configID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[systemmodule.ConfigDetailResp](resp.GetData())
}

func (c *Client) ListAuditLogs(ctx context.Context, req *systemmodule.ListAuditLogReq) (*dto.ListResp[systemmodule.AuditLogResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("system grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode system audit list request: %w", err)
	}
	resp, err := c.rpc.ListAuditLogs(ctx, &systemv1.ListAuditLogsRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[systemmodule.AuditLogResp]](resp.GetData())
}

func (c *Client) GetAuditLog(ctx context.Context, auditLogID int) (*systemmodule.AuditLogDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("system grpc client is not configured")
	}
	resp, err := c.rpc.GetAuditLog(ctx, &systemv1.GetResourceRequest{Id: int64(auditLogID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[systemmodule.AuditLogDetailResp](resp.GetData())
}

func (c *Client) ListNamespaceLocks(ctx context.Context) (*systemmodule.ListNamespaceLockResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("system grpc client is not configured")
	}
	resp, err := c.rpc.ListNamespaceLocks(ctx, &systemv1.PingRequest{})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[systemmodule.ListNamespaceLockResp](resp.GetData())
}

func (c *Client) ListQueuedTasks(ctx context.Context) (*systemmodule.QueuedTasksResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("system grpc client is not configured")
	}
	resp, err := c.rpc.ListQueuedTasks(ctx, &systemv1.PingRequest{})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[systemmodule.QueuedTasksResp](resp.GetData())
}

func (c *Client) GetSystemMetrics(ctx context.Context) (*systemmetricmodule.SystemMetricsResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("system grpc client is not configured")
	}
	resp, err := c.rpc.GetSystemMetrics(ctx, &systemv1.PingRequest{})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[systemmetricmodule.SystemMetricsResp](resp.GetData())
}

func (c *Client) GetSystemMetricsHistory(ctx context.Context) (*systemmetricmodule.SystemMetricsHistoryResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("system grpc client is not configured")
	}
	resp, err := c.rpc.GetSystemMetricsHistory(ctx, &systemv1.PingRequest{})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[systemmetricmodule.SystemMetricsHistoryResp](resp.GetData())
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
		return nil, fmt.Errorf("system payload is nil")
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
		return fmt.Errorf("system rpc failed: %w", err)
	}
}
