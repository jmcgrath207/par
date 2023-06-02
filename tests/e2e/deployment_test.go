package e2e

import (
	"context"
	"fmt"
	dnsv1alpha1 "github.com/jmcgrath207/par/apis/dns/v1alpha1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"strings"
	"testing"
	"time"
)

// REF: https://superorbital.io/blog/testing-production-controllers/
// REF: https://github.com/superorbital/random-number-controller

var (
	timeout   = time.Second * 10
	duration  = time.Second * 10
	interval  = time.Millisecond * 250
	clientset *kubernetes.Clientset
	k8sClient client.Client
	namespace = "default"
	records   *dnsv1alpha1.Records
)

func boolPointer(b bool) *bool {
	return &b
}

func cleanupResource(object client.Object) {
	err := k8sClient.Delete(context.Background(), object)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
}

func addRecords() *dnsv1alpha1.Records {
	dnsv1alpha1.AddToScheme(scheme.Scheme)
	yamlFile, err := os.ReadFile("../resources/test_dns_v1alpha1_records.yaml")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	records := &dnsv1alpha1.Records{}
	_, _, err = decode(yamlFile, nil, records)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	err = k8sClient.Create(context.Background(), records)
	gomega.Expect(err).Should(gomega.Succeed())
	return records
}

func createDeployment(deploymentPath string) *appsv1.Deployment {
	yamlFile, err := os.ReadFile(deploymentPath)
	if err != nil {
		fmt.Println(err)
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	deployment := &appsv1.Deployment{}
	_, _, err = decode(yamlFile, nil, deployment)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	gomega.Expect(k8sClient.Create(context.Background(), deployment)).Should(gomega.Succeed())
	return deployment
}

func GetManagerAddress() string {
	// Find all services that match the labels in of par.dev/manager: true
	serviceList := &corev1.ServiceList{}
	opts := []client.ListOption{
		client.InNamespace("par"),
		client.MatchingLabels(map[string]string{"par.dev/manager": "true"}),
	}

	err := k8sClient.List(context.Background(), serviceList, opts...)
	if err != nil {
		ginkgo.Fail("could not find par manager service")
	}
	return serviceList.Items[0].Spec.ClusterIP
}
func GetProxyAddress() string {
	serviceList := &corev1.ServiceList{}
	opts := []client.ListOption{
		client.InNamespace("par"),
		client.MatchingLabels(map[string]string{"par.dev/proxy": "true"}),
	}

	err := k8sClient.List(context.Background(), serviceList, opts...)
	if err != nil {
		ginkgo.Fail("could not find par proxy service")
	}
	return serviceList.Items[0].Spec.ClusterIP

}
func CheckPodLogsFromDeployment(namespace string, labels map[string]string, checkSlice []string) {
	// TODO: add timeout param
	ifFound := make(map[string]bool)
	var fail int
	var checkOuput string

	podList := corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(labels),
	}
	k8sClient.List(context.Background(), &podList, opts...)
	gomega.Expect(len(podList.Items)).Should(gomega.BeNumerically(">", 0))

	for _, pod := range podList.Items {
		req := clientset.CoreV1().Pods(pod.ObjectMeta.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			Container: pod.Spec.Containers[0].Name,
		})

		podLogs, err := req.Stream(context.Background())
		if err != nil {
			ginkgo.GinkgoWriter.Printf("Pod name: \n %v", pod.Name)
			ginkgo.Fail("Unable to get pod logs")
		}
		defer podLogs.Close()

		buffer := make([]byte, 1024)
		for {
			bytesRead, err := podLogs.Read(buffer)
			if err != nil {
				break
			}
			if bytesRead > 0 {
				output := string(buffer[:bytesRead])
				checkOuput = checkOuput + output
				//ginkgo.GinkgoWriter.Printf("checkSlice %v \n", checkSlice)
				for _, a := range checkSlice {
					if ifFound[a] {
						continue
					}
					if strings.Contains(output, a) {
						ifFound[a] = true
						//ginkgo.GinkgoWriter.Printf("Found it %v \n", a)
						continue
					} else {
						ifFound[a] = false
					}
				}
				continue
			}
		}
	}

	for key, value := range ifFound {
		if value {
			ginkgo.GinkgoWriter.Printf("Found value %v in pod logs\n", key)
			continue
		}
		ginkgo.GinkgoWriter.Printf("Did not find value %v in pod logs\n", key)
		ginkgo.GinkgoWriter.Printf("Pod logs output: \n %v", checkOuput)
		fail = 1
	}
	gomega.Expect(fail).Should(gomega.Equal(0))

}

// TODO: add test following test
// kill manager pods and make sure it works - working
// update a entry a Arecord ip address entry - working
// remove a Arecord entry and make sure it's evicted from DNS cache - working
var _ = ginkgo.Describe("Test Deployments\n", func() {

	ginkgo.Context("A Record: wget with PROXY\n", func() {
		createDeployment("../resources/test_wget_a_record_deployment.yaml")
		time.Sleep(120 * time.Second)
		ginkgo.Specify("\nReturn A Record IP addresses and Proxy IP address", func() {
			var checkSlice []string
			checkSlice = append(checkSlice, "google.com", GetProxyAddress(), "Found A record in storage for Proxy",
				records.Spec.A[0].IPAddresses[0], records.Spec.A[0].IPAddresses[1])
			CheckPodLogsFromDeployment("par", map[string]string{"par.dev/manager": "true"}, checkSlice)
		})
	})

	ginkgo.Context("No Record: wget from PROXY\n", func() {
		createDeployment("../resources/test_wget_no_record_deployment.yaml")
		time.Sleep(120 * time.Second)
		ginkgo.Specify("\nReturn A Record Upstream IP addresses and Proxy IP address", func() {
			var checkSlice []string
			checkSlice = append(checkSlice, "yahoo.com", "Found A record in storage for Proxy", GetProxyAddress())
			CheckPodLogsFromDeployment("par", map[string]string{"par.dev/manager": "true"}, checkSlice)
		})
	})

	ginkgo.Context("A Record: Lookup from Manager", func() {
		deployment := createDeployment("../resources/test_a_record_deployment.yaml")
		time.Sleep(10 * time.Second)
		ginkgo.Specify("\nReturn A Record IP addresses and Manager IP address", func() {
			var checkSlice []string
			checkSlice = append(checkSlice, "google.com",
				records.Spec.A[1].IPAddresses[0], records.Spec.A[1].IPAddresses[1])
			CheckPodLogsFromDeployment(namespace, deployment.Spec.Template.ObjectMeta.Labels, checkSlice)
		})
	})

	ginkgo.Context("No Record: lookup from Manager\n", func() {
		deployment := createDeployment("../resources/test_no_record_deployment.yaml")
		time.Sleep(10 * time.Second)
		ginkgo.Specify("\nReturn IP addresses from Upstream DNS and Manager IP address\n", func() {
			var checkSlice []string
			checkSlice = append(checkSlice, "yahoo.com", GetManagerAddress())
			CheckPodLogsFromDeployment(namespace, deployment.Spec.Template.ObjectMeta.Labels, checkSlice)
		})
	})
})

func TestDeployments(t *testing.T) {
	// https://onsi.github.io/ginkgo/#ginkgo-and-gomega-patterns
	time.Sleep(20 * time.Second)
	gomega.RegisterFailHandler(ginkgo.Fail)
	//set up a client
	env := envtest.Environment{
		UseExistingCluster: boolPointer(true),
	}
	config, err := env.Start()
	clientset, err = kubernetes.NewForConfig(config)
	k8sClient, err = client.New(config, client.Options{Scheme: scheme.Scheme})
	records = addRecords()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	ginkgo.RunSpecs(t, "Test Deployments")
}
