package gateway

import (
	"context"

	"aegis/dto"
	"aegis/internalclient/resourceclient"
	chaossystem "aegis/module/chaossystem"
	container "aegis/module/container"
	dataset "aegis/module/dataset"
	evaluation "aegis/module/evaluation"
	label "aegis/module/label"
	project "aegis/module/project"
)

type remoteAwareProjectService struct {
	project.HandlerService
	resource *resourceclient.Client
}

func (s remoteAwareProjectService) GetProjectDetail(ctx context.Context, projectID int) (*project.ProjectDetailResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.GetProject(ctx, projectID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareProjectService) ListProjects(ctx context.Context, req *project.ListProjectReq) (*dto.ListResp[project.ProjectResp], error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListProjects(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

type remoteAwareContainerService struct {
	container.HandlerService
	resource *resourceclient.Client
}

func (s remoteAwareContainerService) GetContainer(ctx context.Context, containerID int) (*container.ContainerDetailResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.GetContainer(ctx, containerID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareContainerService) ListContainers(ctx context.Context, req *container.ListContainerReq) (*dto.ListResp[container.ContainerResp], error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListContainers(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

type remoteAwareDatasetService struct {
	dataset.HandlerService
	resource *resourceclient.Client
}

func (s remoteAwareDatasetService) GetDataset(ctx context.Context, datasetID int) (*dataset.DatasetDetailResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.GetDataset(ctx, datasetID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareDatasetService) ListDatasets(ctx context.Context, req *dataset.ListDatasetReq) (*dto.ListResp[dataset.DatasetResp], error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListDatasets(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

type remoteAwareEvaluationService struct {
	evaluation.HandlerService
	resource *resourceclient.Client
}

func (s remoteAwareEvaluationService) ListDatapackEvaluationResults(ctx context.Context, req *evaluation.BatchEvaluateDatapackReq, userID int) (*evaluation.BatchEvaluateDatapackResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListDatapackEvaluationResults(ctx, req, userID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareEvaluationService) ListDatasetEvaluationResults(ctx context.Context, req *evaluation.BatchEvaluateDatasetReq, userID int) (*evaluation.BatchEvaluateDatasetResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListDatasetEvaluationResults(ctx, req, userID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareEvaluationService) ListEvaluations(ctx context.Context, req *evaluation.ListEvaluationReq) (*dto.ListResp[evaluation.EvaluationResp], error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListEvaluations(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareEvaluationService) GetEvaluation(ctx context.Context, evaluationID int) (*evaluation.EvaluationResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.GetEvaluation(ctx, evaluationID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareEvaluationService) DeleteEvaluation(ctx context.Context, evaluationID int) error {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.DeleteEvaluation(ctx, evaluationID)
	}
	return missingRemoteDependency("resource-service")
}

type remoteAwareLabelService struct {
	label.HandlerService
	resource labelResourceClient
}

type labelResourceClient interface {
	Enabled() bool
	BatchDeleteLabels(context.Context, []int) error
	CreateLabel(context.Context, *label.CreateLabelReq) (*label.LabelResp, error)
	DeleteLabel(context.Context, int) error
	GetLabel(context.Context, int) (*label.LabelDetailResp, error)
	ListLabels(context.Context, *label.ListLabelReq) (*dto.ListResp[label.LabelResp], error)
	UpdateLabel(context.Context, *label.UpdateLabelReq, int) (*label.LabelResp, error)
}

func (s remoteAwareLabelService) BatchDelete(ctx context.Context, ids []int) error {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.BatchDeleteLabels(ctx, ids)
	}
	return missingRemoteDependency("resource-service")
}

func (s remoteAwareLabelService) Create(ctx context.Context, req *label.CreateLabelReq) (*label.LabelResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.CreateLabel(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareLabelService) Delete(ctx context.Context, labelID int) error {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.DeleteLabel(ctx, labelID)
	}
	return missingRemoteDependency("resource-service")
}

func (s remoteAwareLabelService) GetDetail(ctx context.Context, labelID int) (*label.LabelDetailResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.GetLabel(ctx, labelID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareLabelService) List(ctx context.Context, req *label.ListLabelReq) (*dto.ListResp[label.LabelResp], error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListLabels(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareLabelService) Update(ctx context.Context, req *label.UpdateLabelReq, labelID int) (*label.LabelResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.UpdateLabel(ctx, req, labelID)
	}
	return nil, missingRemoteDependency("resource-service")
}

type remoteAwareChaosSystemService struct {
	chaossystem.HandlerService
	resource chaosSystemResourceClient
}

type chaosSystemResourceClient interface {
	Enabled() bool
	ListChaosSystems(context.Context, *chaossystem.ListChaosSystemReq) (*dto.ListResp[chaossystem.ChaosSystemResp], error)
	GetChaosSystem(context.Context, int) (*chaossystem.ChaosSystemResp, error)
	CreateChaosSystem(context.Context, *chaossystem.CreateChaosSystemReq) (*chaossystem.ChaosSystemResp, error)
	UpdateChaosSystem(context.Context, *chaossystem.UpdateChaosSystemReq, int) (*chaossystem.ChaosSystemResp, error)
	DeleteChaosSystem(context.Context, int) error
	UpsertChaosSystemMetadata(context.Context, int, *chaossystem.BulkUpsertSystemMetadataReq) error
	ListChaosSystemMetadata(context.Context, int, string) ([]chaossystem.SystemMetadataResp, error)
}

func (s remoteAwareChaosSystemService) ListSystems(ctx context.Context, req *chaossystem.ListChaosSystemReq) (*dto.ListResp[chaossystem.ChaosSystemResp], error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListChaosSystems(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareChaosSystemService) GetSystem(ctx context.Context, id int) (*chaossystem.ChaosSystemResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.GetChaosSystem(ctx, id)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareChaosSystemService) CreateSystem(ctx context.Context, req *chaossystem.CreateChaosSystemReq) (*chaossystem.ChaosSystemResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.CreateChaosSystem(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareChaosSystemService) UpdateSystem(ctx context.Context, id int, req *chaossystem.UpdateChaosSystemReq) (*chaossystem.ChaosSystemResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.UpdateChaosSystem(ctx, req, id)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareChaosSystemService) DeleteSystem(ctx context.Context, id int) error {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.DeleteChaosSystem(ctx, id)
	}
	return missingRemoteDependency("resource-service")
}

func (s remoteAwareChaosSystemService) UpsertMetadata(ctx context.Context, id int, req *chaossystem.BulkUpsertSystemMetadataReq) error {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.UpsertChaosSystemMetadata(ctx, id, req)
	}
	return missingRemoteDependency("resource-service")
}

func (s remoteAwareChaosSystemService) ListMetadata(ctx context.Context, id int, metadataType string) ([]chaossystem.SystemMetadataResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListChaosSystemMetadata(ctx, id, metadataType)
	}
	return nil, missingRemoteDependency("resource-service")
}
