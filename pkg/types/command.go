package types

import v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type CommandType string

const (
	ApplyCommand  CommandType = "APPLY"
	DeleteCommand             = "DELETE"
)

type Command struct {
	Type     CommandType         `json:"type"`
	GVK      v1.GroupVersionKind `json:"gvk"`
	Resource string              `json:"resource"`
	Content  string              `json:"content"`
}

type CommandResult struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}
