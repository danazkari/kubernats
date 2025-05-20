/*
Copyright 2024 xiloss.

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

// NATSUserAccountSpec defines the desired state of NATSUserAccount
type NATSUserAccountSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

  // +kubebuilder:validation:Required
  Account string `json:"account"`
  // +kubebuilder:validation:Optional
  Permissions PermissionsSpec `json:"permissions,omitempty"`
}

type PermissionsSpec struct {
  Pub SubjectRules `json:"pub,omitempty"`
  Sub SubjectRules `json:"sub,omitempty"`
}

type SubjectRules struct {
  Allow []string `json:"allow,omitempty"`
  Deny []string `json:"deny,omitempty"`
}

// NATSUserAccountStatus defines the observed state of NATSUserAccount
type NATSUserAccountStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// NATSUserAccount is the Schema for the natsuseraccounts API
type NATSUserAccount struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NATSUserAccountSpec   `json:"spec,omitempty"`
	Status NATSUserAccountStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NATSUserAccountList contains a list of NATSUserAccount
type NATSUserAccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NATSUserAccount `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NATSUserAccount{}, &NATSUserAccountList{})
}
