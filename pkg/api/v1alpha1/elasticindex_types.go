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

// ElasticIndexSpec defines the desired state of ElasticIndex
type ElasticIndexSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Index name in elasticsearch server
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9-_\.]+$`
	IndexName *string `json:"indexName"`

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

	// Index mappings, settings and aliases
	// +kubebuilder:validation:Required
	Model *string `json:"model"`
}

// ElasticIndexStatus defines the observed state of ElasticIndex
type ElasticIndexStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Status indicates whether index was created successfully in elasticsearch server. Possible values: Created, Error, Retry
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
// +kubebuilder:resource:shortName=ei
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="INDEX_NAME",type="string",JSONPath=".spec.indexName"
// +kubebuilder:printcolumn:name="SHARDS",type="integer",JSONPath=".spec.numberOfShards"
// +kubebuilder:printcolumn:name="REPLICAS",type="integer",JSONPath=".spec.numberOfReplicas"
// +kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// ElasticIndex is the Schema for the elasticindices API
type ElasticIndex struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticIndexSpec   `json:"spec,omitempty"`
	Status ElasticIndexStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ElasticIndexList contains a list of ElasticIndex
type ElasticIndexList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ElasticIndex `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ElasticIndex{}, &ElasticIndexList{})
}
