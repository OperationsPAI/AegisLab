package grpcorchestrator

import (
	"aegis/dto"
	project "aegis/module/project"
)

type projectStatisticsReader interface {
	ListProjectStatistics([]int) (map[int]*dto.ProjectStatistics, error)
}

type projectRepositoryStatisticsReader struct {
	repo *project.Repository
}

func newProjectStatisticsReader(repo *project.Repository) projectStatisticsReader {
	return &projectRepositoryStatisticsReader{repo: repo}
}

func (r *projectRepositoryStatisticsReader) ListProjectStatistics(projectIDs []int) (map[int]*dto.ProjectStatistics, error) {
	return r.repo.ListProjectStatistics(projectIDs)
}
