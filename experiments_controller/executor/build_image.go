package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	con "dagger/rcabench/config"

	"github.com/docker/cli/cli/config"
	"github.com/k0kubun/pp/v3"
	"github.com/moby/buildkit/client"
	gateway "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/progress/progresswriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type BuildOptions struct {
	DockerfilePath string
	ImageName      string
	Target         string
	BuildArgs      map[string]string
	ContextDir     string
}

func executeBuildImages(ctx context.Context, taskID string, payload map[string]interface{}) error {
	return buildAlgos()
}
func getAllSubDirectories(root string) ([]string, error) {
	var directories []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != root {
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			directories = append(directories, absPath)
		}
		return nil
	})

	return directories, err
}

func buildAlgos() error {
	pwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	parentDir := filepath.Dir(pwd)
	algos, err := getAllSubDirectories(filepath.Join(parentDir, "algorithms"))
	if err != nil {
		return err
	}

	var eg errgroup.Group
	for _, algo := range algos {
		a := algo
		eg.Go(func() error {
			return buildAlgo(a, nil, a)
		})
	}
	return eg.Wait()
}

func buildAlgo(algopath string, args map[string]string, ctxDir string) error {
	logrus.Infof("building algo %s...", algopath)
	algoName := filepath.Base(algopath)
	t := time.Now().UnixNano()
	err := buildDockerfileAndPush(context.Background(), BuildOptions{
		DockerfilePath: fmt.Sprintf("%s/builder.Dockerfile", algopath), // todo: remove the absolute path
		ImageName:      fmt.Sprintf("%s/%s:%d", con.GetString("harbor.repository"), algoName, t),
		BuildArgs:      args,
		ContextDir:     ctxDir,
		Target:         "",
	})
	return errors.Wrap(err, fmt.Sprintf("build algo %s failed", algopath))
}

func buildDockerfileAndPush(
	ctx context.Context,
	options BuildOptions,
) error {
	c, err := client.New(ctx, con.GetString("buildkitd_address"))
	if err != nil {
		return errors.Wrapf(err, "could not connect to buildkitd at %s", con.GetString("buildkitd_address"))
	}
	defer c.Close()

	dockerConfig := config.LoadDefaultConfigFile(os.Stderr)
	attachable := []session.Attachable{
		authprovider.NewDockerAuthProvider(dockerConfig, nil),
	}

	exports := []client.ExportEntry{
		{
			Type: "image",
			Attrs: map[string]string{
				"name": options.ImageName,
				"push": "true",
			},
		},
	}

	frontendAttrs := map[string]string{
		"filename": filepath.Base(options.DockerfilePath),
	}
	if options.Target != "" {
		frontendAttrs["target"] = options.Target
	}
	for k, v := range options.BuildArgs {
		frontendAttrs[fmt.Sprintf("build-arg:%s", k)] = v
	}

	solveOpt := client.SolveOpt{
		Exports:       exports,
		Session:       attachable,
		Ref:           identity.NewID(),
		Frontend:      "dockerfile.v0",
		FrontendAttrs: frontendAttrs,
		LocalDirs: map[string]string{
			"dockerfile": filepath.Dir(options.DockerfilePath),
			"context":    options.ContextDir,
		},
	}
	pp.Println(solveOpt)

	pw, err := progresswriter.NewPrinter(ctx, os.Stderr, "plain")
	if err != nil {
		return err
	}

	eg, ctx2 := errgroup.WithContext(ctx)
	eg.Go(func() error {
		_, err := c.Build(ctx2, solveOpt, "buildctl-dockerfile",
			func(ctx context.Context, gwClient gateway.Client) (*gateway.Result, error) {
				return nil, nil
			},
			pw.Status(),
		)
		return err
	})
	eg.Go(func() error {
		<-pw.Done()
		return pw.Err()
	})

	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "failed to build and push image")
	}

	return nil
}
