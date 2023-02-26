# par


# Example
https://github.com/kubernetes-sigs/controller-runtime/tree/main/examples/crd



go get sigs.k8s.io/controller-tools/cmd/controller-gen@latest
cd $(go env GOROOT)
$GOROOT/

Used for generating utility code and Kubernetes YAML.
https://book.kubebuilder.io/reference/controller-gen.html#controller-gen-cli
https://github.com/kubernetes-sigs/controller-tools


## test
kubectl apply -f test_pod.yaml 
kubectl delete -f test_pod.yaml


# Problems.
