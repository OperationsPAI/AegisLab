package common

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/repository"
	"aegis/utils"
	"fmt"
	"sort"

	"gorm.io/gorm"
)

// ConvertLabelFiltersToConditions converts a slice of LabelFilter to a slice of map conditions
func ConvertLabelFiltersToConditions(labelItems []dto.LabelItem) []map[string]string {
	if len(labelItems) == 0 {
		return []map[string]string{}
	}

	labelConditions := make([]map[string]string, 0, len(labelItems))
	for _, label := range labelItems {
		labelConditions = append(labelConditions, map[string]string{
			"key":   label.Key,
			"value": label.Value,
		})
	}

	return labelConditions
}

// CreateOrUpdateLabelsFromItems creates or updates labels based on the provided label items
// Returns labels with correct IDs and updates usage_count for existing labels
func CreateOrUpdateLabelsFromItems(db *gorm.DB, labelItems []dto.LabelItem, category consts.LabelCategory) ([]model.Label, error) {
	if len(labelItems) == 0 {
		return []model.Label{}, nil
	}

	// Build key -> value map for quick lookup
	kvMap := make(map[string]dto.LabelItem, len(labelItems))
	for _, item := range labelItems {
		kvMap[item.Key] = item
	}

	// Find existing labels using slice conditions for repository
	labelConditions := dto.ConvertLabelItemsToConditions(labelItems)
	existingLabels, err := repository.ListLabelsByConditions(db, labelConditions)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing labels: %w", err)
	}

	// Separate existing and new labels
	result := make([]model.Label, 0, len(labelItems))
	existingIDs := make([]int, 0, len(existingLabels))
	for _, existing := range existingLabels {
		if item, ok := kvMap[existing.Key]; ok && item.Value == existing.Value {
			result = append(result, existing)
			existingIDs = append(existingIDs, existing.ID)
			delete(kvMap, existing.Key)
		}
	}

	// Increase usage count for existing labels
	if len(existingIDs) > 0 {
		if err := repository.BatchIncreaseLabelUsages(db, existingIDs, 1); err != nil {
			return nil, fmt.Errorf("failed to increase usage for existing labels: %w", err)
		}
	}

	// Create new labels (only those not found in existing)
	if len(kvMap) > 0 {
		newLabels := make([]model.Label, 0, len(kvMap))

		for key, item := range kvMap {
			newLabels = append(newLabels, model.Label{
				Key:         key,
				Value:       item.Value,
				Category:    category,
				Description: fmt.Sprintf(consts.CustomLabelDescriptionTemplate, key, consts.GetLabelCategoryName(category)),
				Color:       utils.GenerateColorFromKey(key),
				Usage:       consts.DefaultLabelUsage,
				IsSystem:    item.IsSystem,
				Status:      consts.CommonEnabled,
			})
		}

		if err := repository.BatchCreateLabels(db, newLabels); err != nil {
			return nil, fmt.Errorf("failed to create new labels: %w", err)
		}

		result = append(result, newLabels...)
	}

	// Sort by ID ascending
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

func GetLabelConditionsByItems(labelItems []dto.LabelItem) []map[string]string {
	labelConditions := make([]map[string]string, 0, len(labelItems))
	for _, item := range labelItems {
		labelConditions = append(labelConditions, map[string]string{
			"key":   item.Key,
			"value": item.Value,
		})
	}
	return labelConditions
}
