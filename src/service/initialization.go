package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/repository"
	producer "aegis/service/prodcuer"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type InitialDataContainer struct {
	Type     consts.ContainerType      `json:"type"`
	Name     string                    `json:"name"`
	IsPublic bool                      `json:"is_public"`
	Status   consts.StatusType         `json:"status"`
	Versions []InitialContainerVersion `json:"versions"`
}

func (c *InitialDataContainer) ConvertToDBContainer() *database.Container {
	return &database.Container{
		Type:     c.Type,
		Name:     c.Name,
		IsPublic: c.IsPublic,
		Status:   c.Status,
	}
}

type InitialContainerVersion struct {
	Name       string             `json:"name"`
	GithubLink string             `json:"github_link"`
	ImageRef   string             `json:"image_ref"`
	Command    string             `json:"command"`
	EnvVars    string             `json:"env_vars"`
	Status     consts.StatusType  `json:"status"`
	HelmConfig *InitialHelmConfig `json:"helm_config"`
}

func (cv *InitialContainerVersion) ConvertToDBContainerVersion() *database.ContainerVersion {
	return &database.ContainerVersion{
		Name:       cv.Name,
		GithubLink: cv.GithubLink,
		ImageRef:   cv.ImageRef,
		Command:    cv.Command,
		EnvVars:    cv.EnvVars,
		Status:     cv.Status,
	}
}

type InitialHelmConfig struct {
	ChartName    string         `json:"chart_name"`
	RepoName     string         `json:"repo_name"`
	RepoURL      string         `json:"repo_url"`
	NsPrefix     string         `json:"ns_prefix"`
	PortTemplate string         `json:"port_template"`
	Values       map[string]any `json:"values"`
}

func (hc *InitialHelmConfig) ConvertToDBHelmConfig() (*database.HelmConfig, error) {
	valuesBytes, err := json.Marshal(hc.Values)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal helm values: %w", err)
	}

	return &database.HelmConfig{
		RepoURL:      hc.RepoURL,
		RepoName:     hc.RepoName,
		ChartName:    hc.ChartName,
		NsPrefix:     hc.NsPrefix,
		PortTemplate: hc.PortTemplate,
		Values:       string(valuesBytes),
	}, nil
}

type InitialDataUser struct {
	Username string            `json:"username"`
	Email    string            `json:"email"`
	Password string            `json:"password"`
	FullName string            `json:"full_name"`
	Status   consts.StatusType `json:"status"`
	IsActive bool              `json:"is_active"`
}

func (u *InitialDataUser) ConvertToDBUser() *database.User {
	return &database.User{
		Username: u.Username,
		Email:    u.Email,
		Password: u.Password,
		FullName: u.FullName,
		Status:   u.Status,
		IsActive: u.IsActive,
	}
}

type InitialData struct {
	Containers []InitialDataContainer `json:"containers"`
	Projects   []database.Project     `json:"projects"`
	AdminUser  InitialDataUser        `json:"admin_user"`
}

var ResourceIDMap map[consts.ResourceName]int

// InitConcurrencyLock initializes the concurrency lock counter
func InitConcurrencyLock(ctx context.Context) {
	if err := repository.InitConcurrencyLock(ctx); err != nil {
		logrus.Fatalf("error setting concurrency lock to 0: %v", err)
	}
}

// InitializeData initializes data (roles, permissions, resources)
func InitializeData(ctx context.Context) {
	if !repository.IsInitialDataSeeded(ctx) {
		if err := initialize(); err != nil {
			logrus.Errorf("Failed to initialize system data: %v", err)
			logrus.Warn("System will continue running without initial system data")
		} else {
			logrus.Info("System data initialized successfully")
			if markErr := repository.MarkDataSeedingComplete(ctx); markErr != nil {
				logrus.Fatalf("Failed to mark data seeding as complete: %v", markErr)
			}
		}
	} else {
		logrus.Info("Initial system data already seeded, skipping initialization")
	}
}

// InitializeSystemData initializes system data (roles, permissions, resources)
func initialize() error {
	dataPath := config.GetString("initialization.data_path")
	filePath := filepath.Join(dataPath, consts.InitialFilename)
	initialData, err := loadInitialDataFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load initial data from file: %w", err)
	}

	resources := []database.Resource{
		{Name: consts.ResourceSystem, Type: consts.ResourceTypeSystem, Category: consts.ResourceAdmin},
		{Name: consts.ResourceAudit, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceContainer, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceContainerVersion, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceDataset, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceDatasetVersion, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceProject, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceLabel, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceUser, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceRole, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourcePermission, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceTask, Type: consts.ResourceTypeTable, Category: consts.ResourceCore},
		{Name: consts.ResourceTrace, Type: consts.ResourceTypeTable, Category: consts.ResourceCore},
		{Name: consts.ResourceInjection, Type: consts.ResourceTypeTable, Category: consts.ResourceCore},
		{Name: consts.ResourceExecution, Type: consts.ResourceTypeTable, Category: consts.ResourceCore},
	}

	systemRoles := make([]database.Role, 0)
	for role, displayName := range consts.SystemRoleDisplayNames {
		systemRoles = append(systemRoles, database.Role{
			Name:        string(role),
			DisplayName: displayName,
			IsSystem:    true,
			Status:      consts.CommonEnabled,
		})
	}

	actions := make([]consts.ActionName, 0, len(consts.ValidActions))
	for action := range consts.ValidActions {
		actions = append(actions, action)
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Create system resources
		if err := repository.BatchUpsertResources(tx, resources); err != nil {
			return fmt.Errorf("failed to create system resources: %w", err)
		}

		resourceNames := make([]consts.ResourceName, 0, len(resources))
		for _, res := range resources {
			resourceNames = append(resourceNames, res.Name)
		}

		allResourcesInDB, err := repository.ListResourcesByNames(tx, resourceNames)
		if err != nil {
			return fmt.Errorf("failed to get system resources from database: %w", err)
		}

		if len(allResourcesInDB) != len(resources) {
			return fmt.Errorf("mismatch in number of resources created and fetched")
		}

		resourceMap := make(map[consts.ResourceName]*database.Resource, len(allResourcesInDB))
		for _, res := range allResourcesInDB {
			ResourceIDMap[res.Name] = res.ID
			resourceMap[res.Name] = &res
		}

		resourceMap[consts.ResourceContainerVersion].ParentID = utils.IntPtr(ResourceIDMap[consts.ResourceContainer])
		resourceMap[consts.ResourceDatasetVersion].ParentID = utils.IntPtr(ResourceIDMap[consts.ResourceDataset])

		toUpdatedResources := []database.Resource{
			*resourceMap[consts.ResourceContainerVersion],
			*resourceMap[consts.ResourceDatasetVersion],
		}

		if err := repository.BatchUpsertResources(tx, toUpdatedResources); err != nil {
			return fmt.Errorf("failed to update resource parent IDs: %w", err)
		}

		// Create system permissions
		var permissionsToCreate []database.Permission
		for _, resource := range allResourcesInDB {
			for _, action := range actions {
				permission := database.Permission{
					Name:        producer.GetPermissionName(action, resource.Name),
					DisplayName: producer.GetPermissionDisplayName(action, resource.DisplayName),
					Action:      string(action),
					ResourceID:  ResourceIDMap[resource.Name],
					IsSystem:    true,
					Status:      consts.CommonEnabled,
				}
				permissionsToCreate = append(permissionsToCreate, permission)
			}
		}

		if err := repository.BatchUpsertPermissions(tx, permissionsToCreate); err != nil {
			return fmt.Errorf("failed to create system permissions: %w", err)
		}

		// Create system roles
		if err := repository.BatchUpsertRoles(tx, systemRoles); err != nil {
			return fmt.Errorf("failed to create system roles: %w", err)
		}

		// Assign permissions to system roles
		if err := assignSystemRolePermissions(tx); err != nil {
			return fmt.Errorf("failed to assign system role permissions: %w", err)
		}

		// Create super admin user and default project
		adminUser, err := initializeAdminUserAndProjects(tx, initialData)
		if err != nil {
			return fmt.Errorf("failed to initialize admin user and projects: %w", err)
		}

		if err := initializeContainers(tx, initialData, adminUser.ID); err != nil {
			return fmt.Errorf("failed to initialize containers: %w", err)
		}

		// Initialize execution result labels
		if err := initializeExecutionLabels(tx); err != nil {
			return fmt.Errorf("failed to initialize execution labels: %w", err)
		}

		return nil
	})
}

// loadInitialDataFromFile loads initial data from a JSON file
func loadInitialDataFromFile(filePath string) (*InitialData, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read initial data file: %w", err)
	}

	var initialData InitialData
	if err := json.Unmarshal(data, &initialData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal initial data: %w", err)
	}

	return &initialData, nil
}

// assignSystemRolePermissions assigns permissions to system roles
func assignSystemRolePermissions(tx *gorm.DB) error {
	for role, permissions := range consts.SystemRolePermissions {
		if role == consts.RoleSuperAdmin {
			if err := assignAllPermissionsToRole(tx, consts.RoleSuperAdmin); err != nil {
				return err
			}
		}

		var permissionStrs []string
		for _, perm := range permissions {
			permissionStrs = append(permissionStrs, string(perm))
		}

		if err := assignPermissionsToRole(tx, role, permissionStrs); err != nil {
			return err
		}
	}

	return nil
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
			continue
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
	adminUser := data.AdminUser.ConvertToDBUser()

	if err := repository.CreateUser(tx, adminUser); err != nil {
		if errors.Is(err, consts.ErrAlreadyExists) {
			return nil, fmt.Errorf("admin user already exists")
		}
		return nil, fmt.Errorf("failed to create admin user: %w", err)
	}

	superAdminRole, err := repository.GetRoleByName(tx, "super_admin")
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("super_admin role not found, ensure system roles are initialized first")
		}
		return nil, fmt.Errorf("failed to get super_admin role: %w", err)
	}

	userRole := database.UserRole{
		UserID: adminUser.ID,
		RoleID: superAdminRole.ID,
	}
	if err := repository.CreateUserRole(tx, &userRole); err != nil {
		if errors.Is(err, consts.ErrAlreadyExists) {
			return nil, fmt.Errorf("admin user already has super_admin role")
		}
		return nil, fmt.Errorf("failed to assign super_admin role to admin user: %w", err)
	}

	// 3. Create default projects
	for _, project := range data.Projects {
		if err := repository.CreateProject(tx, &project); err != nil {
			if errors.Is(err, consts.ErrAlreadyExists) {
				return nil, fmt.Errorf("project %s already exists", project.Name)
			}
			return nil, fmt.Errorf("failed to create project %s: %w", project.Name, err)
		}

		if err := repository.CreateUserProject(tx, &database.UserProject{
			UserID:    adminUser.ID,
			ProjectID: project.ID,
			RoleID:    superAdminRole.ID,
			Status:    consts.CommonEnabled,
		}); err != nil {
			return nil, fmt.Errorf("failed to add admin user to project %s: %w", project.Name, err)
		}
	}

	return adminUser, nil
}

// initializeContainers initializes default containers from initial data
func initializeContainers(tx *gorm.DB, data *InitialData, userID int) error {
	for _, containerData := range data.Containers {
		container := containerData.ConvertToDBContainer()

		versions := make([]database.ContainerVersion, 0, len(containerData.Versions))
		helmConfigs := make([]*database.HelmConfig, 0, len(containerData.Versions))
		for _, versionData := range containerData.Versions {
			version := versionData.ConvertToDBContainerVersion()
			versions = append(versions, *version)

			if versionData.HelmConfig != nil {
				helmConfig, err := versionData.HelmConfig.ConvertToDBHelmConfig()
				if err != nil {
					return fmt.Errorf("failed to convert helm config for container %s version %s: %w",
						containerData.Name, versionData.Name, err)
				}
				helmConfigs = append(helmConfigs, helmConfig)
			}
		}

		_, err := producer.CreateContainerCore(tx, container, versions, helmConfigs, userID)
		if err != nil {
			return fmt.Errorf("failed to create container %s: %w", containerData.Name, err)
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
		_, err := producer.CreateLabelCore(tx, &database.Label{
			Key:         consts.ExecutionLabelSource,
			Value:       labelInfo.value,
			Category:    consts.ExecutionCategory,
			Description: labelInfo.description,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize execution label %s=%s: %w",
				consts.ExecutionLabelSource, labelInfo.value, err)
		}
	}

	return nil
}
