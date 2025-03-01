package handlers

type TaskReq struct {
	TaskID string `uri:"task_id" binding:"required"`
}
