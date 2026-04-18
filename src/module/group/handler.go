package groupmodule

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

// GetGroupStats handles retrieval of group trace statistics
//
//	@Summary		Get statistics for a group of traces
//	@Description	Retrieves statistics such as total traces, average duration, and state distribution for a specified group of traces.
//	@Tags			Groups
//	@ID				get_group_stats
//	@Produce		json
//	@Security		BearerAuth
//	@Param			group_id	path		string								true	"Group ID (UUID)"
//	@Success		200			{object}	dto.GenericResponse[GroupStats]		"Group trace statistics"
//	@Failure		400			{object}	dto.GenericResponse[any]			"Invalid request format/parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/groups/{group_id}/stats [get]
//	@x-api-type		{}
func (h *Handler) GetGroupStats(c *gin.Context) {
	groupID := c.Param(consts.URLPathGroupID)
	if !utils.IsValidUUID(groupID) {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid group ID")
		return
	}

	req := GetGroupStatsReq{GroupID: groupID}
	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	stats, err := h.service.GetGroupStats(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, stats)
}

// GetGroupStream handles streaming of group-level trace completion events via SSE.
// It pushes a lightweight event whenever a trace in the group reaches a terminal state
// (Completed/Failed), enabling real-time progress tracking for batch operations.
//
//	@Summary		Stream group trace events in real-time
//	@Description	Establishes an SSE connection to stream trace completion/failure events for all traces in a group. Each event contains trace_id, state, and last_event.
//	@Tags			Groups
//	@ID				get_group_stream
//	@Produce		text/event-stream
//	@Security		BearerAuth
//	@Param			group_id	path		string						true	"Group ID (UUID)"
//	@Param			last_id		query		string						false	"Last event ID received"	default("0")
//	@Success		200			{string}	string						"A stream of group trace events"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid group ID or request parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/groups/{group_id}/stream [get]
//	@x-request-type	{"stream":"true"}
//	@x-api-type		{}
func (h *Handler) GetGroupStream(c *gin.Context) {
	groupID := c.Param(consts.URLPathGroupID)
	if !utils.IsValidUUID(groupID) {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid group ID")
		return
	}

	var req GetGroupStreamReq
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

	streamKey := fmt.Sprintf(consts.StreamGroupLogKey, groupID)
	logEntry := logrus.WithFields(logrus.Fields{
		"group_id":   groupID,
		"stream_key": streamKey,
	})

	processor, err := h.service.NewGroupStreamProcessor(ctx, groupID)
	if err != nil {
		logEntry.Errorf("Failed to initialize group stream processor: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize group stream: %v", err))
		return
	}

	historical, err := h.service.ReadGroupStreamMessages(ctx, streamKey, req.LastID, 100, 0)
	if err != nil {
		logEntry.Errorf("Failed to read historical group stream events: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to read group event history")
		return
	}

	if len(historical) > 0 {
		lastID, completed, err := sendGroupSSEEvents(c, processor, historical)
		if err != nil {
			logEntry.Errorf("Failed to send historical group stream events: %v", err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to send group stream events")
			return
		}

		if completed {
			return
		}

		req.LastID = lastID
	}

	for {
		select {
		case <-ctx.Done():
			return
		default:
			newMessages, err := h.service.ReadGroupStreamMessages(ctx, streamKey, req.LastID, 10, time.Second)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					return
				}

				logEntry.Errorf("Error reading group stream: %v", err)
				dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to read group stream events")
				return
			}

			if len(newMessages) == 0 {
				continue
			}

			lastID, completed, err := sendGroupSSEEvents(c, processor, newMessages)
			if err != nil {
				logEntry.Errorf("Failed to send group stream events: %v", err)
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

func sendGroupSSEEvents(c *gin.Context, processor *GroupStreamProcessor, streams []redis.XStream) (string, bool, error) {
	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return "", false, fmt.Errorf("no messages to process")
	}

	var lastID string
	for _, msg := range streams[0].Messages {
		lastID = msg.ID

		event, err := processor.ProcessGroupMessage(msg)
		if err != nil {
			logrus.Errorf("Failed to process group stream message %s: %v", msg.ID, err)
			continue
		}

		c.Render(-1, sse.Event{
			Id:    lastID,
			Event: string(consts.EventUpdate),
			Data:  event,
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
