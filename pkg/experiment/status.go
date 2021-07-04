package experiment

import (
	hackathonv1 "github.com/kaiyuanshe/cloudengine/api/v1"
	"github.com/kaiyuanshe/cloudengine/pkg/common/event"
	"reflect"
)

type Status struct {
	*event.Recorder
	Experiment *hackathonv1.Experiment
	Status     *hackathonv1.ExperimentStatus
}

func (s *Status) UpdateExperimentStatus(state *ResourceState) {
	s.Status.Cluster = s.Experiment.Spec.ClusterName
	s.Status.ClusterSync = false
}

func (s *Status) Apply() ([]event.Event, *hackathonv1.Experiment) {
	pre, crt := s.Experiment.Status, s.Status
	if reflect.DeepEqual(pre, crt) {
		return s.Events, nil
	}
	expr := s.Experiment
	expr.Status = *crt
	return s.Events, expr
}

func NewStatus(expr *hackathonv1.Experiment) *Status {
	return &Status{
		Recorder:   event.NewEventRecorder(),
		Experiment: expr,
		Status:     expr.Status.DeepCopy(),
	}
}
