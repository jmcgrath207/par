# par

# Creation Commands

operator-sdk init  --repo github.com/jmcgrath207/par --domain jmcgrath207.github.com
operator-sdk create api --group record --version v1alpha1 --kind A --resource --controller



Core types or Native kinds like Pods are not available in operators in the operator sdk.
https://github.com/kubernetes-sigs/kubebuilder/issues/1999

However, you can mutate a Core type via code  with kube builder.
https://book.kubebuilder.io/reference/webhook-for-core-types.html


Create cert for local testing.

mkdir -p /tmp/k8s-webhook-server/serving-certs/
cd /tmp/k8s-webhook-server/serving-certs/
openssl req -x509 -newkey  rsa:4096 -nodes -keyout tls.key -out tls.crt -sha256 -days 365



## test
kubectl apply -f test_pod.yaml 
kubectl delete -f test_pod.yaml


# Problems.
