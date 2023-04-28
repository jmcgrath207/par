package e2e

import (
	"context"
	"fmt"
	dnsv1 "github.com/jmcgrath207/par/apis/dns/v1"
	"github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"strings"
)

// REF: https://superorbital.io/blog/testing-production-controllers/
// REF: https://github.com/superorbital/random-number-controller

var (
	timeout   = time.Second * 10
	duration  = time.Second * 10
	interval  = time.Millisecond * 250
	clientset *kubernetes.Clientset
	k8sClient client.Client
	g         *gomega.WithT
)

func TestMain(m *testing.M) {
	exitCode := m.Run()
	os.Exit(exitCode)
}

func boolPointer(b bool) *bool {
	return &b
}

func cleanupResource(object client.Object) {
	err := k8sClient.Delete(context.TODO(), object)
	g.Expect(err).ToNot(HaveOccurred())
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
	err = k8sClient.Create(context.TODO(), aRecord)
	g.Expect(err).Should(Succeed())
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
	g.Expect(k8sClient.Create(context.TODO(), deployment)).Should(Succeed())
	return deployment
}

func testNoRecordDeployment() {
	deployment := createDeployment("../resources/test_no_record_deployment.yaml")
	defer cleanupResource(deployment)
	podList := v1.PodList{}
	opts := []client.ListOption{
		client.InNamespace("default"),
		client.MatchingLabels(deployment.Spec.Template.ObjectMeta.Labels),
	}
	k8sClient.List(context.TODO(), &podList, opts...)

	for _, pod := range podList.Items {
		req := clientset.CoreV1().Pods(pod.ObjectMeta.Namespace).GetLogs(pod.Name, &v1.PodLogOptions{
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
				//TODO get dns IP address on pod.
				t := []string{"yahoo.com", pod.Spec.DNSConfig.Nameservers[0]}

				for _, a := range t {
					if strings.Contains(output, a) {
						continue

					} else {
						panic(fmt.Sprintf("Output doesn't container %v", a))
					}

				}

			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func TestAll(t *testing.T) {
	g = NewWithT(t)
	env := envtest.Environment{
		UseExistingCluster: boolPointer(true),
	}

	config, err := env.Start()
	clientset, err = kubernetes.NewForConfig(config)
	k8sClient, err = client.New(config, client.Options{Scheme: scheme.Scheme})
	g.Expect(err).ToNot(HaveOccurred())
	arecord := addArecord()
	defer cleanupResource(arecord)
	testNoRecordDeployment()
}
