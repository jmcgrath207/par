package e2e

import (
	"context"
	"fmt"
	"github.com/jmcgrath207/par/apis/dns/v1"
	//operatorFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"os"
	"testing"
)

var (
	//framework *operatorFramework.Framework
	ns = "default"
)

func testCreateClients(t *testing.T) {

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	// Read file contents into a byte array
	fileContents, err := os.ReadFile("./resources/test_dns_v1_arecord.yaml")
	if err != nil {
		fmt.Println(err)
		return
	}

	// Create an Unstructured object from the YAML file contents

	j := json.Serializer{}

	gvk := schema.GroupVersionKind{
		Group:   v1.GroupVersion.Group,
		Version: v1.GroupVersion.Version,
		Kind:    "ARecord",
	}

	decode, s, err := j.Decode(fileContents, &gvk, nil)
	if err != nil {
		return
	}
	_ = decode
	_ = s

	////testCtx := framework.Framework{}NewTestCtx(t)
	//defer testCtx.Cleanup(t)

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
