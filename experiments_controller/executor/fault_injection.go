package executor

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"regexp"
	"strconv"
	"time"

	chaos "github.com/LGU-SE-Internal/chaos-experiment/handler"
	"github.com/LGU-SE-Internal/rcabench/client"
	"github.com/LGU-SE-Internal/rcabench/client/k8s"
	"github.com/LGU-SE-Internal/rcabench/config"
	"github.com/LGU-SE-Internal/rcabench/consts"
	"github.com/LGU-SE-Internal/rcabench/database"
	"github.com/LGU-SE-Internal/rcabench/dto"
	"github.com/LGU-SE-Internal/rcabench/repository"
	"github.com/LGU-SE-Internal/rcabench/tracing"
	"github.com/LGU-SE-Internal/rcabench/utils"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/trace"
)

type nsConfig struct {
	Port      string `json:"port"`
	ChartName string `json:"chart_name"`
	ImageName string `json:"image_name"`
	ImageTag  string `json:"image_tag"`
	RepoName  string `json:"repo_name"`
	RepoURL   string `json:"repo_url"`
}

type injectionPayload struct {
	benchmark   string
	faultType   int
	namespace   string
	preDuration int
	displayData string
	conf        *chaos.InjectionConf
	node        *chaos.Node
	labels      []dto.LabelItem
}

type restartPayload struct {
	interval      int
	faultDuration int
	injectPayload map[string]any
}

type RestartServiceRateLimiter struct {
	redisClient *redis.Client
	maxTokens   int
	waitTimeout time.Duration
}

func NewRestartServiceRateLimiter() *RestartServiceRateLimiter {
	return &RestartServiceRateLimiter{
		redisClient: client.GetRedisClient(),
		maxTokens:   consts.MaxConcurrentRestarts,
		waitTimeout: time.Duration(consts.TokenWaitTimeout) * time.Second,
	}
}

func (r *RestartServiceRateLimiter) AcquireToken(ctx context.Context, taskID, traceID string) (bool, error) {
	span := trace.SpanFromContext(ctx)

	script := redis.NewScript(`
		local bucket_key = KEYS[1]
		local max_tokens = tonumber(ARGV[1])
		local task_id = ARGV[2]
		local trace_id = ARGV[3]
		local expire_time = tonumber(ARGV[4])
		
		-- 获取当前令牌数量
		local current_tokens = redis.call('SCARD', bucket_key)
		
		if current_tokens < max_tokens then
			-- 有可用令牌，添加任务ID到集合中
			redis.call('SADD', bucket_key, task_id)
			redis.call('EXPIRE', bucket_key, expire_time)
			return 1
		else
			return 0
		end
	`)

	// 令牌过期时间设置为10分钟，防止死锁
	expireTime := 10 * 60

	result, err := script.Run(ctx, r.redisClient, []string{consts.RestartServiceTokenBucket},
		r.maxTokens, taskID, traceID, expireTime).Result()
	if err != nil {
		span.RecordError(err)
		return false, fmt.Errorf("failed to acquire token: %v", err)
	}

	acquired := result.(int64) == 1
	if acquired {
		span.AddEvent("token acquired successfully")
		logrus.WithFields(logrus.Fields{
			"task_id":  taskID,
			"trace_id": traceID,
		}).Info("Successfully acquired restart service token")
	}

	return acquired, nil
}

// ReleaseToken 释放令牌
func (r *RestartServiceRateLimiter) ReleaseToken(ctx context.Context, taskID, traceID string) error {
	span := trace.SpanFromContext(ctx)

	// 从集合中移除任务ID
	result, err := r.redisClient.SRem(ctx, consts.RestartServiceTokenBucket, taskID).Result()
	if err != nil {
		span.RecordError(err)
		return fmt.Errorf("failed to release token: %v", err)
	}

	if result > 0 {
		span.AddEvent("token released successfully")
		logrus.WithFields(logrus.Fields{
			"task_id":  taskID,
			"trace_id": traceID,
		}).Info("Successfully released restart service token")
	}

	return nil
}

// WaitForToken 等待获取令牌，如果超时则返回 false
func (r *RestartServiceRateLimiter) WaitForToken(ctx context.Context, taskID, traceID string) (bool, error) {
	span := trace.SpanFromContext(ctx)
	span.AddEvent("waiting for token")

	timeoutCtx, cancel := context.WithTimeout(ctx, r.waitTimeout)
	defer cancel()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeoutCtx.Done():
			span.AddEvent("token wait timeout")
			logrus.WithFields(logrus.Fields{
				"task_id":  taskID,
				"trace_id": traceID,
				"timeout":  r.waitTimeout,
			}).Warn("Token wait timeout")
			return false, nil
		case <-ticker.C:
			acquired, err := r.AcquireToken(ctx, taskID, traceID)
			if err != nil {
				return false, err
			}
			if acquired {
				return true, nil
			}
		}
	}
}

func rescheduleTask(ctx context.Context, task *dto.UnifiedTask, reason string) error {
	span := trace.SpanFromContext(ctx)

	var executeTime time.Time

	randomFactor := 0.3 + rand.Float64()*0.7 // Random factor between 0.3 and 1.0
	deltaTime := time.Duration(math.Min(math.Pow(2, float64(task.ReStartNum)), 5.0)*randomFactor) * consts.DefaultTimeUnit
	executeTime = time.Now().Add(deltaTime)

	eventPayload := executeTime.String()

	span.AddEvent(fmt.Sprintf("rescheduling task: %s", reason))
	logrus.WithFields(logrus.Fields{
		"task_id":  task.TaskID,
		"trace_id": task.TraceID,
	}).Warnf("%s: %s", reason, executeTime)

	tracing.SetSpanAttribute(ctx, consts.TaskStatusKey, string(consts.TaskStatusPending))

	repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
		TaskID:    task.TaskID,
		TaskType:  consts.TaskTypeRestartService,
		EventName: consts.EventNoNamespaceAvailable,
		Payload:   eventPayload,
	})

	updateTaskStatus(
		ctx,
		task.TraceID,
		task.TaskID,
		reason,
		consts.TaskStautsRescheduled,
		consts.TaskTypeRestartService,
	)

	if _, _, err := SubmitTask(ctx, &dto.UnifiedTask{
		TaskID:       task.TaskID,
		Type:         consts.TaskTypeRestartService,
		Immediate:    false,
		ExecuteTime:  executeTime.Unix(),
		ReStartNum:   task.ReStartNum + 1,
		Payload:      task.Payload,
		Status:       consts.TaskStautsRescheduled,
		TraceID:      task.TraceID,
		GroupID:      task.GroupID,
		TraceCarrier: task.TraceCarrier,
	}); err != nil {
		span.RecordError(err)
		span.AddEvent("failed to submit rescheduled task")
		return fmt.Errorf("failed to submit rescheduled restart task: %v", err)
	}

	return nil
}

// 执行故障注入任务
func executeFaultInjection(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(ctx)

		payload, err := parseInjectionPayload(childCtx, task.Payload)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to parse injection payload")
			return err
		}

		monitor := k8s.GetMonitor()
		if err := monitor.CheckNamespaceToInject(payload.namespace, time.Now(), task.TraceID); err != nil {
			monitor.ReleaseLock(payload.namespace, task.TraceID)
			span.RecordError(fmt.Errorf("failed to get namespace to inject fault: %v", err))
			span.AddEvent("failed to get namespace to inject fault")
			return err
		}

		annotations, err := getAnnotations(childCtx, task)
		if err != nil {
			monitor.ReleaseLock(payload.namespace, task.TraceID)
			span.RecordError(err)
			span.AddEvent("failed to get annotations")
			return err
		}

		name, err := payload.conf.Create(
			childCtx,
			payload.namespace,
			annotations,
			map[string]string{
				consts.CRDTaskID:      task.TaskID,
				consts.CRDTraceID:     task.TraceID,
				consts.CRDGroupID:     task.GroupID,
				consts.CRDBenchmark:   payload.benchmark,
				consts.CRDPreDuration: strconv.Itoa(payload.preDuration),
			})
		if err != nil {
			monitor.ReleaseLock(payload.namespace, task.TraceID)
			span.RecordError(err)
			span.AddEvent("failed to inject fault")
			return fmt.Errorf("failed to inject fault: %v", err)
		}

		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeFaultInjection,
			EventName: consts.EventFaultInjectionStarted,
		})

		engineData, err := json.Marshal(payload.node)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to marshal injection spec to engine config")
			return fmt.Errorf("failed to marshal injection spec to engine config: %v", err)
		}

		faultRecord := database.FaultInjectionSchedule{
			TaskID:        task.TaskID,
			FaultType:     payload.faultType,
			DisplayConfig: payload.displayData,
			EngineConfig:  string(engineData),
			PreDuration:   payload.preDuration,
			Status:        consts.DatasetInitial,
			Description:   fmt.Sprintf("Fault for task %s", task.TaskID),
			Benchmark:     payload.benchmark,
			InjectionName: name,
			Labels:        make(database.LabelsMap),
		}

		for _, label := range payload.labels {
			faultRecord.Labels[label.Key] = label.Value
		}

		if err = database.DB.Create(&faultRecord).Error; err != nil {
			span.RecordError(err)
			span.AddEvent("failed to write fault injection schedule to database")
			logrus.Errorf("failed to write fault injection schedule to database: %v", err)
			return fmt.Errorf("failed to write to database")
		}

		return nil
	})
}

func executeRestartService(ctx context.Context, task *dto.UnifiedTask) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		span := trace.SpanFromContext(ctx)
		span.AddEvent(fmt.Sprintf("Starting retry attempt %d", task.ReStartNum+1))

		rateLimiter := NewRestartServiceRateLimiter()

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
			}).Info("No restart service token available, waiting...")

			acquired, err = rateLimiter.WaitForToken(childCtx, task.TaskID, task.TraceID)
			if err != nil {
				span.RecordError(err)
				span.AddEvent("failed to wait for token")
				return fmt.Errorf("failed to wait for token: %v", err)
			}

			if !acquired {
				if err := rescheduleTask(childCtx, task, "rate limited, retrying later"); err != nil {
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
					}).Error("Failed to release restart service token")
				}
			}
		}()

		payload, err := parseRestartPayload(childCtx, task.Payload)
		if err != nil {
			span.RecordError(err)
			span.AddEvent("failed to parse restart payload")
			return err
		}

		monitor := k8s.GetMonitor()

		t := time.Now()
		deltaTime := time.Duration(payload.interval) * consts.DefaultTimeUnit
		namespace := monitor.GetNamespaceToRestart(t.Add(deltaTime), task.TraceID)
		if namespace == "" {
			// 没有获取到 namespace 锁，立即释放限流令牌
			if releaseErr := rateLimiter.ReleaseToken(ctx, task.TaskID, task.TraceID); releaseErr != nil {
				logrus.WithFields(logrus.Fields{
					"task_id":  task.TaskID,
					"trace_id": task.TraceID,
					"error":    releaseErr,
				}).Error("Failed to release restart service token after namespace lock failure")
			}
			tokenAcquired = false // 标记令牌已释放，避免 defer 中重复释放

			if err := rescheduleTask(childCtx, task, "failed to acquire lock for namespace, retrying"); err != nil {
				return err
			}

			return nil
		}

		payload.injectPayload[consts.InjectNamespace] = namespace
		deltaTime = time.Duration(payload.interval-payload.faultDuration) * consts.DefaultTimeUnit
		injectTime := t.Add(deltaTime)

		nsPrefix, index, err := extractNamespace(namespace)
		if err != nil {
			monitor.ReleaseLock(namespace, task.TraceID)
			span.RecordError(err)
			span.AddEvent("failed to read namespace index")
			return fmt.Errorf("failed to read namespace index: %v", err)
		}

		updateTaskStatus(
			ctx,
			task.TraceID,
			task.TaskID,
			fmt.Sprintf("Restarting service in namespace %s", namespace),
			consts.TaskStatusRunning,
			consts.TaskTypeRestartService,
		)

		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeRestartService,
			EventName: consts.EventRestartServiceStarted,
		})

		if err := installTS(childCtx, namespace, nsPrefix, index); err != nil {
			monitor.ReleaseLock(namespace, task.TraceID)
			span.RecordError(err)
			span.AddEvent("failed to install Train Ticket")

			repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
				TaskID:    task.TaskID,
				TaskType:  consts.TaskTypeRestartService,
				EventName: consts.EventRestartServiceFailed,
				Payload:   err.Error(),
			})

			return err
		}

		message := fmt.Sprintf("Injection start at %s, duration %dm", injectTime.Local().String(), payload.faultDuration)
		repository.PublishEvent(ctx, fmt.Sprintf(consts.StreamLogKey, task.TraceID), dto.StreamEvent{
			TaskID:    task.TaskID,
			TaskType:  consts.TaskTypeRestartService,
			EventName: consts.EventRestartServiceCompleted,
			Payload:   message,
		})

		updateTaskStatus(
			ctx,
			task.TraceID,
			task.TaskID,
			message,
			consts.TaskStatusCompleted,
			consts.TaskTypeRestartService,
		)

		tracing.SetSpanAttribute(ctx, consts.TaskStatusKey, string(consts.TaskStatusCompleted))

		injectTask := &dto.UnifiedTask{
			Type:         consts.TaskTypeFaultInjection,
			Payload:      payload.injectPayload,
			Immediate:    false,
			ExecuteTime:  injectTime.Unix(),
			TraceID:      task.TraceID,
			GroupID:      task.GroupID,
			TraceCarrier: task.TraceCarrier,
		}
		if _, _, err := SubmitTask(childCtx, injectTask); err != nil {
			monitor.ReleaseLock(namespace, task.TraceID)
			span.RecordError(err)
			span.AddEvent("failed to submit inject task")
			return fmt.Errorf("failed to submit inject task: %v", err)
		}

		return nil
	})
}

func installTS(ctx context.Context, namespace, nsPrefix string, namespaceIdx int) error {
	return tracing.WithSpan(ctx, func(childCtx context.Context) error {
		nsConfigMap, err := config.GetNsConfigMap()
		if err != nil {
			return fmt.Errorf("failed to get namespace config map: %v", err)
		}

		payload, exists := nsConfigMap[nsPrefix]
		if !exists {
			return fmt.Errorf("namespace %s not found in config map", nsPrefix)
		}

		nsConfig, err := utils.ConvertToType[nsConfig](payload)
		if err != nil {
			return err
		}

		helmClient, err := client.NewHelmClient(namespace)
		if err != nil {
			return fmt.Errorf("failed to create Helm client: %v", err)
		}

		// Add Train Ticket repository
		if err := helmClient.AddRepo(nsConfig.RepoName, nsConfig.RepoURL); err != nil {
			return fmt.Errorf("failed to add repository: %v", err)
		}

		// Update repositories
		if err := helmClient.UpdateRepo(); err != nil {
			return fmt.Errorf("failed to update repositories: %v", err)
		}

		container, err := repository.GetContaineInfo(&dto.GetContainerFilterOptions{
			Type: consts.ContainerTypeNamespace,
			Name: nsPrefix,
		})
		if err != nil {
			return fmt.Errorf("failed to get container info: %v", err)
		}

		port := fmt.Sprintf(nsConfig.Port, namespaceIdx)
		chartName := fmt.Sprintf("%s/%s", nsConfig.RepoName, nsConfig.ChartName)
		if err := helmClient.InstallTrainTicket(ctx,
			namespace,
			chartName,
			container.Image,
			container.Tag,
			port,
			600*time.Second,
			360*time.Second,
		); err != nil {
			return fmt.Errorf("failed to install Train Ticket: %v", err)
		}

		logrus.Infof("Train Ticket installed successfully in namespace %s", namespace)
		return nil
	})
}

func parseInjectionPayload(ctx context.Context, payload map[string]any) (*injectionPayload, error) {
	return tracing.WithSpanReturnValue(ctx, func(childCtx context.Context) (*injectionPayload, error) {
		message := "invalid or missing '%s' in task payload"

		benchmark, ok := payload[consts.InjectBenchmark].(string)
		if !ok {
			return nil, fmt.Errorf(message, consts.InjectBenchmark)
		}

		faultTypeFloat, ok := payload[consts.InjectFaultType].(float64)
		if !ok || faultTypeFloat < 0 {
			return nil, fmt.Errorf(message, consts.InjectFaultType)
		}
		faultType := int(faultTypeFloat)

		namespace, ok := payload[consts.InjectNamespace].(string)
		if !ok || namespace == "" {
			return nil, fmt.Errorf(message, consts.InjectNamespace)
		}

		preDurationFloat, ok := payload[consts.InjectPreDuration].(float64)
		if !ok || preDurationFloat <= 0 {
			return nil, fmt.Errorf(message, consts.InjectPreDuration)
		}
		preDuration := int(preDurationFloat)

		displayData, ok := payload[consts.InjectDisplayData].(string)
		if !ok || displayData == "" {
			return nil, fmt.Errorf(message, consts.InjectDisplayData)
		}

		conf, err := utils.MapToStruct[chaos.InjectionConf](payload, consts.InjectConf, message)
		if err != nil {
			return nil, err
		}

		node, err := utils.MapToStruct[chaos.Node](payload, consts.InjectNode, message)
		if err != nil {
			return nil, err
		}

		labels, err := utils.ConvertToType[[]dto.LabelItem](payload[consts.InjectLabels])
		if err != nil {
			return nil, fmt.Errorf(message, consts.InjectLabels)
		}

		return &injectionPayload{
			benchmark:   benchmark,
			faultType:   faultType,
			namespace:   namespace,
			preDuration: preDuration,
			displayData: displayData,
			conf:        conf,
			node:        node,
			labels:      labels,
		}, nil
	})
}

func parseRestartPayload(ctx context.Context, payload map[string]any) (*restartPayload, error) {
	return tracing.WithSpanReturnValue(ctx, func(childCtx context.Context) (*restartPayload, error) {
		message := "invalid or missing '%s' in task payload"

		intervalFloat, ok := payload[consts.RestartIntarval].(float64)
		if !ok || intervalFloat <= 0 {
			return nil, fmt.Errorf(message, consts.RestartIntarval)
		}
		interval := int(intervalFloat)

		faultDurationFloat, ok := payload[consts.RestartFaultDuration].(float64)
		if !ok || faultDurationFloat <= 0 {
			return nil, fmt.Errorf(message, consts.RestartFaultDuration)
		}
		faultDuration := int(faultDurationFloat)

		injectPayload, ok := payload[consts.RestartInjectPayload].(map[string]any)
		if !ok {
			return nil, fmt.Errorf(message, consts.RestartInjectPayload)
		}

		return &restartPayload{
			interval:      interval,
			faultDuration: faultDuration,
			injectPayload: injectPayload,
		}, nil
	})
}

func extractNamespace(namespace string) (string, int, error) {
	pattern := `^([a-zA-Z]+)(\d+)$`
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(namespace)

	if len(match) < 3 {
		return "", 0, fmt.Errorf("failed to extract index from namespace %s", namespace)
	}

	num, err := strconv.Atoi(match[2])
	if err != nil {
		return "", 0, fmt.Errorf("failed to convert extracted index to integer: %v", err)
	}

	return match[1], num, nil
}
