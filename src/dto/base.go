package dto

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
	Code      int         `json:"code" example:"200"`
	Message   string      `json:"message" example:"success"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp" example:"2024-01-01T12:00:00Z"`
}

// ListResponse represents a generic list response with pagination
type ListResponse[T any] struct {
	Items      []T            `json:"items"`
	Pagination PaginationInfo `json:"pagination"`
}

// PaginationInfo represents pagination information in responses
type PaginationInfo struct {
	Page       int   `json:"page" example:"1"`
	Size       int   `json:"size" example:"20"`
	Total      int64 `json:"total" example:"100"`
	TotalPages int   `json:"total_pages" example:"5"`
}

// IDRequest represents a simple ID request
type IDRequest struct {
	ID int `uri:"id" binding:"required,min=1" example:"1"`
}

// BatchDeleteRequest represents batch deletion request
type BatchDeleteRequest struct {
	IDs []int `json:"ids" binding:"required,min=1" example:"[1,2,3]"`
}

// StatusUpdateRequest represents status update request
type StatusUpdateRequest struct {
	Status int `json:"status" binding:"required" example:"1"`
}

// ActivateRequest represents activation request
type ActivateRequest struct {
	IsActive bool `json:"is_active" binding:"required" example:"true"`
}
