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
		role, err := repository.GetRoleByName(tx, consts.RoleProjectAdmin.String())
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

	return dto.NewProjectResp(createdProject, nil), nil
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

	// Get project statistics
	statsMap, err := repository.BatchGetProjectStatistics(database.DB, []int{project.ID})
	if err != nil {
		return nil, fmt.Errorf("failed to get project statistics: %w", err)
	}

	stats := statsMap[project.ID]
	resp := dto.NewProjectDetailResp(project, stats)

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
	if req == nil {
		return nil, fmt.Errorf("list project request is nil")
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

	// Batch get statistics for all projects
	statsMap, err := repository.BatchGetProjectStatistics(database.DB, projectIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get project statistics: %w", err)
	}

	projectResps := make([]dto.ProjectResp, 0, len(projects))
	for i := range projects {
		// Convert repository stats to dto stats
		var stats *dto.ProjectStatistics
		if repoStats, exists := statsMap[projects[i].ID]; exists {
			stats = &dto.ProjectStatistics{
				InjectionCount:  repoStats.InjectionCount,
				ExecutionCount:  repoStats.ExecutionCount,
				LastInjectionAt: repoStats.LastInjectionAt,
				LastExecutionAt: repoStats.LastExecutionAt,
			}
		}

		if labels, exists := labelsMap[projects[i].ID]; exists {
			projects[i].Labels = labels
		}
		projectResps = append(projectResps, *dto.NewProjectResp(&projects[i], stats))
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

	return dto.NewProjectResp(updatedProject, nil), nil
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
				if err := repository.ClearProjectLabels(tx, []int{projectID}, labelIDs); err != nil {
					return fmt.Errorf("failed to clear project labels: %w", err)
				}

				if err := repository.BatchDecreaseLabelUsages(tx, labelIDs, 1); err != nil {
					return fmt.Errorf("failed to decrease label usage counts: %w", err)
				}
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

	return dto.NewProjectResp(managedProject, nil), nil
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

// ===================== Project-Injection =====================

// ListProjectInjections lists all fault injections for a specific project
func ListProjectInjections(req *dto.ListInjectionReq, projectID int) (*dto.ListResp[dto.InjectionResp], error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify project exists
	if _, err := repository.GetProjectByID(database.DB, projectID); err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: project id %d not found", consts.ErrNotFound, projectID)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	limit, offset := req.ToGormParams()

	injections, total, err := repository.ListInjectionsByProjectID(database.DB, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list injections for project %d: %w", projectID, err)
	}

	injectionResps := make([]dto.InjectionResp, 0, len(injections))
	for _, injection := range injections {
		injectionResps = append(injectionResps, *dto.NewInjectionResp(&injection))
	}

	resp := dto.ListResp[dto.InjectionResp]{
		Items:      injectionResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// ===================== Project-Execution =====================

// ListProjectExecutions lists all algorithm executions for a specific project
func ListProjectExecutions(req *dto.ListExecutionReq, projectID int) (*dto.ListResp[dto.ExecutionResp], error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Verify project exists
	if _, err := repository.GetProjectByID(database.DB, projectID); err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: project id %d not found", consts.ErrNotFound, projectID)
		}
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	limit, offset := req.ToGormParams()

	executions, total, err := repository.ListExecutionsByProjectID(database.DB, projectID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list executions for project %d: %w", projectID, err)
	}

	executionResps := make([]dto.ExecutionResp, 0, len(executions))
	for _, execution := range executions {
		executionResps = append(executionResps, *dto.NewExecutionResp(&execution, nil))
	}

	resp := dto.ListResp[dto.ExecutionResp]{
		Items:      executionResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// ============================================================================
// Project Permission Check Helper Functions (exported for middleware)
// ============================================================================

// IsUserInProject checks if a user is a member of a project
func IsUserInProject(userID int, projectID int) (bool, error) {
	up, err := repository.GetUserProjectRole(database.DB, userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return up != nil, nil
}

// IsUserProjectAdmin checks if a user has project admin role in a specific project
func IsUserProjectAdmin(userID int, projectID int) (bool, error) {
	up, err := repository.GetUserProjectRole(database.DB, userID, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return up != nil && up.Role != nil && up.Role.Name == consts.RoleProjectAdmin.String(), nil
}

// IsProjectPublic checks if a project is publicly accessible
func IsProjectPublic(projectID int) (bool, error) {
	project, err := repository.GetProjectByID(database.DB, projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return project.IsPublic, nil
}

// GetProjectTeamID gets the team ID for a project
func GetProjectTeamID(projectID int) (int, error) {
	teamID, err := repository.GetProjectTeamID(database.DB, projectID)
	if err != nil {
		return 0, err
	}
	return teamID, nil
}
