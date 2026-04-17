package taskmodule

import (
	"aegis/httpx"
	"errors"
	"net/http"

	"aegis/consts"
	"aegis/dto"
	"aegis/utils"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 4096,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// BatchDeleteTasks
//
//	@Summary		Batch delete tasks
//	@Description	Batch delete tasks by IDs
//	@Tags			Tasks
//	@ID				batch_delete_tasks
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			batch_delete	body		BatchDeleteTaskReq			true	"Batch delete request"
//	@Success		200				{object}	dto.GenericResponse[any]	"Tasks deleted successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]	"Invalid request format or parameters"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/tasks/batch-delete [post]
//	@x-api-type		{}
func (h *Handler) BatchDelete(c *gin.Context) {
	var req BatchDeleteTaskReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters"+err.Error())
		return
	}

	err := h.service.BatchDelete(c.Request.Context(), req.IDs)
	if httpx.HandleServiceError(c, err) {
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
//	@Success		200		{object}	dto.GenericResponse[TaskDetailResp]		"Task retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid task ID"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]				"Task not found"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/tasks/{task_id} [get]
//	@x-api-type		{}
func (h *Handler) Get(c *gin.Context) {
	taskID := c.Param(consts.URLPathTaskID)
	if !utils.IsValidUUID(taskID) {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid task ID")
		return
	}

	resp, err := h.service.GetDetail(c.Request.Context(), taskID)
	if httpx.HandleServiceError(c, err) {
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
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[TaskResp]]	"Tasks retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/tasks [get]
//	@x-api-type		{}
func (h *Handler) List(c *gin.Context) {
	var req ListTaskReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format : "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.List(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// GetTaskLogsWS handles WebSocket connections for real-time job log streaming
//
//	@Summary		Stream task logs via WebSocket
//	@Description	Establishes a WebSocket connection to stream real-time logs.
//	@Description	Process: 1. Validate Token -> 2. Push historical logs from Loki -> 3. Subscribe to Redis for real-time updates -> 4. Close on task completion.
//	@Tags			Tasks
//	@ID				get_task_logs_ws
//	@Param			task_id	path		string				true	"Task ID"
//	@Param			token	query		string				true	"JWT authentication token"
//	@Success		101		{object}	WSLogMessage	"WebSocket connection established"
//	@Failure		400		"Invalid task ID"
//	@Failure		401		"Authentication failed"
//	@Failure		404		"Task not found"
//	@Router			/api/v2/tasks/{task_id}/logs/ws [get]
//	@x-api-type		{}
func (h *Handler) LogsWS(c *gin.Context) {
	taskID := c.Param(consts.URLPathTaskID)
	if taskID == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "task_id is required")
		return
	}

	token := c.Query("token")
	if token == "" {
		dto.ErrorResponse(c, http.StatusUnauthorized, "token query parameter is required")
		return
	}

	if _, err := utils.ValidateToken(token); err != nil {
		dto.ErrorResponse(c, http.StatusUnauthorized, "invalid token: "+err.Error())
		return
	}

	task, err := h.service.GetForLogStream(c.Request.Context(), taskID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			dto.ErrorResponse(c, http.StatusNotFound, "task not found")
			return
		}

		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to retrieve task: "+err.Error())
		return
	}

	conn, err := wsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logrus.Errorf("WebSocket upgrade failed for task %s: %v", taskID, err)
		return
	}
	defer func() { _ = conn.Close() }()

	h.service.StreamLogs(c.Request.Context(), conn, task)
}
