package repository

import (
	"fmt"
	"time"

	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
)

func ListDetectorResultsByExecutionID(executionID int) ([]database.Detector, error) {
	var results []database.Detector
	if err := database.DB.Model(&database.Detector{}).
		Where("execution_id = ?", executionID).
		Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to list detector results by execution ID: %v", err)
	}

	return results, nil
}

// GetDatapackDetectorResults retrieves detector results for multiple datapacks
func GetDatapackDetectorResults(req dto.DatapackDetectorReq) (*dto.DatapackDetectorResp, error) {
	resp := &dto.DatapackDetectorResp{
		TotalCount:    len(req.Datapacks),
		FoundCount:    0,
		NotFoundCount: 0,
		Items:         make([]dto.DatapackDetectorItem, 0, len(req.Datapacks)),
	}

	for _, datapackName := range req.Datapacks {
		item := dto.DatapackDetectorItem{
			Datapack:    datapackName,
			ExecutionID: 0,
			Found:       false,
			ExecutedAt:  "",
			Results:     []dto.DetectorRecord{},
		}

		// Find the latest execution for this datapack that has detector results
		var execution database.ExecutionResult
		query := database.DB.Joins("LEFT JOIN fault_injection_schedules ON fault_injection_schedules.id = execution_results.datapack_id").
			Joins("INNER JOIN detectors ON detectors.execution_id = execution_results.id").
			Where("fault_injection_schedules.injection_name = ? AND execution_results.status = ?", datapackName, consts.ExecutionSuccess)

		if req.Tag != "" {
			query = query.Joins("LEFT JOIN fault_injection_labels ON fault_injection_labels.fault_injection_id = fault_injection_schedules.id").
				Joins("LEFT JOIN labels ON labels.id = fault_injection_labels.label_id").
				Where("labels.key = ? AND labels.value = ?", consts.LabelKeyTag, req.Tag)
		}

		err := query.Order("execution_results.created_at DESC").
			First(&execution).Error

		if err != nil {
			// No execution with detector results found, add item with found=false
			resp.Items = append(resp.Items, item)
			resp.NotFoundCount++
			continue
		}

		// Found execution with detector results, now get all detector results for this execution
		var detectors []database.Detector
		err = database.DB.Where("execution_id = ?", execution.ID).Find(&detectors).Error
		if err != nil {
			return nil, fmt.Errorf("failed to get detector results for execution %d: %v", execution.ID, err)
		}

		// Convert database.Detector to dto.DetectorRecord
		detectorRecords := make([]dto.DetectorRecord, len(detectors))
		for i, detector := range detectors {
			detectorRecords[i] = dto.DetectorRecord{
				SpanName:            detector.SpanName,
				Issues:              detector.Issues,
				AbnormalAvgDuration: detector.AbnormalAvgDuration,
				NormalAvgDuration:   detector.NormalAvgDuration,
				AbnormalSuccRate:    detector.AbnormalSuccRate,
				NormalSuccRate:      detector.NormalSuccRate,
				AbnormalP90:         detector.AbnormalP90,
				NormalP90:           detector.NormalP90,
				AbnormalP95:         detector.AbnormalP95,
				NormalP95:           detector.NormalP95,
				AbnormalP99:         detector.AbnormalP99,
				NormalP99:           detector.NormalP99,
			}
		}

		item.ExecutionID = execution.ID
		item.Found = true
		item.ExecutedAt = execution.CreatedAt.Format(time.RFC3339)
		item.Results = detectorRecords

		resp.Items = append(resp.Items, item)
		resp.FoundCount++
	}

	return resp, nil
}
