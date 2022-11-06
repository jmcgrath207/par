package controllers

import (
	"context"
	"fmt"
	"github.com/snorwin/k8s-generic-webhook/pkg/webhook"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type PodDNSWebhook struct {
	webhook.MutatingWebhook
}

func (w *PodDNSWebhook) SetupWebhookWithManager(mgr manager.Manager) error {
	return webhook.NewGenericWebhookManagedBy(mgr).
		For(&corev1.Pod{}).
		Complete(w)
}

func (w *PodDNSWebhook) Mutate(ctx context.Context, request admission.Request, object runtime.Object) admission.Response {
	_ = log.FromContext(ctx)

	pod := object.(*corev1.Pod)
	fmt.Sprint(pod.Spec.AutomountServiceAccountToken)
	// TODO add your programmatic mutation logic here

	return admission.Allowed("")
}
