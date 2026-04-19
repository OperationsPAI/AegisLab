package team

import (
	"context"
	"fmt"

	"aegis/consts"
	"aegis/dto"
	"aegis/internalclient/resourceclient"
	"aegis/model"
	project "aegis/module/project"

	"go.uber.org/fx"
)

type projectReader interface {
	CountProjects(context.Context, int) (int, error)
	ListProjects(context.Context, *TeamProjectListReq, int) (*dto.ListResp[TeamProjectItem], error)
}

type projectReaderParams struct {
	fx.In

	Repository *Repository
	Resource   *resourceclient.Client `optional:"true"`
}

type projectReaderAdapter struct {
	repo          *Repository
	resource      *resourceclient.Client
	requireRemote bool
}

func newProjectReader(params projectReaderParams) projectReader {
	return projectReaderAdapter{
		repo:     params.Repository,
		resource: params.Resource,
	}
}

func newRemoteProjectReader(params projectReaderParams) projectReader {
	return projectReaderAdapter{
		repo:          params.Repository,
		resource:      params.Resource,
		requireRemote: true,
	}
}

func (r projectReaderAdapter) CountProjects(ctx context.Context, teamID int) (int, error) {
	if r.resource != nil && r.resource.Enabled() {
		includeStatistics := false
		resp, err := r.resource.ListProjects(ctx, &project.ListProjectReq{
			PaginationReq:     dto.PaginationReq{Page: 1, Size: 10},
			TeamID:            &teamID,
			IncludeStatistics: &includeStatistics,
		})
		if err != nil {
			return 0, fmt.Errorf("list team projects via resource-service: %w", err)
		}
		if resp.Pagination == nil {
			return len(resp.Items), nil
		}
		return int(resp.Pagination.Total), nil
	}
	if r.requireRemote {
		return 0, fmt.Errorf("resource-service project reader is not configured")
	}

	var projectCount int64
	if err := r.repo.db.Model(&model.Project{}).
		Where("team_id = ? AND status != ?", teamID, consts.CommonDeleted).
		Count(&projectCount).Error; err != nil {
		return 0, fmt.Errorf("failed to get team project count: %w", err)
	}
	return int(projectCount), nil
}

func (r projectReaderAdapter) ListProjects(ctx context.Context, req *TeamProjectListReq, teamID int) (*dto.ListResp[TeamProjectItem], error) {
	if r.resource != nil && r.resource.Enabled() {
		if req == nil {
			req = &TeamProjectListReq{}
		}

		resourceReq := *req
		resourceReq.TeamID = &teamID
		resp, err := r.resource.ListProjects(ctx, &resourceReq)
		if err != nil {
			return nil, fmt.Errorf("list team projects via resource-service: %w", err)
		}

		items := make([]TeamProjectItem, len(resp.Items))
		copy(items, resp.Items)
		return &dto.ListResp[TeamProjectItem]{
			Items:      items,
			Pagination: resp.Pagination,
		}, nil
	}
	if r.requireRemote {
		return nil, fmt.Errorf("resource-service project reader is not configured")
	}

	if req == nil {
		req = &TeamProjectListReq{}
	}
	limit, offset := req.ToGormParams()
	projects, statsMap, total, err := r.repo.listTeamProjectViews(teamID, limit, offset, req.IsPublic, req.Status)
	if err != nil {
		return nil, err
	}

	items := make([]TeamProjectItem, 0, len(projects))
	for i := range projects {
		items = append(items, *project.NewProjectResp(&projects[i], statsMap[projects[i].ID]))
	}

	return &dto.ListResp[TeamProjectItem]{
		Items:      items,
		Pagination: req.ConvertToPaginationInfo(total),
	}, nil
}

// RemoteProjectReaderOption forces the dedicated iam-service path to use resource RPC only.
func RemoteProjectReaderOption() fx.Option {
	return fx.Decorate(newRemoteProjectReader)
}

var _ projectReader = (*projectReaderAdapter)(nil)
