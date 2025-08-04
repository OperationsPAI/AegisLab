package repository

import (
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
)

func ListDetectorResultsByExecutionID(executionID int) ([]database.Detector, error) {
	var results []database.Detector
	if err := database.DB.Model(&database.Detector{}).
		Where("execution_id = ?", executionID).
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to check detector result by execution ID: %v", err)
	}

	return results, nil
}
