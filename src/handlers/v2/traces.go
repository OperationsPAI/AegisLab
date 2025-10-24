package v2

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/repository"
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
//	@Summary      Stream trace events in real-time
//	@Description  Establishes a Server-Sent Events (SSE) connection to stream trace logs and task execution events in real-time. Returns historical events first, then switches to live monitoring.
//	@Tags         Traces
//	@Accept       json
//	@Produce      text/event-stream
//	@Security     BearerAuth
//	@Param        id  		path      string  true   "Trace ID"
//	@Param        last_id   query     string  false  "Last event ID received" default("0")
//	@Failure      400       {object}  dto.GenericResponse[any]	"Invalid request"
//	@Failure      500       {object}  dto.GenericResponse[any]  "Internal server error"
//	@Router       /api/v2/traces/{id}/stream [get]
func GetTraceStream(c *gin.Context) {
	traceID := c.Param("id")
	lastID := c.Query("last_id")
	if lastID == "" {
		lastID = "0"
	}

	streamKey := fmt.Sprintf(consts.StreamLogKey, traceID)
	logEntry := logrus.WithFields(logrus.Fields{
		"trace_id":   traceID,
		"stream_key": streamKey,
	})

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	if c.IsAborted() {
		return
	}

	listReq := &dto.ListTasksReq{
		TraceID: traceID,
	}

	if err := listReq.Validate(); err != nil {
		logEntry.Errorf("Invalid request: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	_, tasks, err := repository.ListTasks(listReq)
	if err != nil {
		logrus.Errorf("failed to fetch tasks: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch tasks")
		return
	}

	headTask := tasks[0]
	var algorithms []dto.AlgorithmItem
	if consts.TaskType(headTask.Type) == consts.TaskTypeRestartService {
		if repository.CheckCachedField(ctx, consts.InjectionAlgorithmsKey, headTask.GroupID) {
			algorithms, err = repository.GetCachedAlgorithmItemsFromRedis(ctx, consts.InjectionAlgorithmsKey, headTask.GroupID)
			if err != nil {
				logEntry.Errorf("failed to get algorithms from Redis: %v", err)
				dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get algorithms")
				return
			}
		}
	}

	processor := repository.NewStreamProcessor(algorithms)

	logEntry.Infof("Reading historical events from Stream")
	historicalMessages, err := repository.ReadStreamEvents(ctx, streamKey, lastID, 100, 0)
	if err != nil && err != redis.Nil {
		logEntry.Errorf("failed to read historical events from redis: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read event history")
		return
	}

	var completed bool
	if len(historicalMessages) > 0 {
		lastID, completed, err = sendSSEEvents(c, processor, historicalMessages)
		if err != nil {
			logEntry.Errorf("failed to read historical stream events of ID %s: %v", lastID, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read stream events")
			return
		}

		if completed {
			logEntry.Info("Trace completed during historical events, closing stream connection")
			return
		}
	}

	logEntry.Infof("Switching to real-time event monitoring from ID: %s", lastID)
	for {
		select {
		case <-c.Done():
			logEntry.Info("Request context done")
			return

		default:
			newMessages, err := repository.ReadStreamEvents(ctx, streamKey, lastID, 10, time.Second)
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
				continue
			}

			lastID, completed, err = sendSSEEvents(c, processor, newMessages)
			if err != nil {
				logEntry.Errorf("failed to read stream events of ID %s: %v", lastID, err)
				return
			}

			if completed {
				logEntry.Info("Trace completed, closing stream connection")
				return
			}

			logrus.Info("Sent SSE messages, lastID:", lastID)
		}
	}
}

func sendSSEEvents(c *gin.Context, processor *repository.StreamProcessor, streams []redis.XStream) (string, bool, error) {
	lastID, stremEvent, err := processor.ProcessMessageForSSE(streams[0].Messages[0])
	if err != nil {
		c.SSEvent(string(consts.EventEnd), nil)
		c.Writer.Flush()
		return lastID, true, err
	}

	c.Render(-1, sse.Event{
		Id:    lastID,
		Event: string(consts.EventUpdate),
		Data:  stremEvent,
	})
	c.Writer.Flush()

	completed := processor.IsCompleted()
	if completed {
		c.SSEvent(string(consts.EventEnd), nil)
		c.Writer.Flush()
	}

	return lastID, completed, nil
}
