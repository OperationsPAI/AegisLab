package repository

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/LGU-SE-Internal/rcabench/dto"
	"gorm.io/gorm"
)

// SearchQueryBuilder provides methods to build complex database queries from SearchRequest
type SearchQueryBuilder struct {
	db    *gorm.DB
	query *gorm.DB
}

// NewSearchQueryBuilder creates a new search query builder
func NewSearchQueryBuilder(db *gorm.DB) *SearchQueryBuilder {
	return &SearchQueryBuilder{
		db:    db,
		query: db,
	}
}

// ApplySearchRequest applies filters, sorting, and pagination from SearchRequest
func (qb *SearchQueryBuilder) ApplySearchRequest(searchReq *dto.SearchRequest, modelType interface{}) *gorm.DB {
	// Start with the base query
	qb.query = qb.db.Model(modelType)

	// Apply filters
	qb.applyFilters(searchReq.Filters)

	// Apply keyword search if provided
	if searchReq.Keyword != "" {
		qb.applyKeywordSearch(searchReq.Keyword, modelType)
	}

	// Apply sorting
	qb.applySorting(searchReq.Sort)

	return qb.query
}

// ApplyPagination applies pagination to the query
func (qb *SearchQueryBuilder) ApplyPagination(searchReq *dto.SearchRequest) *gorm.DB {
	offset := searchReq.GetOffset()
	return qb.query.Offset(offset).Limit(searchReq.Size)
}

// GetCount gets the total count before pagination
func (qb *SearchQueryBuilder) GetCount() (int64, error) {
	var count int64
	err := qb.query.Count(&count).Error
	return count, err
}

// applyFilters applies all filters to the query
func (qb *SearchQueryBuilder) applyFilters(filters []dto.SearchFilter) {
	for _, filter := range filters {
		qb.applySingleFilter(filter)
	}
}

// applySingleFilter applies a single filter to the query
func (qb *SearchQueryBuilder) applySingleFilter(filter dto.SearchFilter) {
	field := qb.sanitizeFieldName(filter.Field)

	switch filter.Operator {
	case dto.OpEqual:
		qb.query = qb.query.Where(fmt.Sprintf("%s = ?", field), filter.Value)

	case dto.OpNotEqual:
		qb.query = qb.query.Where(fmt.Sprintf("%s != ?", field), filter.Value)

	case dto.OpGreater:
		qb.query = qb.query.Where(fmt.Sprintf("%s > ?", field), filter.Value)

	case dto.OpGreaterEq:
		qb.query = qb.query.Where(fmt.Sprintf("%s >= ?", field), filter.Value)

	case dto.OpLess:
		qb.query = qb.query.Where(fmt.Sprintf("%s < ?", field), filter.Value)

	case dto.OpLessEq:
		qb.query = qb.query.Where(fmt.Sprintf("%s <= ?", field), filter.Value)

	case dto.OpLike:
		qb.query = qb.query.Where(fmt.Sprintf("%s LIKE ?", field), "%"+fmt.Sprintf("%v", filter.Value)+"%")

	case dto.OpStartsWith:
		qb.query = qb.query.Where(fmt.Sprintf("%s LIKE ?", field), fmt.Sprintf("%v", filter.Value)+"%")

	case dto.OpEndsWith:
		qb.query = qb.query.Where(fmt.Sprintf("%s LIKE ?", field), "%"+fmt.Sprintf("%v", filter.Value))

	case dto.OpNotLike:
		qb.query = qb.query.Where(fmt.Sprintf("%s NOT LIKE ?", field), "%"+fmt.Sprintf("%v", filter.Value)+"%")

	case dto.OpIn:
		if len(filter.Values) > 0 {
			qb.query = qb.query.Where(fmt.Sprintf("%s IN ?", field), filter.Values)
		}

	case dto.OpNotIn:
		if len(filter.Values) > 0 {
			qb.query = qb.query.Where(fmt.Sprintf("%s NOT IN ?", field), filter.Values)
		}

	case dto.OpIsNull:
		qb.query = qb.query.Where(fmt.Sprintf("%s IS NULL", field))

	case dto.OpIsNotNull:
		qb.query = qb.query.Where(fmt.Sprintf("%s IS NOT NULL", field))

	case dto.OpDateEqual:
		qb.query = qb.query.Where(fmt.Sprintf("DATE(%s) = DATE(?)", field), filter.Value)

	case dto.OpDateAfter:
		qb.query = qb.query.Where(fmt.Sprintf("DATE(%s) > DATE(?)", field), filter.Value)

	case dto.OpDateBefore:
		qb.query = qb.query.Where(fmt.Sprintf("DATE(%s) < DATE(?)", field), filter.Value)

	case dto.OpDateBetween:
		if len(filter.Values) == 2 {
			qb.query = qb.query.Where(fmt.Sprintf("DATE(%s) BETWEEN DATE(?) AND DATE(?)", field), filter.Values[0], filter.Values[1])
		}
	}
}

// applySorting applies sorting to the query
func (qb *SearchQueryBuilder) applySorting(sortOptions []dto.SortOption) {
	for _, sort := range sortOptions {
		field := qb.sanitizeFieldName(sort.Field)
		direction := strings.ToUpper(string(sort.Direction))

		// Validate direction
		if direction != "ASC" && direction != "DESC" {
			direction = "ASC"
		}

		qb.query = qb.query.Order(fmt.Sprintf("%s %s", field, direction))
	}

	// Add default sorting if no sort options provided
	if len(sortOptions) == 0 {
		qb.query = qb.query.Order("id DESC")
	}
}

// applyKeywordSearch applies general keyword search across searchable fields
func (qb *SearchQueryBuilder) applyKeywordSearch(keyword string, modelType interface{}) {
	// Get searchable fields from model type
	searchableFields := qb.getSearchableFields(modelType)

	if len(searchableFields) == 0 {
		return
	}

	// Build OR conditions for keyword search
	var conditions []string
	var values []interface{}

	for _, field := range searchableFields {
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
		values = append(values, "%"+keyword+"%")
	}

	whereClause := strings.Join(conditions, " OR ")
	qb.query = qb.query.Where(whereClause, values...)
}

// getSearchableFields returns fields that can be searched with keywords
func (qb *SearchQueryBuilder) getSearchableFields(modelType interface{}) []string {
	// This is a simplified implementation
	// In a real application, you might want to use struct tags or configuration
	// to mark fields as searchable

	searchableFields := map[string][]string{
		"User":       {"username", "email", "full_name"},
		"Role":       {"name", "display_name", "description"},
		"Permission": {"name", "display_name", "description"},
		"Project":    {"name", "description"},
		"Task":       {"name", "description"},
		"Dataset":    {"name", "description"},
		"Container":  {"name"},
	}

	typeName := qb.getTypeName(modelType)
	if fields, exists := searchableFields[typeName]; exists {
		return fields
	}

	return []string{}
}

// getTypeName gets the type name from interface
func (qb *SearchQueryBuilder) getTypeName(modelType interface{}) string {
	t := reflect.TypeOf(modelType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// sanitizeFieldName sanitizes field names to prevent SQL injection
func (qb *SearchQueryBuilder) sanitizeFieldName(field string) string {
	// Convert snake_case to database field names
	// Remove any special characters that could be used for injection
	field = strings.ReplaceAll(field, ";", "")
	field = strings.ReplaceAll(field, "--", "")
	field = strings.ReplaceAll(field, "/*", "")
	field = strings.ReplaceAll(field, "*/", "")

	// Convert common field names
	fieldMap := map[string]string{
		"full_name":     "full_name",
		"is_active":     "is_active",
		"is_system":     "is_system",
		"created_at":    "created_at",
		"updated_at":    "updated_at",
		"last_login_at": "last_login_at",
		"user_id":       "user_id",
		"role_id":       "role_id",
		"permission_id": "permission_id",
		"project_id":    "project_id",
		"resource_id":   "resource_id",
		"display_name":  "display_name",
	}

	if dbField, exists := fieldMap[field]; exists {
		return dbField
	}

	return field
}

// BuildSearchResponse builds a standard search response
func BuildSearchResponse[T any](items []T, totalCount int64, searchReq *dto.SearchRequest) dto.SearchResponse[T] {
	totalPages := int((totalCount + int64(searchReq.Size) - 1) / int64(searchReq.Size))

	return dto.SearchResponse[T]{
		Items: items,
		Pagination: dto.PaginationInfo{
			Page:       searchReq.Page,
			Size:       searchReq.Size,
			Total:      totalCount,
			TotalPages: totalPages,
		},
		Filters: searchReq.Filters,
		Sort:    searchReq.Sort,
	}
}

// ExecuteSearch executes a complete search operation
func ExecuteSearch[T any](db *gorm.DB, searchReq *dto.SearchRequest, modelType T) (dto.SearchResponse[T], error) {
	qb := NewSearchQueryBuilder(db)

	// Apply search conditions
	qb.ApplySearchRequest(searchReq, modelType)

	// Get total count
	totalCount, err := qb.GetCount()
	if err != nil {
		return dto.SearchResponse[T]{}, fmt.Errorf("failed to get count: %w", err)
	}

	// Apply pagination and execute query
	var items []T
	err = qb.ApplyPagination(searchReq).Find(&items).Error
	if err != nil {
		return dto.SearchResponse[T]{}, fmt.Errorf("failed to execute search: %w", err)
	}

	return BuildSearchResponse(items, totalCount, searchReq), nil
}
