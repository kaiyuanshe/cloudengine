/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/kaiyuanshe/cloudengine/pkg/common/event"
	"github.com/kaiyuanshe/cloudengine/pkg/common/results"
	"github.com/kaiyuanshe/cloudengine/pkg/customcluster"
	"github.com/kaiyuanshe/cloudengine/pkg/eventbus"
	"github.com/kaiyuanshe/cloudengine/pkg/metainfo"
	"github.com/kaiyuanshe/cloudengine/pkg/utils/logtool"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hackathonv1 "github.com/kaiyuanshe/cloudengine/api/v1"
)

// CustomClusterReconciler reconciles a CustomCluster object
type CustomClusterReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Log      logr.Logger
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=hackathon.kaiyuanshe.cn,resources=customclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hackathon.kaiyuanshe.cn,resources=customclusters/status,verbs=get;update;patch

func (r *CustomClusterReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("customcluster", req.NamespacedName)
	defer logtool.SpendTimeRecord(logger, "reconcile custom cluster")()

	ctx := context.Background()
	result := results.NewResults(ctx)
	cluster, err := r.fetchCustomCluster(ctx, req.NamespacedName)
	if err != nil {
		return ctrl.Result{}, err
	}

	if cluster == nil || r.isMarkedForDeletion(cluster) {
		eventbus.Publish(eventbus.CustomClusterDeletedTopic, req.NamespacedName)
		return ctrl.Result{}, nil
	}

	if !r.ReconcileCompatibility(cluster) {
		logger.Info("cluster not managed by this controller")
		return ctrl.Result{}, nil
	}

	if err = metainfo.UpdateClusterAnnotations(ctx, cluster, r.Client); err != nil {
		return ctrl.Result{}, fmt.Errorf("update cluster anntations failed: %s", err.Error())
	}

	status := customcluster.NewStatus(cluster)
	reconcileResult := r.internalReconcile(ctx, cluster, status)
	err = r.updateStatus(ctx, status)
	if err != nil {
		logger.Error(err, "update cluster status failed")
		return ctrl.Result{Requeue: true}, err
	}
	return result.WithError(err).WithResult(reconcileResult).Aggregate()
}

func (r *CustomClusterReconciler) fetchCustomCluster(ctx context.Context, name types.NamespacedName) (*hackathonv1.CustomCluster, error) {
	cluster := &hackathonv1.CustomCluster{}
	if err := r.Get(ctx, name, cluster); err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
		r.Log.Error(err, "get custom cluster cr failed", "namespace", name.Namespace, "name", name.Name)
		return nil, err
	}
	return cluster, nil
}

func (r *CustomClusterReconciler) ReconcileCompatibility(cluster *hackathonv1.CustomCluster) bool {
	return true
}

func (r *CustomClusterReconciler) internalReconcile(ctx context.Context, cluster *hackathonv1.CustomCluster, status *customcluster.Status) *results.Results {
	logger := r.Log.WithValues("customcluster", cluster.Name, "namespace", cluster.Namespace)
	result := results.NewResults(ctx)

	warnings := cluster.CheckForWarning()
	if warnings != nil {
		logger.Info("cluster validation has warning",
			"namespace", cluster.Namespace,
			"name", cluster.Name,
			"warning", warnings.Error(),
		)
		status.AddEvent(corev1.EventTypeWarning, event.ReasonValidation, warnings.Error())
	}

	driver := customcluster.Driver{
		Client:   r.Client,
		Cluster:  cluster,
		Recorder: r.Recorder,
		Log:      logger.WithName("ClusterDriver"),
	}
	reconcileResult := driver.Reconcile(ctx, status)
	return result.WithResult(reconcileResult)
}

func (r *CustomClusterReconciler) isMarkedForDeletion(cluster *hackathonv1.CustomCluster) bool {
	return !cluster.ObjectMeta.DeletionTimestamp.IsZero()
}

func (r *CustomClusterReconciler) updateStatus(ctx context.Context, status *customcluster.Status) error {
	events, crt := status.Apply()
	if crt == nil {
		return nil
	}

	for _, evt := range events {
		r.Recorder.Event(crt, evt.EventType, evt.Reason, evt.Message)
	}

	r.Log.Info("update custom cluster status",
		"namespace", crt.Namespace,
		"name", crt.Name,
	)
	return r.Client.Status().Update(ctx, crt)
}

func (r *CustomClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hackathonv1.CustomCluster{}).
		Complete(r)
}

func NewCustomClusterController(mgr ctrl.Manager) error {
	var (
		cli    = mgr.GetClient()
		logger = ctrl.Log.WithName("controllers").WithName("CustomCluster")
	)
	err := (&CustomClusterReconciler{
		Client:   cli,
		Recorder: mgr.GetEventRecorderFor("cluster-controller"),
		Log:      logger,
		Scheme:   mgr.GetScheme(),
	}).SetupWithManager(mgr)

	return err
}
