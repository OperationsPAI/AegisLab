package resourceclient

import (
	"context"
	"encoding/json"
	"fmt"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	"aegis/httpx"
	chaossystem "aegis/module/chaossystem"
	container "aegis/module/container"
	dataset "aegis/module/dataset"
	evaluation "aegis/module/evaluation"
	label "aegis/module/label"
	project "aegis/module/project"
	resourcev1 "aegis/proto/resource/v1"

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
	rpc    resourcev1.ResourceServiceClient
}

func NewClient(lc fx.Lifecycle) (*Client, error) {
	target := config.GetString("clients.resource.target")
	if target == "" {
		target = config.GetString("resource.grpc.target")
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
		return nil, fmt.Errorf("create resource grpc client: %w", err)
	}

	client := &Client{
		target: target,
		conn:   conn,
		rpc:    resourcev1.NewResourceServiceClient(conn),
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

func (c *Client) ListProjects(ctx context.Context, req *project.ListProjectReq) (*dto.ListResp[project.ProjectResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode project list request: %w", err)
	}
	resp, err := c.rpc.ListProjects(ctx, &resourcev1.ListProjectsRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[project.ProjectResp]](resp.GetData())
}

func (c *Client) GetProject(ctx context.Context, projectID int) (*project.ProjectDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	resp, err := c.rpc.GetProject(ctx, &resourcev1.GetResourceRequest{Id: int64(projectID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[project.ProjectDetailResp](resp.GetData())
}

func (c *Client) ListContainers(ctx context.Context, req *container.ListContainerReq) (*dto.ListResp[container.ContainerResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode container list request: %w", err)
	}
	resp, err := c.rpc.ListContainers(ctx, &resourcev1.ListContainersRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[container.ContainerResp]](resp.GetData())
}

func (c *Client) GetContainer(ctx context.Context, containerID int) (*container.ContainerDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	resp, err := c.rpc.GetContainer(ctx, &resourcev1.GetResourceRequest{Id: int64(containerID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[container.ContainerDetailResp](resp.GetData())
}

func (c *Client) ListDatasets(ctx context.Context, req *dataset.ListDatasetReq) (*dto.ListResp[dataset.DatasetResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode dataset list request: %w", err)
	}
	resp, err := c.rpc.ListDatasets(ctx, &resourcev1.ListDatasetsRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[dataset.DatasetResp]](resp.GetData())
}

func (c *Client) GetDataset(ctx context.Context, datasetID int) (*dataset.DatasetDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	resp, err := c.rpc.GetDataset(ctx, &resourcev1.GetResourceRequest{Id: int64(datasetID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dataset.DatasetDetailResp](resp.GetData())
}

func (c *Client) CreateLabel(ctx context.Context, req *label.CreateLabelReq) (*label.LabelResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode label create request: %w", err)
	}
	resp, err := c.rpc.CreateLabel(ctx, &resourcev1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[label.LabelResp](resp.GetData())
}

func (c *Client) GetLabel(ctx context.Context, labelID int) (*label.LabelDetailResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	resp, err := c.rpc.GetLabel(ctx, &resourcev1.GetResourceRequest{Id: int64(labelID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[label.LabelDetailResp](resp.GetData())
}

func (c *Client) ListLabels(ctx context.Context, req *label.ListLabelReq) (*dto.ListResp[label.LabelResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode label list request: %w", err)
	}
	resp, err := c.rpc.ListLabels(ctx, &resourcev1.QueryRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[label.LabelResp]](resp.GetData())
}

func (c *Client) UpdateLabel(ctx context.Context, req *label.UpdateLabelReq, labelID int) (*label.LabelResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode label update request: %w", err)
	}
	resp, err := c.rpc.UpdateLabel(ctx, &resourcev1.UpdateByIDRequest{
		Id:   int64(labelID),
		Body: body,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[label.LabelResp](resp.GetData())
}

func (c *Client) DeleteLabel(ctx context.Context, labelID int) error {
	if !c.Enabled() {
		return fmt.Errorf("resource grpc client is not configured")
	}
	_, err := c.rpc.DeleteLabel(ctx, &resourcev1.GetResourceRequest{Id: int64(labelID)})
	if err != nil {
		return mapRPCError(err)
	}
	return nil
}

func (c *Client) BatchDeleteLabels(ctx context.Context, ids []int) error {
	if !c.Enabled() {
		return fmt.Errorf("resource grpc client is not configured")
	}
	_, err := c.rpc.BatchDeleteLabels(ctx, &resourcev1.BatchDeleteRequest{Ids: intsToInt64s(ids)})
	if err != nil {
		return mapRPCError(err)
	}
	return nil
}

func (c *Client) ListChaosSystems(ctx context.Context, req *chaossystem.ListChaosSystemReq) (*dto.ListResp[chaossystem.ChaosSystemResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode chaos system list request: %w", err)
	}
	resp, err := c.rpc.ListChaosSystems(ctx, &resourcev1.QueryRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[chaossystem.ChaosSystemResp]](resp.GetData())
}

func (c *Client) GetChaosSystem(ctx context.Context, systemID int) (*chaossystem.ChaosSystemResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	resp, err := c.rpc.GetChaosSystem(ctx, &resourcev1.GetResourceRequest{Id: int64(systemID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[chaossystem.ChaosSystemResp](resp.GetData())
}

func (c *Client) CreateChaosSystem(ctx context.Context, req *chaossystem.CreateChaosSystemReq) (*chaossystem.ChaosSystemResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode chaos system create request: %w", err)
	}
	resp, err := c.rpc.CreateChaosSystem(ctx, &resourcev1.MutationRequest{Body: body})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[chaossystem.ChaosSystemResp](resp.GetData())
}

func (c *Client) UpdateChaosSystem(ctx context.Context, req *chaossystem.UpdateChaosSystemReq, systemID int) (*chaossystem.ChaosSystemResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode chaos system update request: %w", err)
	}
	resp, err := c.rpc.UpdateChaosSystem(ctx, &resourcev1.UpdateByIDRequest{
		Id:   int64(systemID),
		Body: body,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[chaossystem.ChaosSystemResp](resp.GetData())
}

func (c *Client) DeleteChaosSystem(ctx context.Context, systemID int) error {
	if !c.Enabled() {
		return fmt.Errorf("resource grpc client is not configured")
	}
	_, err := c.rpc.DeleteChaosSystem(ctx, &resourcev1.GetResourceRequest{Id: int64(systemID)})
	if err != nil {
		return mapRPCError(err)
	}
	return nil
}

func (c *Client) UpsertChaosSystemMetadata(ctx context.Context, systemID int, req *chaossystem.BulkUpsertSystemMetadataReq) error {
	if !c.Enabled() {
		return fmt.Errorf("resource grpc client is not configured")
	}
	body, err := toStructPB(req)
	if err != nil {
		return fmt.Errorf("encode chaos system metadata request: %w", err)
	}
	_, err = c.rpc.UpsertChaosSystemMetadata(ctx, &resourcev1.UpdateByIDRequest{
		Id:   int64(systemID),
		Body: body,
	})
	if err != nil {
		return mapRPCError(err)
	}
	return nil
}

func (c *Client) ListChaosSystemMetadata(ctx context.Context, systemID int, metadataType string) ([]chaossystem.SystemMetadataResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	query, err := toStructPB(map[string]any{"type": metadataType})
	if err != nil {
		return nil, fmt.Errorf("encode chaos system metadata query: %w", err)
	}
	resp, err := c.rpc.ListChaosSystemMetadata(ctx, &resourcev1.IDQueryRequest{
		Id:    int64(systemID),
		Query: query,
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	items, err := decodeStruct[struct {
		Items []chaossystem.SystemMetadataResp `json:"items"`
	}](resp.GetData())
	if err != nil {
		return nil, err
	}
	return items.Items, nil
}

func (c *Client) ListDatapackEvaluationResults(ctx context.Context, req *evaluation.BatchEvaluateDatapackReq, userID int) (*evaluation.BatchEvaluateDatapackResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode datapack evaluation request: %w", err)
	}
	resp, err := c.rpc.ListDatapackEvaluationResults(ctx, &resourcev1.ListDatapackEvaluationsRequest{
		Query:  query,
		UserId: int64(userID),
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[evaluation.BatchEvaluateDatapackResp](resp.GetData())
}

func (c *Client) ListDatasetEvaluationResults(ctx context.Context, req *evaluation.BatchEvaluateDatasetReq, userID int) (*evaluation.BatchEvaluateDatasetResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode dataset evaluation request: %w", err)
	}
	resp, err := c.rpc.ListDatasetEvaluationResults(ctx, &resourcev1.ListDatasetEvaluationsRequest{
		Query:  query,
		UserId: int64(userID),
	})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[evaluation.BatchEvaluateDatasetResp](resp.GetData())
}

func (c *Client) ListEvaluations(ctx context.Context, req *evaluation.ListEvaluationReq) (*dto.ListResp[evaluation.EvaluationResp], error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	query, err := toStructPB(req)
	if err != nil {
		return nil, fmt.Errorf("encode evaluation list request: %w", err)
	}
	resp, err := c.rpc.ListEvaluations(ctx, &resourcev1.ListEvaluationsRequest{Query: query})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[dto.ListResp[evaluation.EvaluationResp]](resp.GetData())
}

func (c *Client) GetEvaluation(ctx context.Context, evaluationID int) (*evaluation.EvaluationResp, error) {
	if !c.Enabled() {
		return nil, fmt.Errorf("resource grpc client is not configured")
	}
	resp, err := c.rpc.GetEvaluation(ctx, &resourcev1.GetResourceRequest{Id: int64(evaluationID)})
	if err != nil {
		return nil, mapRPCError(err)
	}
	return decodeStruct[evaluation.EvaluationResp](resp.GetData())
}

func (c *Client) DeleteEvaluation(ctx context.Context, evaluationID int) error {
	if !c.Enabled() {
		return fmt.Errorf("resource grpc client is not configured")
	}
	_, err := c.rpc.DeleteEvaluation(ctx, &resourcev1.GetResourceRequest{Id: int64(evaluationID)})
	if err != nil {
		return mapRPCError(err)
	}
	return nil
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
		return nil, fmt.Errorf("resource payload is nil")
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
		return fmt.Errorf("resource rpc failed: %w", err)
	}
}
