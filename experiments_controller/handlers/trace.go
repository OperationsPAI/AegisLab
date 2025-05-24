package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor/analyzer"
	"github.com/LGU-SE-Internal/rcabench/repository"
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
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
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
	historicalMessages, err := repository.ReadStreamEvents(ctx, streamKey, lastID, 100, 0)
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
			newMessages, err := repository.ReadStreamEvents(ctx, streamKey, lastID, 10, time.Second)
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
	lastID, sseMessages, err := repository.ProcessStreamMessagesForSSE(messages)
	if err != nil {
		return "", err
	}

	for _, message := range sseMessages {
		c.Render(-1, sse.Event{
			Id:    message.ID,
			Event: consts.EventUpdate,
			Data:  message.Data,
		})

		c.Writer.Flush()

		if message.IsCompleted {
			c.SSEvent(consts.EventEnd, nil)
			c.Writer.Flush()
			return lastID, nil
		}
	}

	return lastID, nil
}

func GetTaskEventMap(c *gin.Context) {
	dto.SuccessResponse(c, dto.ValidTaskEventMap)
}

func GetValidTaskTypes(c *gin.Context) {
	dto.SuccessResponse(c, dto.ValidTaskTypes)
}

// AnalyzeTrace 处理链路分析请求
// @Summary     分析链路数据
// @Description 使用多种筛选条件分析链路数据
// @Tags        trace
// @Produce     json
// @Param       first_task_type      query   string  false  "子任务类型筛选"
// @Param       lookback         query   string  false  "时间回溯范围(5m,15m,30m,1h,2h,3h,6h,12h,1d,2d,custom)"
// @Param       custom_start_time query   string  false  "当lookback=custom时必需，自定义开始时间(RFC3339格式)"
// @Param       custom_end_time  query   string  false  "当lookback=custom时必需，自定义结束时间(RFC3339格式)"
// @Success     200  {object}    dto.GenericResponse[any]
// @Failure     400  {object}    dto.GenericResponse[any]
// @Failure     500  {object}    dto.GenericResponse[any]
// @Router      /api/v1/traces/analyze [get]
func AnalyzeTrace(c *gin.Context) {
	var req dto.TraceAnalyzeReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	filterOptions, err := req.Convert()
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid filter options: %v", err))
		return
	}

	stats, err := analyzer.AnalyzeTrace(c.Request.Context(), *filterOptions)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to analyze trace")
		return
	}

	dto.SuccessResponse(c, stats)
}

// GetCompletedMap 获取完成状态的链路
// @Summary     获取完成状态的链路
// @Description 根据指定的时间范围获取完成状态的链路
// @Tags        trace
// @Produce     json
// @Param       lookback         query   string   false "时间回溯范围(5m,15m,30m,1h,2h,3h,6h,12h,1d,2d,custom)"
// @Param       custom_start_time query   string   false "当lookback=custom时必需，自定义开始时间(RFC3339格式)"
// @Param       custom_end_time  query   string   false "当lookback=custom时必需，自定义结束时间(RFC3339格式)"
// @Success     200 {object}     dto.GenericResponse[any]
// @Failure     400 {object}     dto.GenericResponse[any]
// @Failure     500 {object}     dto.GenericResponse[any]
// @Router      /api/v1/traces/completed [get]
func GetCompletedMap(c *gin.Context) {
	var req dto.GetCompletedMapReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid request: %v", err))
		return
	}

	filterOptions, err := req.Convert()
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, fmt.Sprintf("Invalid filter options: %v", err))
		return
	}

	result, err := analyzer.GetCompletedMap(c.Request.Context(), *filterOptions)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to analyze trace")
		return
	}

	dto.SuccessResponse(c, result)
}
