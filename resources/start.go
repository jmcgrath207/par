package resources

import (
	"context"
	"encoding/json"
	"github.com/jmcgrath207/par/store"
	"github.com/patrickmn/go-cache"
	"hash/fnv"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"net"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strconv"
	"time"
)

var namespaceLabels map[string][]map[string]string
var hashLabels map[string]ResourcePayload
var resourceObjectCache *cache.Cache

type ResourcePayload struct {
	Namespace        string
	Id               string
	DnsServerAddress string
	Labels           map[string]string
	ForwardType      string
}

func Start() {
	namespaceLabels = make(map[string][]map[string]string)
	hashLabels = make(map[string]ResourcePayload)
	resourceObjectCache = cache.New(1*time.Minute, 1*time.Minute)
}

func Update(deployment appsv1.Deployment) {
	// Only update deployments that have been observed by the records controller.
	labels, ok := namespaceLabels[deployment.Namespace]
	if !ok {
		return
	}

	for _, label := range labels {
		if haveSameKeys(label, deployment.Labels) {
			a := getHashLabels(label)
			if _, found := resourceObjectCache.Get(a); found {
				return
			}
			resourceObjectCache.Set(a, true, cache.DefaultExpiration)
			UpdateDnsClient(deployment, hashLabels[a])

		}
	}
}

func UpdateByNSLabels(namespace string, labels map[string]string) {

	var deploymentList appsv1.DeploymentList
	opts := []client.ListOption{
		client.MatchingLabels(labels),
		client.InNamespace(namespace),
	}
	store.ClientK8s.List(context.Background(), &deploymentList, opts...)
	for _, deployment := range deploymentList.Items {
		Update(deployment)
	}
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

func getHashLabels(labels map[string]string) string {
	out, err := json.Marshal(labels)
	if err != nil {
		panic(err)
	}
	h := fnv.New32a()
	h.Write(out)
	return strconv.Itoa(int(h.Sum32()))
}

func Observe(resourcePayload ResourcePayload) {

	hashLabels[getHashLabels(resourcePayload.Labels)] = resourcePayload
	namespaceLabels[resourcePayload.Namespace] = append(namespaceLabels[resourcePayload.Namespace], resourcePayload.Labels)
}

func UpdateDnsClient(deployment appsv1.Deployment, payload ResourcePayload) error {

	// Add a new DNS configuration to the deployment's pod template with the updated IP address.
	deployment.Spec.Template.Spec.DNSConfig = &corev1.PodDNSConfig{
		Nameservers: []string{payload.DnsServerAddress},
	}

	deployment.Spec.Template.Spec.DNSPolicy = corev1.DNSNone
	log.FromContext(context.Background()).Info("updating deployment dns policy to point to service dnsIP of par manager",
		"deployment", deployment.Name, "dnsIP", payload.DnsServerAddress)

	go setClientData(payload)
	return nil
}

func setClientData(payload ResourcePayload) {
	// Wait for a Deployments pods to come up and collect their IP address for DNS forwarding decisions.
	var podList corev1.PodList
	opts := []client.ListOption{
		client.MatchingLabels(payload.Labels),
		client.InNamespace(payload.Namespace),
	}

	for {
		status := 0
		store.ClientK8s.List(context.Background(), &podList, opts...)

		for _, pod := range podList.Items {
			if pod.Status.Phase != "Running" || pod.Spec.DNSConfig == nil {
				continue
			}

			if pod.Spec.DNSConfig.Nameservers[0] == payload.DnsServerAddress {

				switch payload.ForwardType {
				case "manager":
					store.ClientId[pod.Status.PodIP] = payload.Id
				case "proxy":
					store.ProxyWaitGroup.Wait()
					store.ToProxySourceHostMap[pod.Status.PodIP] = net.ParseIP(store.ProxyAddress)
				}
				status = 1

			}
		}
		if status == 1 {
			break
		}
	}
	store.DNSWaitGroup.Done()
}
