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
	Pause       bool   `json:"pause"`
	Template    string `json:"template"`
	ClusterName string `json:"clusterName"`
}

type ExperimentEnvStatus string

const (
	ExperimentCreated ExperimentEnvStatus = "Created"
	ExperimentRunning ExperimentEnvStatus = "Running"
	ExperimentStopped ExperimentEnvStatus = "Stopped"
	ExperimentError   ExperimentEnvStatus = "Error"
)

type ExperimentConditionStatus string

const (
	ExperimentConditionTrue    ExperimentConditionStatus = "True"
	ExperimentConditionFalse   ExperimentConditionStatus = "False"
	ExperimentConditionUnknown ExperimentConditionStatus = "Unknown"
)

type ExperimentConditionType string

const (
	ExperimentInitialized   ExperimentConditionType = "Initialized"
	ExperimentPodReady      ExperimentConditionType = "PodReady"
	ExperimentVolumeCreated ExperimentConditionType = "VolumeCreated"
	ExperimentReady         ExperimentConditionType = "Ready"
)

type ExperimentCondition struct {
	Type               ExperimentConditionType   `json:"type"`
	Status             ExperimentConditionStatus `json:"status"`
	Reason             string                    `json:"reason"`
	Message            string                    `json:"message"`
	LastProbeTime      metav1.Time               `json:"lastProbeTime"`
	LastTransitionTime metav1.Time               `json:"lastTransitionTime"`
}

// ExperimentStatus defines the observed state of Experiment
type ExperimentStatus struct {
	Status      ExperimentEnvStatus   `json:"status"`
	ClusterSync bool                  `json:"clusterSync"`
	Conditions  []ExperimentCondition `json:"conditions"`
}

func NewExperimentCondition(conditionType ExperimentConditionType, status ExperimentConditionStatus, reason, message string) ExperimentCondition {
	return ExperimentCondition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastProbeTime:      metav1.Now(),
		LastTransitionTime: metav1.Now(),
	}
}

func QueryExperimentCondition(conditions []ExperimentCondition, conditionType ExperimentConditionType) *ExperimentCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			condition := conditions[i]
			return &condition
		}
	}
	return nil
}

func CheckExperimentCondition(conditions []ExperimentCondition, conditionType ExperimentConditionType, status ExperimentConditionStatus) bool {
	cond := QueryExperimentCondition(conditions, conditionType)
	if cond == nil {
		if status == ExperimentConditionTrue {
			return false
		}
		return true
	}
	return cond.Status == status
}

func UpdateExperimentConditions(conditions []ExperimentCondition, condition ExperimentCondition) []ExperimentCondition {
	isFound := false
	for i := range conditions {
		if conditions[i].Type == condition.Type {
			isFound = true
			conditions[i] = condition
		}
	}
	if !isFound {
		conditions = append(conditions, condition)
	}
	return conditions
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
