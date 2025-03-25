package dto

type StreamReq struct {
	TaskID  string `form:"task_id"`
	TraceID string `form:"trace_id"`
}
