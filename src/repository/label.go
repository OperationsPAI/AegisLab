package repository

import (
	"errors"
	"fmt"

	"aegis/consts"
	"aegis/database"
	"aegis/dto"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// =====================================================================
// Label Repository Functions
// =====================================================================

// BatchDeleteLabels marks multiple labels as deleted in batch
func BatchDeleteLabels(db *gorm.DB, labelIDs []int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	if err := db.Model(&database.Label{}).
		Where("id IN (?) AND status != ?", labelIDs, consts.CommonDeleted).
		Update("status", consts.CommonDeleted).Error; err != nil {
		return fmt.Errorf("failed to batch delete labels: %w", err)
	}
	return nil
}

// BatchIncreaseLabelUsages increases the usage counts of multiple labels
func BatchIncreaseLabelUsages(db *gorm.DB, labelIDs []int, increament int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	expr := fmt.Sprintf("usage_count + %d", increament)
	if err := db.Model(&database.Label{}).
		Where("id IN (?)", labelIDs).
		Clauses(clause.Returning{}).
		UpdateColumn("usage_count", expr).Error; err != nil {
		return fmt.Errorf("failed to batch decrease label usages: %w", err)
	}
	return nil
}

// BatchDecreaseLabelUsages decreases the usage counts of multiple labels
func BatchDecreaseLabelUsages(db *gorm.DB, labelIDs []int, decrement int) error {
	if len(labelIDs) == 0 {
		return nil
	}

	expr := fmt.Sprintf("GREATEST(0, usage_count - %d)", decrement)
	if err := db.Model(&database.Label{}).
		Where("id IN (?)", labelIDs).
		Clauses(clause.Returning{}).
		UpdateColumn("usage_count", expr).Error; err != nil {
		return fmt.Errorf("failed to batch decrease label usages: %w", err)
	}
	return nil
}

// BatchUpsertLabels upserts multiple labels
func BatchUpsertLabels(db *gorm.DB, labels []database.Label) error {
	if len(labels) == 0 {
		return fmt.Errorf("no labels to upsert")
	}

	if err := db.Clauses(clause.OnConflict{
		OnConstraint: "idx_key_value_unique",
		DoNothing:    true,
	}).Create(&labels).Error; err != nil {
		return fmt.Errorf("failed to batch upsert labels: %w", err)
	}

	return nil
}

// BatchUpdateLabels updates multiple labels
func BatchUpdateLabels(db *gorm.DB, labels []database.Label) error {
	if len(labels) == 0 {
		return fmt.Errorf("no labels to update")
	}

	if err := db.Save(&labels).Error; err != nil {
		return fmt.Errorf("failed to batch update labels: %w", err)
	}

	return nil
}

// CreateLabel creates a label
func CreateLabel(db *gorm.DB, label *database.Label) error {
	if err := db.Create(label).Error; err != nil {
		return fmt.Errorf("failed to create label: %w", err)
	}
	return nil
}

// DeleteLabel soft deletes a label by setting its status to deleted
func DeleteLabel(db *gorm.DB, labelID int) (int64, error) {
	result := db.Model(&database.Label{}).
		Where("id = ? AND status != ?", labelID, consts.CommonDeleted).
		Update("status", consts.CommonDeleted)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to soft delete project %d: %w", labelID, result.Error)
	}
	return result.RowsAffected, nil
}

// GetLabelByID gets label by ID
func GetLabelByID(db *gorm.DB, id int) (*database.Label, error) {
	var label database.Label
	if err := db.First(&label, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("label with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to get label: %w", err)
	}
	return &label, nil
}

// GetLabelByKeyAndValue gets label by key and value
func GetLabelByKeyAndValue(db *gorm.DB, key, value string, status ...consts.StatusType) (*database.Label, error) {
	query := db.Where("label_key = ? AND label_value = ?", key, value)

	if len(status) == 0 {
		query = query.Where("status != ?", consts.CommonDeleted)
	} else if len(status) == 1 {
		query = query.Where("status = ?", status[0])
	} else {
		query = query.Where("status IN (?)", status)
	}

	var label database.Label
	if err := query.First(&label).Error; err != nil {
		return nil, fmt.Errorf("failed to get label: %w", err)
	}

	return &label, nil
}

// ListLabels gets the label list
func ListLabels(db *gorm.DB, limit, offset int, filterOptions *dto.ListLabelFilters) ([]database.Label, int64, error) {
	var labels []database.Label
	var total int64

	query := db.Model(&database.Label{})
	if filterOptions.Key != "" {
		query = query.Where("label_key = ?", filterOptions.Key)
	}
	if filterOptions.Value != "" {
		query = query.Where("label_value = ?", filterOptions.Value)
	}
	if filterOptions.Category != nil {
		query = query.Where("category = ?", *filterOptions.Category)
	}
	if filterOptions.IsSystem != nil {
		query = query.Where("is_system = ?", *filterOptions.IsSystem)
	}
	if filterOptions.Status != nil {
		query = query.Where("status = ?", *filterOptions.Status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count labels: %w", err)
	}

	if err := query.Limit(limit).Offset(offset).Order("usage_count DESC, created_at DESC").Find(&labels).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to list labels: %w", err)
	}

	return labels, total, nil
}

// ListLabelsByID lists labels by their IDs
func ListLabelsByID(db *gorm.DB, labelIDs []int) ([]database.Label, error) {
	if len(labelIDs) == 0 {
		return []database.Label{}, nil
	}

	var labels []database.Label
	if err := db.
		Where("id IN (?) AND status != ?", labelIDs, consts.CommonDeleted).
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list labels by IDs: %w", err)
	}
	return labels, nil
}

// ListLabelsGroupByCategory lists labels grouped by their categories
func ListLabelsGroupByCategory(db *gorm.DB) (map[consts.LabelCategory][]database.Label, error) {
	var labels []database.Label
	if err := db.
		Where("status != ?", consts.CommonDeleted).
		Order("usage_count DESC, created_at DESC").
		Find(&labels).Error; err != nil {
		return nil, fmt.Errorf("failed to list labels: %w", err)
	}

	groupedLabels := make(map[consts.LabelCategory][]database.Label)
	for _, label := range labels {
		groupedLabels[label.Category] = append(groupedLabels[label.Category], label)
	}

	return groupedLabels, nil
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
		return nil, fmt.Errorf("failed to search labels: %w", err)
	}

	return labels, nil
}

func UpdateLabel(db *gorm.DB, label *database.Label) error {
	if err := db.Save(label).Error; err != nil {
		return fmt.Errorf("failed to update label: %w", err)
	}
	return nil
}
