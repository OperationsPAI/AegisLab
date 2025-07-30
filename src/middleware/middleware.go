package middleware

import (
	"regexp"

	"context"

	"github.com/google/uuid"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	TracerKey      = "otel-tracer"
	SpanContextKey = "otel-span-context"
)

func SSEPath() gin.HandlerFunc {
	return func(c *gin.Context) {
		sseRegex := regexp.MustCompile(`^/stream(/.*)?$`)
		if sseRegex.MatchString(c.Request.URL.Path) {
			// Set SSE response headers
			c.Writer.Header().Set("Content-Type", "text/event-stream")
			c.Writer.Header().Set("Cache-Control", "no-cache")
			c.Writer.Header().Set("Connection", "keep-alive")
			c.Writer.Header().Set("Transfer-Encoding", "chunked")
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

func TracerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		groupID := c.GetString("groupID")

		// Use request method and path for span name
		spanName := c.Request.Method + " " + c.Request.URL.Path

		ctx, span := otel.Tracer("rcabench/group").Start(
			context.Background(),
			spanName,
			trace.WithAttributes(
				attribute.String("group_id", groupID),
			),
		)
		defer span.End()

		c.Set(SpanContextKey, ctx)

		c.Next()
	}
}
