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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// PuppetCAIssuerSpec defines the desired state of PuppetCAIssuer
type PuppetCAIssuerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of PuppetCAIssuer. Edit PuppetCAIssuer_types.go to remove/update
	Provisioner PuppetCAProvisioner `json:"provisioner"`
}

// PuppetCAIssuerStatus defines the observed state of PuppetCAIssuer
type PuppetCAIssuerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// PuppetCAIssuer is the Schema for the puppetcaissuers API
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
	URL SecretKeySelector `json:"url"`

	// Reference to certificate to access the Puppet CA
	Cert SecretKeySelector `json:"cert"`

	// Reference to certificate to access the Puppet CA
	Key SecretKeySelector `json:"key"`

	// Reference to certificate to access the Puppet CA
	CaCert SecretKeySelector `json:"cacert"`
}

func init() {
	SchemeBuilder.Register(&PuppetCAIssuer{}, &PuppetCAIssuerList{})
}
