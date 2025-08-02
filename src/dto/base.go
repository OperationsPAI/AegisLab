package dto

// PaginationInfo represents pagination information in responses
type PaginationInfo struct {
	Page       int   `json:"page" example:"1"`
	Size       int   `json:"size" example:"20"`
	Total      int64 `json:"total" example:"100"`
	TotalPages int   `json:"total_pages" example:"5"`
}
