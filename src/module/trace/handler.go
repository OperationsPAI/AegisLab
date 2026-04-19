package trace

import (
	"aegis/httpx"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"aegis/consts"
	"aegis/dto"
	"aegis/utils"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	service HandlerService
}

func NewHandler(service HandlerService) *Handler {
	return &Handler{service: service}
}

// GetTrace handles getting a single trace by ID
//
//	@Summary		Get trace by ID
//	@Description	Get detailed information about a specific trace, including its associated tasks
//	@Tags			Traces
//	@ID				get_trace_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			trace_id	path		string									true	"Trace ID"
//	@Success		200			{object}	dto.GenericResponse[TraceDetailResp]	"Trace retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid trace ID"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]				"Trace not found"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/traces/{trace_id} [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) GetTrace(c *gin.Context) {
	traceID := c.Param(consts.URLPathTraceID)
	if !utils.IsValidUUID(traceID) {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid trace ID")
		return
	}

	resp, err := h.service.GetTrace(c.Request.Context(), traceID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListTraces handles listing traces with filtering
//
//	@Summary		List traces
//	@Description	Get a list of traces with filtering via query parameters
//	@Tags			Traces
//	@ID				list_traces
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int												false	"Page number"	default(1)
//	@Param			size		query		int												false	"Page size"		default(20)
//	@Param			trace_type	query		consts.TraceType								false	"Filter by trace type"
//	@Param			group_id	query		string											false	"Filter by group ID (uuid format)"
//	@Param			project_id	query		int												false	"Filter by project ID"
//	@Param			state		query		consts.TraceState								false	"Filter by state"
//	@Param			status		query		consts.StatusType								false	"Filter by status"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[TraceResp]]	"Traces retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/traces [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListTraces(c *gin.Context) {
	var req ListTraceReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListTraces(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// GetTraceStream handles streaming of trace events via Server-Sent Events (SSE)
//
//	@Summary		Stream trace events in real-time
//	@Description	Establishes a Server-Sent Events (SSE) connection to stream trace logs and task execution events in real-time. Returns historical events first, then switches to live monitoring.
//	@Tags			Traces
//	@ID				get_trace_events
//	@Produce		text/event-stream
//	@Security		BearerAuth
//	@Param			trace_id	path		string						true	"Trace ID"
//	@Param			last_id		query		string						false	"Last event ID received"	default("0")
//	@Success		200			{string}	string						"A stream of event messages (e.g., log entries, task status updates)."
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid trace ID or invalid request format/parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/traces/{trace_id}/stream [get]
//	@x-request-type	{"stream":"true"}
//	@x-api-type		{"portal":"true"}
func (h *Handler) GetTraceStream(c *gin.Context) {
	traceID := c.Param(consts.URLPathTraceID)
	if !utils.IsValidUUID(traceID) {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid trace ID")
		return
	}

	var req GetTraceStreamReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format")
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	if c.IsAborted() {
		return
	}

	streamKey := fmt.Sprintf(consts.StreamTraceLogKey, traceID)
	logEntry := logrus.WithFields(logrus.Fields{
		"trace_id":   traceID,
		"stream_key": streamKey,
	})

	processor, err := h.service.GetTraceStreamProcessor(ctx, traceID)
	if err != nil {
		logEntry.Errorf("Failed to initialize stream processor: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize trace stream: %v", err))
		return
	}

	historicalMessages, err := h.service.ReadTraceStreamMessages(ctx, streamKey, req.LastID, 100, 0)
	if err != nil {
		logEntry.Errorf("failed to read historical events from redis: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read event history")
		return
	}

	if len(historicalMessages) > 0 {
		lastID, completed, err := sendTraceSSEEvents(c, processor, historicalMessages)
		if err != nil {
			logEntry.Errorf("failed to send historical stream events of ID %s: %v", req.LastID, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "failed to send stream events")
			return
		}

		if completed {
			return
		}

		req.LastID = lastID
	}

	for {
		select {
		case <-c.Done():
			return
		default:
			newMessages, err := h.service.ReadTraceStreamMessages(ctx, streamKey, req.LastID, 10, time.Second)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}

				logEntry.Errorf("Error reading stream: %v", err)
				dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read stream events")
				return
			}

			if len(newMessages) == 0 {
				continue
			}

			lastID, completed, err := sendTraceSSEEvents(c, processor, newMessages)
			if err != nil {
				logEntry.Errorf("failed to send stream events of ID %s: %v", lastID, err)
				return
			}

			req.LastID = lastID
			if completed {
				time.Sleep(time.Second)
				return
			}
		}
	}
}

func sendTraceSSEEvents(c *gin.Context, processor *StreamProcessor, streams []redis.XStream) (string, bool, error) {
	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return "", false, fmt.Errorf("no messages to process")
	}

	var lastID string
	for _, msg := range streams[0].Messages {
		id, streamEvent, err := processor.ProcessMessageForSSE(msg)
		lastID = id
		if err != nil {
			c.SSEvent(string(consts.EventEnd), nil)
			c.Writer.Flush()
			return lastID, true, err
		}

		c.Render(-1, sse.Event{
			Id:    lastID,
			Event: string(consts.EventUpdate),
			Data:  streamEvent,
		})
		c.Writer.Flush()
	}

	completed := processor.IsCompleted()
	if completed {
		c.SSEvent(string(consts.EventEnd), nil)
		c.Writer.Flush()
	}

	return lastID, completed, nil
}
