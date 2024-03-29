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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TemplateType string

const (
	PodTemplateType TemplateType = "Pod"
)

type PodTemplate struct {
	Image   string            `json:"image"`
	Env     map[string]string `json:"env,omitempty"`
	Command []string          `json:"command,omitempty"`
}

// TemplateData defines the desired state of Template
type TemplateData struct {
	Type            TemplateType              `json:"type"`
	PodTemplate     *PodTemplate              `json:"podTemplate"`
	IngressProtocol ExperimentIngressProtocol `json:"ingressProtocol"`
	IngressPort     int32                     `json:"ingressPort"`

	VNC *VNCConfig `json:"vnc,omitempty"`
	SSH *SSHConfig `json:"ssh,omitempty"`
}

// +kubebuilder:object:root=true

// Template is the Schema for the templates API
type Template struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Data TemplateData `json:"data,omitempty"`
}

// +kubebuilder:object:root=true

// TemplateList contains a list of Template
type TemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Template `json:"items"`
}

type VNCConfig struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SSHConfig struct {
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Key      string `json:"key,omitempty"`
}

func init() {
	SchemeBuilder.Register(&Template{}, &TemplateList{})
}
