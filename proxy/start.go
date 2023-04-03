package proxy

import (
	"context"
	"github.com/jmcgrath207/par/storage"
	corev1 "k8s.io/api/core/v1"
	"net"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var k8sClient client.Client
var namespace []byte

func Start(clientK8s client.Client) {
	namespace, _ = os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	// TODO: keeps on requests are cluster level and not namespaced
	k8sClient = client.NewNamespacedClient(clientK8s, string(namespace))
}

func SetProxyServiceIP(optsClient []client.ListOption) {
	// Set the proxy service IP in the storage map when source pod is requested

	serviceList := &corev1.ServiceList{}
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{"par.dev/proxy": "true"}),
	}

	err := k8sClient.List(context.Background(), serviceList, opts...)
	if err != nil {
		panic(err)
	}

	//TODO: put error logging it can't find service in namespace of par chart
	proxyIP := serviceList.Items[0].Spec.ClusterIP

	var podList corev1.PodList

	k8sClient.List(context.Background(), &podList, optsClient...)

	for _, pod := range podList.Items {
		storage.SourceHostMap[pod.Status.PodIP] = net.ParseIP(proxyIP)
	}

	renderProxyConfig()
	storage.ProxyReady <- true

}

func renderProxyConfig() {

	serviceList := &corev1.ConfigMapList{}

	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{"par.dev/proxy-config": "true"}),
	}

	// TODO: keeps on requests are cluster level and not namespaced
	err := k8sClient.List(context.Background(), serviceList, opts...)
	if err != nil {
		panic(err)
	}

}
