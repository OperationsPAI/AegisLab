package producer

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/sirupsen/logrus"
)

// validParquetFiles is the set of recognized parquet files in a datapack archive
var validParquetFiles = map[string]bool{
	"abnormal_traces.parquet":  true,
	"abnormal_metrics.parquet": true,
	"abnormal_logs.parquet":    true,
	"normal_traces.parquet":    true,
	"normal_metrics.parquet":   true,
	"normal_logs.parquet":      true,
}

// UploadDatapack handles the business logic for uploading a manual datapack
func UploadDatapack(req *dto.UploadDatapackReq, file io.Reader, fileSize int64) (*dto.UploadDatapackResp, error) {
	// Parse labels and groundtruths from request
	labels, err := req.ParseLabels()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", consts.ErrBadRequest, err.Error())
	}

	groundtruths, err := req.ParseGroundtruths()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", consts.ErrBadRequest, err.Error())
	}

	// Check name uniqueness
	existing, _ := repository.GetInjectionByName(database.DB, req.Name, false)
	if existing != nil {
		return nil, fmt.Errorf("%w: injection with name %s already exists", consts.ErrAlreadyExists, req.Name)
	}

	// Save uploaded file to temp location
	tmpFile, err := os.CreateTemp("", "datapack-upload-*.zip")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	if _, err := io.Copy(tmpFile, file); err != nil {
		_ = tmpFile.Close()
		return nil, fmt.Errorf("failed to save uploaded file: %w", err)
	}
	_ = tmpFile.Close()

	// Validate archive contents
	if err := validateDatapackArchive(tmpPath); err != nil {
		return nil, fmt.Errorf("%w: %s", consts.ErrBadRequest, err.Error())
	}

	// Get target directory
	datasetPath := config.GetString("jfs.dataset_path")
	if datasetPath == "" {
		return nil, fmt.Errorf("dataset path not configured")
	}
	targetDir := filepath.Join(datasetPath, req.Name)

	// Ensure target directory does not already exist
	if _, err := os.Stat(targetDir); err == nil {
		return nil, fmt.Errorf("%w: directory %s already exists", consts.ErrAlreadyExists, req.Name)
	}

	// Extract zip to target directory
	if err := extractZipToDir(tmpPath, targetDir); err != nil {
		// Clean up on failure
		_ = os.RemoveAll(targetDir)
		return nil, fmt.Errorf("failed to extract archive: %w", err)
	}

	// Determine ground truth source
	groundtruthSource := ""
	if len(groundtruths) > 0 {
		// Ground truth was provided in the request
		groundtruthSource = consts.GroundtruthSourceManual
	} else {
		// Try to extract ground truth from injection.json if not provided in request
		groundtruths = extractGroundtruthFromInjectionJSON(targetDir)
		if len(groundtruths) > 0 {
			groundtruthSource = consts.GroundtruthSourceImported
		}
	}

	// Create FaultInjection record
	category := chaos.SystemType("")
	if req.Category != "" {
		category = chaos.SystemType(req.Category)
	}

	injection := &database.FaultInjection{
		Name:              req.Name,
		Source:            consts.DatapackSourceManual,
		FaultType:         chaos.ChaosType(0),
		Category:          category,
		Description:       req.Description,
		EngineConfig:      "",
		Groundtruths:      groundtruths,
		GroundtruthSource: groundtruthSource,
		PreDuration:       0,
		BenchmarkID:       nil,
		PedestalID:        nil,
		State:             consts.DatapackBuildSuccess,
		Status:            consts.CommonEnabled,
	}

	if err := CreateInjection(injection, labels); err != nil {
		// Clean up extracted files on DB failure
		_ = os.RemoveAll(targetDir)
		return nil, err
	}

	return &dto.UploadDatapackResp{
		ID:   injection.ID,
		Name: injection.Name,
	}, nil
}

// validateDatapackArchive checks that the zip archive contains at least one recognized parquet file
func validateDatapackArchive(zipPath string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		name := filepath.Base(f.Name)
		if validParquetFiles[name] {
			return nil
		}
	}

	return fmt.Errorf("archive must contain at least one parquet file from: abnormal_traces.parquet, abnormal_metrics.parquet, abnormal_logs.parquet, normal_traces.parquet, normal_metrics.parquet, normal_logs.parquet")
}

// extractZipToDir extracts a zip archive to the target directory with path traversal protection
func extractZipToDir(zipPath, targetDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer func() { _ = r.Close() }()

	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	for _, f := range r.File {
		// Path traversal protection
		destPath := filepath.Join(targetDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(targetDir)+string(os.PathSeparator)) &&
			filepath.Clean(destPath) != filepath.Clean(targetDir) {
			return fmt.Errorf("illegal file path in archive: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", f.Name, err)
			}
			continue
		}

		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for %s: %w", f.Name, err)
		}

		outFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return fmt.Errorf("failed to create file %s: %w", f.Name, err)
		}

		rc, err := f.Open()
		if err != nil {
			_ = outFile.Close()
			return fmt.Errorf("failed to open file in archive %s: %w", f.Name, err)
		}

		_, err = io.Copy(outFile, rc)
		_ = rc.Close()
		_ = outFile.Close()
		if err != nil {
			return fmt.Errorf("failed to extract file %s: %w", f.Name, err)
		}
	}

	return nil
}

// injectionJSONGroundtruth represents the ground truth structure in injection.json
type injectionJSONGroundtruth struct {
	Service   []string `json:"service,omitempty"`
	Pod       []string `json:"pod,omitempty"`
	Container []string `json:"container,omitempty"`
	Metric    []string `json:"metric,omitempty"`
	Function  []string `json:"function,omitempty"`
	Span      []string `json:"span,omitempty"`
}

type injectionJSONFile struct {
	Groundtruths []injectionJSONGroundtruth `json:"ground_truths"`
	GroundTruth  []injectionJSONGroundtruth `json:"ground_truth"`
}

// extractGroundtruthFromInjectionJSON tries to read ground truth from injection.json in the directory
func extractGroundtruthFromInjectionJSON(dir string) []database.Groundtruth {
	jsonPath := filepath.Join(dir, "injection.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		logrus.Debugf("No injection.json found in %s: %v", dir, err)
		return nil
	}

	var parsed injectionJSONFile
	if err := json.Unmarshal(data, &parsed); err != nil {
		logrus.Warnf("Failed to parse injection.json in %s: %v", dir, err)
		return nil
	}

	// Try ground_truths first, then ground_truth
	rawGTs := parsed.Groundtruths
	if len(rawGTs) == 0 {
		rawGTs = parsed.GroundTruth
	}

	if len(rawGTs) == 0 {
		return nil
	}

	result := make([]database.Groundtruth, 0, len(rawGTs))
	for _, gt := range rawGTs {
		result = append(result, database.Groundtruth{
			Service:   gt.Service,
			Pod:       gt.Pod,
			Container: gt.Container,
			Metric:    gt.Metric,
			Function:  gt.Function,
			Span:      gt.Span,
		})
	}

	return result
}
