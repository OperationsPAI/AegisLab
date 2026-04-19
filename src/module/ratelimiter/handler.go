package ratelimiter

import (
	"fmt"
	"net/http"

	"aegis/consts"
	"aegis/dto"
	"aegis/httpx"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// ListRateLimiters
//
//	@Summary		List rate limiters
//	@Description	List all token-bucket rate limiters and their holders.
//	@Tags			RateLimiters
//	@ID				list_rate_limiters
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GenericResponse[RateLimiterListResp]
//	@Router			/api/v2/rate-limiters [get]
//	@x-api-type		{"sdk":"true"}
func (h *Handler) ListRateLimiters(c *gin.Context) {
	resp, err := h.service.List(c.Request.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to list rate limiters")
		httpx.HandleServiceError(c, fmt.Errorf("%w: %v", consts.ErrInternal, err))
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
func (h *Handler) ResetRateLimiter(c *gin.Context) {
	bucket := c.Param("bucket")
	if bucket == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "bucket is required")
		return
	}
	if err := h.service.Reset(c.Request.Context(), bucket); err != nil {
		if httpx.HandleServiceError(c, err) {
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
//	@Success		200	{object}	dto.GenericResponse[RateLimiterGCResp]
//	@Router			/api/v2/rate-limiters/gc [post]
//	@x-api-type		{"sdk":"true"}
func (h *Handler) GCRateLimiters(c *gin.Context) {
	released, buckets, err := h.service.GC(c.Request.Context())
	if err != nil {
		logrus.WithError(err).Error("Failed to gc rate limiters")
		httpx.HandleServiceError(c, fmt.Errorf("%w: %v", consts.ErrInternal, err))
		return
	}
	dto.JSONResponse(c, http.StatusOK, "Garbage collection complete", &RateLimiterGCResp{
		Released:       released,
		TouchedBuckets: buckets,
	})
}
