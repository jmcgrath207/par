package helm

import (
	"context"
	helm "github.com/mittwald/go-helm-client"
	"helm.sh/helm/v3/pkg/repo"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func Start(mgr manager.Manager) {

	opt := &helm.RestConfClientOptions{
		Options: &helm.Options{
			Namespace:        "par", // Change this to the namespace you wish the client to operate in.
			RepositoryCache:  "/tmp/.helmcache",
			RepositoryConfig: "/tmp/.helmrepo",
			Debug:            true,
			Linting:          true, // Change this to false if you don't want linting.
			DebugLog: func(format string, v ...interface{}) {
				// Change this to your own logger. Default is 'log.Printf(format, v...)'.
			},
		},
		RestConfig: mgr.GetConfig(),
	}

	helmClient, err := helm.NewClientFromRestConf(opt)
	if err != nil {
		panic(err)
	}

	// Define a public chart repository.
	chartRepo := repo.Entry{
		Name: "stable",
		URL:  "https://charts.helm.sh/stable",
	}

	// Add a chart-repository to the client.
	if err := helmClient.AddOrUpdateChartRepo(chartRepo); err != nil {
		panic(err)
	}

	chartSpec := helm.ChartSpec{
		ReleaseName: "nginx",
		ChartName:   "http://helm.whatever.com/repo/etcd-operator.tar.gz",
		Namespace:   "par",
		UpgradeCRDs: true,
		Wait:        true,
	}

	if _, err := helmClient.InstallOrUpgradeChart(context.Background(), &chartSpec, nil); err != nil {
		panic(err)
	}

}
