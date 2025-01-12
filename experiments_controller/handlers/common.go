package handlers

import (
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
