package injection

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
	"aegis/model"
	"aegis/utils"

	"github.com/sirupsen/logrus"
)

type DatapackStore struct {
	basePath string
}

func NewDatapackStore() *DatapackStore {
	return &DatapackStore{basePath: config.GetString("jfs.dataset_path")}
}

func (s *DatapackStore) RootDir(datapackName string) string {
	return filepath.Join(s.basePath, datapackName)
}

func (s *DatapackStore) Package(zipWriter *zip.Writer, datapackName string, excludeRules []utils.ExculdeRule) error {
	workDir := s.RootDir(datapackName)
	if !utils.IsAllowedPath(workDir) {
		return fmt.Errorf("invalid path access to %s", workDir)
	}
	return packageDatapackDirectoryToZip(zipWriter, workDir, excludeRules)
}

func (s *DatapackStore) BuildFileTree(datapackName, baseURL string, datapackID int) (*DatapackFilesResp, error) {
	workDir := s.RootDir(datapackName)
	if !utils.IsAllowedPath(workDir) {
		return nil, fmt.Errorf("invalid path access to %s", workDir)
	}
	if _, err := os.Stat(workDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("datapack directory not found for datapack id %d", datapackID)
	}

	resp := &DatapackFilesResp{
		Files:     []DatapackFileItem{},
		FileCount: 0,
		DirCount:  0,
	}

	rootItems, err := buildFileTree(workDir, "", baseURL, datapackID, resp)
	if err != nil {
		return nil, err
	}
	resp.Files = rootItems
	return resp, nil
}

func (s *DatapackStore) OpenFile(datapackName, filePath string) (string, string, int64, io.ReadSeekCloser, error) {
	fullPath, err := s.resolveFilePath(datapackName, filePath)
	if err != nil {
		return "", "", 0, nil, err
	}

	file, err := os.Open(fullPath)
	if err != nil {
		return "", "", 0, nil, fmt.Errorf("failed to open file: %w", err)
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return "", "", 0, nil, fmt.Errorf("failed to stat file: %w", err)
	}

	fileName := filepath.Base(fullPath)
	contentType := "application/octet-stream"
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

func (s *DatapackStore) ResolveFilePath(datapackName, filePath string) (string, error) {
	return s.resolveFilePath(datapackName, filePath)
}

func (s *DatapackStore) resolveFilePath(datapackName, filePath string) (string, error) {
	workDir := s.RootDir(datapackName)
	if !utils.IsAllowedPath(workDir) {
		return "", fmt.Errorf("invalid path access to %s", workDir)
	}

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

func buildFileTree(workDir, relPath string, baseURL string, datapackID int, resp *DatapackFilesResp) ([]DatapackFileItem, error) {
	_ = baseURL
	_ = datapackID
	currentPath := filepath.Join(workDir, relPath)
	entries, err := os.ReadDir(currentPath)
	if err != nil {
		return nil, err
	}

	var items []DatapackFileItem
	for _, entry := range entries {
		itemRelPath := filepath.Join(relPath, entry.Name())
		fileInfo, err := entry.Info()
		if err != nil {
			return nil, err
		}

		item := DatapackFileItem{
			Name: entry.Name(),
			Path: filepath.ToSlash(itemRelPath),
		}

		if entry.IsDir() {
			children, err := buildFileTree(workDir, itemRelPath, baseURL, datapackID, resp)
			if err != nil {
				return nil, err
			}
			item.Children = children

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

func formatFileSize(bytes int64) string {
	const (
		kb = 1024
		mb = 1024 * 1024
	)

	if bytes < mb {
		return fmt.Sprintf("%.1fKB", float64(bytes)/float64(kb))
	}
	return fmt.Sprintf("%.1fMB", float64(bytes)/float64(mb))
}

func (s *DatapackStore) CreateUploadTempFile() (*os.File, error) {
	return os.CreateTemp("", "datapack-upload-*.zip")
}

func (s *DatapackStore) ValidateArchive(zipPath string) error {
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

func (s *DatapackStore) EnsureDatapackDirAvailable(datapackName string) (string, error) {
	if s.basePath == "" {
		return "", fmt.Errorf("dataset path not configured")
	}
	targetDir := s.RootDir(datapackName)
	if _, err := os.Stat(targetDir); err == nil {
		return "", fmt.Errorf("%w: directory %s already exists", consts.ErrAlreadyExists, datapackName)
	}
	return targetDir, nil
}

func (s *DatapackStore) ExtractArchive(zipPath, targetDir string) error {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("failed to open zip archive: %w", err)
	}
	defer func() { _ = r.Close() }()

	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	for _, f := range r.File {
		destPath := filepath.Join(targetDir, f.Name)
		if !strings.HasPrefix(filepath.Clean(destPath), filepath.Clean(targetDir)+string(os.PathSeparator)) &&
			filepath.Clean(destPath) != filepath.Clean(targetDir) {
			return fmt.Errorf("illegal file path in archive: %s", f.Name)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0o755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", f.Name, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
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

func (s *DatapackStore) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

func (s *DatapackStore) Remove(path string) error {
	return os.Remove(path)
}

func (s *DatapackStore) ExtractGroundtruths(dir string) []model.Groundtruth {
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

	rawGTs := parsed.Groundtruths
	if len(rawGTs) == 0 {
		rawGTs = parsed.GroundTruth
	}
	if len(rawGTs) == 0 {
		return nil
	}

	result := make([]model.Groundtruth, 0, len(rawGTs))
	for _, gt := range rawGTs {
		result = append(result, model.Groundtruth{
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

var validParquetFiles = map[string]bool{
	"abnormal_traces.parquet":  true,
	"abnormal_metrics.parquet": true,
	"abnormal_logs.parquet":    true,
	"normal_traces.parquet":    true,
	"normal_metrics.parquet":   true,
	"normal_logs.parquet":      true,
}
