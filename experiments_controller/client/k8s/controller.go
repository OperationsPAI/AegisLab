package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/CUHK-SE-Group/rcabench/config"
	"github.com/sirupsen/logrus"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/tools/cache"
)

const resyncPeriod = 10 * time.Second

type JobEnv struct {
	Namespace string
	Service   string
	StartTime time.Time
	EndTime   time.Time
}

type Callback interface {
	HandleJobAdd(labels map[string]string)
	HandleJobUpdate(labels map[string]string, status string)
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
			logrus.Infof("Job %s created successfully in namespace %s", job.Name, job.Namespace)
			callback.HandleJobAdd(job.Labels)
		},
		UpdateFunc: func(oldObj, newObj any) {
			oldJob := oldObj.(*batchv1.Job)
			newJob := newObj.(*batchv1.Job)

			if callback != nil && oldJob.Name == newJob.Name {
				if oldJob.Status.Succeeded == 0 && newJob.Status.Succeeded > 0 {
					callback.HandleJobUpdate(newJob.Labels, "Completed")
					if err := DeleteJob(context.Background(), config.GetString("k8s.namespace"), newJob.Name); err != nil {
						logrus.Error(err)
					}
				}

				if oldJob.Status.Failed == 0 && newJob.Status.Failed > 0 {
					callback.HandleJobUpdate(newJob.Labels, "Error")
				}
			}
		},
		DeleteFunc: func(obj any) {
			job := obj.(*batchv1.Job)
			logrus.Infof("Job %s deleted successfully in namespace %s", job.Name, job.Namespace)
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
