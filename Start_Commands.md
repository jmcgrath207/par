


https://book.kubebuilder.io/quick-start.html

kubebuilder init --domain  --repo my.domain/guestbook
kubebuilder create api --group dns --version v1 --kind aRecord
kubebuilder edit --multigroup=true
kubebuilder create api --group proxy --version  v1alpha1 --kind HttpProxy


### Check the makefile for the kubebuilder version.
Version is in the PROJECT file in root