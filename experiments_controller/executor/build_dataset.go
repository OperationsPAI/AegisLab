package executor

import (
	"context"
	"errors"
	"fmt"
	"time"

	chaosCli "github.com/CUHK-SE-Group/chaos-experiment/client"
	"github.com/CUHK-SE-Group/rcabench/database"
	"gorm.io/gorm"
)

type DatasetPayload struct {
	DatasetName string
}

func parseDatasetPayload(payload map[string]interface{}) (*DatasetPayload, error) {
	datasetName, ok := payload[EvalPayloadDataset].(string)
	if !ok || datasetName == "" {
		return nil, fmt.Errorf("missing or invalid '%s' key in payload", EvalPayloadDataset)
	}
	return &DatasetPayload{
		DatasetName: datasetName,
	}, nil
}

func executeBuildDataset(ctx context.Context, taskID string, payload map[string]interface{}) error {
	datasetPayload, err := parseAlgorithmExecutionPayload(payload)
	if err != nil {
		return err
	}

	var faultRecord database.FaultInjectionSchedule
	err = database.DB.Where("injection_name = ?", datasetPayload.DatasetName).First(&faultRecord).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("no matching fault injection record found for dataset: %s", datasetPayload.DatasetName)
		}
		return fmt.Errorf("failed to query database for dataset: %s, error: %v", datasetPayload.DatasetName, err)
	}

	var startTime, endTime time.Time
	if faultRecord.Status == database.DatasetSuccess {
		startTime = faultRecord.StartTime
		endTime = faultRecord.EndTime
	} else if faultRecord.Status == database.DatasetInitial {
		startTime, endTime, err = chaosCli.QueryCRDByName("ts", datasetPayload.DatasetName)
		if err != nil {
			return fmt.Errorf("failed to QueryCRDByName: %s, error: %v", datasetPayload.DatasetName, err)
		}
		if err := database.DB.Model(&faultRecord).Where("injection_name = ?", datasetPayload.DatasetName).
			Updates(map[string]interface{}{
				"start_time": startTime,
				"end_time":   endTime,
			}).Error; err != nil {
			return fmt.Errorf("failed to update start_time and end_time for dataset: %s, error: %v", datasetPayload.DatasetName, err)
		}
	}

}
