package k8s

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
)

func deleteCRD(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	logEntry := logrus.WithField("namespace", namespace).WithField("name", name)

	// 1. 检查资源是否存在
	obj, err := k8sDynamicClient.Resource(gvr).Namespace(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to get CRD: %v", err)
	}

	// 2. 检查是否已在删除中
	if !obj.GetDeletionTimestamp().IsZero() {
		logEntry.Info("CRD is already being deleted")
		return nil
	}

	// 3. 执行删除（幂等操作）
	err = k8sDynamicClient.Resource(gvr).Namespace(namespace).Delete(ctx, name, deleteOptions)
	if err != nil && !errors.IsNotFound(err) {
		if timeoutCtx.Err() != nil {
			return fmt.Errorf("timeout while deleting CRD %s/%s: %v", namespace, name, timeoutCtx.Err())
		}

		return fmt.Errorf("failed to delete CRD: %v", err)
	}

	return nil
}

func cleanFinalizers(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) error {
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	logEntry := logrus.WithFields(logrus.Fields{
		"namespace": namespace,
		"name":      name,
	})

	patchBytes := []byte(`{"metadata":{"finalizers":[]}}`)
	_, err := k8sDynamicClient.Resource(gvr).Namespace(namespace).Patch(
		timeoutCtx,
		name,
		types.MergePatchType,
		patchBytes,
		metav1.PatchOptions{},
	)

	if err != nil && !errors.IsNotFound(err) {
		if timeoutCtx.Err() != nil {
			return fmt.Errorf("timeout while patching resource %s/%s: %v", namespace, name, timeoutCtx.Err())
		}

		return fmt.Errorf("failed to patch finalizers: %v", err)
	}

	logEntry.Info("Successfully cleared finalizers")
	return nil
}
