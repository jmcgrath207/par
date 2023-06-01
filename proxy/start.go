package proxy

import (
	"bytes"
	"context"
	"github.com/jmcgrath207/par/storage"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"text/template"
	"time"
)

var k8sClient client.Client
var namespace []byte
var proxyIP string

func Start() {
	namespace, _ = os.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
	// TODO: keeps on requests are cluster level and not namespaced in RBAC rules
	//k8sClient = clientK8s
	k8sClient = client.NewNamespacedClient(storage.ClientK8s, string(namespace))
	GetProxyServiceIP()
	//TODO need to pass the managerIP address instead.
}

func setProxyPodsClientId(replicas int) {
	// Set the Client ID to the Pods of the Proxy

	var podList corev1.PodList
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{"par.dev/proxy": "true"}),
	}

	for {
		k8sClient.List(context.Background(), &podList, opts...)
		count := 0
		podCount := len(podList.Items)
		if podCount != replicas {
			time.Sleep(2 * time.Second)
			continue
		}

		for _, pod := range podList.Items {
			if pod.Status.Phase != "Running" {
				count = count - 1
				break
			}
			storage.ClientId[pod.Status.PodIP] = "1"
			count = count + 1
		}
		if count == podCount {
			break
		}
		time.Sleep(5 * time.Second)
	}
}

func GetProxyServiceIP() {
	serviceList := &corev1.ServiceList{}
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{"par.dev/proxy": "true"}),
	}
	for {
		log.FromContext(context.Background()).Info("Looking for proxy service", "namespace", string(namespace), "label", "par.dev/proxy=true")
		err := k8sClient.List(context.Background(), serviceList, opts...)
		if err != nil {
			log.FromContext(context.Background()).Info("Waiting for proxy service to be created", "namespace", string(namespace), "label", "par.dev/proxy=true")
			time.Sleep(5 * time.Second)
			continue
		}

		if serviceList.Items == nil {
			log.FromContext(context.Background()).Info("Waiting for proxy service to be created", "namespace", string(namespace), "label", "par.dev/proxy=true")
			time.Sleep(5 * time.Second)
			continue
		}
		break
	}

	log.FromContext(context.Background()).Info("Found proxy service", "namespace", string(namespace), "label", "par.dev/proxy=true")

	proxyIP = serviceList.Items[0].Spec.ClusterIP
	storage.ProxyAddress = proxyIP

}

func RenderProxyConfig(managerIP string) {
	var buf bytes.Buffer

	// Wait for Proxy IP address
	for {
		if proxyIP != "" {
			break
		}
		time.Sleep(5 * time.Second)
	}

	serviceList := &corev1.ConfigMapList{}

	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(map[string]string{"par.dev/proxy-config": "true"}),
	}

	// TODO: keeps on requests are cluster level and not namespaced
	log.FromContext(context.Background()).Info("Looking for proxy config maps",
		"namespace", string(namespace), "label", "par.dev/proxy-config=true")
	err := k8sClient.List(context.Background(), serviceList, opts...)
	if err != nil {
		log.FromContext(context.Background()).Error(err, "Error getting proxy config maps",
			"namespace", string(namespace), "label", "par.dev/proxy-config=true")
		panic(err)
	}
	log.FromContext(context.Background()).Info("Found proxy config maps", "namespace",
		string(namespace), "label", "par.dev/proxy-config=true")

	// Update config maps with proxy-config label
	for _, configMap := range serviceList.Items {
		for k, v := range configMap.Data {
			templ := template.Must(template.New("").Parse(v))
			templ.Execute(&buf, map[string]interface{}{
				"dnsResolver": managerIP,
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
		if err != nil {
			panic(err)
		}
		log.FromContext(context.Background()).Info("Updated proxy config map with Dns Resolver Ip",
			"namespace", string(namespace), "configMap", configMap.Name,
			"dnsResolver", managerIP, "label", "par.dev/proxy-config=true")
	}

	var deployments appsv1.DeploymentList

	// Get all deployments that have the proxy label
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
			log.FromContext(context.Background()).Error(err, "Error updating deployment",
				"namespace", deployment.Namespace, "name", deployment.Name)
			panic(err)
		}
		log.FromContext(context.Background()).Info("Updated proxy deployment to trigger a restart",
			"namespace", deployment.Namespace, "name", deployment.Name)
		setProxyPodsClientId(int(*deployment.Spec.Replicas))

	}
	storage.ProxyReady <- true
	log.FromContext(context.Background()).Info("Proxy Ready")

}
