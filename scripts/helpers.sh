
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
	kubectl delete  mutatingwebhookconfigurations.admissionregistration.k8s.io  par-mutating-webhook --ignore-not-found
	jobs -p | xargs kill -SIGSTOP
	jobs -p | xargs kill -9
	sudo ss -aK '( dport = :8080 or sport = :8080 )'
	} &> /dev/null
}


function add_test_clients() {
  while [ "$(kubectl get pods -n par -l par.dev/manager=true -o=jsonpath='{.items[*].status.phase}')" != "Running" ]; do
    echo  "waiting for manager pod to start. Sleep 10" && sleep 10
done
  kubectl apply -f tests/resources/test_dns_v1alpha1_records.yaml
	kubectl apply -f tests/resources/test_a_record_deployment.yaml
	kubectl apply -f tests/resources/test_no_record_deployment.yaml
	kubectl apply -f tests/resources/test_wget_a_record_deployment.yaml
	kubectl apply -f tests/resources/test_wget_no_record_deployment.yaml
	kubectl apply -f tests/resources/debug_service.yaml
}
