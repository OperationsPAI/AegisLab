package producer

import (
	"aegis/client"
	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"aegis/utils"
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	chaos "github.com/OperationsPAI/chaos-experiment/handler"
	"github.com/apache/arrow-go/v18/arrow/ipc"
	"github.com/duckdb/duckdb-go/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// injectionProcessItem represents a batch of parallel fault injections
type injectionProcessItem struct {
	index         int          // Batch index in the original request
	faultDuration int          // Maximum duration among all faults in this batch
	nodes         []chaos.Node // Multiple fault nodes to be injected in parallel
	executeTime   time.Time    // Execution time for this batch
}

// BatchDeleteInjectionsByIDs deletes fault injections based on their IDs
func BatchDeleteInjectionsByIDs(injectionIDs []int) error {
	if len(injectionIDs) == 0 {
		return nil
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		return batchDeleteExecutionsCore(tx, injectionIDs)
	})
}

// BatchDeleteInjectionsByLabels deletes fault injections based on label conditions
func BatchDeleteInjectionsByLabels(labelItems []dto.LabelItem) error {
	if len(labelItems) == 0 {
		return nil
	}

	labelConditions := make([]map[string]string, 0, len(labelItems))
	for _, item := range labelItems {
		labelConditions = append(labelConditions, map[string]string{
			"key":   item.Key,
			"value": item.Value,
		})
	}

	return database.DB.Transaction(func(tx *gorm.DB) error {
		injectionIDs, err := repository.ListInjectionIDsByLabels(tx, labelConditions)
		if err != nil {
			return fmt.Errorf("failed to list injection ids by labels: %w", err)
		}

		return batchDeleteInjectionsCore(tx, injectionIDs)
	})
}

// CreateInjection creates a new fault injection along with its associated project-container relationships and labels
func CreateInjection(injection *database.FaultInjection, labelItems []dto.LabelItem) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := repository.CreateInjection(tx, injection); err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return fmt.Errorf("%w: injection with name %s already exists", consts.ErrAlreadyExists, injection.Name)
			}
			return fmt.Errorf("failed to create injection: %w", err)
		}

		if len(labelItems) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, labelItems, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			// Collect label IDs
			labelIDs := make([]int, 0, len(labels))
			for _, label := range labels {
				labelIDs = append(labelIDs, label.ID)
			}

			// AddInjectionLabels now takes injectionID and labelIDs (stores as TaskLabel internally)
			if err := repository.AddInjectionLabels(tx, injection.ID, labelIDs); err != nil {
				return fmt.Errorf("failed to add injection labels: %w", err)
			}
		}

		return nil
	})
}

// GetInjectionDetail retrieves detailed information about a specific fault injection
func GetInjectionDetail(injectionID int) (*dto.InjectionDetailResp, error) {
	logEntry := logrus.WithFields(logrus.Fields{
		"injectionID": injectionID,
	})

	injection, err := repository.GetInjectionByID(database.DB, injectionID)
	if err != nil {
		logEntry.Error("failed to get injection from repository: %w", err)
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, injectionID)
		}
		return nil, fmt.Errorf("failed to get injection: %w", err)
	}

	labels, err := repository.ListLabelsByInjectionID(database.DB, injection.ID)
	if err != nil {
		logEntry.Error("failed to get injection labels from repository: %w", err)
		return nil, fmt.Errorf("failed to get injection labels: %w", err)
	}

	injection.Labels = labels
	resp := dto.NewInjectionDetailResp(injection)

	return resp, err
}

// CloneInjection clones an existing injection with a new name
func CloneInjection(injectionID int, req *dto.CloneInjectionReq) (*dto.InjectionDetailResp, error) {
	original, err := repository.GetInjectionByID(database.DB, injectionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, injectionID)
		}
		return nil, fmt.Errorf("failed to get injection: %w", err)
	}

	cloned := &database.FaultInjection{
		Name:          req.Name,
		FaultType:     original.FaultType,
		Category:      original.Category,
		Description:   original.Description,
		DisplayConfig: original.DisplayConfig,
		EngineConfig:  original.EngineConfig,
		Groundtruths:  original.Groundtruths,
		PreDuration:   original.PreDuration,
		StartTime:     original.StartTime,
		EndTime:       original.EndTime,
		BenchmarkID:   original.BenchmarkID,
		PedestalID:    original.PedestalID,
		State:         consts.DatapackInitial,
		Status:        consts.CommonEnabled,
	}

	if err := CreateInjection(cloned, req.Labels); err != nil {
		return nil, err
	}

	labels, err := repository.ListLabelsByInjectionID(database.DB, cloned.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cloned injection labels: %w", err)
	}

	cloned.Labels = labels
	return dto.NewInjectionDetailResp(cloned), nil
}

// GetInjectionLogs retrieves execution logs for an injection
func GetInjectionLogs(injectionID int) (*dto.InjectionLogsResp, error) {
	injection, err := repository.GetInjectionByID(database.DB, injectionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, injectionID)
		}
		return nil, fmt.Errorf("failed to get injection: %w", err)
	}

	resp := &dto.InjectionLogsResp{
		InjectionID: injectionID,
		Logs:        []string{},
	}

	if injection.TaskID != nil {
		resp.TaskID = *injection.TaskID
		// TODO: Implement actual log retrieval from task execution
		// For now, return empty logs as placeholder
	}

	return resp, nil
}

// ListInjections lists fault injections based on the provided filters
func ListInjections(req *dto.ListInjectionReq) (*dto.ListResp[dto.InjectionResp], error) {
	limit, offset := req.ToGormParams()
	fitlerOptions := req.ToFilterOptions()

	injections, total, err := repository.ListInjections(database.DB, limit, offset, fitlerOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to list injections: %w", err)
	}

	injectionIDs := make([]int, 0, len(injections))
	for _, injection := range injections {
		injectionIDs = append(injectionIDs, injection.ID)
	}

	labelsMap, err := repository.ListInjectionLabels(database.DB, injectionIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to list injection labels: %w", err)
	}

	injectionResps := make([]dto.InjectionResp, 0, len(injections))
	for _, injection := range injections {
		if labels, exists := labelsMap[injection.ID]; exists {
			injection.Labels = labels
		}
		injectionResps = append(injectionResps, *dto.NewInjectionResp(&injection))
	}

	resp := dto.ListResp[dto.InjectionResp]{
		Items:      injectionResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}
	return &resp, nil
}

// SearchInjections performs advanced search on fault injections
func SearchInjections(req *dto.SearchInjectionReq, projectID *int) (*dto.SearchResp[dto.InjectionDetailResp], error) {
	if req == nil {
		return nil, fmt.Errorf("search injection request is nil")
	}

	searchReq := req.ConvertToSearchReq()

	// Add project filter if projectID is provided
	if projectID != nil {
		searchReq.AddFilter("project_id", dto.OpEqual, *projectID)
	}

	injections, total, err := repository.ExecuteSearch(database.DB, searchReq, database.FaultInjection{})
	if err != nil {
		return nil, fmt.Errorf("failed to search injections: %w", err)
	}

	labelConditions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		labelConditions = append(labelConditions, map[string]string{
			"key":   item.Key,
			"value": item.Value,
		})
	}

	filteredInjections := []database.FaultInjection{}
	if len(labelConditions) > 0 {
		injectionIDs, err := repository.ListInjectionIDsByLabels(database.DB, labelConditions)
		if err != nil {
			return nil, fmt.Errorf("failed to list injection ids by labels: %w", err)
		}

		injectionIDMap := make(map[int]struct{}, len(injectionIDs))
		for _, id := range injectionIDs {
			injectionIDMap[id] = struct{}{}
		}

		for _, injection := range injections {
			if _, exists := injectionIDMap[injection.ID]; exists {
				filteredInjections = append(filteredInjections, injection)
			}
		}
	} else {
		filteredInjections = injections
	}

	// Convert to response format
	injectionResps := make([]dto.InjectionDetailResp, 0, len(filteredInjections))
	for _, injection := range filteredInjections {
		injectionResps = append(injectionResps, *dto.NewInjectionDetailResp(&injection))
	}

	resp := &dto.SearchResp[dto.InjectionDetailResp]{
		Items:      injectionResps,
		Pagination: req.ConvertToPaginationInfo(total),
	}

	return resp, nil
}

// GetDatapackFilename returns the filename for downloading a datapack
func GetDatapackFilename(injectionID int) (string, error) {
	injection, err := repository.GetInjectionByID(database.DB, injectionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return "", fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, injectionID)
		}
		return "", fmt.Errorf("failed to get injection: %w", err)
	}

	if injection.State < consts.DatapackBuildSuccess {
		return "", fmt.Errorf("datapack for injection id %d is not ready for download", injectionID)
	}

	return injection.Name, nil
}

// DownloadDatapack handles the downloading of a specific datapack
func DownloadDatapack(zipWriter *zip.Writer, excludeRules []utils.ExculdeRule, injectionID int) error {
	if zipWriter == nil {
		return fmt.Errorf("zip writer cannot be nil")
	}

	injection, err := repository.GetInjectionByID(database.DB, injectionID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, injectionID)
		}
		return fmt.Errorf("failed to get injection: %w", err)
	}

	if err := packageDatapackToZip(zipWriter, injection, excludeRules); err != nil {
		return fmt.Errorf("failed to package injection to zip: %w", err)
	}

	return nil
}

// GetDatapackFiles retrieves the file structure of a datapack in tree format
func GetDatapackFiles(datapackID int, baseURL string) (*dto.DatapackFilesResp, error) {
	datapack, err := repository.GetInjectionByID(database.DB, datapackID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return nil, fmt.Errorf("%w: datapack id: %d", consts.ErrNotFound, datapackID)
		}
		return nil, fmt.Errorf("failed to get datapack: %w", err)
	}

	if datapack.State < consts.DatapackBuildSuccess {
		return nil, fmt.Errorf("datapack %d is not ready", datapackID)
	}

	workDir := filepath.Join(config.GetString("jfs.dataset_path"), datapack.Name)
	if !utils.IsAllowedPath(workDir) {
		return nil, fmt.Errorf("invalid path access to %s", workDir)
	}

	// Check if directory exists
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("datapack directory not found for datapack id %d", datapackID)
	}

	resp := &dto.DatapackFilesResp{
		Files:     []dto.DatapackFileItem{},
		FileCount: 0,
		DirCount:  0,
	}

	// Build tree structure
	rootItems, err := buildFileTree(workDir, "", baseURL, datapackID, resp)
	if err != nil {
		return nil, fmt.Errorf("failed to build file tree: %w", err)
	}

	resp.Files = rootItems
	return resp, nil
}

// DownloadDatapackFile downloads a specific file from a datapack
func DownloadDatapackFile(datapackID int, filePath string) (string, string, int64, io.ReadSeekCloser, error) {
	fullPath, err := getFileFullPath(datapackID, filePath)
	if err != nil {
		return "", "", 0, nil, fmt.Errorf("invalid file path: %w", err)
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return "", "", 0, nil, fmt.Errorf("failed to open file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		file.Close()
		return "", "", 0, nil, fmt.Errorf("failed to stat file: %w", err)
	}

	fileName := filepath.Base(fullPath)
	contentType := "application/octet-stream"

	// Determine content type based on file extension
	switch filepath.Ext(fileName) {
	case ".json":
		contentType = "application/json"
	case ".yaml", ".yml":
		contentType = "application/x-yaml"
	case ".txt", ".log":
		contentType = "text/plain"
	case ".csv":
		contentType = "text/csv"
	case ".xml":
		contentType = "application/xml"
	case ".html", ".htm":
		contentType = "text/html"
	case ".pdf":
		contentType = "application/pdf"
	case ".zip":
		contentType = "application/zip"
	case ".tar", ".gz", ".tgz":
		contentType = "application/x-tar"
	}

	return fileName, contentType, stat.Size(), file, nil
}

// QueryDatapackFileContent reads a parquet file and returns it along with file size
func QueryDatapackFileContent(ctx context.Context, datapackID int, filePath string) (string, int64, io.ReadCloser, error) {
	fullPath, err := getFileFullPath(datapackID, filePath)
	if err != nil {
		return "", 0, nil, fmt.Errorf("invalid file path: %w", err)
	}
	if filepath.Ext(fullPath) != ".parquet" {
		return "", 0, nil, fmt.Errorf("file is not a parquet file: %s", filePath)
	}

	connector, err := duckdb.NewConnector("", nil)
	if err != nil {
		return "", 0, nil, err
	}

	countConn, err := connector.Connect(ctx)
	if err != nil {
		logrus.Errorf("connect failed: %v", err)
		return "", 0, nil, err
	}
	defer countConn.Close()

	var totalRows int64
	countQuery := fmt.Sprintf("SELECT count(*) FROM read_parquet('%s')", fullPath)

	db := sql.OpenDB(connector)
	if err := db.QueryRowContext(ctx, countQuery).Scan(&totalRows); err != nil {
		return "", 0, nil, err
	}

	// Inspect schema and build a SELECT that casts unsupported unsigned integer types
	safeSQL, err := buildSafeParquetSQL(ctx, db, fullPath)
	if err != nil {
		return "", 0, nil, fmt.Errorf("failed to build safe parquet SQL: %w", err)
	}

	pr, pw := io.Pipe()
	go func() {
		defer pw.Close()

		conn, err := connector.Connect(ctx)
		if err != nil {
			logrus.Errorf("connect failed: %v", err)
			return
		}
		defer conn.Close()

		arrow, err := duckdb.NewArrowFromConn(conn)
		if err != nil {
			logrus.Errorf("failed to get arrow interface: %v", err)
			return
		}

		rdr, err := arrow.QueryContext(ctx, safeSQL)
		if err != nil {
			logrus.Errorf("query context failed: %v", err)
			return
		}
		defer rdr.Release()

		writer := ipc.NewWriter(pw, ipc.WithSchema(rdr.Schema()), ipc.WithZstd())
		defer writer.Close()

		for rdr.Next() {
			record := rdr.RecordBatch()
			if err := writer.Write(record); err != nil {
				logrus.Errorf("failed to write arrow record: %v", err)
				record.Release()
				break
			}
			record.Release()
		}

		if err := rdr.Err(); err != nil {
			logrus.Errorf("reader error: %v", err)
		}
	}()

	return filepath.Base(fullPath), totalRows, pr, nil
}

// ListInjectionsNoissues handles the request to list fault injections without issues
func ListInjectionsNoIssues(req *dto.ListInjectionNoIssuesReq, projectID *int) ([]dto.InjectionNoIssuesResp, error) {
	if len(req.Labels) == 0 {
		return nil, nil
	}

	labelConditions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		parts := strings.SplitN(item, ":", 2)
		labelConditions = append(labelConditions, map[string]string{
			"key":   parts[0],
			"value": parts[1],
		})
	}

	opts, err := req.TimeRangeQuery.Convert()
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	records, err := repository.ListInjectionsNoIssues(database.DB, labelConditions, &opts.CustomStartTime, &opts.CustomEndTime, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list fault injections without issues: %w", err)
	}

	var items []dto.InjectionNoIssuesResp
	for i, record := range records {
		resp, err := dto.NewInjectionNoIssuesResp(record)
		if err != nil {
			return nil, fmt.Errorf("failed to create InjectionNoIssuesResp at index %d: %w", i, err)
		}

		items = append(items, *resp)
	}

	return items, nil
}

// ListInjectionsNoissues handles the request to list fault injections without issues
func ListInjectionsWithIssues(req *dto.ListInjectionWithIssuesReq, projectID *int) ([]dto.InjectionWithIssuesResp, error) {
	if len(req.Labels) == 0 {
		return nil, nil
	}

	labelConditions := make([]map[string]string, 0, len(req.Labels))
	for _, item := range req.Labels {
		parts := strings.SplitN(item, ":", 2)
		labelConditions = append(labelConditions, map[string]string{
			"key":   parts[0],
			"value": parts[1],
		})
	}

	opts, err := req.TimeRangeQuery.Convert()
	if err != nil {
		return nil, fmt.Errorf("invalid time range: %w", err)
	}

	records, err := repository.ListInjectionsWithIssues(database.DB, labelConditions, &opts.CustomStartTime, &opts.CustomEndTime, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to list fault injections without issues: %w", err)
	}

	var items []dto.InjectionWithIssuesResp
	for _, record := range records {
		resp, err := dto.NewInjectionWithIssuesResp(record)
		if err != nil {
			return nil, fmt.Errorf("failed to create InjectionNoIssuesResp: %w", err)
		}

		items = append(items, *resp)
	}

	return items, nil
}

// ManageInjectionTags manages labels associated with a fault injection
func ManageInjectionLabels(req *dto.ManageInjectionLabelReq, injectionID int) (*dto.InjectionResp, error) {
	if req == nil {
		return nil, fmt.Errorf("manage injection labels request is nil")
	}

	var managedInjection *database.FaultInjection

	err := database.DB.Transaction(func(tx *gorm.DB) error {
		injection, err := repository.GetInjectionByID(database.DB, injectionID)
		if err != nil {
			if errors.Is(err, consts.ErrNotFound) {
				return fmt.Errorf("%w: injection id: %d", consts.ErrNotFound, injectionID)
			}
			return fmt.Errorf("failed to get injection: %w", err)
		}

		if len(req.AddLabels) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, req.AddLabels, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			// Collect label IDs
			labelIDs := make([]int, 0, len(labels))
			for _, label := range labels {
				labelIDs = append(labelIDs, label.ID)
			}

			// AddInjectionLabels now takes injectionID and labelIDs (stores as TaskLabel internally)
			if err := repository.AddInjectionLabels(tx, injection.ID, labelIDs); err != nil {
				return fmt.Errorf("failed to add injection labels: %w", err)
			}
		}

		if len(req.RemoveLabels) > 0 {
			labelIDs, err := repository.ListLabelIDsByKeyAndInjectionID(tx, injection.ID, req.RemoveLabels)
			if err != nil {
				return fmt.Errorf("failed to find label ids by keys: %w", err)
			}

			if len(labelIDs) > 0 {
				if err := repository.ClearInjectionLabels(tx, []int{injectionID}, labelIDs); err != nil {
					return fmt.Errorf("failed to clear injection labels: %w", err)
				}

				if err := repository.BatchDecreaseLabelUsages(tx, labelIDs, 1); err != nil {
					return fmt.Errorf("failed to decrease label usage counts: %w", err)
				}
			}
		}

		labels, err := repository.ListLabelsByInjectionID(database.DB, injectionID)
		if err != nil {
			return fmt.Errorf("failed to get injection labels: %w", err)
		}

		injection.Labels = labels
		managedInjection = injection
		return nil
	})
	if err != nil {
		return nil, err
	}

	return dto.NewInjectionResp(managedInjection), nil
}

// BatchManageInjectionLabels adds or removes labels from multiple injections
// Each injection can have its own set of label operations
func BatchManageInjectionLabels(req *dto.BatchManageInjectionLabelReq) (*dto.BatchManageInjectionLabelResp, error) {
	if req == nil {
		return nil, fmt.Errorf("batch manage injection labels request is nil")
	}

	resp := &dto.BatchManageInjectionLabelResp{
		FailedCount:  0,
		FailedItems:  []string{},
		SuccessCount: 0,
		SuccessItems: []dto.InjectionResp{},
	}

	if len(req.Items) == 0 {
		return resp, nil
	}

	// Process all operations in a single transaction
	return resp, database.DB.Transaction(func(tx *gorm.DB) error {
		// Step 1: Collect all injection IDs and verify they exist (batch query)
		allInjectionIDs := make([]int, 0, len(req.Items))
		operationMap := make(map[int]*dto.InjectionLabelOperation)

		for i := range req.Items {
			item := &req.Items[i]
			allInjectionIDs = append(allInjectionIDs, item.InjectionID)
			operationMap[item.InjectionID] = item
		}

		injections, err := repository.ListFaultInjectionsByID(tx, allInjectionIDs)
		if err != nil {
			return fmt.Errorf("failed to list injections: %w", err)
		}

		foundIDMap := make(map[int]*database.FaultInjection)
		for i := range injections {
			foundIDMap[injections[i].ID] = &injections[i]
		}

		// Track which IDs were not found
		validIDs := make([]int, 0, len(foundIDMap))
		for _, id := range allInjectionIDs {
			if _, found := foundIDMap[id]; !found {
				resp.FailedItems = append(resp.FailedItems, fmt.Sprintf("Injection ID %d not found", id))
				resp.FailedCount++
				delete(operationMap, id) // Remove from operations
			} else {
				validIDs = append(validIDs, id)
			}
		}

		if len(validIDs) == 0 {
			return fmt.Errorf("no valid injection IDs found")
		}

		// Step 2: Collect all unique labels from all operations and create them in batch
		allAddLabels := make([]dto.LabelItem, 0)
		allRemoveLabels := make([]dto.LabelItem, 0)
		labelKeySet := make(map[string]bool)

		for _, op := range operationMap {
			for _, label := range op.AddLabels {
				key := label.Key + ":" + label.Value
				if !labelKeySet[key] {
					labelKeySet[key] = true
					allAddLabels = append(allAddLabels, label)
				}
			}
			for _, label := range op.RemoveLabels {
				key := label.Key + ":" + label.Value
				if !labelKeySet[key] {
					labelKeySet[key] = true
					allRemoveLabels = append(allRemoveLabels, label)
				}
			}
		}

		// Create or update all labels in batch
		var labelMap map[string]int // key:value -> label_id
		if len(allAddLabels) > 0 {
			labels, err := common.CreateOrUpdateLabelsFromItems(tx, allAddLabels, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to create or update labels: %w", err)
			}

			labelMap = make(map[string]int)
			for _, label := range labels {
				key := label.Key + ":" + label.Value
				labelMap[key] = label.ID
			}
		}

		// Get label IDs for removal labels
		var removeLabelMap map[string]int // key:value -> label_id
		if len(allRemoveLabels) > 0 {
			labelConditions := make([]map[string]string, 0, len(allRemoveLabels))
			for _, item := range allRemoveLabels {
				labelConditions = append(labelConditions, map[string]string{
					"key":   item.Key,
					"value": item.Value,
				})
			}

			labelIDs, err := repository.ListLabelIDsByConditions(tx, labelConditions, consts.InjectionCategory)
			if err != nil {
				return fmt.Errorf("failed to find labels to remove: %w", err)
			}

			// Map them back for quick lookup
			if len(labelIDs) > 0 {
				labels, err := repository.ListLabelsByID(tx, labelIDs)
				if err != nil {
					return fmt.Errorf("failed to list labels by IDs: %w", err)
				}

				removeLabelMap = make(map[string]int)
				for _, label := range labels {
					key := label.Key + ":" + label.Value
					removeLabelMap[key] = label.ID
				}
			}
		}

		// Step 3: Process each injection's operations
		for _, injectionID := range validIDs {
			op := operationMap[injectionID]

			if len(op.AddLabels) > 0 {
				labelIDsToAdd := make([]int, 0, len(op.AddLabels))
				for _, label := range op.AddLabels {
					key := label.Key + ":" + label.Value
					if labelID, exists := labelMap[key]; exists {
						labelIDsToAdd = append(labelIDsToAdd, labelID)
					}
				}

				if len(labelIDsToAdd) > 0 {
					if err := repository.AddInjectionLabels(tx, injectionID, labelIDsToAdd); err != nil {
						resp.FailedItems = append(resp.FailedItems, fmt.Sprintf("Injection ID %d: failed to add labels - %s", injectionID, err.Error()))
						resp.FailedCount++
						delete(foundIDMap, injectionID)
						continue
					}
				}
			}

			if len(op.RemoveLabels) > 0 && removeLabelMap != nil {
				labelIDsToRemove := make([]int, 0, len(op.RemoveLabels))
				for _, label := range op.RemoveLabels {
					key := label.Key + ":" + label.Value
					if labelID, exists := removeLabelMap[key]; exists {
						labelIDsToRemove = append(labelIDsToRemove, labelID)
					}
				}

				if len(labelIDsToRemove) > 0 {
					if err := repository.ClearInjectionLabels(tx, []int{injectionID}, labelIDsToRemove); err != nil {
						resp.FailedItems = append(resp.FailedItems, fmt.Sprintf("Injection ID %d: failed to remove labels - %s", injectionID, err.Error()))
						resp.FailedCount++
						delete(foundIDMap, injectionID)
						continue
					}
				}
			}
		}

		// Step 4: Fetch updated injection data with labels (batch query)
		if len(foundIDMap) > 0 {
			successIDs := make([]int, 0, len(foundIDMap))
			for id := range foundIDMap {
				successIDs = append(successIDs, id)
			}

			updatedInjections, err := repository.ListFaultInjectionsByID(tx, successIDs)
			if err != nil {
				return fmt.Errorf("failed to fetch updated injections: %w", err)
			}

			labelsMap, err := repository.ListInjectionLabels(tx, successIDs)
			if err != nil {
				return fmt.Errorf("failed to list injection labels: %w", err)
			}

			for i := range updatedInjections {
				injection := &updatedInjections[i]
				if labels, exists := labelsMap[injection.ID]; exists {
					injection.Labels = labels
				}
				injectionResp := dto.NewInjectionResp(injection)
				resp.SuccessItems = append(resp.SuccessItems, *injectionResp)
				resp.SuccessCount++
			}
		}

		return nil
	})
}

// ProduceRestartPedestalTasks produces pedestal restart tasks with support for parallel fault injection
func ProduceRestartPedestalTasks(ctx context.Context, req *dto.SubmitInjectionReq, groupID string, userID int, projectID *int) (*dto.SubmitInjectionResp, error) {
	if req == nil {
		return nil, fmt.Errorf("submit injection request is nil")
	}

	if projectID == nil {
		project, err := repository.GetProjectByName(database.DB, req.ProjectName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("%w: project %s not found", consts.ErrNotFound, req.ProjectName)
			}
			return nil, fmt.Errorf("failed to get project: %w", err)
		}
		projectID = &project.ID
	}

	pedestalVersionResults, err := common.MapRefsToContainerVersions([]*dto.ContainerRef{&req.Pedestal.ContainerRef}, consts.ContainerTypePedestal, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map pedestal container ref to version: %w", err)
	}

	pedestalVersion, exists := pedestalVersionResults[&req.Pedestal.ContainerRef]
	if !exists {
		return nil, fmt.Errorf("pedestal version not found for container: %s (version: %s)", req.Pedestal.Name, req.Pedestal.Version)
	}

	helmConfig, err := repository.GetHelmConfigByContainerVersionID(database.DB, pedestalVersion.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("%w: helm config not found for pedestal version id %d", consts.ErrNotFound, pedestalVersion.ID)
		}
		return nil, fmt.Errorf("failed to get helm config: %w", err)
	}

	params := flattenYAMLToParameters(req.Pedestal.Payload, "")
	helmValues, err := common.ListHelmConfigValues(params, helmConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to render pedestal helm values: %w", err)
	}

	helmConfigItem := dto.NewHelmConfigItem(helmConfig)
	helmConfigItem.DynamicValues = helmValues

	pedestalItem := dto.NewContainerVersionItem(&pedestalVersion)
	pedestalItem.Extra = helmConfigItem

	benchmarkVersionResults, err := common.MapRefsToContainerVersions([]*dto.ContainerRef{&req.Benchmark.ContainerRef}, consts.ContainerTypeBenchmark, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map benchmark container ref to version: %w", err)
	}

	benchmarkVersion, exists := benchmarkVersionResults[&req.Benchmark.ContainerRef]
	if !exists {
		return nil, fmt.Errorf("benchmark version not found for container: %s (version: %s)", req.Benchmark.Name, req.Benchmark.Version)
	}

	benchmarkVersionItem := dto.NewContainerVersionItem(&benchmarkVersion)
	envVars, err := common.ListContainerVersionEnvVars(req.Benchmark.EnvVars, &benchmarkVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to list benchmark env vars: %w", err)
	}

	benchmarkVersionItem.EnvVars = envVars

	// Parse each batch and collect items
	processedItems := make([]injectionProcessItem, 0, len(req.Specs))
	var parseWarnings []string
	for i := range req.Specs {
		item, warning, err := parseBatchInjectionSpecs(ctx, pedestalItem.ContainerName, i, req.Specs[i])
		if err != nil {
			return nil, fmt.Errorf("failed to parse injection spec batch %d: %w", i, err)
		}

		if warning != "" {
			parseWarnings = append(parseWarnings, warning)
		} else {
			processedItems = append(processedItems, *item)
		}
	}

	// Remove duplicated batches
	uniqueItems, duplicatedInRequest, alreadyExisted, err := removeDuplicated(processedItems)
	if err != nil {
		return nil, fmt.Errorf("failed to remove duplicated batches: %w", err)
	}

	// Collect warnings about duplications
	var warnings *dto.InjectionWarnings
	if len(parseWarnings) > 0 || len(duplicatedInRequest) > 0 || len(alreadyExisted) > 0 {
		warnings = &dto.InjectionWarnings{
			DuplicateServicesInBatch:  parseWarnings,
			DuplicateBatchesInRequest: duplicatedInRequest,
			BatchesExistInDatabase:    alreadyExisted,
		}
	}

	if len(req.Algorithms) > 0 {
		refs := make([]*dto.ContainerRef, 0, len(req.Algorithms))
		for i := range req.Algorithms {
			refs = append(refs, &req.Algorithms[i].ContainerRef)
		}

		algorithmVersionsResults, err := common.MapRefsToContainerVersions(refs, consts.ContainerTypeAlgorithm, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
		}

		var algorithmVersionItems []dto.ContainerVersionItem
		for i := range req.Algorithms {
			spec := &req.Algorithms[i]
			algorithmVersion, exists := algorithmVersionsResults[&spec.ContainerRef]
			if !exists {
				return nil, fmt.Errorf("algorithm version not found for %v", spec)
			}

			algorithmVersionItem := dto.NewContainerVersionItem(&algorithmVersion)
			envVars, err := common.ListContainerVersionEnvVars(spec.EnvVars, &algorithmVersion)
			if err != nil {
				return nil, fmt.Errorf("failed to list algorithm env vars: %w", err)
			}

			algorithmVersionItem.EnvVars = envVars
			algorithmVersionItems = append(algorithmVersionItems, algorithmVersionItem)
		}

		if len(algorithmVersionItems) > 0 {
			if err := client.SetHashField(ctx, consts.InjectionAlgorithmsKey, groupID, algorithmVersionItems); err != nil {
				return nil, fmt.Errorf("failed to store injection algorithms: %w", err)
			}
		}
	}

	injectionItems := make([]dto.SubmitInjectionItem, 0, len(uniqueItems))
	for _, item := range uniqueItems {
		payload := map[string]any{
			consts.RestartPedestal:      pedestalItem,
			consts.RestartHelmConfig:    helmConfig,
			consts.RestartIntarval:      req.Interval,
			consts.RestartFaultDuration: item.faultDuration,
			consts.RestartInjectPayload: map[string]any{
				consts.InjectBenchmark:   benchmarkVersionItem,
				consts.InjectPreDuration: req.PreDuration,
				consts.InjectNodes:       item.nodes,
				consts.InjectLabels:      req.Labels,
				consts.InjectSystem:      chaos.SystemType(pedestalItem.ContainerName),
			},
		}

		task := &dto.UnifiedTask{
			Type:        consts.TaskTypeRestartPedestal,
			Immediate:   false,
			ExecuteTime: item.executeTime.Unix(),
			Payload:     payload,
			GroupID:     groupID,
			ProjectID:   *projectID,
			UserID:      userID,
			State:       consts.TaskPending,
			Extra: map[consts.TaskExtra]any{
				consts.TaskExtraInjectionAlgorithms: len(req.Algorithms),
			},
		}
		task.SetGroupCtx(ctx)

		err := common.SubmitTask(ctx, task)
		if err != nil {
			return nil, fmt.Errorf("failed to submit fault injection task: %w", err)
		}

		injectionItems = append(injectionItems, dto.SubmitInjectionItem{
			Index:   item.index,
			TraceID: task.TraceID,
			TaskID:  task.TaskID,
		})
	}

	sort.Slice(injectionItems, func(i, j int) bool {
		return injectionItems[i].Index < injectionItems[j].Index
	})

	return &dto.SubmitInjectionResp{
		GroupID:       groupID,
		Items:         injectionItems,
		OriginalCount: len(processedItems),
		Warnings:      warnings,
	}, nil
}

// ProduceDatapackBuildingTasks produces datapack building tasks into Redis based on the request specifications
func ProduceDatapackBuildingTasks(ctx context.Context, req *dto.SubmitDatapackBuildingReq, groupID string, userID int, projectID *int) (*dto.SubmitDatapackBuildingResp, error) {
	if req == nil {
		return nil, fmt.Errorf("submit datapack building request is nil")
	}

	if projectID == nil {
		// Use project name from request
		project, err := repository.GetProjectByName(database.DB, req.ProjectName)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, fmt.Errorf("%w: project %s not found", consts.ErrNotFound, req.ProjectName)
			}
			return nil, fmt.Errorf("failed to get project: %w", err)
		}
		projectID = &project.ID
	}

	refs := make([]*dto.ContainerRef, 0, len(req.Specs))
	for _, spec := range req.Specs {
		refs = append(refs, &spec.Benchmark.ContainerRef)
	}

	benchmarkVersionResults, err := common.MapRefsToContainerVersions(refs, consts.ContainerTypeBenchmark, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to map container refs to versions: %w", err)
	}

	var allBuildingItems []dto.SubmitBuildingItem
	for idx, spec := range req.Specs {
		datapacks, datasetVersionID, err := extractDatapacks(database.DB, spec.Datapack, spec.Dataset, userID, consts.TaskTypeBuildDatapack)
		if err != nil {
			return nil, fmt.Errorf("failed to extract datapacks: %w", err)
		}

		benchmarkVersion, exists := benchmarkVersionResults[refs[idx]]
		if !exists {
			return nil, fmt.Errorf("benchmark version not found for %v", spec.Benchmark)
		}

		benchmarkVersionItem := dto.NewContainerVersionItem(&benchmarkVersion)
		envVars, err := common.ListContainerVersionEnvVars(spec.Benchmark.EnvVars, &benchmarkVersion)
		if err != nil {
			return nil, fmt.Errorf("failed to list benchmark env vars: %w", err)
		}

		benchmarkVersionItem.EnvVars = envVars

		var buildingItems []dto.SubmitBuildingItem
		for _, datapack := range datapacks {
			if datapack.StartTime == nil || datapack.EndTime == nil {
				return nil, fmt.Errorf("datapack %s does not have valid start_time and end_time", datapack.Name)
			}

			payload := map[string]any{
				consts.BuildBenchmark:        benchmarkVersionItem,
				consts.BuildDatapack:         dto.NewInjectionItem(&datapack),
				consts.BuildDatasetVersionID: datasetVersionID,
				consts.BuildLabels:           req.Labels,
			}

			task := &dto.UnifiedTask{
				Type:      consts.TaskTypeBuildDatapack,
				Immediate: true,
				Payload:   payload,
				GroupID:   groupID,
				ProjectID: *projectID,
				UserID:    userID,
				State:     consts.TaskPending,
			}
			task.SetGroupCtx(ctx)

			err = common.SubmitTask(ctx, task)
			if err != nil {
				return nil, fmt.Errorf("failed to submit datapack building task: %w", err)
			}

			buildingItems = append(buildingItems, dto.SubmitBuildingItem{
				Index:   idx,
				TraceID: task.TraceID,
				TaskID:  task.TaskID,
			})
		}

		allBuildingItems = append(allBuildingItems, buildingItems...)
	}

	resp := &dto.SubmitDatapackBuildingResp{
		GroupID: groupID,
		Items:   allBuildingItems,
	}
	return resp, nil
}

func batchDeleteInjectionsCore(db *gorm.DB, injectionIDs []int) error {
	executions, err := repository.ListExecutionsByDatapackIDs(db, injectionIDs)
	if err != nil {
		return fmt.Errorf("failed to list executions by datapack ids: %w", err)
	}

	if len(executions) == 0 {
		return fmt.Errorf("no executions found for the given injection ids")
	}

	executionIDs := make([]int, 0, len(executions))
	for _, execution := range executions {
		executionIDs = append(executionIDs, execution.ID)
	}

	if err := batchDeleteExecutionsCore(db, executionIDs); err != nil {
		return fmt.Errorf("failed to batch delete executions: %v", err)
	}

	if err := repository.ClearInjectionLabels(db, injectionIDs, nil); err != nil {
		return fmt.Errorf("failed to clear injection labels: %w", err)
	}

	if err := repository.BatchDeleteInjections(db, injectionIDs); err != nil {
		return fmt.Errorf("failed to delete injections: %w", err)
	}

	return nil
}

// parseBatchInjectionSpecs parses a single batch of fault injection specifications for parallel execution
// Returns the processed item, a warning message (if any), and an error
func parseBatchInjectionSpecs(ctx context.Context, pedestal string, batchIndex int, specs []chaos.Node) (*injectionProcessItem, string, error) {
	if len(specs) == 0 {
		return nil, "", fmt.Errorf("empty fault injection batch at index %d", batchIndex)
	}

	// Extract fault duration - use the maximum duration among all faults in the batch
	maxDuration := 0
	nodes := make([]chaos.Node, 0, len(specs))

	for idx, spec := range specs {
		childNode, exists := spec.Children[strconv.Itoa(spec.Value)]
		if !exists {
			return nil, "", fmt.Errorf("failed to find key %d in the children at index %d", spec.Value, idx)
		}

		if len(childNode.Children) < 3 {
			return nil, "", fmt.Errorf("no child nodes found for fault spec at index %d", idx)
		}

		faultDuration := childNode.Children[consts.DurationNodeKey].Value
		if faultDuration > maxDuration {
			maxDuration = faultDuration
		}

		systemIdx := childNode.Children[consts.SystemNodeKey].Value
		system := chaos.GetAllSystemTypes()[systemIdx]
		if pedestal != system.String() {
			return nil, "", fmt.Errorf("mismatched system type %s for pedestal %s at index %d", system.String(), pedestal, idx)
		}

		nodes = append(nodes, spec)
	}

	uniqueServices := make(map[string]int, len(nodes))
	var duplicateServiceWarnings []string
	for idx, node := range nodes {
		conf, err := chaos.NodeToStruct[chaos.InjectionConf](ctx, &node)
		if err != nil {
			return nil, "", fmt.Errorf("failed to convert node to InjectionConf at index %d: %w", idx, err)
		}

		groundtruth, err := conf.GetGroundtruth(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get groundtruth from InjectionConf at index %d: %w", idx, err)
		}

		for _, service := range groundtruth.Service {
			if service != "" {
				if oldIdx, exists := uniqueServices[service]; exists {
					duplicateServiceWarnings = append(duplicateServiceWarnings,
						fmt.Sprintf("service '%s' at positions %d and %d", service, oldIdx, idx))
					continue
				}
				uniqueServices[service] = idx
			}
		}
	}

	// Sort nodes to ensure consistent ordering
	nodes = sortNodes(nodes)

	var warning string
	if len(duplicateServiceWarnings) > 0 {
		warning = fmt.Sprintf("Batch %d contains duplicate service injections: %s",
			batchIndex, strings.Join(duplicateServiceWarnings, "; "))
	}

	return &injectionProcessItem{
		index:         batchIndex,
		faultDuration: maxDuration,
		nodes:         nodes,
	}, warning, nil
}

// flattenYAMLToParameters converts nested YAML map to flat parameter specs
func flattenYAMLToParameters(data map[string]any, prefix string) []dto.ParameterSpec {
	var params []dto.ParameterSpec

	for key, value := range data {
		fullKey := key
		if prefix != "" {
			fullKey = prefix + "." + key
		}

		switch v := value.(type) {
		case map[string]any:
			// Recursively flatten nested structures
			params = append(params, flattenYAMLToParameters(v, fullKey)...)
		case []any:
			// Convert array to JSON string
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				logrus.Warnf("Failed to marshal array for key %s: %v", fullKey, err)
				continue
			}
			params = append(params, dto.ParameterSpec{
				Key:   fullKey,
				Value: string(jsonBytes),
			})
		default:
			// Primitive values (string, int, bool, etc.)
			params = append(params, dto.ParameterSpec{
				Key:   fullKey,
				Value: v,
			})
		}
	}

	return params
}

// removeDuplicated filters out batches that already exist in DB and removes duplicates within the request
func removeDuplicated(items []injectionProcessItem) ([]injectionProcessItem, []int, []int, error) {
	engineConfigStrs := make([]string, len(items))
	for i, item := range items {
		if len(item.nodes) == 0 {
			engineConfigStrs[i] = ""
			continue
		}

		// Marshal the entire batch of nodes as the engine config
		b, err := json.Marshal(item.nodes)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("failed to marshal engine config at batch index %d: %w", i, err)
		}

		engineConfigStrs[i] = string(b)
	}

	orderedUniqueIdx := make([]int, 0, len(engineConfigStrs))
	seen := make(map[string]struct{}, len(engineConfigStrs))
	duplicatedInRequest := make([]int, 0)
	for i, key := range engineConfigStrs {
		if key == "" {
			orderedUniqueIdx = append(orderedUniqueIdx, i)
			continue
		}
		if _, ok := seen[key]; ok {
			duplicatedInRequest = append(duplicatedInRequest, items[i].index)
			continue
		}

		seen[key] = struct{}{}
		orderedUniqueIdx = append(orderedUniqueIdx, i)
	}

	existed := make(map[string]struct{})
	keys := make([]string, 0, len(seen))
	for k := range seen {
		if k != "" {
			keys = append(keys, k)
		}
	}

	batchSize := 100
	for start := 0; start < len(keys); start += batchSize {
		end := min(start+batchSize, len(keys))

		batch := keys[start:end]
		existing, err := repository.ListExistingEngineConfigs(database.DB, batch)
		if err != nil {
			return nil, nil, nil, err
		}

		for _, v := range existing {
			existed[v] = struct{}{}
		}
	}

	out := make([]injectionProcessItem, 0, len(orderedUniqueIdx))
	alreadyExisted := make([]int, 0) // Track batch indices that already exist in DB
	for _, idx := range orderedUniqueIdx {
		key := engineConfigStrs[idx]
		if key == "" {
			out = append(out, items[idx])
			continue
		}
		if _, ok := existed[key]; ok {
			alreadyExisted = append(alreadyExisted, items[idx].index)
			continue
		}

		items[idx].executeTime = time.Now().Add(time.Duration(idx*2) * time.Second)
		out = append(out, items[idx])
	}

	return out, duplicatedInRequest, alreadyExisted, nil
}

// sortNodes sorts chaos nodes by their Value field and then by their JSON representation for consistency
func sortNodes(nodes []chaos.Node) []chaos.Node {
	if len(nodes) <= 1 {
		return nodes
	}

	// Create a copy to avoid modifying the original slice
	sortedNodes := make([]chaos.Node, len(nodes))
	copy(sortedNodes, nodes)

	// Sort nodes by their Value field first, then by serialized representation for consistency
	// Using a stable sort to maintain relative order for equal elements
	for i := 0; i < len(sortedNodes)-1; i++ {
		for j := i + 1; j < len(sortedNodes); j++ {
			// Primary sort: by Value field
			if sortedNodes[i].Value > sortedNodes[j].Value {
				sortedNodes[i], sortedNodes[j] = sortedNodes[j], sortedNodes[i]
				continue
			}

			// Secondary sort: if Values are equal, sort by JSON representation for consistency
			if sortedNodes[i].Value == sortedNodes[j].Value {
				iJSON, _ := json.Marshal(sortedNodes[i])
				jJSON, _ := json.Marshal(sortedNodes[j])
				if string(iJSON) > string(jJSON) {
					sortedNodes[i], sortedNodes[j] = sortedNodes[j], sortedNodes[i]
				}
			}
		}
	}

	return sortedNodes
}

// buildFileTree recursively builds a tree structure of files and directories
func buildFileTree(workDir, relPath string, baseURL string, datapackID int, resp *dto.DatapackFilesResp) ([]dto.DatapackFileItem, error) {
	currentPath := filepath.Join(workDir, relPath)
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return nil, err
	}

	var items []dto.DatapackFileItem
	for _, entry := range entries {
		itemRelPath := filepath.Join(relPath, entry.Name())

		fileInfo, err := entry.Info()
		if err != nil {
			return nil, err
		}

		item := dto.DatapackFileItem{
			Name: entry.Name(),
			Path: filepath.ToSlash(itemRelPath),
		}

		if entry.IsDir() {
			children, err := buildFileTree(workDir, itemRelPath, baseURL, datapackID, resp)
			if err != nil {
				return nil, err
			}
			item.Children = children

			// Count direct subfolders and files
			subFolderCount := 0
			fileCount := 0
			for _, child := range children {
				if len(child.Children) > 0 {
					subFolderCount++
				} else {
					fileCount++
				}
			}

			item.Size = fmt.Sprintf("%d subfolders, %d files", subFolderCount, fileCount)
			resp.DirCount++
		} else {
			fileSize := fileInfo.Size()
			item.Size = formatFileSize(fileSize)
			modTime := fileInfo.ModTime()
			item.ModTime = &modTime
			resp.FileCount++
		}

		items = append(items, item)
	}

	return items, nil
}

// buildSafeParquetSQL inspects the parquet file schema via DuckDB DESCRIBE and builds
// a SELECT query that casts unsupported Arrow types (unsigned integers like UBIGINT)
// to compatible signed types, avoiding "Unsupported type 'uint64'" errors in Arrow IPC.
func buildSafeParquetSQL(ctx context.Context, db *sql.DB, filePath string) (string, error) {
	unsupportedCast := map[string]string{
		"UTINYINT":  "SMALLINT", // uint8  -> int16
		"USMALLINT": "INTEGER",  // uint16 -> int32
		"UINTEGER":  "INTEGER",  // uint32 -> int32
		"UBIGINT":   "INTEGER",  // uint64 -> int32 (values > 2^63-1 will overflow)
	}

	fallbackSQL := fmt.Sprintf("SELECT * FROM read_parquet('%s')", filePath)

	describeQuery := fmt.Sprintf("DESCRIBE SELECT * FROM read_parquet('%s')", filePath)
	rows, err := db.QueryContext(ctx, describeQuery)
	if err != nil {
		logrus.Warnf("failed to describe parquet file, falling back to SELECT *: %v", err)
		return fallbackSQL, nil
	}
	defer rows.Close()

	colDescs, err := rows.ColumnTypes()
	if err != nil {
		logrus.Warnf("failed to get DESCRIBE column types, falling back to SELECT *: %v", err)
		return fallbackSQL, nil
	}
	scanArgs := make([]any, len(colDescs))
	for i := range scanArgs {
		scanArgs[i] = new(sql.NullString)
	}

	var columns []string
	needsCast := false

	for rows.Next() {
		if err := rows.Scan(scanArgs...); err != nil {
			logrus.Warnf("failed to scan DESCRIBE row, falling back to SELECT *: %v", err)
			return fallbackSQL, nil
		}

		colName := scanArgs[0].(*sql.NullString).String
		colType := scanArgs[1].(*sql.NullString).String

		upperColType := strings.ToUpper(colType)

		// Check if it's an array type (ends with [])
		isArray := strings.HasSuffix(upperColType, "[]")
		baseType := upperColType
		if isArray {
			baseType = strings.TrimSuffix(upperColType, "[]")
		}

		if castType, ok := unsupportedCast[baseType]; ok {
			targetType := castType
			if isArray {
				targetType = castType + "[]"
			}
			columns = append(columns, fmt.Sprintf("CAST(\"%s\" AS %s) AS \"%s\"", colName, targetType, colName))
			needsCast = true
			logrus.Infof("parquet column %q: casting %s -> %s", colName, colType, targetType)
		} else {
			columns = append(columns, fmt.Sprintf("\"%s\"", colName))
		}
	}

	if !needsCast || len(columns) == 0 {
		return fallbackSQL, nil
	}

	safeSQL := fmt.Sprintf("SELECT %s FROM read_parquet('%s')", strings.Join(columns, ", "), filePath)
	logrus.Infof("parquet query uses type casting: %s", safeSQL)
	return safeSQL, nil
}

// formatFileSize formats bytes to human readable format (KB or MB) with one decimal place
func formatFileSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * 1024
	)

	if bytes < MB {
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(KB))
	}
	return fmt.Sprintf("%.1fMB", float64(bytes)/float64(MB))
}

func getFileFullPath(datapackID int, filePath string) (string, error) {
	datapack, err := repository.GetInjectionByID(database.DB, datapackID)
	if err != nil {
		if errors.Is(err, consts.ErrNotFound) {
			return "", fmt.Errorf("%w: datapack id: %d", consts.ErrNotFound, datapackID)
		}
		return "", fmt.Errorf("failed to get datapack: %w", err)
	}

	if datapack.State < consts.DatapackBuildSuccess {
		return "", fmt.Errorf("datapack %d is not ready for download", datapackID)
	}

	workDir := filepath.Join(config.GetString("jfs.dataset_path"), datapack.Name)
	if !utils.IsAllowedPath(workDir) {
		return "", fmt.Errorf("invalid path access to %s", workDir)
	}

	// Clean and validate the file path
	cleanPath := filepath.Clean(filePath)
	fullPath := filepath.Join(workDir, cleanPath)

	if !strings.HasPrefix(fullPath, workDir) {
		return "", fmt.Errorf("invalid file path: path traversal detected")
	}
	if !utils.IsAllowedPath(fullPath) {
		return "", fmt.Errorf("invalid file path access")
	}

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: file not found: %s", consts.ErrNotFound, cleanPath)
		}
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	if fileInfo.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", cleanPath)
	}

	return fullPath, nil
}
