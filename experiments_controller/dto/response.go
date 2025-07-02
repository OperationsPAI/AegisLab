package dto

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type GenericResponse[T any] struct {
	Code      int    `json:"code"`                // 状态码
	Message   string `json:"message"`             // 响应消息
	Data      T      `json:"data,omitempty"`      // 泛型类型的数据
	Timestamp int64  `json:"timestamp,omitempty"` // 响应生成时间
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
