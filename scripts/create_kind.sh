#!/bin/bash

function trap_func_kind() {
  kind export logs
  kind delete cluster -n par-cluster
}


trap 'trap_func_kind' ERR

kind_cluster=$(kind get clusters | grep -o par-cluster)

if [[ -z $kind_cluster ]]; then

kind create cluster \
  --verbosity=6 \
  --config scripts/kind.yaml \
  --retain \
  --name par-cluster \
  --image "kindest/node:v${ENVTEST_K8S_VERSION}"
fi



kubectl config set-context par-cluster
echo "Kubernetes cluster:"
kubectl get nodes -o wide
