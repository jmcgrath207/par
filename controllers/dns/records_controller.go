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

package dns

import (
	"context"
	"github.com/jmcgrath207/par/controllers/deployment"
	"github.com/jmcgrath207/par/dns/types"
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

var managerAddress string
var initReconcile int

// RecordsReconciler reconciles a Records object
type RecordsReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=dns.par.dev,resources=records,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=dns.par.dev,resources=records/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=dns.par.dev,resources=records/finalizers,verbs=update

func (r *RecordsReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	// Needs to happen here since the Read Cache of the client vaild until Reconcile is Invoked.
	// https://sdk.operatorframework.io/docs/building-operators/golang/references/client/#default-client
	if initReconcile == 0 {
		r.SetManagerAddress(ctx)
		r.BackFillRecords(ctx)
		initReconcile = 1
	}
	var records types.Records

	if err := r.Get(ctx, req.NamespacedName, &records); err != nil {
		// Handle error if the MyResource object cannot be fetched
		if errors.IsNotFound(err) {
			// The MyResource object has been deleted, so we can stop reconciling
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err

	}
	if records.Spec.ManagerAddress == "" {
		r.UpdateRecords(ctx, records)
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RecordsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&types.Records{}).
		Complete(r)
}

func (r *RecordsReconciler) BackFillRecords(ctx context.Context) (ctrl.Result, error) {
	// Gather existing Arecords in cluster and create a controller for them
	recordsList := types.RecordsList{}
	// Create a client.MatchingLabels object with the annotation key and value
	err := r.List(ctx, &recordsList)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, records := range recordsList.Items {
		if records.Spec.ManagerAddress != "" {
			r.UpdateRecords(ctx, records)
		}

	}

	return ctrl.Result{}, nil
}

func (r *RecordsReconciler) InvokeDeploymentManager(ctx context.Context, records types.Records) {
	if err := (&deployment.DeploymentReconciler{
		Client: storage.Mgr.GetClient(),
		Scheme: storage.Mgr.GetScheme(),
	}).SetupWithManager(storage.Mgr, records); err != nil {
		log.FromContext(ctx).Error(err, "unable to create controller", "controller", "Deployment")
		os.Exit(1)
	}
}

func (r *RecordsReconciler) SetManagerAddress(ctx context.Context) {

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
	log.FromContext(ctx).Info("searching for par manager service", "namespace", namespace)

	err = r.List(ctx, serviceList, opts...)
	if err != nil {
		log.FromContext(ctx).Error(err, "could not find par manager service", "namespace", namespace)
		panic(err)
	}
	managerAddress = serviceList.Items[0].Spec.ClusterIP
	log.FromContext(ctx).Info("found service par manager service", "service", serviceList.Items[0].Spec.ClusterIP)
	proxy.SetProxyServiceIP(opts)
}

func (r *RecordsReconciler) UpdateRecords(ctx context.Context, records types.Records) {
	records.Spec.ManagerAddress = managerAddress
	records.Set()
	for _, x := range records.RecordItems {
		storage.SetRecord(x.HostName, x)
		log.FromContext(ctx).Info("Reconciling record", "Record Type", x.RecordType, "Hostname", x.HostName)

	}

	r.InvokeDeploymentManager(ctx, records)
}
