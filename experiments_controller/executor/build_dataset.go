package executor

import "context"

func executeBuildDataset(ctx context.Context, taskID string, payload map[string]interface{}) error {
	return buildAlgos()
}
