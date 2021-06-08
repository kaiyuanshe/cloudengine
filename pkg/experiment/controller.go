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

type Controller struct {
	Client client.Client
	Logger logr.Logger
}

func (c *Controller) Reconcile(ctx context.Context, status *Status) *results.Results {
	result := results.NewResults(ctx)
	if hackathonv1.CheckExperimentCondition(status.Experiment.Status.Conditions,
		hackathonv1.ExperimentInitialized, hackathonv1.ExperimentConditionFalse) {
		c.Logger.Info("init experiment")
		initResult := c.firstInitExperiment(ctx, status)
		result = result.WithResult(initResult)
	}

	resourceState, err := NewExprResourceStatus(ctx, c.Client, status.Experiment)
	if err != nil {
		c.Logger.Error(err, "query experiment state failed")
		return result.WithError(err)
	}

	result.WithResult((&DataVolume{
		client:        c.Client,
		status:        status,
		resourceState: resourceState,
	}).Reconcile())

	podResult := c.reconcileExperimentPods(ctx, status, resourceState)
	status.UpdateExperimentStatus(resourceState)
	return result.WithResult(podResult)
}

func (c *Controller) firstInitExperiment(ctx context.Context, status *Status) *results.Results {
	result := results.NewResults(ctx)
	status.Status.Conditions = hackathonv1.UpdateExperimentConditions(
		status.Status.Conditions,
		hackathonv1.NewExperimentCondition(hackathonv1.ExperimentInitialized, hackathonv1.ExperimentConditionTrue, "", ""))
	status.AddEvent(corev1.EventTypeNormal, event.ReasonCreated, "init experiment")
	return result
}

func (c *Controller) reconcileExperimentPods(ctx context.Context, status *Status, resState *ResourceState) *results.Results {
	result := results.NewResults(ctx)
	return result
}
