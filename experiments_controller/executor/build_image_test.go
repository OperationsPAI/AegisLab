package executor

import (
	"context"

	"testing"
)

func TestBuildImage(t *testing.T) {
	err := buildDockerfileAndPush(context.Background(), BuildOptions{
		DockerfilePath: "/home/nn/workspace/rcabench/algorithms/detector/builder.Dockerfile", // todo: remove the absolute path
		ImageName:      "10.10.10.240/library/detector11:latest",
		BuildArgs:      map[string]string{},
		ContextDir:     ".",
		Target:         "",
	})
	if err != nil {
		t.Error(err)
	}
}
