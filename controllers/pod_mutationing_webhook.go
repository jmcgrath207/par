package controllers

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type WebHookManager struct {
	mgr manager.Manager
}

// TODO: kubebuilder is not generate this manifest when make manifest is ran.
// +kubebuilder:webhook:path=/mutate-v1-pod,mutating=true,failurePolicy=fail,groups="",resources=pods,verbs=create;update,versions=v1,name=a.par.jmcgrath207.github.com
func (w *WebHookManager) createWebhooks() {

	w.mgr.GetWebhookServer().Register("/mutate-v1-pod", &webhook.Admission{Handler: &PodDnsUpdater{Client: w.mgr.GetClient()}})

}

func (w *WebHookManager) deleteWebhook() {}

func (w *WebHookManager) InitWebhooks(mgr manager.Manager) error {
	w.mgr = mgr
	w.createWebhooks()
	//defer w.deleteWebhook()

	return nil
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
