package repository

import (
	"aegis/database"
	"fmt"

	"gorm.io/gorm"
)

func CreateHelmConfig(helmConfig *database.HelmConfig) error {
	if err := database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(helmConfig).Error; err != nil {
			return fmt.Errorf("failed to create helm config: %v", err)
		}

		return nil
	}); err != nil {
		return err
	}

	return nil
}
