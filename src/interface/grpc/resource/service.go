package grpcresource

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"aegis/consts"
	"aegis/dto"
	chaossystem "aegis/module/chaossystem"
	container "aegis/module/container"
	dataset "aegis/module/dataset"
	evaluation "aegis/module/evaluation"
	label "aegis/module/label"
	project "aegis/module/project"
	resourcev1 "aegis/proto/resource/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/structpb"
)

const resourceServiceName = "resource-service"

type projectReader interface {
	GetProjectDetail(context.Context, int) (*project.ProjectDetailResp, error)
	ListProjects(context.Context, *project.ListProjectReq) (*dto.ListResp[project.ProjectResp], error)
}

type containerReader interface {
	GetContainer(context.Context, int) (*container.ContainerDetailResp, error)
	ListContainers(context.Context, *container.ListContainerReq) (*dto.ListResp[container.ContainerResp], error)
}

type datasetReader interface {
	GetDataset(context.Context, int) (*dataset.DatasetDetailResp, error)
	ListDatasets(context.Context, *dataset.ListDatasetReq) (*dto.ListResp[dataset.DatasetResp], error)
}

type evaluationReader interface {
	ListDatapackEvaluationResults(context.Context, *evaluation.BatchEvaluateDatapackReq, int) (*evaluation.BatchEvaluateDatapackResp, error)
	ListDatasetEvaluationResults(context.Context, *evaluation.BatchEvaluateDatasetReq, int) (*evaluation.BatchEvaluateDatasetResp, error)
	ListEvaluations(context.Context, *evaluation.ListEvaluationReq) (*dto.ListResp[evaluation.EvaluationResp], error)
	GetEvaluation(context.Context, int) (*evaluation.EvaluationResp, error)
	DeleteEvaluation(context.Context, int) error
}

type labelReader interface {
	BatchDelete(context.Context, []int) error
	Create(context.Context, *label.CreateLabelReq) (*label.LabelResp, error)
	Delete(context.Context, int) error
	GetDetail(context.Context, int) (*label.LabelDetailResp, error)
	List(context.Context, *label.ListLabelReq) (*dto.ListResp[label.LabelResp], error)
	Update(context.Context, *label.UpdateLabelReq, int) (*label.LabelResp, error)
}

type chaosSystemReader interface {
	ListSystems(context.Context, *chaossystem.ListChaosSystemReq) (*dto.ListResp[chaossystem.ChaosSystemResp], error)
	GetSystem(context.Context, int) (*chaossystem.ChaosSystemResp, error)
	CreateSystem(context.Context, *chaossystem.CreateChaosSystemReq) (*chaossystem.ChaosSystemResp, error)
	UpdateSystem(context.Context, int, *chaossystem.UpdateChaosSystemReq) (*chaossystem.ChaosSystemResp, error)
	DeleteSystem(context.Context, int) error
	UpsertMetadata(context.Context, int, *chaossystem.BulkUpsertSystemMetadataReq) error
	ListMetadata(context.Context, int, string) ([]chaossystem.SystemMetadataResp, error)
}

type chaosSystemMetadataListResponse struct {
	Items []chaossystem.SystemMetadataResp `json:"items"`
}

type resourceServer struct {
	resourcev1.UnimplementedResourceServiceServer
	projects     projectReader
	containers   containerReader
	datasets     datasetReader
	labels       labelReader
	chaosSystems chaosSystemReader
	evaluations  evaluationReader
}

func newResourceServer(
	projects *project.Service,
	containers *container.Service,
	datasets *dataset.Service,
	labels label.HandlerService,
	chaosSystems chaossystem.HandlerService,
	evaluations *evaluation.Service,
) *resourceServer {
	return &resourceServer{
		projects:     projects,
		containers:   containers,
		datasets:     datasets,
		labels:       labels,
		chaosSystems: chaosSystems,
		evaluations:  evaluations,
	}
}

func (s *resourceServer) Ping(context.Context, *resourcev1.PingRequest) (*resourcev1.PingResponse, error) {
	return &resourcev1.PingResponse{
		Service:       resourceServiceName,
		AppId:         consts.AppID,
		Status:        "ok",
		TimestampUnix: time.Now().Unix(),
	}, nil
}

func (s *resourceServer) ListProjects(ctx context.Context, req *resourcev1.ListProjectsRequest) (*resourcev1.ResourceListResponse, error) {
	query, err := decodeQuery[project.ListProjectReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.projects.ListProjects(ctx, query)
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeListResponse(resp)
}

func (s *resourceServer) GetProject(ctx context.Context, req *resourcev1.GetResourceRequest) (*resourcev1.ResourceItemResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	resp, err := s.projects.GetProjectDetail(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(resp)
}

func (s *resourceServer) ListContainers(ctx context.Context, req *resourcev1.ListContainersRequest) (*resourcev1.ResourceListResponse, error) {
	query, err := decodeQuery[container.ListContainerReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.containers.ListContainers(ctx, query)
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeListResponse(resp)
}

func (s *resourceServer) GetContainer(ctx context.Context, req *resourcev1.GetResourceRequest) (*resourcev1.ResourceItemResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	resp, err := s.containers.GetContainer(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(resp)
}

func (s *resourceServer) ListDatasets(ctx context.Context, req *resourcev1.ListDatasetsRequest) (*resourcev1.ResourceListResponse, error) {
	query, err := decodeQuery[dataset.ListDatasetReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.datasets.ListDatasets(ctx, query)
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeListResponse(resp)
}

func (s *resourceServer) GetDataset(ctx context.Context, req *resourcev1.GetResourceRequest) (*resourcev1.ResourceItemResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	resp, err := s.datasets.GetDataset(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(resp)
}

func (s *resourceServer) CreateLabel(ctx context.Context, req *resourcev1.MutationRequest) (*resourcev1.ResourceItemResponse, error) {
	body, err := decodeQuery[label.CreateLabelReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.labels.Create(ctx, body)
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(resp)
}

func (s *resourceServer) GetLabel(ctx context.Context, req *resourcev1.GetResourceRequest) (*resourcev1.ResourceItemResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	resp, err := s.labels.GetDetail(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(resp)
}

func (s *resourceServer) ListLabels(ctx context.Context, req *resourcev1.QueryRequest) (*resourcev1.ResourceListResponse, error) {
	query, err := decodeQuery[label.ListLabelReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.labels.List(ctx, query)
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeListResponse(resp)
}

func (s *resourceServer) UpdateLabel(ctx context.Context, req *resourcev1.UpdateByIDRequest) (*resourcev1.ResourceItemResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	body, err := decodeQuery[label.UpdateLabelReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := body.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.labels.Update(ctx, body, int(req.GetId()))
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(resp)
}

func (s *resourceServer) DeleteLabel(ctx context.Context, req *resourcev1.GetResourceRequest) (*emptypb.Empty, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if err := s.labels.Delete(ctx, int(req.GetId())); err != nil {
		return nil, mapResourceError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *resourceServer) BatchDeleteLabels(ctx context.Context, req *resourcev1.BatchDeleteRequest) (*emptypb.Empty, error) {
	if err := validatePositiveInt64s(req.GetIds(), "ids"); err != nil {
		return nil, err
	}
	if err := s.labels.BatchDelete(ctx, int64sToInts(req.GetIds())); err != nil {
		return nil, mapResourceError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *resourceServer) ListChaosSystems(ctx context.Context, req *resourcev1.QueryRequest) (*resourcev1.ResourceListResponse, error) {
	query, err := decodeQuery[chaossystem.ListChaosSystemReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.chaosSystems.ListSystems(ctx, query)
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeListResponse(resp)
}

func (s *resourceServer) GetChaosSystem(ctx context.Context, req *resourcev1.GetResourceRequest) (*resourcev1.ResourceItemResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	resp, err := s.chaosSystems.GetSystem(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(resp)
}

func (s *resourceServer) CreateChaosSystem(ctx context.Context, req *resourcev1.MutationRequest) (*resourcev1.ResourceItemResponse, error) {
	body, err := decodeQuery[chaossystem.CreateChaosSystemReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.chaosSystems.CreateSystem(ctx, body)
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(resp)
}

func (s *resourceServer) UpdateChaosSystem(ctx context.Context, req *resourcev1.UpdateByIDRequest) (*resourcev1.ResourceItemResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	body, err := decodeQuery[chaossystem.UpdateChaosSystemReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.chaosSystems.UpdateSystem(ctx, int(req.GetId()), body)
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(resp)
}

func (s *resourceServer) DeleteChaosSystem(ctx context.Context, req *resourcev1.GetResourceRequest) (*emptypb.Empty, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if err := s.chaosSystems.DeleteSystem(ctx, int(req.GetId())); err != nil {
		return nil, mapResourceError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *resourceServer) UpsertChaosSystemMetadata(ctx context.Context, req *resourcev1.UpdateByIDRequest) (*emptypb.Empty, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	body, err := decodeQuery[chaossystem.BulkUpsertSystemMetadataReq](req.GetBody())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := s.chaosSystems.UpsertMetadata(ctx, int(req.GetId()), body); err != nil {
		return nil, mapResourceError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *resourceServer) ListChaosSystemMetadata(ctx context.Context, req *resourcev1.IDQueryRequest) (*resourcev1.ResourceItemResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	query, err := decodeQuery[struct {
		Type string `json:"type"`
	}](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.chaosSystems.ListMetadata(ctx, int(req.GetId()), query.Type)
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(chaosSystemMetadataListResponse{Items: resp})
}

func (s *resourceServer) ListDatapackEvaluationResults(ctx context.Context, req *resourcev1.ListDatapackEvaluationsRequest) (*resourcev1.ResourceItemResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	query, err := decodeQuery[evaluation.BatchEvaluateDatapackReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.evaluations.ListDatapackEvaluationResults(ctx, query, int(req.GetUserId()))
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(resp)
}

func (s *resourceServer) ListDatasetEvaluationResults(ctx context.Context, req *resourcev1.ListDatasetEvaluationsRequest) (*resourcev1.ResourceItemResponse, error) {
	if req.GetUserId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	query, err := decodeQuery[evaluation.BatchEvaluateDatasetReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.evaluations.ListDatasetEvaluationResults(ctx, query, int(req.GetUserId()))
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(resp)
}

func (s *resourceServer) ListEvaluations(ctx context.Context, req *resourcev1.ListEvaluationsRequest) (*resourcev1.ResourceListResponse, error) {
	query, err := decodeQuery[evaluation.ListEvaluationReq](req.GetQuery())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}
	if err := query.Validate(); err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	resp, err := s.evaluations.ListEvaluations(ctx, query)
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeListResponse(resp)
}

func (s *resourceServer) GetEvaluation(ctx context.Context, req *resourcev1.GetResourceRequest) (*resourcev1.ResourceItemResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}

	resp, err := s.evaluations.GetEvaluation(ctx, int(req.GetId()))
	if err != nil {
		return nil, mapResourceError(err)
	}
	return encodeItemResponse(resp)
}

func (s *resourceServer) DeleteEvaluation(ctx context.Context, req *resourcev1.GetResourceRequest) (*emptypb.Empty, error) {
	if req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id is required")
	}
	if err := s.evaluations.DeleteEvaluation(ctx, int(req.GetId())); err != nil {
		return nil, mapResourceError(err)
	}
	return &emptypb.Empty{}, nil
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

func encodeItemResponse(value any) (*resourcev1.ResourceItemResponse, error) {
	item, err := toStruct(value)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &resourcev1.ResourceItemResponse{Data: item}, nil
}

func encodeListResponse(value any) (*resourcev1.ResourceListResponse, error) {
	item, err := toStruct(value)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &resourcev1.ResourceListResponse{Data: item}, nil
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

func mapResourceError(err error) error {
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

func validatePositiveInt64s(items []int64, field string) error {
	if len(items) == 0 {
		return status.Errorf(codes.InvalidArgument, "%s is required", field)
	}
	for _, item := range items {
		if item <= 0 {
			return status.Errorf(codes.InvalidArgument, "%s must contain positive integers", field)
		}
	}
	return nil
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
