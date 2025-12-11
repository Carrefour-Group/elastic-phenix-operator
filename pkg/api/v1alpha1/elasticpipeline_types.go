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

// ElasticPipelineSpec defines the desired state of ElasticPipeline
type ElasticPipelineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// PipelineName is an example field of ElasticPipeline. Edit elasticpipeline_types.go to remove/update
	// Pipeline name in elasticsearch server
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Pattern=`^[a-z0-9-_\.]+$`
	PipelineName string `json:"pipelineName,omitempty"`
	// Elasticsearch URI with this format <scheme>://<user>:<password>@<hostname>:<port> from a key of a secret in the local namespace
	// +kubebuilder:validation:Required
	ElasticURI ElasticURISource `json:"elasticURI"`
	// +kubebuilder:validation:Required
	Model string `json:"model"`
}

// ElasticPipelineStatus defines the observed state of ElasticPipeline
type ElasticPipelineStatus struct {
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

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="PIPELINE_NAME",type="string",JSONPath=".spec.pipelineName"
//+kubebuilder:printcolumn:name="STATUS",type="string",JSONPath=".status.status"
//+kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"

// ElasticPipeline is the Schema for the elasticpipelines API
type ElasticPipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticPipelineSpec   `json:"spec,omitempty"`
	Status ElasticPipelineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ElasticPipelineList contains a list of ElasticPipeline
type ElasticPipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ElasticPipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ElasticPipeline{}, &ElasticPipelineList{})
}
