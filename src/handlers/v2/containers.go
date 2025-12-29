package v2

import (
	"context"
	"net/http"
	"path/filepath"
	"strconv"

	"aegis/consts"
	"aegis/dto"
	"aegis/handlers"
	"aegis/middleware"
	producer "aegis/service/prodcuer"

	"github.com/gin-gonic/gin"
)

// ===================== Container =====================

// CreateContainer handles container creation for v2 API
//
//	@Summary		Create container
//	@Description	Create a new container without build configuration.
//	@Tags			Containers
//	@ID				create_container
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.CreateContainerReq					true	"Container creation request"
//	@Success		201		{object}	dto.GenericResponse[dto.ContainerResp]	"Container created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		409		{object}	dto.GenericResponse[any]				"Conflict error"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/containers [post]
//	@x-api-type		{"sdk":"true"}
func CreateContainer(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req dto.CreateContainerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.CreateContainer(&req, userID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusCreated, "Container created successfully", resp)
}

// DeleteContainer handles container deletion
//
//	@Summary		Delete container
//	@Description	Delete a container (soft delete by setting status to -1)
//	@Tags			Containers
//	@ID				delete_container
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_id	path		int							true	"Container ID"
//	@Success		204				{object}	dto.GenericResponse[any]	"Container deleted successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]	"Invalid container ID"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]	"Container not found"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/containers/{container_id} [delete]
func DeleteContainer(c *gin.Context) {
	containerIDStr := c.Param(consts.URLPathContainerID)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	err = producer.DeleteContainer(containerID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Container deleted successfully", nil)
}

// GetContainer handles getting a single container by ID
//
//	@Summary		Get container by ID
//	@Description	Get detailed information about a specific container
//	@Tags			Containers
//	@ID				get_container_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_id	path		int												true	"Container ID"
//	@Success		200				{object}	dto.GenericResponse[dto.ContainerDetailResp]	"Container retrieved successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]						"Invalid container ID"
//	@Failure		401				{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]						"Container not found"
//	@Failure		500				{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/containers/{container_id} [get]
//	@x-api-type		{"sdk":"true"}
func GetContainer(c *gin.Context) {
	containerIDStr := c.Param(consts.URLPathContainerID)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil || containerID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	resp, err := producer.GetContainerDetail(containerID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListContainers handles listing containers with pagination and filtering
//
//	@Summary		List containers
//	@Description	Get paginated list of containers with pagination and filtering
//	@Tags			Containers
//	@ID				list_containers
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int														false	"Page number"	default(1)
//	@Param			size		query		consts.PageSize											false	"Page size"		default(20)
//	@Param			type		query		consts.ContainerType									false	"Container type filter"
//	@Param			is_public	query		bool													false	"Container public visibility filter"
//	@Param			status		query		consts.StatusType										false	"Container status filter"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.ContainerResp]]	"Containers retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]								"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/containers [get]
//	@x-api-type		{"sdk":"true"}
func ListContainers(c *gin.Context) {
	var req dto.ListContainerReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListContainers(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// UpdateContainer handles container updates
//
//	@Summary		Update container
//	@Description	Update an existing container's information
//	@Tags			Containers
//	@ID				update_container
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_id	path		int										true	"Container ID"
//	@Param			request			body		dto.UpdateContainerReq					true	"Container update request"
//	@Success		202				{object}	dto.GenericResponse[dto.ContainerResp]	"Container updated successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]				"Invalid container ID/request"
//	@Failure		401				{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]				"Container not found"
//	@Failure		500				{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/containers/{container_id} [patch]
func UpdateContainer(c *gin.Context) {
	containerIDStr := c.Param(consts.URLPathContainerID)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	var req dto.UpdateContainerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	resp, err := producer.UpdateContainer(&req, containerID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Container updated successfully", resp)
}

// ===================== Container Version =====================

// CreateContainerVersion handles container version creation for v2 API
//
//	@Summary		Create container version
//	@Description	Create a new container version for an existing container.
//	@Tags			Containers
//	@ID				create_container_version
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_id	path		int												true	"Container ID"
//	@Param			request			body		dto.CreateContainerVersionReq					true	"Container version creation request"
//	@Success		201				{object}	dto.GenericResponse[dto.ContainerVersionResp]	"Container version created successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]						"Invalid container ID or invalid request format or parameters"
//	@Failure		401				{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		409				{object}	dto.GenericResponse[any]						"Conflict error"
//	@Failure		500				{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/containers/{container_id}/versions [post]
//	@x-api-type		{"sdk":"true"}
func CreateContainerVersion(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	containerIDStr := c.Param(consts.URLPathContainerID)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil || containerID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	var req dto.CreateContainerVersionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.CreateContainerVersion(&req, containerID, userID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusCreated, "Container version created successfully", resp)
}

// DeleteContainerVersion handles container version deletion
//
//	@Summary		Delete container version
//	@Description	Delete a container version (soft delete by setting status to false)
//	@Tags			Containers
//	@ID				delete_container_version
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_id	path		int							true	"Container ID"
//	@Param			version_id		path		int							true	"Container Version ID"
//	@Success		204				{object}	dto.GenericResponse[any]	"Container version deleted successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]	"Invalid container ID or container version ID"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]	"Container or version not found"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/containers/{container_id}/versions/{version_id} [delete]
func DeleteContainerVersion(c *gin.Context) {
	versionIDStr := c.Param(consts.URLPathVersionID)
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil || versionID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container version ID")
		return
	}

	err = producer.DeleteContainerVersion(versionID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Container version deleted successfully", nil)
}

// GetContainerVersion handles getting a single container version by ID
//
//	@Summary		Get container version by ID
//	@Description	Get detailed information about a specific container version
//	@Tags			Containers
//	@ID				get_container_version_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_id	path		int													true	"Container ID"
//	@Param			version_id		path		int													true	"Container Version ID"
//	@Success		200				{object}	dto.GenericResponse[dto.ContainerVersionDetailResp]	"Container version retrieved successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]							"Invalid container ID/container version ID"
//	@Failure		401				{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]							"Container or version not found"
//	@Failure		500				{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/containers/{container_id}/versions/{version_id} [get]
//	@x-api-type		{"sdk":"true"}
func GetContainerVersion(c *gin.Context) {
	containerIDStr := c.Param(consts.URLPathContainerID)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil || containerID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	versionIDStr := c.Param(consts.URLPathVersionID)
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil || versionID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container version ID")
		return
	}

	resp, err := producer.GetContainerVersionDetail(containerID, versionID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// ListContainerVersions handles listing container versions with pagination and filtering
//
//	@Summary		List container versions
//	@Description	Get paginated list of container versions for a specific container
//	@Tags			Containers
//	@ID				list_container_versions
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_id	path		int															true	"Container ID"
//	@Param			page			query		int															false	"Page number"	default(1)
//	@Param			size			query		int															false	"Page size"		default(20)
//	@Param			status			query		consts.StatusType											false	"Container version status filter"
//	@Success		200				{object}	dto.GenericResponse[dto.ListResp[dto.ContainerVersionResp]]	"Container versions retrieved successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]									"Invalid request format or parameters"
//	@Failure		401				{object}	dto.GenericResponse[any]									"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]									"Permission denied"
//	@Failure		500				{object}	dto.GenericResponse[any]									"Internal server error"
//	@Router			/api/v2/containers/{container_id}/versions [get]
//	@x-api-type		{"sdk":"true"}
func ListContainerVersions(c *gin.Context) {
	containerIDStr := c.Param(consts.URLPathContainerID)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil || containerID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	var req dto.ListContainerVersionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListContainerVersions(&req, containerID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Container versions retrieved successfully", resp)
}

// UpdateContainerVersion handles container version updates
//
//	@Summary		Update container version
//	@Description	Update an existing container version's information
//	@Tags			Containers
//	@ID				update_container_version
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_id	path		int												true	"Container ID"
//	@Param			version_id		path		int												true	"Container Version ID"
//	@Param			request			body		dto.UpdateContainerVersionReq					true	"Container version update request"
//	@Success		202				{object}	dto.GenericResponse[dto.ContainerVersionResp]	"Container version updated successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]						"Invalid container ID/container version ID/request"
//	@Failure		401				{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]						"Container not found"
//	@Failure		500				{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/containers/{container_id}/versions/{version_id} [patch]
func UpdateContainerVersion(c *gin.Context) {
	containerIDStr := c.Param(consts.URLPathContainerID)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	versionIDStr := c.Param(consts.URLPathVersionID)
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container version ID")
		return
	}

	var req dto.UpdateContainerVersionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	resp, err := producer.UpdateContainerVersion(&req, containerID, versionID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Container version updated successfully", resp)
}

// ManageContainerCustomLabels manages container custom labels (key-value pairs)
//
//	@Summary		Manage container custom labels
//	@Description	Add or remove custom labels (key-value pairs) for a container
//	@Tags			Containers
//	@ID				manage_container_labels
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_id	path		int										true	"Container ID"
//	@Param			manage			body		dto.ManageContainerLabelReq				true	"Label management request"
//	@Success		200				{object}	dto.GenericResponse[dto.ContainerResp]	"Labels managed successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]				"Invalid container ID or invalid request format/parameters"
//	@Failure		401				{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]				"Container not found"
//	@Failure		500				{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/containers/{container_id}/labels [patch]
func ManageContainerCustomLabels(c *gin.Context) {
	containerIDStr := c.Param(consts.URLPathContainerID)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil || containerID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	var req dto.ManageContainerLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ManageContainerLabels(&req, containerID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// SubmitContainerBuilding handles submitting a container build task
//
//	@Summary		Submit container building
//	@Description	Submit a container build task to build a container image from provided source files.
//	@Tags			Containers
//	@ID				build_container_image
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		dto.SubmitBuildContainerReq							true	"Container build request"
//	@Success		200		{object}	dto.GenericResponse[dto.SubmitContainerBuildResp]	"Container build task submitted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]							"Required files not found"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/containers/build [post]
//	@x-api-type		{"sdk":"true"}
func SubmitContainerBuilding(c *gin.Context) {
	groupID := c.GetString("groupID")
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to get span context")
		return
	}
	spanCtx := ctx.(context.Context)

	var req dto.SubmitBuildContainerReq
	if err := c.ShouldBind(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ProduceContainerBuildingTask(spanCtx, &req, groupID, userID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Container building task submitted successfully", resp)
}

// UploadHelmValueFile handles uploading Helm values file
//
//	@Summary		Upload Helm values file
//	@Description	Upload a Helm values YAML file and save it to JuiceFS storage
//	@Tags			Containers
//	@ID				upload_helm_value_file
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_id	path		int									true	"Container ID"
//	@Param			version_id		path		int									true	"Container Version ID"
//	@Param			file			formData	file								true	"Helm values YAML file"
//	@Success		200				{object}	dto.GenericResponse[dto.UploadHelmValueFileResp]	"File uploaded successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]			"Invalid request or file"
//	@Failure		401				{object}	dto.GenericResponse[any]			"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]			"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]			"Container or version not found"
//	@Failure		500				{object}	dto.GenericResponse[any]			"Internal server error"
//	@Router			/api/v2/containers/{container_id}/versions/{version_id}/helm-values [post]
func UploadHelmValueFile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	containerIDStr := c.Param(consts.URLPathContainerID)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil || containerID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return
	}

	versionIDStr := c.Param(consts.URLPathVersionID)
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil || versionID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container version ID")
		return
	}

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "No file uploaded or invalid file: "+err.Error())
		return
	}

	filename := file.Filename
	ext := filepath.Ext(filename)
	if ext != ".yaml" && ext != ".yml" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid file type: only .yaml or .yml files are allowed")
		return
	}

	// Call service layer to handle file upload
	resp, err := producer.UploadHelmValueFile(file, containerID, versionID, userID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}
