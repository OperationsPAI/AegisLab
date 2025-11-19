package producer

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"aegis/utils"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// CreateProject handles the business logic for creating a new project
func CreateProject(req *dto.CreateProjectReq, userID int) (*dto.ProjectResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	project := req.ConvertToProject()

	var createdProject *database.Project
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		role, err := repository.GetRoleByName(tx, consts.RoleProjectAdmin)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role %v not found", err, consts.RoleProjectAdmin)
			}
			return fmt.Errorf("failed to get project owner role: %w", err)
		}

		if err := repository.CreateProject(tx, project); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: project with name %s already exists", consts.ErrAlreadyExists, project.Name)
			}
			return err
		}

		if err := repository.CreateUserProject(tx, &database.UserProject{
			UserID:    userID,
			ProjectID: project.ID,
			RoleID:    role.ID,
			Status:    consts.CommonEnabled,
		}); err != nil {
			return fmt.Errorf("failed to assign project owner: %w", err)
		}

		createdProject = project
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewProjectResp(createdProject), nil
}

// DeleteProject deletes an existing project by marking its status as deleted
func DeleteProject(projectID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if _, err := repository.RemoveUsersFromProject(tx, projectID); err != nil {
			return fmt.Errorf("failed to remove users from project: %w", err)
		}

		rows, err := repository.DeleteProject(tx, projectID)
		if err != nil {
			return fmt.Errorf("failed to delete project: %w", err)
		}
		if rows == 0 {
			return fmt.Errorf("%w: project id %d not found", consts.ErrNotFound, projectID)
		}

		return nil
	})
}

// GetProjectDetail retrieves detailed information about a project by its ID
func GetProjectDetail(projectID int) (*dto.ProjectDetailResp, error) {
	project, err := repository.GetProjectByID(database.DB, projectID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: project with ID %d not found", consts.ErrNotFound, projectID)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	resp := dto.NewProjectDetailResp(project)

	userCount, err := repository.GetProjectUserCount(database.DB, project.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get project user count: %w", err)
	}
	resp.UserCount = userCount

	// TODO add more project details if needed (container, dataset, etc.)

	return resp, nil
}

// ListProjects lists projects based on the provided filters
func ListProjects(req *dto.ListProjectReq) (*dto.ListResp[dto.ProjectResp], error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	limit, offset := req.ToGormParams()

	projects, total, err := repository.ListProjects(database.DB, limit, offset, req.IsPublic, req.Status)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	projectIDs := make([]int, 0, len(projects))
	for _, p := range projects {
		projectIDs = append(projectIDs, p.ID)
	}

	labelsMap, err := repository.ListProjectLabels(database.DB, projectIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list project labels: %w", err)
	}

	projectResps := make([]dto.ProjectResp, 0, len(projects))
	for _, project := range projects {
		if labels, exists := labelsMap[project.ID]; exists {
			project.Labels = labels
		}
		projectResps = append(projectResps, *dto.NewProjectResp(&project))
	}

	resp := dto.ListResp[dto.ProjectResp]{
		Items:      projectResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// UpdateProject updates an existing project's details
func UpdateProject(req *dto.UpdateProjectReq, projectID int) (*dto.ProjectResp, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	var updatedProject *database.Project

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		existingProject, err := repository.GetProjectByID(tx, projectID)
		if err != nil {
			return fmt.Errorf("failed to get project: %w", err)
		}

		req.PatchProjectModel(existingProject)

		if err := repository.UpdateProject(tx, existingProject); err != nil {
			return fmt.Errorf("failed to update project: %w", err)
		}

		updatedProject = existingProject
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewProjectResp(updatedProject), nil
}

// ===================== Project-Label =====================

// ManageProjectLabels manages project labels (key-value pairs)
func ManageProjectLabels(req *dto.ManageProjectLabelReq, projectID int) (*dto.ProjectResp, error) {
	if req == nil {
		return nil, fmt.Errorf("manage project labels request is nil")
	}

	var managedProject *database.Project
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		project, err := repository.GetProjectByID(tx, projectID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: project id: %d", consts.ErrNotFound, projectID)
			}
			return fmt.Errorf("failed to get project: %w", err)
		}

		if len(req.AddLabels) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, req.AddLabels, consts.ProjectCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			projectLabels := make([]database.ProjectLabel, 0, len(labels))
			for _, label := range labels {
				projectLabels = append(projectLabels, database.ProjectLabel{
					ProjectID: projectID,
					LabelID:   label.ID,
				})
			}

			if err := repository.AddProjectLabels(tx, projectLabels); err != nil {
				return fmt.Errorf("failed to add project labels: %w", err)
			}
		}

		if len(req.RemoveLabels) > 0 {
			labelIDs, err := repository.ListLabelIDsByKeyAndProjectID(tx, projectID, req.RemoveLabels)
			if err != nil {
				return fmt.Errorf("failed to find label ids by keys: %w", err)
			}

			if len(labelIDs) == 0 {
				return fmt.Errorf("no labels found for the given keys")
			}

			if err := repository.ClearProjectLabels(tx, []int{projectID}, labelIDs); err != nil {
				return fmt.Errorf("failed to clear project labels: %w", err)
			}

			if err := repository.BatchDecreaseLabelUsages(tx, labelIDs, 1); err != nil {
				return fmt.Errorf("failed to decrease label usage counts: %w", err)
			}
		}

		labels, err := repository.ListLabelsByProjectID(database.DB, project.ID)
		if err != nil {
			return fmt.Errorf("failed to get project labels: %w", err)
		}

		project.Labels = labels
		managedProject = project
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewProjectResp(managedProject), nil
}

func fetchProjectsMapByIDBatch(db *gorm.DB, projectIDs []int) (map[int]database.Project, error) {
	if len(projectIDs) == 0 {
		return make(map[int]database.Project), nil
	}

	projects, err := repository.ListProjectsByID(db, utils.ToUniqueSlice(projectIDs))
	if err != nil {
		return nil, fmt.Errorf("failed to list projects by IDs: %w", err)
	}

	projectMap := make(map[int]database.Project, len(projectIDs))
	for _, p := range projects {
		projectMap[p.ID] = p
	}

	return projectMap, nil
}
