package project

import (
	"context"
	"fmt"

	"aegis/dto"
	"aegis/internalclient/orchestratorclient"

	"go.uber.org/fx"
)

type projectStatisticsSource interface {
	ListProjectStatistics(context.Context, []int) (map[int]*dto.ProjectStatistics, error)
}

type projectStatisticsSourceParams struct {
	fx.In

	Repository   *Repository
	Orchestrator *orchestratorclient.Client `optional:"true"`
}

type projectStatisticsAdapter struct {
	orchestrator  *orchestratorclient.Client
	repository    *Repository
	requireRemote bool
}

func newProjectStatisticsSource(params projectStatisticsSourceParams) projectStatisticsSource {
	return projectStatisticsAdapter{
		orchestrator: params.Orchestrator,
		repository:   params.Repository,
	}
}

func newRemoteProjectStatisticsSource(params projectStatisticsSourceParams) projectStatisticsSource {
	return projectStatisticsAdapter{
		orchestrator:  params.Orchestrator,
		repository:    params.Repository,
		requireRemote: true,
	}
}

func (a projectStatisticsAdapter) ListProjectStatistics(ctx context.Context, projectIDs []int) (map[int]*dto.ProjectStatistics, error) {
	if a.orchestrator != nil && a.orchestrator.Enabled() {
		return a.orchestrator.ListProjectStatistics(ctx, projectIDs)
	}
	if a.requireRemote {
		return nil, fmt.Errorf("orchestrator-service project statistics source is not configured")
	}
	if a.repository == nil {
		return nil, fmt.Errorf("project statistics source is not configured")
	}
	return a.repository.ListProjectStatistics(projectIDs)
}

// RemoteStatisticsOption forces the dedicated resource-service path to use orchestrator RPC only.
func RemoteStatisticsOption() fx.Option {
	return fx.Decorate(newRemoteProjectStatisticsSource)
}
