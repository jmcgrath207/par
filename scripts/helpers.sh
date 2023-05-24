
function trap_func() {
  set +e
  {
  kubectl delete -f tests/resources/test_dns_v1alpha1_records.yaml  --ignore-not-found
	helm delete par -n par
	helm delete nginx -n par
	kubectl delete -f tests/resources/test_a_record_deployment.yaml --ignore-not-found
	kubectl delete -f tests/resources/test_no_record_deployment.yaml --ignore-not-found
	kubectl delete -f tests/resources/test_wget_a_record_deployment.yaml --ignore-not-found
	kubectl delete -f tests/resources/test_wget_no_record_deployment.yaml --ignore-not-found
	kubectl delete -f tests/resources/debug_service.yaml --ignore-not-found
	kill $(jobs -p)
	} &> /dev/null
}


function add_test_clients() {
  kubectl apply -f tests/resources/test_dns_v1alpha1_records.yaml
	kubectl apply -f tests/resources/test_a_record_deployment.yaml
	kubectl apply -f tests/resources/test_no_record_deployment.yaml
	kubectl apply -f tests/resources/test_wget_a_record_deployment.yaml
	kubectl apply -f tests/resources/test_wget_no_record_deployment.yaml
	kubectl apply -f tests/resources/debug_service.yaml
}
