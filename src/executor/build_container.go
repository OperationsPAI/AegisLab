package executor

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"

	"aegis/config"
	"aegis/consts"
	"aegis/database"
	"aegis/dto"
	"aegis/repository"
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
	containerType consts.ContainerType
	name          string
	image         string
	tag           string
	command       string
	envVars       string
	sourcePath    string
	buildOptions  dto.BuildOptions
}

func rescheduleBuildTask(ctx context.Context, task *dto.UnifiedTask, reason string) error {
	span := trace.SpanFromContext(ctx)

	var executeTime time.Time

	minDelayMinutes := 1
	maxDelayMinutes := 5
	randomDelayMinutes := minDelayMinutes + rand.Intn(maxDelayMinutes-minDelayMinutes+1)
	executeTime = time.Now().Add(time.Duration(randomDelayMinutes) * time.Minute)

	eventPayload := executeTime.String()

	span.AddEvent(fmt.Sprintf("rescheduling build task: %s", reason))
	logrus.WithFields(logrus.Fields{
		"task_id":     task.TaskID,
		"trace_id":    task.TraceID,
		"delay_mins":  randomDelayMinutes,
		"retry_count": task.ReStartNum + 1,
	}).Warnf("%s: scheduled for %s", reason, executeTime.Format("2006-01-02 15:04:05"))

	tracing.SetSpanAttribute(ctx, consts.TaskStatusKey, string(consts.TaskStatusPending))

	repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
		TaskID:    task.TaskID,
		TaskType:  consts.TaskTypeBuildImage,
		EventName: consts.EventNoTokenAvailable,
		Payload:   eventPayload,
	})

	updateTaskStatus(
		ctx,
		task.TraceID,
		task.TaskID,
		reason,
		consts.TaskStautsRescheduled,
		consts.TaskTypeBuildImage,
	)

	if _, _, err := SubmitTask(ctx, &dto.UnifiedTask{
		TaskID:       task.TaskID,
		Type:         consts.TaskTypeBuildImage,
		Immediate:    false,
		ExecuteTime:  executeTime.Unix(),
		ReStartNum:   task.ReStartNum + 1,
		Payload:      task.Payload,
		Status:       consts.TaskStautsRescheduled,
		TraceID:      task.TraceID,
		GroupID:      task.GroupID,
		ProjectID:    task.ProjectID,
		TraceCarrier: task.TraceCarrier,
	}); err != nil {
		span.RecordError(err)
		span.AddEvent("failed to submit rescheduled task")
		return fmt.Errorf("failed to submit rescheduled build task: %v", err)
	}

	return nil
}

func executeBuildImage(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)
		span.AddEvent(fmt.Sprintf("Starting build attempt %d", task.ReStartNum+1))

		rateLimiter := utils.NewBuildContainerRateLimiter()

		acquired, err := rateLimiter.AcquireToken(childCtx, task.TaskID, task.TraceID)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to acquire rate limit token")
			return fmt.Errorf("failed to acquire rate limit token: %v", err)
		}

		if !acquired {
			span.AddEvent("no token available, waiting")
			logrus.WithFields(logrus.Fields{
				"task_id":  task.TaskID,
				"trace_id": task.TraceID,
			}).Info("No build container token available, waiting...")

			acquired, err = rateLimiter.WaitForToken(childCtx, task.TaskID, task.TraceID)
			if err != nil {
				span.RecordError(err)
				span.AddEvent("failed to wait for token")
				return fmt.Errorf("failed to wait for token: %v", err)
			}

			if !acquired {
				if err := rescheduleBuildTask(childCtx, task, "failed to acquire build token within timeout, retrying later"); err != nil {
					return err
				}
				return nil
			}
		}

		var tokenAcquired = true
		defer func() {
			if tokenAcquired {
				if releaseErr := rateLimiter.ReleaseToken(ctx, task.TaskID, task.TraceID); releaseErr != nil {
					logrus.WithFields(logrus.Fields{
						"task_id":  task.TaskID,
						"trace_id": task.TraceID,
						"error":    releaseErr,
					}).Error("Failed to release build container token")
				}
			}
		}()

		payload, err := parseBuildPayload(task.Payload)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to parse build payload")
			return err
		}

		updateTaskStatus(
			childCtx,
			task.TraceID,
			task.TaskID,
			fmt.Sprintf("building image for task %s", task.TaskID),
			consts.TaskStatusRunning,
			task.Type,
		)

		if err := buildImagendPush(childCtx, payload); err != nil {
			return err
		}

		if err := repository.CreateContainer(&database.Container{
			Type:    string(payload.containerType),
			Name:    payload.name,
			Image:   payload.image,
			Tag:     payload.tag,
			Command: payload.command,
			EnvVars: payload.envVars,
		}); err != nil {
			span.RecordError(err)
			span.AddEvent("failed to create container record")
			return err
		}

		repository.PublishEvent(childCtx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  task.Type,
			EventName: consts.EventImageBuildSucceed,
			Payload: dto.AlgorithmItem{
				Name:  payload.name,
				Image: payload.image,
				Tag:   payload.tag,
			},
		})

		updateTaskStatus(
			childCtx,
			task.TraceID,
			task.TaskID,
			fmt.Sprintf(consts.TaskMsgCompleted, task.TaskID),
			consts.TaskStatusCompleted,
			task.Type,
		)

		if err := os.RemoveAll(payload.sourcePath); err != nil {
			logrus.WithField("source_path", payload.sourcePath).Warnf("failed to remove source path after build: %v", err)
		}

		logrus.WithField("task_id", task.TaskID).Info("Build image task completed successfully")
		return nil
	})
}

func parseBuildPayload(payload map[string]any) (*containerPayload, error) {
	message := "missing or invalid '%s' key in payload"

	containerType, ok := payload[consts.BuildContainerType].(string)
	if !ok || containerType == "" {
		return nil, fmt.Errorf(message, consts.BuildContainerType)
	}

	name, ok := payload[consts.BuildName].(string)
	if !ok || name == "" {
		return nil, fmt.Errorf(message, consts.BuildName)
	}

	image, ok := payload[consts.BuildImage].(string)
	if !ok || image == "" {
		return nil, fmt.Errorf(message, consts.BuildImage)
	}

	tag, ok := payload[consts.BuildTag].(string)
	if !ok || tag == "" {
		return nil, fmt.Errorf(message, consts.BuildTag)
	}

	command, ok := payload[consts.BuildCommand].(string)
	if !ok || command == "" {
		return nil, fmt.Errorf(message, consts.BuildCommand)
	}

	envVarsArray, err := utils.ConvertToType[[]string](payload[consts.BuildImageEnvVars])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to []string: %v)", consts.BuildImageEnvVars, err)
	}

	sourcePath, ok := payload[consts.BuildSourcePath].(string)
	if !ok || sourcePath == "" {
		return nil, fmt.Errorf(message, consts.BuildSourcePath)
	}

	buildOptions, err := utils.ConvertToType[dto.BuildOptions](payload[consts.BuildBuildOptions])
	if err != nil {
		return nil, fmt.Errorf("failed to convert '%s' to BuildOptions: %v", consts.BuildBuildOptions, err)
	}

	return &containerPayload{
		containerType: consts.ContainerType(containerType),
		name:          name,
		image:         image,
		tag:           tag,
		command:       command,
		envVars:       strings.Join(envVarsArray, ","),
		sourcePath:    sourcePath,
		buildOptions:  buildOptions,
	}, nil
}

func buildImagendPush(ctx context.Context, payload *containerPayload) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(childCtx)

		address := fmt.Sprintf("tcp://%s", config.GetString("buildkit.address"))
		if address == "" {
			err := fmt.Errorf("buildkit address is not configured")
			span.RecordError(err)
			span.AddEvent("buildkit address is not configured")
			return err
		}

		c, err := buildkitclient.New(childCtx, address)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to create buildkit client")
			return err
		}

		defer c.Close()

		dockerConfig := con.LoadDefaultConfigFile(os.Stderr)
		attachable := []session.Attachable{
			authprovider.NewDockerAuthProvider(dockerConfig, nil),
		}

		fullImage := fmt.Sprintf("%s:%s", payload.image, payload.tag)
		exports := []buildkitclient.ExportEntry{
			{
				Type: buildkitclient.ExporterImage,
				Attrs: map[string]string{
					"name": fullImage,
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
			span.RecordError(err)
			span.AddEvent("failed to create local mount for context")
			return err
		}

		dockerfilePath := filepath.Join(payload.sourcePath, opts.DockerfilePath)
		dockerfileLocalMount, err := fsutil.NewFS(filepath.Dir(dockerfilePath))
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to create local mount for dockerfile")
			return err
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
			span.RecordError(err)
			span.AddEvent("failed to create progress writer")
			return err
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
				bklog.G(ctx).Errorf("build failed: %v", err)
				return err
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
			span.RecordError(err)
			span.AddEvent("failed to build and push image")
			return err
		}

		return nil
	})
}
