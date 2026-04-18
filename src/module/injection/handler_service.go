package injectionmodule

import (
	"archive/zip"
	"context"
	"io"

	"aegis/dto"
	"aegis/utils"
)

// HandlerService captures the injection operations consumed by the HTTP handler.
type HandlerService interface {
	ListProjectInjections(context.Context, *ListInjectionReq, int) (*dto.ListResp[InjectionResp], error)
	Search(context.Context, *SearchInjectionReq, *int) (*dto.SearchResp[InjectionDetailResp], error)
	ListNoIssues(context.Context, *ListInjectionNoIssuesReq, *int) ([]InjectionNoIssuesResp, error)
	ListWithIssues(context.Context, *ListInjectionWithIssuesReq, *int) ([]InjectionWithIssuesResp, error)
	SubmitFaultInjection(context.Context, *SubmitInjectionReq, string, int, *int) (*SubmitInjectionResp, error)
	SubmitDatapackBuilding(context.Context, *SubmitDatapackBuildingReq, string, int, *int) (*SubmitDatapackBuildingResp, error)
	ListInjections(context.Context, *ListInjectionReq) (*dto.ListResp[InjectionResp], error)
	GetInjection(context.Context, int) (*InjectionDetailResp, error)
	ManageLabels(context.Context, *ManageInjectionLabelReq, int) (*InjectionResp, error)
	BatchManageLabels(context.Context, *BatchManageInjectionLabelReq) (*BatchManageInjectionLabelResp, error)
	BatchDelete(context.Context, *BatchDeleteInjectionReq) error
	Clone(context.Context, int, *CloneInjectionReq) (*InjectionDetailResp, error)
	GetLogs(context.Context, int) (*InjectionLogsResp, error)
	GetDatapackFilename(context.Context, int) (string, error)
	DownloadDatapack(context.Context, *zip.Writer, []utils.ExculdeRule, int) error
	GetDatapackFiles(context.Context, int, string) (*DatapackFilesResp, error)
	DownloadDatapackFile(context.Context, int, string) (string, string, int64, io.ReadSeekCloser, error)
	QueryDatapackFile(context.Context, int, string) (string, int64, io.ReadCloser, error)
	UpdateGroundtruth(context.Context, int, *UpdateGroundtruthReq) error
	UploadDatapack(context.Context, *UploadDatapackReq, io.Reader, int64) (*UploadDatapackResp, error)
}

func AsHandlerService(service *Service) HandlerService {
	return service
}
