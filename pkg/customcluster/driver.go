package customcluster

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	hackathonv1 "github.com/kaiyuanshe/cloudengine/api/v1"
	"github.com/kaiyuanshe/cloudengine/pkg/common/event"
	"github.com/kaiyuanshe/cloudengine/pkg/common/results"
	"github.com/kaiyuanshe/cloudengine/pkg/utils/k8stools"
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
	if d.isMetaCluster() {
		d.Log.Info("handle meta cluster")
		return d.reconcileMetaCluster(ctx, status)
	}

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
		d.Log.Info("wait first connect")
		return results.NewResults(ctx)
	}

	status.Status.Status = hackathonv1.ClusterUnknown
	hbCond := hackathonv1.QueryClusterCondition(d.Cluster.Status.Conditions, hackathonv1.ClusterHeartbeat)
	if hbCond == nil || hbCond.Status == hackathonv1.ClusterStatusFalse {
		status.Status.Status = hackathonv1.ClusterLost
		d.Log.Info("cluster lost")
		return results.NewResults(ctx)
	}

	timeout := d.Cluster.Spec.ClusterTimeoutSeconds
	if timeout == 0 {
		timeout = HeartbeatTimeoutSeconds
	}

	if time.Since(hbCond.LastProbeTime.Time) > time.Duration(timeout)*time.Second {
		status.Status.Status = hackathonv1.ClusterLost
		status.AddEvent(corev1.EventTypeWarning, event.ReasonUnhealthy, "cluster heartbeat timeout")
		status.Status.Conditions = hackathonv1.UpdateClusterConditions(
			status.Status.Conditions,
			hackathonv1.NewClusterCondition(hackathonv1.ClusterHeartbeat, hackathonv1.ClusterStatusFalse, event.ReasonUnhealthy, "time out"))
		d.Log.Info("cluster heartbeat timeout")
		return results.NewResults(ctx)
	}

	status.Status.Status = hackathonv1.ClusterReady
	if cond := hackathonv1.QueryClusterCondition(d.Cluster.Status.Conditions, hackathonv1.ClusterResourceSync); cond.Status == hackathonv1.ClusterStatusFalse {
		status.Status.Status = hackathonv1.ClusterOutOfControl
		status.AddEvent(corev1.EventTypeWarning, event.ReasonUnexpected, fmt.Sprintf("resource sync error: %s", cond.Message))
	}
	if cond := hackathonv1.QueryClusterCondition(d.Cluster.Status.Conditions, hackathonv1.ClusterCommandApply); cond.Status == hackathonv1.ClusterStatusFalse {
		status.Status.Status = hackathonv1.ClusterOutOfControl
		status.AddEvent(corev1.EventTypeWarning, event.ReasonUnexpected, fmt.Sprintf("cluster command apply error: %s", cond.Message))
	}

	return results.NewResults(ctx).With("wait-next-heartbeat-check", func() (reconcile.Result, error) {
		timeoutAt := hbCond.LastProbeTime.Time.Add(time.Duration(d.Cluster.Spec.ClusterTimeoutSeconds+1) * time.Second)
		return reconcile.Result{
			RequeueAfter: timeoutAt.Sub(time.Now()),
		}, nil
	})
}

func (d *Driver) InitCustomCluster(ctx context.Context, status *Status) *results.Results {
	return results.NewResults(ctx).With("init-custom-cluster", func() (reconcile.Result, error) {
		status.Status.Status = hackathonv1.ClusterCreated
		status.Status.ClusterID = uuid.New().String()
		d.Log.Info("init new cluster")
		status.Status.Conditions = hackathonv1.UpdateClusterConditions(status.Status.Conditions, hackathonv1.NewClusterCondition(hackathonv1.ClusterInit, hackathonv1.ClusterStatusTrue, "", ""))
		status.AddEvent(corev1.EventTypeNormal, event.ReasonCreated, "wait for first heartbeat")
		return reconcile.Result{Requeue: true}, nil
	})
}

func (d *Driver) isMetaCluster() bool {
	metaObj := d.Cluster.GetObjectMeta()
	labels := metaObj.GetLabels()
	if labels == nil {
		return false
	}
	_, ok := labels[k8stools.MetaClusterMark]
	return ok
}

func (d *Driver) reconcileMetaCluster(ctx context.Context, status *Status) *results.Results {
	status.Status.Status = hackathonv1.ClusterReady
	status.Status.Conditions = hackathonv1.UpdateClusterConditions(
		status.Status.Conditions,
		hackathonv1.NewClusterCondition(hackathonv1.ClusterInit, hackathonv1.ClusterStatusTrue, "Ready", "meta cluster ready"))
	status.Status.Conditions = hackathonv1.UpdateClusterConditions(
		status.Status.Conditions,
		hackathonv1.NewClusterCondition(hackathonv1.ClusterFirstConnect, hackathonv1.ClusterStatusTrue, "Ready", "meta cluster ready"))
	status.Status.Conditions = hackathonv1.UpdateClusterConditions(
		status.Status.Conditions,
		hackathonv1.NewClusterCondition(hackathonv1.ClusterHeartbeat, hackathonv1.ClusterStatusTrue, "Ready", "meta cluster ready"))
	status.Status.Conditions = hackathonv1.UpdateClusterConditions(
		status.Status.Conditions,
		hackathonv1.NewClusterCondition(hackathonv1.ClusterResourceSync, hackathonv1.ClusterStatusTrue, "Ready", "meta cluster ready"))
	status.Status.Conditions = hackathonv1.UpdateClusterConditions(
		status.Status.Conditions,
		hackathonv1.NewClusterCondition(hackathonv1.ClusterCommandApply, hackathonv1.ClusterStatusTrue, "Ready", "meta cluster ready"))
	return results.NewResults(ctx).With("update-meta-cluster", func() (reconcile.Result, error) {
		return reconcile.Result{RequeueAfter: time.Duration(status.Cluster.Spec.ClusterTimeoutSeconds) * time.Second}, nil
	})
}
