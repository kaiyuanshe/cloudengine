package experiment

import (
	hackathonv1 "cloudengine/api/v1"
	"cloudengine/pkg/common/event"
	"cloudengine/pkg/common/reconciler"
	"cloudengine/pkg/common/results"
	"cloudengine/pkg/utils/logtool"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	hostPathDir   = "/opt/open-hackathon/cloud-engine/data"
	containerPath = "/data"
	volumeSizeGi  = 10
)

type DataVolume struct {
	client        client.Client
	status        *Status
	resourceState *ResourceState
	logger        logr.Logger
}

func (v *DataVolume) Reconcile(ctx context.Context) *results.Results {
	return results.NewResults(ctx).
		WithResult(v.reconcileVolumeClaim(ctx)).
		WithResult(v.reconcileVolume(ctx))
}

func (v *DataVolume) reconcileVolume(ctx context.Context) *results.Results {
	defer logtool.SpendTimeRecord(v.logger, "reconcile volume claim")()
	result := results.NewResults(ctx)
	var (
		pvName = dataVolumeName(v.status.Experiment)
	)

	expected := buildExpectedDataVolume(v.status.Experiment)
	reconciled := v.resourceState.DataVolume
	if reconciled == nil {
		reconciled = expected
	}

	config := &reconciler.ResourceConfig{
		Client:     v.client,
		Owner:      v.status.Experiment,
		Expected:   expected,
		Reconciled: reconciled,
		NeedUpdate: func() bool {
			return !reflect.DeepEqual(expected.Spec.PersistentVolumeSource, reconciled.Spec.PersistentVolumeSource)
		},
		NeedRecreate: func() bool {
			return false
		},
		PreCreateHook: func() error {
			v.status.AddEvent(corev1.EventTypeNormal, event.ReasonCreated, "create data volume")
			return nil
		},
		PreUpdateHook: func() error {
			reconciled.Spec.StorageClassName = expected.Spec.StorageClassName
			reconciled.Spec.PersistentVolumeSource = expected.Spec.PersistentVolumeSource
			return nil
		},
		PostUpdateHook: func() error {
			v.status.AddEvent(corev1.EventTypeNormal, event.ReasonUpdated, "update data volume config")
			return nil
		},
		Logger: v.logger.WithValues("pv", pvName),
	}

	return result.WithError(reconciler.ReconcileResource(ctx, config))
}

func (v *DataVolume) reconcileVolumeClaim(ctx context.Context) *results.Results {
	defer logtool.SpendTimeRecord(v.logger, "reconcile data volume")()
	result := results.NewResults(ctx)
	var (
		pvcName = dataVolumeClaimName(v.status.Experiment)
	)
	expected := buildExpectedDataVolumeClaim(v.status.Experiment)
	v.logger.Info("build expected pvc", "pvc", pvcName)
	reconciled := v.resourceState.DataVolumeClaim
	if reconciled == nil {
		reconciled = expected
	}

	config := &reconciler.ResourceConfig{
		Client:     v.client,
		Owner:      v.status.Experiment,
		Expected:   expected,
		Reconciled: reconciled,
		NeedUpdate: func() bool {
			return false
		},
		NeedRecreate: func() bool {
			return false
		},
		PreCreateHook: func() error {
			v.status.AddEvent(corev1.EventTypeNormal, event.ReasonCreated, "create data volume claim")
			return nil
		},
		PreUpdateHook: func() error {
			return nil
		},
		PostUpdateHook: func() error {
			v.status.AddEvent(corev1.EventTypeNormal, event.ReasonUpdated, "update data volume claim")
			return nil
		},
		Logger: v.logger.WithValues("pvc", pvcName),
	}

	return result.WithError(reconciler.ReconcileResource(ctx, config))
}

func dataVolumeName(experiment *hackathonv1.Experiment) string {
	return fmt.Sprintf("pv-%s", experiment.Name)
}

func dataVolumeClaimName(experiment *hackathonv1.Experiment) string {
	return fmt.Sprintf("pvc-%s", experiment.Name)
}

func buildExpectedDataVolume(experiment *hackathonv1.Experiment) *corev1.PersistentVolume {
	return &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: experiment.Namespace,
			Name:      dataVolumeName(experiment),
			Labels: map[string]string{
				LabelKeyClusterName:    experiment.Spec.ClusterName,
				LabelKeyExperimentName: experiment.Name,
			},
		},
		Spec: corev1.PersistentVolumeSpec{
			Capacity: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceStorage: resource.MustParse(fmt.Sprintf("%dGi", volumeSizeGi)),
			},
			PersistentVolumeSource: corev1.PersistentVolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: fmt.Sprintf("%s/%s", hostPathDir, experiment.UID),
				}},
			AccessModes:                   []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
			StorageClassName:              DataVolumeStorageClass,
			ClaimRef:                      &corev1.ObjectReference{Namespace: experiment.Namespace, Name: dataVolumeClaimName(experiment)},
		},
	}
}

func buildExpectedDataVolumeClaim(experiment *hackathonv1.Experiment) *corev1.PersistentVolumeClaim {
	return &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: experiment.Namespace,
			Name:      dataVolumeClaimName(experiment),
			Labels: map[string]string{
				LabelKeyClusterName:    experiment.Spec.ClusterName,
				LabelKeyExperimentName: experiment.Name,
			},
		},
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			Resources: corev1.ResourceRequirements{Requests: map[corev1.ResourceName]resource.Quantity{
				corev1.ResourceStorage: resource.MustParse(fmt.Sprintf("%dGi", volumeSizeGi)),
			}},
			VolumeName:       dataVolumeName(experiment),
			StorageClassName: &DataVolumeStorageClass,
			DataSource:       nil,
		},
	}
}
