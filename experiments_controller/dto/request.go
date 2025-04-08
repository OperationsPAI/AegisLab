package dto

type PaginationReq struct {
	PageNum  int `form:"page_num" binding:"required,min=1"`
	PageSize int `form:"page_size" binding:"required,oneof=10 20 50"`
}

var PaginationFieldMap = map[string]string{
	"PageNum":  "page_num",
	"PageSize": "page_size",
}
