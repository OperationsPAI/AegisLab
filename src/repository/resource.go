package repository

import (
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

// GetResourceByID gets resource by ID
func GetResourceByID(db *gorm.DB, id int) (*database.Resource, error) {
	var resource database.Resource
	if err := db.Where("id = ? and status != ?", id, consts.CommonDeleted).First(&resource).Error; err != nil {
		return nil, fmt.Errorf("failed to find resource with id %d: %w", id, err)
	}
	return &resource, nil
}

// GetResourceByName gets resource by name
func GetResourceByName(db *gorm.DB, name consts.ResourceName) (*database.Resource, error) {
	var resource database.Resource
	if err := db.
		Where("name = ? and status != ?", name, consts.CommonDeleted).
		First(&resource).Error; err != nil {
		return nil, fmt.Errorf("failed to find resource with name %s: %w", name, err)
	}
	return &resource, nil
}

// ListResources gets resource list
func ListResources(db *gorm.DB, limit, offset int, resourceType *consts.ResourceType, category *consts.ResourceCategory) ([]database.Resource, int64, error) {
	var resources []database.Resource
	var total int64

	query := database.DB.Model(&database.Resource{}).Preload("Parent")
	if resourceType != nil {
		query = query.Where("type = ?", resourceType)
	}
	if category != nil {
		query = query.Where("category = ?", category)
	}

	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count resources: %v", err)
	}

	if err := query.Limit(limit).Offset(offset).Find(&resources).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list resources: %v", err)
	}

	return resources, total, nil
}

// ListResourcesByNames lists resources by names
func ListResourcesByNames(db *gorm.DB, names []consts.ResourceName) ([]database.Resource, error) {
	if len(names) == 0 {
		return nil, fmt.Errorf("no resource names provided")
	}

	var resources []database.Resource
	if err := db.Where("name IN (?)", names).
		Find(&resources).Error; err != nil {
		return nil, fmt.Errorf("failed to list resources by names: %v", err)
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
