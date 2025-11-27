package dto

import (
	"fmt"
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

// SortOption represents a sort option
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

// SearchReq represents a complex search request
type SearchReq struct {
	// Pagination
	PaginationReq

	// Filters
	Filters []SearchFilter `json:"filters,omitempty"`

	// Sort
	Sort []SortOption `json:"sort,omitempty"`

	// Search keyword (for general text search)
	Keyword string `json:"keyword,omitempty" form:"keyword"`

	// Include/Exclude fields
	IncludeFields []string `json:"include_fields,omitempty"`
	ExcludeFields []string `json:"exclude_fields,omitempty"`

	// Include related entities
	Includes []string `json:"includes,omitempty" form:"include"`
}

// GetOffset calculates the offset for pagination
func (sr *SearchReq) GetOffset() int {
	return (sr.Page - 1) * int(sr.Size)
}

// HasFilter checks if a specific filter exists
func (sr *SearchReq) HasFilter(field string) bool {
	for _, filter := range sr.Filters {
		if filter.Field == field {
			return true
		}
	}
	return false
}

// GetFilter gets a specific filter by field name
func (sr *SearchReq) GetFilter(field string) *SearchFilter {
	for _, filter := range sr.Filters {
		if filter.Field == field {
			return &filter
		}
	}
	return nil
}

// AddFilter adds a new filter
func (sr *SearchReq) AddFilter(field string, operator FilterOperator, value any) {
	// Convert value to string
	valueStr := fmt.Sprintf("%v", value)

	sr.Filters = append(sr.Filters, SearchFilter{
		Field:    field,
		Operator: operator,
		Value:    valueStr,
	})
}

// AddInclude adds a new include option
func (sr *SearchReq) AddInclude(field string) {
	sr.Includes = append(sr.Includes, field)
}

// AddSort adds a new sort option
func (sr *SearchReq) AddSort(field string, direction SortDirection) {
	sr.Sort = append(sr.Sort, SortOption{
		Field:     field,
		Direction: direction,
	})
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

// AdvancedSearchReq extends SearchRequest with common filter shortcuts
type AdvancedSearchReq struct {
	PaginationReq
	Sort []SortOption `json:"sort" binding:"omitempty"`

	Statuses  []consts.StatusType `json:"status" binding:"omitempty"`
	CreatedAt *DateRange          `json:"created_at" binding:"omitempty"`
	UpdatedAt *DateRange          `json:"updated_at" binding:"omitempty"`
}

func (req *AdvancedSearchReq) Validate() error {
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

// ConvertAdvancedToSearch converts AdvancedSearchRequest to SearchRequest with additional filters
func (req *AdvancedSearchReq) ConvertAdvancedToSearch() *SearchReq {
	sr := &SearchReq{}

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

// AlgorithmSearchReq represents algorithm search request
type AlgorithmSearchReq struct {
	AdvancedSearchReq

	// Algorithm-specific filters
	Name  *string `json:"name,omitempty"`
	Image *string `json:"image,omitempty"`
	Tag   *string `json:"tag,omitempty"`
	Type  *string `json:"type,omitempty"`
}

// ConvertToSearchRequest converts AlgorithmSearchRequest to SearchRequest
func (req *AlgorithmSearchReq) ConvertToSearchRequest() *SearchReq {
	sr := req.AdvancedSearchReq.ConvertAdvancedToSearch()

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

// SearchResp represents the response for search operations
type SearchResp[T any] struct {
	Items      []T             `json:"items"`
	Pagination *PaginationInfo `json:"pagination"`
}
