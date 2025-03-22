package k8s

import (
	"context"
	"fmt"
	"time"

	"slices"

	chaosCli "github.com/CUHK-SE-Group/chaos-experiment/client"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/CUHK-SE-Group/rcabench/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

const (
	jobKind      = "Job"
	resyncPeriod = 10 * time.Second
)

type JobEnv struct {
	Namespace   string
	Service     string
	PreDuration int
	StartTime   time.Time
	EndTime     time.Time
}

type Callback interface {
	HandleCRDUpdate(namespace, pod, name string)
	HandleJobAdd(labels map[string]string)
	HandleJobUpdate(labels map[string]string, status string)
	HandlePodUpdate()
}

type Controller struct {
	crdInformers map[schema.GroupVersionResource]cache.SharedIndexInformer
	jobInformer  cache.SharedIndexInformer
	podInformer  cache.SharedIndexInformer
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

	return &Controller{
		crdInformers: crdInformers,
		jobInformer:  factory.Batch().V1().Jobs().Informer(),
		podInformer:  factory.Core().V1().Pods().Informer(),
	}
}

func (c *Controller) initEventHandlers(callback Callback) {
	for gvr, informer := range c.crdInformers {
		informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj any) {
				u := obj.(*unstructured.Unstructured)
				if isTargetNamespace(u.GetNamespace()) {
					logrus.WithFields(logrus.Fields{
						"type":      gvr.Resource,
						"namespace": u.GetNamespace(),
						"name":      u.GetName(),
					}).Info("Chaos experiment created successfully")
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

					oldConditions, _, _ := unstructured.NestedSlice(oldU.Object, "status", "conditions")
					newConditions, _, _ := unstructured.NestedSlice(newU.Object, "status", "conditions")

					// 解析关键条件状态
					oldAllInjected := getCRDConditionStatus(oldConditions, "AllInjected")
					oldAllRecovered := getCRDConditionStatus(oldConditions, "AllRecovered")

					newAllInjected := getCRDConditionStatus(newConditions, "AllInjected")
					newAllRecovered := getCRDConditionStatus(newConditions, "AllRecovered")

					if !oldAllInjected && newAllInjected {
						logEntry.Infof("All targets injected in chaos experiment")
					}

					if !oldAllRecovered && newAllRecovered {
						logEntry.Infof("All targets recoverd in chaos experiment")

						kind := newU.GetKind()
						pod, _, _ := unstructured.NestedString(newU.Object, "spec", "selector", "labelSelectors", "app")

						chaosGVRMapping := make(map[string]schema.GroupVersionResource)
						for gvr, obj := range chaosCli.GetCRDMapping() {
							chaosGVRMapping[utils.GetTypeName(obj)] = gvr
						}

						gvr, ok := chaosGVRMapping[kind]
						if !ok {
							logEntry.Error("The gvr resource can not be found")
						}

						callback.HandleCRDUpdate(newU.GetNamespace(), pod, newU.GetName())
						if err := DeleteCRD(context.Background(), gvr, newU.GetNamespace(), newU.GetName()); err != nil {
							logEntry.Errorf("Failed to delete CRD: %v", err)
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
		})
	}

	c.jobInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj any) {
			job := obj.(*batchv1.Job)
			logrus.WithField("namespace", job.Namespace).WithField("job_name", job.Name).Info("Job created successfully")
			callback.HandleJobAdd(job.Labels)
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldJob := oldObj.(*batchv1.Job)
			newJob := newObj.(*batchv1.Job)

			if callback != nil && oldJob.Name == newJob.Name {
				if oldJob.Status.Succeeded == 0 && newJob.Status.Succeeded > 0 {
					callback.HandleJobUpdate(newJob.Labels, "Completed")
					if err := DeleteJob(context.Background(), config.GetString("k8s.namespace"), newJob.Name); err != nil {
						logrus.WithField("namespace", newJob.Namespace).WithField("job_name", newJob.Name).Errorf("Failed to delete job: %v", err)
					}
				}

				if oldJob.Status.Failed == 0 && newJob.Status.Failed > 0 {
					callback.HandleJobUpdate(newJob.Labels, "Error")
				}
			}
		},
		DeleteFunc: func(obj any) {
			job := obj.(*batchv1.Job)
			logrus.WithField("namespace", job.Namespace).WithField("job_name", job.Name).Infof("Job deleted successfully")
		},
	})

	c.podInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
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
							logrus.WithField("job_name", jobOwnerRef.Name).WithError(err)
						}

						if job != nil {
							handlePodError(newPod, job, reason)
							break
						}
					}
				}
			}
		},
		DeleteFunc: func(obj any) {
			pod := obj.(*corev1.Pod)
			logrus.WithField("namespace", pod.Namespace).WithField("pod_name", pod.Name).Infof("Pod deleted successfully")
		},
	})
}

func (c *Controller) Run(ctx context.Context, callback Callback) {
	defer runtime.HandleCrash()

	c.initEventHandlers(callback)

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
		message := "Timed out waiting for caches to sync"
		runtime.HandleError(fmt.Errorf(message))
		logrus.Error(message)
		return
	}

	<-ctx.Done()
	logrus.Info("Stopping informer controller...")
}

func isTargetNamespace(namespace string) bool {
	return slices.Contains(config.GetStringSlice("injection.namespace"), namespace)
}

func getCRDConditionStatus(conditions []any, conditionType string) bool {
	for _, c := range conditions {
		condition, ok := c.(map[string]interface{})
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
		logrus.WithField("pod_name", pod.Name).WithError(err).Errorf("Failed to get events for Pod")
		return
	}

	var messages []string
	for _, event := range events.Items {
		if event.Type == "Warning" && event.Reason == "Failed" {
			messages = append(messages, event.Message)
		}
	}

	fields := logrus.Fields{
		"job_name": job.Name,
		"pod_name": pod.Name,
		"reason":   reason,
	}
	logrus.WithFields(fields).Error(messages)

	if err := DeleteJob(context.TODO(), job.Namespace, job.Name); err != nil {
		logrus.WithField("namespace", job.Namespace).WithField("job_name", job.Name).WithError(err).Error("Failed to delete Job")
	}
}
