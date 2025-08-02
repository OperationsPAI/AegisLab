package dto

<<<<<<< HEAD
import "time"

// BaseEntity represents common fields for all entities
type BaseEntity struct {
	ID        int       `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BaseRequest represents common request fields
type BaseRequest struct {
	Page int `json:"page,omitempty" form:"page" binding:"min=1" example:"1"`
	Size int `json:"size,omitempty" form:"size" binding:"min=1,max=100" example:"20"`
}

// BaseResponse represents common response fields
type BaseResponse struct {
	Code      int       `json:"code" example:"200"`
	Message   string    `json:"message" example:"success"`
	Data      any       `json:"data,omitempty"`
	Timestamp time.Time `json:"timestamp" example:"2024-01-01T12:00:00Z"`
}

// ListResponse represents a generic list response with pagination
type ListResponse[T any] struct {
	Items      []T            `json:"items"`
	Pagination PaginationInfo `json:"pagination"`
}

=======
>>>>>>> e80cacbd880fbb4043936bbdf42dcf6ff48f2812
// PaginationInfo represents pagination information in responses
type PaginationInfo struct {
	Page       int   `json:"page" example:"1"`
	Size       int   `json:"size" example:"20"`
	Total      int64 `json:"total" example:"100"`
	TotalPages int   `json:"total_pages" example:"5"`
}
