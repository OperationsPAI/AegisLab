package dto

import "time"

// PaginationInfo represents pagination information in responses
type PaginationInfo struct {
	Page       int   `json:"page" example:"1"`
	Size       int   `json:"size" example:"20"`
	Total      int64 `json:"total" example:"100"`
	TotalPages int   `json:"total_pages" example:"5"`
}

// PaginationRequest represents pagination parameters in requests
type PaginationRequest struct {
	Page int `json:"page,omitempty" binding:"min=1" example:"1"`
	Size int `json:"size,omitempty" binding:"min=1,max=100" example:"20"`
}

// SortField represents a single sort field
type SortField struct {
	Field string `json:"field" binding:"required" example:"created_at"`
	Order string `json:"order" binding:"required,oneof=asc desc" example:"desc"`
}

// BaseResponse represents the standard API response structure
type BaseResponse struct {
	Code      int         `json:"code" example:"200"`
	Message   string      `json:"message" example:"success"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp time.Time   `json:"timestamp" example:"2024-01-01T12:00:00Z"`
}

// ListResponse represents a generic list response with pagination
type ListResponse struct {
	Items      interface{}    `json:"items"`
	Pagination PaginationInfo `json:"pagination"`
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
