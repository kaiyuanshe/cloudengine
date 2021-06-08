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
	"k8s.io/apimachinery/pkg/util/validation/field"
)

func init() {
	SchemeBuilder.Register(&CustomCluster{}, &CustomClusterList{})
}

// CustomClusterSpec defines the desired state of CustomCluster
type CustomClusterSpec struct {
	ClusterTimeoutSeconds int      `json:"clusterTimeoutSeconds"`
	PublishIps            []string `json:"publishIPs,omitempty"`
	PrivateIps            []string `json:"privateIPs,omitempty"`
	EnablePrivateIP       bool     `json:"enablePrivateIP"`
}

type ClusterStatus string

const (
	ClusterCreated      ClusterStatus = "Created"
	ClusterReady        ClusterStatus = "Ready"
	ClusterOutOfControl ClusterStatus = "OutOfControl"
	ClusterLost         ClusterStatus = "Lost"
	ClusterUnknown      ClusterStatus = "Unknown"
)

// CustomClusterStatus defines the observed state of CustomCluster
type CustomClusterStatus struct {
	Status     ClusterStatus      `json:"status"`
	Conditions []ClusterCondition `json:"conditions,omitempty"`
	ClusterID  string             `json:"clusterId"`
}

// +kubebuilder:object:root=true

// CustomCluster is the Schema for the customclusters API
// +kubebuilder:subresource:status
type CustomCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomClusterSpec   `json:"spec,omitempty"`
	Status CustomClusterStatus `json:"status,omitempty"`
}

func (c *CustomCluster) CheckForWarning() error {
	errs := field.ErrorList{}
	for _, f := range clusterSpecWarnings {
		err := f(c)
		if err != nil {
			errs = append(errs, err...)
		}
	}
	return errs.ToAggregate()
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

func NewClusterCondition(conditionType ClusterConditionType, status ClusterConditionStatus, reason, message string) ClusterCondition {
	return ClusterCondition{
		Type:               conditionType,
		Status:             status,
		Reason:             reason,
		Message:            message,
		LastProbeTime:      metav1.Now(),
		LastTransitionTime: metav1.Now(),
	}
}

func QueryClusterCondition(conditions []ClusterCondition, conditionType ClusterConditionType) *ClusterCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			condition := conditions[i]
			return &condition
		}
	}
	return nil
}

func CheckClusterCondition(conditions []ClusterCondition, conditionType ClusterConditionType, status ClusterConditionStatus) bool {
	cond := QueryClusterCondition(conditions, conditionType)
	if cond == nil {
		if status == ClusterStatusTrue {
			return false
		}
		return true
	}
	return cond.Status == status
}

func UpdateClusterConditions(conditions []ClusterCondition, condition ClusterCondition) []ClusterCondition {
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

type clusterValidation func(cluster *CustomCluster) field.ErrorList

var (
	clusterSpecWarnings = []clusterValidation{}
)
