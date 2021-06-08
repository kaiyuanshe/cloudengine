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
	"cloudengine/pkg/common/results"
	"cloudengine/pkg/eventbus"
	"cloudengine/pkg/experiment"
	"cloudengine/pkg/utils/logtool"
	"context"
	"k8s.io/apimachinery/pkg/types"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	hackathonv1 "cloudengine/api/v1"
)

// ExperimentReconciler reconciles a Experiment object
type ExperimentReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=hackathon.kaiyuanshe.cn,resources=experiments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=hackathon.kaiyuanshe.cn,resources=experiments/status,verbs=get;update;patch

func (r *ExperimentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	logger := r.Log.WithValues("experiment", req.NamespacedName)
	result := results.NewResults(ctx)
	defer logtool.SpendTimeRecord(logger, "reconcile experiment")

	expr, err := r.fetchExperiment(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, "fetch experiment failed")
		return ctrl.Result{}, err
	}

	// expr deleted
	if !expr.DeletionTimestamp.IsZero() {
		eventbus.Publish(eventbus.ExperimentDeletedTopic, expr)
		return ctrl.Result{}, nil
	}

	status := experiment.NewStatus(expr)
	result.WithResult((&experiment.Controller{
		Client: r.Client,
		Logger: logger,
	}).Reconcile(ctx, status))
	return result.Aggregate()
}

func (r *ExperimentReconciler) fetchExperiment(ctx context.Context, name types.NamespacedName) (*hackathonv1.Experiment, error) {
	expr := &hackathonv1.Experiment{}
	err := r.Client.Get(ctx, name, expr)
	return expr, err
}

func (r *ExperimentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&hackathonv1.Experiment{}).
		Complete(r)
}
