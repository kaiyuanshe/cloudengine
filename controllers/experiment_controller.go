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
	"github.com/kaiyuanshe/cloudengine/pkg/common/results"
	"github.com/kaiyuanshe/cloudengine/pkg/eventbus"
	"github.com/kaiyuanshe/cloudengine/pkg/experiment"
	"github.com/kaiyuanshe/cloudengine/pkg/utils/logtool"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hackathonv1 "github.com/kaiyuanshe/cloudengine/api/v1"
)

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	client.Client
	Recorder record.EventRecorder
	Log      logr.Logger
	Scheme   *runtime.Scheme
}

// +kubebuilder:rbac:groups=hackathon.kaiyuanshe.cn,resources=experiments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hackathon.kaiyuanshe.cn,resources=experiments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=hackathon.kaiyuanshe.cn,resources=templates,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hackathon.kaiyuanshe.cn,resources=templates/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;create;delete;patch;update
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;delete;patch;update

func (r *ExperimentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("experiment", req.NamespacedName)
	result := results.NewResults(ctx)
	defer logtool.SpendTimeRecord(logger, "reconcile experiment")()

	expr, err := r.fetchExperiment(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, "fetch experiment failed")
		return ctrl.Result{}, err
	}

	// expr deleted
	if expr == nil || !expr.DeletionTimestamp.IsZero() {
		logger.Info("experiment has deleted, publish topic")
		eventbus.Publish(eventbus.ExperimentDeletedTopic, req.NamespacedName)
		return ctrl.Result{}, nil
	}

	status := experiment.NewStatus(expr)
	result.WithResult((&experiment.Controller{
		Client: r.Client,
		Logger: logger.WithName("ExperimentController"),
	}).Reconcile(ctx, status))
	err = r.updateStatus(ctx, status)
	if err != nil {
		logger.Error(err, "update experiment status failed")
	}
	return result.WithError(err).Aggregate()
}

func (r *ExperimentReconciler) fetchExperiment(ctx context.Context, name types.NamespacedName) (*hackathonv1.Experiment, error) {
	expr := &hackathonv1.Experiment{}
	err := r.Client.Get(ctx, name, expr)
	if errors.IsNotFound(err) {
		return nil, nil
	}
	return expr, err
}

func (r *ExperimentReconciler) updateStatus(ctx context.Context, status *experiment.Status) error {
	log := r.Log.WithValues("name", status.Experiment.Name, "namespace", status.Experiment.Namespace)
	events, crt := status.Apply()
	if crt == nil {
		log.Info("not need update status")
		return nil
	}

	for _, evt := range events {
		r.Recorder.Event(crt, evt.EventType, evt.Reason, evt.Message)
	}

	log.Info("update experiment status")
	err := r.Client.Status().Update(ctx, crt)
	if err != nil && errors.IsConflict(err) {
		log.Info("update experiment status conflict, retry.")
		newCrt := &hackathonv1.Experiment{}
		e := r.Client.Get(ctx, types.NamespacedName{
			Namespace: crt.Namespace,
			Name:      crt.Name,
		}, newCrt)
		if e != nil {
			return e
		}
		newCrt.Status = crt.Status
		return r.Client.Status().Update(ctx, newCrt)
	}
	return err
}

func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hackathonv1.Experiment{}).
		Owns(&corev1.Pod{}).
		Complete(r)
}
