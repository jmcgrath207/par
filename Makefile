


minikube_deploy_debug:
	eval $(minikube docker-env)
	docker build -f DockerfileDebug -t par:debug-latest .
	controller-gen resources paths=./...
	helm upgrade par charts/par --install --set controllerManager.manager.image.repository="par" \
										   --set controllerManager.manager.image.tag="debug-latest" \
										   --set controllerManager.manager.image.imagePullPolicy="Never" \
										   --set controllerManager.manager.enableProbes=false \
										   --set controllerManager.manager.securityContext.runAsNonRoot=false \
										   --create-namespace \
										   --namespace par
	kubectl port-forward -n par deployments/par-controller-manager 56268:56268


minikube_destroy_debug:
	helm uninstall par -n par
