
function trap_func() {
  set +e
	helm delete par -n par
	helm delete nginx -n par
	kubectl delete -f tests/resources/test_a_record_deployment.yaml --ignore-not-found
	kubectl delete -f tests/resources/test_no_record_deployment.yaml --ignore-not-found
	kubectl delete -f tests/resources/test_wget_a_record_deployment.yaml --ignore-not-found
	kubectl delete -f tests/resources/test_wget_no_record_deployment.yaml --ignore-not-found
	kubectl delete -f tests/resources/debug_service.yaml --ignore-not-found
}


function add_test_clients() {
  helm install nginx nginx/nginx -f tests/resources/test_proxy.yaml -n par
	kubectl apply -f tests/resources/test_a_record_deployment.yaml
	kubectl apply -f tests/resources/test_no_record_deployment.yaml
	kubectl apply -f tests/resources/test_wget_a_record_deployment.yaml
	kubectl apply -f tests/resources/test_wget_no_record_deployment.yaml
	kubectl apply -f tests/resources/debug_service.yaml
}
