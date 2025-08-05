package repository

import (
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
)

func ListGranularityResultsByExecutionID(executionID int) ([]database.GranularityResult, error) {
	var results []database.GranularityResult
	if err := database.DB.Model(&database.GranularityResult{}).
		Where("execution_id = ?", executionID).
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to list granularity results by execution ID: %v", err)
	}

	return results, nil
}
