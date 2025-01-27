package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	con "github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/utils"

	"github.com/docker/cli/cli/config"
	"github.com/moby/buildkit/client"
	"github.com/moby/buildkit/frontend"
	gateway "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/bklog"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/moby/buildkit/util/progress/progresswriter"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/tonistiigi/fsutil"
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

func buildAlgos() error {
	wd := con.GetString("workspace")
	algos, err := utils.GetAllSubDirectories(filepath.Join(wd, "algorithms"))
	if err != nil {
		return err
	}

	var eg errgroup.Group
	for _, algo := range algos {
		eg.Go(func() error {
			return buildAlgo(algo, nil, wd)
		})
	}
	return eg.Wait()
}

func buildAlgo(algoDir string, args map[string]string, ctxDir string) error {
	logrus.Infof("building algo %s...", algoDir)
	algoName := filepath.Base(algoDir)
	t := time.Now().UnixNano()

	err := buildDockerfileAndPush(context.Background(), BuildOptions{
		DockerfilePath: filepath.Join(algoDir, "builder.Dockerfile"), // todo: remove the absolute path
		ImageName:      fmt.Sprintf("%s/%s:%d", con.GetString("harbor.repository"), algoName, t),
		BuildArgs:      args,
		ContextDir:     ctxDir,
		Target:         "",
	})

	return errors.Wrap(err, fmt.Sprintf("Build algo %s failed", algoDir))
}

func buildDockerfileAndPush(ctx context.Context, options BuildOptions) error {
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
			Type: client.ExporterImage,
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

	ctxLocalMount, err := fsutil.NewFS(options.ContextDir)
	if err != nil {
		return errors.Wrap(err, "Failed to create local mount for context")
	}
	dockerfileLocalMount, err := fsutil.NewFS(filepath.Dir(options.DockerfilePath))
	if err != nil {
		return errors.Wrap(err, "Failed to create local mount for dockerfile")
	}
	solveOpt := client.SolveOpt{
		Exports:       exports,
		Session:       attachable,
		Ref:           identity.NewID(),
		Frontend:      "dockerfile.v0",
		FrontendAttrs: frontendAttrs,
		LocalMounts: map[string]fsutil.FS{
			"context":    ctxLocalMount,
			"dockerfile": dockerfileLocalMount,
		},
	}
	pw, err := progresswriter.NewPrinter(context.TODO(), os.Stderr, string(progressui.AutoMode))
	if err != nil {
		return err
	}
	mw := progresswriter.NewMultiWriter(pw)

	var writers []progresswriter.Writer
	for _, at := range attachable {
		if s, ok := at.(interface {
			SetLogger(progresswriter.Logger)
		}); ok {
			w := mw.WithPrefix("", false)
			s.SetLogger(func(s *client.SolveStatus) {
				w.Status() <- s
			})
			writers = append(writers, w)
		}
	}

	eg, ctx2 := errgroup.WithContext(ctx)
	eg.Go(func() error {
		defer func() {
			for _, w := range writers {
				close(w.Status())
			}
		}()
		sreq := gateway.SolveRequest{
			Frontend:    solveOpt.Frontend,
			FrontendOpt: solveOpt.FrontendAttrs,
		}
		sreq.CacheImports = make([]frontend.CacheOptionsEntry, len(solveOpt.CacheImports))
		for i, e := range solveOpt.CacheImports {
			sreq.CacheImports[i] = frontend.CacheOptionsEntry{
				Type:  e.Type,
				Attrs: e.Attrs,
			}
		}

		resp, err := c.Build(ctx2, solveOpt, "buildctl",
			func(ctx context.Context, c gateway.Client) (*gateway.Result, error) {
				logrus.Info("begin to solve")
				res, err := c.Solve(ctx, sreq)

				return res, err
			},
			progresswriter.ResetTime(mw.WithPrefix("", false)).Status(),
		)
		logrus.Info("Build finished")
		if err != nil {
			bklog.G(ctx).Errorf("Build failed: %v", err)
		}
		for k, v := range resp.ExporterResponse {
			bklog.G(ctx).Debugf("exporter response: %s=%s", k, v)
		}
		return err

	})

	eg.Go(func() error {
		<-pw.Done()
		logrus.Info("Build finished")
		return pw.Err()
	})

	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "Failed to build and push image")
	}

	return nil
}
