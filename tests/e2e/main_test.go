package e2e

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"testing"

	"github.com/blang/semver/v4"
	operatorFramework "github.com/prometheus-operator/prometheus-operator/test/framework"
)

var (
	previousVersionFramework *operatorFramework.Framework
	framework                *operatorFramework.Framework
	opImage                  *string
)

func TestMain(m *testing.M) {
	kubeconfig := flag.String(
		"kubeconfig",
		"",
		"kube config path, e.g. $HOME/.kube/config",
	)
	opImage = flag.String(
		"operator-image",
		"",
		"operator image, e.g. quay.io/prometheus-operator/prometheus-operator",
	)
	flag.Parse()

	var (
		err      error
		exitCode int
	)

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

	prevStableVersionURL := fmt.Sprintf("https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/release-%d.%d/VERSION", currentSemVer.Major, currentSemVer.Minor-1)
	reader, err := operatorFramework.URLToIOReader(prevStableVersionURL)
	if err != nil {
		logger.Printf("failed to get previous version file content: %v\n", err)
		os.Exit(1)
	}

	prevStableVersion, err := io.ReadAll(reader)
	if err != nil {
		logger.Printf("failed to read previous stable version: %v\n", err)
		os.Exit(1)
	}

	prometheusOperatorGithubBranchURL := "https://raw.githubusercontent.com/prometheus-operator/prometheus-operator"

	prevSemVer, err := semver.ParseTolerant(string(prevStableVersion))
	if err != nil {
		logger.Printf("failed to parse previous stable version: %v\n", err)
		os.Exit(1)
	}
	prevStableOpImage := fmt.Sprintf("%s:v%s", "quay.io/prometheus-operator/prometheus-operator", strings.TrimSpace(string(prevStableVersion)))
	prevExampleDir := fmt.Sprintf("%s/release-%d.%d/example", prometheusOperatorGithubBranchURL, prevSemVer.Major, prevSemVer.Minor)
	prevResourcesDir := fmt.Sprintf("%s/release-%d.%d/test/framework/resources", prometheusOperatorGithubBranchURL, prevSemVer.Major, prevSemVer.Minor)

	if previousVersionFramework, err = operatorFramework.New(*kubeconfig, prevStableOpImage, prevExampleDir, prevResourcesDir, prevSemVer); err != nil {
		logger.Printf("failed to setup previous version framework: %v\n", err)
		os.Exit(1)
	}

	exampleDir := "../../example"
	resourcesDir := "../framework/resources"

	nextSemVer, err := semver.ParseTolerant(fmt.Sprintf("0.%d.0", currentSemVer.Minor))
	if err != nil {
		logger.Printf("failed to parse next version: %v\n", err)
		os.Exit(1)
	}

	// init with next minor version since we are developing toward it.
	if framework, err = operatorFramework.New(*kubeconfig, *opImage, exampleDir, resourcesDir, nextSemVer); err != nil {
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
