package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TraceID() gin.HandlerFunc {
	return func(c *gin.Context) {
		traceID := uuid.New().String()
		c.Set("traceID", traceID)
		c.Writer.Header().Set("X-Trace-ID", traceID)
		c.Next()
	}
}
