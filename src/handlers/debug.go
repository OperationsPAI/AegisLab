package handlers

import (
	"net/http"

	"github.com/LGU-SE-Internal/rcabench/client/debug"
	"github.com/LGU-SE-Internal/rcabench/client/k8s"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
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

// GetNSLock
//
//	@Summary		获取命名空间锁状态
//	@Description	获取命名空间锁状态信息
//	@Tags			debug
//	@Produce		json
//	@Success		200	{object}	dto.GenericResponse[any]
//	@Failure		500	{object}	dto.GenericResponse[any]
//	@Router			/api/v1/debug/ns/status [get]
func GetNSLock(c *gin.Context) {
	cli := k8s.GetMonitor()
	items, err := cli.InspectLock()
	if err != nil {
		logrus.Error("failed to inspect namespace locks:", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to inspect lock")
		return
	}

	dto.SuccessResponse(c, items)
}
