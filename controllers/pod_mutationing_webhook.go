package controllers

import (
	"context"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type WebHookManager struct {
	Mgr    manager.Manager
	Client client.Client
}

func (w *WebHookManager) CreateWebhooks(namespace string) {

	mutatingwebhook := &admissionregistrationv1.MutatingWebhookConfiguration{}
	w.Client.Get(context.Background(), client.ObjectKey{
		Namespace: os.Getenv("CURRENT_NAMESPACE"),
		Name:      os.Getenv("WEBHOOK_NAME"),
	}, mutatingwebhook)

	// TODO: get Webhook path from object.
	//w.Mgr.GetWebhookServer().Register(mutatingwebhook.Webhooks, &webhook.Admission{Handler: &PodDnsUpdater{Client: w.Mgr.GetClient()}})
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
	// TODO: handle not picking up when pod oject is applied.
	pod := &corev1.Pod{}
	err := p.decoder.Decode(request, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	panic("implement me")
}
