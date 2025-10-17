package utils

import (
	"aegis/config"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/sirupsen/logrus"
)

type JobLogWriter struct {
	logDir string
}

func NewJobLogWriter() (*JobLogWriter, error) {
	logDir := filepath.Join(config.GetString("logging.dir"), config.GetString("logging.job.dir"))
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create job log directory %s: %v", logDir, err)
	}

	return &JobLogWriter{
		logDir: logDir,
	}, nil
}

// CleanOldLogs removes log files older than retention days
func (w *JobLogWriter) CleanOldLogs() error {
	cutoffTime := time.Now().AddDate(0, 0, -config.GetInt("logging.job.log_retention_days"))

	err := filepath.Walk(w.logDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		if info.ModTime().Before(cutoffTime) {
			if err := os.Remove(path); err != nil {
				logrus.Warnf("Failed to remove old log file %s: %v", path, err)
			} else {
				logrus.Debugf("Removed old log file: %s", path)
			}
		}

		return nil
	})

	return err
}

// WriteJobLogs writes job logs to a file
// Returns the file path where logs were written
func (w *JobLogWriter) WriteJobLogs(jobName, namespace, traceID string, logMap map[string][]string) (string, error) {
	if len(logMap) == 0 {
		return "", fmt.Errorf("no logs to write for job %s in namespace %s", jobName, namespace)
	}

	// Generate log file path: jobs/{namespace}/{date}/{timestamp}-{jobName}.log
	now := time.Now()
	dateDir := now.Format("2006-01-02")
	logSubDir := filepath.Join(w.logDir, namespace, dateDir, jobName)

	if err := os.MkdirAll(logSubDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create log subdirectory: %v", err)
	}

	timestamp := now.Format("150405") // HHMMSS
	filename := fmt.Sprintf("%s-%s.json", jobName, timestamp)
	filePath := filepath.Join(logSubDir, filename)

	content := map[string]any{
		"job_name":  jobName,
		"namespace": namespace,
		"trace_id":  traceID,
		"timestamp": now.Format(time.RFC3339),
		"pod_logs":  logMap,
	}

	jsonData, err := json.MarshalIndent(content, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal log data: %v", err)
	}

	if err := os.WriteFile(filePath, jsonData, 0644); err != nil {
		return "", fmt.Errorf("failed to write log file: %v", err)
	}

	logrus.WithFields(logrus.Fields{
		"job_name": jobName,
	}).Infof("Job logs written to file %s", filePath)

	return filePath, nil
}
