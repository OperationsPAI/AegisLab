package grpcresourceinterface

import (
	"context"
	"errors"
	"testing"

	"aegis/consts"
	"aegis/dto"
	chaossystemmodule "aegis/module/chaossystem"
	containermodule "aegis/module/container"
	datasetmodule "aegis/module/dataset"
	evaluationmodule "aegis/module/evaluation"
	labelmodule "aegis/module/label"
	projectmodule "aegis/module/project"
	resourcev1 "aegis/proto/resource/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

type projectReaderStub struct {
	listResp *dto.ListResp[projectmodule.ProjectResp]
	getResp  *projectmodule.ProjectDetailResp
	err      error
}

func (s projectReaderStub) GetProjectDetail(_ context.Context, projectID int) (*projectmodule.ProjectDetailResp, error) {
	if projectID <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.getResp, s.err
}

func (s projectReaderStub) ListProjects(_ context.Context, req *projectmodule.ListProjectReq) (*dto.ListResp[projectmodule.ProjectResp], error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return s.listResp, s.err
}

type containerReaderStub struct {
	listResp *dto.ListResp[containermodule.ContainerResp]
	getResp  *containermodule.ContainerDetailResp
	err      error
}

func (s containerReaderStub) GetContainer(_ context.Context, containerID int) (*containermodule.ContainerDetailResp, error) {
	if containerID <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.getResp, s.err
}

func (s containerReaderStub) ListContainers(_ context.Context, req *containermodule.ListContainerReq) (*dto.ListResp[containermodule.ContainerResp], error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return s.listResp, s.err
}

type datasetReaderStub struct {
	listResp *dto.ListResp[datasetmodule.DatasetResp]
	getResp  *datasetmodule.DatasetDetailResp
	err      error
}

func (s datasetReaderStub) GetDataset(_ context.Context, datasetID int) (*datasetmodule.DatasetDetailResp, error) {
	if datasetID <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.getResp, s.err
}

func (s datasetReaderStub) ListDatasets(_ context.Context, req *datasetmodule.ListDatasetReq) (*dto.ListResp[datasetmodule.DatasetResp], error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return s.listResp, s.err
}

type evaluationReaderStub struct {
	datapackResp *evaluationmodule.BatchEvaluateDatapackResp
	datasetResp  *evaluationmodule.BatchEvaluateDatasetResp
	listResp     *dto.ListResp[evaluationmodule.EvaluationResp]
	getResp      *evaluationmodule.EvaluationResp
	err          error
}

func (s evaluationReaderStub) ListDatapackEvaluationResults(_ context.Context, req *evaluationmodule.BatchEvaluateDatapackReq, userID int) (*evaluationmodule.BatchEvaluateDatapackResp, error) {
	if req == nil || userID <= 0 {
		return nil, errors.New("invalid request")
	}
	return s.datapackResp, s.err
}

func (s evaluationReaderStub) ListDatasetEvaluationResults(_ context.Context, req *evaluationmodule.BatchEvaluateDatasetReq, userID int) (*evaluationmodule.BatchEvaluateDatasetResp, error) {
	if req == nil || userID <= 0 {
		return nil, errors.New("invalid request")
	}
	return s.datasetResp, s.err
}

func (s evaluationReaderStub) ListEvaluations(_ context.Context, req *evaluationmodule.ListEvaluationReq) (*dto.ListResp[evaluationmodule.EvaluationResp], error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return s.listResp, s.err
}

func (s evaluationReaderStub) GetEvaluation(_ context.Context, id int) (*evaluationmodule.EvaluationResp, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.getResp, s.err
}

func (s evaluationReaderStub) DeleteEvaluation(_ context.Context, id int) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return s.err
}

type labelReaderStub struct {
	listResp *dto.ListResp[labelmodule.LabelResp]
	getResp  *labelmodule.LabelDetailResp
	itemResp *labelmodule.LabelResp
	err      error
}

func (s labelReaderStub) BatchDelete(_ context.Context, ids []int) error {
	if len(ids) == 0 {
		return errors.New("ids required")
	}
	return s.err
}

func (s labelReaderStub) Create(_ context.Context, req *labelmodule.CreateLabelReq) (*labelmodule.LabelResp, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return s.itemResp, s.err
}

func (s labelReaderStub) Delete(_ context.Context, id int) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return s.err
}

func (s labelReaderStub) GetDetail(_ context.Context, id int) (*labelmodule.LabelDetailResp, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.getResp, s.err
}

func (s labelReaderStub) List(_ context.Context, req *labelmodule.ListLabelReq) (*dto.ListResp[labelmodule.LabelResp], error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return s.listResp, s.err
}

func (s labelReaderStub) Update(_ context.Context, req *labelmodule.UpdateLabelReq, id int) (*labelmodule.LabelResp, error) {
	if req == nil || id <= 0 {
		return nil, errors.New("invalid request")
	}
	return s.itemResp, s.err
}

type chaosSystemReaderStub struct {
	listResp     *dto.ListResp[chaossystemmodule.ChaosSystemResp]
	getResp      *chaossystemmodule.ChaosSystemResp
	metadataResp []chaossystemmodule.SystemMetadataResp
	err          error
}

func (s chaosSystemReaderStub) ListSystems(_ context.Context, req *chaossystemmodule.ListChaosSystemReq) (*dto.ListResp[chaossystemmodule.ChaosSystemResp], error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return s.listResp, s.err
}

func (s chaosSystemReaderStub) GetSystem(_ context.Context, id int) (*chaossystemmodule.ChaosSystemResp, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.getResp, s.err
}

func (s chaosSystemReaderStub) CreateSystem(_ context.Context, req *chaossystemmodule.CreateChaosSystemReq) (*chaossystemmodule.ChaosSystemResp, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	return s.getResp, s.err
}

func (s chaosSystemReaderStub) UpdateSystem(_ context.Context, id int, req *chaossystemmodule.UpdateChaosSystemReq) (*chaossystemmodule.ChaosSystemResp, error) {
	if id <= 0 || req == nil {
		return nil, errors.New("invalid request")
	}
	return s.getResp, s.err
}

func (s chaosSystemReaderStub) DeleteSystem(_ context.Context, id int) error {
	if id <= 0 {
		return errors.New("invalid id")
	}
	return s.err
}

func (s chaosSystemReaderStub) UpsertMetadata(_ context.Context, id int, req *chaossystemmodule.BulkUpsertSystemMetadataReq) error {
	if id <= 0 || req == nil {
		return errors.New("invalid request")
	}
	return s.err
}

func (s chaosSystemReaderStub) ListMetadata(_ context.Context, id int, _ string) ([]chaossystemmodule.SystemMetadataResp, error) {
	if id <= 0 {
		return nil, errors.New("invalid id")
	}
	return s.metadataResp, s.err
}

func TestResourceServerListProjects(t *testing.T) {
	server := &resourceServer{
		projects: projectReaderStub{listResp: &dto.ListResp[projectmodule.ProjectResp]{
			Items: []projectmodule.ProjectResp{{ID: 1, Name: "demo"}},
			Pagination: &dto.PaginationInfo{
				Page: 1, Size: 20, Total: 1, TotalPages: 1,
			},
		}},
		containers:   containerReaderStub{},
		datasets:     datasetReaderStub{},
		labels:       labelReaderStub{},
		chaosSystems: chaosSystemReaderStub{},
		evaluations:  evaluationReaderStub{},
	}

	query, err := structpb.NewStruct(map[string]any{"page": 1, "size": 20})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	resp, err := server.ListProjects(context.Background(), &resourcev1.ListProjectsRequest{Query: query})
	if err != nil {
		t.Fatalf("ListProjects() error = %v", err)
	}
	if got := resp.GetData().AsMap()["items"]; got == nil {
		t.Fatalf("ListProjects() missing items in response: %+v", resp.GetData().AsMap())
	}
}

func TestResourceServerGetDatasetNotFound(t *testing.T) {
	server := &resourceServer{
		projects:     projectReaderStub{},
		containers:   containerReaderStub{},
		datasets:     datasetReaderStub{err: consts.ErrNotFound},
		labels:       labelReaderStub{},
		chaosSystems: chaosSystemReaderStub{},
		evaluations:  evaluationReaderStub{},
	}

	_, err := server.GetDataset(context.Background(), &resourcev1.GetResourceRequest{Id: 8})
	if err == nil {
		t.Fatal("GetDataset() error = nil, want error")
	}
	if status.Code(err) != codes.NotFound {
		t.Fatalf("GetDataset() code = %s, want %s", status.Code(err), codes.NotFound)
	}
}

func TestResourceServerListContainersInvalidQuery(t *testing.T) {
	server := &resourceServer{
		projects:     projectReaderStub{},
		containers:   containerReaderStub{},
		datasets:     datasetReaderStub{},
		labels:       labelReaderStub{},
		chaosSystems: chaosSystemReaderStub{},
		evaluations:  evaluationReaderStub{},
	}

	query, err := structpb.NewStruct(map[string]any{"page": -1})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	_, err = server.ListContainers(context.Background(), &resourcev1.ListContainersRequest{Query: query})
	if err == nil {
		t.Fatal("ListContainers() error = nil, want error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("ListContainers() code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestResourceServerListDatapackEvaluationsRequiresUserID(t *testing.T) {
	server := &resourceServer{
		projects:     projectReaderStub{},
		containers:   containerReaderStub{},
		datasets:     datasetReaderStub{},
		labels:       labelReaderStub{},
		chaosSystems: chaosSystemReaderStub{},
		evaluations:  evaluationReaderStub{},
	}

	query, err := structpb.NewStruct(map[string]any{
		"specs": []any{
			map[string]any{
				"algorithm": map[string]any{"name": "algo", "version": "v1.0.0"},
				"datapack":  "pack-a",
			},
		},
	})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	_, err = server.ListDatapackEvaluationResults(context.Background(), &resourcev1.ListDatapackEvaluationsRequest{Query: query})
	if err == nil {
		t.Fatal("ListDatapackEvaluationResults() error = nil, want error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("ListDatapackEvaluationResults() code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestResourceServerListEvaluations(t *testing.T) {
	server := &resourceServer{
		projects:     projectReaderStub{},
		containers:   containerReaderStub{},
		datasets:     datasetReaderStub{},
		labels:       labelReaderStub{},
		chaosSystems: chaosSystemReaderStub{},
		evaluations: evaluationReaderStub{listResp: &dto.ListResp[evaluationmodule.EvaluationResp]{
			Items: []evaluationmodule.EvaluationResp{{ID: 3, EvalType: consts.EvalTypeDataset}},
			Pagination: &dto.PaginationInfo{
				Page: 1, Size: 20, Total: 1, TotalPages: 1,
			},
		}},
	}

	query, err := structpb.NewStruct(map[string]any{"page": 1, "size": 20})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	resp, err := server.ListEvaluations(context.Background(), &resourcev1.ListEvaluationsRequest{Query: query})
	if err != nil {
		t.Fatalf("ListEvaluations() error = %v", err)
	}
	if got := resp.GetData().AsMap()["items"]; got == nil {
		t.Fatalf("ListEvaluations() missing items in response: %+v", resp.GetData().AsMap())
	}
}

func TestResourceServerListLabels(t *testing.T) {
	server := &resourceServer{
		projects:   projectReaderStub{},
		containers: containerReaderStub{},
		datasets:   datasetReaderStub{},
		labels: labelReaderStub{listResp: &dto.ListResp[labelmodule.LabelResp]{
			Items: []labelmodule.LabelResp{{ID: 9, Key: "env", Value: "prod"}},
			Pagination: &dto.PaginationInfo{
				Page: 1, Size: 20, Total: 1, TotalPages: 1,
			},
		}},
		chaosSystems: chaosSystemReaderStub{},
		evaluations:  evaluationReaderStub{},
	}

	query, err := structpb.NewStruct(map[string]any{"page": 1, "size": 20})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	resp, err := server.ListLabels(context.Background(), &resourcev1.QueryRequest{Query: query})
	if err != nil {
		t.Fatalf("ListLabels() error = %v", err)
	}
	if got := resp.GetData().AsMap()["items"]; got == nil {
		t.Fatalf("ListLabels() missing items in response: %+v", resp.GetData().AsMap())
	}
}

func TestResourceServerBatchDeleteLabelsRequiresIDs(t *testing.T) {
	server := &resourceServer{
		projects:     projectReaderStub{},
		containers:   containerReaderStub{},
		datasets:     datasetReaderStub{},
		labels:       labelReaderStub{},
		chaosSystems: chaosSystemReaderStub{},
		evaluations:  evaluationReaderStub{},
	}

	_, err := server.BatchDeleteLabels(context.Background(), &resourcev1.BatchDeleteRequest{})
	if err == nil {
		t.Fatal("BatchDeleteLabels() error = nil, want error")
	}
	if status.Code(err) != codes.InvalidArgument {
		t.Fatalf("BatchDeleteLabels() code = %s, want %s", status.Code(err), codes.InvalidArgument)
	}
}

func TestResourceServerListChaosSystems(t *testing.T) {
	server := &resourceServer{
		projects:   projectReaderStub{},
		containers: containerReaderStub{},
		datasets:   datasetReaderStub{},
		labels:     labelReaderStub{},
		chaosSystems: chaosSystemReaderStub{listResp: &dto.ListResp[chaossystemmodule.ChaosSystemResp]{
			Items: []chaossystemmodule.ChaosSystemResp{{ID: 4, Name: "k8s"}},
			Pagination: &dto.PaginationInfo{
				Page: 1, Size: 20, Total: 1, TotalPages: 1,
			},
		}},
		evaluations: evaluationReaderStub{},
	}

	query, err := structpb.NewStruct(map[string]any{"page": 1, "size": 20})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	resp, err := server.ListChaosSystems(context.Background(), &resourcev1.QueryRequest{Query: query})
	if err != nil {
		t.Fatalf("ListChaosSystems() error = %v", err)
	}
	if got := resp.GetData().AsMap()["items"]; got == nil {
		t.Fatalf("ListChaosSystems() missing items in response: %+v", resp.GetData().AsMap())
	}
}

func TestResourceServerListChaosSystemMetadata(t *testing.T) {
	server := &resourceServer{
		projects:   projectReaderStub{},
		containers: containerReaderStub{},
		datasets:   datasetReaderStub{},
		labels:     labelReaderStub{},
		chaosSystems: chaosSystemReaderStub{metadataResp: []chaossystemmodule.SystemMetadataResp{
			{ID: 1, SystemName: "k8s", MetadataType: "service", ServiceName: "api"},
		}},
		evaluations: evaluationReaderStub{},
	}

	query, err := structpb.NewStruct(map[string]any{"type": "service"})
	if err != nil {
		t.Fatalf("NewStruct() error = %v", err)
	}

	resp, err := server.ListChaosSystemMetadata(context.Background(), &resourcev1.IDQueryRequest{Id: 4, Query: query})
	if err != nil {
		t.Fatalf("ListChaosSystemMetadata() error = %v", err)
	}
	items, ok := resp.GetData().AsMap()["items"].([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("ListChaosSystemMetadata() unexpected response: %+v", resp.GetData().AsMap())
	}
}
