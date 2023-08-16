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
  # Install Par Chart
  helm upgrade --install par ./chart \
    --set image.repository="local.io/local/par" \
    --set image.tag="${image_tag}" \
    --set image.imagePullPolicy="Never" \
    --set metrics="true" \
    --create-namespace \
    --namespace par --wait
  # Patch deploy so Kind image upload to work.
  if [[ $ENV == "debug" ]]; then
    # Disable for Debugging of Delve.
    kubectl patch deployments.apps -n par par-manager -p \
      '{ "spec": {"template": { "spec":{"securityContext": null, "containers":[{"name":"manager", "livenessProbe": null, "readinessProbe": null, "securityContext": null, "command": null, "args": null  }]}}}}'
  fi

  # kill dangling port forwards if found.
  sudo ss -aK '( dport = :8080 or sport = :8080 )' | true

  # Deploy Proxy
  helm install nginx oci://registry-1.docker.io/bitnamicharts/nginx -f tests/resources/test_proxy.yaml -n par

  # Start Prometheus Port Forward
  (
    sleep 10
    printf "\n\n" && while :; do kubectl port-forward -n par service/par-manager-metrics 8080:8080 || sleep 5; done
  ) &

  if [[ $ENV == "debug" ]]; then
    # Background log following for manager
    (
      sleep 10
      printf "\n\n" && while :; do kubectl logs -n par -l par.dev/manager="true" -f || sleep 5; done
    ) &
    add_test_clients
    kubectl port-forward -n par service/par-manager-debug 30002:9999

  elif [[ $ENV == "e2e" ]]; then
    ${LOCALBIN}/setup-envtest use ${ENVTEST_K8S_VERSION} --bin-dir ${LOCALBIN} -p path
    ${LOCALBIN}/ginkgo -v -r --race --randomize-all --randomize-suites ./tests/e2e/...

  elif [[ $ENV == "e2e-debug" ]]; then
    ${LOCALBIN}/setup-envtest use ${ENVTEST_K8S_VERSION} --bin-dir ${LOCALBIN} -p path
    sleep infinity
  else
    # Assume make local deploy
    # Background log following for manager
    add_test_clients
    (
      sleep 10
      printf "\n\n" && while :; do kubectl logs -n par -l par.dev/manager="true" -f || sleep 5; done
    ) &
    sleep infinity
  fi
}

main "$@"
