package deployment

import (
	"context"
	dnsv1 "github.com/jmcgrath207/par/apis/dns/v1"
	"github.com/jmcgrath207/par/storage"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

type DeploymentReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	aRecord dnsv1.Arecord
}

// Define a predicate function to filter out unwanted events
func (w *DeploymentReconciler) deploymentPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if w.aRecord.Spec.Namespace == e.Object.GetNamespace() {
				if reflect.DeepEqual(w.aRecord.Spec.Labels, e.Object.GetLabels()) {
					return true
				}
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if w.aRecord.Spec.Namespace == e.ObjectNew.GetNamespace() {
				if reflect.DeepEqual(w.aRecord.Spec.Labels, e.ObjectNew.GetLabels()) {
					return true
				}
			}
			return false
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (w *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager, aRecord dnsv1.Arecord) error {
	w.aRecord = aRecord
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Named(aRecord.Name).
		WithEventFilter(w.deploymentPredicate()).
		Complete(w)
}

func (w *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// after processing a New Deployment.
	deployment := &appsv1.Deployment{}
	w.Get(ctx, req.NamespacedName, deployment)
	UpdateDnsClient(*deployment, w.aRecord.Spec.ManagerAddress)
	return ctrl.Result{}, nil
}

func UpdateDnsClient(deployment appsv1.Deployment, dnsIP string) {
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
