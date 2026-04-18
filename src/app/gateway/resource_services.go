package gatewayapp

import (
	"context"

	"aegis/dto"
	"aegis/internalclient/resourceclient"
	chaossystemmodule "aegis/module/chaossystem"
	containermodule "aegis/module/container"
	datasetmodule "aegis/module/dataset"
	evaluationmodule "aegis/module/evaluation"
	labelmodule "aegis/module/label"
	projectmodule "aegis/module/project"
)

type remoteAwareProjectService struct {
	projectmodule.HandlerService
	resource *resourceclient.Client
}

func (s remoteAwareProjectService) GetProjectDetail(ctx context.Context, projectID int) (*projectmodule.ProjectDetailResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.GetProject(ctx, projectID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareProjectService) ListProjects(ctx context.Context, req *projectmodule.ListProjectReq) (*dto.ListResp[projectmodule.ProjectResp], error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListProjects(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

type remoteAwareContainerService struct {
	containermodule.HandlerService
	resource *resourceclient.Client
}

func (s remoteAwareContainerService) GetContainer(ctx context.Context, containerID int) (*containermodule.ContainerDetailResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.GetContainer(ctx, containerID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareContainerService) ListContainers(ctx context.Context, req *containermodule.ListContainerReq) (*dto.ListResp[containermodule.ContainerResp], error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListContainers(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

type remoteAwareDatasetService struct {
	datasetmodule.HandlerService
	resource *resourceclient.Client
}

func (s remoteAwareDatasetService) GetDataset(ctx context.Context, datasetID int) (*datasetmodule.DatasetDetailResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.GetDataset(ctx, datasetID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareDatasetService) ListDatasets(ctx context.Context, req *datasetmodule.ListDatasetReq) (*dto.ListResp[datasetmodule.DatasetResp], error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListDatasets(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

type remoteAwareEvaluationService struct {
	evaluationmodule.HandlerService
	resource *resourceclient.Client
}

func (s remoteAwareEvaluationService) ListDatapackEvaluationResults(ctx context.Context, req *evaluationmodule.BatchEvaluateDatapackReq, userID int) (*evaluationmodule.BatchEvaluateDatapackResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListDatapackEvaluationResults(ctx, req, userID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareEvaluationService) ListDatasetEvaluationResults(ctx context.Context, req *evaluationmodule.BatchEvaluateDatasetReq, userID int) (*evaluationmodule.BatchEvaluateDatasetResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListDatasetEvaluationResults(ctx, req, userID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareEvaluationService) ListEvaluations(ctx context.Context, req *evaluationmodule.ListEvaluationReq) (*dto.ListResp[evaluationmodule.EvaluationResp], error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListEvaluations(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareEvaluationService) GetEvaluation(ctx context.Context, evaluationID int) (*evaluationmodule.EvaluationResp, error) {
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
	labelmodule.HandlerService
	resource labelResourceClient
}

type labelResourceClient interface {
	Enabled() bool
	BatchDeleteLabels(context.Context, []int) error
	CreateLabel(context.Context, *labelmodule.CreateLabelReq) (*labelmodule.LabelResp, error)
	DeleteLabel(context.Context, int) error
	GetLabel(context.Context, int) (*labelmodule.LabelDetailResp, error)
	ListLabels(context.Context, *labelmodule.ListLabelReq) (*dto.ListResp[labelmodule.LabelResp], error)
	UpdateLabel(context.Context, *labelmodule.UpdateLabelReq, int) (*labelmodule.LabelResp, error)
}

func (s remoteAwareLabelService) BatchDelete(ctx context.Context, ids []int) error {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.BatchDeleteLabels(ctx, ids)
	}
	return missingRemoteDependency("resource-service")
}

func (s remoteAwareLabelService) Create(ctx context.Context, req *labelmodule.CreateLabelReq) (*labelmodule.LabelResp, error) {
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

func (s remoteAwareLabelService) GetDetail(ctx context.Context, labelID int) (*labelmodule.LabelDetailResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.GetLabel(ctx, labelID)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareLabelService) List(ctx context.Context, req *labelmodule.ListLabelReq) (*dto.ListResp[labelmodule.LabelResp], error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListLabels(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareLabelService) Update(ctx context.Context, req *labelmodule.UpdateLabelReq, labelID int) (*labelmodule.LabelResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.UpdateLabel(ctx, req, labelID)
	}
	return nil, missingRemoteDependency("resource-service")
}

type remoteAwareChaosSystemService struct {
	chaossystemmodule.HandlerService
	resource chaosSystemResourceClient
}

type chaosSystemResourceClient interface {
	Enabled() bool
	ListChaosSystems(context.Context, *chaossystemmodule.ListChaosSystemReq) (*dto.ListResp[chaossystemmodule.ChaosSystemResp], error)
	GetChaosSystem(context.Context, int) (*chaossystemmodule.ChaosSystemResp, error)
	CreateChaosSystem(context.Context, *chaossystemmodule.CreateChaosSystemReq) (*chaossystemmodule.ChaosSystemResp, error)
	UpdateChaosSystem(context.Context, *chaossystemmodule.UpdateChaosSystemReq, int) (*chaossystemmodule.ChaosSystemResp, error)
	DeleteChaosSystem(context.Context, int) error
	UpsertChaosSystemMetadata(context.Context, int, *chaossystemmodule.BulkUpsertSystemMetadataReq) error
	ListChaosSystemMetadata(context.Context, int, string) ([]chaossystemmodule.SystemMetadataResp, error)
}

func (s remoteAwareChaosSystemService) ListSystems(ctx context.Context, req *chaossystemmodule.ListChaosSystemReq) (*dto.ListResp[chaossystemmodule.ChaosSystemResp], error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListChaosSystems(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareChaosSystemService) GetSystem(ctx context.Context, id int) (*chaossystemmodule.ChaosSystemResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.GetChaosSystem(ctx, id)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareChaosSystemService) CreateSystem(ctx context.Context, req *chaossystemmodule.CreateChaosSystemReq) (*chaossystemmodule.ChaosSystemResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.CreateChaosSystem(ctx, req)
	}
	return nil, missingRemoteDependency("resource-service")
}

func (s remoteAwareChaosSystemService) UpdateSystem(ctx context.Context, id int, req *chaossystemmodule.UpdateChaosSystemReq) (*chaossystemmodule.ChaosSystemResp, error) {
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

func (s remoteAwareChaosSystemService) UpsertMetadata(ctx context.Context, id int, req *chaossystemmodule.BulkUpsertSystemMetadataReq) error {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.UpsertChaosSystemMetadata(ctx, id, req)
	}
	return missingRemoteDependency("resource-service")
}

func (s remoteAwareChaosSystemService) ListMetadata(ctx context.Context, id int, metadataType string) ([]chaossystemmodule.SystemMetadataResp, error) {
	if s.resource != nil && s.resource.Enabled() {
		return s.resource.ListChaosSystemMetadata(ctx, id, metadataType)
	}
	return nil, missingRemoteDependency("resource-service")
}
