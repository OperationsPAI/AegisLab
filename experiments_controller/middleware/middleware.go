package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GroupID() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "POST" {
			groupID := uuid.New().String()
			c.Set("groupID", groupID)
			c.Writer.Header().Set("X-Group-ID", groupID)
			c.Next()
		}
	}
}
