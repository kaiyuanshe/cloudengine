package experiment

import (
	hackathonv1 "cloudengine/api/v1"
	"cloudengine/pkg/common/event"
	"cloudengine/pkg/common/results"
	"cloudengine/pkg/utils/k8stools"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
		status.AddEvent(corev1.EventTypeWarning, event.ReasonUnexpected, err.Error())
		return result.WithError(err)
	}

	result.WithResult((&DataVolume{
		client:        c.Client,
		status:        status,
		resourceState: resourceState,
	}).Reconcile(ctx))

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
	if resState.Template == nil {
		status.AddEvent(corev1.EventTypeWarning, event.ReasonCreated, "template not found")
		return result.WithError(fmt.Errorf("template not found"))
	}

	expected, err := buildExpectedEnvPod(status.Experiment, resState.Template)
	if err != nil {
		return result.WithError(err)
	}
	if len(resState.EnvPod) == 0 {
		status.AddEvent(corev1.EventTypeNormal, event.ReasonCreated, "create env pod")
		return result.WithError(c.Client.Create(ctx, expected))
	}

	reconciled := resState.EnvPod[0]
	return result.With("check-env-pod", func() (reconcile.Result, error) {
		if k8stools.IsPodReady(&reconciled) {
			status.Status.Conditions = hackathonv1.UpdateExperimentConditions(
				status.Status.Conditions, hackathonv1.NewExperimentCondition(
					hackathonv1.ExperimentPodReady, hackathonv1.ExperimentConditionTrue, "", ""))
		}
		return reconcile.Result{}, nil
	})
}

func buildExpectedEnvPod(experiment *hackathonv1.Experiment, template *hackathonv1.Template) (*corev1.Pod, error) {
	podCfg := template.Data.PodTemplate
	if podCfg == nil {
		return nil, fmt.Errorf("pod template is nil")
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      experiment.Name,
			Namespace: experiment.Namespace,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "experiment",
					Image:   podCfg.Image,
					Command: podCfg.Command,
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "data-volume",
							MountPath: containerPath,
						},
					},
				},
			},
			Volumes: []corev1.Volume{
				{
					Name: "data-volume",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: dataVolumeClaimName(experiment),
							ReadOnly:  false,
						},
					},
				},
			},
		},
	}

	metaObj := pod.GetObjectMeta()
	err := controllerutil.SetControllerReference(experiment, metaObj, scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("set ref failed: %s", err.Error())
	}

	return pod, nil
}
