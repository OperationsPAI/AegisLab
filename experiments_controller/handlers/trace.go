package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/gin-contrib/sse"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

func GetTraceStream(c *gin.Context) {
	var traceReq dto.TraceReq
	if err := c.BindUri(&traceReq); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid URI")
		return
	}

	var req dto.TraceStreamReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Param")
		return
	}

	lastID := req.LastID
	if lastID == "" {
		lastID = "0"
	}

	streamKey := fmt.Sprintf(consts.StreamLogKey, traceReq.TraceID)
	logEntry := logrus.WithFields(logrus.Fields{
		"trace_id":   traceReq.TraceID,
		"stream_key": streamKey,
	})

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	if c.IsAborted() {
		return
	}

	logEntry.Infof("Reading historical events from Stream")
	historicalMessages, err := client.ReadStreamEvents(ctx, streamKey, lastID, 100, 0)
	if err != nil && err != redis.Nil {
		logEntry.Errorf("failed to read historical events: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read event history")
		return
	}

	if len(historicalMessages) > 0 {
		lastID, err = sendSSEMessages(c, historicalMessages)
		if err != nil {
			logEntry.Error(err)
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
			newMessages, err := client.ReadStreamEvents(ctx, streamKey, lastID, 10, time.Second)
			if err != nil && err != redis.Nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					logEntry.Infof("Context done while reading stream: %v", err)
					return
				}
				logEntry.Errorf("Error reading stream: %v", err)
				continue
			}
			if err == redis.Nil {
				continue
			}

			lastID, err = sendSSEMessages(c, newMessages)
			logrus.Info("Sent SSE messages, lastID:", lastID)
			if err != nil {
				logEntry.Error(err)
				return
			}
		}
	}
}

func sendSSEMessages(c *gin.Context, messages []redis.XStream) (string, error) {
	var lastID string
	for _, stream := range messages {
		for _, msg := range stream.Messages {
			lastID = msg.ID

			streamEvent, err := client.ParseEventFromValues(msg.Values)
			if err != nil {
				return "", fmt.Errorf("failed to parse stream message value: %v", err)
			}

			sseMessage, err := streamEvent.ToSSE()
			if err != nil {
				return "", fmt.Errorf("failed to parse streamEvent to sse message: %v", err)
			}

			c.Render(-1, sse.Event{
				Id:    msg.ID,
				Event: consts.EventUpdate,
				Data:  sseMessage,
			})

			c.Writer.Flush()

			if streamEvent.TaskType == consts.TaskTypeCollectResult {
				if payload, ok := streamEvent.Payload.(client.InfoPayloadTemplate); ok {
					if payload.Status == consts.TaskStatusCompleted {
						c.SSEvent(consts.EventEnd, nil)
						c.Writer.Flush()
						return lastID, nil
					}
				}
			}
		}
	}

	return lastID, nil
}
