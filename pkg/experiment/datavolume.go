package experiment

import (
	hackathonv1 "cloudengine/api/v1"
	"cloudengine/pkg/common/event"
	"cloudengine/pkg/common/results"
	"context"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
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
	if hackathonv1.CheckExperimentCondition(v.status.Status.Conditions, hackathonv1.ExperimentVolumeCreated, hackathonv1.ExperimentConditionFalse) {
		return v.createDataVolume(ctx)
	}

	if v.resourceState.PV == nil || v.resourceState.PVC == nil {
		v.status.AddEvent(corev1.EventTypeWarning, event.ReasonDeleted, "data volume was deleted, recreate")
		return v.createDataVolume(ctx)
	}

	return nil
}

func (v *DataVolume) createDataVolume(ctx context.Context) *results.Results {
	result := results.NewResults(ctx)
	v.status.AddEvent(corev1.EventTypeNormal, event.ReasonCreated, "create data volume")
	return result
}
