package handlers

import (
	"github.com/gin-gonic/gin"
)

type EvalResp struct {
	TaskID string `json:"task_id"`
}

// CancelEvaluation
//
//	@Summary		取消评估的执行
//	@Description	取消评估的执行
//	@Tags			evaluation
//	@Router			/api/v1/evaluation/cancel [post]
func CancelEvaluation(c *gin.Context) {

}

// GetEvaluationList
//
//	@Summary		获取评估列表
//	@Description	获取评估列表
//	@Tags			evaluation
//	@Router			/api/v1/evaluation [get]
func GetEvaluationList(c *gin.Context) {

}

// GetEvaluationLogs
//
//	@Summary		查看评估的日志
//	@Description	查看评估的日志
//	@Tags			evaluation
//	@Router			/api/v1/evaluation/getlogs [post]
func GetEvaluationLogs(c *gin.Context) {

}

// GetEvaluationStatus
//
//	@Summary		查看评估状态
//	@Description	查看评估状态
//	@Tags			evaluation
//	@Router			/api/v1/evaluation/getstatus [post]
func GetEvaluationStatus(c *gin.Context) {

}

// SubmitEvaluation
// TODO 批量评估
//
//	@Summary		提交评估
//	@Description	提交评估
//	@Tags			evaluation
//	@Produce		application/json
//	@Consumes		application/json
//	@Param			type	query		string								true	"任务类型"
//	@Param			body	body		executor.AlgorithmExecutionPayload	true	"请求体"
//	@Success		200		{object}	GenericResponse[EvalResp]
//	@Failure		400		{object}	GenericResponse[any]
//	@Failure		500		{object}	GenericResponse[any]
//	@Router			/api/v1/evaluation [post]
func SubmitEvaluation(c *gin.Context) {
}
