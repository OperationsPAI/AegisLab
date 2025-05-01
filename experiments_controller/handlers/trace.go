package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
	"github.com/CUHK-SE-Group/rcabench/consts"
	"github.com/CUHK-SE-Group/rcabench/dto"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// GetTraceStream
//
//	@Summary      获取 trace 任务状态事件流
//	@Description  通过Server-Sent Events (SSE) 实时获取任务的执行状态更新，直到任务完成或连接关闭
//	@Tags         injection
//	@Produce      text/event-stream
//	@Consumes	  application/json
//	@Param        trace_id path      string  				true  "需要监控的链路ID"
//	@Success      200      {object}  nil     				"成功建立SSE连接，持续推送事件流"
//	@Failure      400      {object}  dto.GenericResponse[any]	"无效的链路ID格式"
//	@Failure      404      {object}  dto.GenericResponse[any]  	"指定ID的链路不存在"
//	@Failure      500      {object}  dto.GenericResponse[any]  	"服务器内部错误"
func GetTraceStream(c *gin.Context) {
	var req dto.TraceReq
	if err := c.BindUri(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid param")
		return
	}

	streamKey := fmt.Sprintf(consts.StreamLogKey, req.TraceID)
	logEntry := logrus.WithFields(logrus.Fields{
		"trace_id":   req.TraceID,
		"stream_key": streamKey,
	})

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	if c.IsAborted() {
		return
	}

	lastID := "0"
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
			newMessages, err := client.ReadStreamEvents(ctx, streamKey, lastID, 5, time.Second)
			if err != nil && err != redis.Nil {
				logEntry.Errorf("Error reading stream: %v", err)
				continue
			}

			lastID, err = sendSSEMessages(c, newMessages)
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

			c.SSEvent(consts.EventUpdate, sseMessage)
			c.Writer.Flush()

			if isTerminatingMessage(streamEvent, consts.TaskTypeCollectResult) {
				c.SSEvent(consts.EventEnd, nil)
				c.Writer.Flush()
				return lastID, nil
			}
		}
	}

	return lastID, nil
}

func isTerminatingMessage(streamEvent *client.StreamEvent, expectedTaskType consts.TaskType) bool {
	return streamEvent.TaskType == expectedTaskType
}
