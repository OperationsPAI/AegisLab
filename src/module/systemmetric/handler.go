package systemmetricmodule

import (
	"net/http"

	"aegis/dto"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service HandlerService
}

func NewHandler(service HandlerService) *Handler {
	return &Handler{service: service}
}

// GetSystemMetrics retrieves current system metrics
//
//	@Summary		Get current system metrics
//	@Description	Get current CPU, memory, and disk usage metrics
//	@Tags			System
//	@ID				get_system_metrics
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GenericResponse[SystemMetricsResp]	"System metrics retrieved successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/system/metrics [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) GetSystemMetrics(c *gin.Context) {
	resp, err := h.service.GetSystemMetrics(c.Request.Context())
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get system metrics: "+err.Error())
		return
	}

	dto.JSONResponse(c, http.StatusOK, "System metrics retrieved successfully", resp)
}

// GetSystemMetricsHistory retrieves historical system metrics (24 hours)
//
//	@Summary		Get historical system metrics
//	@Description	Get 24-hour historical CPU and memory usage metrics
//	@Tags			System
//	@ID				get_system_metrics_history
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	dto.GenericResponse[SystemMetricsHistoryResp]	"System metrics history retrieved successfully"
//	@Failure		401	{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		500	{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/system/metrics/history [get]
//	@x-api-type		{"admin":"true"}
func (h *Handler) GetSystemMetricsHistory(c *gin.Context) {
	resp, err := h.service.GetSystemMetricsHistory(c.Request.Context())
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get system metrics history: "+err.Error())
		return
	}

	dto.JSONResponse(c, http.StatusOK, "System metrics history retrieved successfully", resp)
}
