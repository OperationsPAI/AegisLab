package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const (
	jobKind      = "Job"
	resyncPeriod = 10 * time.Second
)

type JobEnv struct {
	Namespace string
	Service   string
	StartTime time.Time
	EndTime   time.Time
}

type Callback interface {
	HandleJobAdd(labels map[string]string)
	HandleJobUpdate(labels map[string]string, status string)
	HandlePodUpdate()
}

type Controller struct {
	jobInformer cache.SharedIndexInformer
	podInformer cache.SharedIndexInformer
}

func NewController() *Controller {
	factory := informers.NewSharedInformerFactoryWithOptions(
		k8sClient,
		resyncPeriod,
		informers.WithNamespace(config.GetString("k8s.namespace")),
	)

	return &Controller{
		jobInformer: factory.Batch().V1().Jobs().Informer(),
		podInformer: factory.Core().V1().Pods().Informer(),
	}
}

func (c *Controller) initEventHandlers(callback Callback) {
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
						logrus.WithField("namespace", newJob.Namespace).WithField("job_name", newJob.Name).WithError(err).Error("Failed to delete")
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

	logrus.Info("Starting k8s controller")
	go c.jobInformer.Run(ctx.Done())
	go c.podInformer.Run(ctx.Done())

	if !cache.WaitForCacheSync(ctx.Done(), c.jobInformer.HasSynced, c.podInformer.HasSynced) {
		message := "Timed out waiting for caches to sync"
		runtime.HandleError(fmt.Errorf(message))
		logrus.Error(message)
		return
	}

	<-ctx.Done()
	logrus.Info("Stopping informer...")
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
