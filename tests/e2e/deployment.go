package e2e

import (
	"context"
	testFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
	"testing"
)

func testCreateClients(t *testing.T) {

	testCtx := framework.NewTestCtx(t)
	defer testCtx.Cleanup(t)

	ns := framework.CreateNamespace(context.Background(), t, testCtx)

	////testCtx := framework.Framework{}NewTestCtx(t)
	//defer testCtx.Cleanup(t)

	simple, err := testFramework.MakeDeployment("../resources/test_a_record_deployment.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if err := framework.CreateDeployment(context.Background(), ns, simple); err != nil {
		t.Fatal("Creating simple basic auth app failed: ", err)
	}
}
