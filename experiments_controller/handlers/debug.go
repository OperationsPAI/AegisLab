package handlers

import (
	"net/http"

	"github.com/LGU-SE-Internal/rcabench/client/debug"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/gin-gonic/gin"
)

func GetAllVars(c *gin.Context) {
	dto.SuccessResponse[any](c, debug.NewDebugRegistry().GetAll())
}

func GetVar(c *gin.Context) {
	var req dto.DebugGetReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "invalid JSON format")
		return
	}

	data, err := debug.NewDebugRegistry().Get(req.Name)
	if err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	dto.SuccessResponse[any](c, data)
}

func SetVar(c *gin.Context) {
	var req dto.DebugSetReq
	if err := c.BindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "invalid JSON format")
		return
	}

	if err := debug.NewDebugRegistry().Set(req.Name, req.Value); err != nil {
		dto.ErrorResponse(c, http.StatusNotFound, err.Error())
		return
	}

	dto.SuccessResponse[any](c, nil)
}
