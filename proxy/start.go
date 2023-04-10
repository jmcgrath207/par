package proxy

import (
	"bytes"
	"context"
	"github.com/jmcgrath207/par/storage"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"text/template"
	"time"
)

var k8sClient client.Client
var namespace []byte

func Start(clientK8s client.Client) {
	namespace, _ = os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	// TODO: keeps on requests are cluster level and not namespaced
	//k8sClient = clientK8s
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

	renderProxyConfig(proxyIP)
	storage.ProxyReady <- true

}

func renderProxyConfig(proxyIP string) {

	var buf bytes.Buffer

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

	// Update config maps with proxy IP

	for _, configMap := range serviceList.Items {
		for k, v := range configMap.Data {
			templ := template.Must(template.New("").Parse(v))
			templ.Execute(&buf, map[string]interface{}{
				"dnsResolver": proxyIP,
			})
			configMap.Data[k] = buf.String()
			buf.Reset()
		}

		err = k8sClient.Patch(context.TODO(), &configMap, client.MergeFrom(&corev1.ConfigMap{
			ObjectMeta: v1.ObjectMeta{
				Name:      configMap.Name,
				Namespace: configMap.Namespace,
			},
		}))
	}

	var deployments appsv1.DeploymentList

	// Get all deployments that match the labels and namespace in the A record
	opts = []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{"par.dev/proxy": "true"}),
	}
	err = k8sClient.List(context.Background(), &deployments, opts...)
	if err != nil {
		panic(err)
	}

	for _, deployment := range deployments.Items {

		// Set the annotation to trigger a restart
		annotations := deployment.GetAnnotations()
		if annotations == nil {
			annotations = make(map[string]string)
		}
		annotations["kubectl.kubernetes.io/restartedAt"] = time.Now().Format(time.RFC3339Nano)
		deployment.SetAnnotations(annotations)

		// Update the deployment object
		err = k8sClient.Update(context.TODO(), &deployment)
		if err != nil {
			panic(err)
		}
	}

}
