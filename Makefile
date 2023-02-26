
## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin

## Tool Binaries
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen

## Tool Versions
CONTROLLER_TOOLS_VERSION ?= latest

# TODO: working on CRD generating error.
generate_resources:
	 $(LOCALBIN)/controller-gen paths=./resources/...


minikube_deploy_debug: controller-gen generate_resources
	eval $(minikube docker-env)
	docker build -f DockerfileDebug -t par:debug-latest .
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

controller-gen:
 	## Download controller-gen locally if necessary.
	test -s $(LOCALBIN)/controller-gen || mkdir -p $(LOCALBIN) && GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

