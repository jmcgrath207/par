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

func CheckPodLogsFromDeployment(deployment *appsv1.Deployment, checkSlice []string) {
	ifFound := make(map[string]bool)
	var fail int
	var checkOuput string

	podList := corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(deployment.Spec.Template.ObjectMeta.Labels),
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

		buffer := make([]byte, 512)
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
// kill manager pods and make sure it works
// update a entry a Arecord ip address entry
// remove a Arecord entry and make sure it's evicted from DNS cache
// add HostAlias Test
// add Non Proxy Test
// Add Cname Test
var _ = ginkgo.Describe("Test Deployments on Arecords\n", func() {

	ginkgo.Context("Test Deployment that queries a domain that is IN ARecord\n", func() {
		deployment := createDeployment("../resources/test_a_record_deployment.yaml")
		//defer cleanupResource(deployment)
		// TODO: check logs on manager if deployment is ready
		time.Sleep(10 * time.Second)
		ginkgo.Specify("Does return a Proxy IP address upon DNS lookup from Par Manager Address\n", func() {
			var checkSlice []string
			checkSlice = append(checkSlice, "google.com", records.Spec.A[0].IPAddress)
			CheckPodLogsFromDeployment(deployment, checkSlice)
		})
	})

	ginkgo.Context("Test Deployment that queries a domain that is IN ARecord with PROXY\n", func() {
		deployment := createDeployment("../resources/test_wget_a_record_deployment.yaml")
		//defer cleanupResource(deployment)
		// TODO: check logs on manager if deployment is ready
		time.Sleep(10 * time.Second)
		ginkgo.Specify("Does return a Proxy IP address upon DNS lookup from Par Manager Address\n", func() {
			var checkSlice []string
			// TODO: how to get the cluster dns address and add it to the checkSlice
			checkSlice = append(checkSlice, "google.com")
			CheckPodLogsFromDeployment(deployment, checkSlice)
		})
	})

	ginkgo.Context("Test Deployment that queries a domain NOT IN ARecord with PROXY\n", func() {
		deployment := createDeployment("../resources/test_wget_no_record_deployment.yaml")
		//defer cleanupResource(deployment)
		// TODO: check logs on manager if deployment is ready
		time.Sleep(10 * time.Second)
		ginkgo.Specify("Does not return a Proxy IP address upon DNS lookup from Par Manager Address, only Upstream DNS\n", func() {
			var checkSlice []string
			// TODO: Query Cluster DNS and compare with CheckSlice
			checkSlice = append(checkSlice, "yahoo.com")
			CheckPodLogsFromDeployment(deployment, checkSlice)
		})
	})

	ginkgo.Context("Test Deployment that queries a domain NOT IN ARecord\n", func() {
		deployment := createDeployment("../resources/test_no_record_deployment.yaml")
		//defer cleanupResource(deployment)
		// TODO: check logs on manager if deployment is ready
		time.Sleep(10 * time.Second)
		ginkgo.Specify("Does not return a Proxy IP address upon DNS lookup from Par Manager Address, only Upstream DNS\n", func() {
			var checkSlice []string
			checkSlice = append(checkSlice, "yahoo.com", GetManagerAddress())
			CheckPodLogsFromDeployment(deployment, checkSlice)
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
