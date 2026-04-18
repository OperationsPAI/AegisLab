package runtimeclient

import (
	"context"
	"encoding/json"
	"fmt"

	"aegis/config"
	"aegis/consts"
	"aegis/httpx"
	systemmetricmodule "aegis/module/systemmetric"
	taskmodule "aegis/module/task"
	runtimev1 "aegis/proto/runtime/v1"

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
	rpc    runtimev1.RuntimeServiceClient
}

func NewClient(lc fx.Lifecycle) (*Client, error) {
	target := config.GetString("clients.runtime.target")
	if target == "" {
		target = config.GetString("runtime_worker.grpc.target")
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
		return nil, fmt.Errorf("create runtime grpc client: %w", err)
	}

	client := &Client{
		target: target,
		conn:   conn,
		rpc:    runtimev1.NewRuntimeServiceClient(conn),
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

func (c *Client) GetNamespaceLocks(ctx context.Context) (*systemmetricmodule.ListNamespaceLockResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("runtime grpc client is not configured")
	}
	resp, err := c.rpc.GetNamespaceLocks(ctx, &runtimev1.PingRequest{})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[systemmetricmodule.ListNamespaceLockResp](resp.GetData())
}

func (c *Client) GetQueuedTasks(ctx context.Context) (*taskmodule.QueuedTasksResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("runtime grpc client is not configured")
	}
	resp, err := c.rpc.GetQueuedTasks(ctx, &runtimev1.PingRequest{})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[taskmodule.QueuedTasksResp](resp.GetData())
}

func decodeStruct[T any](payload *structpb.Struct) (*T, error) {
	if payload == nil {
		return nil, fmt.Errorf("runtime payload is nil")
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
		return fmt.Errorf("runtime rpc failed: %w", err)
	}
}
