package dto

type TraceReq struct {
	TraceID string `uri:"trace_id" binding:"required"`
}
