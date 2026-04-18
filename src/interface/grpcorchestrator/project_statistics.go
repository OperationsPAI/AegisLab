package grpcorchestratorinterface

import (
	"aegis/dto"
	projectmodule "aegis/module/project"
)

type projectStatisticsReader interface {
	ListProjectStatistics([]int) (map[int]*dto.ProjectStatistics, error)
}

type projectRepositoryStatisticsReader struct {
	repo *projectmodule.Repository
}

func newProjectStatisticsReader(repo *projectmodule.Repository) projectStatisticsReader {
	return &projectRepositoryStatisticsReader{repo: repo}
}

func (r *projectRepositoryStatisticsReader) ListProjectStatistics(projectIDs []int) (map[int]*dto.ProjectStatistics, error) {
	return r.repo.ListProjectStatistics(projectIDs)
}
