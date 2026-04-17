package projectmodule

import (
	"context"
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/service/common"

	"gorm.io/gorm"
)

type Service struct {
	repository *Repository
}

func NewService(repository *Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) CreateProject(ctx context.Context, req *CreateProjectReq, userID int) (*ProjectResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	project := req.ConvertToProject()

	var createdProject *model.Project
	err := s.repository.Transaction(func(tx *gorm.DB) error {
		if err := s.repository.withDB(tx).createProjectWithOwner(project, userID); err != nil {
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
	return s.repository.Transaction(func(tx *gorm.DB) error {
		rows, err := s.repository.withDB(tx).deleteProjectCascade(projectID)
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
	project, stats, userCount, err := s.repository.loadProjectDetail(projectID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: project with ID %d not found", consts.ErrNotFound, projectID)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
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

	projects, statsMap, total, err := s.repository.listProjectViews(limit, offset, req.IsPublic, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
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

	err := s.repository.Transaction(func(tx *gorm.DB) error {
		project, err := s.repository.withDB(tx).updateMutableProject(projectID, func(existingProject *model.Project) {
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
	err := s.repository.Transaction(func(tx *gorm.DB) error {
		repo := s.repository.withDB(tx)
		addLabelIDs := make([]int, 0, len(req.AddLabels))
		if len(req.AddLabels) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, req.AddLabels, consts.ProjectCategory)
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
