package webhook

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type DeploymentUpdate struct{}

func (d *DeploymentUpdate) Default(ctx context.Context, obj runtime.Object) error {
	log := logf.FromContext(ctx)
	deployment, ok := obj.(*appsv1.Deployment)
	if !ok {
		return fmt.Errorf("expected a Pod but got a %T", obj)
	}

	if deployment.Annotations == nil {
		deployment.Annotations = map[string]string{}
	}
	deployment.Annotations["example-mutating-admission-webhook"] = "foo"
	log.Info("Annotated Pod")

	return nil
}
