package middleware

import (
	"github.com/gin-gonic/gin"
)

func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Do something before the request (log, etc.)
		c.Next() // Process the rest of the request
		// Do something after the request (log response, etc.)
	}
}
