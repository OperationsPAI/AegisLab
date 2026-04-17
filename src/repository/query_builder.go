package repository

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"aegis/dto"

	"gorm.io/gorm"
)

// SearchQueryBuilder provides methods to build complex database queries from SearchRequest
type SearchQueryBuilder[F ~string] struct {
	db                *gorm.DB
	query             *gorm.DB
	allowedSortFields map[F]string // user field name -> DB column name (whitelist for sort/group)
}

// NewSearchQueryBuilder creates a new search query builder.
// allowedSortFields is a whitelist mapping user-facing field names to DB column names
// for sort and group_by operations. If nil, sorting defaults to "id DESC" only.
func NewSearchQueryBuilder[F ~string](db *gorm.DB, allowedSortFields map[F]string) *SearchQueryBuilder[F] {
	return &SearchQueryBuilder[F]{
		db:                db,
		query:             db,
		allowedSortFields: allowedSortFields,
	}
}

// ApplySearchReq applies filters, sorting, and pagination from SearchRequest.
func (qb *SearchQueryBuilder[F]) ApplySearchReq(filters []dto.SearchFilter, keyword string, sortOptions []dto.TypedSortOption[F], groupBy []F, modelType interface{}) *gorm.DB {
	// Start with the base query
	qb.query = qb.db.Model(modelType)

	// Apply filters
	qb.applyFilters(filters)

	// Apply keyword search if provided
	if keyword != "" {
		qb.applyKeywordSearch(keyword, modelType)
	}

	// Apply sorting (group_by fields first, then user sort)
	qb.applySorting(sortOptions, groupBy)

	return qb.query
}

// applyFilters applies all filters to the query
func (qb *SearchQueryBuilder[F]) applyFilters(filters []dto.SearchFilter) {
	for _, filter := range filters {
		qb.applySingleFilter(filter)
	}
}

// applyInclude applies include options to the query
func (qb *SearchQueryBuilder[F]) applyIncludes(includes []string) {
	for _, include := range includes {
		qb.query = qb.query.Preload(include)
	}
}

func (qb *SearchQueryBuilder[F]) ApplyIncludes(includes []string) {
	qb.applyIncludes(includes)
}

// applyIncludeFields includes specified fields in the query
func (qb *SearchQueryBuilder[F]) applyIncludeFields(includeFields []string) {
	for _, field := range includeFields {
		qb.query = qb.query.Select(field)
	}
}

func (qb *SearchQueryBuilder[F]) ApplyIncludeFields(includeFields []string) {
	qb.applyIncludeFields(includeFields)
}

// applyExcludeFields excludes specified fields from the query
func (qb *SearchQueryBuilder[F]) applyExcludeFields(excludeFields []string, modelType interface{}) {
	// Get all fields from model type
	t := reflect.TypeOf(modelType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var allFields []string
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		dbTag := field.Tag.Get("gorm")
		if dbTag != "" {
			dbField := strings.Split(dbTag, ";")[0]
			allFields = append(allFields, dbField)
		} else {
			allFields = append(allFields, field.Name)
		}
	}

	// Determine fields to select
	fieldsToSelect := make([]string, 0, len(allFields))
	excludeMap := make(map[string]struct{})
	for _, field := range excludeFields {
		excludeMap[field] = struct{}{}
	}

	for _, field := range allFields {
		if _, excluded := excludeMap[field]; !excluded {
			fieldsToSelect = append(fieldsToSelect, field)
		}
	}

	if len(fieldsToSelect) > 0 {
		qb.query = qb.query.Select(strings.Join(fieldsToSelect, ", "))
	}
}

func (qb *SearchQueryBuilder[F]) ApplyExcludeFields(excludeFields []string, modelType interface{}) {
	qb.applyExcludeFields(excludeFields, modelType)
}

// applyKeywordSearch applies general keyword search across searchable fields
func (qb *SearchQueryBuilder[F]) applyKeywordSearch(keyword string, modelType interface{}) {
	// Get searchable fields from model type
	searchableFields := qb.getSearchableFields(modelType)

	if len(searchableFields) == 0 {
		return
	}

	// Build OR conditions for keyword search
	var conditions []string
	var values []any

	for _, field := range searchableFields {
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
		values = append(values, "%"+keyword+"%")
	}

	whereClause := strings.Join(conditions, " OR ")
	qb.query = qb.query.Where(whereClause, values...)
}

// applySingleFilter applies a single filter to the query
func (qb *SearchQueryBuilder[F]) applySingleFilter(filter dto.SearchFilter) {
	field := qb.sanitizeFieldName(filter.Field)
	if field == "" {
		return
	}

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
		if values := resolveMultiValues(filter); len(values) > 0 {
			qb.query = qb.query.Where(fmt.Sprintf("%s IN (?)", field), values)
		}

	case dto.OpNotIn:
		if values := resolveMultiValues(filter); len(values) > 0 {
			qb.query = qb.query.Where(fmt.Sprintf("%s NOT IN (?)", field), values)
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

// applyPagination applies pagination to the query
func (qb *SearchQueryBuilder[F]) applyPagination(pagination *dto.PaginationReq) *gorm.DB {
	offset := (pagination.Page - 1) * int(pagination.Size)
	return qb.query.Offset(offset).Limit(int(pagination.Size))
}

// applySorting applies sorting to the query using a whitelist approach.
// GroupBy fields are applied first (ASC) to ensure items in the same group are adjacent,
// then user sort options are applied within each group.
func (qb *SearchQueryBuilder[F]) applySorting(sortOptions []dto.TypedSortOption[F], groupBy []F) {
	applied := false

	// Apply group_by fields first for consistent grouping order
	for _, field := range groupBy {
		if dbField, ok := qb.allowedSortFields[field]; ok {
			qb.query = qb.query.Order(dbField + " ASC")
			applied = true
		}
	}

	// Apply user sort options (whitelist validated via typed key lookup)
	for _, sort := range sortOptions {
		dbField, ok := qb.allowedSortFields[sort.Field]
		if !ok {
			continue // skip fields not in whitelist
		}
		direction := "ASC"
		if strings.ToUpper(string(sort.Direction)) == "DESC" {
			direction = "DESC"
		}
		qb.query = qb.query.Order(dbField + " " + direction)
		applied = true
	}

	if !applied {
		qb.query = qb.query.Order("id DESC")
	}
}

// GetCount gets the total count before pagination
func (qb *SearchQueryBuilder[F]) getCount() (int64, error) {
	var count int64
	err := qb.query.Count(&count).Error
	return count, err
}

func (qb *SearchQueryBuilder[F]) GetCount() (int64, error) {
	return qb.getCount()
}

func (qb *SearchQueryBuilder[F]) Query() *gorm.DB {
	return qb.query
}

// getSearchableFields returns fields that can be searched with keywords
func (qb *SearchQueryBuilder[F]) getSearchableFields(modelType interface{}) []string {
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
func (qb *SearchQueryBuilder[F]) getTypeName(modelType interface{}) string {
	t := reflect.TypeOf(modelType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// sanitizeFieldName validates that a field name contains only safe characters
// (alphanumeric, underscore, dot for table.column notation).
// Returns empty string if any unsafe character is detected.
func (qb *SearchQueryBuilder[F]) sanitizeFieldName(field string) string {
	if field == "" {
		return ""
	}
	for _, c := range field {
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '_' && c != '.' {
			return ""
		}
	}
	return field
}

// resolveMultiValues returns the effective []string for IN/NOT IN operators.
// It prefers filter.Values when populated; otherwise it tries to parse filter.Value
// as a JSON array (e.g. "[\"a\",\"b\"]" or "[1,2]").
// A bare non-JSON single value is wrapped in a one-element slice.
func resolveMultiValues(filter dto.SearchFilter) []string {
	if len(filter.Values) > 0 {
		return filter.Values
	}
	if filter.Value == "" {
		return nil
	}
	// Try JSON array parse
	var parsed []any
	if err := json.Unmarshal([]byte(filter.Value), &parsed); err == nil {
		result := make([]string, len(parsed))
		for i, v := range parsed {
			result[i] = fmt.Sprintf("%v", v)
		}
		return result
	}
	// Fallback: treat the whole value as a single element
	return []string{filter.Value}
}
