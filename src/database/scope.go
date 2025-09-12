package database

import (
	"fmt"

	"gorm.io/gorm"
)

// Fuzzy search Scope
func KeywordSearch(keyword string, fields ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if keyword == "" {
			return db
		}
		query := ""
		for i, field := range fields {
			if i > 0 {
				query += " OR "
			}
			query += fmt.Sprintf("%s LIKE ?", field)
		}
		return db.Where(query, "%"+keyword+"%")
	}
}

func CursorPaginate(lastID uint, size int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if lastID > 0 {
			db = db.Where("id > ?", lastID)
		}
		return db.Limit(size)
	}
}

// Pagination Scope
func Paginate(pageNum, pageSize int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		offset := (pageNum - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

// Sort Scope
func Sort(sort string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if sort == "" {
			sort = "id desc"
		}
		return db.Order(sort)
	}
}
