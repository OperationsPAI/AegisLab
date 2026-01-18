package v2

import (
	"aegis/consts"
	"aegis/dto"
	producer "aegis/service/producer"
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

// GetNotificationStream handles streaming of global workflow notifications via Server-Sent Events (SSE)
//
//	@Summary		Stream global notifications in real-time
//	@Description	Establishes a Server-Sent Events (SSE) connection to stream workflow notifications (injection completed, datapack ready, execution completed, etc.) in real-time.
//	@Tags			Notifications
//	@ID				get_notification_stream
//	@Produce		text/event-stream
//	@Security		BearerAuth
//	@Param			last_id	query		string	false	"Last event ID received"	default("0")
//	@Success		200		{string}	string	"A stream of notification events"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/notifications/stream [get]
//	@x-api-type		{"sdk":"true"}
//	@x-request-type	{"stream":"true"}
func GetNotificationStream(c *gin.Context) {
	var req dto.GetNotificationStreamReq
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

	streamKey := consts.NotificationStreamKey
	logEntry := logrus.WithFields(logrus.Fields{
		"stream_key": streamKey,
	})

	logEntry.Infof("Reading historical notifications from Stream")
	historicalMessages, err := producer.ReadNotificationStreamMessages(ctx, streamKey, req.LastID, 100, 0)
	if err != nil {
		logEntry.Errorf("failed to read historical notifications from redis: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read notification history")
		return
	}

	if len(historicalMessages) > 0 {
		lastID, err := sendNotificationSSEEvents(c, historicalMessages)
		if err != nil {
			logEntry.Errorf("failed to send historical notification events of ID %s: %v", req.LastID, err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "failed to send notification events")
			return
		}
		req.LastID = lastID
	}

	logEntry.Infof("Switching to real-time notification monitoring from ID: %s", req.LastID)
	for {
		select {
		case <-c.Done():
			logEntry.Info("Request context done")
			return

		default:
			newMessages, err := producer.ReadNotificationStreamMessages(ctx, streamKey, req.LastID, 10, time.Second)
			if err != nil {
				if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
					logEntry.Infof("Context done while reading stream: %v", err)
					return
				}

				logEntry.Errorf("Error reading notification stream: %v", err)
				dto.ErrorResponse(c, http.StatusInternalServerError, "failed to read notification events")
				return
			}

			if len(newMessages) == 0 {
				logEntry.Debug("No new notifications, continuing")
				continue
			}

			lastID, err := sendNotificationSSEEvents(c, newMessages)
			if err != nil {
				logEntry.Errorf("failed to send notification events of ID %s: %v", lastID, err)
				return
			}

			req.LastID = lastID
			logrus.Info("Sent notification SSE messages, lastID:", lastID)
		}
	}
}

// sendNotificationSSEEvents processes and sends notification messages as SSE events
func sendNotificationSSEEvents(c *gin.Context, streams []redis.XStream) (string, error) {
	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return "", fmt.Errorf("no messages to process")
	}

	var lastID string
	for _, msg := range streams[0].Messages {
		lastID = msg.ID

		// Parse notification event from message
		notification := parseNotificationMessage(msg)

		c.Render(-1, sse.Event{
			Id:    lastID,
			Event: "notification",
			Data:  notification,
		})
		c.Writer.Flush()
	}

	return lastID, nil
}

// parseNotificationMessage converts a Redis stream message to a notification event
func parseNotificationMessage(msg redis.XMessage) dto.NotificationEvent {
	notification := dto.NotificationEvent{
		Timestamp: time.Now(),
	}

	for key, val := range msg.Values {
		switch key {
		case "type":
			notification.Type = val.(string)
		case "entity_id":
			notification.EntityID = val.(string)
		case "message":
			notification.Message = val.(string)
		case "status":
			notification.Status = val.(string)
		}
	}

	return notification
}
