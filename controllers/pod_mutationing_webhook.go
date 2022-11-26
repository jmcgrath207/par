package controllers

import (
	"context"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type WebHookManager struct {
	Mgr             manager.Manager
	CurrentWebHooks []client.Object
}

func strPtr(s string) *string { return &s }

// TODO: Working on Mutating hook logic
func (w *WebHookManager) newMutatingIsReadyWebhookFixture(service corev1.Service) admissionregistrationv1.MutatingWebhook {
	sideEffectsNone := admissionregistrationv1.SideEffectClassNone
	failOpen := admissionregistrationv1.Ignore
	return admissionregistrationv1.MutatingWebhook{
		Name: "mutating-is-webhook-configuration-ready.k8s.io",
		Rules: []admissionregistrationv1.RuleWithOperations{{
			Operations: []admissionregistrationv1.OperationType{admissionregistrationv1.Create},
			Rule: admissionregistrationv1.Rule{
				APIGroups:   []string{""},
				APIVersions: []string{"v1"},
				Resources:   []string{"configmaps"},
			},
		}},
		ClientConfig: admissionregistrationv1.WebhookClientConfig{
			Service: &admissionregistrationv1.ServiceReference{
				Namespace: service.Namespace,
				Name:      service.Name,
				Path:      strPtr("/always-deny"),
				Port:      pointer.Int32(9999),
			},
			CABundle: w.Mgr.GetConfig().CAData,
		},
		// network failures while the service network routing is being set up should be ignored by the marker
		FailurePolicy:           &failOpen,
		SideEffects:             &sideEffectsNone,
		AdmissionReviewVersions: []string{"v1", "v1beta1"},
		// Scope the webhook to just the markers namespace
		//NamespaceSelector: &metav1.LabelSelector{
		//	MatchLabels: map[string]string{f.UniqueName + "-markers": "true"},
		//},
		//// appease createMutatingWebhookConfiguration isolation requirements
		//ObjectSelector: &metav1.LabelSelector{
		//	MatchLabels: map[string]string{f.UniqueName: "true"},
		//},
	}
}

func (w *WebHookManager) CreateWebhooks(namespace string) {

	mutatingWebhookService := corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "par-webhook", Namespace: namespace}}

	mutatingWebhook := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "par-webhook"},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			w.newMutatingIsReadyWebhookFixture(mutatingWebhookService),
		},
	}

	// TODO: Fix issue with Rbac So mutating Webhook can be created.
	// mutatingwebhookconfigurations.admissionregistration.k8s.io is forbidden: User "system:serviceaccount:par-dev:par-dev-controller-manager"
	// cannot create resource "mutatingwebhookconfigurations" in API group "admissionregistration.k8s.io" at the cluster scope
	err := w.Mgr.GetClient().Create(context.Background(), mutatingWebhook)
	if err != nil {
		return
	}
	// Track Current webhooks for clean up
	w.CurrentWebHooks = append(w.CurrentWebHooks, mutatingWebhook)

	w.Mgr.GetWebhookServer().Register("/mutate-v1-pod", &webhook.Admission{Handler: &PodDnsUpdater{Client: w.Mgr.GetClient()}})

}

func (w *WebHookManager) DeleteWebhook() {
	// TODO: Delete webhook manifest objects
	//	iterate current webhooks and delete the objects

}

type PodDnsUpdater struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (p PodDnsUpdater) Handle(ctx context.Context, request admission.Request) admission.Response {
	pod := &corev1.Pod{}
	err := p.decoder.Decode(request, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	panic("implement me")
}
