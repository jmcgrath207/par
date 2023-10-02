package webhook

import (
	"context"
	"fmt"
	"github.com/jmcgrath207/par/resources"
	"github.com/open-policy-agent/cert-controller/pkg/rotator"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

var _ webhook.CustomDefaulter = &DeploymentUpdate{}

type DeploymentUpdate struct{}

func (d *DeploymentUpdate) Default(ctx context.Context, obj runtime.Object) error {
	deployment, _ := obj.(*appsv1.Deployment)
	resources.Update(*deployment)
	return nil
}

// Based on this code snippet
// https://github.com/metallb/metallb/blob/main/internal/k8s/webhook.go#L18
func Start(mgr manager.Manager) {
	webHookCertRdy := make(chan struct{})
	log := logf.FromContext(context.Background())
	webhooks := []rotator.WebhookInfo{
		{
			Name: "par-mutating-webhook",
			Type: rotator.Mutating,
		},
	}

	namespace, _ := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")

	err := rotator.AddRotator(mgr, &rotator.CertRotator{
		SecretKey: types.NamespacedName{
			Namespace: string(namespace),
			Name:      "webhook-server-certs",
		},
		CertDir:        "/tmp/k8s-webhook-server/serving-certs",
		CAName:         "par",
		CAOrganization: "par",
		DNSName:        fmt.Sprintf("%s.%s.svc", "par-manager-webhook", string(namespace)),
		Webhooks:       webhooks,
		IsReady:        webHookCertRdy,
		//RestartOnSecretRefresh: true,
	})
	if err != nil {
		log.Error(err, "cert rotation failed")
	}

	go func() {
		<-webHookCertRdy
		if err := builder.WebhookManagedBy(mgr).
			For(&appsv1.Deployment{}).
			WithDefaulter(&DeploymentUpdate{}).
			Complete(); err != nil {
			log.Error(err, "unable to create webhook", "webhook", "Deployment")
			os.Exit(1)
		}

	}()
}
