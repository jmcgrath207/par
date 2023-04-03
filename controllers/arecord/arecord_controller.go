/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package arecord

import (
	"context"
	dnsv1 "github.com/jmcgrath207/par/apis/dns/v1"
	"github.com/jmcgrath207/par/proxy"
	"github.com/jmcgrath207/par/storage"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ArecordReconciler reconciles a Arecord object
type ArecordReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dns.par.dev,resources=arecords,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dns.par.dev,resources=arecords/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dns.par.dev,resources=arecords/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// the Arecord object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ArecordReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	var aRecord dnsv1.Arecord

	if err := r.Get(ctx, req.NamespacedName, &aRecord); err != nil {
		// Handle error if the MyResource object cannot be fetched
		if errors.IsNotFound(err) {
			// The MyResource object has been deleted, so we can stop reconciling
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	// Find all deployments that match the labels in of par.dev/manager: true
	// Get the IP address of the service that matches the labels in the A record
	serviceList := &corev1.ServiceList{}
	namespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return ctrl.Result{}, err
	}
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{"par.dev/manager": "true"}),
	}
	err = r.List(ctx, serviceList, opts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	var deployments appsv1.DeploymentList

	// Get all deployments that match the labels and namespace in the A record
	opts = []client.ListOption{
		client.InNamespace(aRecord.Spec.Namespace),
		client.MatchingLabels(aRecord.Spec.Labels),
	}

	err = r.List(ctx, &deployments, opts...)
	if err != nil {
		return ctrl.Result{}, err
	}

	// Update the deployment's DNS server to point to the service IP address of the Manager
	for _, deployment := range deployments.Items {
		r.UpdateDnsClient(deployment, serviceList.Items[0].Spec.ClusterIP)
	}
	_ = log.FromContext(ctx)

	storage.SetRecord("A", aRecord.Spec.HostName, aRecord.Spec.IPAddress)

	proxy.SetProxyServiceIP(opts)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ArecordReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dnsv1.Arecord{}).
		Complete(r)
}

func (r *ArecordReconciler) UpdateDnsClient(deployment appsv1.Deployment, dnsIP string) {
	// Update Pods DNS server it points to the service IP address of the Manager
	// TODO: Check in memory cache first if these labels have already been processed.

	deploymentClone := deployment.DeepCopy()

	// Add a new DNS configuration to the deployment's pod template with the updated IP address.
	deploymentClone.Spec.Template.Spec.DNSConfig = &corev1.PodDNSConfig{
		Nameservers: []string{dnsIP},
	}

	deploymentClone.Spec.Template.Spec.DNSPolicy = corev1.DNSNone
	err := r.Patch(context.TODO(), deploymentClone, client.MergeFrom(&deployment))
	if err != nil {
		panic(err)
	}

}

func (r *ArecordReconciler) HostAlias(ctx context.Context, deployment appsv1.Deployment, aRecord dnsv1.Arecord) {
	// Update the deployment object's hostAliases field
	deployment.Spec.Template.Spec.HostAliases = []corev1.HostAlias{
		{
			IP:        aRecord.Spec.IPAddress,
			Hostnames: []string{aRecord.Spec.HostName},
		},
	}

	if err := r.Client.Update(ctx, &deployment); err != nil {
		panic(err)
	}
}

//func getPodByIP(client client.Client, podIP string) (*corev1.Pod, error) {
//	podList := &corev1.PodList{}
//	err := client.List(context.Background(), podList)
//	if err != nil {
//		return nil, err
//	}
//
//	for _, pod := range podList.Items {
//		if pod.Status.PodIP == podIP {
//			return &pod, nil
//		}
//	}
//
//	return nil, fmt.Errorf("pod with IP %s not found", podIP)
//}
