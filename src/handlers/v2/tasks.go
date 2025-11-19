package v2

import (
	"aegis/consts"
	"net/http"

	"aegis/dto"
	"aegis/handlers"
	producer "aegis/service/prodcuer"
	"aegis/utils"

	"github.com/gin-gonic/gin"
)

// BatchDeleteTasks
//
//	@Summary		Batch delete tasks
//	@Description	Batch delete tasks by IDs
//	@Tags			Tasks
//	@ID				batch_delete_tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			batch_delete	body		dto.BatchDeleteTaskReq		true	"Batch delete request"
//	@Success		200				{object}	dto.GenericResponse[any]	"Tasks deleted successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]	"Invalid request format or parameters"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/tasks/batch-delete [post]
func BatchDeleteTasks(c *gin.Context) {
	var req dto.BatchDeleteTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters"+err.Error())
		return
	}

	err := producer.BatchDeleteTasks(req.IDs)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Tasks deleted successfully", nil)
}

// GetTask handles getting a single task by ID
//
//	@Summary		Get task by ID
//	@Description	Get detailed information about a specific task
//	@Tags			Tasks
//	@ID				get_task_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			task_id	path		string									true	"Task ID"
//	@Success		200		{object}	dto.GenericResponse[dto.TaskDetailResp]	"Task retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid task ID"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]				"Task not found"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/tasks/{task_id} [get]
//	@x-api-type		{"sdk":"true"}
func GetTask(c *gin.Context) {
	taskID := c.Param(consts.URLPathTaskID)
	if !utils.IsValidUUID(taskID) {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	resp, err := producer.GetTaskDetail(taskID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListTasks handles simple task listing
//
//	@Summary		List tasks
//	@Description	Get a simple list of tasks with basic filtering via query parameters
//	@Tags			Tasks
//	@ID				list_tasks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int												false	"Page number"	default(1)
//	@Param			size		query		int												false	"Page size"		default(20)
//	@Param			task_type	query		consts.TaskType									false	"Filter by task type"
//	@Param			immediate	query		bool											false	"Filter by immediate execution"
//	@Param			trace_id	query		string											false	"Filter by trace ID (uuid format)"
//	@Param			group_id	query		string											false	"Filter by group ID (uuid format)"
//	@Param			project_id	query		int												false	"Filter by project ID"
//	@Param			state		query		consts.TaskState								false	"Filter by state"
//	@Param			status		query		consts.StatusType								false	"Filter by status"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.TaskResp]]	"Tasks retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/tasks [get]
//	@x-api-type		{"sdk":"true"}
func ListTasks(c *gin.Context) {
	var req dto.ListTaskReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format : "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListTasks(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}
