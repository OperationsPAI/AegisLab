package executor

import (
	"context"

	"testing"
)

func TestBuildImage(t *testing.T) {
	err := buildDockerfileAndPush(context.Background(), BuildOptions{
		DockerfilePath: "/home/nn/workspace/rcabench/algorithms/detector/builder.Dockerfile", // todo: remove the absolute path
		ImageName:      "10.10.10.240/library/detecto1:latest12",
		BuildArgs:      map[string]string{},
		ContextDir:     "/home/nn/workspace/rcabench/algorithms/detector",
		Target:         "",
	})
	if err != nil {
		t.Error(err)
	}
}
