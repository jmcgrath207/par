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
	"fmt"
	"github.com/google/uuid"
	dnsv1alpha1 "github.com/jmcgrath207/par/apis/dns/v1alpha1"
	"github.com/jmcgrath207/par/controllers/deployment"
	"github.com/jmcgrath207/par/storage"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"os"
	"reflect"
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
		return ctrl.Result{}, nil
	}
	var records dnsv1alpha1.Records

	if err := r.Get(ctx, req.NamespacedName, &records); err != nil {
		// Handle error if the MyResource object cannot be fetched
		if errors.IsNotFound(err) {
			// The MyResource object has been deleted, so we can stop reconciling
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err

	}
	r.UpdateRecords(ctx, records)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *RecordsReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dnsv1alpha1.Records{}).
		Complete(r)
}

func (r *RecordsReconciler) BackFillRecords(ctx context.Context) (ctrl.Result, error) {
	// Gather existing Arecords in cluster and create a controller for them

	recordsList := dnsv1alpha1.RecordsList{}
	// Create a client.MatchingLabels object with the annotation key and value
	err := r.List(ctx, &recordsList)
	if err != nil {
		return ctrl.Result{}, err
	}

	for _, records := range recordsList.Items {
		r.UpdateRecords(ctx, records)
	}

	return ctrl.Result{}, nil
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
}

func (r *RecordsReconciler) UpdateRecords(ctx context.Context, records dnsv1alpha1.Records) {

	val := reflect.ValueOf(records.Spec)
	for i := 0; i < val.NumField(); i++ {
		attrName := val.Type().Field(i).Name
		// TODO: make this better
		if attrName == "A" {
			count := 1
			for _, x := range records.Spec.A {
				u, _ := uuid.NewRandom()
				id := u.String()
				storage.SetRecord(attrName, x.HostName+"."+id, x)
				log.FromContext(ctx).Info("Reconciling record", "Record Type", attrName, "Hostname", x.HostName)
				r.InvokeDeploymentManager(ctx, managerAddress, records.ObjectMeta.Namespace, fmt.Sprintf("A record "+string(rune(count))), x.Labels, id, x.ForwardType)
				count = count + 1
			}
		}
	}

}
func (r *RecordsReconciler) InvokeDeploymentManager(ctx context.Context, dnsServerAddress string, namespace string, name string, labels map[string]string, id string, forwardType string) {
	if err := (&deployment.DeploymentReconciler{
		Client: storage.Mgr.GetClient(),
		Scheme: storage.Mgr.GetScheme(),
	}).SetupWithManager(storage.Mgr, dnsServerAddress, namespace, name, labels, id, forwardType); err != nil {
		log.FromContext(ctx).Error(err, "unable to create controller", "controller", "Deployment"+name)
		os.Exit(1)
	}
}
