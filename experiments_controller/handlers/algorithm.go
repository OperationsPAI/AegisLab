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
//	@Summary		获取算法列表
//	@Description	获取系统中所有可用的算法列表，包括算法的镜像信息、标签和更新时间。只返回状态为激活的算法容器
//	@Tags			algorithm
//	@Produce		json
//	@Success		200	{object}	dto.GenericResponse[dto.ListAlgorithmsResp]	"成功返回算法列表"
//	@Failure		500	{object}	dto.GenericResponse[any]	"服务器内部错误"
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
//	@Summary		提交算法执行任务
//	@Description	批量提交算法执行任务，支持多个算法和数据集的组合执行。系统将为每个执行任务分配唯一的 TraceID 用于跟踪任务状态和结果
//	@Tags			algorithm
//	@Produce		json
//	@Consumes		application/json
//	@Param			body	body		dto.SubmitExecutionReq	true	"算法执行请求列表，包含算法名称、数据集和环境变量"
//	@Success		202		{object}	dto.GenericResponse[dto.SubmitResp]	"成功提交算法执行任务，返回任务跟踪信息"
//	@Failure		400		{object}	dto.GenericResponse[any]	"请求参数错误，如JSON格式不正确、算法名称或数据集名称无效、环境变量名称不支持等"
//	@Failure		500		{object}	dto.GenericResponse[any]	"服务器内部错误"
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
