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

package controllers

import (
	"context"
	"github.com/google/uuid"
	dnsv1 "github.com/jmcgrath207/par/api/v1"
	apps "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

var controllerSessionID uuid.UUID

func init() {
	controllerSessionID = uuid.New()
}

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

	var deployment apps.DeploymentList

	// TODO: need to for loop this logic
	opts := []client.ListOption{
		client.InNamespace(aRecord.Spec.Namespace),
		client.MatchingLabels(aRecord.Spec.Labels),
	}
	// TODO: E0317 04:22:26.109656      12 reflector.go:140] pkg/mod/k8s.io/client-go@v0.26.0/tools/cache/reflector.go:169: Failed to watch *v1.Deployment: failed to list *v1.Deployment: deployments.apps is forbidden: User "system:serviceaccount:par:par-chart-controller-manager" cannot list resource "deployments" in API group "apps" at the cluster scope
	// code is dying here due to rbac.
	if err := r.List(ctx, &deployment, opts...); err != nil {
		// Handle error if the MyResource object cannot be fetched
		if errors.IsNotFound(err) {
			// The MyResource object has been deleted, so we can stop reconciling
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, err
	}

	_ = log.FromContext(ctx)

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ArecordReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&dnsv1.Arecord{}).
		Complete(r)
}
