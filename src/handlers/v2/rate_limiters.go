package v2

import (
	"fmt"
	"net/http"

	"aegis/consts"
	"aegis/dto"
	"aegis/handlers"
	producer "aegis/service/producer"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ListRateLimiters
//
//	@Summary		List rate limiters
//	@Description	List all token-bucket rate limiters and their holders.
//	@Tags			RateLimiters
//	@ID				list_rate_limiters
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GenericResponse[dto.RateLimiterListResp]
//	@Router			/api/v2/rate-limiters [get]
//	@x-api-type		{"sdk":"true"}
func ListRateLimiters(c *gin.Context) {
	resp, err := producer.ListRateLimiters(c.Request.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list rate limiters")
		handlers.HandleServiceError(c, fmt.Errorf("%w: %v", consts.ErrInternal, err))
		return
	}
	dto.JSONResponse(c, http.StatusOK, "Rate limiters retrieved successfully", resp)
}

// ResetRateLimiter
//
//	@Summary		Reset a rate limiter bucket
//	@Description	Delete the given token-bucket key from Redis. Admin-only.
//	@Tags			RateLimiters
//	@ID				reset_rate_limiter
//	@Produce		json
//	@Security		BearerAuth
//	@Param			bucket	path	string	true	"Bucket short name"
//	@Success		200	{object}	dto.GenericResponse[any]
//	@Router			/api/v2/rate-limiters/{bucket} [delete]
//	@x-api-type		{"sdk":"true"}
func ResetRateLimiter(c *gin.Context) {
	bucket := c.Param("bucket")
	if bucket == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "bucket is required")
		return
	}
	if err := producer.ResetRateLimiter(c.Request.Context(), bucket); err != nil {
		if handlers.HandleServiceError(c, err) {
			return
		}
	}
	dto.JSONResponse(c, http.StatusOK, "Rate limiter reset successfully", gin.H{"bucket": bucket})
}

// GCRateLimiters
//
//	@Summary		Garbage-collect leaked tokens across all rate limiters
//	@Description	Scan every known bucket and release tokens held by terminal-state tasks. Admin-only.
//	@Tags			RateLimiters
//	@ID				gc_rate_limiters
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GenericResponse[dto.RateLimiterGCResp]
//	@Router			/api/v2/rate-limiters/gc [post]
//	@x-api-type		{"sdk":"true"}
func GCRateLimiters(c *gin.Context) {
	released, buckets, err := producer.GCRateLimiters(c.Request.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to gc rate limiters")
		handlers.HandleServiceError(c, fmt.Errorf("%w: %v", consts.ErrInternal, err))
		return
	}
	dto.JSONResponse(c, http.StatusOK, "Garbage collection complete", &dto.RateLimiterGCResp{
		Released:       released,
		TouchedBuckets: buckets,
	})
}
