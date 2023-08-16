

* Webhooks needs to know about meta information about all records, currently a single deployment controller is started to bound to that information. This is important so we can filter and apply the correct updates.
  * This information includes:
    * Namepaces
    * Labels
    * ID of the DNS client
    * DNS address

* With this Cached information will look up a hash based on Namespace + Labels
  * This will return
    * ID of the client
    * DNS Address




### Getting this error. Cert rotator is not completing.
# https://github.com/open-policy-agent/cert-controller

[john@john-labtop chart]$ kubectl logs -n par par-manager-585d6d7b9d-4pp6s manager
2023-08-16T05:30:53Z	INFO	controller-runtime.metrics	Metrics server is starting to listen	{"addr": ":8080"}
2023-08-16T05:30:53Z	INFO	staring cert rotation
2023-08-16T05:30:53Z	INFO	Starting DNS server	{"port": "9000"}
2023-08-16T05:30:53Z	INFO	successful cert rotation
2023-08-16T05:30:53Z	INFO	controller-runtime.builder	Registering a mutating webhook	{"GVK": "apps/v1, Kind=Deployment", "path": "/mutate-apps-v1-deployment"}
2023-08-16T05:30:53Z	INFO	controller-runtime.webhook	Registering webhook	{"path": "/mutate-apps-v1-deployment"}
2023-08-16T05:30:53Z	INFO	controller-runtime.builder	skip registering a validating webhook, object does not implement admission.Validator or WithValidator wasn't called	{"GVK": "apps/v1, Kind=Deployment"}
2023-08-16T05:30:53Z	INFO	setup	starting manager
2023-08-16T05:30:53Z	INFO	Starting server	{"kind": "health probe", "addr": "[::]:8081"}
2023-08-16T05:30:53Z	INFO	controller-runtime.webhook.webhooks	Starting webhook server
2023-08-16T05:30:53Z	INFO	cert-rotation	starting cert rotator controller
2023-08-16T05:30:53Z	INFO	starting server	{"path": "/metrics", "kind": "metrics", "addr": "[::]:8080"}
2023-08-16T05:30:53Z	INFO	Starting EventSource	{"controller": "cert-rotator", "source": "kind source: *v1.Secret"}
2023-08-16T05:30:53Z	INFO	Stopping and waiting for non leader election runnables
2023-08-16T05:30:53Z	INFO	Starting EventSource	{"controller": "cert-rotator", "source": "kind source: *unstructured.Unstructured"}
2023-08-16T05:30:53Z	INFO	Starting Controller	{"controller": "cert-rotator"}
2023-08-16T05:30:53Z	INFO	Starting workers	{"controller": "cert-rotator", "worker count": 1}
2023-08-16T05:30:53Z	INFO	Shutdown signal received, waiting for all workers to finish	{"controller": "cert-rotator"}
2023-08-16T05:30:53Z	INFO	shutting down server	{"path": "/metrics", "kind": "metrics", "addr": "[::]:8080"}
2023-08-16T05:30:53Z	INFO	All workers finished	{"controller": "cert-rotator"}
2023-08-16T05:30:53Z	ERROR	controller-runtime.source.EventHandler	failed to get informer from cache	{"error": "Timeout: failed waiting for *v1.Secret Informer to sync"}
sigs.k8s.io/controller-runtime/pkg/internal/source.(*Kind).Start.func1.1
/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.15.0/pkg/internal/source/kind.go:68
k8s.io/apimachinery/pkg/util/wait.loopConditionUntilContext.func1
/go/pkg/mod/k8s.io/apimachinery@v0.27.2/pkg/util/wait/loop.go:62
k8s.io/apimachinery/pkg/util/wait.loopConditionUntilContext
/go/pkg/mod/k8s.io/apimachinery@v0.27.2/pkg/util/wait/loop.go:63
k8s.io/apimachinery/pkg/util/wait.PollUntilContextCancel
/go/pkg/mod/k8s.io/apimachinery@v0.27.2/pkg/util/wait/poll.go:33
sigs.k8s.io/controller-runtime/pkg/internal/source.(*Kind).Start.func1
/go/pkg/mod/sigs.k8s.io/controller-runtime@v0.15.0/pkg/internal/source/kind.go:56
