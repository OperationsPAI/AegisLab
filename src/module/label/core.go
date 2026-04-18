package labelmodule

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	"aegis/utils"
	"errors"
	"fmt"
	"sort"

	"gorm.io/gorm"
)

func (r *Repository) CreateLabelCore(db *gorm.DB, label *model.Label) (*model.Label, error) {
	query := r.useDB(db).Where("label_key = ? AND label_value = ?", label.Key, label.Value).
		Where("status != ?", consts.CommonDeleted)

	var existingLabel model.Label
	err := query.First(&existingLabel).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing label: %w", err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := r.useDB(db).Omit(labelKeyOmitFields).Create(label).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return nil, fmt.Errorf("%w: label with key %s and value %s already exists", consts.ErrAlreadyExists, label.Key, label.Value)
			}
			return nil, fmt.Errorf("failed to create label: %w", err)
		}
		return label, nil
	}

	existingLabel.Category = label.Category
	existingLabel.Description = label.Description
	existingLabel.Color = label.Color
	existingLabel.Status = consts.CommonEnabled
	if err := r.useDB(db).Omit(labelKeyOmitFields).Save(&existingLabel).Error; err != nil {
		return nil, fmt.Errorf("failed to update existing label: %w", err)
	}
	return &existingLabel, nil
}

func (r *Repository) CreateOrUpdateLabelsFromItems(db *gorm.DB, labelItems []dto.LabelItem, category consts.LabelCategory) ([]model.Label, error) {
	if len(labelItems) == 0 {
		return []model.Label{}, nil
	}

	repo := r
	if db != nil {
		repo = NewRepository(db)
	}
	kvMap := make(map[string]dto.LabelItem, len(labelItems))
	for _, item := range labelItems {
		kvMap[item.Key] = item
	}

	existingLabels, err := repo.listLabelsByConditions(dto.ConvertLabelItemsToConditions(labelItems))
	if err != nil {
		return nil, fmt.Errorf("failed to find existing labels: %w", err)
	}

	result := make([]model.Label, 0, len(labelItems))
	existingIDs := make([]int, 0, len(existingLabels))
	for _, existing := range existingLabels {
		if item, ok := kvMap[existing.Key]; ok && item.Value == existing.Value {
			result = append(result, existing)
			existingIDs = append(existingIDs, existing.ID)
			delete(kvMap, existing.Key)
		}
	}

	if len(existingIDs) > 0 {
		if err := repo.batchIncreaseLabelUsages(existingIDs, 1); err != nil {
			return nil, fmt.Errorf("failed to increase usage for existing labels: %w", err)
		}
	}

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
		if err := repo.batchCreateLabels(newLabels); err != nil {
			return nil, fmt.Errorf("failed to create new labels: %w", err)
		}
		result = append(result, newLabels...)
	}

	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result, nil
}
