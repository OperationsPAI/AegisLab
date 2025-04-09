package repository

import (
	"github.com/CUHK-SE-Group/rcabench/database"
)

func listGranularityWithFilters(executionIDs []int, levels []string, rank int) ([]database.GranularityResult, error) {
	query := database.DB.Model(&database.GranularityResult{}).
		Where("rank <= ?", rank)

	if len(levels) > 0 {
		query = query.Where("level IN (?)", levels)
	}

	var results []database.GranularityResult
	if err := query.Find(&results).Error; err != nil {
		return nil, err
	}

	return results, nil
}
