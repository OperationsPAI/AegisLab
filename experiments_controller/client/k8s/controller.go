package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"slices"

	chaosCli "github.com/CUHK-SE-Group/chaos-experiment/client"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type ResouceType string

const (
	jobKind      = "Job"
	resyncPeriod = 10 * time.Second

	CRDResourceType ResouceType = "CRD"
	JobResourceType ResouceType = "Job"
)

type JobEnv struct {
	Namespace   string
	Service     string
	PreDuration int
	StartTime   time.Time
	EndTime     time.Time
}

type timeRange struct {
	Start time.Time
	End   time.Time
}

// 接口避免循环引用
type Callback interface {
	HandleCRDFailed(name, errorMsg string, labels map[string]string)
	HandleCRDSucceeded(namespace, pod, name string, startTime, endTime time.Time, labels map[string]string) error
	HandleJobAdd(labels map[string]string) error
	HandleJobFailed(labels map[string]string, errorMsg string) error
	HandleJobSucceeded(labels map[string]string) error
}

type QueueItem struct {
	Type      ResouceType
	Namespace string
	Name      string
	GVR       *schema.GroupVersionResource
}

type Controller struct {
	crdInformers map[schema.GroupVersionResource]cache.SharedIndexInformer
	jobInformer  cache.SharedIndexInformer
	podInformer  cache.SharedIndexInformer

	queue workqueue.TypedRateLimitingInterface[QueueItem]
}

func NewController(restConfig *rest.Config) *Controller {
	factory := informers.NewSharedInformerFactoryWithOptions(
		k8sClient,
		resyncPeriod,
		informers.WithNamespace(config.GetString("k8s.namespace")),
	)

	chaosGVRs := make([]schema.GroupVersionResource, 0, len(chaosCli.GetCRDMapping()))
	for gvr := range chaosCli.GetCRDMapping() {
		chaosGVRs = append(chaosGVRs, gvr)
	}

	// 初始化所有 CRD Informer
	chaosFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		k8sDynamicClient,
		resyncPeriod,
		metav1.NamespaceAll,
		nil,
	)

	crdInformers := make(map[schema.GroupVersionResource]cache.SharedIndexInformer)
	for _, gvr := range chaosGVRs {
		crdInformers[gvr] = chaosFactory.ForResource(gvr).Informer()
	}

	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[QueueItem]())

	return &Controller{
		crdInformers: crdInformers,
		jobInformer:  factory.Batch().V1().Jobs().Informer(),
		podInformer:  factory.Core().V1().Pods().Informer(),
		queue:        queue,
	}
}

func (c *Controller) Run(ctx context.Context, callback Callback) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	c.registerEventHandlers(callback)

	logrus.Info("Starting informer controller")
	for _, informer := range c.crdInformers {
		go informer.Run(ctx.Done())
	}
	go c.jobInformer.Run(ctx.Done())
	go c.podInformer.Run(ctx.Done())

	allSyncs := []cache.InformerSynced{c.jobInformer.HasSynced, c.podInformer.HasSynced}
	for _, informer := range c.crdInformers {
		allSyncs = append(allSyncs, informer.HasSynced)
	}

	if !cache.WaitForCacheSync(ctx.Done(), allSyncs...) {
		message := "timed out waiting for caches to sync"
		runtime.HandleError(fmt.Errorf("%s", message))
		logrus.Error(message)
		return
	}

	go wait.Until(c.runWorker, time.Second, ctx.Done())

	<-ctx.Done()
	logrus.Info("Stopping informer controller...")
}

func (c *Controller) registerEventHandlers(callback Callback) {
	for gvr, informer := range c.crdInformers {
		if _, err := informer.AddEventHandler(c.genCRDEventHandlerFuncs(gvr, callback)); err != nil {
			logrus.WithField("gvr", gvr.Resource).Error("failed to add event handler")
			return
		}
	}

	if _, err := c.jobInformer.AddEventHandler(c.genJobEventHandlerFuncs(callback)); err != nil {
		logrus.WithField("func", "genJobEventHandlerFuncs").Error("failed to add event handler")
		return
	}

	if _, err := c.podInformer.AddEventHandler(c.genPodEventHandlerFuncs()); err != nil {
		logrus.WithField("func", "genPodEventHandlerFuncs").Error("failed to add event handler")
		return
	}
}

func (c *Controller) genCRDEventHandlerFuncs(gvr schema.GroupVersionResource, callback Callback) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			u := obj.(*unstructured.Unstructured)
			if isTargetNamespace(u.GetNamespace()) {
				logrus.WithFields(logrus.Fields{
					"type":      gvr.Resource,
					"namespace": u.GetNamespace(),
					"name":      u.GetName(),
				}).Info("chaos experiment created successfully")
			}
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldU := oldObj.(*unstructured.Unstructured)
			newU := newObj.(*unstructured.Unstructured)

			if callback != nil && oldU.GetName() == newU.GetName() && isTargetNamespace(newU.GetNamespace()) {
				logEntry := logrus.WithFields(logrus.Fields{
					"type":      gvr.Resource,
					"namespace": newU.GetNamespace(),
					"name":      newU.GetName(),
				})

				labels, _, _ := unstructured.NestedStringMap(newU.Object, "metadata", "labels")

				oldPhase, _, _ := unstructured.NestedString(oldU.Object, "status", "experiment", "desiredPhase")
				newPhase, _, _ := unstructured.NestedString(newU.Object, "status", "experiment", "desiredPhase")
				if oldPhase == "Run" && newPhase == "Stop" {
					conditions, _, _ := unstructured.NestedSlice(newU.Object, "status", "conditions")

					selected := getCRDConditionStatus(conditions, "Selected")
					if !selected {
						message := "failed to select app in the chaos experiment"
						logEntry.Error(message)
						callback.HandleCRDFailed(newU.GetName(), message, labels)
						return
					}

					// 会花费 duration 的时间去尝试注入
					allInjected := getCRDConditionStatus(conditions, "AllInjected")
					if !allInjected {
						message := "failed to inject all targets in the chaos experiment"
						logEntry.Error(message)
						callback.HandleCRDFailed(newU.GetName(), message, labels)
						return
					}
				}

				oldConditions, _, _ := unstructured.NestedSlice(oldU.Object, "status", "conditions")
				newConditions, _, _ := unstructured.NestedSlice(newU.Object, "status", "conditions")

				// 判断是否注入
				oldAllInjected := getCRDConditionStatus(oldConditions, "AllInjected")
				newAllInjected := getCRDConditionStatus(newConditions, "AllInjected")
				if !oldAllInjected && newAllInjected {
					logEntry.Infof("all targets injected in the chaos experiment")
					durationStr, _, _ := unstructured.NestedString(newU.Object, "spec", "duration")

					pattern := `(\d+)m`
					re := regexp.MustCompile(pattern)
					match := re.FindStringSubmatch(durationStr)
					if len(match) <= 1 {
						message := "failed to get the duration"
						logEntry.Error(message)
						callback.HandleCRDFailed(newU.GetName(), message, labels)
						return
					}

					duration, err := strconv.Atoi(match[1])
					if err != nil {
						message := "failed to get the duration of the chaos experiement"
						logEntry.Error(message)
						callback.HandleCRDFailed(newU.GetName(), message, labels)
						return
					}

					// 计时协程去判断是否恢复成功
					if duration > 0 {
						go func(namespace, name string, injectDuration int) {
							timer := time.NewTimer(60 * time.Duration(injectDuration) * time.Second)
							<-timer.C

							obj, err := k8sDynamicClient.Resource(gvr).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
							if err != nil {
								if errors.IsNotFound(err) {
									logEntry.Info("the chaos experiment has been deleted")
									return
								}

								message := "failed to get the CRD resource object"
								logEntry.Errorf("%s: %v", message, err)
								callback.HandleCRDFailed(newU.GetName(), message, labels)
								return
							}

							conditions, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
							recovered := getCRDConditionStatus(conditions, "AllRecovered")

							if !recovered {
								message := "faied to recover all targets in the chaos experiment"
								logEntry.Error(message)
								callback.HandleCRDFailed(newU.GetName(), message, labels)
							}
						}(newU.GetNamespace(), newU.GetName(), duration)
					}
				}

				oldAllRecovered := getCRDConditionStatus(oldConditions, "AllRecovered")
				newAllRecovered := getCRDConditionStatus(newConditions, "AllRecovered")
				if !oldAllRecovered && newAllRecovered {
					logEntry.Infof("all targets recoverd in the chaos experiment")

					pod, _, _ := unstructured.NestedString(newU.Object, "spec", "selector", "labelSelectors", "app")

					chaosGVRMapping := make(map[string]schema.GroupVersionResource)
					for gvr, obj := range chaosCli.GetCRDMapping() {
						chaosGVRMapping[utils.GetTypeName(obj)] = gvr
					}

					kind := newU.GetKind()
					gvr, ok := chaosGVRMapping[kind]
					if !ok {
						message := "failed to get the CRD resource gvr"
						logEntry.Error(message)
						callback.HandleCRDFailed(newU.GetName(), message, labels)
						return
					}

					newRecords, _, _ := unstructured.NestedSlice(newU.Object, "status", "experiment", "containerRecords")
					timeRanges := getCRDEventTimeRanges(newRecords)
					if len(timeRanges) == 0 {
						message := "failed to get the start_time and end_time"
						logEntry.Error(message)
						callback.HandleCRDFailed(newU.GetName(), message, labels)
						return
					}

					timeRange := timeRanges[0]
					if err := callback.HandleCRDSucceeded(newU.GetNamespace(), pod, newU.GetName(), timeRange.Start, timeRange.End, labels); err != nil {
						logEntry.Error(err)
						callback.HandleCRDFailed(newU.GetName(), err.Error(), labels)
						return
					}

					if !config.GetBool("debugging.enable") {
						c.queue.Add(QueueItem{
							Type:      CRDResourceType,
							Namespace: newU.GetNamespace(),
							Name:      newU.GetName(),
							GVR:       &gvr,
						})
					}
				}
			}
		},
		DeleteFunc: func(obj any) {
			u := obj.(*unstructured.Unstructured)
			if isTargetNamespace(u.GetNamespace()) {
				logrus.WithFields(logrus.Fields{
					"type":      gvr.Resource,
					"namespace": u.GetNamespace(),
					"name":      u.GetName(),
				}).Info("Chaos experiment deleted successfully")
			}
		},
	}
}

func (c *Controller) genJobEventHandlerFuncs(callback Callback) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			job := obj.(*batchv1.Job)
			logrus.WithField("namespace", job.Namespace).WithField("job_name", job.Name).Info("Job created successfully")
			callback.HandleJobAdd(job.Labels)
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldJob := oldObj.(*batchv1.Job)
			newJob := newObj.(*batchv1.Job)

			if callback != nil && oldJob.Name == newJob.Name {
				if oldJob.Status.Failed == 0 && newJob.Status.Failed > 0 {
					errorMsg := extractJobError(newJob)
					callback.HandleJobFailed(newJob.Labels, errorMsg)
				}

				if oldJob.Status.Succeeded == 0 && newJob.Status.Succeeded > 0 {
					callback.HandleJobSucceeded(newJob.Labels)
					if !config.GetBool("debugging.enable") {
						c.queue.Add(QueueItem{
							Type:      JobResourceType,
							Namespace: newJob.Namespace,
							Name:      newJob.Name,
						})
					}
				}
			}
		},
		DeleteFunc: func(obj any) {
			job := obj.(*batchv1.Job)
			logrus.WithField("namespace", job.Namespace).WithField("job_name", job.Name).Infof("job delete successfully")
		},
	}
}

func (c *Controller) genPodEventHandlerFuncs() cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj any) {
			newPod := newObj.(*corev1.Pod)

			if newPod.Status.Phase == corev1.PodPending {
				podReasons := []string{"ImagePullBackOff"}
				ownerRefs := newPod.OwnerReferences
				var jobOwnerRef *metav1.OwnerReference
				for _, ref := range ownerRefs {
					if ref.Kind == jobKind {
						jobOwnerRef = &ref
						break
					}
				}

				if jobOwnerRef == nil {
					return
				}

				for _, reason := range podReasons {
					if checkPodReason(newPod, reason) {
						job, err := GetJob(context.TODO(), newPod.Namespace, jobOwnerRef.Name)
						if err != nil {
							logrus.WithField("job_name", jobOwnerRef.Name).Error(err)
						}

						if job != nil {
							handlePodError(newPod, job, reason)
							if !config.GetBool("debugging.enable") {
								c.queue.Add(QueueItem{
									Type:      JobResourceType,
									Namespace: job.Namespace,
									Name:      job.Name,
								})
							}

							break
						}
					}
				}
			}
		},
		DeleteFunc: func(obj any) {
			pod := obj.(*corev1.Pod)
			logrus.WithField("namespace", pod.Namespace).WithField("pod_name", pod.Name).Infof("pod delete successfully")
		},
	}
}

func (c *Controller) runWorker() {
	for c.processQueueItem() {
	}
}

func (c *Controller) processQueueItem() bool {
	item, quit := c.queue.Get()
	if quit {
		return false
	}

	logrus.Infof("Processing item: %+v", item)

	defer c.queue.Done(item)

	var err error
	switch item.Type {
	case CRDResourceType:
		if item.GVR == nil {
			logrus.Error("The groupVersionResource can not be nil")
			c.queue.Forget(item)
			return true
		}
		err = DeleteCRD(context.Background(), *item.GVR, item.Namespace, item.Name)
	case JobResourceType:
		err = DeleteJob(context.Background(), item.Namespace, item.Name)
	default:
		logrus.Errorf("unknown resource type: %s", item.Type)
		return true
	}

	if err != nil {
		if errors.IsNotFound(err) {
			logrus.Warnf("%s already deleted: %v", item.Type, err)
			c.queue.Forget(item)
			return true
		}

		logrus.WithField("namespace", item.Namespace).WithField("name", item.Name).Error(err)
		c.queue.AddRateLimited(item)
		return true
	}

	c.queue.Forget(item)
	return true
}

func isTargetNamespace(namespace string) bool {
	return slices.Contains(config.GetStringSlice("injection.namespace"), namespace)
}

func getCRDConditionStatus(conditions []any, conditionType string) bool {
	for _, c := range conditions {
		condition, ok := c.(map[string]any)
		if !ok {
			continue
		}

		t, _, _ := unstructured.NestedString(condition, "type")
		status, _, _ := unstructured.NestedString(condition, "status")
		if t == conditionType {
			return status == "True"
		}
	}

	return false
}

func getCRDEventTimeRanges(records []any) []timeRange {
	var timeRanges []timeRange
	for _, r := range records {
		record, ok := r.(map[string]any)
		if !ok {
			continue
		}

		var startTime, endTime *time.Time
		events, _, _ := unstructured.NestedSlice(record, "events")
		for _, e := range events {
			event, ok := e.(map[string]any)
			if !ok {
				continue
			}

			operation, _, _ := unstructured.NestedString(event, "operation")
			eventType, _, _ := unstructured.NestedString(event, "type")

			if eventType == "Succeeded" && operation == "Apply" {
				startTime, _ = parseEventTime(event)
			}

			if eventType == "Succeeded" && operation == "Recover" {
				endTime, _ = parseEventTime(event)
			}
		}

		if startTime != nil && endTime != nil {
			timeRanges = append(timeRanges, timeRange{Start: *startTime, End: *endTime})
		}
	}

	if len(timeRanges) != 0 {
		return timeRanges
	}

	return []timeRange{}
}

func parseEventTime(event map[string]any) (*time.Time, error) {
	t, _, _ := unstructured.NestedString(event, "timestamp")
	if t, err := time.Parse(time.RFC3339, t); err == nil {
		return &t, nil
	}

	return nil, fmt.Errorf("parse event time failed")
}

func extractJobError(job *batchv1.Job) string {
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobFailed && condition.Status == "True" {
			return fmt.Sprintf("Reason: %s, Message: %s", condition.Reason, condition.Message)
		}
	}

	return "Unknown error"
}

func checkPodReason(pod *corev1.Pod, reason string) bool {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		// 检查容器的 Waiting 状态
		if containerStatus.State.Waiting != nil {
			if containerStatus.State.Waiting.Reason == reason {
				return true
			}
		}

		// 某些情况下，镜像拉取失败可能导致容器直接终止（例如重试次数耗尽）
		if containerStatus.State.Terminated != nil {
			if containerStatus.State.Terminated.Reason == reason {
				return true
			}
		}
	}

	return false
}

func handlePodError(pod *corev1.Pod, job *batchv1.Job, reason string) {
	// 获取 Pod 事件
	events, err := k8sClient.CoreV1().Events(pod.Namespace).List(context.TODO(), metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s", pod.Name),
	})
	if err != nil {
		logrus.WithField("pod_name", pod.Name).Errorf("failed to get events for Pod: %v", err)
		return
	}

	var messages []string
	for _, event := range events.Items {
		if event.Type == "Warning" && event.Reason == "Failed" {
			messages = append(messages, event.Message)
		}
	}

	logrus.WithFields(logrus.Fields{
		"job_name": job.Name,
		"pod_name": pod.Name,
		"reason":   reason,
	}).Error(messages)
}
