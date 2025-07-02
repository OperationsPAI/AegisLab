package repository

import (
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type QueryBuilder func(*gorm.DB) *gorm.DB

type genericQueryParams struct {
	builder       QueryBuilder
	sortField     string
	limit         int
	pageNum       int
	pageSize      int
	selectColumns []string
}

func genericQueryWithBuilder[T any](params *genericQueryParams) (total int64, records []T, err error) {
	var model T

	baseQuery := database.DB.Model(&model)
	query := params.builder(baseQuery)

	if params.sortField != "" {
		query = query.Scopes(database.Sort(params.sortField))
	}

	if params.limit > 0 {
		query = query.Limit(params.limit)
	}

	if params.pageNum > 0 && params.pageSize > 0 {
		if err = query.Count(&total).Error; err != nil {
			return 0, nil, fmt.Errorf("count error: %v", err)
		}

		query = query.Scopes(database.Paginate(params.pageNum, params.pageSize))
	}

	if len(params.selectColumns) > 0 {
		query = query.Select(params.selectColumns)
	}

	if err = query.Find(&records).Error; err != nil {
		logrus.Errorf("failed to find records: %v", err)
		return total, nil, fmt.Errorf("failed to find records")
	}

	return total, records, nil
}
