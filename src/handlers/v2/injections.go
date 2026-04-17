package v2

import (
	"aegis/consts"
	"aegis/utils"
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"aegis/dto"
	"aegis/handlers"
	"aegis/middleware"
	producer "aegis/service/producer"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
)

// BatchDeleteInjections
//
//	@Summary		Batch delete injections
//	@Description	Batch delete injections by IDs or labels or tags with cascading deletion of related records
//	@Tags			Injections
//	@ID				batch_delete_injections
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			batch_delete	body		dto.BatchDeleteInjectionReq	true	"Batch delete request"
//	@Success		200				{object}	dto.GenericResponse[any]	"Injections deleted successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]	"Invalid request"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/batch-delete [post]
func BatchDeleteInjections(c *gin.Context) {
	var req dto.BatchDeleteInjectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	var err error
	if len(req.IDs) > 0 {
		err = producer.BatchDeleteInjectionsByIDs(req.IDs)
	} else {
		err = producer.BatchDeleteInjectionsByLabels(req.Labels)
	}

	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse[any](c, http.StatusNoContent, "Injections deleted successfully", nil)
}

// GetInjection handles getting a single injection by ID
//
//	@Summary		Get injection by ID
//	@Description	Get detailed information about a specific injection
//	@Tags			Injections
//	@ID				get_injection_by_id
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int												true	"Injection ID"
//	@Success		200	{object}	dto.GenericResponse[dto.InjectionDetailResp]	"Injection retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]						"Invalid injection ID"
//	@Failure		401	{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403	{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]						"Injection not found"
//	@Failure		500	{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/injections/{id} [get]
//	@x-api-type		{"sdk":"true"}
func GetInjection(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "injection ID")
	if !ok {
		logrus.WithField("idStr", idStr).Warn("GetInjection: invalid ID format or ID <= 0")
		return
	}

	resp, err := producer.GetInjectionDetail(id)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":    id,
			"error": err.Error(),
		}).Error("GetInjection: failed to get injection detail")
	}

	if handlers.HandleServiceError(c, err) {
		return
	}

	logrus.WithField("id", id).Info("GetInjection: successfully retrieved injection")
	dto.SuccessResponse(c, resp)
}

// GetInjectionMetadata
//
//	@Summary		Get Injection Metadata
//	@Description	Get injection-related metadata including configuration, field mappings, and system resources
//	@Tags			Injections
//	@ID				get_injection_metadata
//	@Produce		json
//	@Security		BearerAuth
//	@Param			system	query		chaos.SystemType								true	"System for config and resources metadata"
//	@Success		200		{object}	dto.GenericResponse[dto.InjectionMetadataResp]	"Successfully returned metadata"
//	@Failure		400		{object}	dto.GenericResponse[any]						"Invalid system"
//	@Failure		401		{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]						"Resource not found"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/injections/metadata [get]
//	@x-api-type		{"sdk":"true"}
func GetInjectionMetadata(c *gin.Context) {
	systemStr := c.Query("system")

	ctx := context.Background()
	system := chaos.SystemType(systemStr)

	confNode, err := chaos.StructToNode[chaos.InjectionConf](string(system))
	if err != nil {
		// K8s namespace/pods may not exist in dev environment — return partial metadata
		logrus.Warnf("Failed to build injection config node: %v, continuing with nil config", err)
	}

	faultResourceMap, err := chaos.GetChaosTypeResourceMappings()
	if err != nil {
		handlers.HandleServiceError(c, err)
		return
	}

	resourceMap, err := chaos.GetSystemResourceMap(ctx)
	if err != nil {
		// Some systems may not be deployed in the current environment
		logrus.Warnf("Failed to get system resource map: %v, using empty map", err)
		resourceMap = make(map[chaos.SystemType]chaos.SystemResource)
	}

	resource := resourceMap[system]

	dto.SuccessResponse(c, &dto.InjectionMetadataResp{
		Config:           confNode,
		FaultTypeMap:     chaos.ChaosTypeMap,
		FaultResourceMap: faultResourceMap,
		SystemResource:   resource,
	})
}

// ListInjections handles listing injections with pagination and filtering
//
//	@Summary		List injections
//	@Description	Get a paginated list of injections with pagination and filtering
//	@Tags			Injections
//	@ID				list_injections
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int														false	"Page number"	default(1)
//	@Param			size		query		int														false	"Page size"		default(20)
//	@Param			type		query		chaos.ChaosType											false	"Filter by fault type"
//	@Param			benchmark	query		string													false	"Filter by benchmark"
//	@Param			state		query		consts.DatapackState									false	"Filter by injection state"
//	@Param			status		query		int														false	"Filter by status"
//	@Param			labels		query		[]string												false	"Filter by labels (array of key:value strings, e.g., 'type:chaos')"
//	@Success		200			{object}	dto.GenericResponse[dto.ListResp[dto.InjectionResp]]	"Injections retrieved successfully"
//	@Failure		400			{object}	dto.GenericResponse[any]								"Invalid request format or parameters"
//	@Failure		401			{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403			{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500			{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/injections [get]
//	@x-api-type		{"sdk":"true"}
func ListInjections(c *gin.Context) {
	var req dto.ListInjectionReq
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ListInjections(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// SearchInjections
//
//	@Summary		Search injections
//	@Description	Advanced search for injections with complex filtering including name search, custom labels, tags, and time ranges
//	@Tags			Injections
//	@ID				search_injections
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			search	body		dto.SearchInjectionReq											true	"Search criteria"
//	@Success		200		{object}	dto.GenericResponse[dto.SearchResp[dto.InjectionDetailResp]]	"Search results"
//	@Failure		400		{object}	dto.GenericResponse[any]										"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]										"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]										"Permission denied"
//	@Failure		500		{object}	dto.GenericResponse[any]										"Internal server error"
//	@Router			/api/v2/injections/search [post]
//	@x-api-type		{"sdk":"true"}
func SearchInjections(c *gin.Context) {
	searchInjectionsCommon(c, nil)
}

// ListFaultInjectionNoIssues
//
//	@Summary		Query Fault Injection Records Without Issues
//	@Description	Query all fault injection records without issues based on time range, returning detailed records including configuration information
//	@Tags			Injections
//	@ID				list_failed_injections
//	@Produce		json
//	@Param			labels				query		[]string											false	"Filter by labels (array of key:value strings, e.g., 'type:chaos')"
//	@Param			lookback			query		string												false	"Time range query, supports custom relative time (1h/24h/7d) or custom, default not set"
//	@Param			custom_start_time	query		string												false	"Custom start time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Param			custom_end_time		query		string												false	"Custom end time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[[]dto.InjectionNoIssuesResp]	"Successfully returned fault injection records without issues"
//	@Failure		400					{object}	dto.GenericResponse[any]							"Request parameter error, such as incorrect time format or parameter validation failure, etc."
//	@Failure		500					{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/injections/analysis/no-issues [get]
//	@x-api-type		{"sdk":"true"}
func ListFaultInjectionNoIssues(c *gin.Context) {
	listFaultInjectionNoIssuesCommon(c, nil)
}

// ListFaultInjectionWithIssues
//
//	@Summary		Query Fault Injection Records With Issues
//	@Description	Query all fault injection records with issues based on time range
//	@Tags			Injections
//	@ID				list_successful_injections
//	@Produce		json
//	@Param			labels				query		[]string	false	"Filter by labels (array of key:value strings, e.g., 'type:chaos')"
//	@Param			lookback			query		string		false	"Time range query, supports custom relative time (1h/24h/7d) or custom, default not set"
//	@Param			custom_start_time	query		string		false	"Custom start time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Param			custom_end_time		query		string		false	"Custom end time, RFC3339 format, required when lookback=custom"	Format(date-time)
//	@Success		200					{object}	dto.GenericResponse[[]dto.InjectionWithIssuesResp]
//	@Failure		400					{object}	dto.GenericResponse[any]	"Request parameter error, such as incorrect time format or parameter validation failure, etc."
//	@Failure		500					{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/analysis/with-issues [get]
//	@x-api-type		{"sdk":"true"}
func ListFaultInjectionWithIssues(c *gin.Context) {
	listFaultInjectionWithIssuesCommon(c, nil)
}

// ManageInjectionCustomLabels manages injection custom labels (key-value pairs)
//
//	@Summary		Manage injection custom labels
//	@Description	Add or remove custom labels (key-value pairs) for an injection
//	@Tags			Injections
//	@ID				manage_injection_labels
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int										true	"Injection ID"
//	@Param			manage	body		dto.ManageInjectionLabelReq				true	"Custom label management request"
//	@Success		200		{object}	dto.GenericResponse[dto.InjectionResp]	"Custom labels managed successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]				"Invalid injection ID or request format/parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]				"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]				"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]				"Injection not found"
//	@Failure		500		{object}	dto.GenericResponse[any]				"Internal server error"
//	@Router			/api/v2/injections/{id}/labels [patch]
//	@x-api-type		{"sdk":"true"}
func ManageInjectionCustomLabels(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "injection ID")
	if !ok {
		return
	}

	var req dto.ManageInjectionLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	resp, err := producer.ManageInjectionLabels(&req, id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// BatchManageInjectionLabels
//
//	@Summary		Batch manage injection labels
//	@Description	Add or remove labels from multiple injections by IDs with success/failure tracking
//	@Tags			Injections
//	@ID				batch_manage_injection_labels
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			batch_manage	body		dto.BatchManageInjectionLabelReq						true	"Batch manage label request"
//	@Success		200				{object}	dto.GenericResponse[dto.BatchManageInjectionLabelResp]	"Injection labels managed successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]								"Invalid request"
//	@Failure		401				{object}	dto.GenericResponse[any]								"Authentication required"
//	@Failure		403				{object}	dto.GenericResponse[any]								"Permission denied"
//	@Failure		500				{object}	dto.GenericResponse[any]								"Internal server error"
//	@Router			/api/v2/injections/labels/batch [patch]
//	@x-api-type		{"sdk":"true"}
func BatchManageInjectionLabels(c *gin.Context) {
	var req dto.BatchManageInjectionLabelReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	resp, err := producer.BatchManageInjectionLabels(&req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// SubmitFaultInjection submits batch fault injections
//
//	@Summary		Submit batch fault injections
//	@Description	Submit multiple fault injection tasks in batch
//	@Tags			Injections
//	@ID				inject_fault
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		dto.SubmitInjectionReq							true	"Fault injection request body"
//	@Success		200		{object}	dto.GenericResponse[dto.SubmitInjectionResp]	"Fault injection submitted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]						"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]						"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]						"Resource not found"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/injections/inject [post]
//	@x-api-type		{"sdk":"true"}
func SubmitFaultInjection(c *gin.Context) {
	submitFaultInjectionCommon(c, nil)
}

// SubmitDatapackBuilding submits batch datapack buildings
//
//	@Summary		Submit batch datapack buildings
//	@Description.	Submit multiple datapack building tasks in batch
//	@Tags			Injections
//	@ID				build_datapack
//	@Accept			json
//	@Produce		json
//	@Param			body	body		dto.SubmitDatapackBuildingReq						true	"Datapack building request body"
//	@Success		202		{object}	dto.GenericResponse[dto.SubmitDatapackBuildingResp]	"Datapack building submitted successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]							"Invalid request format or parameters"
//	@Failure		401		{object}	dto.GenericResponse[any]							"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]							"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]							"Resource not found"
//	@Failure		500		{object}	dto.GenericResponse[any]							"Internal server error"
//	@Router			/api/v2/injections/build [post]
//	@x-api-type		{"sdk":"true"}
func SubmitDatapackBuilding(c *gin.Context) {
	submitDatapackBuildingCommon(c, nil)
}

// CloneInjection handles cloning an injection configuration
//
//	@Summary		Clone injection
//	@Description	Clone an existing injection configuration for reuse
//	@Tags			Injections
//	@ID				clone_injection
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int												true	"Injection ID"
//	@Param			body	body		dto.CloneInjectionReq							true	"Clone request"
//	@Success		201		{object}	dto.GenericResponse[dto.InjectionDetailResp]	"Injection cloned successfully"
//	@Failure		400		{object}	dto.GenericResponse[any]						"Invalid request"
//	@Failure		401		{object}	dto.GenericResponse[any]						"Authentication required"
//	@Failure		404		{object}	dto.GenericResponse[any]						"Injection not found"
//	@Failure		500		{object}	dto.GenericResponse[any]						"Internal server error"
//	@Router			/api/v2/injections/{id}/clone [post]
//	@x-api-type		{"sdk":"true"}
func CloneInjection(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "injection ID")
	if !ok {
		return
	}

	var req dto.CloneInjectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	resp, err := producer.CloneInjection(id, &req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "Injection cloned successfully", resp)
}

// GetInjectionLogs handles getting injection execution logs
//
//	@Summary		Get injection logs
//	@Description	Get execution logs for a specific injection
//	@Tags			Injections
//	@ID				get_injection_logs
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int											true	"Injection ID"
//	@Success		200	{object}	dto.GenericResponse[dto.InjectionLogsResp]	"Logs retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]					"Invalid injection ID"
//	@Failure		401	{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]					"Injection not found"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/injections/{id}/logs [get]
//	@x-api-type		{"sdk":"true"}
func GetInjectionLogs(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "injection ID")
	if !ok {
		return
	}

	resp, err := producer.GetInjectionLogs(id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusOK, "Logs retrieved successfully", resp)
}

// DownloadDatapack handles datapack file download
//
//	@Summary		Download datapack
//	@Description	Download datapack file by injection ID
//	@Tags			Injections
//	@ID				download_datapack
//	@Produce		application/zip
//	@Security		BearerAuth
//	@Param			id	path		int							true	"Injection ID"
//	@Success		200	{file}		binary						"Datapack zip file"
//	@Failure		400	{object}	dto.GenericResponse[any]	"Invalid injection ID"
//	@Failure		403	{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404	{object}	dto.GenericResponse[any]	"Injection not found"
//	@Failure		500	{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/{id}/download [get]
//	@x-api-type		{"sdk":"true"}
func DownloadDatapack(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "injection ID")
	if !ok {
		return
	}

	filename, err := producer.GetDatapackFilename(id)
	if handlers.HandleServiceError(c, err) {
		return
	}

	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", filename))

	zipWriter := zip.NewWriter(c.Writer)
	defer func() { _ = zipWriter.Close() }()

	if err := producer.DownloadDatapack(zipWriter, []utils.ExculdeRule{}, id); err != nil {
		delete(c.Writer.Header(), "Content-Disposition")
		c.Header("Content-Type", "application/json; charset=utf-8")
		handlers.HandleServiceError(c, err)
	}
}

// ListDatapackFiles handles getting the file structure of an injection datapack
//
//	@Summary		List datapack files
//	@Description	Get the file structure of an injection datapack
//	@Tags			Injections
//	@ID				list_datapack_files
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int											true	"Injection ID"
//	@Success		200	{object}	dto.GenericResponse[dto.DatapackFilesResp]	"Files retrieved successfully"
//	@Failure		400	{object}	dto.GenericResponse[any]					"Invalid injection ID"
//	@Failure		401	{object}	dto.GenericResponse[any]					"Authentication required"
//	@Failure		404	{object}	dto.GenericResponse[any]					"Datapack not found or not ready"
//	@Failure		500	{object}	dto.GenericResponse[any]					"Internal server error"
//	@Router			/api/v2/injections/{id}/files [get]
func ListDatapackFiles(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "datapack ID")
	if !ok {
		return
	}

	// Get base URL from request
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	resp, err := producer.GetDatapackFiles(id, baseURL)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// DownloadDatapackFile handles downloading a specific file from a datapack.
// Supports HTTP Range requests for resumable downloads.
//
//	@Summary		Download datapack file
//	@Description	Download a specific file from a datapack. Supports Range requests for resumable download.
//	@Tags			Injections
//	@ID				download_datapack_file
//	@Produce		application/octet-stream
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Injection ID"
//	@Param			path	query		string						true	"Relative path to the file"
//	@Success		200		{file}		binary						"Complete file content"
//	@Success		206		{file}		binary						"Partial file content (Range request)"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid injection ID or file path"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Datapack or file not found"
//	@Failure		416		{object}	dto.GenericResponse[any]	"Range not satisfiable"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/{id}/files/download [get]
func DownloadDatapackFile(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "datapack ID")
	if !ok {
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "File path is required")
		return
	}

	fileName, contentType, fileSize, fileReader, err := producer.DownloadDatapackFile(id, filePath)
	if handlers.HandleServiceError(c, err) {
		return
	}
	defer func() { _ = fileReader.Close() }()

	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", fileName))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("Accept-Ranges", "bytes")

	// Handle Range request for resumable download
	rangeHeader := c.GetHeader("Range")
	if rangeHeader != "" {
		serveRangeRequest(c, fileReader, fileSize, rangeHeader)
		return
	}

	// Full file response
	c.Header("Content-Length", strconv.FormatInt(fileSize, 10))
	c.Status(http.StatusOK)

	if _, err := io.Copy(c.Writer, fileReader); err != nil {
		logrus.WithError(err).Error("failed to stream file content")
		return
	}
}

// QueryDatapackFile handles querying the content of a specific file in the datapack.
// Returns the complete file with Content-Length for download progress tracking.
//
// NOTE: Arrow IPC is a structured stream that must be read sequentially from the
// beginning — Range requests are intentionally NOT supported here. Use
// DownloadDatapackFile for resumable downloads of raw files.
//
//	@Summary		Query datapack file content
//	@Description	Query the content of a parquet file in the datapack, returned as a complete stream. Content-Length header is provided for progress tracking.
//	@Tags			Injections
//	@ID				query_datapack_file
//	@Produce		application/vnd.apache.arrow.stream
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Injection ID"
//	@Param			path	query		string						true	"Relative path to the file"
//	@Success		200		{file}		binary						"Complete Arrow IPC stream"
//	@Failure		400		{object}	dto.GenericResponse[any]	"Invalid injection ID or file path"
//	@Failure		401		{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		403		{object}	dto.GenericResponse[any]	"Permission denied"
//	@Failure		404		{object}	dto.GenericResponse[any]	"Datapack or file not found"
//	@Failure		500		{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/{id}/files/query [get]
func QueryDatapackFile(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "datapack ID")
	if !ok {
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		dto.ErrorResponse(c, http.StatusBadRequest, "File path is required")
		return
	}

	ctx := c.Request.Context()

	fileName, totalRows, reader, err := producer.QueryDatapackFileContent(ctx, id, filePath)
	if err != nil {
		if handlers.HandleServiceError(c, err) {
			return
		}
	}
	defer func() { _ = reader.Close() }()

	// Content-Length enables axios onDownloadProgress to calculate percentage
	c.Header("Content-Type", "application/vnd.apache.arrow.stream")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.arrow", fileName))
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Header("X-Total-Rows", strconv.FormatInt(totalRows, 10))
	c.Header("X-Accel-Buffering", "no")
	c.Status(http.StatusOK)

	if _, err := io.Copy(c.Writer, reader); err != nil {
		logrus.Errorf("failed to stream file content: %v", err)
		return
	}
}

// ===================== Private Helper Functions =====================

// serveRangeRequest handles HTTP Range requests for partial content delivery.
// Supports single range requests in the format "bytes=start-end".
func serveRangeRequest(c *gin.Context, reader io.ReadSeeker, fileSize int64, rangeHeader string) {
	// Parse "bytes=start-end" format
	const prefix = "bytes="
	if !strings.HasPrefix(rangeHeader, prefix) {
		dto.ErrorResponse(c, http.StatusRequestedRangeNotSatisfiable, "Invalid range format")
		return
	}

	rangeSpec := strings.TrimPrefix(rangeHeader, prefix)
	// Only support single range (no multi-range)
	if strings.Contains(rangeSpec, ",") {
		dto.ErrorResponse(c, http.StatusRequestedRangeNotSatisfiable, "Multi-range not supported")
		return
	}

	parts := strings.SplitN(rangeSpec, "-", 2)
	if len(parts) != 2 {
		dto.ErrorResponse(c, http.StatusRequestedRangeNotSatisfiable, "Invalid range format")
		return
	}

	var start, end int64
	var err error

	if parts[0] == "" {
		// Suffix range: "bytes=-500" means last 500 bytes
		suffix, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil || suffix <= 0 || suffix > fileSize {
			c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
			dto.ErrorResponse(c, http.StatusRequestedRangeNotSatisfiable, "Invalid range")
			return
		}
		start = fileSize - suffix
		end = fileSize - 1
	} else {
		start, err = strconv.ParseInt(parts[0], 10, 64)
		if err != nil || start < 0 || start >= fileSize {
			c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
			dto.ErrorResponse(c, http.StatusRequestedRangeNotSatisfiable, "Invalid range start")
			return
		}

		if parts[1] == "" {
			// Open-ended range: "bytes=100-" means from 100 to end
			end = fileSize - 1
		} else {
			end, err = strconv.ParseInt(parts[1], 10, 64)
			if err != nil || end < start || end >= fileSize {
				c.Header("Content-Range", fmt.Sprintf("bytes */%d", fileSize))
				dto.ErrorResponse(c, http.StatusRequestedRangeNotSatisfiable, "Invalid range end")
				return
			}
		}
	}

	contentLength := end - start + 1

	// Seek to start position
	if _, err := reader.Seek(start, io.SeekStart); err != nil {
		logrus.Errorf("failed to seek to range start: %v", err)
		dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to seek to range start")
		return
	}

	c.Header("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, fileSize))
	c.Header("Content-Length", strconv.FormatInt(contentLength, 10))
	c.Status(http.StatusPartialContent)

	if _, err := io.CopyN(c.Writer, reader, contentLength); err != nil {
		logrus.Errorf("failed to stream partial content: %v", err)
		return
	}
}

// searchInjectionsCommon is the common logic for searching injections
func searchInjectionsCommon(c *gin.Context, projectID *int) {
	var req dto.SearchInjectionReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	// Note: Project filtering should be handled at the service layer
	// For project-scoped calls, the service layer will filter by project
	resp, err := producer.SearchInjections(&req, projectID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, resp)
}

// listFaultInjectionNoIssuesCommon is the common logic for listing injections without issues
func listFaultInjectionNoIssuesCommon(c *gin.Context, projectID *int) {
	var req dto.ListInjectionNoIssuesReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("failed to bind query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("invalid query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Note: Project filtering should be handled at the service layer
	// For project-scoped calls, the service layer will filter by project

	items, err := producer.ListInjectionsNoIssues(&req, projectID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, items)
}

// listFaultInjectionWithIssuesCommon is the common logic for listing injections with issues
func listFaultInjectionWithIssuesCommon(c *gin.Context, projectID *int) {
	var req dto.ListInjectionWithIssuesReq
	if err := c.BindQuery(&req); err != nil {
		logrus.Errorf("failed to bind query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid query parameters")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Errorf("invalid query parameters: %v", err)
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Note: Project filtering should be handled at the service layer
	// For project-scoped calls, the service layer will filter by project

	items, err := producer.ListInjectionsWithIssues(&req, projectID)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.SuccessResponse(c, items)
}

// submitFaultInjectionCommon is the common logic for submitting fault injections
func submitFaultInjectionCommon(c *gin.Context, projectID *int) {
	groupID := c.GetString("groupID")
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("Failed to get span context from gin.Context in SubmitFaultInjection")
		dto.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	spanCtx := ctx.(context.Context)
	span := trace.SpanFromContext(spanCtx)

	var req dto.SubmitInjectionReq
	if err := c.BindJSON(&req); err != nil {
		span.SetStatus(codes.Error, "validation error in SubmitFaultInjection: "+err.Error())
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		span.SetStatus(codes.Error, "validation error in SubmitFaultInjection: "+err.Error())
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	if req.ProjectName == "" && projectID == nil {
		span.SetStatus(codes.Error, "validation error in SubmitFaultInjection: project name is required")
		dto.ErrorResponse(c, http.StatusBadRequest, "Project name or ID is required")
		return
	}

	resp, err := producer.ProduceRestartPedestalTasks(spanCtx, &req, groupID, userID, projectID)
	if err != nil {
		span.SetStatus(codes.Error, "service error in SubmitFaultInjection: "+err.Error())
		logrus.Errorf("Failed to submit fault injection: %v", err)
		handlers.HandleServiceError(c, err)
		return
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("Successfully submitted %d fault injections with groupID: %s", len(resp.Items), groupID))
	dto.SuccessResponse(c, resp)
}

// submitDatapackBuildingCommon is the common logic for submitting datapack buildings
func submitDatapackBuildingCommon(c *gin.Context, projectID *int) {
	groupID := c.GetString("groupID")
	userID, exists := middleware.GetCurrentUserID(c)
	if !exists {
		dto.ErrorResponse(c, http.StatusUnauthorized, "Authentication required")
		return
	}

	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("Failed to get span context from gin.Context in SubmitDatapackBuilding")
		dto.ErrorResponse(c, http.StatusInternalServerError, "Internal server error")
		return
	}

	spanCtx := ctx.(context.Context)
	span := trace.SpanFromContext(spanCtx)

	var req dto.SubmitDatapackBuildingReq
	if err := c.BindJSON(&req); err != nil {
		span.SetStatus(codes.Error, "validation error in SubmitDatapackBuilding: "+err.Error())
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request format: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		span.SetStatus(codes.Error, "validation error in SubmitDatapackBuilding: "+err.Error())
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request parameters: "+err.Error())
		return
	}

	if req.ProjectName == "" && projectID == nil {
		span.SetStatus(codes.Error, "validation error in SubmitFaultInjection: project name is required")
		dto.ErrorResponse(c, http.StatusBadRequest, "Project name or ID is required")
		return
	}

	resp, err := producer.ProduceDatapackBuildingTasks(spanCtx, &req, groupID, userID, projectID)
	if err != nil {
		span.SetStatus(codes.Error, "service error in SubmitDatapackBuilding: "+err.Error())
		logrus.Errorf("Failed to submit datapack building: %v", err)
		handlers.HandleServiceError(c, err)
		return
	}

	span.SetStatus(codes.Ok, fmt.Sprintf("Successfully submitted %d datapack buildings with groupID: %s", len(resp.Items), groupID))
	dto.SuccessResponse(c, resp)
}

// UploadDatapack handles manual datapack upload
//
//	@Summary		Upload a manual datapack
//	@Description	Upload a zip archive as a manual datapack data source
//	@Tags			Injections
//	@ID				upload_datapack
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			name			formData	string						true	"Datapack name"
//	@Param			description		formData	string						false	"Description"
//	@Param			category		formData	string						false	"Category"
//	@Param			labels			formData	string						false	"JSON-encoded labels"
//	@Param			ground_truths	formData	string						false	"JSON-encoded ground truths"
//	@Param			file			formData	file						true	"Zip archive file"
//	@Success		201				{object}	dto.GenericResponse[dto.UploadDatapackResp]	"Datapack uploaded successfully"
//	@Failure		400				{object}	dto.GenericResponse[any]	"Invalid request"
//	@Failure		401				{object}	dto.GenericResponse[any]	"Authentication required"
//	@Failure		500				{object}	dto.GenericResponse[any]	"Internal server error"
//	@Router			/api/v2/injections/upload [post]
func UploadDatapack(c *gin.Context) {
	var req dto.UploadDatapackReq
	if err := c.ShouldBind(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "File is required: "+err.Error())
		return
	}
	defer func() { _ = file.Close() }()

	// Validate .zip extension
	if !strings.HasSuffix(strings.ToLower(header.Filename), ".zip") {
		dto.ErrorResponse(c, http.StatusBadRequest, "Only .zip files are accepted")
		return
	}

	// Max 2GB
	const maxSize = 2 << 30 // 2GB
	if header.Size > maxSize {
		dto.ErrorResponse(c, http.StatusBadRequest, "File size exceeds maximum allowed size of 2GB")
		return
	}

	resp, err := producer.UploadDatapack(&req, file, header.Size)
	if handlers.HandleServiceError(c, err) {
		return
	}

	dto.JSONResponse(c, http.StatusCreated, "Datapack uploaded successfully", resp)
}

// UpdateGroundtruth handles updating ground truth for a datapack
//
//	@Summary		Update datapack ground truth
//	@Description	Update or set ground truth labels for a datapack (fault injection)
//	@Tags			Injections
//	@ID				update_groundtruth
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"Injection ID"
//	@Param			request	body		dto.UpdateGroundtruthReq		true	"Ground truth data"
//	@Success		200		{object}	dto.GenericResponse[any]		"Ground truth updated"
//	@Failure		400		{object}	dto.GenericResponse[any]		"Invalid request"
//	@Failure		404		{object}	dto.GenericResponse[any]		"Injection not found"
//	@Router			/api/v2/injections/{id}/groundtruth [put]
func UpdateGroundtruth(c *gin.Context) {
	idStr := c.Param(consts.URLPathID)
	id, ok := handlers.ParsePositiveID(c, idStr, "injection ID")
	if !ok {
		return
	}

	var req dto.UpdateGroundtruthReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid request: "+err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Validation failed: "+err.Error())
		return
	}

	err := producer.UpdateGroundtruth(id, &req)
	if handlers.HandleServiceError(c, err) {
		return
	}

	logrus.WithField("id", id).Info("UpdateGroundtruth: successfully updated ground truth")
	dto.JSONResponse[any](c, http.StatusOK, "Ground truth updated", nil)
}
