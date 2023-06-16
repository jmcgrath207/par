package deployment

import (
	"context"
	"fmt"
	"github.com/jmcgrath207/par/storage"
	"github.com/patrickmn/go-cache"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"time"
)

type DeploymentReconciler struct {
	client.Client
	Scheme              *runtime.Scheme
	deploymentNameCache *cache.Cache
	controllerName      string
	dnsServerAddress    string
	namespace           string
	labels              map[string]string
	id                  string
	clientIDSet         int
	fowardType          string
}

func haveSameKeys(map1, map2 map[string]string) bool {
	// Check if each key-value pair in map1 is present in map2
	for key, value := range map1 {
		if val, ok := map2[key]; !ok || val != value {
			return false
		}
	}
	return true
}

// Define a predicate function to filter out unwanted events
func (w *DeploymentReconciler) deploymentPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if w.namespace == e.Object.GetNamespace() {
				log.FromContext(context.Background()).Info("Reconcile Create", "deployment", e.Object.GetName(), "controller", w.controllerName)
				return haveSameKeys(w.labels, e.Object.GetLabels())
			}
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			if w.namespace == e.ObjectNew.GetNamespace() {
				if haveSameKeys(w.labels, e.ObjectNew.GetLabels()) {
					_, value := w.deploymentNameCache.Get(e.ObjectNew.GetName())
					if !value {
						log.FromContext(context.Background()).Info("Reconcile Update", "deployment", e.ObjectNew.GetName(), "controller", w.controllerName)
						return true
					}

				}
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			log.FromContext(context.Background()).Info("Reconcile Update", "deployment", e.Object.GetName(), "controller", w.controllerName)
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			log.FromContext(context.Background()).Info("Reconcile Generic", "deployment", e.Object.GetName(), "controller", w.controllerName)
			return true
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (w *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager, dnsServerAddress string, namespace string, name string, labels map[string]string, id string, forwardType string) error {
	w.deploymentNameCache = cache.New(30*time.Second, 1*time.Minute)
	w.fowardType = forwardType
	w.dnsServerAddress = dnsServerAddress
	w.namespace = namespace
	w.labels = labels
	w.id = id
	w.controllerName = fmt.Sprintf(name + " deployment")
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Named(w.controllerName).
		WithEventFilter(
			predicate.And(
				w.deploymentPredicate(),
				predicate.Or(
					predicate.GenerationChangedPredicate{},
					predicate.LabelChangedPredicate{},
				))).
		Complete(w)
}

func (w *DeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// after processing a New Deployment.
	deployment := &appsv1.Deployment{}
	w.Get(ctx, req.NamespacedName, deployment)
	return w.UpdateDnsClient(*deployment)

}

func (w *DeploymentReconciler) UpdateDnsClient(deployment appsv1.Deployment) (ctrl.Result, error) {

	w.deploymentNameCache.Set(deployment.Name, 1, cache.DefaultExpiration)
	deploymentClone := deployment.DeepCopy()

	// Add a new DNS configuration to the deployment's pod template with the updated IP address.
	deploymentClone.Spec.Template.Spec.DNSConfig = &corev1.PodDNSConfig{
		Nameservers: []string{w.dnsServerAddress},
	}

	deploymentClone.Spec.Template.Spec.DNSPolicy = corev1.DNSNone
	log.FromContext(context.Background()).Info("updating deployment dns policy to point to service dnsIP of par manager",
		"deployment", deploymentClone.Name, "dnsIP", w.dnsServerAddress)

	err := storage.ClientK8s.Patch(context.TODO(), deploymentClone, client.MergeFrom(&deployment))
	if err != nil {
		log.FromContext(context.Background()).Error(err, "could not update deployment dns policy to point to service dnsIP of par manager",
			"deployment", deploymentClone.Name, "dnsIP", w.dnsServerAddress)

		return ctrl.Result{}, err
	}
	log.FromContext(context.Background()).Info("updated deployment dns policy to point to service IP of par manager",
		"deployment", deploymentClone.Name, "dnsIP", w.dnsServerAddress)

	w.SetClientData(int(*deployment.Spec.Replicas), deployment.Spec.Template.Labels)
	return ctrl.Result{}, nil
}

func (w *DeploymentReconciler) SetClientData(replicas int, labels map[string]string) {
	var podList corev1.PodList
	opts := []client.ListOption{
		client.MatchingLabels(labels),
		client.InNamespace(w.namespace),
	}

	for {
		w.List(context.Background(), &podList, opts...)
		count := 0
		podCount := len(podList.Items)
		if podCount != replicas {
			continue
		}

		for _, pod := range podList.Items {
			if pod.Status.Phase != "Running" {
				break
			}
			if w.fowardType == "manager" {
				storage.ClientId[pod.Status.PodIP] = w.id
			}
			if w.fowardType == "proxy" {
				log.FromContext(context.Background()).Info("Setting DNS Server to return only proxy IP for source pod", "pod", pod.Name, "proxyIP", storage.ProxyAddress)
				storage.ProxyWaitGroup.Wait()
				storage.ToProxySourceHostMap[pod.Status.PodIP] = net.ParseIP(storage.ProxyAddress)
			}
			count = count + 1
		}
		if count == replicas {
			break
		}
	}
}

// TODO: Add Host Alias Feature later
//func HostAlias(ctx context.Context, deployment appsv1.Deployment, aRecord dnsv1.Arecord) {
//	// Update the deployment object's hostAliases field
//	deployment.Spec.Template.Spec.HostAliases = []corev1.HostAlias{
//		{
//			IP:        aRecord.Spec.IPAddress,
//			Hostnames: []string{aRecord.Spec.HostName},
//		},
//	}
//
//	if err := storage.ClientK8s.Update(ctx, &deployment); err != nil {
//		panic(err)
//	}
//}
