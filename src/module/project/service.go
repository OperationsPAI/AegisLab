package projectmodule

import (
	"context"
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	labelmodule "aegis/module/label"

	"gorm.io/gorm"
)

type Service struct {
	repository *Repository
	stats      projectStatisticsSource
}

func NewService(repository *Repository, stats projectStatisticsSource) *Service {
	return &Service{
		repository: repository,
		stats:      stats,
	}
}

func (s *Service) CreateProject(ctx context.Context, req *CreateProjectReq, userID int) (*ProjectResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	project := req.ConvertToProject()

	var createdProject *model.Project
	err := s.repository.db.Transaction(func(tx *gorm.DB) error {
		if err := NewRepository(tx).createProjectWithOwner(project, userID); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: project with name %s already exists", consts.ErrAlreadyExists, project.Name)
			}
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role %v not found", err, consts.RoleProjectAdmin)
			}
			return err
		}
		createdProject = project
		return nil
	})
	if err != nil {
		return nil, err
	}

	return NewProjectResp(createdProject, nil), nil
}

func (s *Service) DeleteProject(ctx context.Context, projectID int) error {
	return s.repository.db.Transaction(func(tx *gorm.DB) error {
		rows, err := NewRepository(tx).deleteProjectCascade(projectID)
		if err != nil {
			return err
		}
		if rows == 0 {
			return fmt.Errorf("%w: project id %d not found", consts.ErrNotFound, projectID)
		}

		return nil
	})
}

func (s *Service) GetProjectDetail(ctx context.Context, projectID int) (*ProjectDetailResp, error) {
	project, userCount, err := s.repository.loadProjectDetailBase(projectID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: project with ID %d not found", consts.ErrNotFound, projectID)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}
	statsMap, err := s.stats.ListProjectStatistics(ctx, []int{project.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to get project statistics: %w", err)
	}
	stats := statsMap[project.ID]
	if stats == nil {
		stats = &dto.ProjectStatistics{}
	}
	resp := NewProjectDetailResp(project, stats)
	resp.UserCount = userCount

	return resp, nil
}

func (s *Service) ListProjects(ctx context.Context, req *ListProjectReq) (*dto.ListResp[ProjectResp], error) {
	if req == nil {
		return nil, fmt.Errorf("list project request is nil")
	}

	limit, offset := req.ToGormParams()
	includeStatistics := req.IncludeStatistics == nil || *req.IncludeStatistics

	projects, total, err := s.repository.listProjectViews(limit, offset, req.IsPublic, req.Status, req.TeamID)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	statsMap := make(map[int]*dto.ProjectStatistics, len(projects))
	for i := range projects {
		statsMap[projects[i].ID] = &dto.ProjectStatistics{}
	}
	if includeStatistics && len(projects) > 0 {
		projectIDs := make([]int, 0, len(projects))
		for i := range projects {
			projectIDs = append(projectIDs, projects[i].ID)
		}
		statsMap, err = s.stats.ListProjectStatistics(ctx, projectIDs)
		if err != nil {
			return nil, fmt.Errorf("failed to list project statistics: %w", err)
		}
		for _, projectID := range projectIDs {
			if statsMap[projectID] == nil {
				statsMap[projectID] = &dto.ProjectStatistics{}
			}
		}
	}

	projectResps := make([]ProjectResp, 0, len(projects))
	for i := range projects {
		var stats *dto.ProjectStatistics
		if repoStats, exists := statsMap[projects[i].ID]; exists {
			stats = &dto.ProjectStatistics{
				InjectionCount:  repoStats.InjectionCount,
				ExecutionCount:  repoStats.ExecutionCount,
				LastInjectionAt: repoStats.LastInjectionAt,
				LastExecutionAt: repoStats.LastExecutionAt,
			}
		}

		projectResps = append(projectResps, *NewProjectResp(&projects[i], stats))
	}

	resp := dto.ListResp[ProjectResp]{
		Items:      projectResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

func (s *Service) UpdateProject(ctx context.Context, req *UpdateProjectReq, projectID int) (*ProjectResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var updatedProject *model.Project

	err := s.repository.db.Transaction(func(tx *gorm.DB) error {
		project, err := NewRepository(tx).updateMutableProject(projectID, func(existingProject *model.Project) {
			req.PatchProjectModel(existingProject)
		})
		if err != nil {
			return fmt.Errorf("failed to get project: %w", err)
		}
		updatedProject = project
		return nil
	})
	if err != nil {
		return nil, err
	}

	return NewProjectResp(updatedProject, nil), nil
}

func (s *Service) ManageProjectLabels(ctx context.Context, req *ManageProjectLabelReq, projectID int) (*ProjectResp, error) {
	if req == nil {
		return nil, fmt.Errorf("manage project labels request is nil")
	}

	var managedProject *model.Project
	err := s.repository.db.Transaction(func(tx *gorm.DB) error {
		repo := NewRepository(tx)
		addLabelIDs := make([]int, 0, len(req.AddLabels))
		if len(req.AddLabels) > 0 {
			labels, err := labelmodule.NewRepository(tx).CreateOrUpdateLabelsFromItems(tx, req.AddLabels, consts.ProjectCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			for _, label := range labels {
				addLabelIDs = append(addLabelIDs, label.ID)
			}
		}

		project, err := repo.manageProjectLabels(projectID, addLabelIDs, req.RemoveLabels)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) || errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: project id: %d", consts.ErrNotFound, projectID)
			}
			return fmt.Errorf("failed to manage project labels: %w", err)
		}
		managedProject = project
		return nil
	})
	if err != nil {
		return nil, err
	}

	return NewProjectResp(managedProject, nil), nil
}
