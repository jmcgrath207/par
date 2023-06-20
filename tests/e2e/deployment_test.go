package e2e

import (
	"context"
	"fmt"
	dnsv1alpha1 "github.com/jmcgrath207/par/apis/dns/v1alpha1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"net/http"
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
	clientset *kubernetes.Clientset
	k8sClient client.Client
	namespace = "default"
	records   *dnsv1alpha1.Records
)

func boolPointer(b bool) *bool {
	return &b
}

func addRecords() {
	dnsv1alpha1.AddToScheme(scheme.Scheme)
	yamlFile, err := os.ReadFile("../resources/test_dns_v1alpha1_records.yaml")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	records = &dnsv1alpha1.Records{}
	_, _, err = decode(yamlFile, nil, records)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	err = k8sClient.Create(context.Background(), records)
	gomega.Expect(err).Should(gomega.Succeed())
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

func CheckValues(ifFound map[string]bool) int {
	var status int
	for _, value := range ifFound {
		if value {
			status = 1
			continue
		}
		status = 0
	}
	return status
}

func checkPrometheus(checkSlice []string) {
	var status int
	timeout := time.Second * 180
	startTime := time.Now()
	ifFound := make(map[string]bool)

	// Add values to IFound Map
	for _, a := range checkSlice {
		ifFound[a] = false
	}
	httpClient := &http.Client{}

	for {
		elapsed := time.Since(startTime)
		if elapsed >= timeout {
			break
		}
		status = 0
		time.Sleep(1 * time.Second)

		// Create a new GET request
		req, err := http.NewRequest("GET", "http://127.0.0.1:8080/metrics", nil)
		if err != nil {
			continue
		}

		// Send the request
		resp, err := httpClient.Do(req)
		if err != nil {
			continue
		}

		// Read the response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			continue
		}
		output := string(body)
		resp.Body.Close()

		// Print the response
		for _, a := range checkSlice {
			if ifFound[a] {
				continue
			}
			if strings.Contains(output, a) {
				ifFound[a] = true
				status = CheckValues(ifFound)
			}
		}

		if status == 1 {
			break
		}

	}
	for key, value := range ifFound {
		if value {
			ginkgo.GinkgoWriter.Printf("\nFound value: [ %v ] in prometheus exporter\n", key)
			continue
		}
		ginkgo.GinkgoWriter.Printf("\nDid not find value: [ %v ] in prometheus exporter\n", key)
	}

	gomega.Expect(status).Should(gomega.Equal(1))

}

// TODO: add test following test
// kill manager pods and make sure it works - working
// update a entry a Arecord ip address entry - working
// remove a Arecord entry and make sure it's evicted from DNS cache - working
var _ = ginkgo.Describe("Test Deployments\n", func() {

	ginkgo.Context("A Record: wget with PROXY\n", func() {
		createDeployment("../resources/test_wget_a_record_deployment.yaml")
		ginkgo.Specify("\nReturn A Record IP addresses and Proxy IP address", func() {
			var checkSlice []string
			checkSlice = append(checkSlice, "google.com", GetProxyAddress(),
				records.Spec.A[0].IPAddresses[0], records.Spec.A[0].IPAddresses[1])
			checkPrometheus(checkSlice)
		})
	})

	ginkgo.Context("No Record: wget from PROXY\n", func() {
		createDeployment("../resources/test_wget_no_record_deployment.yaml")
		ginkgo.Specify("\nReturn A Record Upstream IP addresses and Proxy IP address", func() {
			var checkSlice []string
			checkSlice = append(checkSlice, "yahoo.com", GetProxyAddress())
			checkPrometheus(checkSlice)
		})
	})

	ginkgo.Context("A Record: Lookup from Manager", func() {
		createDeployment("../resources/test_a_record_deployment.yaml")
		ginkgo.Specify("\nReturn A Record IP addresses and Manager IP address", func() {
			var checkSlice []string
			checkSlice = append(checkSlice, "google.com",
				records.Spec.A[1].IPAddresses[0], records.Spec.A[1].IPAddresses[1], GetManagerAddress())
			checkPrometheus(checkSlice)
		})
	})

	ginkgo.Context("No Record: lookup from Manager\n", func() {
		createDeployment("../resources/test_no_record_deployment.yaml")
		ginkgo.Specify("\nReturn IP addresses from Upstream DNS and Manager IP address\n", func() {
			var checkSlice []string
			checkSlice = append(checkSlice, "yahoo.com", GetManagerAddress())
			checkPrometheus(checkSlice)
		})
	})
})

func TestDeployments(t *testing.T) {
	// https://onsi.github.io/ginkgo/#ginkgo-and-gomega-patterns
	gomega.RegisterFailHandler(ginkgo.Fail)
	//set up a client
	env := envtest.Environment{
		UseExistingCluster: boolPointer(true),
	}
	config, err := env.Start()
	clientset, err = kubernetes.NewForConfig(config)
	k8sClient, err = client.New(config, client.Options{Scheme: scheme.Scheme})
	addRecords()
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
	ginkgo.RunSpecs(t, "Test Deployments")
}
