package dataset

import (
	"aegis/httpx"
	"archive/zip"
	"fmt"
	"net/http"
	"strconv"

	"aegis/consts"
	"aegis/dto"
	"aegis/middleware"
	"aegis/utils"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service HandlerService
}

func NewHandler(service HandlerService) *Handler {
	return &Handler{service: service}
}

// CreateDataset handles dataset creation
//
//	@Summary		Create dataset
//	@Description	Create a new dataset with an initial version
//	@Tags			Datasets
//	@ID				create_dataset
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateDatasetReq					true	"Dataset creation request"
//	@Success		201		{object}	dto.GenericResponse[DatasetResp]	"Dataset created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		409		{object}	dto.GenericResponse[any]			"Conflict error"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/datasets [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) CreateDataset(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateDatasetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.CreateDataset(c.Request.Context(), &req, userID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusCreated, "Dataset created successfully", resp)
}

// DeleteDataset handles dataset deletion
//
//	@Summary		Delete dataset
//	@Description	Delete a dataset (soft delete by setting status to -1)
//	@Tags			Datasets
//	@ID				delete_dataset
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int							true	"Dataset ID"
//	@Success		204			{object}	dto.GenericResponse[any]	"Dataset deleted successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid dataset ID"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]	"Dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id} [delete]
//	@x-api-type		{"portal":"true"}
func (h *Handler) DeleteDataset(c *gin.Context) {
	datasetID, ok := parseDatasetID(c)
	if !ok {
		return
	}

	if httpx.HandleServiceError(c, h.service.DeleteDataset(c.Request.Context(), datasetID)) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Dataset deleted successfully", nil)
}

// GetDataset handles getting a single dataset by ID
//
//	@Summary		Get dataset by ID
//	@Description	Get detailed information about a specific dataset
//	@Tags			Datasets
//	@ID				get_dataset_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int										true	"Dataset ID"
//	@Success		200			{object}	dto.GenericResponse[DatasetDetailResp]	"Dataset retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid dataset ID"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]				"Dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id} [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) GetDataset(c *gin.Context) {
	datasetID, ok := parseDatasetID(c)
	if !ok {
		return
	}

	resp, err := h.service.GetDataset(c.Request.Context(), datasetID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListDatasets handles listing datasets with pagination and filtering
//
//	@Summary		List datasets
//	@Description	Get paginated list of datasets with pagination and filtering
//	@Tags			Datasets
//	@ID				list_datasets
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int												false	"Page number"	default(1)
//	@Param			size		query		int												false	"Page size"		default(20)
//	@Param			type		query		string											false	"Dataset type filter"
//	@Param			is_public	query		bool											false	"Dataset public visibility filter"
//	@Param			status		query		consts.StatusType								false	"Dataset status filter"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[DatasetResp]]	"Datasets retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/datasets [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListDatasets(c *gin.Context) {
	var req ListDatasetReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.ListDatasets(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// SearchDataset handles searching datasets with advanced filtering
//
//	@Summary		Search datasets
//	@Description	Search datasets with advanced filtering options
//	@Tags			Datasets
//	@ID				search_datasets
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		SearchDatasetReq										true	"Dataset search request"
//	@Success		200		{object}	dto.GenericResponse[dto.ListResp[DatasetDetailResp]]	"Datasets retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]								"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/datasets/search [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) SearchDataset(c *gin.Context) {
	var req SearchDatasetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.SearchDatasets(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// UpdateDataset handles dataset updates
//
//	@Summary		Update dataset
//	@Description	Update an existing dataset's information
//	@Tags			Datasets
//	@ID				update_dataset
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int									true	"Dataset ID"
//	@Param			request		body		UpdateDatasetReq					true	"Dataset update request"
//	@Success		202			{object}	dto.GenericResponse[DatasetResp]	"Dataset updated successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]			"Invalid dataset ID/request"
//	@Failure		401			{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]			"Dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id} [patch]
//	@x-api-type		{"portal":"true"}
func (h *Handler) UpdateDataset(c *gin.Context) {
	datasetID, ok := parseDatasetID(c)
	if !ok {
		return
	}

	var req UpdateDatasetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.UpdateDataset(c.Request.Context(), &req, datasetID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Dataset updated successfully", resp)
}

// ManageDatasetCustomLabels manages dataset custom labels (key-value pairs)
//
//	@Summary		Manage dataset custom labels
//	@Description	Add or remove custom labels (key-value pairs) for a dataset
//	@Tags			Datasets
//	@ID				update_dataset_labels
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int									true	"Dataset ID"
//	@Param			manage		body		ManageDatasetLabelReq				true	"Label management request"
//	@Success		200			{object}	dto.GenericResponse[DatasetResp]	"Labels managed successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]			"Invalid dataset ID or invalid request format/parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]			"Dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/labels [patch]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ManageDatasetCustomLabels(c *gin.Context) {
	datasetID, ok := parseDatasetID(c)
	if !ok {
		return
	}

	var req ManageDatasetLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.ManageDatasetLabels(c.Request.Context(), &req, datasetID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// CreateDatasetVersion handles dataset version creation for v2 API
//
//	@Summary		Create dataset version
//	@Description	Create a new dataset version for an existing dataset.
//	@Tags			Datasets
//	@ID				create_dataset_version
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int										true	"Dataset ID"
//	@Param			request		body		CreateDatasetVersionReq					true	"Dataset version creation request"
//	@Success		201			{object}	dto.GenericResponse[DatasetVersionResp]	"Dataset version created successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid request"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		409			{object}	dto.GenericResponse[any]				"Conflict error"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/versions [post]
//	@x-api-type		{"portal":"true"}
func (h *Handler) CreateDatasetVersion(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	datasetID, ok := parseDatasetID(c)
	if !ok {
		return
	}

	var req CreateDatasetVersionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.CreateDatasetVersion(c.Request.Context(), &req, datasetID, userID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusCreated, "Dataset version created successfully", resp)
}

// DeleteDatasetVersion handles dataset version deletion
//
//	@Summary		Delete dataset version
//	@Description	Delete a dataset version (soft delete by setting status to false)
//	@Tags			Datasets
//	@ID				delete_dataset_version
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int							true	"Dataset ID"
//	@Param			version_id	path		int							true	"Dataset Version ID"
//	@Success		204			{object}	dto.GenericResponse[any]	"Dataset version deleted successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid dataset ID/dataset version ID"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]	"Dataset or version not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/versions/{version_id} [delete]
//	@x-api-type		{"portal":"true"}
func (h *Handler) DeleteDatasetVersion(c *gin.Context) {
	versionID, ok := parseDatasetVersionID(c)
	if !ok {
		return
	}

	if httpx.HandleServiceError(c, h.service.DeleteDatasetVersion(c.Request.Context(), versionID)) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Dataset version deleted successfully", nil)
}

// GetDatasetVersion handles getting a single dataset version by ID
//
//	@Summary		Get dataset version by ID
//	@Description	Get detailed information about a specific dataset version
//	@Tags			Datasets
//	@ID				get_dataset_version_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int												true	"Dataset ID"
//	@Param			version_id	path		int												true	"Dataset Version ID"
//	@Success		200			{object}	dto.GenericResponse[DatasetVersionDetailResp]	"Dataset version retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid dataset ID/dataset version ID"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]						"Dataset or version not found"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/versions/{version_id} [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) GetDatasetVersion(c *gin.Context) {
	datasetID, ok := parseDatasetID(c)
	if !ok {
		return
	}
	versionID, ok := parseDatasetVersionID(c)
	if !ok {
		return
	}

	resp, err := h.service.GetDatasetVersion(c.Request.Context(), datasetID, versionID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListDatasetVersions handles listing dataset versions with pagination and filtering
//
//	@Summary		List dataset versions
//	@Description	Get paginated list of dataset versions for a specific dataset
//	@Tags			Datasets
//	@ID				list_dataset_versions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int														true	"Dataset ID"
//	@Param			page		query		int														false	"Page number"	default(1)
//	@Param			size		query		int														false	"Page size"		default(20)
//	@Param			status		query		consts.StatusType										false	"Dataset version status filter"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[DatasetVersionResp]]	"Dataset versions retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]								"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/versions [get]
//	@x-api-type		{"portal":"true"}
func (h *Handler) ListDatasetVersions(c *gin.Context) {
	datasetID, ok := parseDatasetID(c)
	if !ok {
		return
	}

	var req ListDatasetVersionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.ListDatasetVersions(c.Request.Context(), &req, datasetID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// UpdateDatasetVersion handles dataset version updates
//
//	@Summary		Update dataset version
//	@Description	Update an existing dataset version's information
//	@Tags			Datasets
//	@ID				update_dataset_version
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int										true	"Dataset ID"
//	@Param			version_id	path		int										true	"Dataset Version ID"
//	@Param			request		body		UpdateDatasetVersionReq					true	"Dataset version update request"
//	@Success		202			{object}	dto.GenericResponse[DatasetVersionResp]	"Dataset version updated successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid dataset ID/dataset version ID/request format/request parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]				"Dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/versions/{version_id} [patch]
//	@x-api-type		{"portal":"true"}
func (h *Handler) UpdateDatasetVersion(c *gin.Context) {
	datasetID, ok := parseDatasetID(c)
	if !ok {
		return
	}
	versionID, ok := parseDatasetVersionID(c)
	if !ok {
		return
	}

	var req UpdateDatasetVersionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.UpdateDatasetVersion(c.Request.Context(), &req, datasetID, versionID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Dataset version updated successfully", resp)
}

// DownloadDatasetVersion handles dataset file download
//
//	@Summary		Download dataset version
//	@Description	Download dataset file by version ID
//	@Tags			Datasets
//	@ID				download_dataset_version
//	@Produce		application/octet-stream
//	@Security		BearerAuth
//	@Param			dataset_id	path		int							true	"Dataset ID"
//	@Param			version_id	path		int							true	"Dataset Version ID"
//	@Success		200			{file}		binary						"Dataset file"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid dataset ID/dataset version ID"
//	@Failure		403			{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]	"Dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/versions/{version_id}/download [get]
//	@x-api-type		{"portal":"true","sdk":"true"}
func (h *Handler) DownloadDatasetVersion(c *gin.Context) {
	datasetID, ok := parseDatasetID(c)
	if !ok {
		return
	}
	versionID, ok := parseDatasetVersionID(c)
	if !ok {
		return
	}

	filename, err := h.service.GetDatasetVersionFilename(c.Request.Context(), datasetID, versionID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", filename))

	zipWriter := zip.NewWriter(c.Writer)
	defer func() { _ = zipWriter.Close() }()

	if err := h.service.DownloadDatasetVersion(c.Request.Context(), zipWriter, []utils.ExculdeRule{}, versionID); err != nil {
		delete(c.Writer.Header(), "Content-Disposition")
		c.Header("Content-Type", "application/json; charset=utf-8")
		httpx.HandleServiceError(c, err)
	}
}

// ManageDatasetInjections manages dataset injections
//
//	@Summary		Manage dataset injections
//	@Description	Add or remove injections for a dataset
//	@Tags			Datasets
//	@ID				manage_dataset_version_injections
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int												true	"Dataset ID"
//	@Param			version_id	path		int												true	"Dataset Version ID"
//	@Param			manage		body		ManageDatasetVersionInjectionReq				true	"Injection management request"
//	@Success		200			{object}	dto.GenericResponse[DatasetVersionDetailResp]	"Injections managed successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]						"Invalid dataset ID or invalid request format/parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]						"Dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/version/{version_id}/injections [patch]
//	@x-api-type		{"portal":"true","sdk":"true"}
func (h *Handler) ManageDatasetVersionInjections(c *gin.Context) {
	_, ok := parseDatasetID(c)
	if !ok {
		return
	}
	versionID, ok := parseDatasetVersionID(c)
	if !ok {
		return
	}

	var req ManageDatasetVersionInjectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := h.service.ManageDatasetVersionInjections(c.Request.Context(), &req, versionID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

func parseDatasetID(c *gin.Context) (int, bool) {
	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil || datasetID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return 0, false
	}
	return datasetID, true
}

func parseDatasetVersionID(c *gin.Context) (int, bool) {
	versionIDStr := c.Param(consts.URLPathVersionID)
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil || versionID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset version ID")
		return 0, false
	}
	return versionID, true
}
