#!/bin/bash
#
# Brief description of your script
# Copyright 2023 john

set -e

function trap_func() {
	kubectl delete -f tests/resources/test_a_record_deployment.yaml --ignore-not-found
	kubectl delete -f tests/resources/test_no_record_deployment.yaml --ignore-not-found
	kubectl delete -f tests/resources/test_wget_a_record_deployment.yaml --ignore-not-found
	kubectl delete -f tests/resources/test_wget_no_record_deployment.yaml --ignore-not-found
	helm uninstall par -n par
	helm delete nginx -n par
	kubectl delete -f config/samples
    kill $(jobs -p)
}


function main() {
	trap 'trap_func' EXIT ERR
	docker build -f Dockerfile -t local.io/local/par:latest .
	minikube image load  local.io/local/par:latest --overwrite --daemon
	helm upgrade --install par ./chart --install --set controllerManager.manager.image.repository="local.io/local/par" \
											   	 --set controllerManager.manager.image.tag="latest" \
											     --create-namespace \
											     --namespace par
	kubectl patch deployments.apps -n par par-chart-controller-manager -p \
	'{ "spec": {"template": { "spec":{"containers":[{"name":"manager", "imagePullPolicy": "Never" }]}}}}'
	kubectl expose deployment -n par par-chart-controller-manager --type=LoadBalancer --port=56268 || true
	helm install nginx nginx/nginx -f test_proxy.yaml -n par
#	|| helm repo add nginx https://charts.bitnami.com/bitnami && helm install nginx nginx/nginx --set commonLabels."par\.dev"=proxy
	minikube tunnel &
	minikube_pid=$!
	sleep 10
	echo -e "\nUse the External-IP to connect to debugging"
	kubectl get svc -n par par-chart-controller-manager
	kubectl apply -f tests/resources/test_a_record_deployment.yaml
	kubectl apply -f tests/resources/test_no_record_deployment.yaml
	kubectl apply -f tests/resources/test_wget_a_record_deployment.yaml
	kubectl apply -f tests/resources/test_wget_no_record_deployment.yaml
	kubectl apply -f config/samples
	wait ${minikube_pid}

}

main "$@"
