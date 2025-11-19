package v2

import (
	"aegis/consts"
	"net/http"
	"strconv"

	"aegis/dto"
	"aegis/handlers"
	producer "aegis/service/prodcuer"

	"github.com/gin-gonic/gin"
)

// BatchDeleteLabels handles batch deletion of labels
//
//	@Summary		Batch delete labels
//	@Description	Batch delete labels by IDs with cascading deletion of related records
//	@Tags			Labels
//	@ID				batch_delete_labels
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.BatchDeleteLabelReq		true	"Batch delete request"
//	@Success		200		{object}	dto.GenericResponse[any]	"Labels deleted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/labels/batch-delete [post]
func BatchDeleteLabels(c *gin.Context) {
	var req dto.BatchDeleteLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	err := producer.BatchDeleteLabels(req.IDs)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Labels deleted successfully", nil)
}

// CreateLabel creates a new label
//
//	@Summary		Create label
//	@Description	Create a new label with key-value pair. If a deleted label with same key-value exists, it will be restored and updated.
//	@Tags			Labels
//	@ID				create_label
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			label	body		dto.CreateLabelReq					true	"Label creation request"
//	@Success		201		{object}	dto.GenericResponse[dto.LabelResp]	"Label created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]			"Invalid request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		409		{object}	dto.GenericResponse[any]			"Label already exists"
//	@Failure		500		{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/labels [post]
func CreateLabel(c *gin.Context) {
	var req dto.CreateLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format:"+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.CreateLabel(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "Label created successfully", resp)
}

// DeleteLabel handles label deletion
//
//	@Summary		Delete label
//	@Description	Delete a label and remove all its associations
//	@Tags			Labels
//	@ID				delete_label
//	@Produce		json
//	@Security		BearerAuth
//	@Param			label_id	path		int							true	"Label ID"
//	@Success		204			{object}	dto.GenericResponse[any]	"Label deleted successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]	"Invalid label ID"
//	@Failure		401			{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]	"Label not found"
//	@Failure		500			{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/labels/{label_id} [delete]
func DeleteLabel(c *gin.Context) {
	labelIdStr := c.Param(consts.URLPathLabelID)
	labelID, err := strconv.Atoi(labelIdStr)
	if err != nil || labelID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid label ID")
		return
	}

	err = producer.DeleteLabel(labelID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Label deleted successfully", nil)
}

// GetLabelDetail handles getting a single label by ID
//
//	@Summary		Get label by ID
//	@Description	Get detailed information about a specific label
//	@Tags			Labels
//	@ID				get_label_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			label_id	path		int											true	"Label ID"
//	@Success		200			{object}	dto.GenericResponse[dto.LabelDetailResp]	"Label retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]					"Invalid label ID"
//	@Failure		401			{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]					"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]					"Label not found"
//	@Failure		500			{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/labels/{label_id} [get]
func GetLabelDetail(c *gin.Context) {
	labelIdStr := c.Param(consts.URLPathLabelID)
	labelID, err := strconv.Atoi(labelIdStr)
	if err != nil || labelID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid label ID")
		return
	}

	resp, err := producer.GetLabelDetail(labelID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListLabels handles listing labels with pagination and filtering
//
//	@Summary		List labels
//	@Description	Get paginated list of labels with filtering
//	@Tags			Labels
//	@ID				list_labels
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int													false	"Page number"	default(1)
//	@Param			size		query		int													false	"Page size"		default(20)
//	@Param			key			query		string												false	"Filter by label key"
//	@Param			value		query		string												false	"Filter by label value"
//	@Param			category	query		consts.LabelCategory								false	"Filter by category"
//	@Param			is_system	query		bool												false	"Filter by system label"
//	@Param			status		query		consts.StatusType									false	"Filter by status"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.LabelResp]]	"Labels retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]							"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/labels [get]
func ListLabels(c *gin.Context) {
	var req dto.ListLabelReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListLabels(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// UpdateLabel handles label updates
//
//	@Summary		Update label
//	@Description	Update an existing label's information
//	@Tags			Labels
//	@ID				update_label
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			label_id	path		int									true	"Label ID"
//	@Param			request		body		dto.UpdateLabelReq					true	"Label update request"
//	@Success		202			{object}	dto.GenericResponse[dto.LabelResp]	"Label updated successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]			"Invalid label ID or invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404			{object}	dto.GenericResponse[any]			"Label not found"
//	@Failure		500			{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/labels/{label_id} [patch]
func UpdateLabel(c *gin.Context) {
	labelIdStr := c.Param(consts.URLPathLabelID)
	labelID, err := strconv.Atoi(labelIdStr)
	if err != nil || labelID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid label ID")
		return
	}

	var req dto.UpdateLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.UpdateLabel(&req, labelID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusAccepted, "Label updated successfully", resp)
}
