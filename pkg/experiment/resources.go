package experiment

import (
	hackathonv1 "cloudengine/api/v1"
	"context"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceState struct {
	Cluster  *hackathonv1.CustomCluster
	Template *hackathonv1.Template
	EnvPod   *corev1.Pod
	PV       *corev1.PersistentVolume
	PVC      *corev1.PersistentVolumeClaim
}

func NewExprResourceStatus(ctx context.Context, client client.Client, expr *hackathonv1.Experiment) (*ResourceState, error) {
	return nil, nil
}
