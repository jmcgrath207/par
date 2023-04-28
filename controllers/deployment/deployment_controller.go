package deployment

import (
	"context"
	dnsv1 "github.com/jmcgrath207/par/apis/dns/v1"
	"github.com/jmcgrath207/par/storage"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// SetupWithManager sets up the controller with the Manager.
func (w *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// TODO: filter on namespace and labels
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(w)
}

func (w *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	msg := storage.ArecordQueue.Pop()
	aRecord := msg.ARecord
	dnsServerIP := msg.DnsServerIP
	if dnsServerIP == "" {
		return ctrl.Result{}, nil
	}

	var deployments appsv1.DeploymentList

	// Get all deployments that match the labels and namespace in the A record
	opts := []client.ListOption{
		client.InNamespace(aRecord.Spec.Namespace),
		client.MatchingLabels(aRecord.Spec.Labels),
	}

	log.FromContext(ctx).Info("searching for deployments with labels", "labels", aRecord.Spec.Labels, "namespace", aRecord.Spec.Namespace)
	err := w.List(ctx, &deployments, opts...)
	if err != nil {
		log.FromContext(ctx).Error(err, "could not find deployments with labels", "labels", aRecord.Spec.Labels, "namespace", aRecord.Spec.Namespace)
		return ctrl.Result{}, err
	}

	// Update the deployment's DNS server to point to the service IP address of the Manager
	for _, deployment := range deployments.Items {
		log.FromContext(ctx).Info("found client deployment", "deployment", deployment.Name)
		UpdateDnsClient(deployment, dnsServerIP)
	}
	return ctrl.Result{}, nil

}
func UpdateDnsClient(deployment appsv1.Deployment, dnsIP string) {
	// Update Pods DNS server it points to the service IP address of the Manager
	// TODO: Check in memory cache first if these labels have already been processed.

	deploymentClone := deployment.DeepCopy()

	// Add a new DNS configuration to the deployment's pod template with the updated IP address.
	deploymentClone.Spec.Template.Spec.DNSConfig = &corev1.PodDNSConfig{
		Nameservers: []string{dnsIP},
	}

	deploymentClone.Spec.Template.Spec.DNSPolicy = corev1.DNSNone
	log.FromContext(context.Background()).Info("updating deployment dns policy to point to service dnsIP of par manager", "deployment", deploymentClone.Name, "dnsIP", dnsIP)
	err := storage.ClientK8s.Patch(context.TODO(), deploymentClone, client.MergeFrom(&deployment))
	if err != nil {
		log.FromContext(context.Background()).Error(err, "could not update deployment dns policy to point to service dnsIP of par manager", "deployment", deploymentClone.Name, "dnsIP", dnsIP)
		panic(err)
	}
	log.FromContext(context.Background()).Info("updated deployment dns policy to point to service IP of par manager", "deployment", deploymentClone.Name, "dnsIP", dnsIP)

}

func HostAlias(ctx context.Context, deployment appsv1.Deployment, aRecord dnsv1.Arecord) {
	// Update the deployment object's hostAliases field
	deployment.Spec.Template.Spec.HostAliases = []corev1.HostAlias{
		{
			IP:        aRecord.Spec.IPAddress,
			Hostnames: []string{aRecord.Spec.HostName},
		},
	}

	if err := storage.ClientK8s.Update(ctx, &deployment); err != nil {
		panic(err)
	}
}
