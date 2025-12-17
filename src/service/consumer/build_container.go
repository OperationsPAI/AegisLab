package consumer

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"time"

	"aegis/config"
	"aegis/consts"
	"aegis/dto"
	"aegis/service/common"
	"aegis/tracing"
	"aegis/utils"

	con "github.com/docker/cli/cli/config"
	buildkitclient "github.com/moby/buildkit/client"
	"github.com/moby/buildkit/frontend"
	gateway "github.com/moby/buildkit/frontend/gateway/client"
	"github.com/moby/buildkit/identity"
	"github.com/moby/buildkit/session"
	"github.com/moby/buildkit/session/auth/authprovider"
	"github.com/moby/buildkit/util/bklog"
	"github.com/moby/buildkit/util/progress/progressui"
	"github.com/moby/buildkit/util/progress/progresswriter"
	"github.com/sirupsen/logrus"
	"github.com/tonistiigi/fsutil"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/errgroup"
)

type containerPayload struct {
	imageRef     string
	sourcePath   string
	buildOptions dto.BuildOptions
}

// executeBuildContainer handles the execution of a build container task
func executeBuildContainer(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		span.AddEvent(fmt.Sprintf("Starting build attempt %d", task.ReStartNum+1))
		logEntry := logrus.WithFields(logrus.Fields{
			"task_id":  task.TaskID,
			"trace_id": task.TraceID,
		})

		rateLimiter := GetBuildContainerRateLimiter()
		acquired, err := rateLimiter.AcquireToken(childCtx, task.TaskID, task.TraceID)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to acquire rate limit token", err)
		}

		if !acquired {
			span.AddEvent("no token available, waiting")
			logEntry.Info("No build container token available, waiting...")

			acquired, err = rateLimiter.WaitForToken(childCtx, task.TaskID, task.TraceID)
			if err != nil {
				return handleExecutionError(span, logEntry, "failed to wait for token", err)
			}

			if !acquired {
				if err := rescheduleContainerBuildingTask(childCtx, task, "failed to acquire build token within timeout, retrying later"); err != nil {
					return err
				}
				return nil
			}
		}

		defer func() {
			if acquired {
				if releaseErr := rateLimiter.ReleaseToken(childCtx, task.TaskID, task.TraceID); releaseErr != nil {
					logEntry.Error("Failed to release build container token")
				}
			}
		}()

		payload, err := parseContainerPayload(task.Payload)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to parse build payload", err)
		}

		if err := buildImageAndPush(childCtx, payload, logEntry); err != nil {
			return err
		}

		updateTaskState(childCtx,
			newTaskStateUpdate(
				task.TraceID,
				task.TaskID,
				task.Type,
				consts.TaskCompleted,
				fmt.Sprintf("Container image %s built and pushed successfully", payload.imageRef),
			).withEvent(consts.EventImageBuildSucceed, payload.imageRef),
		)

		if err := os.RemoveAll(payload.sourcePath); err != nil {
			logrus.WithField("source_path", payload.sourcePath).Warnf("failed to remove source path after build: %v", err)
		}

		logrus.WithField("task_id", task.TaskID).Info("Container building task completed successfully")
		return nil
	})
}

// rescheduleContainerBuildingTask reschedules a container building task with a random delay between 1 to 5 minutes
func rescheduleContainerBuildingTask(ctx context.Context, task *dto.UnifiedTask, reason string) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)

		randomDelayMinutes := minDelayMinutes + rand.Intn(maxDelayMinutes-minDelayMinutes+1)
		executeTime := time.Now().Add(time.Duration(randomDelayMinutes) * time.Minute)

		span.AddEvent(fmt.Sprintf("rescheduling build task: %s", reason))
		logrus.WithFields(logrus.Fields{
			"task_id":     task.TaskID,
			"trace_id":    task.TraceID,
			"delay_mins":  randomDelayMinutes,
			"retry_count": task.ReStartNum + 1,
		}).Warnf("%s: scheduled for %s", reason, executeTime.Format(time.DateTime))

		tracing.SetSpanAttribute(childCtx, consts.TaskStateKey, consts.GetTaskStateName(consts.TaskRescheduled))

		updateTaskState(childCtx,
			newTaskStateUpdate(
				task.TraceID,
				task.TaskID,
				task.Type,
				consts.TaskRescheduled,
				reason,
			).withEvent(consts.EventNoTokenAvailable, executeTime.String()),
		)

		task.Reschedule(executeTime)
		if err := common.SubmitTask(childCtx, task); err != nil {
			span.RecordError(err)
			span.AddEvent("failed to submit rescheduled task")
			return fmt.Errorf("failed to submit rescheduled container building task: %v", err)
		}

		return nil
	})
}

// parseContainerPayload extracts and validates the container build payload from the task
func parseContainerPayload(payload map[string]any) (*containerPayload, error) {
	message := "missing or invalid '%s' key in payload"

	imageRef, ok := payload[consts.BuildImageRef].(string)
	if !ok || imageRef == "" {
		return nil, fmt.Errorf(message, consts.BuildImageRef)
	}

	sourcePath, ok := payload[consts.BuildSourcePath].(string)
	if !ok || sourcePath == "" {
		return nil, fmt.Errorf(message, consts.BuildSourcePath)
	}

	buildOptions, err := utils.ConvertToType[dto.BuildOptions](payload[consts.BuildBuildOptions])
	if err != nil {
		return nil, fmt.Errorf("invalid build configuration in payload: %v", err)
	}

	return &containerPayload{
		imageRef:     imageRef,
		sourcePath:   sourcePath,
		buildOptions: buildOptions,
	}, nil
}

// buildImageAndPush builds the container image using BuildKit and pushes it to the registry
func buildImageAndPush(ctx context.Context, payload *containerPayload, logEntry *logrus.Entry) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)

		address := fmt.Sprintf("tcp://%s", config.GetString("buildkit.address"))
		if address == "" {
			err := fmt.Errorf("buildkit address is not configured")
			return handleExecutionError(span, logEntry, "buildkit address is not configured", err)
		}

		c, err := buildkitclient.New(childCtx, address)
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to create buildkit client", err)
		}

		defer c.Close()

		dockerConfig := con.LoadDefaultConfigFile(os.Stderr)
		attachable := []session.Attachable{
			authprovider.NewDockerAuthProvider(dockerConfig, nil),
		}

		exports := []buildkitclient.ExportEntry{
			{
				Type: buildkitclient.ExporterImage,
				Attrs: map[string]string{
					"name": payload.imageRef,
					"push": "true",
				},
			},
		}

		opts := payload.buildOptions
		frontendAttrs := map[string]string{
			"filename": filepath.Base(opts.DockerfilePath),
		}
		if opts.Target != "" {
			frontendAttrs["target"] = opts.Target
		}

		if opts.BuildArgs != nil {
			for k, v := range opts.BuildArgs {
				frontendAttrs[fmt.Sprintf("build-arg:%s", k)] = v
			}
		}

		ctxLocalMount, err := fsutil.NewFS(filepath.Join(payload.sourcePath, opts.ContextDir))
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to create local mount for context", err)
		}

		dockerfilePath := filepath.Join(payload.sourcePath, opts.DockerfilePath)
		dockerfileLocalMount, err := fsutil.NewFS(filepath.Dir(dockerfilePath))
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to create local mount for dockerfile", err)
		}

		solveOpt := buildkitclient.SolveOpt{
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
		pw, err := progresswriter.NewPrinter(childCtx, os.Stderr, string(progressui.AutoMode))
		if err != nil {
			return handleExecutionError(span, logEntry, "failed to create progress writer", err)
		}

		mw := progresswriter.NewMultiWriter(pw)
		var writers []progresswriter.Writer
		for _, at := range attachable {
			if s, ok := at.(interface {
				SetLogger(progresswriter.Logger)
			}); ok {
				w := mw.WithPrefix("", false)
				s.SetLogger(func(s *buildkitclient.SolveStatus) {
					w.Status() <- s
				})
				writers = append(writers, w)
			}
		}

		eg, ctx2 := errgroup.WithContext(childCtx)
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
				bklog.G(childCtx).Errorf("build failed: %v", err)
				return err
			}

			for k, v := range resp.ExporterResponse {
				bklog.G(childCtx).Debugf("exporter response: %s=%s", k, v)
			}
			return err
		})

		eg.Go(func() error {
			<-pw.Done()
			logrus.Info("Build finished")
			return pw.Err()
		})

		if err := eg.Wait(); err != nil {
			return handleExecutionError(span, logEntry, "failed to build and push image", err)
		}

		return nil
	})
}
