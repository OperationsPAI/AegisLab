package repository

import (
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
)

const (
	containerVersionOmitFields = "active_version_key"
)

// BatchCreateContainerVersions creates multiple container versions
func BatchCreateContainerVersions(db *gorm.DB, versions []database.ContainerVersion) error {
	if len(versions) == 0 {
		return fmt.Errorf("no container versions to create")
	}

	for i := range versions {
		versionPtr := &versions[i]
		err := db.Omit(containerVersionOmitFields).Create(versionPtr).Error
		if err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: index %d (version: %s)", consts.ErrAlreadyExists, i, versions[i].Name)
			}
			return fmt.Errorf("failed to create record index %d: %w", i, err)
		}
	}

	return nil
}

// BatchCreateHelmConfigs creates multiple helm configs
func BatchCreateHelmConfigs(db *gorm.DB, helmConfigs []database.HelmConfig) error {
	if len(helmConfigs) == 0 {
		return fmt.Errorf("no helm configs to create")
	}

	if err := db.Create(helmConfigs).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return fmt.Errorf("helm config already exists: %v", consts.ErrAlreadyExists)
		}
		return fmt.Errorf("failed to batch create helm configs: %v", err)
	}

	return nil
}

// BatchSoftDeleteContainerVersions soft-deletes all versions of a container by setting their status to deleted
func BatchSoftDeleteContainerVersions(db *gorm.DB, containerID int) error {
	var versionIDs []int
	if err := db.Model(&database.ContainerVersion{}).
		Where("container_id = ?", containerID).
		Select("id").
		Find(&versionIDs).Error; err != nil {
		return fmt.Errorf("failed to retrieve version IDs for container ID %d: %w", containerID, err)
	}

	if len(versionIDs) == 0 {
		return nil
	}

	if err := db.Model(&database.ContainerVersion{}).
		Where("id IN (?)", versionIDs).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to batch soft-delete versions for container ID %d: %w", containerID, err)
	}

	return nil
}

// BatchSoftDeleteHelmConfigsByVersionIDs soft-deletes helm configs associated with the given version IDs
func BatchSoftDeleteHelmConfigsByVersionIDs(db *gorm.DB, versionIDs []int) error {
	if len(versionIDs) == 0 {
		return nil
	}

	if err := db.Where("container_version_id IN (?)", versionIDs).
		Delete(&database.HelmConfig{}).Error; err != nil {
		return fmt.Errorf("failed to batch delete helm configs by version IDs: %w", err)
	}

	return nil
}

// CheckContainerExists checks if a container with the given ID exists
func CheckContainerExists(id int) (bool, error) {
	var container database.Container
	if err := database.DB.
		Where("id = ?", id).
		First(&container).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("failed to check container: %v", err)
	}

	return true, nil
}

// CreateContainer creates a new container record
func CreateContainer(db *gorm.DB, container *database.Container) error {
	return createModel(db.Omit(commonOmitFields), container)
}

// CreateContainerVersion creates a new container version record
func CreateContainerVersion(db *gorm.DB, version *database.ContainerVersion) error {
	return createModel(db.Omit(containerVersionOmitFields), version)
}

// CreateHelmConfig creates a new helm config record
func CreateHelmConfig(db *gorm.DB, helmConfig *database.HelmConfig) error {
	return createModel(db, helmConfig)
}

func GetContainerByID(db *gorm.DB, id int) (*database.Container, error) {
	return findModel[database.Container](db, "id = ? and status != ?", id, consts.CommonDeleted)
}

func GetContainerByName(db *gorm.DB, name string) (*database.Container, error) {
	return findModel[database.Container](db, "name = ? and status != ?", name, consts.CommonDeleted)
}

// GetContainerCountByType returns count of containers grouped by type
func GetContainerCountByType() (map[string]int64, error) {
	type TypeCount struct {
		Type  string `json:"type"`
		Count int64  `json:"count"`
	}

	var results []TypeCount
	err := database.DB.Model(&database.Container{}).
		Select("type, COUNT(*) as count").
		Where("status != -1").
		Group("type").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to count containers by type: %v", err)
	}

	typeCounts := make(map[string]int64)
	for _, result := range results {
		typeCounts[result.Type] = result.Count
	}

	return typeCounts, nil
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

// GetContainerVersion retrieves a specific container version based on type, name, user ID, and requested version
func GetContainerVersion(containerType consts.ContainerType, containerName string, userID int, requestedVersion string) (*database.ContainerVersion, error) {
	var versions []database.ContainerVersion
	if err := database.DB.Model(&database.ContainerVersion{}).
		Preload("Container").
		Preload("HelmConfig").
		Joins("JOIN containers ON containers.id = container_versions.container_id").
		Where("containers.type = ? AND containers.name = ? AND containers.userID = ? AND containers.status = ? AND container_versions.status = ?",
			containerType, containerName, userID, consts.CommonEnabled, consts.CommonEnabled).
		Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to query container versions: %v", err)
	}

	existingVersions := make(map[string]database.ContainerVersion)
	for _, v := range versions {
		existingVersions[v.Name] = v
	}

	var selectedVersion *database.ContainerVersion
	if requestedVersion != "" {
		if version, exists := existingVersions[requestedVersion]; exists {
			selectedVersion = &version
		} else {
			return nil, fmt.Errorf("requested version '%s' not found for container '%s'", requestedVersion, containerName)
		}
	} else {
		selectedVersion = &versions[0]
	}

	return selectedVersion, nil
}

// GetContainerVersionByID retrieves a ContainerVersion by its ID
func GetContainerVersionByID(db *gorm.DB, versionID int) (*database.ContainerVersion, error) {
	return findModel[database.ContainerVersion](db, "id = ?", versionID)
}

func GetContainerVersionsByContainerID(db *gorm.DB, containerID int) ([]database.ContainerVersion, error) {
	var versions []database.ContainerVersion
	if err := db.
		Where("container_id = ?", containerID).
		Find(&versions).Error; err != nil {
		return nil, fmt.Errorf("failed to retrieve container versions for container ID %d: %w", containerID, err)
	}
	return versions, nil
}

// GetHelmConfigByContainerVersionID retrieves the HelmConfig associated with a specific ContainerVersion ID
func GetHelmConfigByContainerVersionID(db *gorm.DB, versionID int) (*database.HelmConfig, error) {
	return findModel[database.HelmConfig](db, "container_version_id = ?", versionID)
}

// ListContainers lists containers based on filter options
func ListContainers(db *gorm.DB, limit, offset int, contaierType consts.ContainerType, status *int) ([]database.Container, int64, error) {
	var containers []database.Container
	var total int64

	query := db.Model(&database.Role{})
	if contaierType != "" {
		query = query.Where("type = ?", string(contaierType))
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count containers: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("updated_at DESC").Find(&containers).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list containers: %w", err)
	}

	return containers, total, nil
}

// UpdateContainer updates a container
func UpdateContainer(db *gorm.DB, container *database.Container) error {
	return updateModel(db.Omit(commonOmitFields), container)
}

// UpdateContainerVersion updates a container version
func UpdateContainerVersion(db *gorm.DB, version *database.ContainerVersion) error {
	return updateModel(db.Omit(containerVersionOmitFields), version)
}

// UpdateHelmConfig updates a helm config
func UpdateHelmConfig(db *gorm.DB, helmConfig *database.HelmConfig) error {
	return updateModel(db, helmConfig)
}
