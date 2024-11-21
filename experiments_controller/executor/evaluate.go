package executor

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"dagger.io/dagger"
	"dagger.io/dagger/dag"
	"github.com/sirupsen/logrus"
)

type Rcabench struct{}

func (m *Rcabench) BuildBenchmarkDataImage(
	ctx context.Context,
	src *dagger.Directory,
	abnorStartTime, abnorEndTime, norStartTime, norEndTime time.Time,
) *dagger.Container {
	workspace := dag.Container().
		WithDirectory("/app", src).
		WithWorkdir("/app").
		Directory("/app")

	logrus.Infof("timestamp: %v %v %v %v", abnorStartTime, abnorEndTime, norStartTime, norEndTime)
	return dag.Container().
		WithEnvVariable("CACHEBUSTER", time.Now().String()). // 强制重新编译
		Build(workspace, dagger.ContainerBuildOpts{
			BuildArgs: []dagger.BuildArg{
				{
					Name:  "ABNORMAL_START_VAR",
					Value: strconv.Itoa(int(abnorStartTime.Unix())),
				},
				{
					Name:  "ABNORMAL_END_VAR",
					Value: strconv.Itoa(int(abnorEndTime.Unix())),
				},
				{
					Name:  "NORMAL_START_VAR",
					Value: strconv.Itoa(int(norStartTime.Unix())),
				},
				{
					Name:  "NORMAL_END_VAR",
					Value: strconv.Itoa(int(norEndTime.Unix())),
				},
			},
		})
}

func (m *Rcabench) BuildAlgoBuilderImage(ctx context.Context, src *dagger.Directory) *dagger.Container {
	workspace := dag.Container().
		WithDirectory("/app", src).
		WithWorkdir("/app").
		Directory("/app")

	return dag.Container().
		Build(workspace, dagger.ContainerBuildOpts{
			Dockerfile: "builder.Dockerfile",
		})
}

func (m *Rcabench) BuildAlgoRunnerImage(
	ctx context.Context, bench_dir, src *dagger.Directory, start_script *dagger.File,
	abnorStartTime, abnorEndTime, norStartTime, norEndTime time.Time,
) *dagger.Container {
	data := m.BuildBenchmarkDataImage(ctx, bench_dir, abnorStartTime, abnorEndTime, norStartTime, norEndTime)
	builder := m.BuildAlgoBuilderImage(ctx, src)
	runner := builder.
		WithWorkdir("/app").
		WithDirectory("/app/input", data.Directory("/app/input")).
		WithFile("/app/rca.py", src.File("rca.py")).
		WithFile("/app/run_exp.py", start_script).
		WithEnvVariable("WORKSPACE", "/app").
		WithEnvVariable("CACHEBUSTER", time.Now().String()). // 强制重新编译
		WithEnvVariable("ABNORMAL_START", strconv.Itoa(int(abnorStartTime.Unix()))).
		WithEnvVariable("ABNORMAL_END", strconv.Itoa(int(abnorEndTime.Unix()))).
		WithEnvVariable("NORMAL_START", strconv.Itoa(int(norStartTime.Unix()))).
		WithEnvVariable("NORMAL_END", strconv.Itoa(int(norEndTime.Unix())))

	return runner
}
func (m *Rcabench) Evaluate(
	ctx context.Context, bench_dir, src *dagger.Directory, start_script *dagger.File,
	abnorStartTime, abnorEndTime, norStartTime, norEndTime time.Time,
) *dagger.Directory {
	return m.BuildAlgoRunnerImage(ctx, bench_dir, src, start_script, abnorStartTime, abnorEndTime, norStartTime, norEndTime).
		WithExec([]string{"python", "run_exp.py"}).
		Directory("/app/output")
}

func (m *Rcabench) Publish(ctx context.Context, registry string, username string, password *dagger.Secret,
) (string, error) {
	return dag.Container().
		From("nginx:1.23-alpine").
		WithNewFile(
			"/usr/share/nginx/html/index.html",
			"Hello from Dagger!",
			dagger.ContainerWithNewFileOpts{Permissions: 0o400},
		).
		WithRegistryAuth(registry, username, password).
		Publish(ctx, fmt.Sprintf("%s/library/my-nginx", registry))
}
