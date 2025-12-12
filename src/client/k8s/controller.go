package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"time"

	chaosCli "github.com/LGU-SE-Internal/chaos-experiment/client"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"

	"aegis/config"
	"aegis/consts"
	"aegis/utils"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
)

type ActionType string

const (
	jobKind      = "Job"
	resyncPeriod = 5 * time.Second

	CheckRecovery ActionType = "CheckRecovery"
	DeleteCRD     ActionType = "DeleteCRD"
	DeleteJob     ActionType = "DeleteJob"
)

type timeRange struct {
	Start time.Time
	End   time.Time
}

type Callback interface {
	HandleCRDAdd(name string, annotations map[string]string, labels map[string]string)
	HandleCRDDelete(namespace string, annotations map[string]string, labels map[string]string)
	HandleCRDFailed(name string, annotations map[string]string, labels map[string]string, errMsg string)
	HandleCRDSucceeded(namespace, pod, name string, startTime, endTime time.Time, annotations map[string]string, labels map[string]string)
	HandleJobAdd(annotations map[string]string, labels map[string]string)
	HandleJobFailed(job *batchv1.Job, annotations map[string]string, labels map[string]string)
	HandleJobSucceeded(job *batchv1.Job, annotations map[string]string, labels map[string]string)
}

type QueueItem struct {
	Type       ActionType
	Namespace  string
	Name       string
	Duration   time.Duration
	GVR        *schema.GroupVersionResource
	RetryCount int
	MaxRetries int
}

type Controller struct {
	callback     Callback
	crdInformers map[string]map[schema.GroupVersionResource]cache.SharedIndexInformer
	jobInformer  cache.SharedIndexInformer
	podInformer  cache.SharedIndexInformer
	queue        workqueue.TypedRateLimitingInterface[QueueItem]
}

func NewController() *Controller {
	namespaces, err := utils.GetAllNamespaces()
	if err != nil {
		logrus.WithField("func", "config.GetAllNamespaces").Error(err)
		panic(err)
	}

	chaosGVRs := make([]schema.GroupVersionResource, 0, len(chaosCli.GetCRDMapping()))
	for gvr := range chaosCli.GetCRDMapping() {
		chaosGVRs = append(chaosGVRs, gvr)
	}

	tweakListOptions := func(options *metav1.ListOptions) {
		options.LabelSelector = fmt.Sprintf("%s=%s", consts.K8sLabelAppID, consts.AppID)
	}

	crdInformers := make(map[string]map[schema.GroupVersionResource]cache.SharedIndexInformer, len(namespaces))
	for _, namespace := range namespaces {
		chaosFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
			k8sDynamicClient,
			resyncPeriod,
			namespace,
			tweakListOptions,
		)

		gvrInformers := make(map[schema.GroupVersionResource]cache.SharedIndexInformer, len(chaosGVRs))
		for _, gvr := range chaosGVRs {
			gvrInformers[gvr] = chaosFactory.ForResource(gvr).Informer()
		}

		crdInformers[namespace] = gvrInformers
	}

	platformFactory := informers.NewSharedInformerFactoryWithOptions(
		k8sClient,
		resyncPeriod,
		informers.WithNamespace(config.GetString("k8s.namespace")),
		informers.WithTweakListOptions(tweakListOptions),
	)

	queue := workqueue.NewTypedRateLimitingQueue(
		workqueue.DefaultTypedControllerRateLimiter[QueueItem](),
	)

	return &Controller{
		crdInformers: crdInformers,
		jobInformer:  platformFactory.Batch().V1().Jobs().Informer(),
		podInformer:  platformFactory.Core().V1().Pods().Informer(),
		queue:        queue,
	}
}

func (c *Controller) Run(ctx context.Context, callback Callback) {
	defer runtime.HandleCrash()
	defer c.queue.ShutDown()

	c.callback = callback
	c.registerEventHandlers(ctx)

	logrus.Info("Starting informer controller")

	for _, gvrInformers := range c.crdInformers {
		for _, informer := range gvrInformers {
			go informer.Run(ctx.Done())
		}
	}
	go c.jobInformer.Run(ctx.Done())
	go c.podInformer.Run(ctx.Done())

	allSyncs := []cache.InformerSynced{c.jobInformer.HasSynced, c.podInformer.HasSynced}
	for _, gvrInformers := range c.crdInformers {
		for _, informer := range gvrInformers {
			allSyncs = append(allSyncs, informer.HasSynced)
		}
	}

	if !cache.WaitForCacheSync(ctx.Done(), allSyncs...) {
		message := "timed out waiting for caches to sync"
		runtime.HandleError(fmt.Errorf("%s", message))
		logrus.Error(message)
		return
	}

	logrus.Info("starting queue worker...")
	go wait.Until(c.runWorker, time.Second, ctx.Done())

	<-ctx.Done()
	logrus.Info("Stopping informer controller...")
}

func (c *Controller) registerEventHandlers(ctx context.Context) {
	for _, gvrInformers := range c.crdInformers {
		for gvr, informer := range gvrInformers {
			if _, err := informer.AddEventHandler(c.genCRDEventHandlerFuncs(gvr)); err != nil {
				logrus.WithFields(logrus.Fields{
					"gvr":  gvr.Resource,
					"func": "genCRDEventHandlerFuncs",
				}).Error("failed to add event handler")
				return
			}
		}
	}

	if _, err := c.jobInformer.AddEventHandler(c.genJobEventHandlerFuncs()); err != nil {
		logrus.WithField("func", "genJobEventHandlerFuncs").Error("failed to add event handler")
		return
	}

	if _, err := c.podInformer.AddEventHandler(c.genPodEventHandlerFuncs(ctx)); err != nil {
		logrus.WithField("func", "genPodEventHandlerFuncs").Error("failed to add event handler")
		return
	}
}

func (c *Controller) genCRDEventHandlerFuncs(gvr schema.GroupVersionResource) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			u := obj.(*unstructured.Unstructured)

			if consts.InitialTime != nil {
				creationTime := u.GetCreationTimestamp().Time
				if creationTime.Before(*consts.InitialTime) {
					return
				}
			}

			c.callback.HandleCRDAdd(u.GetName(), u.GetAnnotations(), u.GetLabels())
			logrus.WithFields(logrus.Fields{
				"type":      gvr.Resource,
				"namespace": u.GetNamespace(),
				"name":      u.GetName(),
			}).Info("chaos experiment created successfully")
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldU := oldObj.(*unstructured.Unstructured)
			newU := newObj.(*unstructured.Unstructured)

			if oldU.GetName() == newU.GetName() {
				logEntry := logrus.WithFields(logrus.Fields{
					"type":      gvr.Resource,
					"namespace": newU.GetNamespace(),
					"name":      newU.GetName(),
				})

				oldPhase, _, _ := unstructured.NestedString(oldU.Object, "status", "experiment", "desiredPhase")
				newPhase, _, _ := unstructured.NestedString(newU.Object, "status", "experiment", "desiredPhase")
				if oldPhase == "Run" && newPhase == "Stop" {
					conditions, _, _ := unstructured.NestedSlice(newU.Object, "status", "conditions")

					selected := getCRDConditionStatus(conditions, "Selected")
					if !selected {
						c.handleCRDFailed(gvr, newU, "failed to select app in the chaos experiment")
						return
					}

					allInjected := getCRDConditionStatus(conditions, "AllInjected")
					if !allInjected {
						c.handleCRDFailed(gvr, newU, "failed to inject all targets in the chaos experiment")
						return
					}
				}

				oldConditions, _, _ := unstructured.NestedSlice(oldU.Object, "status", "conditions")
				newConditions, _, _ := unstructured.NestedSlice(newU.Object, "status", "conditions")

				// Check if injected
				oldAllInjected := getCRDConditionStatus(oldConditions, "AllInjected")
				newAllInjected := getCRDConditionStatus(newConditions, "AllInjected")
				if !oldAllInjected && newAllInjected {
					logEntry.Infof("all targets injected in the chaos experiment")
					durationStr, _, _ := unstructured.NestedString(newU.Object, "spec", "duration")

					pattern := `(\d+)m`
					re := regexp.MustCompile(pattern)
					match := re.FindStringSubmatch(durationStr)
					if len(match) <= 1 {
						c.handleCRDFailed(gvr, newU, "failed to get the duration")
						return
					}

					duration, err := strconv.Atoi(match[1])
					if err != nil {
						c.handleCRDFailed(gvr, newU, "failed to get the duration of the chaos experiement")
						return
					}

					if duration > 0 {
						c.queue.AddAfter(QueueItem{
							Type:       CheckRecovery,
							Namespace:  newU.GetNamespace(),
							Name:       newU.GetName(),
							Duration:   time.Duration(duration) * consts.DefaultTimeUnit,
							GVR:        &gvr,
							RetryCount: 0,
							MaxRetries: 2,
						}, time.Duration(duration)*time.Minute)
					}
				}
			}
		},
		DeleteFunc: func(obj any) {
			u := obj.(*unstructured.Unstructured)
			logrus.WithFields(logrus.Fields{
				"type":      gvr.Resource,
				"namespace": u.GetNamespace(),
				"name":      u.GetName(),
			}).Info("Chaos experiment deleted successfully")
			c.callback.HandleCRDDelete(u.GetNamespace(), u.GetAnnotations(), u.GetLabels())
		},
	}
}

func (c *Controller) genJobEventHandlerFuncs() cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			if consts.InitialTime != nil {
				creationTime := obj.(*batchv1.Job).CreationTimestamp.Time
				if creationTime.Before(*consts.InitialTime) {
					return
				}
			}

			job := obj.(*batchv1.Job)
			logrus.WithFields(logrus.Fields{
				"namespace": job.Namespace,
				"job_name":  job.Name,
				"task_type": job.Labels[consts.JobLabelTaskType],
			}).Info("job created successfully")
			c.callback.HandleJobAdd(job.Annotations, job.Labels)
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldJob := oldObj.(*batchv1.Job)
			newJob := newObj.(*batchv1.Job)

			if oldJob.Name == newJob.Name {
				if oldJob.Status.Failed == *oldJob.Spec.BackoffLimit && newJob.Status.Failed == *newJob.Spec.BackoffLimit+1 {
					c.callback.HandleJobFailed(newJob, newJob.Annotations, newJob.Labels)
				}

				if oldJob.Status.Succeeded == 0 && newJob.Status.Succeeded > 0 {
					c.callback.HandleJobSucceeded(newJob, newJob.Annotations, newJob.Labels)
					if !config.GetBool("debugging.enabled") {
						c.queue.Add(QueueItem{
							Type:      DeleteJob,
							Namespace: newJob.Namespace,
							Name:      newJob.Name,
						})
					}
				}
			}
		},
		DeleteFunc: func(obj any) {
			job := obj.(*batchv1.Job)
			logrus.WithFields(logrus.Fields{
				"namespace": job.Namespace,
				"job_name":  job.Name,
				"task_type": job.Labels[consts.JobLabelTaskType],
			}).Infof("job delete successfully")
		},
	}
}

func (c *Controller) genPodEventHandlerFuncs(ctx context.Context) cache.ResourceEventHandlerFuncs {
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
						job, err := GetJob(ctx, newPod.Namespace, jobOwnerRef.Name)
						if err != nil {
							logrus.WithField("job_name", jobOwnerRef.Name).Error(err)
						}

						if job != nil {
							handlePodError(ctx, newPod, job, reason)
							// Trigger job failed callback to ensure proper cleanup (e.g., token release)
							c.callback.HandleJobFailed(job, job.Annotations, job.Labels)

							if !config.GetBool("debugging.enabled") {
								c.queue.Add(QueueItem{
									Type:      DeleteJob,
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
	case CheckRecovery:
		err = c.checkRecoveryStatus(item)
	case DeleteCRD:
		if item.GVR == nil {
			logrus.Error("The groupVersionResource can not be nil")
			c.queue.Forget(item)
			return true
		}

		err = deleteCRD(context.Background(), item.GVR, item.Namespace, item.Name)
	case DeleteJob:
		if !config.GetBool("debugging.enabled") {
			err = deleteJob(context.Background(), item.Namespace, item.Name)
		} else {
			logrus.WithFields(logrus.Fields{
				"namespace": item.Namespace,
				"name":      item.Name,
			}).Info("Skipping job deletion due to debugging mode enabled")
		}

	default:
		logrus.Errorf("unknown resource type: %s", item.Type)
		return true
	}

	if err != nil {
		logrus.WithField("namespace", item.Namespace).WithField("name", item.Name).Error(err)
		c.queue.AddRateLimited(item)
		return true
	}

	c.queue.Forget(item)
	return true
}

func (c *Controller) checkRecoveryStatus(item QueueItem) error {
	logEntry := logrus.WithFields(logrus.Fields{
		"type":      item.GVR.Resource,
		"namespace": item.Namespace,
		"name":      item.Name,
	})

	obj, err := k8sDynamicClient.
		Resource(*item.GVR).
		Namespace(item.Namespace).
		Get(context.Background(), item.Name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			logEntry.Info("the chaos experiment has been deleted")
			return nil
		}

		return fmt.Errorf("failed to get the CRD resource object: %w", err)
	}

	conditions, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	recovered := getCRDConditionStatus(conditions, "AllRecovered")
	if recovered {
		logEntry.Infof("all targets recoverd in the chaos experiment after %d attempts", item.RetryCount+1)
		c.handleCRDSuccess(*item.GVR, obj, item.Duration)
		return nil
	}

	if item.RetryCount < item.MaxRetries {
		logEntry.Warnf("Recovery not complete (attempt %d/%d), scheduling retry after 1 minute",
			item.RetryCount+1, item.MaxRetries+1)
		c.queue.AddAfter(QueueItem{
			Type:       CheckRecovery,
			Namespace:  item.Namespace,
			Name:       item.Name,
			Duration:   item.Duration,
			GVR:        item.GVR,
			RetryCount: item.RetryCount + 1,
			MaxRetries: item.MaxRetries,
		}, time.Duration(1)*time.Minute)
	} else {
		// If the retry count exceeds the maximum, log it and handle it normally.
		logEntry.Warningf("Recovery not complete after %d retries, giving up but processing as success", item.MaxRetries+1)
		c.handleCRDSuccess(*item.GVR, obj, item.Duration)
	}

	return nil
}

func (c *Controller) handleCRDSuccess(gvr schema.GroupVersionResource, u *unstructured.Unstructured, duration time.Duration) {
	newRecords, _, _ := unstructured.NestedSlice(u.Object, "status", "experiment", "containerRecords")
	timeRange := getCRDEventTimeRanges(newRecords, duration)
	if timeRange == nil {
		c.handleCRDFailed(gvr, u, "failed to get the start_time and end_time")
		return
	}

	pod, _, _ := unstructured.NestedString(u.Object, "spec", "selector", "labelSelectors", "app")
	c.callback.HandleCRDSucceeded(u.GetNamespace(), pod, u.GetName(), timeRange.Start, timeRange.End, u.GetAnnotations(), u.GetLabels())
	if !config.GetBool("debugging.enabled") {
		c.queue.Add(QueueItem{
			Type:      DeleteCRD,
			Namespace: u.GetNamespace(),
			Name:      u.GetName(),
			GVR:       &gvr,
		})
	}
}

func (c *Controller) handleCRDFailed(gvr schema.GroupVersionResource, u *unstructured.Unstructured, errMsg string) {
	logrus.WithFields(logrus.Fields{
		"type":      gvr.Resource,
		"namespace": u.GetNamespace(),
		"name":      u.GetName(),
	}).Errorf("CRD failed: %s", errMsg)

	c.callback.HandleCRDFailed(u.GetName(), u.GetAnnotations(), u.GetLabels(), errMsg)
	if !config.GetBool("debugging.enabled") {
		c.queue.Add(QueueItem{
			Type:      DeleteCRD,
			Namespace: u.GetNamespace(),
			Name:      u.GetName(),
			GVR:       &gvr,
		})
	}
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

func getCRDEventTimeRanges(records []any, duration time.Duration) *timeRange {
	r := records[0]
	record, ok := r.(map[string]any)
	if !ok {
		logrus.Error("invalid record format")
		return nil
	}

	var startTimePtr, endTimePtr *time.Time
	events, _, _ := unstructured.NestedSlice(record, "events")
	for _, e := range events {
		event, ok := e.(map[string]any)
		if !ok {
			continue
		}

		operation, _, _ := unstructured.NestedString(event, "operation")
		eventType, _, _ := unstructured.NestedString(event, "type")

		if eventType == "Succeeded" && operation == "Apply" {
			startTimePtr, _ = parseEventTime(event)
		}

		if eventType == "Succeeded" && operation == "Recover" {
			endTimePtr, _ = parseEventTime(event)
		}
	}

	if startTimePtr == nil {
		logrus.Error("start time not found in events")
		return nil
	}

	startTime := *startTimePtr
	var endTime time.Time

	if endTimePtr != nil {
		endTime = *endTimePtr
	} else {
		endTime = startTime.Add(duration)
		logrus.Infof("end time not found, calculated from start time + duration: %v", endTime)
	}

	return &timeRange{Start: startTime, End: endTime}
}

func parseEventTime(event map[string]any) (*time.Time, error) {
	t, _, _ := unstructured.NestedString(event, "timestamp")
	if t, err := time.Parse(time.RFC3339, t); err == nil {
		return &t, nil
	}

	return nil, fmt.Errorf("parse event time failed")
}

func checkPodReason(pod *corev1.Pod, reason string) bool {
	for _, containerStatus := range pod.Status.ContainerStatuses {
		// Check container Waiting status
		if containerStatus.State.Waiting != nil {
			if containerStatus.State.Waiting.Reason == reason {
				return true
			}
		}

		// In some cases, image pull failure may cause container to terminate directly (e.g., retry count exhausted)
		if containerStatus.State.Terminated != nil {
			if containerStatus.State.Terminated.Reason == reason {
				return true
			}
		}
	}

	return false
}

func handlePodError(ctx context.Context, pod *corev1.Pod, job *batchv1.Job, reason string) {
	// Get Pod events
	events, err := k8sClient.CoreV1().Events(pod.Namespace).List(ctx, metav1.ListOptions{
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
