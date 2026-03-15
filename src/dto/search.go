package dto

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"aegis/consts"
)

// SortDirection represents sort direction
type SortDirection string

const (
	SortASC  SortDirection = "asc"
	SortDESC SortDirection = "desc"
)

// FilterOperator represents filter operations
type FilterOperator string

const (
	// Comparison operators
	OpEqual     FilterOperator = "eq"  // ==
	OpNotEqual  FilterOperator = "ne"  // !=
	OpGreater   FilterOperator = "gt"  // >
	OpGreaterEq FilterOperator = "gte" // >=
	OpLess      FilterOperator = "lt"  // <
	OpLessEq    FilterOperator = "lte" // <=

	// String operators
	OpLike       FilterOperator = "like"   // LIKE %value%
	OpStartsWith FilterOperator = "starts" // LIKE value%
	OpEndsWith   FilterOperator = "ends"   // LIKE %value
	OpNotLike    FilterOperator = "nlike"  // NOT LIKE %value%

	// Array operators
	OpIn    FilterOperator = "in"  // IN (value1, value2, ...)
	OpNotIn FilterOperator = "nin" // NOT IN (value1, value2, ...)

	// Null operators
	OpIsNull    FilterOperator = "null"  // IS NULL
	OpIsNotNull FilterOperator = "nnull" // IS NOT NULL

	// Date operators
	OpDateEqual   FilterOperator = "deq"      // DATE(field) = DATE(value)
	OpDateAfter   FilterOperator = "dafter"   // DATE(field) > DATE(value)
	OpDateBefore  FilterOperator = "dbefore"  // DATE(field) < DATE(value)
	OpDateBetween FilterOperator = "dbetween" // DATE(field) BETWEEN date1 AND date2
)

// SearchFilter represents a single filter condition
type SearchFilter struct {
	Field    string         `json:"field"`            // Field name
	Operator FilterOperator `json:"operator"`         // Operator
	Value    string         `json:"value"`            // Value (can be string, number, boolean, etc.)
	Values   []string       `json:"values,omitempty"` // Multiple values (for IN operations etc.)
}

// SortOption represents a sort option with a plain string field (used internally in SearchReq).
type SortOption struct {
	Field     string        `json:"field" binding:"omitempty"`     // Sort field
	Direction SortDirection `json:"direction" binding:"omitempty"` // Sort direction
}

func (so *SortOption) Validate() error {
	if so.Direction != SortASC && so.Direction != SortDESC {
		return fmt.Errorf("invalid sort direction: %s", so.Direction)
	}
	return nil
}

// TypedSortOption is a sort option whose field name is constrained to a specific type F.
// Use with a consts field type (e.g. consts.InjectionField) to enforce compile-time type safety.
type TypedSortOption[F ~string] struct {
	Field     F             `json:"field" binding:"omitempty"`     // Sort field (constrained to allowed values)
	Direction SortDirection `json:"direction" binding:"omitempty"` // Sort direction
}

func (so TypedSortOption[F]) Validate() error {
	if so.Direction != SortASC && so.Direction != SortDESC {
		return fmt.Errorf("invalid sort direction: %s", so.Direction)
	}
	return nil
}

// ToSortOption converts a TypedSortOption to a plain SortOption for the query layer.
func (so TypedSortOption[F]) ToSortOption() SortOption {
	return SortOption{Field: string(so.Field), Direction: so.Direction}
}

// SearchReq represents a complex search request.
// F constrains the sort/group_by fields to a typed constant set; use string when unconstrained.
type SearchReq[F ~string] struct {
	// Pagination
	PaginationReq

	// Filters
	Filters []SearchFilter `json:"filters,omitempty"`

	// Sort — typed field names (constrained to F)
	Sort []TypedSortOption[F] `json:"sort,omitempty"`

	// GroupBy fields for tree-structured results (ordered by nesting level)
	GroupBy []F `json:"group_by,omitempty"`

	// Search keyword (for general text search)
	Keyword string `json:"keyword,omitempty" form:"keyword"`

	// Include/Exclude fields
	IncludeFields []string `json:"include_fields,omitempty"`
	ExcludeFields []string `json:"exclude_fields,omitempty"`

	// Include related entities
	Includes []string `json:"includes,omitempty" form:"include"`
}

// GetOffset calculates the offset for pagination
func (sr *SearchReq[F]) GetOffset() int {
	return (sr.Page - 1) * int(sr.Size)
}

// HasFilter checks if a specific filter exists
func (sr *SearchReq[F]) HasFilter(field string) bool {
	for _, filter := range sr.Filters {
		if filter.Field == field {
			return true
		}
	}
	return false
}

// GetFilter gets a specific filter by field name
func (sr *SearchReq[F]) GetFilter(field string) *SearchFilter {
	for _, filter := range sr.Filters {
		if filter.Field == field {
			return &filter
		}
	}
	return nil
}

// AddFilter adds a new filter
func (sr *SearchReq[F]) AddFilter(field string, operator FilterOperator, value any) {
	filter := SearchFilter{
		Field:    field,
		Operator: operator,
	}

	// For IN/NOT IN operators, convert slice values to the Values field
	if operator == OpIn || operator == OpNotIn {
		rv := reflect.ValueOf(value)
		if rv.Kind() == reflect.Slice {
			values := make([]string, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				values[i] = fmt.Sprintf("%v", rv.Index(i).Interface())
			}
			filter.Values = values
			sr.Filters = append(sr.Filters, filter)
			return
		}
	}

	filter.Value = fmt.Sprintf("%v", value)
	sr.Filters = append(sr.Filters, filter)
}

// AddInclude adds a new include option
func (sr *SearchReq[F]) AddInclude(field string) {
	sr.Includes = append(sr.Includes, field)
}

// AddSort adds a typed sort option
func (sr *SearchReq[F]) AddSort(field F, direction SortDirection) {
	sr.Sort = append(sr.Sort, TypedSortOption[F]{Field: field, Direction: direction})
}

// SortOptions returns the sort options as plain []SortOption for the repository/query layer.
// The string conversion from F happens only here, at the SQL boundary.
func (sr *SearchReq[F]) SortOptions() []SortOption {
	opts := make([]SortOption, len(sr.Sort))
	for i, s := range sr.Sort {
		opts[i] = s.ToSortOption()
	}
	return opts
}

// GroupByStrings returns the group_by fields as plain []string for the repository/query layer.
func (sr *SearchReq[F]) GroupByStrings() []string {
	groups := make([]string, len(sr.GroupBy))
	for i, f := range sr.GroupBy {
		groups[i] = string(f)
	}
	return groups
}

// DateRange represents a date range filter
type DateRange struct {
	From *time.Time `json:"from,omitempty" binding:"omitempty"`
	To   *time.Time `json:"to,omitempty" binding:"omitempty"`
}

func (dr *DateRange) Validate() error {
	if dr.From != nil && dr.To != nil && dr.From.After(*dr.To) {
		return fmt.Errorf("invalid date range: 'from' date is after 'to' date")
	}
	return nil
}

// NumberRange represents a number range filter
type NumberRange struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

func (nr *NumberRange) Validate() error {
	if nr.Min != nil && nr.Max != nil && *nr.Min > *nr.Max {
		return fmt.Errorf("invalid number range: min is greater than max")
	}
	return nil
}

// AdvancedSearchReq extends SearchRequest with common filter shortcuts.
// F constrains the allowed field names for sort and group_by to a typed constant set.
// Use a consts field type (e.g. consts.InjectionField) for compile-time safety;
// use 'string' when no field type constraint is needed.
type AdvancedSearchReq[F ~string] struct {
	PaginationReq
	Sort    []TypedSortOption[F] `json:"sort" binding:"omitempty"`
	GroupBy []F                  `json:"group_by" binding:"omitempty"` // Group results into tree structure by these fields (ordered)

	Statuses  []consts.StatusType `json:"status" binding:"omitempty"`
	CreatedAt *DateRange          `json:"created_at" binding:"omitempty"`
	UpdatedAt *DateRange          `json:"updated_at" binding:"omitempty"`
}

func (req *AdvancedSearchReq[F]) Validate() error {
	if err := req.PaginationReq.Validate(); err != nil {
		return err
	}

	if len(req.Sort) > 0 {
		for i, so := range req.Sort {
			if err := so.Validate(); err != nil {
				return fmt.Errorf("invalid sort option at index %d: %w", i, err)
			}
		}
	}

	for i, field := range req.GroupBy {
		if strings.TrimSpace(string(field)) == "" {
			return fmt.Errorf("empty group_by field at index %d", i)
		}
	}

	if len(req.Statuses) > 0 {
		for i, status := range req.Statuses {
			if _, exists := consts.ValidStatuses[status]; !exists {
				return fmt.Errorf("invalid status value at index %d: %d", i, status)
			}
		}
	}

	if req.CreatedAt != nil {
		if err := req.CreatedAt.Validate(); err != nil {
			return fmt.Errorf("invalid created_at range: %w", err)
		}
	}
	if req.UpdatedAt != nil {
		if err := req.UpdatedAt.Validate(); err != nil {
			return fmt.Errorf("invalid updated_at range: %w", err)
		}
	}

	return nil
}

// ConvertAdvancedToSearch converts AdvancedSearchReq to SearchReq[F].
// Sort and GroupBy are assigned directly — no string conversion; types are preserved.
func (req *AdvancedSearchReq[F]) ConvertAdvancedToSearch() *SearchReq[F] {
	sr := &SearchReq[F]{}
	sr.PaginationReq = req.PaginationReq // Copy pagination fields
	sr.Sort = req.Sort                   // []TypedSortOption[F] — direct, no conversion
	sr.GroupBy = req.GroupBy             // []F — direct, no conversion

	if len(req.Statuses) > 0 {
		values := make([]string, len(req.Statuses))
		for i, v := range req.Statuses {
			values[i] = fmt.Sprintf("%v", v)
		}
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "status",
			Operator: OpIn,
			Values:   values,
		})
	} else {
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "status",
			Operator: OpNotEqual,
			Value:    fmt.Sprintf("%v", consts.CommonDeleted),
		})
	}

	if req.CreatedAt != nil {
		if req.CreatedAt.From != nil && req.CreatedAt.To != nil {
			sr.AddFilter("created_at", OpDateBetween, []any{req.CreatedAt.From, req.CreatedAt.To})
		} else if req.CreatedAt.From != nil {
			sr.AddFilter("created_at", OpDateAfter, req.CreatedAt.From)
		} else if req.CreatedAt.To != nil {
			sr.AddFilter("created_at", OpDateBefore, req.CreatedAt.To)
		}
	}
	if req.UpdatedAt != nil {
		if req.UpdatedAt.From != nil && req.UpdatedAt.To != nil {
			sr.AddFilter("updated_at", OpDateBetween, []any{req.UpdatedAt.From, req.UpdatedAt.To})
		} else if req.UpdatedAt.From != nil {
			sr.AddFilter("updated_at", OpDateAfter, req.UpdatedAt.From)
		} else if req.UpdatedAt.To != nil {
			sr.AddFilter("updated_at", OpDateBefore, req.UpdatedAt.To)
		}
	}

	return sr
}

// AlgorithmSearchReq represents algorithm search request.
// Uses AdvancedSearchReq[string] since algorithm sort fields have no typed constant set yet.
type AlgorithmSearchReq struct {
	AdvancedSearchReq[string]

	// Algorithm-specific filters
	Name  *string `json:"name,omitempty"`
	Image *string `json:"image,omitempty"`
	Tag   *string `json:"tag,omitempty"`
	Type  *string `json:"type,omitempty"`
}

// ConvertToSearchRequest converts AlgorithmSearchRequest to SearchRequest
func (req *AlgorithmSearchReq) ConvertToSearchRequest() *SearchReq[string] {
	sr := req.ConvertAdvancedToSearch()

	// Add algorithm-specific filters
	if req.Name != nil {
		sr.AddFilter("name", OpLike, *req.Name)
	}
	if req.Image != nil {
		sr.AddFilter("image", OpLike, *req.Image)
	}
	if req.Tag != nil {
		sr.AddFilter("tag", OpEqual, *req.Tag)
	}
	if req.Type != nil {
		sr.AddFilter("type", OpEqual, *req.Type)
	}

	// Default to only active algorithms
	if !sr.HasFilter("status") {
		sr.AddFilter("status", OpEqual, consts.CommonEnabled)
	}

	return sr
}

// ListResp represents the response for list operations
type ListResp[T any] struct {
	Items      []T             `json:"items"`
	Pagination *PaginationInfo `json:"pagination"`
}

// SearchGroupNode represents a node in the grouped tree result
type SearchGroupNode[T any] struct {
	Field    string               `json:"field"`              // Group field name
	Value    string               `json:"value"`              // Group field value
	Count    int                  `json:"count"`              // Number of items in this group
	Children []SearchGroupNode[T] `json:"children,omitempty"` // Sub-groups (non-leaf nodes)
	Items    []T                  `json:"items,omitempty"`    // Actual items (leaf nodes only)
}

// SearchResp represents the response for search operations
type SearchResp[T any] struct {
	Items      []T                  `json:"items"`
	Groups     []SearchGroupNode[T] `json:"groups,omitempty"` // Tree-structured results (when group_by is specified)
	Pagination *PaginationInfo      `json:"pagination"`
}

// BuildGroupTree builds a tree structure from flat items based on groupBy fields.
// Items should already be sorted by groupBy fields (handled by the query layer).
func BuildGroupTree[T any, F ~string](items []T, groupBy []F) []SearchGroupNode[T] {
	if len(groupBy) == 0 || len(items) == 0 {
		return nil
	}
	groupByStr := make([]string, len(groupBy))
	for i, f := range groupBy {
		groupByStr[i] = string(f)
	}
	return buildGroupLevel(items, groupByStr, 0)
}

// buildGroupLevel recursively builds group nodes for a specific level of grouping
func buildGroupLevel[T any](items []T, groupBy []string, level int) []SearchGroupNode[T] {
	if level >= len(groupBy) || len(items) == 0 {
		return nil
	}

	field := groupBy[level]
	isLeaf := level == len(groupBy)-1

	// Group items by field value, preserving order from DB sort
	type group struct {
		key   string
		items []T
	}
	groupMap := make(map[string]*group)
	var groupOrder []*group

	for i := range items {
		val, _ := getJSONFieldValue(items[i], field)
		key := fmt.Sprintf("%v", val)

		if g, exists := groupMap[key]; exists {
			g.items = append(g.items, items[i])
		} else {
			g = &group{key: key, items: []T{items[i]}}
			groupMap[key] = g
			groupOrder = append(groupOrder, g)
		}
	}

	result := make([]SearchGroupNode[T], 0, len(groupOrder))
	for _, g := range groupOrder {
		node := SearchGroupNode[T]{
			Field: field,
			Value: g.key,
			Count: len(g.items),
		}
		if isLeaf {
			node.Items = g.items
		} else {
			node.Children = buildGroupLevel(g.items, groupBy, level+1)
		}
		result = append(result, node)
	}
	return result
}

// getJSONFieldValue retrieves a struct field value by its JSON tag name,
// including fields in embedded structs.
func getJSONFieldValue(item any, jsonTag string) (any, bool) {
	v := reflect.ValueOf(item)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		if v.IsNil() {
			return nil, false
		}
		v = v.Elem()
		t = v.Type()
	}

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		tag := sf.Tag.Get("json")
		if tag != "" {
			tagName := strings.Split(tag, ",")[0]
			if tagName == jsonTag {
				return v.Field(i).Interface(), true
			}
		}
		// Recurse into embedded structs
		if sf.Anonymous && sf.Type.Kind() == reflect.Struct {
			if val, found := getJSONFieldValue(v.Field(i).Interface(), jsonTag); found {
				return val, true
			}
		}
	}
	return nil, false
}
