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
