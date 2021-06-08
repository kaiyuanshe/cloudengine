package experiment

import (
	hackathonv1 "cloudengine/api/v1"
	"cloudengine/pkg/common/event"
	"cloudengine/pkg/utils/k8stools"
	"reflect"
)

type Status struct {
	*event.Recorder
	Experiment *hackathonv1.Experiment
	Status     *hackathonv1.ExperimentStatus
}

func (s *Status) UpdateExperimentStatus(state *ResourceState) {
	s.Status.Status = hackathonv1.ExperimentCreated

	if len(state.EnvPod) > 0 {
		pod := state.EnvPod[0]
		if k8stools.IsPodReady(&pod) {
			s.Status.Conditions = hackathonv1.UpdateExperimentConditions(
				s.Status.Conditions,
				hackathonv1.NewExperimentCondition(
					hackathonv1.ExperimentPodReady,
					hackathonv1.ExperimentConditionTrue, "", ""))
		}
	}

	resources := []hackathonv1.ExperimentConditionType{
		hackathonv1.ExperimentPodReady,
	}

	allSafe := true
	for _, res := range resources {
		cond := hackathonv1.QueryExperimentCondition(s.Status.Conditions, res)
		if cond == nil {
			allSafe = true
			continue
		}

		if cond.Status == hackathonv1.ExperimentConditionFalse {
			allSafe = true
			s.Status.Status = hackathonv1.ExperimentError
			s.Status.Conditions = hackathonv1.UpdateExperimentConditions(
				s.Status.Conditions,
				hackathonv1.NewExperimentCondition(
					hackathonv1.ExperimentReady,
					hackathonv1.ExperimentConditionFalse, cond.Reason, cond.Message))
			return
		}
	}

	if allSafe {
		s.Status.Status = hackathonv1.ExperimentRunning
		s.Status.Conditions = hackathonv1.UpdateExperimentConditions(
			s.Status.Conditions,
			hackathonv1.NewExperimentCondition(
				hackathonv1.ExperimentReady,
				hackathonv1.ExperimentConditionTrue, "", ""))
	}
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
