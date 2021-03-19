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

// ExperimentSpec defines the desired state of Experiment
type ExperimentSpec struct {
	Template          string  `json:"template"`
	CustomClusterName *string `json:"customClusterName"`
}

type ExperimentEnvStatus string

const (
	ExperimentCreated ExperimentEnvStatus = "Created"
	ExperimentRunning ExperimentEnvStatus = "Running"
	ExperimentStop    ExperimentEnvStatus = "Stop"
)

type ConditionStatus string

const (
	ConditionTrue    ConditionStatus = "True"
	ConditionFalse   ConditionStatus = "False"
	ConditionUnknown ConditionStatus = "Unknown"
)

type ExperimentConditionType string

const (
	PodInitialized ExperimentConditionType = "Initialized"
	PodReady       ExperimentConditionType = "Ready"
)

type Condition struct {
	Type    ExperimentConditionType `json:"type"`
	Status  ConditionStatus         `json:"status"`
	Reason  string                  `json:"reason"`
	Message string                  `json:"message"`
}

// ExperimentStatus defines the observed state of Experiment
type ExperimentStatus struct {
	Status     ExperimentEnvStatus `json:"status"`
	Conditions []Condition         `json:"conditions"`
}

// +kubebuilder:object:root=true

// Experiment is the Schema for the experiments API
// +kubebuilder:subresource:status
type Experiment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExperimentSpec   `json:"spec,omitempty"`
	Status ExperimentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ExperimentList contains a list of Experiment
type ExperimentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Experiment `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Experiment{}, &ExperimentList{})
}
