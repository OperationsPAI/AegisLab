package repository

import (
	"fmt"

	"aegis/database"

	"gorm.io/gorm"
)

// ListDetectorResultsByExecutionID lists detector results for a specific execution ID
func ListDetectorResultsByExecutionID(db *gorm.DB, executionID int) ([]database.DetectorResult, error) {
	var results []database.DetectorResult
	if err := db.
		Where("execution_id = ?", executionID).
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to list detectors for execution %d: %w", executionID, err)
	}
	return results, nil
}

// SaveDetectorResults saves multiple detector results
func SaveDetectorResults(db *gorm.DB, results []database.DetectorResult) error {
	if len(results) == 0 {
		return fmt.Errorf("no detector results to save")
	}

	if err := db.Create(&results).Error; err != nil {
		return fmt.Errorf("failed to save detector results: %w", err)
	}

	return nil
}
