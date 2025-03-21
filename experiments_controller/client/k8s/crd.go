package k8s

import (
	"context"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func DeleteCRD(ctx context.Context, gvr schema.GroupVersionResource, namespace, name string) error {
	deletePolicy := metav1.DeletePropagationForeground
	deleteOptions := metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}

	if err := k8sDynamicClient.Resource(gvr).Namespace(namespace).Delete(ctx, name, deleteOptions); err != nil {
		return err
	}

	return nil
}
