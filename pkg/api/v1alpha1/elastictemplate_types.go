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

// ElasticTemplateSpec defines the desired state of ElasticTemplate
type ElasticTemplateSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Template name in elasticsearch server
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9-_\.]+$`
	TemplateName *string `json:"templateName"`

	// Elasticsearch URI with this format <scheme>://<user>:<password>@<hostname>:<port> from a key of a secret in the local namespace
	// +kubebuilder:validation:Required
	ElasticURI ElasticURISource `json:"elasticURI"`

	// Number of elasticsearch shards
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=500
	// +kubebuilder:validation:Required
	NumberOfShards *int32 `json:"numberOfShards"`

	// Number of elasticsearch replicas
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=3
	// +kubebuilder:validation:Required
	NumberOfReplicas *int32 `json:"numberOfReplicas"`

	// Template order
	// +optional
	// +nullable
	Order *int `json:"order,omitempty"`

	// Template mappings, settings, index_patterns and version
	// +kubebuilder:validation:Required
	Model *string `json:"model"`
}

// ElasticTemplateStatus defines the observed state of ElasticTemplate
type ElasticTemplateStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Status indicates whether template was created successfully in elasticsearch server. Possible values: Created, Error, Retry
	// +optional
	Status string `json:"status,omitempty"`

	// The http code status returned by elasticsearch
	// +optional
	HttpCodeStatus string `json:"httpCodeStatus,omitempty"`

	// The message returned by elasticsearch. Useful when Status is Error or Retry
	// +optional
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=et
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="TEMPLATE_NAME",type="string",JSONPath=".spec.templateName"
// +kubebuilder:printcolumn:name="SHARDS",type="integer",JSONPath=".spec.numberOfShards"
// +kubebuilder:printcolumn:name="REPLICAS",type="integer",JSONPath=".spec.numberOfReplicas"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// ElasticTemplate is the Schema for the elastictemplates API
type ElasticTemplate struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticTemplateSpec   `json:"spec,omitempty"`
	Status ElasticTemplateStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ElasticTemplateList contains a list of ElasticTemplate
type ElasticTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ElasticTemplate `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ElasticTemplate{}, &ElasticTemplateList{})
}
