/*

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

package v1alpha2

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

func init() {
	SchemeBuilder.Register(&PuppetCAIssuer{}, &PuppetCAIssuerList{})
}

// PuppetCAIssuerSpec defines the desired state of PuppetCAIssuer
type PuppetCAIssuerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Provisioner contains the Puppet CA certificates provisioner configuration.
	Provisioner PuppetCAProvisioner `json:"provisioner"`
}

// PuppetCAIssuerStatus defines the observed state of PuppetCAIssuer
type PuppetCAIssuerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +optional
	Conditions []PuppetCAIssuerCondition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// PuppetCAIssuer is the Schema for the puppetcaissuers API
// +kubebuilder:subresource:status
type PuppetCAIssuer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PuppetCAIssuerSpec   `json:"spec,omitempty"`
	Status PuppetCAIssuerStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PuppetCAIssuerList contains a list of PuppetCAIssuer
type PuppetCAIssuerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PuppetCAIssuer `json:"items"`
}

// SecretKeySelector contains the reference to a secret.
type SecretKeySelector struct {
	// The key of the secret to select from. Must be a valid secret key.
	// +optional
	Key string `json:"key,omitempty"`
}

// PuppetCAProvisioner contains the configuration for requesting certificate from the Puppet CA
type PuppetCAProvisioner struct {
	// The name of the secret in the pod's namespace to select from.
	Name string `json:"name"`

	// Reference to URL of the Puppet CA
	URLRef SecretKeySelector `json:"url"`

	// Reference to certificate to access the Puppet CA
	CertRef SecretKeySelector `json:"cert"`

	// Reference to certificate to access the Puppet CA
	KeyRef SecretKeySelector `json:"key"`

	// Reference to certificate to access the Puppet CA
	CaCertRef SecretKeySelector `json:"cacert"`
}

// ConditionType represents a PuppetCAIssuer condition type.
// +kubebuilder:validation:Enum=Ready
type ConditionType string

const (
	// ConditionReady indicates that a PuppetCAIssuer is ready for use.
	ConditionReady ConditionType = "Ready"
)

// ConditionStatus represents a condition's status.
// +kubebuilder:validation:Enum=True;False;Unknown
type ConditionStatus string

// These are valid condition statuses. "ConditionTrue" means a resource is in
// the condition; "ConditionFalse" means a resource is not in the condition;
// "ConditionUnknown" means kubernetes can't decide if a resource is in the
// condition or not. In the future, we could add other intermediate
// conditions, e.g. ConditionDegraded.
const (
	// ConditionTrue represents the fact that a given condition is true
	ConditionTrue ConditionStatus = "True"

	// ConditionFalse represents the fact that a given condition is false
	ConditionFalse ConditionStatus = "False"

	// ConditionUnknown represents the fact that a given condition is unknown
	ConditionUnknown ConditionStatus = "Unknown"
)

// PuppetCAIssuerCondition contains condition information for the issuer.
type PuppetCAIssuerCondition struct {
	// Type of the condition, currently ('Ready').
	Type ConditionType `json:"type"`

	// Status of the condition, one of ('True', 'False', 'Unknown').
	// +kubebuilder:validation:Enum=True;False;Unknown
	Status ConditionStatus `json:"status"`

	// LastTransitionTime is the timestamp corresponding to the last status
	// change of this condition.
	// +optional
	LastTransitionTime *metav1.Time `json:"lastTransitionTime,omitempty"`

	// Reason is a brief machine readable explanation for the condition's last
	// transition.
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is a human readable description of the details of the last
	// transition, complementing reason.
	// +optional
	Message string `json:"message,omitempty"`
}
