package handlers

import (
	"net/http"
	"reflect"
	"time"

	"github.com/gin-gonic/gin"
)

type GenericResponse[T any] struct {
	Code      int    `json:"code"`                // 状态码
	Message   string `json:"message"`             // 响应消息
	Data      T      `json:"data"`                // 泛型类型的数据
	Timestamp int64  `json:"timestamp,omitempty"` // 响应生成时间
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

func StructToMap(obj any) map[string]any {
	result := make(map[string]any)
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// 获取 JSON 标签名，如果没有则用字段名
		tag := field.Tag.Get("json")
		if tag == "" {
			tag = field.Name
		}
		result[tag] = v.Field(i).Interface()
	}

	return result
}
