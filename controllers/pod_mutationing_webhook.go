package controllers

import (
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type DynamicWebHooks struct {
	mgr manager.Manager
}

func (w *DynamicWebHooks) createWebhook() {}
func (w *DynamicWebHooks) deleteWebhook() {}

func (w *DynamicWebHooks) initWebhook() {
	w.createWebhook()
	defer w.deleteWebhook()
	PodPatch{}.patch()
}

func (w *DynamicWebHooks) SetupWithManager(mgr manager.Manager) error {
	w.mgr = mgr
	return nil
}
