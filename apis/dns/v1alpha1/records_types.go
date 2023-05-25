/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RecordsSpec defines the desired state of Records

type RecordsSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Records. Edit records_types.go to remove/update
	ManagerAddress string         `json:"manager-address,omitempty"`
	A              []ARecordsSpec `json:"a,omitempty"`
}

type ARecordsSpec struct {
	Namespaces  string            `json:"namespaces"`
	Labels      map[string]string `json:"labels"`
	HostName    string            `json:"hostname"`
	IPAddress   string            `json:"ip-address"`
	ForwardType string            `json:"forward-type"`
}

// RecordsStatus defines the observed state of Records
type RecordsStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Records is the Schema for the records API
type Records struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RecordsSpec   `json:"spec,omitempty"`
	Status RecordsStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// RecordsList contains a list of Records
type RecordsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Records `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Records{}, &RecordsList{})
}
