package executor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	con "github.com/CUHK-SE-Group/rcabench/config"

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

	dockerfileLocalMount, err := fsutil.NewFS(options.ContextDir)
	if err != nil {
		return errors.Wrap(err, "failed to create local mount for dockerfile")
	}
	cxtLocalMount, err := fsutil.NewFS(filepath.Dir(options.DockerfilePath))
	if err != nil {
		return errors.Wrap(err, "failed to create local mount for dockerfile")
	}
	solveOpt := client.SolveOpt{
		Exports:       exports,
		Session:       attachable,
		Ref:           identity.NewID(),
		Frontend:      "dockerfile.v0",
		FrontendAttrs: frontendAttrs,
		LocalMounts: map[string]fsutil.FS{
			"context":    cxtLocalMount,
			"dockerfile": dockerfileLocalMount,
		},
	}
	// traceFile, err := os.OpenFile("tracefile.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	// if err != nil {
	// return err
	// }
	// var traceEnc *json.Encoder
	// if traceFile != nil {
	// 	defer traceFile.Close()
	// 	traceEnc = json.NewEncoder(traceFile)

	// 	bklog.L.Infof("tracing logs to %s", traceFile.Name())
	// }
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
	// if traceEnc != nil {
	// 	traceCh := make(chan *client.SolveStatus)
	// 	pw = progresswriter.Tee(pw, traceCh)
	// 	eg.Go(func() error {
	// 		for s := range traceCh {
	// 			if err := traceEnc.Encode(s); err != nil {
	// 				return err
	// 			}
	// 		}
	// 		return nil
	// 	})
	// }
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
				logrus.Info("fuck")

				return res, err
			},
			progresswriter.ResetTime(mw.WithPrefix("", false)).Status(),
		)
		logrus.Info("build11 finished")
		if err != nil {
			bklog.G(ctx).Errorf("build failed: %v", err)
		}
		for k, v := range resp.ExporterResponse {
			bklog.G(ctx).Debugf("exporter response: %s=%s", k, v)
		}
		return err

	})

	eg.Go(func() error {
		<-pw.Done()
		logrus.Info("build finished")
		return pw.Err()
	})

	if err := eg.Wait(); err != nil {
		return errors.Wrap(err, "failed to build and push image")
	}

	return nil
}
