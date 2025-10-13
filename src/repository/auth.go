package repository

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"aegis/config"
	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
)

type InitialDataContainer struct {
	Type       string             `json:"type"`
	Name       string             `json:"name"`
	Registry   string             `json:"registry"`
	Repository string             `json:"repository"`
	Tag        string             `json:"tag"`
	Command    string             `json:"command"`
	EnvVars    string             `json:"env_vars"`
	IsPublic   bool               `json:"is_public"`
	Status     int                `json:"status"`
	HelmConfig *InitialHelmConfig `json:"helm_config"`
}

type InitialHelmConfig struct {
	ChartName    string         `json:"chart_name"`
	RepoName     string         `json:"repo_name"`
	RepoURL      string         `json:"repo_url"`
	Values       map[string]any `json:"values"`
	NsPrefix     string         `json:"ns_prefix"`
	PortTemplate string         `json:"port_template"`
}

type InitialDataUser struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
	Status   int    `json:"status"`
	IsActive bool   `json:"is_active"`
}

type InitialData struct {
	Containers []InitialDataContainer `json:"containers"`
	Projects   []database.Project     `json:"projects"`
	AdminUser  InitialDataUser        `json:"admin_user"`
}

// InitializeSystemData initializes system data (roles, permissions, resources)
func InitializeSystemData() error {
	dataPath := config.GetString("initialization.data_path")
	filePath := filepath.Join(dataPath, consts.InitialFilename)
	initialData, err := loadInitialDataFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load initial data from file: %v", err)
	}

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
			{Name: consts.ResourceLabel.String(), DisplayName: "Label", Type: "table", Category: "admin", IsSystem: true, Status: 1},
			{Name: consts.ResourceSystem.String(), DisplayName: "System", Type: "system", Category: "admin", IsSystem: true, Status: 1},
			{Name: consts.ResourceAudit.String(), DisplayName: "Audit", Type: "table", Category: "admin", IsSystem: true, Status: 1},
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
		adminUser, err := initializeAdminUserAndProjects(tx, initialData)
		if err != nil {
			return fmt.Errorf("failed to initialize admin user and projects: %v", err)
		}

		if err := initializeContainers(tx, initialData, adminUser.ID); err != nil {
			return fmt.Errorf("failed to initialize containers: %v", err)
		}

		// Initialize execution result labels
		if err := initializeExecutionLabels(tx); err != nil {
			return fmt.Errorf("failed to initialize execution labels: %v", err)
		}

		return nil
	})
}

func loadInitialDataFromFile(filePath string) (*InitialData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read initial data file: %v", err)
	}

	var initialData InitialData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal initial data: %v", err)
	}

	return &initialData, nil
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
func initializeAdminUserAndProjects(tx *gorm.DB, data *InitialData) (*database.User, error) {
	// 1. Create super admin user
	adminUser := database.User{
		Username: data.AdminUser.Username,
		Email:    data.AdminUser.Email,
		Password: data.AdminUser.Password, // Plain password, will be hashed in BeforeCreate hook
		FullName: data.AdminUser.FullName,
		Status:   data.AdminUser.Status,
		IsActive: data.AdminUser.IsActive,
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
	defaultProjects := data.Projects

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

// initializeContainers initializes default containers from initial data
func initializeContainers(tx *gorm.DB, data *InitialData, userID int) error {
	for _, containerData := range data.Containers {
		container := &database.Container{
			Type:         containerData.Type,
			Name:         containerData.Name,
			Registry:     containerData.Registry,
			Repository:   containerData.Repository,
			Command:      containerData.Command,
			EnvVars:      containerData.EnvVars,
			UserID:       userID,
			HelmConfigID: nil,
			IsPublic:     containerData.IsPublic,
			Status:       containerData.Status,
		}

		var helmConfig *database.HelmConfig
		if containerData.HelmConfig != nil {
			valuesBytes, err := json.Marshal(containerData.HelmConfig.Values)
			if err != nil {
				return fmt.Errorf("failed to marshal default helm values for container %s: %v", containerData.Name, err)
			}

			helmConfig = &database.HelmConfig{
				ChartName:    containerData.HelmConfig.ChartName,
				RepoName:     containerData.HelmConfig.RepoName,
				RepoURL:      containerData.HelmConfig.RepoURL,
				Values:       string(valuesBytes),
				NsPrefix:     containerData.HelmConfig.NsPrefix,
				PortTemplate: containerData.HelmConfig.PortTemplate,
			}

			if err := CreateHelmConfig(helmConfig); err != nil {
				return fmt.Errorf("failed to create helm config for container %s: %v", containerData.Name, err)
			}

			container.HelmConfigID = &helmConfig.ID
		}

		if err := CreateContainerWithTx(tx, container, containerData.Tag); err != nil {
			return fmt.Errorf("failed to create container %s: %v", container.Name, err)
		}
	}

	return nil
}

// initializeExecutionLabels initializes system labels for execution results
func initializeExecutionLabels(tx *gorm.DB) error {
	// Initialize source labels
	sourceLabels := []struct {
		value       string
		description string
	}{
		{consts.ExecutionSourceManual, consts.ExecutionManualDescription},
		{consts.ExecutionSourceSystem, consts.ExecutionSystemDescription},
	}

	for _, labelInfo := range sourceLabels {
		_, err := CreateOrGetLabelWithTx(tx, consts.ExecutionLabelSource, labelInfo.value, consts.LabelExecution, labelInfo.description)
		if err != nil {
			return fmt.Errorf("failed to initialize execution label %s=%s: %v",
				consts.ExecutionLabelSource, labelInfo.value, err)
		}
	}

	return nil
}
