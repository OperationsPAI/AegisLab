package repository

import (
	"fmt"

	"rcabench/database"
)

func CheckGranularityResultsByExecutionID(executionID int) (bool, error) {
	var count int64
	if err := database.DB.Model(&database.GranularityResult{}).
		Where("execution_id = ?", executionID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check granularity result existence: %v", err)
	}

	return count > 0, nil
}

func ListGranularityResultsByExecutionID(executionID int) ([]database.GranularityResult, error) {
	var results []database.GranularityResult
	if err := database.DB.Model(&database.GranularityResult{}).
		Where("execution_id = ?", executionID).
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to list granularity results by execution ID: %v", err)
	}

	return results, nil
}

func SaveGranularityResults(results []database.GranularityResult) error {
	if len(results) == 0 {
		return fmt.Errorf("no granularity results to save")
	}

	return database.DB.Create(&results).Error
}
