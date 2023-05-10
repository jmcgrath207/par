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

var managerAddress string
var initReconcile int

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

	// Needs to happen here since the Read Cache of the client vaild until Reconcile is Invoked.
	// https://sdk.operatorframework.io/docs/building-operators/golang/references/client/#default-client
	if initReconcile == 0 {
		r.SetManagerAddress(ctx)
		r.BackFillArecords(ctx)
		initReconcile = 1
	}

	var aRecord dnsv1.Arecord

	if err := r.Get(ctx, req.NamespacedName, &aRecord); err != nil {
		// Handle error if the MyResource object cannot be fetched
		if errors.IsNotFound(err) {
			// The MyResource object has been deleted, so we can stop reconciling
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}
	if aRecord.Spec.ManagerAddress == "" {
		r.UpdateArecord(ctx, aRecord)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ArecordReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// TODO: Doesn't seem to work until  NewControllerManagedBy is called.
	// Could be timing issues.
	return ctrl.NewControllerManagedBy(mgr).
		For(&dnsv1.Arecord{}).
		Complete(r)
}

func (r *ArecordReconciler) BackFillArecords(ctx context.Context) (ctrl.Result, error) {
	// Gather existing Arecords in cluster and create a controller for them
	aRecordList := dnsv1.ArecordList{}
	// Create a client.MatchingLabels object with the annotation key and value
	err := r.List(ctx, &aRecordList)
	if err != nil {
		return ctrl.Result{}, err
	}
	aRecord := dnsv1.Arecord{}
	for _, aRecord = range aRecordList.Items {
		if aRecord.Spec.ManagerAddress != "" {
			r.UpdateArecord(ctx, aRecord)
		}
	}
	return ctrl.Result{}, nil
}

func (r *ArecordReconciler) InvokeDeploymentManager(ctx context.Context, aRecord dnsv1.Arecord) {
	if err := (&deployment.DeploymentReconciler{
		Client: storage.Mgr.GetClient(),
		Scheme: storage.Mgr.GetScheme(),
	}).SetupWithManager(storage.Mgr, aRecord); err != nil {
		log.FromContext(ctx).Error(err, "unable to create controller", "controller", "Deployment")
		os.Exit(1)
	}
}

func (r *ArecordReconciler) SetManagerAddress(ctx context.Context) {

	// Find all services that match the labels in of par.dev/manager: true
	serviceList := &corev1.ServiceList{}
	namespace, err := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	if err != nil {
		panic(err)
	}
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{"par.dev/manager": "true"}),
	}
	log.FromContext(context.Background()).Info("searching for par manager service", "namespace", namespace)

	err = r.List(ctx, serviceList, opts...)
	if err != nil {
		log.FromContext(context.Background()).Error(err, "could not find par manager service", "namespace", namespace)
		panic(err)
	}
	managerAddress = serviceList.Items[0].Spec.ClusterIP
	log.FromContext(context.Background()).Info("found service par manager service", "service", serviceList.Items[0].Spec.ClusterIP)
	proxy.SetProxyServiceIP(opts)
}

func (r *ArecordReconciler) UpdateArecord(ctx context.Context, aRecord dnsv1.Arecord) {
	aRecord.Spec.ManagerAddress = managerAddress
	storage.SetRecord("A", aRecord.Spec.HostName, aRecord.Spec.IPAddress)
	r.Update(ctx, &aRecord)
	log.FromContext(context.Background()).Info("Reconciling A record", "A record",
		aRecord.Spec.HostName, "IP address", aRecord.Spec.IPAddress,
		"Namespace", aRecord.Spec.Namespace, "Labels", aRecord.Spec.Labels)
	r.InvokeDeploymentManager(ctx, aRecord)
}

//func getPodByIP(client client.Client, podIP string) (*corev1.Pod, error) {
//	podList := &corev1.PodList{}
//	err := client.List(ctx, podList)
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
