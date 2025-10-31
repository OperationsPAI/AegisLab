package service

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// AssignPermissionToUser assigns a permission directly to a user
func AssignPermissionToUser(req *dto.AssignUserPermissionRequest, userID int) error {
	userPermisson := req.ConvertToUserPermission()

	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		permission, err := repository.GetPermissionByID(tx, req.PermissionID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: permission not found", consts.ErrNotFound)
			}
			return err
		}

		if req.ProjectID != nil {
			project, err := repository.GetProjectByID(tx, *req.ProjectID)
			if err != nil {
				if errors.Is(err, consts.ErrNotFound) {
					return fmt.Errorf("%w: project not found", consts.ErrNotFound)
				}
				return err
			}
			userPermisson.ProjectID = &project.ID
		}

		if req.ContainerID != nil {
			container, err := repository.GetContainerByID(tx, *req.ContainerID)
			if err != nil {
				if errors.Is(err, consts.ErrNotFound) {
					return fmt.Errorf("%w: container not found", consts.ErrNotFound)
				}
				return err
			}
			userPermisson.ContainerID = &container.ID
		}

		userPermisson.UserID = user.ID
		userPermisson.PermissionID = permission.ID

		if err := repository.CreateUserPermission(tx, userPermisson); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: user already has this permission", consts.ErrAlreadyExists)
			}
			return err
		}

		return nil
	})
}

// RemovePermissionFromUser removes a permission directly from a user
func RemovePermissionFromUser(userID, permissionID int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		permission, err := repository.GetPermissionByID(tx, permissionID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: permission not found", consts.ErrNotFound)
			}
			return err
		}

		// Remove permission from user
		if err := repository.DeleteUserPermission(tx, user.ID, permission.ID); err != nil {
			return err
		}
		return nil
	})
}

// AssignRole assigns a role to a user
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
		// Check if user exists
		user, err := repository.GetUserByID(tx, userID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: user not found", consts.ErrNotFound)
			}
			return err
		}

		// Check if role exists
		role, err := repository.GetRoleByID(tx, roleID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: role not found", consts.ErrNotFound)
			}
			return err
		}

		// Remove role from user
		return repository.DeleteUserRole(tx, user.ID, role.ID)
	})
}

// AssignUserToProject assigns a user to a project with a specific role
func AssignUserToProject(userID, projectID, roleID int) error {
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
				return fmt.Errorf("%w: user already has this role", consts.ErrAlreadyExists)
			}
			return err
		}

		return nil
	})
}

// RemoveUserFromProject removes a user from a project
func RemoveUserFromProject(userID, projectID int) error {
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

		if err := repository.UpdateUserProject(tx, &database.UserProject{
			UserID:    user.ID,
			ProjectID: project.ID,
			Status:    consts.CommonDisabled,
		}); err != nil {
			return err
		}
		return nil
	})
}
