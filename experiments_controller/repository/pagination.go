package repository

import (
	"fmt"

	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/sirupsen/logrus"
)

// 泛型分页查询函数
func paginateQuery[T any](
	condition string, // 查询条件语句
	conditionArgs []any, // 查询条件参数
	sortField string, // 排序字段及方向
	pageNum int, // 页码
	pageSize int, // 每页数量
	selectColumns []string, // 指定查询字段（可选）
) (total int64, records []T, err error) {
	var model T

	db := database.DB.Model(&model).Where(condition, conditionArgs...)
	if err = db.Count(&total).Error; err != nil {
		return 0, nil, fmt.Errorf("count error: %w", err)
	}

	query := db.Scopes(
		database.Sort(sortField),
		database.Paginate(pageNum, pageSize),
	)

	if len(selectColumns) > 0 {
		query = query.Select(selectColumns)
	}

	if err = query.Find(&records).Error; err != nil {
		message := "failed to find records"
		logrus.Errorf("%s: %v", message, err)
		return total, nil, fmt.Errorf("failed to find records")
	}

	return total, records, nil
}
