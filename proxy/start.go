package proxy

import (
	"context"
	"github.com/jmcgrath207/par/storage"
	corev1 "k8s.io/api/core/v1"
	"net"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var Client client.Client

func Start(client client.Client) {
	Client = client
}

func SetProxyServiceIP(optsClient []client.ListOption) {

	serviceList := &corev1.ServiceList{}
	namespace, _ := os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{"par.dev/proxy": "true"}),
	}

	err := Client.List(context.Background(), serviceList, opts...)
	if err != nil {
		panic(err)
	}

	//TODO: put error logging it can't find service in namespace of par chart
	proxyIP := serviceList.Items[0].Spec.ClusterIP

	var podList corev1.PodList

	Client.List(context.Background(), &podList, optsClient...)

	for _, pod := range podList.Items {
		storage.SourceHostMap[pod.Status.PodIP] = net.ParseIP(proxyIP)
	}
	storage.AcquiredProxyServiceIP <- true

}
