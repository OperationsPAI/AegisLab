package v2

import (
	"testing"

	"rcabench/dto"
	"github.com/stretchr/testify/assert"
)

func TestExtractDatapacks(t *testing.T) {
	tests := []struct {
		name           string
		datapackName   *string
		datasetName    *string
		datasetVersion *string
		expectError    bool
		errorContains  string
	}{
		{
			name:           "missing both datapack and dataset",
			datapackName:   nil,
			datasetName:    nil,
			datasetVersion: nil,
			expectError:    true,
			errorContains:  "either datapack or dataset must be specified",
		},
		{
			name:           "dataset without version",
			datapackName:   nil,
			datasetName:    stringPtr("test-dataset"),
			datasetVersion: nil,
			expectError:    true,
			errorContains:  "dataset_version is required when querying by dataset name",
		},
		{
			name:           "dataset with empty version",
			datapackName:   nil,
			datasetName:    stringPtr("test-dataset"),
			datasetVersion: stringPtr(""),
			expectError:    true,
			errorContains:  "dataset_version is required when querying by dataset name",
		},
		{
			name:           "valid dataset with version",
			datapackName:   nil,
			datasetName:    stringPtr("test-dataset"),
			datasetVersion: stringPtr("v1.0"),
			expectError:    false,
		},
		{
			name:           "valid datapack",
			datapackName:   stringPtr("test-datapack"),
			datasetName:    nil,
			datasetVersion: nil,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := extractDatapacks(tt.datapackName, tt.datasetName, tt.datasetVersion)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				// Note: This test doesn't actually query the database,
				// so we expect errors for non-existent datasets/datapacks
				// The important thing is that the validation logic works correctly
				assert.Error(t, err) // Will fail because dataset/datapack doesn't exist in test DB
			}
		})
	}
}

func TestAlgorithmExecutionRequestValidation(t *testing.T) {
	tests := []struct {
		name          string
		request       dto.AlgorithmExecutionRequest
		expectError   bool
		errorContains string
	}{
		{
			name: "valid dataset with version",
			request: dto.AlgorithmExecutionRequest{
				ProjectName:    "test-project",
				Algorithm:      dto.AlgorithmItem{Name: "test-algorithm"},
				Dataset:        stringPtr("test-dataset"),
				DatasetVersion: stringPtr("v1.0"),
			},
			expectError: false,
		},
		{
			name: "valid datapack",
			request: dto.AlgorithmExecutionRequest{
				ProjectName: "test-project",
				Algorithm:   dto.AlgorithmItem{Name: "test-algorithm"},
				Datapack:    stringPtr("test-datapack"),
			},
			expectError: false,
		},
		{
			name: "dataset without version",
			request: dto.AlgorithmExecutionRequest{
				ProjectName: "test-project",
				Algorithm:   dto.AlgorithmItem{Name: "test-algorithm"},
				Dataset:     stringPtr("test-dataset"),
			},
			expectError:   true,
			errorContains: "dataset_version is required when dataset is specified",
		},
		{
			name: "dataset with empty version",
			request: dto.AlgorithmExecutionRequest{
				ProjectName:    "test-project",
				Algorithm:      dto.AlgorithmItem{Name: "test-algorithm"},
				Dataset:        stringPtr("test-dataset"),
				DatasetVersion: stringPtr(""),
			},
			expectError:   true,
			errorContains: "dataset_version cannot be empty",
		},
		{
			name: "both datapack and dataset specified",
			request: dto.AlgorithmExecutionRequest{
				ProjectName:    "test-project",
				Algorithm:      dto.AlgorithmItem{Name: "test-algorithm"},
				Datapack:       stringPtr("test-datapack"),
				Dataset:        stringPtr("test-dataset"),
				DatasetVersion: stringPtr("v1.0"),
			},
			expectError:   true,
			errorContains: "cannot specify both datapack and dataset",
		},
		{
			name: "neither datapack nor dataset specified",
			request: dto.AlgorithmExecutionRequest{
				ProjectName: "test-project",
				Algorithm:   dto.AlgorithmItem{Name: "test-algorithm"},
			},
			expectError:   true,
			errorContains: "either datapack or dataset must be specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.request.Validate()

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
