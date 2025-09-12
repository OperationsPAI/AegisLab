package dto

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GenericResponse[T any] struct {
	Code      int    `json:"code"`                // Status code
	Message   string `json:"message"`             // Response message
	Data      T      `json:"data,omitempty"`      // Generic type data
	Timestamp int64  `json:"timestamp,omitempty"` // Response generation time
}

type GenericCreateResponse[T, E any] struct {
	CreatedCount int    `json:"created_count"`
	CreatedItems []T    `json:"created_items"`
	FailedCount  int    `json:"failed_count,omitempty"`
	FailedItems  []E    `json:"failed_items,omitempty"`
	Message      string `json:"message"`
}

type ListResp[T any] struct {
	Total int64 `json:"total"`
	Items []T   `json:"items"`
}

type PaginationResp[T any] struct {
	Total      int64 `json:"total"`
	TotalPages int64 `json:"total_pages,omitempty"`
	Items      []T   `json:"items"`
}

func NewPaginationResponse[T any](total int64, pageSize int, items []T) *PaginationResp[T] {
	totalPages := int64(0)
	if pageSize > 0 {
		totalPages = (total + int64(pageSize) - 1) / int64(pageSize)
	}

	return &PaginationResp[T]{
		Total:      total,
		TotalPages: totalPages,
		Items:      items,
	}
}

type Trace struct {
	TraceID    string `json:"trace_id"`
	HeadTaskID string `json:"head_task_id"`
	Index      int    `json:"index"`
}

type SubmitResp struct {
	GroupID string  `json:"group_id"`
	Traces  []Trace `json:"traces"`
}

func JSONResponse[T any](c *gin.Context, code int, message string, data T) {
	c.JSON(code, GenericResponse[T]{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

func SuccessResponse[T any](c *gin.Context, data T) {
	c.JSON(http.StatusOK, GenericResponse[T]{
		Code:      http.StatusOK,
		Message:   "Success",
		Data:      data,
		Timestamp: time.Now().Unix(),
	})
}

func ErrorResponse(c *gin.Context, code int, message string) {
	c.JSON(code, GenericResponse[any]{
		Code:      code,
		Message:   message,
		Timestamp: time.Now().Unix(),
	})
}

// AlgorithmResponse represents algorithm response for v2 API
type AlgorithmResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Image     string    `json:"image"`
	Tag       string    `json:"tag"`
	Command   string    `json:"command"`
	EnvVars   string    `json:"env_vars"`
	ProjectID int       `json:"project_id"`
	IsPublic  bool      `json:"is_public"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Related entities (only included when specifically requested)
	Project *ProjectResponse `json:"project,omitempty"`
}

// TaskResponse represents task response for v2 API
type TaskResponse struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	TraceID   string    `json:"trace_id"`
	GroupID   string    `json:"group_id,omitempty"`
	ProjectID int       `json:"project_id"`
	Immediate bool      `json:"immediate"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Logs      []string  `json:"logs,omitempty"` // Only included when specifically requested

	// Related entities (only included when specifically requested)
	Project *ProjectResponse `json:"project,omitempty"`
}

// ContainerResponse represents container response for v2 API
type ContainerResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Image     string    `json:"image"`
	Tag       string    `json:"tag"`
	Command   string    `json:"command"`
	EnvVars   string    `json:"env_vars"`
	UserID    int       `json:"user_id"`
	IsPublic  bool      `json:"is_public"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Related entities (only included when specifically requested)
	User *UserResponse `json:"user,omitempty"`
}

// ProjectResponse represents project response for v2 API
type ProjectResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPublic    bool      `json:"is_public"`
	Status      int       `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// Related entities (only included when specifically requested)
	Members []UserProjectResponse `json:"members,omitempty"` // Project members with their roles
}

// DatasetResponse represents dataset response for v2 API
type DatasetResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Size        int64     `json:"size"`
	FileCount   int       `json:"file_count"`
	DataSource  string    `json:"data_source"`
	Format      string    `json:"format"`
	Status      int       `json:"status"`
	IsPublic    bool      `json:"is_public"`
	DownloadURL string    `json:"download_url,omitempty"`
	Checksum    string    `json:"checksum,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TaskDetailResponse represents detailed task response including logs
type TaskDetailResponse struct {
	TaskResponse
	Logs []string `json:"logs"`
}
