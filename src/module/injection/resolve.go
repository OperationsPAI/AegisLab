package injectionmodule

import (
	"aegis/consts"
	"aegis/dto"
	"aegis/model"
	datasetmodule "aegis/module/dataset"
	"fmt"
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

func (r *Repository) ResolveDatapacks(datapackName *string, datasetRef *dto.DatasetRef, userID int, taskType consts.TaskType) ([]model.FaultInjection, *int, error) {
	states, exists := taskTypeDatapackStates[taskType]
	if !exists {
		return nil, nil, fmt.Errorf("unsupported task type: %s", consts.GetTaskTypeName(taskType))
	}

	validStates := make(map[consts.DatapackState]struct{}, len(states))
	for _, state := range states {
		validStates[state] = struct{}{}
	}

	validateDatapack := func(datapack *model.FaultInjection) error {
		if _, ok := validStates[datapack.State]; !ok {
			return fmt.Errorf("datapack %s is not in a valid state for execution", datapack.Name)
		}
		if len(datapack.Labels) > 0 && taskType == consts.TaskTypeRunAlgorithm &&
			hasLabelKeyValue(datapack.Labels, consts.LabelKeyTag, consts.DetectorNoAnomaly) {
			return fmt.Errorf("cannot execute detector algorithm on no_anomaly datapack: %s", datapack.Name)
		}
		return nil
	}

	if datapackName != nil {
		datapack, err := r.findInjectionByName(*datapackName, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get datapack: %w", err)
		}
		if err := validateDatapack(datapack); err != nil {
			return nil, nil, err
		}
		return []model.FaultInjection{*datapack}, nil, nil
	}

	if datasetRef != nil {
		datasetVersionResults, err := datasetmodule.NewRepository(r.db).ResolveDatasetVersions([]*dto.DatasetRef{datasetRef}, userID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get dataset versions: %w", err)
		}

		version, ok := datasetVersionResults[datasetRef]
		if !ok {
			return nil, nil, fmt.Errorf("dataset version not found for %v", datasetRef)
		}

		datapacks, err := datasetmodule.NewRepository(r.db).ListInjectionsByDatasetVersionID(version.ID, true)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get dataset datapacks: %s", err.Error())
		}
		if len(datapacks) == 0 {
			return nil, nil, fmt.Errorf("dataset contains no datapacks")
		}

		for i := range datapacks {
			if err := validateDatapack(&datapacks[i]); err != nil {
				return nil, nil, err
			}
		}
		return datapacks, &version.ID, nil
	}

	return nil, nil, fmt.Errorf("either datapack or dataset must be specified")
}

func hasLabelKeyValue(labels []model.Label, key, value string) bool {
	for _, label := range labels {
		if label.Key == key && label.Value == value {
			return true
		}
	}
	return false
}
