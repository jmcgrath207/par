package e2e

import (
	"fmt"
	"github.com/blang/semver/v4"
	operatorFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

var (
	previousVersionFramework *operatorFramework.Framework
	framework                *operatorFramework.Framework
	exitCode                 int
)

func cleanupTestDeploy() {
	exec.Command("/bin/bash", "helm uninstall par -n test-par")
	exec.Command("/bin/bash", "helm delete nginx -n test-par")
}

func TestMain(m *testing.M) {

	cmd := exec.Command("/bin/bash", "./test_deploy.sh")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println(err.Error())
		fmt.Println(string(output))
		return
	}
	defer cleanupTestDeploy()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	kubeconfig := filepath.Join(homeDir, ".kube", "config")

	opImage := "local.io/local/party:latest"

	logger := log.New(os.Stdout, "", log.Lshortfile)

	currentVersion, err := os.ReadFile("../../VERSION")
	if err != nil {
		logger.Printf("failed to read version file: %v\n", err)
		os.Exit(1)
	}
	currentSemVer, err := semver.ParseTolerant(string(currentVersion))
	if err != nil {
		logger.Printf("failed to parse current version: %v\n", err)
		os.Exit(1)
	}

	exampleDir := "../../example"
	resourcesDir := "../e2e/resources"

	nextSemVer, err := semver.ParseTolerant(fmt.Sprintf("0.%d.0", currentSemVer.Minor))
	if err != nil {
		logger.Printf("failed to parse next version: %v\n", err)
		os.Exit(1)
	}

	// init with next minor version since we are developing toward it.
	if framework, err = operatorFramework.New(kubeconfig, opImage, exampleDir, resourcesDir, nextSemVer); err != nil {
		logger.Printf("failed to setup framework: %v\n", err)
		os.Exit(1)
	}

	exitCode = m.Run()

	os.Exit(exitCode)
}

func TestAll(t *testing.T) {
	testFuncs := map[string]func(t *testing.T){
		"testCreateClients": testCreateClients,
	}
	for name, f := range testFuncs {
		t.Run(name, f)
	}
}
