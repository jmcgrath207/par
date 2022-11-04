# par

# Creation Commands

operator-sdk init  --repo github.com/jmcgrath207/par --domain jmcgrath207.github.com
operator-sdk create api --group record --version v1alpha1 --kind A --resource --controller



Core types or Native kinds like Pods are not available in operators in the operator sdk.
https://github.com/kubernetes-sigs/kubebuilder/issues/1999

However, you can mutate a Core type via code  with kube builder.
https://book.kubebuilder.io/reference/webhook-for-core-types.html


# Problems.

Need mutation webhook config to code generated/manually create  MutatingWebhookConfiguration
Figure out how we can convert this to helm.