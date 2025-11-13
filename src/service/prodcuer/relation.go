package producer

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// ===================== User-Role =====================

// AssignRoleToUser assigns a role to a user
func AssignRoleToUser(userID, roleID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		role, err := repository.GetRoleByID(tx, roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return err
		}

		// Assign role to user
		if err := repository.CreateUserRole(tx, &database.UserRole{
			UserID: user.ID,
			RoleID: role.ID,
		}); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: user already has this role", consts.ErrAlreadyExists)
			}
			return err
		}

		return nil
	})
}

// RemoveRoleFromUser removes a role from a user
func RemoveRoleFromUser(userID, roleID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		role, err := repository.GetRoleByID(tx, roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return err
		}

		if err := repository.DeleteUserRole(tx, user.ID, role.ID); err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("%w: failed to delete user role association (%d, %d)", err, userID, roleID)
			}
			return err
		}
		return nil
	})
}

// ===================== User-Permission =====================

// BatchAssignUserPermissions assigns multiple permissions to a user
func BatchAssignUserPermissions(req *dto.AssignUserPermissionReq, userID int) error {
	permissionIDs := make([]int, len(req.Items))
	for i, up := range req.Items {
		permissionIDs[i] = up.PermissionID
	}

	containerIDs := make([]int, 0, len(req.Items))
	datasetIDs := make([]int, 0, len(req.Items))
	projectIDs := make([]int, 0, len(req.Items))
	for _, item := range req.Items {
		if item.ContainerID != nil {
			containerIDs = append(containerIDs, *item.ContainerID)
		}
		if item.DatasetID != nil {
			datasetIDs = append(datasetIDs, *item.DatasetID)
		}
		if item.ProjectID != nil {
			projectIDs = append(projectIDs, *item.ProjectID)
		}
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		permissionResults, err := fetchPermissionsMapByIDBatch(tx, permissionIDs)
		if err != nil {
			return fmt.Errorf("failed to fetch permissions: %w", err)
		}

		containerResults, err := fetchContainersMapByIDBatch(tx, containerIDs)
		if err != nil {
			return fmt.Errorf("failed to fetch containers: %w", err)
		}

		datasetResults, err := fetchDatasetsMapByIDBatch(tx, datasetIDs)
		if err != nil {
			return fmt.Errorf("failed to fetch datasets: %w", err)
		}

		projectResults, err := fetchProjectsMapByIDBatch(tx, projectIDs)
		if err != nil {
			return fmt.Errorf("failed to fetch projects: %w", err)
		}

		var userPermissons []database.UserPermission
		for _, item := range req.Items {
			if _, exists := permissionResults[item.PermissionID]; !exists {
				return fmt.Errorf("%w: permission id %d not found", consts.ErrNotFound, item.PermissionID)
			}

			if item.ContainerID != nil {
				if _, exists := containerResults[*item.ContainerID]; !exists {
					return fmt.Errorf("%w: container id %d not found", consts.ErrNotFound, *item.ContainerID)
				}
			}

			if item.DatasetID != nil {
				if _, exists := datasetResults[*item.DatasetID]; !exists {
					return fmt.Errorf("%w: dataset id %d not found", consts.ErrNotFound, *item.DatasetID)
				}
			}

			if item.ProjectID != nil {
				if _, exists := projectResults[*item.ProjectID]; !exists {
					return fmt.Errorf("%w: project id %d not found", consts.ErrNotFound, *item.ProjectID)
				}
			}

			userPermisson := item.ConvertToUserPermission()
			userPermisson.UserID = user.ID
			userPermissons = append(userPermissons, *userPermisson)
		}

		if err := repository.BatchCreateUserPermissions(tx, userPermissons); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: user already has one or more of these permissions", consts.ErrAlreadyExists)
			}
			return fmt.Errorf("failed to assgin permissions to user: %w", err)
		}

		return nil
	})
}

// BatchRemoveUserPermissions removes multiple permissions from a user
func BatchRemoveUserPermissions(req *dto.RemoveUserPermissionReq, userID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		permissionResults, err := fetchPermissionsMapByIDBatch(tx, req.PermissionIDs)
		if err != nil {
			return fmt.Errorf("failed to fetch permissions: %w", err)
		}

		for _, permissionID := range req.PermissionIDs {
			if _, exists := permissionResults[permissionID]; !exists {
				return fmt.Errorf("%w: permission id %d not found", consts.ErrNotFound, permissionID)
			}
		}

		if err := repository.BatchDeleteUserPermisssions(tx, user.ID, req.PermissionIDs); err != nil {
			return fmt.Errorf("")
		}
		return nil
	})
}

// ===================== Role-Permission =====================

// AssginPermissionsToRole assigns multiple permissions to a role
func BatchAssignRolePermissions(permissionIDs []int, roleID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		role, err := repository.GetRoleByID(tx, roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return err
		}

		if role.IsSystem {
			return fmt.Errorf("%w: cannot assign permissions to system role", consts.ErrPermissionDenied)
		}

		permissionResults, err := fetchPermissionsMapByIDBatch(tx, permissionIDs)
		if err != nil {
			return fmt.Errorf("failed to fetch permissions: %w", err)
		}

		var rolePermissions []database.RolePermission
		for _, permissionID := range permissionIDs {
			if _, exists := permissionResults[permissionID]; !exists {
				return fmt.Errorf("%w: permission id %d not found", consts.ErrNotFound, permissionID)
			}

			rolePermissions = append(rolePermissions, database.RolePermission{
				RoleID:       role.ID,
				PermissionID: permissionID,
			})
		}

		if err := repository.BatchCreateRolePermissions(tx, rolePermissions); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: role already has one or more of these permissions", consts.ErrAlreadyExists)
			}
			return fmt.Errorf("failed to assign permissions to role: %w", err)
		}

		return nil
	})
}

// RemovePermissionsFromRole removes permissions from a role
func RemovePermissionsFromRole(permissionIDs []int, roleID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		role, err := repository.GetRoleByID(tx, roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return err
		}

		if role.IsSystem {
			return fmt.Errorf("%w: cannot remove permissions of system role", consts.ErrPermissionDenied)
		}

		if err := repository.BatchDeleteRolePermisssions(tx, roleID, permissionIDs); err != nil {
			return fmt.Errorf("")
		}

		return nil
	})
}

// ListUsersFromRole lists users assigned to a specific role
func ListUsersFromRole(roleID int) ([]dto.UserResp, error) {
	role, err := repository.GetRoleByID(database.DB, roleID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: role not found", consts.ErrNotFound)
		}
		return nil, err
	}

	users, err := repository.ListUsersByRoleID(database.DB, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role users: %w", err)
	}

	var userResps []dto.UserResp
	for _, user := range users {
		userResps = append(userResps, *dto.NewUserResp(&user))
	}

	return userResps, nil
}

// ===================== User-Container =====================

// AssignContainerToUser assigns a user to a container with a specific role
func AssignContainerToUser(userID, containerID, roleID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		container, err := repository.GetContainerByID(tx, containerID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: container not found", consts.ErrNotFound)
			}
			return err
		}

		role, err := repository.GetRoleByID(tx, roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return err
		}

		if err := repository.CreateUserContainer(tx, &database.UserContainer{
			UserID:      user.ID,
			ContainerID: container.ID,
			RoleID:      role.ID,
		}); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: user already assigned to this container", consts.ErrAlreadyExists)
			}
			return err
		}

		return nil
	})
}

// RemoveContainerFromUser removes a user from a container
func RemoveContainerFromUser(userID, containerID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		container, err := repository.GetContainerByID(tx, containerID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: container not found", consts.ErrNotFound)
			}
			return err
		}

		row, err := repository.DeleteUserContainer(tx, user.ID, container.ID)
		if err != nil {
			return fmt.Errorf("failed to remove user from container: %w", err)
		}
		if row == 0 {
			return fmt.Errorf("%w: user is not assigned to this container", consts.ErrNotFound)
		}

		return nil
	})
}

// ===================== User-Dataset =====================

// AssignDatasetToUser assigns a user to a dataset with a specific role
func AssignDatasetToUser(userID, datasetID, roleID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		dataset, err := repository.GetDatasetByID(tx, datasetID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: dataset not found", consts.ErrNotFound)
			}
			return err
		}

		role, err := repository.GetRoleByID(tx, roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return err
		}

		if err := repository.CreateUserDataset(tx, &database.UserDataset{
			UserID:    user.ID,
			DatasetID: dataset.ID,
			RoleID:    role.ID,
		}); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: user already assigned to this dataset", consts.ErrAlreadyExists)
			}
			return err
		}

		return nil
	})
}

// RemoveDatasetFromUser removes a user from a dataset
func RemoveDatasetFromUser(userID, datasetID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		dataset, err := repository.GetDatasetByID(tx, datasetID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: dataset not found", consts.ErrNotFound)
			}
			return err
		}

		row, err := repository.DeleteUserDataset(tx, user.ID, dataset.ID)
		if err != nil {
			return fmt.Errorf("failed to remove user from dataset: %w", err)
		}
		if row == 0 {
			return fmt.Errorf("%w: user is not assigned to this dataset", consts.ErrNotFound)
		}

		return nil
	})
}

// ===================== User-Project =====================

// AssignProjectToUser assigns a user to a project with a specific role
func AssignProjectToUser(userID, projectID, roleID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		project, err := repository.GetProjectByID(tx, projectID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: project not found", consts.ErrNotFound)
			}
			return err
		}

		role, err := repository.GetRoleByID(tx, roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return err
		}

		if err := repository.CreateUserProject(tx, &database.UserProject{
			UserID:    user.ID,
			ProjectID: project.ID,
			RoleID:    role.ID,
		}); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: user already assigned to this project", consts.ErrAlreadyExists)
			}
			return err
		}

		return nil
	})
}

// RemoveProjectFromUser removes a user from a project
func RemoveProjectFromUser(userID, projectID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		project, err := repository.GetProjectByID(tx, projectID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: project not found", consts.ErrNotFound)
			}
			return err
		}

		row, err := repository.DeleteUserProject(tx, project.ID, user.ID)
		if err != nil {
			return fmt.Errorf("failed to remove user from project: %w", err)
		}
		if row == 0 {
			return fmt.Errorf("%w: user is not assigned to this project", consts.ErrNotFound)
		}

		return nil
	})
}
