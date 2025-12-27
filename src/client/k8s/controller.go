package k8s

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"sync"
	"time"

	chaosCli "github.com/LGU-SE-Internal/chaos-experiment/client"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"

	"aegis/config"
	"aegis/consts"

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
	callback         Callback
	crdInformers     map[string]map[schema.GroupVersionResource]cache.SharedIndexInformer
	informerMu       sync.RWMutex // Protects crdInformers map
	activeNamespaces map[string]bool
	namespaceMu      sync.RWMutex
	jobInformer      cache.SharedIndexInformer
	podInformer      cache.SharedIndexInformer
	queue            workqueue.TypedRateLimitingInterface[QueueItem]
	chaosGVRs        []schema.GroupVersionResource
	ctx              context.Context
	cancelFunc       context.CancelFunc
}

func NewController() *Controller {
	crdInformers := make(map[string]map[schema.GroupVersionResource]cache.SharedIndexInformer)
	activeNamespaces := make(map[string]bool)

	tweakListOptions := func(options *metav1.ListOptions) {
		options.LabelSelector = fmt.Sprintf("%s=%s", consts.K8sLabelAppID, consts.AppID)
	}

	platformFactory := informers.NewSharedInformerFactoryWithOptions(
		GetK8sClient(),
		resyncPeriod,
		informers.WithNamespace(config.GetString("k8s.namespace")),
		informers.WithTweakListOptions(tweakListOptions),
	)

	queue := workqueue.NewTypedRateLimitingQueue(
		workqueue.DefaultTypedControllerRateLimiter[QueueItem](),
	)

	jobInformer := platformFactory.Batch().V1().Jobs().Informer()
	podInformer := platformFactory.Core().V1().Pods().Informer()

	chaosGVRs := make([]schema.GroupVersionResource, 0, len(chaosCli.GetCRDMapping()))
	for gvr := range chaosCli.GetCRDMapping() {
		chaosGVRs = append(chaosGVRs, gvr)
	}

	return &Controller{
		crdInformers:     crdInformers,
		activeNamespaces: activeNamespaces,
		jobInformer:      jobInformer,
		podInformer:      podInformer,
		queue:            queue,
		chaosGVRs:        chaosGVRs,
	}
}

func (c *Controller) Initialize(ctx context.Context, cancelFunc context.CancelFunc, callback Callback) {
	logrus.Info("Initializing controller...")

	c.ctx = ctx
	c.cancelFunc = cancelFunc
	c.callback = callback

	if err := c.startJobAndPodInformers(); err != nil {
		logrus.Fatalf("Failed to start Job and Pod informers: %v", err)
	}

	go c.startQueueWorker()

	logrus.Info("Controller initialized successfully")

	<-ctx.Done()
	c.stop()
	logrus.Info("Controller shutdown complete")
}

// AddNamespaceInformers dynamically adds informers for new namespaces
func (c *Controller) AddNamespaceInformers(namespaces []string) error {
	if len(namespaces) == 0 {
		logrus.Debug("No namespaces to add")
		return nil
	}

	logrus.Infof("Adding informers for %d namespace(s): %v", len(namespaces), namespaces)

	tweakListOptions := func(options *metav1.ListOptions) {
		options.LabelSelector = fmt.Sprintf("%s=%s", consts.K8sLabelAppID, consts.AppID)
	}

	addedCount := 0
	newInformers := make(map[string][]cache.InformerSynced) // Track new informers for cache sync

	for _, namespace := range namespaces {
		// Check if informer already exists (with read lock)
		c.informerMu.RLock()
		_, exists := c.crdInformers[namespace]
		c.informerMu.RUnlock()

		if exists {
			logrus.Infof("Informer for namespace %s already exists, skipping", namespace)
			// Still mark as active in case it was previously disabled
			c.namespaceMu.Lock()
			c.activeNamespaces[namespace] = true
			c.namespaceMu.Unlock()
			continue
		}

		// Create new factory for this namespace
		logrus.Infof("Creating new CRD informers for namespace: %s", namespace)
		chaosFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
			GetK8sDynamicClient(),
			resyncPeriod,
			namespace,
			tweakListOptions,
		)

		// Create GVR informers
		gvrInformers := make(map[schema.GroupVersionResource]cache.SharedIndexInformer, len(c.chaosGVRs))
		syncFuncs := make([]cache.InformerSynced, 0, len(c.chaosGVRs))

		// CRITICAL: Mark namespace as active BEFORE starting informers
		// This prevents race condition where informer cache sync triggers AddFunc
		// before namespace is marked active, causing events to be ignored
		c.namespaceMu.Lock()
		c.activeNamespaces[namespace] = true
		c.namespaceMu.Unlock()

		for _, gvr := range c.chaosGVRs {
			informer := chaosFactory.ForResource(gvr).Informer()

			// Register event handler
			logrus.Infof("Registering event handler for %s (Group: %s, Version: %s) in namespace %s",
				gvr.Resource, gvr.Group, gvr.Version, namespace)
			if _, err := informer.AddEventHandler(c.genCRDEventHandlerFuncs(gvr)); err != nil {
				return fmt.Errorf("failed to add event handler for %s in namespace %s: %w",
					gvr.Resource, namespace, err)
			}

			gvrInformers[gvr] = informer
			syncFuncs = append(syncFuncs, informer.HasSynced)

			// Start informer immediately
			logrus.Infof("Starting informer for %s in namespace %s", gvr.Resource, namespace)
			go informer.Run(c.ctx.Done())
		}

		// Update crdInformers map with write lock
		c.informerMu.Lock()
		c.crdInformers[namespace] = gvrInformers
		c.informerMu.Unlock()

		newInformers[namespace] = syncFuncs
		addedCount++
		logrus.Infof("Successfully added and started %d CRD informer(s) for namespace: %s", len(gvrInformers), namespace)
	}

	// Wait for all new informers to sync their caches
	if len(newInformers) > 0 {
		allSyncFuncs := make([]cache.InformerSynced, 0)
		for _, syncFuncs := range newInformers {
			allSyncFuncs = append(allSyncFuncs, syncFuncs...)
		}

		logrus.Infof("Waiting for %d new CRD informer(s) to sync caches...", len(allSyncFuncs))
		if !cache.WaitForCacheSync(c.ctx.Done(), allSyncFuncs...) {
			return fmt.Errorf("timed out waiting for new CRD informer caches to sync")
		}
		logrus.Info("All new CRD informer caches synced successfully")
	}

	logrus.Infof("Namespace informer addition completed: %d new, %d existing", addedCount, len(namespaces)-addedCount)
	return nil
}

// RemoveNamespaceInformers marks namespaces as inactive
// Note: Kubernetes informers cannot be gracefully stopped, so we keep them running
// but mark the namespaces as inactive to filter events in handlers
func (c *Controller) RemoveNamespaceInformers(namespaces []string) {
	if len(namespaces) == 0 {
		logrus.Debug("No namespaces to remove")
		return
	}

	logrus.Infof("Marking %d namespace(s) as inactive: %v", len(namespaces), namespaces)

	c.namespaceMu.Lock()
	defer c.namespaceMu.Unlock()

	deactivatedCount := 0
	for _, ns := range namespaces {
		wasActive := c.activeNamespaces[ns]
		c.activeNamespaces[ns] = false
		if wasActive {
			deactivatedCount++
			logrus.Infof("Namespace %s marked as inactive (events will be ignored)", ns)
		} else {
			logrus.Debugf("Namespace %s was already inactive", ns)
		}
	}

	logrus.Infof("Namespace deactivation completed: %d deactivated, %d already inactive",
		deactivatedCount, len(namespaces)-deactivatedCount)
}

// stop gracefully stops the controller and cancels all informers
func (c *Controller) stop() {
	logrus.Info("Stopping controller and cancelling all informers...")
	if c.cancelFunc != nil {
		c.cancelFunc()
	}
	c.queue.ShutDown()
}

// startJobAndPodInformers starts Job and Pod informers with event handlers
// It should be called after the controller is created but can be called at any time
func (c *Controller) startJobAndPodInformers() error {
	if c.callback == nil {
		return fmt.Errorf("callback must be set before registering Job and Pod informers")
	}

	logrus.Info("Starting Job and Pod informers...")

	// Register event handlers
	if c.jobInformer != nil {
		if _, err := c.jobInformer.AddEventHandler(c.genJobEventHandlerFuncs()); err != nil {
			return fmt.Errorf("failed to add event handler for Job informer: %w", err)
		}
		logrus.Info("Registered event handler for Job informer")

		// Start Job informer
		go c.jobInformer.Run(c.ctx.Done())
		logrus.Info("Started Job informer")
	} else {
		logrus.Warn("Job informer is nil, skipping registration")
	}

	if c.podInformer != nil {
		if _, err := c.podInformer.AddEventHandler(c.genPodEventHandlerFuncs()); err != nil {
			return fmt.Errorf("failed to add event handler for Pod informer: %w", err)
		}
		logrus.Info("Registered event handler for Pod informer")

		// Start Pod informer
		go c.podInformer.Run(c.ctx.Done())
		logrus.Info("Started Pod informer")
	} else {
		logrus.Warn("Pod informer is nil, skipping registration")
	}

	// Wait for cache sync
	syncFuncs := []cache.InformerSynced{}
	if c.jobInformer != nil {
		syncFuncs = append(syncFuncs, c.jobInformer.HasSynced)
	}
	if c.podInformer != nil {
		syncFuncs = append(syncFuncs, c.podInformer.HasSynced)
	}

	if len(syncFuncs) > 0 {
		logrus.Info("Waiting for Job and Pod informer caches to sync...")
		if !cache.WaitForCacheSync(c.ctx.Done(), syncFuncs...) {
			return fmt.Errorf("timed out waiting for Job and Pod informer caches to sync")
		}
		logrus.Info("Job and Pod informer caches synced successfully")
	}

	logrus.Info("Successfully registered and started Job and Pod informers")
	return nil
}

// startQueueWorker starts the queue worker for processing delayed tasks
func (c *Controller) startQueueWorker() {
	logrus.Info("Starting queue worker...")
	wait.Until(c.runWorker, time.Second, c.ctx.Done())
}

// genCRDEventHandlerFuncs generates event handler functions for a given CRD GVR
func (c *Controller) genCRDEventHandlerFuncs(gvr schema.GroupVersionResource) cache.ResourceEventHandlerFuncs {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			u := obj.(*unstructured.Unstructured)

			logrus.Infof("CRD AddFunc triggered: resource=%s, namespace=%s, name=%s",
				gvr.Resource, u.GetNamespace(), u.GetName())

			// Check if namespace is active
			if !c.isNamespaceActive(u.GetNamespace()) {
				logrus.Warnf("Ignoring CRD add event for inactive namespace %s: %s/%s",
					u.GetNamespace(), gvr.Resource, u.GetName())
				return
			}

			if consts.InitialTime != nil {
				creationTime := u.GetCreationTimestamp().Time
				if creationTime.Before(*consts.InitialTime) {
					logrus.Debugf("Ignoring CRD add event for object created before initial time: %s/%s in %s",
						gvr.Resource, u.GetName(), u.GetNamespace())
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

			// Check if namespace is active
			if !c.isNamespaceActive(newU.GetNamespace()) {
				logrus.Debugf("Ignoring CRD update event for inactive namespace %s: %s/%s",
					newU.GetNamespace(), gvr.Resource, newU.GetName())
				return
			}

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

			// Check if namespace is active
			if !c.isNamespaceActive(u.GetNamespace()) {
				logrus.Debugf("Ignoring CRD delete event for inactive namespace %s: %s/%s",
					u.GetNamespace(), gvr.Resource, u.GetName())
				return
			}

			logrus.WithFields(logrus.Fields{
				"type":      gvr.Resource,
				"namespace": u.GetNamespace(),
				"name":      u.GetName(),
			}).Info("Chaos experiment deleted successfully")
			c.callback.HandleCRDDelete(u.GetNamespace(), u.GetAnnotations(), u.GetLabels())
		},
	}
}

// genJobEventHandlerFuncs generates event handler functions for Job informer
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

// genPodEventHandlerFuncs generates event handler functions for Pod informer
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
						job, err := GetJob(c.ctx, newPod.Namespace, jobOwnerRef.Name)
						if err != nil {
							logrus.WithField("job_name", jobOwnerRef.Name).Error(err)
						}

						if job != nil {
							handlePodError(c.ctx, newPod, job, reason)
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

// runWorker processes items from the queue
func (c *Controller) runWorker() {
	for c.processQueueItem() {
	}
}

// processQueueItem processes a single item from the queue
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

	obj, err := GetK8sDynamicClient().
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

// isNamespaceActive checks if a namespace is active and should process events
func (c *Controller) isNamespaceActive(namespace string) bool {
	c.namespaceMu.RLock()
	defer c.namespaceMu.RUnlock()

	active, exists := c.activeNamespaces[namespace]
	logrus.Debugf("Checking namespace %s: exists=%v, active=%v", namespace, exists, active)
	return exists && active
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
	events, err := GetK8sClient().CoreV1().Events(pod.Namespace).List(ctx, metav1.ListOptions{
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
