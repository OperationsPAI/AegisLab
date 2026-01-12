package producer

import (
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
	"aegis/service/common"
	"fmt"

	"gorm.io/gorm"
)

var taskTypeDatapackStates = map[consts.TaskType][]consts.DatapackState{
	consts.TaskTypeBuildDatapack: {
		consts.DatapackInjectSuccess,
		consts.DatapackBuildFailed,
		consts.DatapackBuildSuccess,
		consts.DatapackDetectorFailed,
		consts.DatapackDetectorSuccess,
	},
	consts.TaskTypeRunAlgorithm: {
		consts.DatapackDetectorSuccess,
	},
}

// checkLabelKeyValue checks if a label with the specified key and value exists in the provided label slice
func checkLabelKeyValue(labels []database.Label, key, value string) bool {
	for _, label := range labels {
		if label.Key == key && label.Value == value {
			return true
		}
	}
	return false
}

// extractDatapacks extracts datapacks based on the provided datapack name or dataset ref
func extractDatapacks(db *gorm.DB, datapackName *string, datasetRef *dto.DatasetRef, userID int, taskType consts.TaskType) ([]database.FaultInjection, *int, error) {
	states, exists := taskTypeDatapackStates[taskType]
	if !exists {
		return nil, nil, fmt.Errorf("unsupported task type: %s", consts.GetTaskTypeName(taskType))
	}

	validStates := map[consts.DatapackState]struct{}{}
	for _, state := range states {
		validStates[state] = struct{}{}
	}

	// validateDatapack validates a single datapack's state and labels
	validateDatapack := func(datapack *database.FaultInjection) error {
		if _, exists := validStates[datapack.State]; !exists {
			return fmt.Errorf("datapack %s is not in a valid state for execution", datapack.Name)
		}

		if len(datapack.Labels) > 0 && taskType == consts.TaskTypeRunAlgorithm {
			if exists := checkLabelKeyValue(datapack.Labels, consts.LabelKeyTag, consts.DetectorNoAnomaly); exists {
				return fmt.Errorf("cannot execute detector algorithm on no_anomaly datapack: %s", datapack.Name)
			}
		}

		return nil
	}

	if datapackName != nil {
		datapack, err := repository.GetInjectionByName(db, *datapackName, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get datapack: %w", err)
		}

		if err := validateDatapack(datapack); err != nil {
			return nil, nil, err
		}

		return []database.FaultInjection{*datapack}, nil, nil
	}

	if datasetRef != nil {
		datasetVersionResults, err := common.MapRefsToDatasetVersions([]*dto.DatasetRef{datasetRef}, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get dataset versions: %w", err)
		}

		version, exists := datasetVersionResults[datasetRef]
		if !exists {
			return nil, nil, fmt.Errorf("dataset version not found for %v", datasetRef)
		}

		datapacks, err := repository.ListInjectionsByDatasetVersionID(db, version.ID, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get dataset datapacks: %s", err.Error())
		}

		if len(datapacks) == 0 {
			return nil, nil, fmt.Errorf("dataset contains no datapacks")
		}

		for _, datapack := range datapacks {
			if err := validateDatapack(&datapack); err != nil {
				return nil, nil, err
			}
		}

		return datapacks, &version.ID, nil
	}

	return nil, nil, fmt.Errorf("either datapack or dataset must be specified")
}
