package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"aegis/client/k8s"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/repository"
	"aegis/service/common"
	"aegis/service/consumer"
	producer "aegis/service/prodcuer"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
)

const AdminUsername = "admin"

type InitialDynamicConfig struct {
	Key          string                 `json:"key"`
	Value        string                 `json:"value"`
	DefaultValue string                 `json:"default_value"`
	ValueType    consts.ConfigValueType `json:"value_type"`
	Category     string                 `json:"category"`
	Description  string                 `json:"description"`
	IsSecret     bool                   `json:"is_secret"`
	MinValue     *float64               `json:"min_value,omitempty"`
	MaxValue     *float64               `json:"max_value,omitempty"`
	Pattern      string                 `json:"pattern,omitempty"`
	Options      string                 `json:"options,omitempty"`
}

func (c *InitialDynamicConfig) ConvertToDBDynamicConfig() *database.DynamicConfig {
	return &database.DynamicConfig{
		Key:          c.Key,
		Value:        c.Value,
		DefaultValue: c.DefaultValue,
		ValueType:    c.ValueType,
		Category:     c.Category,
		Description:  c.Description,
		IsSecret:     c.IsSecret,
		MinValue:     c.MinValue,
		MaxValue:     c.MaxValue,
		Pattern:      c.Pattern,
		Options:      c.Options,
	}
}

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
	Name       string                   `json:"name"`
	GithubLink string                   `json:"github_link"`
	ImageRef   string                   `json:"image_ref"`
	Command    string                   `json:"command"`
	EnvVars    []InitialParameterConfig `json:"env_vars"`
	Status     consts.StatusType        `json:"status"`
	HelmConfig *InitialHelmConfig       `json:"helm_config"`
}

func (cv *InitialContainerVersion) ConvertToDBContainerVersion() *database.ContainerVersion {
	return &database.ContainerVersion{
		Name:       cv.Name,
		GithubLink: cv.GithubLink,
		ImageRef:   cv.ImageRef,
		Command:    cv.Command,
		Status:     cv.Status,
	}
}

type InitialHelmConfig struct {
	ChartName string                   `json:"chart_name"`
	RepoName  string                   `json:"repo_name"`
	RepoURL   string                   `json:"repo_url"`
	Values    []InitialParameterConfig `json:"values"`
}

func (hc *InitialHelmConfig) ConvertToDBHelmConfig() *database.HelmConfig {
	return &database.HelmConfig{
		ChartName: hc.ChartName,
		RepoName:  hc.RepoName,
		RepoURL:   hc.RepoURL,
	}
}

type InitialParameterConfig struct {
	Key            string                   `json:"key"`
	Type           consts.ParameterType     `json:"type"`
	Category       consts.ParameterCategory `json:"category"`
	ValueType      consts.ValueDataType     `json:"value_type"`
	DefaultValue   *string                  `json:"default_value"`
	TemplateString *string                  `json:"template_string"`
	Required       bool                     `json:"required"`
	Overridable    *bool                    `json:"overridable"`
}

func (pc *InitialParameterConfig) ConvertToDBParameterConfig() *database.ParameterConfig {
	config := &database.ParameterConfig{
		Key:            pc.Key,
		Type:           pc.Type,
		Category:       pc.Category,
		ValueType:      pc.ValueType,
		DefaultValue:   pc.DefaultValue,
		TemplateString: pc.TemplateString,
		Required:       pc.Required,
		Overridable:    true, // default to true
	}

	// If overridable is explicitly set, use that value
	if pc.Overridable != nil {
		config.Overridable = *pc.Overridable
	}

	return config
}

type InitialDatasaet struct {
	Name        string                  `json:"name"`
	Type        string                  `json:"type"`
	Description string                  `json:"description"`
	IsPublic    bool                    `json:"is_public"`
	Status      consts.StatusType       `json:"status"`
	Versions    []InitialDatasetVersion `json:"versions"`
}

func (d *InitialDatasaet) ConvertToDBDataset() *database.Dataset {
	return &database.Dataset{
		Name:        d.Name,
		Type:        d.Type,
		Description: d.Description,
		IsPublic:    d.IsPublic,
		Status:      d.Status,
	}
}

type InitialDatasetVersion struct {
	Name   string            `json:"name"`
	Status consts.StatusType `json:"status"`
}

func (dv *InitialDatasetVersion) ConvertToDBDatasetVersion() *database.DatasetVersion {
	return &database.DatasetVersion{
		Name:   dv.Name,
		Status: dv.Status,
	}
}

type InitialDataProject struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Status      consts.StatusType `json:"status"`
}

func (p *InitialDataProject) ConvertToDBProject() *database.Project {
	return &database.Project{
		Name:        p.Name,
		Description: p.Description,
		Status:      p.Status,
	}
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
	DynamicConfigs []InitialDynamicConfig `json:"dynamic_configs"`
	Containers     []InitialDataContainer `json:"containers"`
	Datasets       []InitialDatasaet      `json:"datasets"`
	Projects       []InitialDataProject   `json:"projects"`
	AdminUser      InitialDataUser        `json:"admin_user"`
}

type ConsumerData struct {
	configs []database.DynamicConfig
}

var (
	consumerData  *ConsumerData
	resourceIDMap map[consts.ResourceName]int
)

// InitConcurrencyLock initializes the concurrency lock counter
func InitConcurrencyLock(ctx context.Context) {
	if err := repository.InitConcurrencyLock(ctx); err != nil {
		logrus.Fatalf("error setting concurrency lock to 0: %v", err)
	}
}

// InitializeProducer initializes data (roles, permissions, resources)
func InitializeProducer() {
	user, err := repository.GetUserByUsername(database.DB, AdminUsername)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logrus.Info("Seeding initial system data for producer...")
			if err := initializeProducer(); err != nil {
				logrus.Fatalf("Failed to initialize system data for producer: %v", err)
			}
			logrus.Info("Successfully seeded initial system data for producer")
		} else {
			logrus.Fatalf("Failed to check for %s: %v", AdminUsername, err)
		}
	}

	if user != nil {
		logrus.Info("Initial system data for producer already seeded, skipping initialization")
	}
}

// InitializeConsumer initializes dynamic configurations
func InitializeConsumer() {
	configs, err := repository.ListExistingConfigs(database.DB)
	if err != nil {
		logrus.Fatalf("Failed to check existing dynamic configs: %v", err)
	}

	consumerData = &ConsumerData{
		configs: configs,
	}

	if len(consumerData.configs) == 0 {
		logrus.Info("Seeding initial system data for consumer...")
		if err := initializeConsumer(); err != nil {
			logrus.Fatalf("Failed to initialize system data for consumer: %v", err)
		}
		logrus.Info("Successfully seeded initial system data for consumer")
	} else {
		logrus.Info("Initial system data for consumer already seeded, skipping initialization")
	}

	loadedCount := 0
	for _, cfg := range consumerData.configs {
		if err := config.SetViperValue(cfg.Key, cfg.Value, cfg.ValueType); err != nil {
			logrus.Fatalf("Failed to load dynamic config %s into viper: %v", cfg.Key, err)
		}
		loadedCount++

		// Log non-secret configs
		if !cfg.IsSecret {
			logrus.Debugf("Loaded dynamic config: %s = %s", cfg.Key, cfg.Value)
		} else {
			logrus.Debugf("Loaded dynamic config: %s = [REDACTED]", cfg.Key)
		}
	}

	config.GetChaosSystemConfigManager().RegisterCallback(onChaosSystemConfigUpdate)

	// Initialize namespaces on startup to ensure all enabled namespaces are properly initialized
	// This is critical after program restart to re-initialize CRD informers
	logrus.Info("Initializing namespaces on startup...")
	monitor := consumer.GetMonitor()
	if monitor == nil {
		logrus.Warn("Monitor not initialized, skipping namespace initialization")
	} else {
		initialized, err := monitor.InitializeNamespaces()
		if err != nil {
			logrus.Errorf("Failed to initialize namespaces: %v", err)
			return
		}

		if len(initialized) == 0 {
			logrus.Warn("No namespaces to initialize on startup")
			return
		}

		logrus.Infof("Initialized namespaces on startup: %v", initialized)
		if err := updateK8sController(initialized, []string{}); err != nil {
			logrus.Errorf("Failed to update k8s controller: %v", err)
			return
		}
	}
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

// initializeProducer initializes system data (roles, permissions, resources)
func initializeProducer() error {
	dataPath := config.GetString("initialization.data_path")
	filePath := filepath.Join(dataPath, consts.InitialFilename)
	initialData, err := loadInitialDataFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load initial data from file: %w", err)
	}

	resources := []database.Resource{
		{Name: consts.ResourceSystem, Type: consts.ResourceTypeSystem, Category: consts.ResourceAdmin},
		{Name: consts.ResourceAudit, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
		{Name: consts.ResourceConfigruation, Type: consts.ResourceTypeTable, Category: consts.ResourceAdmin},
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

	for i := range resources {
		resources[i].DisplayName = consts.GetResourceDisplayName(resources[i].Name)
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

	return withOptimizedDBSettings(func() error {
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
			resourceIDMap = make(map[consts.ResourceName]int, len(allResourcesInDB))
			for _, res := range allResourcesInDB {
				resourceIDMap[res.Name] = res.ID
				resourceMap[res.Name] = &res
			}

			resourceMap[consts.ResourceContainerVersion].ParentID = utils.IntPtr(resourceIDMap[consts.ResourceContainer])
			resourceMap[consts.ResourceDatasetVersion].ParentID = utils.IntPtr(resourceIDMap[consts.ResourceDataset])

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
						ResourceID:  resourceIDMap[resource.Name],
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

			// Create default containers from initial data
			if err := initializeContainers(tx, initialData, adminUser.ID); err != nil {
				return fmt.Errorf("failed to initialize containers: %w", err)
			}

			// Create default datasets from initial data
			if err := initializeDatasets(tx, initialData, adminUser.ID); err != nil {
				return fmt.Errorf("failed to initialize datasets: %w", err)
			}

			// Initialize execution result labels
			if err := initializeExecutionLabels(tx); err != nil {
				return fmt.Errorf("failed to initialize execution labels: %w", err)
			}

			return nil
		})
	})
}

// assignSystemRolePermissions assigns permissions to system roles
func assignSystemRolePermissions(tx *gorm.DB) error {
	for roleName, permissionNames := range consts.SystemRolePermissions {
		role, err := repository.GetRoleByName(tx, roleName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("role %s not found", roleName)
			}
			return err
		}

		if roleName == consts.RoleSuperAdmin {
			permissions, err := repository.ListSystemPermissions(tx)
			if err != nil {
				return fmt.Errorf("failed to list system permissions: %w", err)
			}

			var rolePermissions []database.RolePermission
			for _, perm := range permissions {
				rolePermissions = append(rolePermissions, database.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				})
			}

			if err := repository.BatchCreateRolePermissions(tx, rolePermissions); err != nil {
				return fmt.Errorf("failed to assign all permissions to super admin role: %w", err)
			}
		} else {
			var permissionStrs []string
			for _, name := range permissionNames {
				permissionStrs = append(permissionStrs, string(name))
			}

			permissions, err := repository.ListPermissionsByNames(tx, permissionStrs)
			if err != nil {
				return fmt.Errorf("failed to list permissions for role %s: %w", roleName, err)
			}

			var rolePermissions []database.RolePermission
			for _, perm := range permissions {
				rolePermissions = append(rolePermissions, database.RolePermission{
					RoleID:       role.ID,
					PermissionID: perm.ID,
				})
			}

			if err := repository.BatchCreateRolePermissions(tx, rolePermissions); err != nil {
				return fmt.Errorf("failed to assign permissions to role %s: %w", roleName, err)
			}
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
	for _, projectData := range data.Projects {
		project := projectData.ConvertToDBProject()
		if err := repository.CreateProject(tx, project); err != nil {
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
	dataPath := config.GetString("initialization.data_path")

	for _, containerData := range data.Containers {
		container := containerData.ConvertToDBContainer()
		if container.Type == consts.ContainerTypePedestal {
			system := chaos.SystemType(container.Name)
			if !system.IsValid() {
				return fmt.Errorf("invalid pedestal name: %s", container.Name)
			}
		}

		versions := make([]database.ContainerVersion, 0, len(containerData.Versions))
		for _, versionData := range containerData.Versions {
			version := versionData.ConvertToDBContainerVersion()

			if len(versionData.EnvVars) > 0 {
				params := make([]database.ParameterConfig, 0, len(versionData.EnvVars))
				for _, paramData := range versionData.EnvVars {
					param := paramData.ConvertToDBParameterConfig()
					params = append(params, *param)
				}
				version.EnvVars = params
			}

			if versionData.HelmConfig != nil {
				helmConfig := versionData.HelmConfig.ConvertToDBHelmConfig()
				if len(versionData.HelmConfig.Values) > 0 {
					params := make([]database.ParameterConfig, 0, len(versionData.HelmConfig.Values))
					for _, paramData := range versionData.HelmConfig.Values {
						param := paramData.ConvertToDBParameterConfig()
						params = append(params, *param)
					}
					helmConfig.DynamicValues = params
				}

				version.HelmConfig = helmConfig
			}

			versions = append(versions, *version)
		}

		container.Versions = versions

		createdContainer, err := producer.CreateContainerCore(tx, container, userID)
		if err != nil {
			return fmt.Errorf("failed to create container %s: %w", containerData.Name, err)
		}

		if createdContainer.Type == consts.ContainerTypePedestal {
			if err := producer.UploadHemlValueFileCore(
				tx,
				container.Versions[0].HelmConfig,
				nil,
				filepath.Join(dataPath, fmt.Sprintf("%s.yaml", createdContainer.Name)),
			); err != nil {
				return fmt.Errorf("failed to upload helm value file for container %s: %w", containerData.Name, err)
			}
		}
	}

	return nil
}

// initializeDatasets initializes default datasets from initial data
func initializeDatasets(tx *gorm.DB, data *InitialData, userID int) error {
	for _, datasetData := range data.Datasets {
		dataset := datasetData.ConvertToDBDataset()

		versions := make([]database.DatasetVersion, 0, len(datasetData.Versions))
		for _, versionData := range datasetData.Versions {
			version := versionData.ConvertToDBDatasetVersion()
			versions = append(versions, *version)
		}

		_, err := producer.CreateDatasetCore(tx, dataset, versions, userID)
		if err != nil {
			return fmt.Errorf("failed to create dataset %s: %w", datasetData.Name, err)
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

func initializeConsumer() error {
	dataPath := config.GetString("initialization.data_path")
	filePath := filepath.Join(dataPath, consts.InitialFilename)
	initialData, err := loadInitialDataFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load initial data from file: %w", err)
	}

	return withOptimizedDBSettings(func() error {
		err := database.DB.Transaction(func(tx *gorm.DB) error {
			if err := initializeDynamicConfigs(tx, initialData); err != nil {
				return fmt.Errorf("failed to initialize dynamic configs for consumer: %w", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("failed to initialize consumer data: %w", err)
		}

		return nil
	})
}

// initializeDynamicConfigs initializes dynamic configuration items from initial data
func initializeDynamicConfigs(tx *gorm.DB, data *InitialData) error {
	var configs []database.DynamicConfig
	for _, configData := range data.DynamicConfigs {
		cfg := configData.ConvertToDBDynamicConfig()
		if err := common.ValidateConfig(cfg); err != nil {
			return fmt.Errorf("invalid config value for key %s: %w", configData.Key, err)
		}

		if err := common.CreateConfig(tx, cfg); err != nil {
			return fmt.Errorf("failed to create dynamic config %s: %w", configData.Key, err)
		}

		configs = append(configs, *cfg)
	}

	consumerData.configs = configs
	return nil
}

// onChaosSystemConfigUpdate is called when chaos system config is updated
func onChaosSystemConfigUpdate() error {
	logrus.Info("Chaos system configuration updated, refreshing namespaces...")

	// Refresh Monitor namespaces
	monitor := consumer.GetMonitor()
	if monitor == nil {
		logrus.Warn("Monitor not initialized, skipping namespace refresh")
		return nil
	}

	result, err := monitor.RefreshNamespaces()
	if err != nil {
		return fmt.Errorf("failed to refresh namespaces: %w", err)
	}

	// Log detailed refresh results
	totalChanges := len(result.Added) + len(result.Recovered) + len(result.Disabled) + len(result.Deleted)
	logrus.Infof("Namespace refresh completed: %d total changes", totalChanges)

	if len(result.Added) > 0 {
		logrus.Infof("Added namespaces: %v", result.Added)
	}
	if len(result.Recovered) > 0 {
		logrus.Infof("Recovered namespaces: %v", result.Recovered)
	}
	if len(result.Disabled) > 0 {
		logrus.Warnf("Disabled namespaces (have active locks): %v", result.Disabled)
	}
	if len(result.Deleted) > 0 {
		logrus.Infof("Deleted namespaces (no active locks): %v", result.Deleted)
	}

	// Add informers for newly added and recovered namespaces
	namespacesToAdd := make([]string, 0, len(result.Added)+len(result.Recovered))
	namespacesToAdd = append(namespacesToAdd, result.Added...)
	namespacesToAdd = append(namespacesToAdd, result.Recovered...)

	// Remove informers for disabled and deleted namespaces
	namespacesToRemove := make([]string, 0, len(result.Disabled)+len(result.Deleted))
	namespacesToRemove = append(namespacesToRemove, result.Disabled...)
	namespacesToRemove = append(namespacesToRemove, result.Deleted...)

	if err := updateK8sController(namespacesToAdd, namespacesToRemove); err != nil {
		return fmt.Errorf("failed to update k8s controller: %w", err)
	}

	return nil
}

// updateK8sController updates the Kubernetes controller with the latest namespace configurations
func updateK8sController(toAdd, toRemove []string) error {
	controller := k8s.GetK8sController()
	if controller == nil {
		logrus.Warn("Controller not initialized, skipping informer update")
		return nil
	}

	if len(toAdd) > 0 {
		logrus.Infof("Adding informers for active namespaces: %v", toAdd)
		if err := controller.AddNamespaceInformers(toAdd); err != nil {
			return fmt.Errorf("failed to add namespace informers: %w", err)
		}
	}

	if len(toRemove) > 0 {
		logrus.Infof("Marking namespaces as inactive: %v", toRemove)
		controller.RemoveNamespaceInformers(toRemove)
	}

	return nil
}

// withOptimizedDBSettings executes a function with optimized database settings for bulk operations.
// It temporarily disables foreign key and unique checks to speed up initialization.
func withOptimizedDBSettings(fn func() error) error {
	// Disable foreign key and unique checks for faster initialization
	if err := database.DB.Exec("SET FOREIGN_KEY_CHECKS=0").Error; err != nil {
		logrus.Warnf("Failed to disable foreign key checks: %v", err)
	}
	if err := database.DB.Exec("SET UNIQUE_CHECKS=0").Error; err != nil {
		logrus.Warnf("Failed to disable unique checks: %v", err)
	}

	// Ensure checks are re-enabled even if the function fails
	defer func() {
		if err := database.DB.Exec("SET FOREIGN_KEY_CHECKS=1").Error; err != nil {
			logrus.Errorf("Failed to re-enable foreign key checks: %v", err)
		}
		if err := database.DB.Exec("SET UNIQUE_CHECKS=1").Error; err != nil {
			logrus.Errorf("Failed to re-enable unique checks: %v", err)
		}
	}()

	return fn()
}
