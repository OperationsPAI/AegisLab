package dto

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/gin-gonic/gin"
)

type GenericResponse[T any] struct {
	Code      int    `json:"code"`                // 状态码
	Message   string `json:"message"`             // 响应消息
	Data      T      `json:"data,omitempty"`      // 泛型类型的数据
	Timestamp int64  `json:"timestamp,omitempty"` // 响应生成时间
}

type PaginationResp[T any] struct {
	Total int64 `json:"total"`
	Data  []T   `json:"-"`
}

type Trace struct {
	TraceID    string `json:"trace_id"`
	HeadTaskID string `json:"head_task_id"`
}

type SubmitResp struct {
	GroupID string  `json:"group_id"`
	Traces  []Trace `json:"traces"`
}

func (p *PaginationResp[T]) MarshalJSON() ([]byte, error) {
	type Alias PaginationResp[T]

	// 获取类型 T 的实际类型
	var t T
	tType := reflect.TypeOf(t)
	if tType != nil && tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	// 处理指针类型
	if tType.Kind() == reflect.Ptr {
		tType = tType.Elem()
	}

	// 处理匿名类型或无效类型
	typeName := "item"
	if tType != nil {
		typeName = tType.Name()
	}

	snakeCase := utils.ToSnakeCase(typeName)
	dataKey := fmt.Sprintf("%ss", strings.Split(snakeCase, "_")[0])

	result := map[string]any{
		"total": p.Total,
		dataKey: p.Data,
	}

	return json.Marshal(result)
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
	c.JSON(http.StatusOK, GenericResponse[any]{
		Code:    code,
		Message: message,
	})
}
