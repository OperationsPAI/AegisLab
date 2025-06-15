package repository

import (
	"errors"
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"gorm.io/gorm"
)

func GetAlgorithmImageInfo(name string) (string, string, error) {
	if name == "" {
		return "", "", fmt.Errorf("algorithm name cannot be empty")
	}

	var record database.Algorithm
	if err := database.DB.
		Where("name = ? AND status = ?", name, true).
		Order("created_at DESC").
		First(&record).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", "", fmt.Errorf("algorithm '%s' not found", name)
		}
		return "", "", fmt.Errorf("failed to query algorithm: %w", err)
	}

	return record.Image, record.Tag, nil
}

func ListAllAlgorithms() ([]database.Algorithm, error) {
	var algorithms []database.Algorithm
	if err := database.DB.
		Order("created_at DESC").
		Find(&algorithms).Error; err != nil {
		return nil, fmt.Errorf("failed to list all algorithms: %w", err)
	}

	return algorithms, nil
}
