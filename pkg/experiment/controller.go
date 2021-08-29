package experiment

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	hackathonv1 "github.com/kaiyuanshe/cloudengine/api/v1"
	"github.com/kaiyuanshe/cloudengine/pkg/common/event"
	"github.com/kaiyuanshe/cloudengine/pkg/common/results"
	"github.com/kaiyuanshe/cloudengine/pkg/utils/k8stools"
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
		logger:        c.Logger.WithName("DataVolume"),
	}).Reconcile(ctx))

	result.WithResult((&IngressService{
		client:        c.Client,
		status:        status,
		resourceState: resourceState,
		logger:        c.Logger.WithName("IngressService"),
	}).Reconcile(ctx))

	podResult := c.reconcileExperimentPods(ctx, status, resourceState)
	status.UpdateExperimentStatus(resourceState)
	return result.WithResult(podResult)
}

func (c *Controller) firstInitExperiment(ctx context.Context, status *Status) *results.Results {
	result := results.NewResults(ctx)
	status.Status.Status = hackathonv1.ExperimentCreated
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

	if status.Experiment.Spec.Pause {
		if status.Status.Status != hackathonv1.ExperimentStopped {
			status.Status.Status = hackathonv1.ExperimentStopped
			status.AddEvent(corev1.EventTypeNormal, event.ReasonStateChange, "pause experiment")
			status.Status.Conditions = hackathonv1.UpdateExperimentConditions(
				status.Status.Conditions, hackathonv1.NewExperimentCondition(
					hackathonv1.ExperimentPodReady, hackathonv1.ExperimentConditionFalse, "PauseExperiment", ""))
		}
		if len(resState.EnvPod) > 0 {
			for i := range resState.EnvPod {
				needDelPod := resState.EnvPod[i]
				_ = c.Client.Delete(ctx, &needDelPod)
			}
		}
		return result
	}

	expected, err := buildExpectedEnvPod(status.Experiment, resState.Template)
	if err != nil {
		c.Logger.Error(err, "build expect env pod failed")
		return result.WithError(err)
	}

	if len(resState.EnvPod) == 0 {
		status.AddEvent(corev1.EventTypeNormal, event.ReasonCreated, "create env pod")
		return result.WithError(c.Client.Create(ctx, expected))
	}

	reconciled := resState.EnvPod[0]
	c.Logger.Info("found event pod", "pod", reconciled.Name, "namespace", reconciled.Namespace, "status", reconciled.Status.Phase)
	return result.With("check-env-pod", func() (reconcile.Result, error) {
		if k8stools.IsPodReady(&reconciled) {
			if status.Status.Status != hackathonv1.ExperimentRunning {
				status.Status.Status = hackathonv1.ExperimentRunning
				status.Status.Conditions = hackathonv1.UpdateExperimentConditions(
					status.Status.Conditions, hackathonv1.NewExperimentCondition(
						hackathonv1.ExperimentPodReady, hackathonv1.ExperimentConditionTrue, "", ""))
			}
		} else {
			if !status.Experiment.Spec.Pause && status.Status.Status == hackathonv1.ExperimentRunning {
				status.Status.Status = hackathonv1.ExperimentError
				status.AddEvent(corev1.EventTypeWarning, event.ReasonUnhealthy, fmt.Sprintf("pod %s not ready", reconciled.Name))
			}
			status.Status.Conditions = hackathonv1.UpdateExperimentConditions(
				status.Status.Conditions, hackathonv1.NewExperimentCondition(
					hackathonv1.ExperimentPodReady, hackathonv1.ExperimentConditionFalse, "PodNotReady", ""))
		}
		return reconcile.Result{}, nil
	})
}

func buildExpectedEnvPod(experiment *hackathonv1.Experiment, template *hackathonv1.Template) (*corev1.Pod, error) {
	podCfg := template.Data.PodTemplate
	if podCfg == nil {
		return nil, fmt.Errorf("pod template is nil")
	}
	envs := make([]corev1.EnvVar, 0)
	for k, v := range template.Data.PodTemplate.Env {
		envs = append(envs, corev1.EnvVar{
			Name:  k,
			Value: v,
		})
	}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      experiment.Name,
			Namespace: experiment.Namespace,
			Labels: map[string]string{
				LabelKeyClusterName:    experiment.Spec.ClusterName,
				LabelKeyExperimentName: experiment.Name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:    "experiment",
					Image:   podCfg.Image,
					Command: podCfg.Command,
					Env:     envs,
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

	err := controllerutil.SetControllerReference(experiment, pod.GetObjectMeta(), scheme.Scheme)
	if err != nil {
		return nil, fmt.Errorf("set pod owner ref failed: %s", err.Error())
	}

	return pod, nil
}
