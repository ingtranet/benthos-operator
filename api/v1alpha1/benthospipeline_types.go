/*
Copyright 2022.

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
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BenthosPipelineSpec defines the desired state of BenthosPipeline
type BenthosPipelineSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of BenthosPipeline. Edit benthospipeline_types.go to remove/update
	//+kubebuilder:default="jeffail/benthos:edge-cgo"
	Image string `json:"image,omitempty"`
	//+kubebuilder:default=1
	Replicas int32  `json:"replicas,omitempty"`
	Config   string `json:"config,omitempty"`
	// +kubebuilder:pruning:PreserveUnknownFields
	// +kubebuilder:validation:Type=object
	ConfigInline runtime.RawExtension `json:"configInline,omitempty"`
}

// BenthosPipelineStatus defines the observed state of BenthosPipeline
type BenthosPipelineStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// BenthosPipeline is the Schema for the benthospipelines API
type BenthosPipeline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BenthosPipelineSpec   `json:"spec,omitempty"`
	Status BenthosPipelineStatus `json:"status,omitempty"`
}

func (p *BenthosPipeline) GetYamlConfig() string {
	if p.Spec.Config != "" {
		return p.Spec.Config
	} else {
		var config map[string]interface{}
		if err := json.Unmarshal(p.Spec.ConfigInline.Raw, &config); err != nil {
			return "config-parsing-error"
		}
		yamlConfig, err := yaml.Marshal(config)
		if err != nil {
			return "config-marshaling-error"
		}
		return string(yamlConfig)
	}
}

func (p *BenthosPipeline) GetConfigHash() string {
	hash := sha1.Sum([]byte(p.GetYamlConfig()))
	return base64.StdEncoding.EncodeToString(hash[:])
}

//+kubebuilder:object:root=true

// BenthosPipelineList contains a list of BenthosPipeline
type BenthosPipelineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BenthosPipeline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BenthosPipeline{}, &BenthosPipelineList{})
}
