package controllers

import (
	"context"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
				Resources:   []string{"configmaps"}, //TODO Should this be pods?
			},
		}},
		ClientConfig: admissionregistrationv1.WebhookClientConfig{
			Service: &admissionregistrationv1.ServiceReference{
				Namespace: service.Namespace,
				Name:      service.Name,
				Path:      strPtr("/mutate-v1-pod-par-dev"),
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

	// TODO make sure ports matches webhook
	mutatingWebhookService := corev1.Service{ObjectMeta: metav1.ObjectMeta{Name: "par-webhook", Namespace: namespace}}

	// Pass service port to mutating webhook creation
	mutatingWebhook := &admissionregistrationv1.MutatingWebhookConfiguration{
		ObjectMeta: metav1.ObjectMeta{Name: "par-webhook"},
		Webhooks: []admissionregistrationv1.MutatingWebhook{
			w.newMutatingIsReadyWebhookFixture(mutatingWebhookService),
		},
	}

	err := w.Mgr.GetClient().Create(context.Background(), mutatingWebhook)
	if err != nil {
		if errors.IsAlreadyExists(err) {
			err := w.Mgr.GetClient().Update(context.Background(), mutatingWebhook)
			// TODO: having issue doing a update if object fails to create.
			// mutatingwebhookconfigurations.admissionregistration.k8s.io "par-webhook" is invalid: metadata.resourceVersion: Invalid value: 0x0: must be specified for an update
			if err != nil {
				panic(err)
			}
		}
		panic(err)
	}

	// Track Current webhooks for clean up
	w.CurrentWebHooks = append(w.CurrentWebHooks, mutatingWebhook)

	w.Mgr.GetWebhookServer().Register(mutatingWebhook.Webhooks[0].ClientConfig.Service.Path, &webhook.Admission{Handler: &PodDnsUpdater{Client: w.Mgr.GetClient()}})

}

func (w *WebHookManager) DeleteWebhook() {
	// TODO: Delete webhook manifest objects
	//	iterate current webhooks and delete the objects

}

type PodDnsUpdater struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (p *PodDnsUpdater) InjectDecoder(d *admission.Decoder) error {
	p.decoder = d
	return nil
}

func (p PodDnsUpdater) Handle(ctx context.Context, request admission.Request) admission.Response {
	pod := &corev1.Pod{}
	err := p.decoder.Decode(request, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	panic("implement me")
}
