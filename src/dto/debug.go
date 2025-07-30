package dto

type DebugGetReq struct {
	Name string `form:"name" binding:"required"`
}

type DebugSetReq struct {
	Name  string `json:"name" binding:"required"`
	Value any    `json:"value"`
}
