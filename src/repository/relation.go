package repository

import (
	"fmt"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
)

const (
	userContainerOmitFields = "active_user_container"
	userProjectOmitFields   = "active_user_project"
)

// CreateUserPermission creates a user-permission record
func CreateUserPermission(db *gorm.DB, userPermission *database.UserPermission) error {
	return createModel(db, userPermission)
}

// DeleteUserPermission deletes a user-permission record
func DeleteUserPermission(db *gorm.DB, userID, permissionID int) error {
	if err := db.Where("user_id = ? AND permission_id = ?", userID, permissionID).
		Delete(&database.UserPermission{}).Error; err != nil {
		return fmt.Errorf("failed to remove permission from user: %v", err)
	}
	return nil
}

// DeleteAllUsersByPermissionID deletes all user-permission associations associated with a given permission
func RemoveAllUsersFromPermission(db *gorm.DB, permissionID int) error {
	if err := db.Where("permission_id = ?", permissionID).
		Delete(&database.UserPermission{}).Error; err != nil {
		return fmt.Errorf("failed to remove all users from permission: %v", err)
	}
	return nil
}

// DeleteAllPermissionsByUserID deletes all user-permission associations associated with a given user
func RemoveAllPermissionsFromUser(db *gorm.DB, userID int) error {
	if err := db.Where("user_id = ?", userID).
		Delete(&database.UserPermission{}).Error; err != nil {
		return fmt.Errorf("failed to remove all permissions from user: %v", err)
	}
	return nil
}

// CreateUserRole creates a user-role association
func CreateUserRole(db *gorm.DB, userRole *database.UserRole) error {
	return createModel(db, userRole)
}

// DeleteUserRole deletes a user-role association
func DeleteUserRole(db *gorm.DB, userID, roleID int) error {
	if err := db.Where("user_id = ? AND role_id = ?", userID, roleID).
		Delete(&database.UserRole{}).Error; err != nil {
		return fmt.Errorf("failed to remove role from user: %v", err)
	}
	return nil
}

// DeleteAllUsersByRoleID deletes all user-role associations associated with a given role
func RemoveAllUsersFromRole(db *gorm.DB, roleID int) error {
	if err := db.Where("role_id = ?", roleID).
		Delete(&database.UserRole{}).Error; err != nil {
		return err
	}
	return nil
}

// DeleteAllRolesByUserID deletes all user-role associations associated with a given user
func RemoveAllRolesFromUser(db *gorm.DB, userID int) error {
	if err := db.Where("user_id = ?", userID).
		Delete(&database.UserRole{}).Error; err != nil {
		return fmt.Errorf("failed to remove all roles from user: %v", err)
	}
	return nil
}

// CreateUserContainer creates a user-container association
func CreateUserContainer(db *gorm.DB, userContainer *database.UserContainer) error {
	return createModel(db.Omit(userContainerOmitFields), userContainer)
}

// UpdateUserContainer updates a user-container association
func UpdateUserContainer(db *gorm.DB, userContainer *database.UserContainer) error {
	return updateModel(db.Omit(userContainerOmitFields), userContainer)
}

// RemoveAllUsersFromContainer removes all user-container associations for a given container
func RemoveAllUsersFromContainer(db *gorm.DB, containerID int) error {
	if err := db.Model(&database.UserContainer{}).
		Where("container_id = ?", containerID).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to remove all users from container: %v", err)
	}
	return nil
}

// RemoveAllContainersFromRole removes all user-container associations for a given role
func RemoveAllContainersFromRole(db *gorm.DB, roleID int) error {
	if err := db.Model(&database.UserContainer{}).
		Where("role_id = ?", roleID).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to remove all containers from role: %v", err)
	}
	return nil
}

// RemoveAllContainersFromUser removes all user-container associations for a given user
func RemoveAllContainersFromUser(db *gorm.DB, userID int) error {
	if err := db.Model(&database.UserContainer{}).
		Where("user_id = ?", userID).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to remove all containers from user: %v", err)
	}
	return nil
}

// CreateUserProject creates a user-project association
func CreateUserProject(db *gorm.DB, userProject *database.UserProject) error {
	return createModel(db.Omit(userProjectOmitFields), userProject)
}

// UpdateUserProject updates a user-project association
func UpdateUserProject(db *gorm.DB, userProject *database.UserProject) error {
	return updateModel(db.Omit(userProjectOmitFields), userProject)
}

// RemoveAllUsersFromProject removes all user-project associations for a given project
func RemoveAllUsersFromProject(db *gorm.DB, projectID int) error {
	if err := db.Model(&database.UserProject{}).
		Where("project_id = ?", projectID).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to remove all users from project: %v", err)
	}
	return nil
}

// RemoveAllProjectsFromRole removes all user-project associations for a given role
func RemoveAllProjectsFromRole(db *gorm.DB, roleID int) error {
	if err := db.Model(&database.UserProject{}).
		Where("role_id = ?", roleID).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to remove all projects from role: %v", err)
	}
	return nil
}

// RemoveAllProjectsFromUser removes all user-project associations for a given user
func RemoveAllProjectsFromUser(db *gorm.DB, userID int) error {
	if err := db.Model(&database.UserProject{}).
		Where("user_id = ?", userID).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to remove all projects from user: %v", err)
	}
	return nil
}

// CreateRolePermission creates a role-permission relation
func CreateRolePermission(db *gorm.DB, rolePermission *database.RolePermission) error {
	return createModel(db, rolePermission)
}

// DeleteRolePermission deletes a role-permission relation
func DeleteRolePermission(db *gorm.DB, roleID int, permissionID int) error {
	if err := db.Where("role_id = ? AND permission_id = ?", roleID, permissionID).
		Delete(&database.RolePermission{}).Error; err != nil {
		return err
	}
	return nil
}

// RemoveAllPermissionsFromRole removes all permissions associated with a given role
func RemoveAllPermissionsFromRole(db *gorm.DB, roleID int) error {
	if err := db.Where("role_id = ?", roleID).
		Delete(&database.RolePermission{}).Error; err != nil {
		return err
	}
	return nil
}

// RemoveAllRolesFromPermission removes all roles associated with a given permission
func RemoveAllRolesFromPermission(db *gorm.DB, permissionID int) error {
	if err := db.Where("permission_id = ?", permissionID).
		Delete(&database.RolePermission{}).Error; err != nil {
		return err
	}
	return nil
}
