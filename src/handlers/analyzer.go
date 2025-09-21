package handlers

import (
	"fmt"
	"net/http"

	"aegis/dto"
	"aegis/executor/analyzer"
	"aegis/repository"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AnalyzeInjections handles fault injection analysis requests
//
//	@Summary     	Analyze fault injection data
//	@Description 	Analyze fault injection data using various filtering conditions, returning statistical information including efficiency, diversity, distance between seeds, etc.
//	@Tags        	analyzer
//	@Produce     	json
//	@Param			project_name		query		string	false	"Project name filter"
//	@Param			env					query		string	false	"Environment label filter"	Enums(dev, prod)	default(prod)
//	@Param			batch				query		string	false	"Batch label filter"
//	@Param			tag					query		string	false	"Classification label filter"	Enums(train, test)	default(train)
//	@Param			benchmark			query		string	false	"Benchmark type filter"	Enums(clickhouse)	default(clickhouse)
//	@Param			status				query		int		false	"Status filter, refer to field mapping interface (/mapping) for specific values"	default(0)
//	@Param			fault_type			query		int		false	"Fault type filter, refer to field mapping interface (/mapping) for specific values"	default(0)
//	@Param			lookback			query		string	false	"Time range query, supports custom relative time (1h/24h/7d) or custom, default is not set"
//	@Param			custom_start_time	query		string	false	"Custom start time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Param			custom_end_time		query		string	false	"Custom end time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Success     	200  				{object}    dto.GenericResponse[dto.AnalyzeInjectionsResp]	"Returns fault injection analysis statistics"
//	@Failure		400					{object}	dto.GenericResponse[any]	"Request parameter error, such as incorrect parameter format, validation failure, etc."
//	@Failure     	500  				{object}    dto.GenericResponse[any]	"Internal server error"
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

// AnalyzeTraces handles trace analysis requests
//
//	@Summary     	Analyze trace data
//	@Description 	Analyze trace data using various filtering conditions, returning statistical information including traces ending with fault injection
//	@Tags        	trace
//	@Produce     	json
//	@Param       	first_task_type     query   	string  false  	"First task type filter"
//	@Param			lookback			query		string	false	"Time range query, supports custom relative time (1h/24h/7d) or custom, default is not set"
//	@Param			custom_start_time	query		string	false	"Custom start time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Param			custom_end_time		query		string	false	"Custom end time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Success     	200  				{object}    dto.GenericResponse[dto.TraceStats]	"Returns trace analysis statistics"
//	@Failure		400					{object}	dto.GenericResponse[any]	"Request parameter error, such as incorrect parameter format, validation failure, etc."
//	@Failure     	500  				{object}    dto.GenericResponse[any]	"Internal server error"
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
