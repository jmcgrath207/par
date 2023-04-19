package e2e

import (
	"context"
	"fmt"
	"github.com/jmcgrath207/par/apis/dns/v1"
	"k8s.io/client-go/kubernetes/scheme"

	operatorFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
	"os"
	"testing"
)

var (
	framework *operatorFramework.Framework
	ns        = "default"
)

func testCreateClients(t *testing.T) {

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	// Read file contents into a byte array
	yamlFile, err := os.ReadFile("./resources/test_dns_v1_arecord.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}
	decode := scheme.Codecs.UniversalDeserializer().Decode

	aRecord := &v1.Arecord{}
	_, _, err = decode(yamlFile, nil, aRecord)
	if err != nil {
		fmt.Printf("%#v", err)
	}

	// Deploy Arecord custom type example
	// https://github.com/prometheus-operator/prometheus-operator/blob/b609a8bb6f9361c71f63f8157f00ffb6ac864b1b/pkg/client/versioned/typed/monitoring/v1/alertmanager.go#L116

	results := &v1.Arecord{}
	framework.KubeClient.CoreV1().RESTClient().Post().Namespace("default").Body(aRecord).Do(context.Background()).Into(results)

	simple, err := testFramework.MakeDeployment("./resources/test_a_record_deployment.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if err := framework.CreateOrUpdateDeploymentAndWaitUntilReady(context.Background(), ns, simple); err != nil {
		t.Fatal("Creating simple basic auth app failed: ", err)
	}

	deployment, err := framework.GetDeployment(context.Background(), ns, simple.Name)
	if err != nil {
		return
	}

	fmt.Println(deployment)

}
