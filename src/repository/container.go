package repository

import (
	"fmt"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	containerOmitFields        = "Versions"
	containerVersionOmitFields = "active_version_key,HelmConfig,EnvVars"
	helmConfigOmitFields       = "Values"
)

type ParameterConfigFetcher func(db *gorm.DB, keys []string, resourceID int) ([]database.ParameterConfig, error)

// =====================================================================
// Container Repository Functions
// =====================================================================

// CreateContainer creates a new container record
func CreateContainer(db *gorm.DB, container *database.Container) error {
	if err := db.Omit(commonOmitFields, containerOmitFields).Create(container).Error; err != nil {
		return fmt.Errorf("failed to create container: %w", err)
	}
	return nil
}

// DeleteContainer soft deletes a container by setting its status to deleted
func DeleteContainer(db *gorm.DB, containerID int) (int64, error) {
	result := db.Model(&database.Container{}).
		Where("id = ? AND status != ?", containerID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to delete container %d: %w", containerID, err)
	}
	return result.RowsAffected, nil
}

// GetContainerByID retrieves a container by its ID
func GetContainerByID(db *gorm.DB, id int) (*database.Container, error) {
	var container database.Container
	if err := db.Where("id = ? AND status != ?", id, consts.CommonDeleted).First(&container).Error; err != nil {
		return nil, fmt.Errorf("failed to find container with id %d: %w", id, err)
	}
	return &container, nil
}

// GetContainerStatistics returns statistics about containers
func GetContainerStatistics() (map[string]int64, error) {
	stats := make(map[string]int64)

	// Total containers
	var total int64
	if err := database.DB.Model(&database.Container{}).Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count total containers: %v", err)
	}
	stats["total"] = total

	// Active containers
	var active int64
	if err := database.DB.Model(&database.Container{}).Where("status = 1").Count(&active).Error; err != nil {
		return nil, fmt.Errorf("failed to count active containers: %v", err)
	}
	stats["active"] = active

	// Disabled containers
	var disabled int64
	if err := database.DB.Model(&database.Container{}).Where("status = 0").Count(&disabled).Error; err != nil {
		return nil, fmt.Errorf("failed to count disabled containers: %v", err)
	}
	stats["disabled"] = disabled

	// Deleted containers
	var deleted int64
	if err := database.DB.Model(&database.Container{}).Where("status = -1").Count(&deleted).Error; err != nil {
		return nil, fmt.Errorf("failed to count deleted containers: %v", err)
	}
	stats["deleted"] = deleted

	return stats, nil
}

// ListContainers lists containers based on filter options
func ListContainers(db *gorm.DB, limit, offset int, contaierType *consts.ContainerType, isPublic *bool, status *consts.StatusType) ([]database.Container, int64, error) {
	var containers []database.Container
	var total int64

	query := db.Model(&database.Container{})
	if contaierType != nil {
		query = query.Where("type = ?", *contaierType)
	}
	if isPublic != nil {
		query = query.Where("is_public = ?", *isPublic)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count containers: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&containers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list containers: %w", err)
	}

	return containers, total, nil
}

// ListContainersByID retrieves multiple containers by their IDs
func ListContainersByID(tx *gorm.DB, containerIDs []int) ([]database.Container, error) {
	if len(containerIDs) == 0 {
		return []database.Container{}, nil
	}

	var containers []database.Container
	if err := tx.
		Where("id IN (?) AND status != ?", containerIDs, consts.CommonDeleted).
		Find(&containers).Error; err != nil {
		return nil, fmt.Errorf("failed to query containers: %w", err)
	}
	return containers, nil
}

// UpdateContainer updates a container
func UpdateContainer(db *gorm.DB, container *database.Container) error {
	if err := db.Omit(commonOmitFields).Save(container).Error; err != nil {
		return fmt.Errorf("failed to update container: %w", err)
	}
	return nil
}

// =====================================================================
// ContainerVersion Repository Functions
// =====================================================================

// BatchCreateContainerVersions creates multiple container versions
func BatchCreateContainerVersions(db *gorm.DB, versions []database.ContainerVersion) error {
	if len(versions) == 0 {
		return fmt.Errorf("no container versions to create")
	}

	if err := db.Omit(containerVersionOmitFields).Create(&versions).Error; err != nil {
		return fmt.Errorf("failed to batch create container versions: %w", err)
	}

	return nil
}

// BatchDeleteContainerVersions soft deletes all versions of a specific container
func BatchDeleteContainerVersions(db *gorm.DB, containerID int) (int64, error) {
	result := db.Model(&database.ContainerVersion{}).
		Where("container_id = ? AND status != ?", containerID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to batch soft delete container versions for container %d: %w", containerID, result.Error)
	}
	return result.RowsAffected, nil
}

// BatchGetContainerVersions retrieves container versions for multiple container names
func BatchGetContainerVersions(db *gorm.DB, containerType consts.ContainerType, containerNames []string, userID int) ([]database.ContainerVersion, error) {
	if len(containerNames) == 0 {
		return []database.ContainerVersion{}, nil
	}

	var versions []database.ContainerVersion

	query := db.Table("container_versions cv").
		Preload("Container").
		Where("cv.status = ?", consts.CommonEnabled).
		Order("cv.container_id DESC, cv.name_major DESC, cv.name_minor DESC, cv.name_patch DESC")

	query = query.Joins("INNER JOIN containers c ON c.id = cv.container_id").
		Where("c.type = ? AND c.name IN (?) AND c.status = ?", containerType, containerNames, consts.CommonEnabled)

	if userID > 0 {
		query = query.Joins(
			"LEFT JOIN user_containers uc ON uc.container_id = c.id AND uc.user_id = ? AND uc.status = ?",
			userID, consts.CommonEnabled,
		).Where(
			db.Where("c.is_public = ?", true).
				Or("uc.container_id IS NOT NULL"),
		)
	}

	if err := query.Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to query container versions: %w", err)
	}

	return versions, nil
}

// CheckContainerExistsWithDifferentType checks if a container exists with a different type
func CheckContainerExistsWithDifferentType(db *gorm.DB, containerName string, requestedType consts.ContainerType, userID int) (bool, consts.ContainerType, error) {
	var container database.Container

	query := db.Table("containers").
		Where("name = ? AND type != ? AND status = ?", containerName, requestedType, consts.CommonEnabled)

	if userID > 0 {
		query = query.Joins(
			"LEFT JOIN user_containers uc ON uc.container_id = containers.id AND uc.user_id = ? AND uc.status = ?",
			userID, consts.CommonEnabled,
		).Where(
			db.Where("containers.is_public = ?", true).
				Or("uc.container_id IS NOT NULL"),
		)
	}

	if err := query.First(&container).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, 0, nil
		}
		return false, 0, fmt.Errorf("failed to check container existence: %w", err)
	}

	return true, container.Type, nil
}

// DeleteContainerVersion soft deletes a container version
func DeleteContainerVersion(db *gorm.DB, versionID int) (int64, error) {
	result := db.Model(&database.ContainerVersion{}).
		Where("id = ? AND status != ?", versionID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to soft delete container version %d: %w", versionID, result.Error)
	}
	return result.RowsAffected, nil
}

// GetContainerVersionByID retrieves a ContainerVersion by its ID
func GetContainerVersionByID(db *gorm.DB, versionID int) (*database.ContainerVersion, error) {
	var version database.ContainerVersion
	if err := db.
		Preload("Container").
		Preload("HelmConfig").
		Where("id = ?", versionID).First(&version).Error; err != nil {
		return nil, fmt.Errorf("failed to find container version with id %d: %w", versionID, err)
	}
	return &version, nil
}

// ListContainerVersions lists container versions with pagination and optional status filtering
func ListContainerVersions(db *gorm.DB, limit, offset int, containerID int, status *consts.StatusType) ([]database.ContainerVersion, int64, error) {
	var versions []database.ContainerVersion
	var total int64

	query := db.Model(&database.ContainerVersion{}).Where("container_id = ?", containerID)
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count container versions: %v", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&versions).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list container versions: %v", err)
	}

	return versions, total, nil
}

// ListContainerVersions lists all versions of a specific container
func ListContainerVersionsByContainerID(db *gorm.DB, containerID int) ([]database.ContainerVersion, error) {
	var versions []database.ContainerVersion
	if err := db.
		Preload("Container").
		Preload("HelmConfig").
		Where("container_id = ?", containerID).
		Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to list container versions for container %d: %w", containerID, err)
	}
	return versions, nil
}

// UpdateContainerVersion updates a container version
func UpdateContainerVersion(db *gorm.DB, version *database.ContainerVersion) error {
	if err := db.Omit(containerVersionOmitFields).Save(version).Error; err != nil {
		return fmt.Errorf("failed to update container version: %w", err)
	}
	return nil
}

// =====================================================================
// HelmConfig Repository Functions
// =====================================================================

// BatchCreateHelmConfigs creates multiple helm configs
func BatchCreateHelmConfigs(db *gorm.DB, helmConfigs []database.HelmConfig) error {
	if len(helmConfigs) == 0 {
		return fmt.Errorf("no helm configs to create")
	}

	if err := db.Omit(helmConfigOmitFields).Create(helmConfigs).Error; err != nil {
		return fmt.Errorf("failed to batch create helm configs: %v", err)
	}

	return nil
}

// GetHelmConfigByContainerVersionID retrieves the HelmConfig associated with a specific ContainerVersion ID
func GetHelmConfigByContainerVersionID(db *gorm.DB, versionID int) (*database.HelmConfig, error) {
	var helmConfig database.HelmConfig
	if err := db.Preload("ContainerVersion").
		Where("container_version_id = ?", versionID).
		First(&helmConfig).Error; err != nil {
		return nil, fmt.Errorf("failed to find helm config for version id %d: %w", versionID, err)
	}
	return &helmConfig, nil
}

// UpdateHelmConfig updates a helm config
func UpdateHelmConfig(db *gorm.DB, helmConfig *database.HelmConfig) error {
	if err := db.Save(helmConfig).Error; err != nil {
		return fmt.Errorf("failed to update helm config: %w", err)
	}
	return nil
}

// =====================================================================
// ParameterConfig Repository Functions
// =====================================================================

// BatchCreateOrFindParameterConfigs creates multiple parameter configs or finds existing ones using upsert
func BatchCreateOrFindParameterConfigs(db *gorm.DB, params []database.ParameterConfig) error {
	if len(params) == 0 {
		return nil
	}

	if err := db.Clauses(clause.OnConflict{
		OnConstraint: "idx_unique_config",
		DoNothing:    true,
	}).Create(&params).Error; err != nil {
		return fmt.Errorf("failed to batch create parameter configs: %w", err)
	}
	return nil
}

// ListParameterConfigsByKeys retrieves ParameterConfigs by their keys, type and category
func ListParameterConfigsByKeys(db *gorm.DB, configs []database.ParameterConfig) ([]database.ParameterConfig, error) {
	if len(configs) == 0 {
		return []database.ParameterConfig{}, nil
	}

	// Build query conditions for batch lookup
	var results []database.ParameterConfig
	query := db.Model(&database.ParameterConfig{})

	// Build OR conditions for each config
	conditions := db.Where("1 = 0") // Start with false condition
	for _, cfg := range configs {
		conditions = conditions.Or(
			db.Where("config_key = ? AND type = ? AND category = ?", cfg.Key, cfg.Type, cfg.Category),
		)
	}

	if err := query.Where(conditions).Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to list parameter configs by keys: %w", err)
	}

	return results, nil
}

// =====================================================================
// ContainerLabel Repository Functions
// =====================================================================

// AddContainerLabels adds multiple container-label associations in a batch
func AddContainerLabels(db *gorm.DB, containerLabels []database.ContainerLabel) error {
	if len(containerLabels) == 0 {
		return nil
	}
	if err := db.Create(&containerLabels).Error; err != nil {
		return fmt.Errorf("failed to add container-label associations: %w", err)
	}
	return nil
}

// ClearContainerLabels removes label associations from specified containers
func ClearContainerLabels(db *gorm.DB, containerIDs []int, labelIDs []int) error {
	if len(containerIDs) == 0 {
		return nil
	}

	query := db.Table("container_labels").
		Where("container_id IN (?)", containerIDs)
	if len(labelIDs) > 0 {
		query = query.Where("label_id IN (?)", labelIDs)
	}

	if err := query.Delete(nil).Error; err != nil {
		return fmt.Errorf("failed to clear container-label associations: %w", err)
	}
	return nil
}

// RemoveLabelsFromContainer removes all label associations from a specific container
func RemoveLabelsFromContainer(db *gorm.DB, containerID int) error {
	if err := db.Where("container_id = ?", containerID).
		Delete(&database.ContainerLabel{}).Error; err != nil {
		return fmt.Errorf("failed to remove all labels from container %d: %w", containerID, err)
	}
	return nil
}

// RemoveContainersFromLabel removes all container associations from a specific label
func RemoveContainersFromLabel(db *gorm.DB, labelID int) (int64, error) {
	result := db.Where("label_id = ?", labelID).
		Delete(&database.ContainerLabel{})
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to remove all containers from label %d: %w", labelID, err)
	}
	return result.RowsAffected, nil
}

// RemoveContainersFromLabels removes all container associations from multiple labels
func RemoveContainersFromLabels(db *gorm.DB, labelIDs []int) (int64, error) {
	if len(labelIDs) == 0 {
		return 0, nil
	}

	result := db.Where("label_id IN (?)", labelIDs).
		Delete(&database.ContainerLabel{})
	if err := result.Error; err != nil {
		return 0, fmt.Errorf("failed to remove all containers from labels %v: %w", labelIDs, err)
	}
	return result.RowsAffected, nil
}

// ListContainerLabels gets labels for multiple containers in batch
func ListContainerLabels(db *gorm.DB, containerIDs []int) (map[int][]database.Label, error) {
	if len(containerIDs) == 0 {
		return nil, nil
	}

	type containerLabelResult struct {
		database.Label
		containerID int `gorm:"column:container_id"`
	}

	var flatResults []containerLabelResult
	if err := db.Model(&database.Label{}).
		Joins("JOIN container_labels cl ON cl.label_id = labels.id").
		Where("cl.container_id IN (?)", containerIDs).
		Select("labels.*, cl.container_id").
		Find(&flatResults).Error; err != nil {
		return nil, fmt.Errorf("failed to batch query container labels: %w", err)
	}

	labelsMap := make(map[int][]database.Label)
	for _, id := range containerIDs {
		labelsMap[id] = []database.Label{}
	}

	for _, res := range flatResults {
		label := res.Label
		labelsMap[res.containerID] = append(labelsMap[res.containerID], label)
	}

	return labelsMap, nil
}

// ListContainerLabelCounts retrieves the count of containers associated with each label ID
func ListContainerLabelCounts(db *gorm.DB, labelIDs []int) (map[int]int64, error) {
	if len(labelIDs) == 0 {
		return make(map[int]int64), nil
	}

	type containerLabelResult struct {
		labelID int `gorm:"column:label_id"`
		count   int64
	}

	var results []containerLabelResult
	if err := db.Model(&database.ContainerLabel{}).
		Select("label_id, count(label_id) as count").
		Where("label_id IN (?)", labelIDs).
		Group("label_id").
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to count associations: %w", err)
	}

	countMap := make(map[int]int64, len(results))
	for _, result := range results {
		countMap[result.labelID] = result.count
	}

	return countMap, nil
}

// ListLabelsByContainerID lists all labels associated with a specific container
func ListLabelsByContainerID(db *gorm.DB, containerID int) ([]database.Label, error) {
	var labels []database.Label
	if err := db.Model(&database.Label{}).
		Joins("JOIN container_labels cl ON cl.label_id = labels.id").
		Where("cl.container_id = ?", containerID).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list labels for container %d: %w", containerID, err)
	}
	return labels, nil
}

// ListLabelIDsByKeyAndContainerID finds label IDs by keys associated with a specific container
func ListLabelIDsByKeyAndContainerID(db *gorm.DB, containerID int, keys []string) ([]int, error) {
	var labelIDs []int

	err := db.Table("labels l").
		Select("l.id").
		Joins("JOIN container_labels cl ON cl.label_id = l.id").
		Where("cl.container_id = ? AND l.label_key IN (?)", containerID, keys).
		Pluck("l.id", &labelIDs).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find label IDs by keys for container %d: %w", containerID, err)
	}

	return labelIDs, nil
}

// =====================================================================
// ContainerVersionEnvVar Repository Functions
// =====================================================================

// AddContainerVersionEnvVars adds multiple environment variable parameters for a specific container version
func AddContainerVersionEnvVars(db *gorm.DB, envVars []database.ContainerVersionEnvVar) error {
	if len(envVars) == 0 {
		return nil
	}
	if err := db.Create(&envVars).Error; err != nil {
		return fmt.Errorf("failed to add container version env vars: %w", err)
	}
	return nil
}

// ListContainerEnvVars lists environment variable parameters for a specific container version
func ListContainerVersionEnvVars(db *gorm.DB, keys []string, containerVersionID int) ([]database.ParameterConfig, error) {
	query := db.Model(&database.ParameterConfig{}).
		Joins("JOIN container_version_env_vars cvev ON cvev.parameter_config_id = parameter_configs.id").
		Where("cvev.container_version_id = ?", containerVersionID).
		Where("parameter_configs.category = ?", consts.ParameterCategoryEnvVars)

	if len(keys) > 0 {
		query = query.Where("parameter_configs.config_key IN (?)", keys)
	}

	var params []database.ParameterConfig
	if err := query.Find(&params).Error; err != nil {
		return nil, fmt.Errorf("failed to list container env vars: %w", err)
	}
	return params, nil
}

// =====================================================================
// HelmConfigValues Repository Functions
// =====================================================================

// AddHelmConfigValues adds multiple helm value parameters for a specific helm config
func AddHelmConfigValues(db *gorm.DB, helmValues []database.HelmConfigValue) error {
	if len(helmValues) == 0 {
		return nil
	}
	if err := db.Create(&helmValues).Error; err != nil {
		return fmt.Errorf("failed to add helm config values: %w", err)
	}
	return nil
}

// ListHelmConfigValues lists helm value parameters for a specific helm config
func ListHelmConfigValues(db *gorm.DB, keys []string, helmConfigID int) ([]database.ParameterConfig, error) {
	query := db.Model(&database.ParameterConfig{}).
		Joins("JOIN helm_config_values hcv ON hcv.parameter_config_id = parameter_configs.id").
		Where("hcv.helm_config_id = ?", helmConfigID)

	if len(keys) > 0 {
		query = query.Where("parameter_configs.config_key IN (?)", keys)
	}

	var params []database.ParameterConfig
	if err := query.Find(&params).Error; err != nil {
		return nil, fmt.Errorf("failed to list helm values: %w", err)
	}
	return params, nil
}
