package e2e

import (
	"context"
	"fmt"
	dnsv1 "github.com/jmcgrath207/par/apis/dns/v1"
	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
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
)

func boolPointer(b bool) *bool {
	return &b
}

func cleanupResource(object client.Object) {
	err := k8sClient.Delete(context.Background(), object)
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
}

func addArecord() *dnsv1.Arecord {
	dnsv1.AddToScheme(scheme.Scheme)
	yamlFile, err := os.ReadFile("../resources/test_dns_v1_arecord.yaml")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	aRecord := &dnsv1.Arecord{}
	_, _, err = decode(yamlFile, nil, aRecord)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	err = k8sClient.Create(context.Background(), aRecord)
	gomega.Expect(err).Should(gomega.Succeed())
	return aRecord
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

func CheckPodLogsFromDeployment(deployment *appsv1.Deployment, searchSlice []string) {
	ifFound := make(map[string]bool)
	var fail int

	podList := corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(deployment.Spec.Template.ObjectMeta.Labels),
	}
	k8sClient.List(context.Background(), &podList, opts...)

	for _, pod := range podList.Items {
		req := clientset.CoreV1().Pods(pod.ObjectMeta.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
			Container: pod.Spec.Containers[0].Name,
		})
		podLogs, err := req.Stream(context.Background())
		if err != nil {
			log.FromContext(context.Background()).Error(err, "unable to get pod logs")
			os.Exit(1)
		}
		defer podLogs.Close()

		buffer := make([]byte, 1024)
		for {
			bytesRead, err := podLogs.Read(buffer)
			if err != nil {
				log.FromContext(context.Background()).Error(err, "unable to read pod logs")
				break
			}
			if bytesRead > 0 {
				output := string(buffer[:bytesRead])
				for _, a := range searchSlice {
					if ifFound[a] {
						continue
					}
					if strings.Contains(output, a) {
						ifFound[a] = true
						continue
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
		fail = 1
	}

	gomega.Expect(fail).Should(gomega.Equal(0))

}

var _ = ginkgo.Describe("Test Deployments that use Par Manager Address as DNS\n", func() {

	ginkgo.Context("Test Deployment that queries a domain NOT IN ARecord\n", func() {
		arecord := addArecord()
		defer cleanupResource(arecord)
		deployment := createDeployment("../resources/test_no_record_deployment.yaml")
		defer cleanupResource(deployment)
		// TODO: check logs on manager if deployment is ready
		time.Sleep(5 * time.Second)
		ginkgo.Specify("Does not return a Proxy IP address upon DNS lookup from Par Manager Address, only Upstream DNS\n", func() {
			var checkSlice []string
			checkSlice = append(checkSlice, "yahoo.com", GetManagerAddress())
			CheckPodLogsFromDeployment(deployment, checkSlice)
		})
	})
	ginkgo.Context("Test Deployment that queries a domain that is IN ARecord\n", func() {
		arecord := addArecord()
		defer cleanupResource(arecord)
		deployment := createDeployment("../resources/test_wget_a_record_deployment.yaml")
		defer cleanupResource(deployment)
		// TODO: check logs on manager if deployment is ready
		time.Sleep(5 * time.Second)
		ginkgo.Specify("Does return a Proxy IP address upon DNS lookup from Par Manager Address\n", func() {
			var checkSlice []string
			// TODO: has issue with check both values in checkSlice when two test run. Might be due to function instead of method in struct
			checkSlice = append(checkSlice, "google.com", arecord.Spec.IPAddress)
			CheckPodLogsFromDeployment(deployment, checkSlice)
		})
	})
})
var _ = ginkgo.BeforeSuite(func() {
	gomega.RegisterFailHandler(ginkgo.Fail)
	//set up a client
	env := envtest.Environment{
		UseExistingCluster: boolPointer(true),
	}
	config, err := env.Start()
	clientset, err = kubernetes.NewForConfig(config)
	k8sClient, err = client.New(config, client.Options{Scheme: scheme.Scheme})
	gomega.Expect(err).ToNot(gomega.HaveOccurred())
})

func TestDeployments(t *testing.T) {
	// https://onsi.github.io/ginkgo/#ginkgo-and-gomega-patterns
	time.Sleep(20 * time.Second)
	ginkgo.RunSpecs(t, "Test Deployments")
}
