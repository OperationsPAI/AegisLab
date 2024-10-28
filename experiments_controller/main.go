package main

import (
	"context"
	"dagger/rcabench/internal/dagger"
	"fmt"
)

type Rcabench struct{}

func (m *Rcabench) BuildBenchmarkDataImage(ctx context.Context, src *dagger.Directory) *dagger.Container {
	workspace := dag.Container().
		WithDirectory("/app", src).
		WithWorkdir("/app").
		Directory("/app")

	return dag.Container().
		Build(workspace)
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

func (m *Rcabench) BuildAlgoRunnerImage(ctx context.Context, bench_dir, src *dagger.Directory, start_script *dagger.File) *dagger.Container {
	data := m.BuildBenchmarkDataImage(ctx, bench_dir)
	builder := m.BuildAlgoBuilderImage(ctx, src)

	runner := builder.
		WithWorkdir("/app").
		WithDirectory("/app/input", data.Directory("/app/input")).
		WithFile("/app/rca.py", src.File("rca.py")).
		WithFile("/app/run_exp.py", start_script).
		WithEnvVariable("WORKSPACE", "/app")

	return runner
}
func (m *Rcabench) Evaluate(ctx context.Context, bench_dir, src *dagger.Directory, start_script *dagger.File) *dagger.Directory {
	return m.BuildAlgoRunnerImage(ctx, bench_dir, src, start_script).
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

func (m *Rcabench) Test1(input string) string {
	fmt.Println(input)
	return input
}
