package types

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterStatus struct {
	GVK        v1.GroupVersionKind `json:"gvk"`
	Cluster    string              `json:"cluster"`
	Conditions []CommonCondition   `json:"conditions,omitempty"`
}

type ResourceStatus struct {
	GVK             v1.GroupVersionKind `json:"gvk"`
	Resource        string              `json:"resource"`
	Conditions      []CommonCondition   `json:"conditions,omitempty"`
	ResourceVersion string              `json:"resourceVersion"`
}

type CommonCondition struct {
	Type    string `json:"type"`
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Message string `json:"message"`
}
