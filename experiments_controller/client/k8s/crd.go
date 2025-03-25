package k8s

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func DeleteCRD(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) error {
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
		return fmt.Errorf("Failed to get CRD: %v", err)
	}

	// 2. 检查是否已在删除中
	if !obj.GetDeletionTimestamp().IsZero() {
		logEntry.Info("CRD is already being deleted")
		return nil
	}

	// 3. 检查是否处于可删除状态
	conditions, _, _ := unstructured.NestedSlice(obj.Object, "status", "conditions")
	if !getCRDConditionStatus(conditions, "AllRecovered") {
		logEntry.Info("CRD is not in AllRecovered state, will retry later")
		return fmt.Errorf("Resource not ready for deletion")
	}

	// 4. 执行删除（幂等操作）
	err = k8sDynamicClient.Resource(gvr).Namespace(namespace).Delete(ctx, name, deleteOptions)
	if err != nil && !errors.IsNotFound(err) {
		return fmt.Errorf("Failed to delete CRD: %v", err)
	}

	return nil
}
