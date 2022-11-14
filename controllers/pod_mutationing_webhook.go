package controllers

import (
	admissionregv1 "k8s.io/api/admissionregistration/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type DynamicWebHooks struct {
	mgr        manager.Manager
	hookServer *webhook.Server
	decoder    *admission.Decoder
}

type PodDnsUpdater struct {
	Client  client.Client
	decoder *admission.Decoder
}

func (d *DynamicWebHooks) SetupWithManager(mgr manager.Manager) error {
	d.mgr = mgr
	d.hookServer = mgr.GetWebhookServer()
	return nil
}
func (d *DynamicWebHooks) createWebhook() {
	//# https://pkg.go.dev/k8s.io/kubernetes/pkg/apis/admissionregistration#MutatingWebhook
	// create mutating webhook spec
	//mutatingWebhookConfig := admissionregv1.MutatingWebhook{}

	hookServer.Register("/mutate-v1-pod", &webhook.Admission{Handler: &podAnnotator{Client: mgr.GetClient()}})
	d.hookServer.Register()

}

func (d *DynamicWebHooks) deleteWebhook() {}

func (d *DynamicWebHooks) initWebhook() {
	d.createWebhook()
	defer d.deleteWebhook()
	PodPatch{}.patch()
}
