package containermodule

import (
	"aegis/httpx"
	"context"
	"net/http"
	"path/filepath"
	"strconv"

	"aegis/consts"
	"aegis/dto"
	"aegis/middleware"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// CreateContainer handles container creation for v2 API
//
//	@Summary		Create container
//	@Description	Create a new container without build configuration.
//	@Tags			Containers
//	@ID				create_container
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		CreateContainerReq						true	"Container creation request"
//	@Success		201		{object}	dto.GenericResponse[ContainerResp]		"Container created successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		409		{object}	dto.GenericResponse[any]				"Conflict error"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/containers [post]
//	@x-api-type		{}
func (h *Handler) CreateContainer(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req CreateContainerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.CreateContainer(c.Request.Context(), &req, userID)
	if httpx.HandleServiceError(c, err) {
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
//	@x-api-type		{}
func (h *Handler) DeleteContainer(c *gin.Context) {
	containerID, ok := parseContainerID(c)
	if !ok {
		return
	}

	if httpx.HandleServiceError(c, h.service.DeleteContainer(c.Request.Context(), containerID)) {
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
//	@Success		200				{object}	dto.GenericResponse[ContainerDetailResp]	"Container retrieved successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]						"Invalid container ID"
//	@Failure		401				{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]						"Container not found"
//	@Failure		500				{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/containers/{container_id} [get]
//	@x-api-type		{}
func (h *Handler) GetContainer(c *gin.Context) {
	containerID, ok := parseContainerID(c)
	if !ok {
		return
	}

	resp, err := h.service.GetContainer(c.Request.Context(), containerID)
	if httpx.HandleServiceError(c, err) {
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
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[ContainerResp]]	"Containers retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]								"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/containers [get]
//	@x-api-type		{}
func (h *Handler) ListContainers(c *gin.Context) {
	var req ListContainerReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListContainers(c.Request.Context(), &req)
	if httpx.HandleServiceError(c, err) {
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
//	@Param			request			body		UpdateContainerReq						true	"Container update request"
//	@Success		202				{object}	dto.GenericResponse[ContainerResp]		"Container updated successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]				"Invalid container ID/request"
//	@Failure		401				{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]				"Container not found"
//	@Failure		500				{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/containers/{container_id} [patch]
//	@x-api-type		{}
func (h *Handler) UpdateContainer(c *gin.Context) {
	containerID, ok := parseContainerID(c)
	if !ok {
		return
	}

	var req UpdateContainerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	resp, err := h.service.UpdateContainer(c.Request.Context(), &req, containerID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Container updated successfully", resp)
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
//	@Param			manage			body		ManageContainerLabelReq					true	"Label management request"
//	@Success		200				{object}	dto.GenericResponse[ContainerResp]		"Labels managed successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]				"Invalid container ID or invalid request format/parameters"
//	@Failure		401				{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]				"Container not found"
//	@Failure		500				{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/containers/{container_id}/labels [patch]
//	@x-api-type		{}
func (h *Handler) ManageContainerCustomLabels(c *gin.Context) {
	containerID, ok := parseContainerID(c)
	if !ok {
		return
	}

	var req ManageContainerLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ManageContainerLabels(c.Request.Context(), &req, containerID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

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
//	@Param			request			body		CreateContainerVersionReq					true	"Container version creation request"
//	@Success		201				{object}	dto.GenericResponse[ContainerVersionResp]	"Container version created successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]						"Invalid container ID or invalid request format or parameters"
//	@Failure		401				{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		409				{object}	dto.GenericResponse[any]						"Conflict error"
//	@Failure		500				{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/containers/{container_id}/versions [post]
//	@x-api-type		{}
func (h *Handler) CreateContainerVersion(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	containerID, ok := parseContainerID(c)
	if !ok {
		return
	}

	var req CreateContainerVersionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.CreateContainerVersion(c.Request.Context(), &req, containerID, userID)
	if httpx.HandleServiceError(c, err) {
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
//	@x-api-type		{}
func (h *Handler) DeleteContainerVersion(c *gin.Context) {
	versionID, ok := parseVersionID(c, "Invalid container version ID")
	if !ok {
		return
	}

	if httpx.HandleServiceError(c, h.service.DeleteContainerVersion(c.Request.Context(), versionID)) {
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
//	@Success		200				{object}	dto.GenericResponse[ContainerVersionDetailResp]	"Container version retrieved successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]							"Invalid container ID/container version ID"
//	@Failure		401				{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]							"Container or version not found"
//	@Failure		500				{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/containers/{container_id}/versions/{version_id} [get]
//	@x-api-type		{}
func (h *Handler) GetContainerVersion(c *gin.Context) {
	containerID, ok := parseContainerID(c)
	if !ok {
		return
	}
	versionID, ok := parseVersionID(c, "Invalid container version ID")
	if !ok {
		return
	}

	resp, err := h.service.GetContainerVersion(c.Request.Context(), containerID, versionID)
	if httpx.HandleServiceError(c, err) {
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
//	@Success		200				{object}	dto.GenericResponse[dto.ListResp[ContainerVersionResp]]	"Container versions retrieved successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]									"Invalid request format or parameters"
//	@Failure		401				{object}	dto.GenericResponse[any]									"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]									"Permission denied"
//	@Failure		500				{object}	dto.GenericResponse[any]									"Internal server error"
//	@Router			/api/v2/containers/{container_id}/versions [get]
//	@x-api-type		{}
func (h *Handler) ListContainerVersions(c *gin.Context) {
	containerID, ok := parseContainerID(c)
	if !ok {
		return
	}

	var req ListContainerVersionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.ListContainerVersions(c.Request.Context(), &req, containerID)
	if httpx.HandleServiceError(c, err) {
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
//	@Param			request			body		UpdateContainerVersionReq					true	"Container version update request"
//	@Success		202				{object}	dto.GenericResponse[ContainerVersionResp]	"Container version updated successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]						"Invalid container ID/container version ID/request"
//	@Failure		401				{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]						"Container not found"
//	@Failure		500				{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/containers/{container_id}/versions/{version_id} [patch]
//	@x-api-type		{}
func (h *Handler) UpdateContainerVersion(c *gin.Context) {
	containerID, ok := parseContainerID(c)
	if !ok {
		return
	}
	versionID, ok := parseVersionID(c, "Invalid container version ID")
	if !ok {
		return
	}

	var req UpdateContainerVersionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	resp, err := h.service.UpdateContainerVersion(c.Request.Context(), &req, containerID, versionID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusAccepted, "Container version updated successfully", resp)
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
//	@Param			request	body		SubmitBuildContainerReq							true	"Container build request"
//	@Success		200		{object}	dto.GenericResponse[SubmitContainerBuildResp]	"Container build task submitted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]							"Required files not found"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/containers/build [post]
//	@x-api-type		{}
func (h *Handler) SubmitContainerBuilding(c *gin.Context) {
	groupID := c.GetString("groupID")
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	var req SubmitBuildContainerReq
	if err := c.ShouldBind(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := h.service.SubmitContainerBuilding(spanContextFromGin(c), &req, groupID, userID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Container building task submitted successfully", resp)
}

// UploadHelmChart handles uploading Helm chart package
//
//	@Summary		Upload Helm chart package
//	@Description	Upload a Helm chart package (.tgz) and save it to local storage as fallback
//	@Tags			Containers
//	@ID				upload_helm_chart
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			container_id	path		int												true	"Container ID"
//	@Param			version_id		path		int												true	"Container Version ID"
//	@Param			file			formData	file											true	"Helm chart package (.tgz)"
//	@Success		200				{object}	dto.GenericResponse[UploadHelmChartResp]	"Chart uploaded successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]						"Invalid request or file"
//	@Failure		401				{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]						"Container or version not found"
//	@Failure		500				{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/containers/{container_id}/versions/{version_id}/helm-chart [post]
//	@x-api-type		{}
func (h *Handler) UploadHelmChart(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	containerID, ok := parseContainerID(c)
	if !ok {
		return
	}
	versionID, ok := parseVersionID(c, "Invalid container version ID")
	if !ok {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "No file uploaded or invalid file: "+err.Error())
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".tgz" && ext != ".gz" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid file type: only .tgz or .tar.gz files are allowed")
		return
	}

	resp, err := h.service.UploadHelmChart(c.Request.Context(), file, containerID, versionID, userID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
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
//	@Param			container_id	path		int													true	"Container ID"
//	@Param			version_id		path		int													true	"Container Version ID"
//	@Param			file			formData	file												true	"Helm values YAML file"
//	@Success		200				{object}	dto.GenericResponse[UploadHelmValueFileResp]	"File uploaded successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]							"Invalid request or file"
//	@Failure		401				{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404				{object}	dto.GenericResponse[any]							"Container or version not found"
//	@Failure		500				{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/containers/{container_id}/versions/{version_id}/helm-values [post]
//	@x-api-type		{}
func (h *Handler) UploadHelmValueFile(c *gin.Context) {
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists || userID <= 0 {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	containerID, ok := parseContainerID(c)
	if !ok {
		return
	}
	versionID, ok := parseVersionID(c, "Invalid container version ID")
	if !ok {
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "No file uploaded or invalid file: "+err.Error())
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext != ".yaml" && ext != ".yml" {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid file type: only .yaml or .yml files are allowed")
		return
	}

	resp, err := h.service.UploadHelmValueFile(c.Request.Context(), file, containerID, versionID, userID)
	if httpx.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

func parseContainerID(c *gin.Context) (int, bool) {
	containerIDStr := c.Param(consts.URLPathContainerID)
	containerID, err := strconv.Atoi(containerIDStr)
	if err != nil || containerID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid container ID")
		return 0, false
	}
	return containerID, true
}

func parseVersionID(c *gin.Context, message string) (int, bool) {
	versionIDStr := c.Param(consts.URLPathVersionID)
	versionID, err := strconv.Atoi(versionIDStr)
	if err != nil || versionID <= 0 {
		dto.ErrorResponse(c, http.StatusBadRequest, message)
		return 0, false
	}
	return versionID, true
}

func spanContextFromGin(c *gin.Context) context.Context {
	ctx, ok := c.Get(middleware.SpanContextKey)
	if ok {
		if spanCtx, ok := ctx.(context.Context); ok {
			return spanCtx
		}
	}
	return c.Request.Context()
}
