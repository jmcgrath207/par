#!/bin/bash
#
# Brief description of your script
# Copyright 2023 john

set -e

function main() {
  echo "$(pwd)"
	docker build -f ../../Dockerfile -t local.io/local/par:latest ../..
	minikube image load  local.io/local/par:latest --overwrite --daemon
	helm upgrade --install par ../../chart --install --set controllerManager.manager.image.repository="local.io/local/par" \
											   	 --set controllerManager.manager.image.tag="latest" \
											     --create-namespace \
											     --namespace test-par
	kubectl patch deployments.apps -n test-par par-chart-controller-manager -p \
	'{ "spec": {"template": { "spec":{"containers":[{"name":"manager", "imagePullPolicy": "Never" }]}}}}'
	helm upgrade --install nginx nginx/nginx -f resources/test_proxy.yaml -n test-par
}

main "$@"
