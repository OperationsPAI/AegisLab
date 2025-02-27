package middleware

import (
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func SSEPath() gin.HandlerFunc {
	return func(c *gin.Context) {
		sseRegex := regexp.MustCompile(`^/stream(/.*)?$`)
		if sseRegex.MatchString(c.Request.URL.Path) {
			// 设置 SSE 响应头
			c.Header("Content-Type", "text/event-stream")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")
		}

		c.Next()
	}
}

func GroupID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" {
			groupID := uuid.New().String()
			c.Set("groupID", groupID)
			c.Writer.Header().Set("X-Group-ID", groupID)
		}

		c.Next()
	}
}
