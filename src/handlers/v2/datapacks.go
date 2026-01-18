package v2

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/handlers"
	producer "aegis/service/producer"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ListDatapacks handles listing datapacks with filtering
//
//	@Summary		List datapacks
//	@Description	Get a paginated list of datapacks with optional filtering
//	@Tags			Datapacks
//	@ID				list_datapacks
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int									false	"Page number"	default(1)
//	@Param			size		query		int									false	"Page size"		default(20)
//	@Param			state		query		string								false	"Datapack state"
//	@Param			benchmark	query		string								false	"Benchmark name"
//	@Param			fault_type	query		string								false	"Fault type"
//	@Success		200			{object}	dto.GenericResponse[dto.ListInjectionResp]	"Datapacks retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]			"Invalid request"
//	@Failure		401			{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		500			{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/datapacks [get]
//	@x-api-type		{"sdk":"true"}
func ListDatapacks(c *gin.Context) {
	var req dto.ListInjectionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	// Default pagination
	if req.Page == 0 {
		req.Page = 1
	}
	if req.Size == 0 {
		req.Size = 20
	}

	resp, err := producer.ListInjections(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Datapacks retrieved successfully", resp)
}

// GetDatapack handles getting a single datapack by ID
//
//	@Summary		Get datapack by ID
//	@Description	Get detailed information about a specific datapack
//	@Tags			Datapacks
//	@ID				get_datapack_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int											true	"Datapack ID"
//	@Success		200	{object}	dto.GenericResponse[dto.InjectionDetailResp]	"Datapack retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]					"Invalid datapack ID"
//	@Failure		401	{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]					"Datapack not found"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/datapacks/{id} [get]
//	@x-api-type		{"sdk":"true"}
func GetDatapack(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "datapack ID")
	if !ok {
		return
	}

	resp, err := producer.GetInjectionDetail(id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Datapack retrieved successfully", resp)
}

// GetDatapackMetadata handles getting datapack metadata
//
//	@Summary		Get datapack metadata
//	@Description	Get metadata about a datapack including trace count, size, etc.
//	@Tags			Datapacks
//	@ID				get_datapack_metadata
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int									true	"Datapack ID"
//	@Success		200	{object}	dto.GenericResponse[dto.DatapackInfo]	"Metadata retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]			"Invalid datapack ID"
//	@Failure		401	{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]			"Datapack not found"
//	@Failure		500	{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/datapacks/{id}/metadata [get]
//	@x-api-type		{"sdk":"true"}
func GetDatapackMetadata(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "datapack ID")
	if !ok {
		return
	}

	resp, err := producer.GetInjectionDetail(id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Metadata retrieved successfully", resp)
}

// DownloadDatapack handles downloading a datapack as a zip archive
//
//	@Summary		Download datapack
//	@Description	Download a datapack as a zip archive containing all trace data
//	@Tags			Datapacks
//	@ID				download_datapack
//	@Produce		application/zip
//	@Security		BearerAuth
//	@Param			id	path		int							true	"Datapack ID"
//	@Success		200	{file}		binary						"Datapack zip file"
//	@Failure		400	{object}	dto.GenericResponse[any]	"Invalid datapack ID"
//	@Failure		401	{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]	"Datapack not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/datapacks/{id}/download [get]
//	@x-api-type		{"sdk":"true"}
func DownloadDatapack(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "datapack ID")
	if !ok {
		return
	}

	zipData, filename, err := producer.DownloadDatapack(id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Length", strconv.Itoa(len(zipData)))
	c.Data(http.StatusOK, "application/zip", zipData)
}
