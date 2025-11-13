package v2

import (
	"archive/zip"
	"fmt"
	"net/http"
	"strconv"

	"aegis/consts"
	"aegis/dto"
	"aegis/handlers"
	"aegis/middleware"
	producer "aegis/service/prodcuer"
	"aegis/utils"

	"github.com/gin-gonic/gin"
)

// CreateDataset handles dataset creation
//
//	@Summary		Create dataset
//	@Description	Create a new dataset with an initial version
//	@Tags			Datasets
//	@ID				create_dataset
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateDatasetReq					true	"Dataset creation request"
//	@Success		201		{object}	dto.GenericResponse[dto.DatasetResp]	"Dataset created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		409		{object}	dto.GenericResponse[any]				"Conflict error"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/datasets [post]
func CreateDataset(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dto.CreateDatasetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.CreateDataset(&req, userID)
	if handlers.HandleServiceError(c, err) {
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
func DeleteDataset(c *gin.Context) {
	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	err = producer.DeleteDataset(datasetID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusCreated, "Dataset deleted successfully", nil)
}

// GetDataset handles getting a single dataset by ID
//
//	@Summary		Get dataset by ID
//	@Description	Get detailed information about a specific dataset
//	@Tags			Datasets
//	@ID				get_dataset_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int											true	"Dataset ID"
//	@Success		200			{object}	dto.GenericResponse[dto.DatasetDetailResp]	"Dataset retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]					"Invalid dataset ID"
//	@Failure		401			{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]					"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]					"Dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id} [get]
func GetDataset(c *gin.Context) {
	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	resp, err := producer.GetDatasetDetail(datasetID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			page	query		int													false	"Page number"	default(1)
//	@Param			size	query		int													false	"Page size"		default(20)
//	@Param			type	query		string												false	"Dataset type filter"
//	@Param			status	query		int													false	"Dataset status filter"
//	@Success		200		{object}	dto.GenericResponse[dto.ListResp[dto.DatasetResp]]	"Datasets retrieved successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/datasets [get]
func ListDatasets(c *gin.Context) {
	var req dto.ListDatasetReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListDatasets(&req)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			dataset_id	path		int										true	"Dataset ID"
//	@Param			request		body		dto.UpdateDatasetReq					true	"Dataset update request"
//	@Success		202			{object}	dto.GenericResponse[dto.DatasetResp]	"Dataset updated successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid dataset ID/request"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]				"Dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id} [patch]
func UpdateDataset(c *gin.Context) {
	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	var req dto.UpdateDatasetReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	resp, err := producer.UpdateDataset(&req, datasetID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Dataset updated successfully", resp)
}

// ===================== Dataset-Label API =====================

// ManageDatasetCustomLabels manages dataset custom labels (key-value pairs)
//
//	@Summary		Manage dataset custom labels
//	@Description	Add or remove custom labels (key-value pairs) for a dataset
//	@Tags			Datasets
//	@ID				update_dataset_labels
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int										true	"Dataset ID"
//	@Param			manage		body		dto.ManageDatasetLabelReq				true	"Label management request"
//	@Success		200			{object}	dto.GenericResponse[dto.DatasetResp]	"Labels managed successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]				"Invalid dataset ID or invalid request format/parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]				"Dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/labels [patch]
func ManageDatasetCustomLabels(c *gin.Context) {
	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil || datasetID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	var req dto.ManageDatasetLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ManageDatasetLabels(&req, datasetID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			dataset_id	path		int											true	"Dataset ID"
//	@Param			request		body		dto.CreateDatasetVersionReq					true	"Dataset version creation request"
//	@Success		201			{object}	dto.GenericResponse[dto.DatasetVersionResp]	"Dataset version created successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]					"Invalid request"
//	@Failure		401			{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]					"Permission denied"
//	@Failure		409			{object}	dto.GenericResponse[any]					"Conflict error"
//	@Failure		500			{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/versions [post]
func CreateDatasetVersion(c *gin.Context) {
	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil || datasetID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	var req dto.CreateDatasetVersionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.CreateDatasetVersion(&req, datasetID)
	if handlers.HandleServiceError(c, err) {
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
func DeleteDatasetVersion(c *gin.Context) {
	versionIDStr := c.Param(consts.URLPathVersionID)
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset version ID")
		return
	}

	err = producer.DeleteDatasetVersion(versionID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			dataset_id	path		int													true	"Dataset ID"
//	@Param			version_id	path		int													true	"Dataset Version ID"
//	@Success		200			{object}	dto.GenericResponse[dto.DatasetVersionDetailResp]	"Dataset version retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]							"Invalid dataset ID/dataset version ID"
//	@Failure		401			{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]							"Dataset or version not found"
//	@Failure		500			{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/versions/{version_id} [get]
func GetDatasetVersion(c *gin.Context) {
	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil || datasetID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	versionIDStr := c.Param(consts.URLPathVersionID)
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset version ID")
		return
	}

	resp, err := producer.GetDatasetVersionDetail(datasetID, versionID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			dataset_id	path		int															true	"Dataset ID"
//	@Param			page		query		int															false	"Page number"	default(1)
//	@Param			size		query		int															false	"Page size"		default(20)
//	@Param			status		query		int															false	"Dataset version status filter"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.DatasetVersionResp]]	"Dataset versions retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]									"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]									"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]									"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]									"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/versions [get]
func ListDatasetVersions(c *gin.Context) {
	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil || datasetID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	var req dto.ListDatasetVersionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListDatasetVersions(&req, datasetID)
	if handlers.HandleServiceError(c, err) {
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
//	@Param			dataset_id	path		int											true	"Dataset ID"
//	@Param			version_id	path		int											true	"Dataset Version ID"
//	@Param			request		body		dto.UpdateDatasetVersionReq					true	"Dataset version update request"
//	@Success		202			{object}	dto.GenericResponse[dto.DatasetVersionResp]	"Dataset version updated successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]					"Invalid dataset ID/dataset version ID/request format/request parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]					"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]					"Dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/versions/{version_id} [patch]
func UpdateDatasetVersion(c *gin.Context) {
	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	versionIDStr := c.Param(consts.URLPathVersionID)
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset version ID")
		return
	}

	var req dto.UpdateDatasetVersionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.UpdateDatasetVersion(&req, datasetID, versionID)
	if handlers.HandleServiceError(c, err) {
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
func DownloadDatasetVersion(c *gin.Context) {
	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil || datasetID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	versionIDStr := c.Param(consts.URLPathVersionID)
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil || versionID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid version ID")
		return
	}

	filename, err := producer.GetFilename(datasetID, versionID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", filename))

	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	if err := producer.DownloadDatasetVersion(zipWriter, []utils.ExculdeRule{}, versionID); err != nil {
		delete(c.Writer.Header(), "Content-Disposition")
		c.Header("Content-Type", "application/json; charset=utf-8")
		handlers.HandleServiceError(c, err)
	}
}

// ===================== DatasetVersion-Injection API =====================

// ManageDatasetInjections manages dataset injections
//
//	@Summary		Manage dataset injections
//	@Description	Add or remove injections for a dataset
//	@Tags			Datasets
//	@ID				link_injections_to_dataset_version
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			dataset_id	path		int													true	"Dataset ID"
//	@Param			version_id	path		int													true	"Dataset Version ID"
//	@Param			manage		body		dto.ManageDatasetVersionInjectionReq				true	"Injection management request"
//	@Success		200			{object}	dto.GenericResponse[dto.DatasetVersionDetailResp]	"Injections managed successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]							"Invalid dataset ID or invalid request format/parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]							"Dataset not found"
//	@Failure		500			{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/datasets/{dataset_id}/version/{version_id}/injections [patch]
func ManageDatasetVersionInjections(c *gin.Context) {
	datasetIDStr := c.Param(consts.URLPathDatasetID)
	datasetID, err := strconv.Atoi(datasetIDStr)
	if err != nil || datasetID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset ID")
		return
	}

	versionIDStr := c.Param(consts.URLPathVersionID)
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil || versionID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset version ID")
		return
	}

	var req dto.ManageDatasetVersionInjectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ManageDatasetVersionInjections(&req, versionID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}
