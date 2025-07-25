package repository

import (
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
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

	query := params.builder(database.DB.Model(&model))

	if params.sortField != "" {
		query = query.Scopes(database.Sort(params.sortField))
	}

	if params.pageNum > 0 && params.pageSize > 0 {
		countQuery := params.builder(database.DB.Model(&model))
		if err = countQuery.Limit(-1).Offset(-1).Count(&total).Error; err != nil {
			return 0, nil, fmt.Errorf("Count error: %v", err)
		}

		query = query.Scopes(database.Paginate(params.pageNum, params.pageSize))
	} else if params.limit > 0 {
		query = query.Limit(params.limit)
	}

	if len(params.selectColumns) > 0 {
		query = query.Select(params.selectColumns)
	}

	if err = query.Find(&records).Error; err != nil {
		return total, nil, fmt.Errorf("Failed to find records :%v", err)
	}

	return total, records, nil
}
