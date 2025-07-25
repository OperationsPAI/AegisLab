package handlers

import (
	"fmt"
	"net/http"

	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor/analyzer"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AnalyzeTrace 处理链路分析请求
// @Summary     分析链路数据
// @Description 使用多种筛选条件分析链路数据，返回包括故障注入结束链路在内的统计信息
// @Tags        trace
// @Produce     json
// @Param       first_task_type     query   string  false  "子任务类型筛选"
// @Param		lookback			query	string	false	"相对时间查询，如 1h, 24h, 7d或者是custom"
// @Param       custom_start_time 	query   string  false  "当lookback=custom时必需，自定义开始时间(RFC3339格式)"
// @Param       custom_end_time  	query   string  false  "当lookback=custom时必需，自定义结束时间(RFC3339格式)"
// @Success     200  {object}    	dto.GenericResponse[any]  "返回统计信息，包含fault_injection_traces字段显示以FaultInjection事件结束的trace_id列表"
// @Failure     400  {object}    	dto.GenericResponse[any]
// @Failure     500  {object}    	dto.GenericResponse[any]
// @Router      /api/v1/traces/analyze [get]
func AnalyzeTraces(c *gin.Context) {
	var req dto.AnalyzeTracesReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("failed to bind query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("invalid query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
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
