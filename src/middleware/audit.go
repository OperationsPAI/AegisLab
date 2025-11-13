package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"aegis/consts"
	producer "aegis/service/prodcuer"
	"aegis/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const (
	respMsgField = "message"
)

// AuditMiddleware automatically logs all API requests for audit purposes
func AuditMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		startTime := time.Now()

		// Get user information (if authenticated)
		userID, _ := GetCurrentUserID(c)

		// Read request body (for recording details)
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// Use Writer wrapper to capture response
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		// Only log important operations (not GET requests, exclude health checks)
		if shouldAudit(c.Request.Method, c.FullPath()) {
			action := determineAction(c.Request.Method)
			resource := determineResource(c.FullPath())

			// Build details JSON
			details := map[string]interface{}{
				"method":      c.Request.Method,
				"path":        c.Request.URL.Path,
				"query":       c.Request.URL.RawQuery,
				"status_code": statusCode,
			}

			// Only include request body for non-sensitive operations and small payloads
			if len(requestBody) > 0 && len(requestBody) < 1024 && !isSensitivePath(c.FullPath()) {
				var reqData map[string]interface{}
				if err := json.Unmarshal(requestBody, &reqData); err == nil {
					// Remove sensitive fields
					sanitizeRequestData(reqData)
					details["request"] = reqData
				}
			}

			detailsJSON, _ := json.Marshal(details)

			// Determine status
			var errorMsg string
			if statusCode >= 400 {
				if len(blw.body.Bytes()) > 0 {
					var respData map[string]any
					if err := json.Unmarshal(blw.body.Bytes(), &respData); err == nil {
						if msg, ok := respData[respMsgField].(string); ok {
							errorMsg = msg
						}
					}
				}
			}

			// Async logging (don't block request)
			go func() {
				ipAddress := c.ClientIP()
				userAgent := c.GetHeader("User-Agent")
				duration := int(duration.Milliseconds())

				//TODO resource instance extraction

				if errorMsg != "" {
					if err := producer.LogFailedAction(ipAddress, userAgent, action, errorMsg, duration, userID, resource); err != nil {
						logrus.Errorf("Failed to log audit action: %v", err)
						return
					}
					return
				}

				if err := producer.LogUserAction(ipAddress, userAgent, action, string(detailsJSON), duration, userID, resource); err != nil {
					logrus.Errorf("Failed to log audit action: %v", err)
					return
				}
			}()
		}
	}
}

// bodyLogWriter is used to capture response body
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// shouldAudit determines if the request should be audited
func shouldAudit(method, path string) bool {
	// Don't log read operations (GET)
	if method == "GET" {
		return false
	}

	// Don't log health checks and monitoring endpoints
	excludedPaths := []string{
		"/system/health",
		"/system/monitor/metrics",
		"/system/monitor/namespace-locks",
	}

	for _, excluded := range excludedPaths {
		if path == excluded {
			return false
		}
	}

	return true
}

// determineAction determines the action type based on HTTP method and path
func determineAction(method string) string {
	switch method {
	case "POST":
		return "CREATE"
	case "PUT", "PATCH":
		return "UPDATE"
	case "DELETE":
		return "DELETE"
	default:
		return "ACCESS"
	}
}

// determineResource extracts the resource type from the path
func determineResource(path string) consts.ResourceName {
	// Simple implementation: extract the first segment after /api/v2/
	// Example: /api/v2/datasets/{id} -> datasets
	parts := strings.Split(strings.Trim(path, "/"), "/")
	var result string
	if len(parts) == 1 {
		result = parts[0]
	} else if len(parts) >= 3 {
		if parts[0] == "api" && strings.HasPrefix(parts[1], "v") {
			result = parts[2]
		} else {
			result = parts[1]
		}
	}

	if result != "" {
		resource := consts.ResourceName(utils.ToSingular(result))

		if resource == consts.ResourceContainer {
			if len(parts) >= 5 {
				resource = consts.ResourceContainerVersion
			}
		}
		if resource == consts.ResourceDataset {
			if len(parts) >= 5 {
				resource = consts.ResourceDatasetVersion
			}
		}

		return resource
	}

	return "unknown"
}

// isSensitivePath checks if the path contains sensitive data
func isSensitivePath(path string) bool {
	sensitivePaths := []string{
		"/auth/login",
		"/auth/register",
		"/users/password",
		"/users/change-password",
	}

	for _, sensitive := range sensitivePaths {
		if strings.Contains(path, sensitive) {
			return true
		}
	}
	return false
}

// sanitizeRequestData removes sensitive fields from request data
func sanitizeRequestData(data map[string]interface{}) {
	sensitiveFields := []string{"password", "token", "secret", "api_key", "apiKey"}
	for _, field := range sensitiveFields {
		if _, exists := data[field]; exists {
			data[field] = "***REDACTED***"
		}
	}
}
