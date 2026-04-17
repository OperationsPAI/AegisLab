package labelmodule

import (
	"aegis/consts"
	"aegis/model"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

func CreateLabelCore(db *gorm.DB, label *model.Label) (*model.Label, error) {
	query := db.Where("label_key = ? AND label_value = ?", label.Key, label.Value).
		Where("status != ?", consts.CommonDeleted)

	var existingLabel model.Label
	err := query.First(&existingLabel).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing label: %w", err)
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if err := db.Omit(labelKeyOmitFields).Create(label).Error; err != nil {
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
	if err := db.Omit(labelKeyOmitFields).Save(&existingLabel).Error; err != nil {
		return nil, fmt.Errorf("failed to update existing label: %w", err)
	}
	return &existingLabel, nil
}
