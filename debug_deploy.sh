#!/bin/bash
#
# Brief description of your script
# Copyright 2023 john

set -e

function trap_func() {
  kill $(jobs -p)

	kubectl delete -f config/samples
	kubectl delete -f test_deployment.yaml
}


function main() {
  local minikube_pid

	trap 'trap_func' EXIT ERR
	docker build -f DockerfileDebug -t local.io/local/par:debug-latest .
	minikube image load  local.io/local/par:debug-latest --overwrite --daemon
#	kubectl delete crds -l app.kubernetes.io/instance=par || true
	helm upgrade --install par ./chart --install --set controllerManager.manager.image.repository="local.io/local/par" \
											   	 --set controllerManager.manager.image.tag="debug-latest" \
											     --create-namespace \
											     --namespace par
	kubectl patch deployments.apps -n par par-chart-controller-manager -p \
	'{ "spec": {"template": { "spec":{"securityContext": null, "containers":[{"name":"manager", "imagePullPolicy": "Never", "livenessProbe": null, "readinessProbe": null, "securityContext": null, "command": null, "args": null  }]}}}}'
	kubectl expose deployment -n par par-chart-controller-manager --type=LoadBalancer --port=56268 || true
	minikube tunnel &
	minikube_pid=$!
	sleep 10
	echo -e "\nUse the External-IP to connect to debugging"
	kubectl get svc -n par par-chart-controller-manager
	kubectl apply -f test_deployment.yaml
	kubectl apply -f config/samples
	wait ${minikube_pid}
}

main "$@"
