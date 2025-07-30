package repository

import (
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"gorm.io/gorm"
)

// PermissionChecker struct for permission checking
type PermissionChecker struct {
	UserID    int
	ProjectID *int
}

// NewPermissionChecker creates a new PermissionChecker
func NewPermissionChecker(userID int, projectID *int) *PermissionChecker {
	return &PermissionChecker{
		UserID:    userID,
		ProjectID: projectID,
	}
}

// HasPermissionTyped checks if the user has a specific permission (using ActionName and ResourceName types)
func (pc *PermissionChecker) HasPermissionTyped(action consts.ActionName, resource consts.ResourceName) (bool, error) {
	return CheckUserPermission(pc.UserID, action.String(), resource.String(), pc.ProjectID)
}

// CanReadResource checks if the user has read permission (using ResourceName type)
func (pc *PermissionChecker) CanReadResource(resource consts.ResourceName) (bool, error) {
	return pc.HasPermissionTyped(consts.ActionRead, resource)
}

// CanWriteResource checks if the user has write permission (using ResourceName type)
func (pc *PermissionChecker) CanWriteResource(resource consts.ResourceName) (bool, error) {
	return pc.HasPermissionTyped(consts.ActionWrite, resource)
}

// CanDeleteResource checks if the user has delete permission (using ResourceName type)
func (pc *PermissionChecker) CanDeleteResource(resource consts.ResourceName) (bool, error) {
	return pc.HasPermissionTyped(consts.ActionDelete, resource)
}

// CanExecuteResource checks if the user has execute permission (using ResourceName type)
func (pc *PermissionChecker) CanExecuteResource(resource consts.ResourceName) (bool, error) {
	return pc.HasPermissionTyped(consts.ActionExecute, resource)
}

// IsAdmin checks if the user is an admin
func (pc *PermissionChecker) IsAdmin() (bool, error) {
	roles, err := GetUserRoles(pc.UserID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name == "admin" || role.Name == "super_admin" {
			return true, nil
		}
	}

	return false, nil
}

// IsProjectAdmin checks if the user is a project admin
func (pc *PermissionChecker) IsProjectAdmin() (bool, error) {
	if pc.ProjectID == nil {
		return false, nil
	}

	roles, err := GetUserProjectRoles(pc.UserID, *pc.ProjectID)
	if err != nil {
		return false, err
	}

	for _, role := range roles {
		if role.Name == "project_admin" || role.Name == "admin" {
			return true, nil
		}
	}

	return false, nil
}

// AuthResult represents the result of a permission check
type AuthResult struct {
	Allowed bool
	Reason  string
}

// CheckMultiplePermissions checks multiple permissions in batch
func (pc *PermissionChecker) CheckMultiplePermissions(permissions map[string]string) (map[string]AuthResult, error) {
	results := make(map[string]AuthResult)

	for key, permission := range permissions {
		// permission format: "action:resource"
		// e.g.: "read:dataset", "write:project"
		parts := splitPermission(permission)
		if len(parts) != 2 {
			results[key] = AuthResult{Allowed: false, Reason: "invalid permission format"}
			continue
		}

		action, resource := parts[0], parts[1]
		allowed, err := pc.HasPermissionTyped(consts.ActionName(action), consts.ResourceName(resource))
		if err != nil {
			results[key] = AuthResult{Allowed: false, Reason: fmt.Sprintf("error: %v", err)}
		} else {
			reason := "allowed"
			if !allowed {
				reason = "permission denied"
			}
			results[key] = AuthResult{Allowed: allowed, Reason: reason}
		}
	}

	return results, nil
}

// InitializeSystemData initializes system data (roles, permissions, resources)
func InitializeSystemData() error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		             // Create system resources
		systemResources := []database.Resource{
			{Name: consts.ResourceProject.String(), DisplayName: "Project", Type: "table", Category: "core", IsSystem: true, Status: 1},
			{Name: consts.ResourceDataset.String(), DisplayName: "Dataset", Type: "table", Category: "core", IsSystem: true, Status: 1},
			{Name: consts.ResourceFaultInjection.String(), DisplayName: "Fault Injection", Type: "table", Category: "core", IsSystem: true, Status: 1},
			{Name: consts.ResourceContainer.String(), DisplayName: "Container", Type: "table", Category: "core", IsSystem: true, Status: 1},
			{Name: consts.ResourceTask.String(), DisplayName: "Task", Type: "table", Category: "core", IsSystem: true, Status: 1},
			{Name: consts.ResourceUser.String(), DisplayName: "User", Type: "table", Category: "admin", IsSystem: true, Status: 1},
			{Name: consts.ResourceRole.String(), DisplayName: "Role", Type: "table", Category: "admin", IsSystem: true, Status: 1},
			{Name: consts.ResourcePermission.String(), DisplayName: "Permission", Type: "table", Category: "admin", IsSystem: true, Status: 1},
		}

		for _, resource := range systemResources {
			var existingResource database.Resource
			if err := tx.Where("name = ?", resource.Name).FirstOrCreate(&existingResource, resource).Error; err != nil {
				return fmt.Errorf("failed to create system resource %s: %v", resource.Name, err)
			}
		}

		             // Create system permissions
		actions := []consts.ActionName{consts.ActionRead, consts.ActionWrite, consts.ActionDelete, consts.ActionExecute, consts.ActionManage}
		for _, resource := range systemResources {
			var resourceRecord database.Resource
			if err := tx.Where("name = ?", resource.Name).First(&resourceRecord).Error; err != nil {
				continue
			}

			for _, action := range actions {
				permission := database.Permission{
					Name:        fmt.Sprintf("%s_%s", action.String(), resource.Name),
					DisplayName: fmt.Sprintf("%s %s", actionDisplayName(action.String()), resource.DisplayName),
					Action:      action.String(),
					ResourceID:  resourceRecord.ID,
					IsSystem:    true,
					Status:      1,
				}

				var existingPermission database.Permission
				if err := tx.Where("name = ?", permission.Name).FirstOrCreate(&existingPermission, permission).Error; err != nil {
					return fmt.Errorf("failed to create system permission %s: %v", permission.Name, err)
				}
			}
		}

		             // Create system roles
		systemRoles := []database.Role{
			{Name: "super_admin", DisplayName: "Super Admin", Type: "system", IsSystem: true, Status: 1},
			{Name: "admin", DisplayName: "Admin", Type: "system", IsSystem: true, Status: 1},
			{Name: "project_admin", DisplayName: "Project Admin", Type: "system", IsSystem: true, Status: 1},
			{Name: "developer", DisplayName: "Developer", Type: "system", IsSystem: true, Status: 1},
			{Name: "viewer", DisplayName: "Viewer", Type: "system", IsSystem: true, Status: 1},
		}

		for _, role := range systemRoles {
			var existingRole database.Role
			if err := tx.Where("name = ?", role.Name).FirstOrCreate(&existingRole, role).Error; err != nil {
				return fmt.Errorf("failed to create system role %s: %v", role.Name, err)
			}
		}

		             // Assign permissions to system roles
		if err := assignSystemRolePermissions(tx); err != nil {
			return fmt.Errorf("failed to assign system role permissions: %v", err)
		}

		             // Create super admin user and default project
		_, err := initializeAdminUserAndProjects(tx)
		if err != nil {
			return fmt.Errorf("failed to initialize admin user and projects: %v", err)
		}

		// if err := initializeContainers(tx, adminUser.ID); err != nil {
		// 	return fmt.Errorf("failed to initialize containers: %v", err)
		// }

		return nil
	})
}

// assignSystemRolePermissions assigns permissions to system roles
func assignSystemRolePermissions(tx *gorm.DB) error {
	// super_admin: all permissions
	if err := assignAllPermissionsToRole(tx, consts.RoleSuperAdmin); err != nil {
		return err
	}

	// admin: all except user management
	adminPermissions := []string{
		string(consts.PermissionReadProject), string(consts.PermissionWriteProject), string(consts.PermissionDeleteProject), string(consts.PermissionManageProject),
		string(consts.PermissionReadDataset), string(consts.PermissionWriteDataset), string(consts.PermissionDeleteDataset), string(consts.PermissionManageDataset),
		string(consts.PermissionReadFaultInjection), string(consts.PermissionWriteFaultInjection), string(consts.PermissionDeleteFaultInjection), string(consts.PermissionExecuteFaultInjection),
		string(consts.PermissionReadContainer), string(consts.PermissionWriteContainer), string(consts.PermissionDeleteContainer), string(consts.PermissionManageContainer),
		string(consts.PermissionReadTask), string(consts.PermissionWriteTask), string(consts.PermissionDeleteTask), string(consts.PermissionExecuteTask),
		string(consts.PermissionReadRole), string(consts.PermissionReadPermission),
	}
	if err := assignPermissionsToRole(tx, consts.RoleAdmin, adminPermissions); err != nil {
		return err
	}

	// project_admin: project-related permissions
	projectAdminPermissions := []string{
		string(consts.PermissionReadProject), string(consts.PermissionWriteProject), string(consts.PermissionManageProject),
		string(consts.PermissionReadDataset), string(consts.PermissionWriteDataset), string(consts.PermissionDeleteDataset),
		string(consts.PermissionReadFaultInjection), string(consts.PermissionWriteFaultInjection), string(consts.PermissionDeleteFaultInjection), string(consts.PermissionExecuteFaultInjection),
		string(consts.PermissionReadContainer), string(consts.PermissionWriteContainer),
		string(consts.PermissionReadTask), string(consts.PermissionWriteTask), string(consts.PermissionExecuteTask),
	}
	if err := assignPermissionsToRole(tx, consts.RoleProjectAdmin, projectAdminPermissions); err != nil {
		return err
	}

	// developer: developer permissions
	developerPermissions := []string{
		string(consts.PermissionReadProject), string(consts.PermissionReadDataset), string(consts.PermissionWriteDataset),
		string(consts.PermissionReadFaultInjection), string(consts.PermissionWriteFaultInjection), string(consts.PermissionExecuteFaultInjection),
		string(consts.PermissionReadContainer), string(consts.PermissionReadTask), string(consts.PermissionWriteTask), string(consts.PermissionExecuteTask),
	}
	if err := assignPermissionsToRole(tx, consts.RoleDeveloper, developerPermissions); err != nil {
		return err
	}

	// viewer: read-only permissions
	viewerPermissions := []string{
		string(consts.PermissionReadProject), string(consts.PermissionReadDataset), string(consts.PermissionReadFaultInjection), string(consts.PermissionReadContainer), string(consts.PermissionReadTask),
	}
	if err := assignPermissionsToRole(tx, consts.RoleViewer, viewerPermissions); err != nil {
		return err
	}

	return nil
}

// Helper functions
func splitPermission(permission string) []string {
	for i, char := range permission {
		if char == ':' {
			return []string{permission[:i], permission[i+1:]}
		}
	}
	return []string{permission}
}

func actionDisplayName(action string) string {
	switch action {
	case "read":
		return "View"
	case "write":
		return "Edit"
	case "delete":
		return "Delete"
	case "execute":
		return "Execute"
	case "manage":
		return "Manage"
	default:
		return action
	}
}

// assignAllPermissionsToRole assigns all permissions to a role
func assignAllPermissionsToRole(tx *gorm.DB, roleName consts.RoleName) error {
	var role database.Role
	if err := tx.Where("name = ?", string(roleName)).First(&role).Error; err != nil {
		return err
	}

	var permissions []database.Permission
	if err := tx.Where("is_system = true AND status = 1").Find(&permissions).Error; err != nil {
		return err
	}

	for _, permission := range permissions {
		rolePermission := database.RolePermission{
			RoleID:       role.ID,
			PermissionID: permission.ID,
		}
		if err := tx.FirstOrCreate(&rolePermission, rolePermission).Error; err != nil {
			return err
		}
	}

	return nil
}

// assignPermissionsToRole assigns specific permissions to a role
func assignPermissionsToRole(tx *gorm.DB, roleName consts.RoleName, permissionNames []string) error {
	var role database.Role
	if err := tx.Where("name = ?", string(roleName)).First(&role).Error; err != nil {
		return err
	}

	for _, permName := range permissionNames {
		var permission database.Permission
		if err := tx.Where("name = ?", permName).First(&permission).Error; err != nil {
			continue // skip non-existent permissions
		}

		rolePermission := database.RolePermission{
			RoleID:       role.ID,
			PermissionID: permission.ID,
		}
		if err := tx.FirstOrCreate(&rolePermission, rolePermission).Error; err != nil {
			return err
		}
	}

	return nil
}

// initializeAdminUserAndProjects initializes the super admin user and default projects
func initializeAdminUserAndProjects(tx *gorm.DB) (*database.User, error) {
	// 1. Create super admin user
	adminUser := database.User{
		Username: "admin",
		Email:    "admin@rcabench.local",
		// password: admin123, encrypted with project standard SHA256+salt
		Password: "60c873a916c7659b9798e17015e9130c0cb9c9f4f7f7c022222c0b869243fd6b:98a126542e7a0e2bf0322965b28885e8eb628c605f3cd0228b74e3d36e5edeee",
		FullName: "System Admin",
		Status:   1,
		IsActive: true,
	}

	var existingUser database.User
	if err := tx.Where("username = ?", adminUser.Username).FirstOrCreate(&existingUser, adminUser).Error; err != nil {
		return nil, fmt.Errorf("failed to create admin user: %v", err)
	}

	// 2. Assign super admin role to the super admin user
	var superAdminRole database.Role
	if err := tx.Where("name = ?", "super_admin").First(&superAdminRole).Error; err != nil {
		return nil, fmt.Errorf("failed to find super_admin role: %v", err)
	}

	userRole := database.UserRole{
		UserID: existingUser.ID,
		RoleID: superAdminRole.ID,
	}
	if err := tx.Where("user_id = ? AND role_id = ?", existingUser.ID, superAdminRole.ID).FirstOrCreate(&userRole).Error; err != nil {
		return nil, fmt.Errorf("failed to assign super_admin role to admin user: %v", err)
	}

	// 3. Create default projects
	defaultProjects := []database.Project{
		{
			Name:        "pair_diagnosis",
			Description: "pair_diagnosis",
			Status:      1,
		},
	}

	for _, project := range defaultProjects {
		var existingProject database.Project
		if err := tx.Where("name = ?", project.Name).FirstOrCreate(&existingProject, project).Error; err != nil {
			return nil, fmt.Errorf("failed to create project %s: %v", project.Name, err)
		}

		// Add super admin to the project
		userProject := database.UserProject{
			UserID:    existingUser.ID,
			ProjectID: existingProject.ID,
			RoleID:    superAdminRole.ID,
			Status:    1,
		}
		if err := tx.Where("user_id = ? AND project_id = ?", existingUser.ID, existingProject.ID).FirstOrCreate(&userProject).Error; err != nil {
			return nil, fmt.Errorf("failed to add admin user to project %s: %v", project.Name, err)
		}
	}

	return &existingUser, nil
}

func initializeContainers(tx *gorm.DB, userID int) error {
	isPublic := true
	containers := []database.Container{
		{
			Type:     "algorithm",
			Name:     config.GetString("algo.detector"),
			Image:    "10.10.10.240/library/detector",
			Tag:      "d3bc0bf",
			Command:  "bash /entrypoint.sh",
			UserID:   userID,
			IsPublic: isPublic,
		},
		{
			Type:     "algorithm",
			Name:     "traceback",
			Image:    "10.10.10.240/library/rca-algo-traceback",
			Tag:      "latest",
			Command:  "bash /entrypoint.sh",
			UserID:   userID,
			IsPublic: isPublic,
		},
		{
			Type:     "benchmark",
			Name:     "clickhouse",
			Image:    "10.10.10.240/library/detector",
			Tag:      "d3bc0bf",
			Command:  "bash /entrypoint.sh",
			UserID:   userID,
			IsPublic: isPublic,
		},
		{
			Type:     "namespace",
			Name:     "ts",
			Image:    "10.10.10.240/library/ts-train-service",
			Tag:      "2ad833a2",
			UserID:   userID,
			IsPublic: isPublic,
		},
	}

	for _, resource := range containers {
		var existingContainer database.Container
		if err := tx.Where("type = ? AND name = ? AND image = ? AND tag = ?",
			resource.Type, resource.Name, resource.Image, resource.Tag).FirstOrCreate(&existingContainer, resource).Error; err != nil {
			return fmt.Errorf("failed to create container %s: %v", resource.Name, err)
		}
	}

	return nil
}
