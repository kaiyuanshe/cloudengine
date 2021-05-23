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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CustomClusterSpec defines the desired state of CustomCluster
type CustomClusterSpec struct {
}

type ClusterStatus string

const (
	ClusterReady        ClusterStatus = "Ready"
	ClusterOutOfControl ClusterStatus = "OutOfControl"
	ClusterLost         ClusterStatus = "Lost"
	ClusterUnknown      ClusterStatus = "Unknown"
)

// CustomClusterStatus defines the observed state of CustomCluster
type CustomClusterStatus struct {
	Status     ClusterStatus `json:"status"`
	Conditions []Condition   `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true

// CustomCluster is the Schema for the customclusters API
type CustomCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomClusterSpec   `json:"spec,omitempty"`
	Status CustomClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CustomClusterList contains a list of CustomCluster
type CustomClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomCluster `json:"items"`
}

type ClusterConditionType string
type ClusterConditionStatus string

const (
	ClusterInit         ClusterConditionType = "Init"
	ClusterFirstConnect ClusterConditionType = "FirstConnect"
	ClusterHeartbeat    ClusterConditionType = "Heartbeat"
	ClusterResourceSync ClusterConditionType = "ResourceSync"
	ClusterCommandApply ClusterConditionType = "CommandApply"

	ClusterStatusTrue    ClusterConditionStatus = "True"
	ClusterStatusFalse   ClusterConditionStatus = "False"
	ClusterStatusUnknown ClusterConditionStatus = "Unknown"
)

type ClusterCondition struct {
	Type               ClusterConditionType   `json:"type"`
	Status             ClusterConditionStatus `json:"status"`
	Reason             string                 `json:"reason,omitempty"`
	Message            string                 `json:"message,omitempty"`
	LastProbeTime      metav1.Time            `json:"lastProbeTime"`
	LastTransitionTime metav1.Time            `json:"lastTransitionTime"`
}

func init() {
	SchemeBuilder.Register(&CustomCluster{}, &CustomClusterList{})
}
