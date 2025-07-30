package handlers

import (
	"context"
	"net/http"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/executor"
	"github.com/LGU-SE-Internal/rcabench/middleware"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// ListAlgorithms
//
//	@Summary		Get algorithm list
//	@Description	Get all available algorithms in the system, including image info, tags, and update time. Only returns containers with active status.
//	@Tags			algorithm
//	@Produce		json
//	@Success		200	{object}	dto.GenericResponse[dto.ListAlgorithmsResp]	"Successfully returned algorithm list"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v1/algorithms [get]
func ListAlgorithms(c *gin.Context) {
	containers, err := repository.ListContainers(&dto.ListContainersFilterOptions{
		Type:   consts.ContainerTypeAlgorithm,
		Status: utils.BoolPtr(true),
	})
	if err != nil {
		logrus.Errorf("failed to list algorithms: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to list algorithms")
		return
	}

	dto.SuccessResponse(c, dto.ListAlgorithmsResp(containers))
}

// SubmitAlgorithmExecution
//
//	@Summary		Submit algorithm execution task
//	@Description	Batch submit algorithm execution tasks, supporting multiple algorithm and dataset combinations. The system assigns a unique TraceID for each execution task to track status and results.
//	@Tags			algorithm
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		dto.SubmitExecutionReq	true	"Algorithm execution request list, including algorithm name, dataset, and environment variables"
//	@Success		202		{object}	dto.GenericResponse[dto.SubmitResp]	"Successfully submitted algorithm execution task, returns task tracking info"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Request parameter error, such as invalid JSON format, algorithm name or dataset name, unsupported environment variable name, etc."
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v1/algorithms [post]
func SubmitAlgorithmExecution(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitAlgorithmExecution called, groupID: %s", groupID)

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("failed to get span context from gin.Context")
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get span context")
		return
	}

	spanCtx := ctx.(context.Context)
	span := trace.SpanFromContext(spanCtx)

	var req dto.SubmitExecutionReq
	if err := c.BindJSON(&req); err != nil {
		logrus.Error(err)
		span.SetStatus(codes.Error, "panic in SubmitAlgorithmExecution")
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Error(err)
		span.SetStatus(codes.Error, "panic in SubmitAlgorithmExecution")
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid execution payload")
		return
	}

	project, err := repository.GetProject("name", req.ProjectName)
	if err != nil {
		logrus.Errorf("failed to get project by name %s: %v", req.ProjectName, err)
		span.SetStatus(codes.Error, "panic in SubmitAlgorithmExecution")
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid project name")
		return
	}

	traces := make([]dto.Trace, 0, len(req.Payloads))
	for idx, payload := range req.Payloads {
		task := &dto.UnifiedTask{
			Type:      consts.TaskTypeRunAlgorithm,
			Payload:   utils.StructToMap(payload),
			Immediate: true,
			GroupID:   groupID,
			ProjectID: &project.ID,
		}
		task.SetGroupCtx(spanCtx)

		taskID, traceID, err := executor.SubmitTask(spanCtx, task)
		if err != nil {
			logrus.Errorf("failed to submit algorithm execution task: %v", err)
			span.SetStatus(codes.Error, "panic in SubmitAlgorithmExecution")
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit algorithm execution task")
			return
		}

		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID, Index: idx})
	}

	dto.JSONResponse(c, http.StatusAccepted,
		"Algorithm executions submitted successfully",
		dto.SubmitResp{GroupID: groupID, Traces: traces},
	)
}
