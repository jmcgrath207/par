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
	"github.com/jmcgrath207/par/controllers/deployment"
	"github.com/jmcgrath207/par/proxy"
	"github.com/jmcgrath207/par/storage"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/uuid"
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

	//TODO add a UUID annotation so we can reference them to a existing deployment upon lookup
	if aRecord.Spec.RecordId == "" {
		aRecord.Spec.RecordId = string(uuid.NewUUID())
	}

	log.FromContext(ctx).Info("Reconciling A record", "A record",
		aRecord.Spec.HostName, "IP address", aRecord.Spec.IPAddress, "Namespace", aRecord.Spec.Namespace, "Labels", aRecord.Spec.Labels)

	// Find all services that match the labels in of par.dev/manager: true
	serviceList := &corev1.ServiceList{}
	namespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		return ctrl.Result{}, err
	}
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{"par.dev/manager": "true"}),
	}
	log.FromContext(ctx).Info("searching for par manager service", "namespace", namespace)

	err = r.List(ctx, serviceList, opts...)
	if err != nil {
		log.FromContext(ctx).Error(err, "could not find par manager service", "namespace", namespace)
	}
	aRecord.Spec.ManagerAddress = serviceList.Items[0].Spec.ClusterIP
	r.Update(context.TODO(), &aRecord)

	if err = (&deployment.DeploymentReconciler{
		Client: storage.Mgr.GetClient(),
		Scheme: storage.Mgr.GetScheme(),
	}).SetupWithManager(storage.Mgr, aRecord); err != nil {
		log.FromContext(ctx).Error(err, "unable to create controller", "controller", "Deployment")
		os.Exit(1)
	}

	log.FromContext(ctx).Info("found service par manager service", "service", serviceList.Items[0].Spec.ClusterIP)

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
