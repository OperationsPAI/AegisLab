package dto

type PaginationReq struct {
	PageNum  *int `form:"page_num" binding:"required,min=1"`
	PageSize *int `form:"page_size" binding:"required,min=5,max=20"`
}

type TaskReq struct {
	TaskID string `uri:"task_id" binding:"required"`
}

var PaginationFieldMap = map[string]string{
	"PageNum":  "page_num",
	"PageSize": "page_size",
}
