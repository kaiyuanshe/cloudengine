package customcluster

import (
	hackathonv1 "cloudengine/api/v1"
	"cloudengine/pkg/common/event"
	"cloudengine/pkg/common/results"
	"context"
	"fmt"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

type Driver struct {
	Client   client.Client
	Cluster  *hackathonv1.CustomCluster
	Recorder record.EventRecorder
	Log      logr.Logger
}

func (d *Driver) Reconcile(ctx context.Context, status *Status) *results.Results {
	status.Status.Status = hackathonv1.ClusterUnknown

	if !hackathonv1.CheckClusterCondition(
		d.Cluster.Status.Conditions,
		hackathonv1.ClusterInit,
		hackathonv1.ClusterStatusTrue) {
		d.Log.Info("cluster not init, do first config")
		return d.InitCustomCluster(ctx, status)
	}

	if hackathonv1.CheckClusterCondition(
		d.Cluster.Status.Conditions,
		hackathonv1.ClusterFirstConnect,
		hackathonv1.ClusterStatusFalse) {
		return results.NewResults(ctx)
	}

	hbCond := hackathonv1.QueryClusterCondition(d.Cluster.Status.Conditions, hackathonv1.ClusterHeartbeat)
	if hbCond == nil || hbCond.Status == hackathonv1.ClusterStatusFalse {
		status.Status.Status = hackathonv1.ClusterLost
		return results.NewResults(ctx)
	}

	if time.Since(hbCond.LastProbeTime.Time) > time.Duration(d.Cluster.Spec.ClusterTimeoutSeconds)*time.Second {
		status.Status.Status = hackathonv1.ClusterLost
		status.AddEvent(corev1.EventTypeWarning, event.ReasonUnhealthy, "cluster heartbeat timeout")
		hackathonv1.UpdateClusterConditions(
			status.Status.Conditions,
			hackathonv1.NewClusterCondition(hackathonv1.ClusterHeartbeat, hackathonv1.ClusterStatusFalse, event.ReasonUnhealthy, "time out"))
		return results.NewResults(ctx)
	}

	status.Status.Status = hackathonv1.ClusterReady
	if cond := hackathonv1.QueryClusterCondition(d.Cluster.Status.Conditions, hackathonv1.ClusterResourceSync);
		cond.Status == hackathonv1.ClusterStatusFalse {
		status.Status.Status = hackathonv1.ClusterOutOfControl
		status.AddEvent(corev1.EventTypeWarning, event.ReasonUnexpected, fmt.Sprintf("resource sync error: %s", cond.Message))
	}
	if cond := hackathonv1.QueryClusterCondition(d.Cluster.Status.Conditions, hackathonv1.ClusterCommandApply);
		cond.Status == hackathonv1.ClusterStatusFalse {
		status.Status.Status = hackathonv1.ClusterOutOfControl
		status.AddEvent(corev1.EventTypeWarning, event.ReasonUnexpected, fmt.Sprintf("cluster command apply error: %s", cond.Message))
	}

	return results.NewResults(ctx).With("wait-next-heartbeat-check", func() (reconcile.Result, error) {
		timeoutAt := hbCond.LastProbeTime.Time.Add(time.Duration(d.Cluster.Spec.ClusterTimeoutSeconds+1) * time.Second)
		return reconcile.Result{
			Requeue:      true,
			RequeueAfter: timeoutAt.Sub(time.Now()),
		}, nil
	})
}

func (d *Driver) InitCustomCluster(ctx context.Context, status *Status) *results.Results {
	status.Status.Status = hackathonv1.ClusterCreated
	hackathonv1.UpdateClusterConditions(status.Status.Conditions, hackathonv1.NewClusterCondition(hackathonv1.ClusterInit, hackathonv1.ClusterStatusFalse, "", ""))
	status.AddEvent(corev1.EventTypeNormal, event.ReasonCreated, "wait for first heartbeat")
	return results.NewResults(ctx)
}
