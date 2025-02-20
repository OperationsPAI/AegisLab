package executor

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"
)

func TestBuildImage(t *testing.T) {
	ctxDir := "/home/nn/workspace/rcabench"
	algoDir := "/home/nn/workspace/rcabench/algorithms/e-diagnose"
	imageName := fmt.Sprintf("%s/%s:%d", "10.10.10.240/library", "e-diagnose", time.Now().UnixNano())

	err := buildDockerfileAndPush(context.Background(), BuildOptions{
		DockerfilePath: filepath.Join(algoDir, "builder.Dockerfile"), // todo: remove the absolute path
		ImageName:      imageName,
		BuildArgs:      nil,
		ContextDir:     ctxDir,
		Target:         "",
	})
	if err != nil {
		t.Error(err)
	}
}
