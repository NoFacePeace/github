/*
Copyright 2024.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ExampleASpec defines the desired state of ExampleA
type ExampleASpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of ExampleA. Edit examplea_types.go to remove/update
	// Foo string `json:"foo,omitempty"`
	GroupName string `json:"groupName,omitempty"`
}

// ExampleAStatus defines the observed state of ExampleA
type ExampleAStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	UnderControl bool `json:"underControl,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ExampleA is the Schema for the examplea API
type ExampleA struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExampleASpec   `json:"spec,omitempty"`
	Status ExampleAStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ExampleAList contains a list of ExampleA
type ExampleAList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExampleA `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExampleA{}, &ExampleAList{})
}
