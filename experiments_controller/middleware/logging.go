package middleware

import (
	"github.com/gin-gonic/gin"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 在请求之前做些事情（记录日志等）
		c.Next() // 处理请求的其余部分
		// 在请求之后做些事情（记录响应等）
	}
}
