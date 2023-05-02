package deployment

import (
	"context"
	"fmt"
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
	Scheme *runtime.Scheme
}

// Define a predicate function to filter out unwanted events
func deploymentPredicate(aRecord dnsv1.Arecord) predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if aRecord.Spec.Namespace == e.Object.GetNamespace() {
				if reflect.DeepEqual(aRecord.Spec.Labels, e.Object.GetLabels()) {
					return true
				}
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if aRecord.Spec.Namespace == e.ObjectNew.GetNamespace() {
				if reflect.DeepEqual(aRecord.Spec.Labels, e.ObjectNew.GetLabels()) {
					return true
				}
			}
			return false
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (w *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager, aRecord dnsv1.Arecord) error {
	// TODO: filter on namespace and labels
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Named("asdfasdf").
		WithEventFilter(deploymentPredicate(aRecord)).
		Complete(w)
}

func (w *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// after processing a New Deployment.
	deployment := &appsv1.Deployment{}
	aRecord := dnsv1.Arecord{}

	w.Get(ctx, req.NamespacedName, deployment)
	deploymentAnnotations := deployment.GetAnnotations()
	value, ok := deploymentAnnotations["par.dev/recordId"]
	if !ok {
		// TODO: work on Arecord Map lookup to get DNS Manager Address.
		aRecord = storage.ArecordMap
		UpdateDnsClient(*deployment, aRecord.Spec.ManagerAddress)
		return ctrl.Result{}, nil
	}

	aRecordList := dnsv1.ArecordList{}
	// Create a client.MatchingLabels object with the annotation key and value
	fieldSelector := client.MatchingFields{
		"metadata.annotations": fmt.Sprintf("%s=%s", "par.dev/recordId", value),
	}
	err := w.List(ctx, &aRecordList, fieldSelector)
	if err != nil {
		//log.FromContext(ctx).Error(err, "could not find deployments with labels", "labels", aRecord.Spec.Labels, "namespace", aRecord.Spec.Namespace)
		return ctrl.Result{}, err
	}
	for _, aRecord = range aRecordList.Items {
		UpdateDnsClient(*deployment, aRecord.Spec.ManagerAddress)
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
