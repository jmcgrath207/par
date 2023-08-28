

* Webhooks needs to know about meta information about all records, currently a single deployment controller is started to bound to that information. This is important so we can filter and apply the correct updates.
  * This information includes:
    * Namepaces
    * Labels
    * ID of the DNS client
    * DNS address

* With this Cached information will look up a hash based on Namespace + Labels
  * This will return
    * ID of the client
    * DNS Address




### Getting this error. Cert rotator is not completing.
# https://github.com/open-policy-agent/cert-controller

issues was caused by not being able to list all mutating hooks in rbac. Need to fix upstream.

E0828 07:15:51.968556       1 reflector.go:148] pkg/mod/k8s.io/client-go@v0.27.2/tools/cache/reflector.go:231: Failed to watch admissionregistration.k8s.io/v1, Kind=MutatingWebhookConfiguration: failed to list admissionregistration.k8s.io/v1, Kind=MutatingWebhookConfiguration: mutatingwebhookconfigurations.admissionregistration.k8s.io is forbidden: User "system:serviceaccount:par:par-manager" cannot list resource "mutatingwebhookconfigurations" in API group "admissionregistration.k8s.io" at the cluster scope

