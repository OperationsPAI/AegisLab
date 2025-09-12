package dto

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
