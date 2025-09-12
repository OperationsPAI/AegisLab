package dto

import (
	"fmt"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
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
	Field    string         `json:"field" binding:"required"`                    // Field name
	Operator FilterOperator `json:"operator" binding:"required"`                 // Operator
	Value    string         `json:"value" swaggertype:"string"`                  // Value (can be string, number, boolean, etc.)
	Values   []string       `json:"values,omitempty" swaggertype:"array,string"` // Multiple values (for IN operations etc.)
}

// SortOption represents a sort option
type SortOption struct {
	Field     string        `json:"field" binding:"required"`     // Sort field
	Direction SortDirection `json:"direction" binding:"required"` // Sort direction
}

// SearchRequest represents a complex search request
type SearchRequest struct {
	// Pagination
	Page int `json:"page" form:"page" binding:"min=1"`
	Size int `json:"size" form:"size" binding:"min=1,max=1000"`

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
	Include []string `json:"include,omitempty" form:"include"`
}

// DateRange represents a date range filter
type DateRange struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// NumberRange represents a number range filter
type NumberRange struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

// AdvancedSearchRequest extends SearchRequest with common filter shortcuts
type AdvancedSearchRequest struct {
	SearchRequest

	// Common filters shortcuts
	CreatedAt *DateRange `json:"created_at,omitempty"`
	UpdatedAt *DateRange `json:"updated_at,omitempty"`
	Status    []int      `json:"status,omitempty"`
	IsActive  *bool      `json:"is_active,omitempty"`
	UserID    *int       `json:"user_id,omitempty"`
	ProjectID *int       `json:"project_id,omitempty"`
}

// SearchResponse represents the response for search operations
type SearchResponse[T any] struct {
	Items      []T             `json:"items"`
	Pagination *PaginationInfo `json:"pagination"`
	Filters    []SearchFilter  `json:"applied_filters,omitempty"`
	Sort       []SortOption    `json:"applied_sort,omitempty"`
}

// ValidateSearchRequest validates the search request
func (sr *SearchRequest) ValidateSearchRequest() error {
	// Set defaults
	if sr.Page == 0 {
		sr.Page = 1
	}
	if sr.Size == 0 {
		sr.Size = 20
	}

	// Validate filters
	for _, filter := range sr.Filters {
		if filter.Field == "" {
			return fmt.Errorf("filter field cannot be empty")
		}
		if filter.Operator == "" {
			return fmt.Errorf("filter operator cannot be empty")
		}

		// Validate operator-specific requirements
		switch filter.Operator {
		case OpIn, OpNotIn:
			if len(filter.Values) == 0 {
				return fmt.Errorf("'%s' operator requires values array", filter.Operator)
			}
		case OpIsNull, OpIsNotNull:
			// These operators don't need values
		case OpDateBetween:
			if len(filter.Values) != 2 {
				return fmt.Errorf("'%s' operator requires exactly 2 values", filter.Operator)
			}
		default:
			if filter.Value == "" {
				return fmt.Errorf("'%s' operator requires a value", filter.Operator)
			}
		}
	}

	// Validate sort options
	for _, sort := range sr.Sort {
		if sort.Field == "" {
			return fmt.Errorf("sort field cannot be empty")
		}
		if sort.Direction != SortASC && sort.Direction != SortDESC {
			return fmt.Errorf("sort direction must be 'asc' or 'desc'")
		}
	}

	return nil
}

// GetOffset calculates the offset for pagination
func (sr *SearchRequest) GetOffset() int {
	return (sr.Page - 1) * sr.Size
}

// HasFilter checks if a specific filter exists
func (sr *SearchRequest) HasFilter(field string) bool {
	for _, filter := range sr.Filters {
		if filter.Field == field {
			return true
		}
	}
	return false
}

// GetFilter gets a specific filter by field name
func (sr *SearchRequest) GetFilter(field string) *SearchFilter {
	for _, filter := range sr.Filters {
		if filter.Field == field {
			return &filter
		}
	}
	return nil
}

// AddFilter adds a new filter
func (sr *SearchRequest) AddFilter(field string, operator FilterOperator, value interface{}) {
	// Convert value to string
	valueStr := fmt.Sprintf("%v", value)

	sr.Filters = append(sr.Filters, SearchFilter{
		Field:    field,
		Operator: operator,
		Value:    valueStr,
	})
}

// AddSort adds a new sort option
func (sr *SearchRequest) AddSort(field string, direction SortDirection) {
	sr.Sort = append(sr.Sort, SortOption{
		Field:     field,
		Direction: direction,
	})
}

// ConvertAdvancedToSearch converts AdvancedSearchRequest to SearchRequest with additional filters
func (asr *AdvancedSearchRequest) ConvertAdvancedToSearch() *SearchRequest {
	sr := &asr.SearchRequest

	// Convert common filters to SearchFilter format
	if asr.CreatedAt != nil {
		if asr.CreatedAt.From != nil && asr.CreatedAt.To != nil {
			sr.AddFilter("created_at", OpDateBetween, []interface{}{asr.CreatedAt.From, asr.CreatedAt.To})
		} else if asr.CreatedAt.From != nil {
			sr.AddFilter("created_at", OpDateAfter, asr.CreatedAt.From)
		} else if asr.CreatedAt.To != nil {
			sr.AddFilter("created_at", OpDateBefore, asr.CreatedAt.To)
		}
	}

	if asr.UpdatedAt != nil {
		if asr.UpdatedAt.From != nil && asr.UpdatedAt.To != nil {
			sr.AddFilter("updated_at", OpDateBetween, []interface{}{asr.UpdatedAt.From, asr.UpdatedAt.To})
		} else if asr.UpdatedAt.From != nil {
			sr.AddFilter("updated_at", OpDateAfter, asr.UpdatedAt.From)
		} else if asr.UpdatedAt.To != nil {
			sr.AddFilter("updated_at", OpDateBefore, asr.UpdatedAt.To)
		}
	}

	if len(asr.Status) > 0 {
		values := make([]string, len(asr.Status))
		for i, v := range asr.Status {
			values[i] = fmt.Sprintf("%v", v)
		}
		sr.Filters = append(sr.Filters, SearchFilter{
			Field:    "status",
			Operator: OpIn,
			Values:   values,
		})
	}

	if asr.IsActive != nil {
		sr.AddFilter("is_active", OpEqual, *asr.IsActive)
	}

	if asr.UserID != nil {
		sr.AddFilter("user_id", OpEqual, *asr.UserID)
	}

	if asr.ProjectID != nil {
		sr.AddFilter("project_id", OpEqual, *asr.ProjectID)
	}

	return sr
}

// TaskSearchRequest represents task search request
type TaskSearchRequest struct {
	AdvancedSearchRequest

	// Task-specific filters
	TaskID    *string `json:"task_id,omitempty"`
	TraceID   *string `json:"trace_id,omitempty"`
	GroupID   *string `json:"group_id,omitempty"`
	TaskType  *string `json:"task_type,omitempty"`
	Status    *string `json:"status,omitempty"`
	Immediate *bool   `json:"immediate,omitempty"`
}

// ConvertToSearchRequest converts TaskSearchRequest to SearchRequest
func (tsr *TaskSearchRequest) ConvertToSearchRequest() *SearchRequest {
	sr := tsr.AdvancedSearchRequest.ConvertAdvancedToSearch()

	// Add task-specific filters
	if tsr.TaskID != nil {
		sr.AddFilter("id", OpEqual, *tsr.TaskID)
	}
	if tsr.TraceID != nil {
		sr.AddFilter("trace_id", OpEqual, *tsr.TraceID)
	}
	if tsr.GroupID != nil {
		sr.AddFilter("group_id", OpEqual, *tsr.GroupID)
	}
	if tsr.TaskType != nil {
		sr.AddFilter("type", OpEqual, *tsr.TaskType)
	}
	if tsr.Status != nil {
		sr.AddFilter("status", OpEqual, *tsr.Status)
	}
	if tsr.Immediate != nil {
		sr.AddFilter("immediate", OpEqual, *tsr.Immediate)
	}

	return sr
}

// AlgorithmSearchRequest represents algorithm search request
type AlgorithmSearchRequest struct {
	AdvancedSearchRequest

	// Algorithm-specific filters
	Name  *string `json:"name,omitempty"`
	Image *string `json:"image,omitempty"`
	Tag   *string `json:"tag,omitempty"`
	Type  *string `json:"type,omitempty"`
}

// ConvertToSearchRequest converts AlgorithmSearchRequest to SearchRequest
func (asr *AlgorithmSearchRequest) ConvertToSearchRequest() *SearchRequest {
	sr := asr.AdvancedSearchRequest.ConvertAdvancedToSearch()

	// Add algorithm-specific filters
	if asr.Name != nil {
		sr.AddFilter("name", OpLike, *asr.Name)
	}
	if asr.Image != nil {
		sr.AddFilter("image", OpLike, *asr.Image)
	}
	if asr.Tag != nil {
		sr.AddFilter("tag", OpEqual, *asr.Tag)
	}
	if asr.Type != nil {
		sr.AddFilter("type", OpEqual, *asr.Type)
	}

	// Default to only active algorithms
	if !sr.HasFilter("status") {
		sr.AddFilter("status", OpEqual, consts.ContainerEnabled)
	}

	return sr
}

// ContainerSearchRequest represents container search request
type ContainerSearchRequest struct {
	AdvancedSearchRequest

	// Container-specific filters
	Name    *string `json:"name,omitempty"`
	Image   *string `json:"image,omitempty"`
	Tag     *string `json:"tag,omitempty"`
	Type    *string `json:"type,omitempty"`
	Command *string `json:"command,omitempty"`
	Status  *int    `json:"status,omitempty"`
}

// ConvertToSearchRequest converts ContainerSearchRequest to SearchRequest
func (csr *ContainerSearchRequest) ConvertToSearchRequest() *SearchRequest {
	sr := csr.AdvancedSearchRequest.ConvertAdvancedToSearch()

	// Add container-specific filters
	if csr.Name != nil {
		sr.AddFilter("name", OpLike, *csr.Name)
	}
	if csr.Image != nil {
		sr.AddFilter("image", OpLike, *csr.Image)
	}
	if csr.Tag != nil {
		sr.AddFilter("tag", OpEqual, *csr.Tag)
	}
	if csr.Type != nil {
		sr.AddFilter("type", OpEqual, *csr.Type)
	}
	if csr.Command != nil {
		sr.AddFilter("command", OpLike, *csr.Command)
	}

	return sr
}
