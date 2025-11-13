package v2

import (
	"aegis/consts"
	"aegis/dto"
	producer "aegis/service/prodcuer"
	"aegis/utils"
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// GetTraceStream handles streaming of trace events via Server-Sent Events (SSE)
//
//	@Summary		Stream trace events in real-time
//	@Description	Establishes a Server-Sent Events (SSE) connection to stream trace logs and task execution events in real-time. Returns historical events first, then switches to live monitoring.
//	@Tags			Traces
//	@ID				stream_trace_events
//	@Accept			json
//	@Produce		text/event-stream
//	@Security		BearerAuth
//	@Param			trace_id	path		string						true	"Trace ID"
//	@Param			last_id		query		string						false	"Last event ID received"	default("0")
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid request"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid trace ID or invalid request format/parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/traces/{trace_id}/stream [get]
func GetTraceStream(c *gin.Context) {
	traceID := c.Param(consts.URLPathTraceID)
	if !utils.IsValidUUID(traceID) {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid trace ID")
		return
	}

	var req dto.GetTraceStreamReq
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

	streamKey := fmt.Sprintf(consts.StreamLogKey, traceID)
	logEntry := logrus.WithFields(logrus.Fields{
		"trace_id":   traceID,
		"stream_key": streamKey,
	})

	processor, err := producer.GetTraceStreamProcessor(ctx, traceID)
	if err != nil {
		logEntry.Errorf("Failed to initialize stream processor: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, fmt.Sprintf("Failed to initialize trace stream: %v", err))
		return
	}

	logEntry.Infof("Reading historical events from Stream")
	historicalMessages, err := producer.ReadTraceStreamMessages(ctx, streamKey, req.LastID, 100, 0)
	if err != nil && err != redis.Nil {
		logEntry.Errorf("failed to read historical events from redis: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read event history")
		return
	}

	if len(historicalMessages) > 0 {
		lastID, completed, err := sendSSEEvents(c, processor, historicalMessages)
		if err != nil {
			logEntry.Errorf("failed to send historical stream events of ID %s: %v", req.LastID, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "failed to send stream events")
			return
		}

		req.LastID = lastID
		if completed {
			logEntry.Info("Trace completed during historical events, closing stream connection")
			return
		}
	}

	logEntry.Infof("Switching to real-time event monitoring from ID: %s", req.LastID)
	for {
		select {
		case <-c.Done():
			logEntry.Info("Request context done")
			return

		default:
			newMessages, err := producer.ReadTraceStreamMessages(ctx, streamKey, req.LastID, 10, time.Second)
			if err != nil && err != redis.Nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					logEntry.Infof("Context done while reading stream: %v", err)
					return
				}

				logEntry.Errorf("Error reading stream: %v", err)
				dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read stream events")
				return
			}

			if err == redis.Nil {
				logEntry.Info("No new messages, continuing")
			}

			lastID, completed, err := sendSSEEvents(c, processor, newMessages)
			if err != nil {
				logEntry.Errorf("failed to send stream events of ID %s: %v", lastID, err)
				return
			}

			req.LastID = lastID
			if completed {
				logEntry.Info("Trace completed, closing stream connection")
				return
			}

			logrus.Info("Sent SSE messages, lastID:", lastID)
		}
	}
}

// sendSSEEvents processes and sends stream messages as SSE events
func sendSSEEvents(c *gin.Context, processor *producer.StreamProcessor, streams []redis.XStream) (string, bool, error) {
	lastID, streamEvent, err := processor.ProcessMessageForSSE(streams[0].Messages[0])
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

	completed := processor.IsCompleted()
	if completed {
		c.SSEvent(string(consts.EventEnd), nil)
		c.Writer.Flush()
	}

	return lastID, completed, nil
}
