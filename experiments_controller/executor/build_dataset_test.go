package executor

import (
	"fmt"
	"testing"
	"time"

	"github.com/CUHK-SE-Group/rcabench/client"
)

func TestCreateDatasetJob(t *testing.T) {
	err := createDatasetJob("testing", "experiment", fmt.Sprintf("10.10.10.240/library/clickhouse_dataset:latest"), []string{"python", "prepare_inputs.py"}, time.Now().Add(-100*time.Minute), time.Now().Add(-95*time.Minute))
	if err != nil {
		t.Error(err)
	}
}
func TestDeleteDatasetJob(t *testing.T) {
	k8sClient := client.NewK8sClient()
	client.DeleteK8sJob(k8sClient, "experiment", "testing")
}
