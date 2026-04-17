package containermodule

import (
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"time"

	"aegis/config"
	"aegis/utils"

	"github.com/sirupsen/logrus"
)

type HelmFileStore struct {
	basePath string
}

func NewHelmFileStore() *HelmFileStore {
	return &HelmFileStore{basePath: config.GetString("jfs.dataset_path")}
}

func (s *HelmFileStore) SaveChart(containerName string, file *multipart.FileHeader) (string, string, error) {
	if s.basePath == "" {
		return "", "", fmt.Errorf("jfs.dataset_path is not configured")
	}

	targetDir := filepath.Join(s.basePath, "helm-charts")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", "", fmt.Errorf("failed to create directory: %w", err)
	}

	targetPath := filepath.Join(
		targetDir,
		fmt.Sprintf("%s_chart_%d%s", containerName, time.Now().Unix(), filepath.Ext(file.Filename)),
	)
	if err := utils.CopyFileFromFileHeader(file, targetPath); err != nil {
		return "", "", fmt.Errorf("failed to save chart file: %w", err)
	}

	checksum, err := utils.CalculateFileSHA256(targetPath)
	if err != nil {
		logrus.WithField("file_path", targetPath).Warnf("failed to calculate checksum: %v", err)
		checksum = ""
	}

	logrus.WithFields(logrus.Fields{
		"file_path": targetPath,
		"checksum":  checksum,
	}).Info("Helm chart package uploaded successfully")

	return targetPath, checksum, nil
}

func (s *HelmFileStore) SaveValueFile(containerName string, srcFileHeader *multipart.FileHeader, srcFilePath string) (string, error) {
	if s.basePath == "" {
		return "", fmt.Errorf("jfs.dataset_path is not configured")
	}

	targetDir := filepath.Join(s.basePath, "helm-values")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	timestamp := time.Now().Unix()
	var targetPath string

	switch {
	case srcFileHeader != nil:
		targetPath = filepath.Join(targetDir, fmt.Sprintf("%s_values_%d%s", containerName, timestamp, filepath.Ext(srcFileHeader.Filename)))
		if err := utils.CopyFileFromFileHeader(srcFileHeader, targetPath); err != nil {
			return "", fmt.Errorf("failed to save file: %w", err)
		}
	case srcFilePath != "":
		targetPath = filepath.Join(targetDir, fmt.Sprintf("%s_values_%d%s", containerName, timestamp, filepath.Ext(srcFilePath)))
		if err := utils.CopyFile(srcFilePath, targetPath); err != nil {
			return "", fmt.Errorf("failed to save file: %w", err)
		}
	default:
		return "", fmt.Errorf("either source file header or source file path is required")
	}

	logrus.WithField("file_path", targetPath).Info("Helm values file uploaded successfully")
	return targetPath, nil
}
