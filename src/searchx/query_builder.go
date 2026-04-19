package searchx

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"aegis/dto"

	"gorm.io/gorm"
)

// QueryBuilder provides methods to build complex database queries from SearchRequest.
type QueryBuilder[F ~string] struct {
	db                *gorm.DB
	query             *gorm.DB
	allowedSortFields map[F]string
}

func NewQueryBuilder[F ~string](db *gorm.DB, allowedSortFields map[F]string) *QueryBuilder[F] {
	return &QueryBuilder[F]{
		db:                db,
		query:             db,
		allowedSortFields: allowedSortFields,
	}
}

func (qb *QueryBuilder[F]) ApplySearchReq(filters []dto.SearchFilter, keyword string, sortOptions []dto.TypedSortOption[F], groupBy []F, modelType any) *gorm.DB {
	qb.query = qb.db.Model(modelType)
	qb.applyFilters(filters)
	if keyword != "" {
		qb.applyKeywordSearch(keyword, modelType)
	}
	qb.applySorting(sortOptions, groupBy)
	return qb.query
}

func (qb *QueryBuilder[F]) ApplyIncludes(includes []string) {
	for _, include := range includes {
		qb.query = qb.query.Preload(include)
	}
}

func (qb *QueryBuilder[F]) ApplyIncludeFields(includeFields []string) {
	for _, field := range includeFields {
		qb.query = qb.query.Select(field)
	}
}

func (qb *QueryBuilder[F]) ApplyExcludeFields(excludeFields []string, modelType any) {
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
			continue
		}
		allFields = append(allFields, field.Name)
	}

	fieldsToSelect := make([]string, 0, len(allFields))
	excludeMap := make(map[string]struct{}, len(excludeFields))
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

func (qb *QueryBuilder[F]) GetCount() (int64, error) {
	var count int64
	err := qb.query.Count(&count).Error
	return count, err
}

func (qb *QueryBuilder[F]) Query() *gorm.DB {
	return qb.query
}

func (qb *QueryBuilder[F]) applyFilters(filters []dto.SearchFilter) {
	for _, filter := range filters {
		qb.applySingleFilter(filter)
	}
}

func (qb *QueryBuilder[F]) applyKeywordSearch(keyword string, modelType any) {
	searchableFields := qb.getSearchableFields(modelType)
	if len(searchableFields) == 0 {
		return
	}

	var conditions []string
	var values []any
	for _, field := range searchableFields {
		conditions = append(conditions, fmt.Sprintf("%s LIKE ?", field))
		values = append(values, "%"+keyword+"%")
	}

	qb.query = qb.query.Where(strings.Join(conditions, " OR "), values...)
}

func (qb *QueryBuilder[F]) applySingleFilter(filter dto.SearchFilter) {
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

func (qb *QueryBuilder[F]) applySorting(sortOptions []dto.TypedSortOption[F], groupBy []F) {
	applied := false

	for _, field := range groupBy {
		if dbField, ok := qb.allowedSortFields[field]; ok {
			qb.query = qb.query.Order(dbField + " ASC")
			applied = true
		}
	}

	for _, sort := range sortOptions {
		dbField, ok := qb.allowedSortFields[sort.Field]
		if !ok {
			continue
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

func (qb *QueryBuilder[F]) getSearchableFields(modelType any) []string {
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

func (qb *QueryBuilder[F]) getTypeName(modelType any) string {
	t := reflect.TypeOf(modelType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

func (qb *QueryBuilder[F]) sanitizeFieldName(field string) string {
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

func resolveMultiValues(filter dto.SearchFilter) []string {
	if len(filter.Values) > 0 {
		return filter.Values
	}

	if strings.TrimSpace(filter.Value) == "" {
		return nil
	}

	var items []string
	if strings.HasPrefix(strings.TrimSpace(filter.Value), "[") && json.Unmarshal([]byte(filter.Value), &items) == nil {
		return items
	}
	return []string{filter.Value}
}
