package repository

import (
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/database"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// BatchUpsertResources upserts multiple resources
func BatchUpsertResources(db *gorm.DB, resources []database.Resource) error {
	if len(resources) == 0 {
		return fmt.Errorf("no resources to upsert")
	}

	if err := db.Omit(commonOmitFields).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}},
		DoNothing: true,
	}).Create(&resources).Error; err != nil {
		return fmt.Errorf("failed to batch upsert resources: %w", err)
	}

	return nil
}

// CreateResource creates a resource
func CreateResource(db *gorm.DB, resource *database.Resource) error {
	return createModel(db.Omit(commonOmitFields), resource)
}

// GetResourceByID gets resource by ID
func GetResourceByID(db *gorm.DB, id int) (*database.Resource, error) {
	return findModel[database.Resource](db, "id = ? and status != ?", id, consts.CommonDeleted)
}

// GetResourceByName gets resource by name
func GetResourceByName(db *gorm.DB, name string) (*database.Resource, error) {
	return findModel[database.Resource](db, "name = ? and status != ?", name, consts.CommonDeleted)
}

// UpdateResource updates resource information
func UpdateResource(db *gorm.DB, resource *database.Resource) error {
	return updateModel(db.Omit(commonOmitFields), resource)
}

// DeleteResource soft deletes resource (sets status to -1)
func DeleteResource(id int) error {
	if err := database.DB.Model(&database.Resource{}).Where("id = ?", id).Update("status", -1).Error; err != nil {
		return fmt.Errorf("failed to delete resource: %v", err)
	}
	return nil
}

// ListResources gets resource list
func ListResources(page, pageSize int, resourceType string, category string, parentID *int, status *int) ([]database.Resource, int64, error) {
	var resources []database.Resource
	var total int64

	query := database.DB.Model(&database.Resource{}).Preload("Parent")

	if status != nil {
		query = query.Where("status = ?", *status)
	}

	if resourceType != "" {
		query = query.Where("type = ?", resourceType)
	}

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if parentID != nil {
		query = query.Where("parent_id = ?", *parentID)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count resources: %v", err)
	}

	// Paginated query
	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("name").Find(&resources).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list resources: %v", err)
	}

	return resources, total, nil
}

// GetResourcesByNamesWithTx gets resources by names within a transaction
func GetResourcesByNamesWithTx(tx *gorm.DB, names []string) ([]database.Resource, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("no resource names provided")
	}

	var allResourcesInDB []database.Resource
	if err := tx.Model(&database.Resource{}).
		Select("id, name").
		Where("name IN ?", names).
		Find(&allResourcesInDB).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch resources by names: %v", err)
	}

	return allResourcesInDB, nil
}

// GetResourcesByType gets resources by type
func GetResourcesByType(resourceType string) ([]database.Resource, error) {
	var resources []database.Resource
	if err := database.DB.Where("type = ? AND status = 1", resourceType).
		Order("name").Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("failed to get resources by type: %v", err)
	}
	return resources, nil
}

// GetResourcesByCategory gets resources by category
func GetResourcesByCategory(category string) ([]database.Resource, error) {
	var resources []database.Resource
	if err := database.DB.Where("category = ? AND status = 1", category).
		Order("name").Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("failed to get resources by category: %v", err)
	}
	return resources, nil
}

// GetChildResources gets child resources
func GetChildResources(parentID int) ([]database.Resource, error) {
	var resources []database.Resource
	if err := database.DB.Where("parent_id = ? AND status = 1", parentID).
		Order("name").Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("failed to get child resources: %v", err)
	}
	return resources, nil
}

// GetRootResources gets root-level resources (no parent resource)
func GetRootResources() ([]database.Resource, error) {
	var resources []database.Resource
	if err := database.DB.Where("parent_id IS NULL AND status = 1").
		Order("name").Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("failed to get root resources: %v", err)
	}
	return resources, nil
}

// GetResourceTree gets resource tree structure
func GetResourceTree() ([]database.Resource, error) {
	var allResources []database.Resource
	if err := database.DB.Where("status = 1").Order("parent_id, name").Find(&allResources).Error; err != nil {
		return nil, fmt.Errorf("failed to get resource tree: %v", err)
	}
	return allResources, nil
}

// GetSystemResources gets system resources
func GetSystemResources() ([]database.Resource, error) {
	var resources []database.Resource
	if err := database.DB.Where("is_system = true AND status = 1").
		Order("category, name").Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("failed to get system resources: %v", err)
	}
	return resources, nil
}

// SearchResources searches resources
func SearchResources(keyword string, resourceType string, category string) ([]database.Resource, error) {
	var resources []database.Resource

	query := database.DB.Model(&database.Resource{}).Where("status = 1")

	if keyword != "" {
		query = query.Where("name ILIKE ? OR display_name ILIKE ? OR description ILIKE ?",
			"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%")
	}

	if resourceType != "" {
		query = query.Where("type = ?", resourceType)
	}

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Order("name").Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("failed to search resources: %v", err)
	}

	return resources, nil
}

// GetResourcePermissions gets resource permissions
func GetResourcePermissions(resourceID int) ([]database.Permission, error) {
	var permissions []database.Permission
	if err := database.DB.Where("resource_id = ? AND status = 1", resourceID).
		Order("action").Find(&permissions).Error; err != nil {
		return nil, fmt.Errorf("failed to get resource permissions: %v", err)
	}
	return permissions, nil
}

// GetOrCreateResource gets or creates resource
func GetOrCreateResource(name, displayName, resourceType, category string, parentID *int) (*database.Resource, error) {
	var resource database.Resource

	// First try to get
	if err := database.DB.Where("name = ?", name).First(&resource).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create if not exists
			resource = database.Resource{
				Name:        name,
				DisplayName: displayName,
				Type:        resourceType,
				Category:    category,
				ParentID:    parentID,
				Status:      1,
			}
			if err := database.DB.Create(&resource).Error; err != nil {
				return nil, fmt.Errorf("failed to create resource: %v", err)
			}
		} else {
			return nil, fmt.Errorf("failed to get resource: %v", err)
		}
	}

	return &resource, nil
}

// CheckResourceExists checks if resource exists
func CheckResourceExists(name string) (bool, error) {
	var count int64
	if err := database.DB.Model(&database.Resource{}).Where("name = ? AND status = 1", name).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check resource existence: %v", err)
	}
	return count > 0, nil
}

// GetResourceHierarchy gets complete hierarchy path of resource
func GetResourceHierarchy(resourceID int) ([]database.Resource, error) {
	var hierarchy []database.Resource
	var currentID *int = &resourceID

	for currentID != nil {
		var resource database.Resource
		if err := database.DB.First(&resource, *currentID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				break
			}
			return nil, fmt.Errorf("failed to get resource hierarchy: %v", err)
		}

		hierarchy = append([]database.Resource{resource}, hierarchy...) // Prepend insert
		currentID = resource.ParentID
	}

	return hierarchy, nil
}
