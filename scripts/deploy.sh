#!/bin/bash
#
# Brief description of your script
# Copyright 2023 john

set -e

source scripts/helpers.sh

function main() {
  local image_tag
  local dockerfile

  trap 'trap_func' EXIT ERR

  if [[ $ENV == "debug" ]]; then
    image_tag="debug-latest"
    dockerfile="DockerfileDebug"
  else
    image_tag="latest"
    dockerfile="Dockerfile"
  fi

  docker build -f ${dockerfile} -t local.io/local/par:$image_tag .
  kind load docker-image -v 9 --name par-cluster --nodes par-cluster-worker local.io/local/par:$image_tag
  #	kubectl delete crds -l app.kubernetes.io/instance=par || true
  helm upgrade --install par ./chart --set controllerManager.manager.image.repository="local.io/local/par" \
    --set controllerManager.manager.image.tag="${image_tag}" \
    --create-namespace \
    --namespace par
  # Patch deploy so Kind image upload to work.
  if [[ $ENV == "debug" ]]; then
    # Disable for Debugging of Delve.
    kubectl patch deployments.apps -n par par-chart-controller-manager -p \
      '{ "spec": {"template": { "spec":{"securityContext": null, "containers":[{"name":"manager", "imagePullPolicy": "Never", "livenessProbe": null, "readinessProbe": null, "securityContext": null, "command": null, "args": null  }]}}}}'
  else
  	kubectl patch deployments.apps -n par par-chart-controller-manager -p \
  	'{ "spec": {"template": { "spec":{"containers":[{"name":"manager", "imagePullPolicy": "Never" }]}}}}'
  fi

  helm install nginx nginx/nginx -f tests/resources/test_proxy.yaml -n par

  # Background log following for manager
  ( sleep 10 ; printf "\n\n" && while :; do kubectl logs -n par -l par.dev/manager="true" -f || sleep 5; done) &

  if [[ $ENV == "debug" ]]; then
    add_test_clients
    kubectl port-forward -n par service/par-manager-debug 30002:9999
  elif [[ $ENV == "e2e" ]]; then
    ${LOCALBIN}/setup-envtest use ${ENVTEST_K8S_VERSION} --bin-dir ${LOCALBIN} -p path
    go test ./tests/e2e/... -coverprofile cover.out
  elif [[ $ENV == "e2e-debug" ]]; then
    ${LOCALBIN}/setup-envtest use ${ENVTEST_K8S_VERSION} --bin-dir ${LOCALBIN} -p path
    sleep infinity
  else
    # Assume make local deploy
    add_test_clients
    sleep infinity
  fi

}

main "$@"