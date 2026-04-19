package dataset

import (
	"archive/zip"
	"context"

	"aegis/dto"
	"aegis/utils"
)

// HandlerService captures the dataset operations consumed by the HTTP handler.
type HandlerService interface {
	CreateDataset(context.Context, *CreateDatasetReq, int) (*DatasetResp, error)
	DeleteDataset(context.Context, int) error
	GetDataset(context.Context, int) (*DatasetDetailResp, error)
	ListDatasets(context.Context, *ListDatasetReq) (*dto.ListResp[DatasetResp], error)
	SearchDatasets(context.Context, *SearchDatasetReq) (*dto.ListResp[DatasetDetailResp], error)
	UpdateDataset(context.Context, *UpdateDatasetReq, int) (*DatasetResp, error)
	ManageDatasetLabels(context.Context, *ManageDatasetLabelReq, int) (*DatasetResp, error)
	CreateDatasetVersion(context.Context, *CreateDatasetVersionReq, int, int) (*DatasetVersionResp, error)
	DeleteDatasetVersion(context.Context, int) error
	GetDatasetVersion(context.Context, int, int) (*DatasetVersionDetailResp, error)
	ListDatasetVersions(context.Context, *ListDatasetVersionReq, int) (*dto.ListResp[DatasetVersionResp], error)
	UpdateDatasetVersion(context.Context, *UpdateDatasetVersionReq, int, int) (*DatasetVersionResp, error)
	GetDatasetVersionFilename(context.Context, int, int) (string, error)
	DownloadDatasetVersion(context.Context, *zip.Writer, []utils.ExculdeRule, int) error
	ManageDatasetVersionInjections(context.Context, *ManageDatasetVersionInjectionReq, int) (*DatasetVersionDetailResp, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
