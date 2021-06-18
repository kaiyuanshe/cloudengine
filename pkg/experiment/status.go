package experiment

import (
	hackathonv1 "cloudengine/api/v1"
	"cloudengine/pkg/common/event"
	"reflect"
)

type Status struct {
	*event.Recorder
	Experiment *hackathonv1.Experiment
	Status     *hackathonv1.ExperimentStatus
}

// TODO
func (s *Status) UpdateExperimentStatus(state *ResourceState) {
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
