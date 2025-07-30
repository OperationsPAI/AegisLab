package handlers

import (
	"fmt"
	"net/http"

	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor/analyzer"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AnalyzeInjections 处理故障注入分析请求
//
//	@Summary     	分析故障注入数据
//	@Description 	使用多种筛选条件分析故障注入数据，返回包括效率、多样性、种子之间的距离等统计信息
//	@Tags        	analyzer
//	@Produce     	json
//	@Param			project_name		query		string	false	"项目名称过滤"
//	@Param			env					query		string	false	"环境标签过滤"	Enums(dev, prod)	default(prod)
//	@Param			batch				query		string	false	"批次标签过滤"
//	@Param			tag					query		string	false	"分类标签过滤"	Enums(train, test)	default(train)
//	@Param			benchmark			query		string	false	"基准测试类型过滤"	Enums(clickhouse)	default(clickhouse)
//	@Param			status				query		int		false	"状态过滤，具体值参考字段映射接口(/mapping)"	default(0)
//	@Param			fault_type			query		int		false	"故障类型过滤，具体值参考字段映射接口(/mapping)"	default(0)
//	@Param			lookback			query		string	false	"时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置"
//	@Param			custom_start_time	query		string	false	"自定义开始时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Param			custom_end_time		query		string	false	"自定义结束时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Success     	200  				{object}    dto.GenericResponse[dto.AnalyzeInjectionsResp]	"返回故障注入分析统计信息"
//	@Failure		400					{object}	dto.GenericResponse[any]	"请求参数错误，如参数格式不正确、验证失败等"
//	@Failure     	500  				{object}    dto.GenericResponse[any]	"服务器内部错误"
//	@Router      	/api/v1/analyzers/injections [get]
func AnalyzeInjections(c *gin.Context) {
	var req dto.AnalyzeInjectionsReq
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

	_, injections, err := repository.ListInjections(req.ToListInjectionsReq())
	if err != nil {
		logrus.Errorf("failed to list injections: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to list injections")
		return
	}

	length := len(injections)
	if length == 0 {
		logrus.Errorf("no injections found")
		dto.ErrorResponse(c, http.StatusNotFound, "No injections found")
		return
	}

	startTime := injections[length-1].CreatedAt
	endTime := injections[0].CreatedAt
	timeDiffHours := endTime.Sub(startTime).Hours()

	efficiency := 0.0
	if timeDiffHours > 0 {
		efficiency = float64(length) / float64(timeDiffHours)
	}

	ia, err := analyzer.NewInjectionAnalyzer()
	if err != nil {
		logrus.Errorf("failed to create injection analyzer: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to create injection analyzer")
		return
	}

	stats, err := ia.Analyze(injections)
	if err != nil {
		logrus.Errorf("failed to analyze injections: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to analyze injections")
		return
	}

	dto.SuccessResponse(c, dto.AnalyzeInjectionsResp{
		Efficiency: fmt.Sprintf("%f/h", efficiency),
		Stats:      stats,
	})
}

// AnalyzeTraces 处理链路分析请求
//
//	@Summary     	分析链路数据
//	@Description 	使用多种筛选条件分析链路数据，返回包括故障注入结束链路在内的统计信息
//	@Tags        	trace
//	@Produce     	json
//	@Param       	first_task_type     query   	string  false  	"首任务类型筛选"
//	@Param			lookback			query		string	false	"时间范围查询，支持自定义相对时间(1h/24h/7d)或custom 默认不设置"
//	@Param			custom_start_time	query		string	false	"自定义开始时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Param			custom_end_time		query		string	false	"自定义结束时间，RFC3339格式，当lookback=custom时必需"	Format(date-time)
//	@Success     	200  				{object}    dto.GenericResponse[dto.TraceStats]	"返回链路分析统计信息"
//	@Failure		400					{object}	dto.GenericResponse[any]	"请求参数错误，如参数格式不正确、验证失败等"
//	@Failure     	500  				{object}    dto.GenericResponse[any]	"服务器内部错误"
//	@Router      	/api/v1/analyzers/traces [get]
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

	stats, err := analyzer.AnalyzeTraces(c.Request.Context(), &req)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to analyze trace")
		return
	}

	dto.SuccessResponse(c, stats)
}
