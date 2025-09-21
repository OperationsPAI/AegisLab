package handlers

import (
	"archive/zip"
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"path/filepath"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	"aegis/executor"
	"aegis/middleware"
	"aegis/repository"
	"aegis/utils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// DeleteDataset
//
//	@Summary      Delete dataset data
//	@Description  Delete dataset data
//	@Tags         dataset
//	@Produce      application/json
//	@Param        names  query  []string  true  "Dataset name list"
//	@Success      200    {object}  dto.GenericResponse[dto.DatasetDeleteResp]
//	@Failure      400    {object}  dto.GenericResponse[any]
//	@Failure      500    {object}  dto.GenericResponse[any]
//	@Router       /api/v1/datasets [delete]
func DeleteDataset(c *gin.Context) {
	var req dto.DatasetDeleteReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	successCount, failedNames, err := repository.DeleteDatasetByName(req.Names)
	if err != nil {
		dto.ErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	dto.SuccessResponse(c, dto.DatasetDeleteResp{SuccessCount: successCount, FailedNames: failedNames})
}

// DownloadDataset handles dataset download requests
//
//	@Summary      Download dataset archive file
//	@Description  Package specified datasets into a ZIP file for download, automatically excluding result.csv and detector conclusion files. Supports downloading by group ID or dataset name (mutually exclusive). Directory structure: when downloading by group ID: datasets/{groupId}/{datasetName}/...; when by name: datasets/{datasetName}/...
//	@Tags         dataset
//	@Produce      application/zip
//	@Param        group_ids  query  []string  false  "List of task group IDs, format: group1,group2,group3. Mutually exclusive with names parameter; group_ids takes precedence"
//	@Param        names      query  []string  false  "List of dataset names, format: dataset1,dataset2,dataset3. Mutually exclusive with group_ids parameter"
//	@Success      200        {string}  binary  "ZIP file stream; the Content-Disposition header contains filename datasets.zip"
//	@Failure      400        {object}  dto.GenericResponse[any]  "Bad request parameters: 1) Parameter binding failed 2) Both parameters are empty 3) Both parameters provided"
//	@Failure      403        {object}  dto.GenericResponse[any]  "Permission error: requested dataset path is not within allowed scope"
//	@Failure      500        {object}  dto.GenericResponse[any]  "Internal server error"
//	@Router       /api/v1/datasets/download [get]
func DownloadDataset(c *gin.Context) {
	var req dto.DatasetDownloadReq
	if err := c.BindQuery(&req); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid Parameters")
		return
	}

	if err := req.Validate(); err != nil {
		dto.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Set the response headers
	c.Header("Content-Type", "application/zip")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s.zip", consts.DownloadFilename))

	zipWriter := zip.NewWriter(c.Writer)
	defer zipWriter.Close()

	excludeRules := []utils.ExculdeRule{
		{Pattern: consts.DetectorConclusionFile, IsGlob: false},
		{Pattern: consts.ExecutionResultFile, IsGlob: false},
	}

	// Define error handler function
	handleError := func(statusCode int, err error) {
		delete(c.Writer.Header(), "Content-Disposition")
		c.Header("Content-Type", "application/json; charset=utf-8")
		dto.ErrorResponse(c, statusCode, err.Error())
	}

	// Select download method based on input
	var (
		downloadFunc func(*zip.Writer, []string, []utils.ExculdeRule) (int, error)
		input        []string
	)

	switch {
	case len(req.GroupIDs) > 0:
		downloadFunc = downloadByGroupIds
		input = req.GroupIDs
	case len(req.Names) > 0:
		downloadFunc = downloadByNames
		input = req.Names
	}

	if statusCode, err := downloadFunc(zipWriter, input, excludeRules); err != nil {
		handleError(statusCode, err)
		return
	}
}

func downloadByGroupIds(zipWriter *zip.Writer, groupIDs []string, excludeRules []utils.ExculdeRule) (int, error) {
	joinedResults, err := repository.GetDatasetWithGroupIDs(groupIDs)
	if err != nil {
		message := "failed to query datasets"
		logrus.Errorf("%s: %v", message, err)
		return http.StatusInternalServerError, fmt.Errorf("failed to query datasets")
	}

	groupedResults := make(map[string][]string)
	for _, joinedResult := range joinedResults {
		if _, exists := groupedResults[joinedResult.GroupID]; !exists {
			groupedResults[joinedResult.GroupID] = []string{}
		}
		groupedResults[joinedResult.GroupID] = append(groupedResults[joinedResult.GroupID], joinedResult.Name)
	}

	// Pre-validate all dataset paths
	for _, datasets := range groupedResults {
		for _, dataset := range datasets {
			workDir := filepath.Join(config.GetString("nfs.path"), dataset)
			if !utils.IsAllowedPath(workDir) {
				logrus.WithField("path", workDir).Errorf("invalid path access")
				return http.StatusInternalServerError, fmt.Errorf("invalid path access")
			}
		}
	}

	for groupID, datasets := range groupedResults {
		folderName := filepath.Join(consts.DownloadFilename, groupID)
		for _, dataset := range datasets {
			workDir := filepath.Join(config.GetString("nfs.path"), dataset)

			err := filepath.WalkDir(workDir, func(path string, dir fs.DirEntry, err error) error {
				if err != nil || dir.IsDir() {
					return err
				}

				relPath, _ := filepath.Rel(workDir, path)
				fullRelPath := filepath.Join(folderName, filepath.Base(workDir), relPath)
				fileName := filepath.Base(path)

				// Apply exclusion rules
				for _, rule := range excludeRules {
					if utils.MatchFile(fileName, rule) {
						return nil
					}
				}

				// Get file info to read modification time
				fileInfo, err := dir.Info()
				if err != nil {
					return err
				}

				// Convert path separators to "/"
				zipPath := filepath.ToSlash(fullRelPath)
				return utils.AddToZip(zipWriter, fileInfo, path, zipPath)
			})
			if err != nil {
				logrus.Errorf("failed to package: %v", err)
				return http.StatusForbidden, fmt.Errorf("failed to package")
			}
		}
	}

	return http.StatusOK, nil
}

func downloadByNames(zipWriter *zip.Writer, names []string, excludeRules []utils.ExculdeRule) (int, error) {
	for _, name := range names {
		workDir := filepath.Join(config.GetString("nfs.path"), name)
		if !utils.IsAllowedPath(workDir) {
			logrus.WithField("path", workDir).Errorf("invalid path access")
			return http.StatusInternalServerError, fmt.Errorf("invalid path access")
		}
	}

	folderName := consts.DownloadFilename
	for _, name := range names {
		workDir := filepath.Join(config.GetString("nfs.path"), name)

		err := filepath.WalkDir(workDir, func(path string, dir fs.DirEntry, err error) error {
			if err != nil || dir.IsDir() {
				return err
			}

			relPath, _ := filepath.Rel(workDir, path)
			fullRelPath := filepath.Join(folderName, filepath.Base(workDir), relPath)
			fileName := filepath.Base(path)

			// Apply exclusion rules
			for _, rule := range excludeRules {
				if utils.MatchFile(fileName, rule) {
					return nil
				}
			}

			// Get file info to read modification time
			fileInfo, err := dir.Info()
			if err != nil {
				return err
			}

			// Convert path separators to "/"
			zipPath := filepath.ToSlash(fullRelPath)
			return utils.AddToZip(zipWriter, fileInfo, path, zipPath)
		})
		if err != nil {
			logrus.Errorf("failed to package: %v", err)
			return http.StatusForbidden, fmt.Errorf("failed to package")
		}
	}

	return http.StatusOK, nil
}

// SubmitDatasetBuilding
//
//	@Summary      Batch build datasets
//	@Description  Batch build datasets based on specified time range and benchmark container
//	@Tags         dataset
//	@Accept       json
//	@Produce      json
//	@Param        body  body  dto.SubmitDatasetBuildingReq  true  "List of dataset build requests; each request includes dataset name, time range, benchmark, and environment variable configuration"
//	@Success      202   {object}  dto.GenericResponse[dto.SubmitResp]  "Successfully submitted dataset building tasks; returns group ID and trace information list"
//	@Failure      400   {object}  dto.GenericResponse[any]  "Bad request parameters: 1) Invalid JSON format 2) Empty dataset name 3) Invalid time range 4) Benchmark does not exist 5) Unsupported environment variable name"
//	@Failure      500   {object}  dto.GenericResponse[any]  "Internal server error"
//	@Router       /api/v1/datasets [post]
func SubmitDatasetBuilding(c *gin.Context) {
	groupID := c.GetString("groupID")
	logrus.Infof("SubmitDatasetBuilding, groupID: %s", groupID)

	var req dto.SubmitDatasetBuildingReq
	if err := c.BindJSON(&req); err != nil {
		logrus.Error(err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if err := req.Validate(); err != nil {
		logrus.Error(err)
		dto.ErrorResponse(c, http.StatusBadRequest, "Invalid dataset building payload")
		return
	}

	// Optimize span output
	ctx, ok := c.Get(middleware.SpanContextKey)
	if !ok {
		logrus.Error("failed to get span context from gin.Context")
		dto.ErrorResponse(c, http.StatusInternalServerError, "failed to get span context")
		return
	}

	spanCtx := ctx.(context.Context)
	traces := make([]dto.Trace, 0, len(req.Payloads))
	for idx, payload := range req.Payloads {
		task := &dto.UnifiedTask{
			Type:      consts.TaskTypeBuildDataset,
			Payload:   utils.StructToMap(payload),
			Immediate: true,
			GroupID:   groupID,
		}
		task.SetGroupCtx(spanCtx)

		taskID, traceID, err := executor.SubmitTask(spanCtx, task)
		if err != nil {
			logrus.Errorf("failed to submit dataset building task: %v", err)
			dto.ErrorResponse(c, http.StatusInternalServerError, "Failed to submit dataset building task")
			return
		}

		traces = append(traces, dto.Trace{TraceID: traceID, HeadTaskID: taskID, Index: idx})
	}

	dto.JSONResponse(c, http.StatusAccepted, "Dataset building submitted successfully",
		dto.SubmitResp{GroupID: groupID, Traces: traces},
	)
}
