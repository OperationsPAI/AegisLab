package consumer

import (
	"aegis/client/k8s"
	"aegis/consts"
	"fmt"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

const (
	minDelayMinutes  = 1
	maxDelayMinutes  = 5
	customTimeFormat = "20060102_150405"
)

// getRequiredVolumeMountConfigs retrieves the volume mount configurations for the specified required keys
func getRequiredVolumeMountConfigs(requiredKeys []consts.VolumeMountName) ([]k8s.VolumeMountConfig, error) {
	volumeMountConfigMap, err := k8s.GetVolumeMountConfigMap()
	if err != nil {
		return nil, fmt.Errorf("failed to get volume mount configuration map: %w", err)
	}

	volumeMountConfigs := make([]k8s.VolumeMountConfig, 0, len(requiredKeys))

	for _, vmName := range requiredKeys {
		cfg, exists := volumeMountConfigMap[vmName]
		if !exists {
			return nil, fmt.Errorf("volume mount configuration %s not found", vmName)
		}
		volumeMountConfigs = append(volumeMountConfigs, cfg)
	}

	return volumeMountConfigs, nil
}

// handleExecutionError is a helper function to handle errors consistently across all executors
// It logs the error, adds span events, records error to span, and returns a wrapped error
func handleExecutionError(span trace.Span, logEntry *logrus.Entry, message string, err error) error {
	logEntry.Errorf("%s: %v", message, err)
	span.AddEvent(message)
	span.RecordError(err)
	return fmt.Errorf("%s: %w", message, err)
}
