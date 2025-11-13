package common

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/utils"
	"fmt"

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
func CreateOrUpdateLabelsFromItems(db *gorm.DB, labelItems []dto.LabelItem, category consts.LabelCategory) ([]database.Label, error) {
	if len(labelItems) == 0 {
		return []database.Label{}, nil
	}

	labels := make([]database.Label, 0, len(labelItems))
	for _, item := range labelItems {
		labels = append(labels, database.Label{
			Key:         item.Key,
			Value:       item.Value,
			Category:    category,
			Description: fmt.Sprintf(consts.CustomLabelDescriptionTemplate, item.Key, consts.GetLabelCategoryName(category)),
			Color:       utils.GenerateColorFromKey(item.Key),
			Usage:       consts.DefaultLabelUsage,
			Status:      consts.CommonEnabled,
		})
	}

	err := repository.BatchUpsertLabels(db, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to batch upsert labels: %w", err)
	}

	return labels, nil
}
