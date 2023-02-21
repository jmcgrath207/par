


minikube_deploy_debug:
	docker build -f DockerfileDebug -t par:debug-latest .
	helm upgrade par charts/par --install --set controllerManager.manager.image.repository="par" \
										   --set controllerManager.manager.image.tag="debug-latest" \
										   --set controllerManager.manager.enableProbes=false \
										   --set controllerManager.manager.securityContext.runAsNonRoot=false \
										   --create-namespace \
										   --namespace par

minikube_destroy_debug:
	helm uninstall par -n par
