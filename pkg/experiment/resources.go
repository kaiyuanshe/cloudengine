package experiment

import (
	hackathonv1 "cloudengine/api/v1"
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceState struct {
	Cluster         *hackathonv1.CustomCluster
	Template        *hackathonv1.Template
	EnvPod          []corev1.Pod
	DataVolume      *corev1.PersistentVolume
	DataVolumeClaim *corev1.PersistentVolumeClaim
}

func NewExprResourceStatus(ctx context.Context, k8sClient client.Client, expr *hackathonv1.Experiment) (*ResourceState, error) {
	var (
		cluster  = &hackathonv1.CustomCluster{}
		template = &hackathonv1.Template{}
		pv       = &corev1.PersistentVolume{}
		pvc      = &corev1.PersistentVolumeClaim{}
		err      error
	)

	if err = k8sClient.Get(ctx, types.NamespacedName{
		Namespace: expr.Namespace,
		Name:      expr.Spec.ClusterName,
	}, cluster); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, fmt.Errorf("query custom cluster failed %s", err.Error())
		}
		return nil, fmt.Errorf("cluster %s not found", expr.Spec.ClusterName)
	}

	if err = k8sClient.Get(ctx, types.NamespacedName{
		Namespace: expr.Namespace,
		Name:      expr.Spec.Template,
	}, template); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, fmt.Errorf("query template failed %s", err.Error())
		}
		return nil, fmt.Errorf("template %s not found", expr.Spec.Template)
	}

	// find pv
	if err = k8sClient.Get(ctx, types.NamespacedName{
		Namespace: expr.Namespace,
		Name:      expr.Name,
	}, pv); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, fmt.Errorf("query pv failed: %s", err.Error())
		}
		pv = nil
	}

	// find pvc
	if err = k8sClient.Get(ctx, types.NamespacedName{
		Namespace: expr.Namespace,
		Name:      expr.Name,
	}, pvc); err != nil {
		if client.IgnoreNotFound(err) != nil {
			return nil, fmt.Errorf("query pvc failed: %s", err.Error())
		}
		pvc = nil
	}

	podList := &corev1.PodList{}
	selector := labels.NewSelector()
	requireExprName, err := labels.NewRequirement(LabelKeyExperimentName, selection.Equals, []string{expr.Name})
	if err != nil {
		return nil, fmt.Errorf("build label selector failed: %s", err.Error())
	}
	selector = selector.Add(*requireExprName)
	err = k8sClient.List(ctx, podList, client.InNamespace(expr.Namespace), client.MatchingLabelsSelector{Selector: selector})
	if err != nil {
		return nil, err
	}

	return &ResourceState{
		Cluster:         cluster,
		Template:        template,
		EnvPod:          podList.Items,
		DataVolume:      pv,
		DataVolumeClaim: pvc,
	}, nil
}
