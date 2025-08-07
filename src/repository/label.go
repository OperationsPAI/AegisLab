package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// CreateOrGetLabel creates a new label or gets existing one
func CreateOrGetLabel(key, value, category, description string) (*database.Label, error) {
	var label database.Label

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(logger.Silent)}).
			Where(&database.Label{Key: key, Value: value, Category: category}).
			First(&label).Error
		if err == nil {
			return tx.Model(&label).UpdateColumn("usage_count", gorm.Expr("usage_count + ?", 1)).Error
		}

		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to query label: %v", err)
		}

		label = database.Label{
			Key:         key,
			Value:       value,
			Category:    category,
			Description: description,
			Color:       "#1890ff",
			IsSystem:    true,
			Usage:       1,
		}
		return tx.Create(&label).Error
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create or get label: %v", err)
	}

	return &label, nil
}

// GetLabelByID gets label by ID
func GetLabelByID(id int) (*database.Label, error) {
	var label database.Label
	if err := database.DB.First(&label, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("label with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get label: %v", err)
	}
	return &label, nil
}

// GetLabelByKeyValue gets label by key-value pair
func GetLabelByKeyValue(key, value string) (*database.Label, error) {
	var label database.Label
	if err := database.DB.Where("label_key = ? AND label_value = ?", key, value).First(&label).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("label '%s:%s' not found", key, value)
		}
		return nil, fmt.Errorf("failed to get label: %v", err)
	}
	return &label, nil
}

// UpdateLabel updates label information
func UpdateLabel(label *database.Label) error {
	if err := database.DB.Save(label).Error; err != nil {
		return fmt.Errorf("failed to update label: %v", err)
	}
	return nil
}

// UpdateLabelUsage updates the usage of a label (any increase or decrease)
func UpdateLabelUsage(id int, increment int) error {
	var operation string
	if increment >= 0 {
		operation = fmt.Sprintf("usage_count + %d", increment)
	} else {
		operation = fmt.Sprintf("GREATEST(0, usage_count + %d)", increment)
	}

	result := database.DB.Model(&database.Label{}).Where("id = ?", id).
		UpdateColumn("usage_count", gorm.Expr(operation))

	if result.Error != nil {
		return fmt.Errorf("failed to update label usage_count: %v", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("label with id %d not found", id)
	}

	return nil
}

// DeleteLabel deletes label (hard delete, because relationships should also be cleaned up when label is deleted)
func DeleteLabel(id int) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// Delete all related relationships first
		if err := tx.Where("label_id = ?", id).Delete(&database.DatasetLabel{}).Error; err != nil {
			return fmt.Errorf("failed to delete dataset label relations: %v", err)
		}

		if err := tx.Where("label_id = ?", id).Delete(&database.FaultInjectionLabel{}).Error; err != nil {
			return fmt.Errorf("failed to delete fault injection label relations: %v", err)
		}

		if err := tx.Where("label_id = ?", id).Delete(&database.ContainerLabel{}).Error; err != nil {
			return fmt.Errorf("failed to delete container label relations: %v", err)
		}

		if err := tx.Where("label_id = ?", id).Delete(&database.ProjectLabel{}).Error; err != nil {
			return fmt.Errorf("failed to delete project label relations: %v", err)
		}

		// Finally delete the label itself
		if err := tx.Delete(&database.Label{}, id).Error; err != nil {
			return fmt.Errorf("failed to delete label: %v", err)
		}

		return nil
	})
}

// ListLabels gets the label list
func ListLabels(page, pageSize int, category string, isSystem *bool) ([]database.Label, int64, error) {
	var labels []database.Label
	var total int64

	query := database.DB.Model(&database.Label{})

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if isSystem != nil {
		query = query.Where("is_system = ?", *isSystem)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count labels: %v", err)
	}

	// Pagination query
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("usage_count DESC, created_at DESC").Find(&labels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list labels: %v", err)
	}

	return labels, total, nil
}

// SearchLabels searches for labels
func SearchLabels(keyword string, category string, limit int) ([]database.Label, error) {
	var labels []database.Label

	query := database.DB.Model(&database.Label{})

	if keyword != "" {
		query = query.Where("key ILIKE ? OR value ILIKE ? OR description ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("usage_count DESC, created_at DESC").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to search labels: %v", err)
	}

	return labels, nil
}

// GetPopularLabels gets popular labels
func GetPopularLabels(category string, limit int) ([]database.Label, error) {
	var labels []database.Label

	query := database.DB.Model(&database.Label{}).Where("usage_count > 0")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("usage_count DESC").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get popular labels: %v", err)
	}

	return labels, nil
}

// GetUnusedLabels gets unused labels
func GetUnusedLabels(category string) ([]database.Label, error) {
	var labels []database.Label

	query := database.DB.Model(&database.Label{}).Where("usage_count = 0")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Order("created_at ASC").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get unused labels: %v", err)
	}

	return labels, nil
}

// CleanupUnusedLabels cleans up unused labels
func CleanupUnusedLabels(olderThanDays int) (int64, error) {
	var count int64

	query := database.DB.Model(&database.Label{}).
		Where("usage_count = 0 AND is_system = false AND created_at < NOW() - INTERVAL ? DAY", olderThanDays)

	// First get the count to be deleted
	if err := query.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count unused labels: %v", err)
	}

	// Execute deletion
	if err := query.Delete(&database.Label{}).Error; err != nil {
		return 0, fmt.Errorf("failed to cleanup unused labels: %v", err)
	}

	return count, nil
}

// GetLabelsByCategory gets labels by category
func GetLabelsByCategory(category string) ([]database.Label, error) {
	var labels []database.Label
	if err := database.DB.Where("category = ?", category).
		Order("usage_count DESC").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get labels by category: %v", err)
	}
	return labels, nil
}

// GetSystemLabels gets system labels
func GetSystemLabels() ([]database.Label, error) {
	var labels []database.Label
	if err := database.DB.Where("is_system = true").
		Order("category, key, value").Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to get system labels: %v", err)
	}
	return labels, nil
}

// GetInjectionLabelsMap gets all labels for multiple injections in batch (optimized)
func GetInjectionLabelsMap(injectionIDs []int) (map[int][]database.Label, error) {
	if len(injectionIDs) == 0 {
		return make(map[int][]database.Label), nil
	}

	// Method 1: Use single query with direct model association
	var relations []database.FaultInjectionLabel
	if err := database.DB.Preload("Label").
		Where("fault_injection_id IN ?", injectionIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get injection label relations: %v", err)
	}

	// Group labels by injection ID
	labelsMap := make(map[int][]database.Label)
	for _, relation := range relations {
		if relation.Label != nil {
			labelsMap[relation.FaultInjectionID] = append(labelsMap[relation.FaultInjectionID], *relation.Label)
		}
	}

	// Initialize empty slices for injections with no labels
	for _, id := range injectionIDs {
		if _, exists := labelsMap[id]; !exists {
			labelsMap[id] = []database.Label{}
		}
	}

	return labelsMap, nil
}

// GetDatasetLabelsMap gets all labels for multiple datasets in batch (optimized)
func GetDatasetLabelsMap(datasetIDs []int) (map[int][]database.Label, error) {
	if len(datasetIDs) == 0 {
		return make(map[int][]database.Label), nil
	}

	var relations []database.DatasetLabel
	if err := database.DB.Preload("Label").
		Where("dataset_id IN ?", datasetIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get dataset label relations: %v", err)
	}

	labelsMap := make(map[int][]database.Label)
	for _, relation := range relations {
		if relation.Label != nil {
			labelsMap[relation.DatasetID] = append(labelsMap[relation.DatasetID], *relation.Label)
		}
	}

	for _, id := range datasetIDs {
		if _, exists := labelsMap[id]; !exists {
			labelsMap[id] = []database.Label{}
		}
	}

	return labelsMap, nil
}

// GetExecutionLabelsMap gets all labels for multiple execution results in batch (optimized)
func GetExecutionLabelsMap(executionIDs []int) (map[int][]database.Label, error) {
	if len(executionIDs) == 0 {
		return make(map[int][]database.Label), nil
	}

	var relations []database.ExecutionResultLabel
	if err := database.DB.Preload("Label").
		Where("execution_id IN ?", executionIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get execution label relations: %v", err)
	}

	labelsMap := make(map[int][]database.Label)
	for _, relation := range relations {
		if relation.Label != nil {
			labelsMap[relation.ExecutionID] = append(labelsMap[relation.ExecutionID], *relation.Label)
		}
	}

	for _, id := range executionIDs {
		if _, exists := labelsMap[id]; !exists {
			labelsMap[id] = []database.Label{}
		}
	}

	return labelsMap, nil
}

// GetContainerLabelsMap gets all labels for multiple containers in batch (optimized)
func GetContainerLabelsMap(containerIDs []int) (map[int][]database.Label, error) {
	if len(containerIDs) == 0 {
		return make(map[int][]database.Label), nil
	}

	var relations []database.ContainerLabel
	if err := database.DB.Preload("Label").
		Where("container_id IN ?", containerIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get container label relations: %v", err)
	}

	labelsMap := make(map[int][]database.Label)
	for _, relation := range relations {
		if relation.Label != nil {
			labelsMap[relation.ContainerID] = append(labelsMap[relation.ContainerID], *relation.Label)
		}
	}

	for _, id := range containerIDs {
		if _, exists := labelsMap[id]; !exists {
			labelsMap[id] = []database.Label{}
		}
	}

	return labelsMap, nil
}

// GetProjectLabelsMap gets all labels for multiple projects in batch (optimized)
func GetProjectLabelsMap(projectIDs []int) (map[int][]database.Label, error) {
	if len(projectIDs) == 0 {
		return make(map[int][]database.Label), nil
	}

	var relations []database.ProjectLabel
	if err := database.DB.Preload("Label").
		Where("project_id IN ?", projectIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get project label relations: %v", err)
	}

	labelsMap := make(map[int][]database.Label)
	for _, relation := range relations {
		if relation.Label != nil {
			labelsMap[relation.ProjectID] = append(labelsMap[relation.ProjectID], *relation.Label)
		}
	}

	for _, id := range projectIDs {
		if _, exists := labelsMap[id]; !exists {
			labelsMap[id] = []database.Label{}
		}
	}

	return labelsMap, nil
}

// User-Role-Permission Relationship Query Functions

// GetUserRolesMap gets all roles for multiple users in batch (optimized)
func GetUserRolesMap(userIDs []int) (map[int][]database.Role, error) {
	if len(userIDs) == 0 {
		return make(map[int][]database.Role), nil
	}

	var relations []database.UserRole
	if err := database.DB.Preload("Role").
		Where("user_id IN ?", userIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get user role relations: %v", err)
	}

	rolesMap := make(map[int][]database.Role)
	for _, relation := range relations {
		if relation.Role != nil {
			rolesMap[relation.UserID] = append(rolesMap[relation.UserID], *relation.Role)
		}
	}

	for _, id := range userIDs {
		if _, exists := rolesMap[id]; !exists {
			rolesMap[id] = []database.Role{}
		}
	}

	return rolesMap, nil
}

// GetRolePermissionsMap gets all permissions for multiple roles in batch (optimized)
func GetRolePermissionsMap(roleIDs []int) (map[int][]database.Permission, error) {
	if len(roleIDs) == 0 {
		return make(map[int][]database.Permission), nil
	}

	var relations []database.RolePermission
	if err := database.DB.Preload("Permission").
		Where("role_id IN ?", roleIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get role permission relations: %v", err)
	}

	permissionsMap := make(map[int][]database.Permission)
	for _, relation := range relations {
		if relation.Permission != nil {
			permissionsMap[relation.RoleID] = append(permissionsMap[relation.RoleID], *relation.Permission)
		}
	}

	for _, id := range roleIDs {
		if _, exists := permissionsMap[id]; !exists {
			permissionsMap[id] = []database.Permission{}
		}
	}

	return permissionsMap, nil
}

// GetUserProjectsMap gets all projects for multiple users in batch (optimized)
func GetUserProjectsMap(userIDs []int) (map[int][]database.UserProject, error) {
	if len(userIDs) == 0 {
		return make(map[int][]database.UserProject), nil
	}

	var relations []database.UserProject
	if err := database.DB.Preload("Project").Preload("Role").
		Where("user_id IN ?", userIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get user project relations: %v", err)
	}

	projectsMap := make(map[int][]database.UserProject)
	for _, relation := range relations {
		projectsMap[relation.UserID] = append(projectsMap[relation.UserID], relation)
	}

	for _, id := range userIDs {
		if _, exists := projectsMap[id]; !exists {
			projectsMap[id] = []database.UserProject{}
		}
	}

	return projectsMap, nil
}

// Project-Container-Dataset Relationship Query Functions

// GetProjectContainersMap gets all containers for multiple projects in batch (optimized)
func GetProjectContainersMap(projectIDs []int) (map[int][]database.Container, error) {
	if len(projectIDs) == 0 {
		return make(map[int][]database.Container), nil
	}

	var relations []database.ProjectContainer
	if err := database.DB.Preload("Container").
		Where("project_id IN ?", projectIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get project container relations: %v", err)
	}

	containersMap := make(map[int][]database.Container)
	for _, relation := range relations {
		if relation.Container != nil {
			containersMap[relation.ProjectID] = append(containersMap[relation.ProjectID], *relation.Container)
		}
	}

	for _, id := range projectIDs {
		if _, exists := containersMap[id]; !exists {
			containersMap[id] = []database.Container{}
		}
	}

	return containersMap, nil
}

// GetProjectDatasetsMap gets all datasets for multiple projects in batch (optimized)
func GetProjectDatasetsMap(projectIDs []int) (map[int][]database.Dataset, error) {
	if len(projectIDs) == 0 {
		return make(map[int][]database.Dataset), nil
	}

	var relations []database.ProjectDataset
	if err := database.DB.Preload("Dataset").
		Where("project_id IN ?", projectIDs).
		Find(&relations).Error; err != nil {
		return nil, fmt.Errorf("failed to get project dataset relations: %v", err)
	}

	datasetsMap := make(map[int][]database.Dataset)
	for _, relation := range relations {
		if relation.Dataset != nil {
			datasetsMap[relation.ProjectID] = append(datasetsMap[relation.ProjectID], *relation.Dataset)
		}
	}

	for _, id := range projectIDs {
		if _, exists := datasetsMap[id]; !exists {
			datasetsMap[id] = []database.Dataset{}
		}
	}

	return datasetsMap, nil
}

func GetProjectInjetionMap(projectIDs []int) (map[int][]database.FaultInjectionSchedule, error) {
	if len(projectIDs) == 0 {
		return make(map[int][]database.FaultInjectionSchedule), nil
	}

	var injections []database.FaultInjectionSchedule
	if err := database.DB.
		Joins("JOIN tasks ON fault_injection_schedules.task_id = tasks.id").
		Where("tasks.project_id IN ?", projectIDs).
		Preload("Task").
		Find(&injections).Error; err != nil {
		return nil, fmt.Errorf("failed to get project fault injection relations: %v", err)
	}

	injectionsMap := make(map[int][]database.FaultInjectionSchedule)
	for _, injection := range injections {
		if injection.Task != nil {
			projectID := *injection.Task.ProjectID
			injectionsMap[projectID] = append(injectionsMap[projectID], injection)
		}
	}

	for _, id := range projectIDs {
		if _, exists := injectionsMap[id]; !exists {
			injectionsMap[id] = []database.FaultInjectionSchedule{}
		}
	}

	return injectionsMap, nil
}
