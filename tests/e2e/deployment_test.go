package e2e

import (
	"context"
	"fmt"
	dnsv1 "github.com/jmcgrath207/par/apis/dns/v1"
	//appsv1 "k8s.io/api/apps/v1"
	"os"
	"testing"
	"time"

	. "github.com/onsi/gomega"

	//corev1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	//"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
)

var (
	timeout  = time.Second * 10
	duration = time.Second * 10
	interval = time.Millisecond * 250
)

func boolPointer(b bool) *bool {
	return &b
}

func TestDeployments(t *testing.T) {

	g := NewWithT(t)

	env := envtest.Environment{
		UseExistingCluster: boolPointer(true),
	}
	config, err := env.Start()
	g.Expect(err).ToNot(HaveOccurred())
	dnsv1.AddToScheme(scheme.Scheme)
	k8sClient, err := client.New(config, client.Options{Scheme: scheme.Scheme})
	g.Expect(err).ToNot(HaveOccurred())

	// Add Read Yaml of Arecord and convert to type dnsv1.Arecord

	yamlFile, err := os.ReadFile("./resources/test_dns_v1_arecord.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	aRecord := &dnsv1.Arecord{}
	_, _, err = decode(yamlFile, nil, aRecord)
	if err != nil {
		fmt.Printf("%#v", err)
	}
	g.Expect(k8sClient.Create(context.TODO(), aRecord)).Should(Succeed())

	// Read Yaml of deployment create a test deployment

	// Add Read Yaml of Arecord and convert to type dnsv1.Arecord

	//yamlFile, err = os.ReadFile("./resources/test_a_record_deployment.yaml")
	//if err != nil {
	//	fmt.Println(err)
	//	return
	//}
	//decode = scheme.Codecs.UniversalDeserializer().Decode
	//
	//testDeployment := &appsv1.Deployment{}
	//_, _, err = decode(yamlFile, nil, testDeployment)
	//if err != nil {
	//	fmt.Printf("%#v", err)
	//}
	//
	//g.Eventually(func() bool {
	//	err = k8sClient.Get(context.TODO(), cmObjectKey, &configMap)
	//
	//	if err != nil {
	//		return false
	//	}
	//	return true
	//}, timeout, interval).Should(BeTrue())

	err = k8sClient.Delete(context.TODO(), aRecord)
	g.Expect(err).ToNot(HaveOccurred())

	// cleanup
	//g.Eventually(func() bool {
	//	var configMapDeleted corev1.ConfigMap
	//	err = k8sClient.Get(context.TODO(), cmObjectKey, &configMapDeleted)
	//	if err == nil {
	//		return false
	//	}
	//	return errors.IsNotFound(err)
	//}, timeout, interval).Should(BeTrue())

}
