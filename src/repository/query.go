package repository

import (
	"fmt"

	"rcabench/database"
	"gorm.io/gorm"
)

// QueryBuilder represents a function that builds database queries
type QueryBuilder func(*gorm.DB) *gorm.DB

// GenericQueryParams represents parameters for generic queries
type GenericQueryParams struct {
	Builder       QueryBuilder
	SortField     string
	Limit         int
	PageNum       int
	PageSize      int
	SelectColumns []string
}

// GenericQueryWithBuilder executes a generic query with the provided builder
func GenericQueryWithBuilder[T any](params *GenericQueryParams) (total int64, records []T, err error) {
	var model T

	query := params.Builder(database.DB.Model(&model))

	if params.SortField != "" {
		query = query.Scopes(database.Sort(params.SortField))
	}

	if params.PageNum > 0 && params.PageSize > 0 {
		countQuery := params.Builder(database.DB.Model(&model))
		if err = countQuery.Limit(-1).Offset(-1).Count(&total).Error; err != nil {
			return 0, nil, fmt.Errorf("count error: %v", err)
		}

		query = query.Scopes(database.Paginate(params.PageNum, params.PageSize))
	} else if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}

	if len(params.SelectColumns) > 0 {
		query = query.Select(params.SelectColumns)
	}

	if err = query.Find(&records).Error; err != nil {
		return total, nil, fmt.Errorf("failed to find records: %v", err)
	}

	return total, records, nil
}
