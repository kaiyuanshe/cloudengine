package experiment

import (
	"fmt"
	hackathonv1 "github.com/kaiyuanshe/cloudengine/api/v1"
	"github.com/kaiyuanshe/cloudengine/pkg/common/event"
	corev1 "k8s.io/api/core/v1"
	"reflect"
)

type Status struct {
	*event.Recorder
	Experiment *hackathonv1.Experiment
	Status     *hackathonv1.ExperimentStatus
}

func (s *Status) UpdateExperimentStatus(state *ResourceState) {
	if state.IngressSvc != nil && state.IngressSvc.Spec.Type == corev1.ServiceTypeNodePort {
		ingressIps := make([]string, 0)
		if len(state.IngressSvc.Spec.ExternalIPs) > 0 {
			ingressIps = state.IngressSvc.Spec.ExternalIPs
		}
		s.Status.IngressIPs = ingressIps
		if len(state.IngressSvc.Spec.Ports) == 1 {
			s.Status.IngressPort = state.IngressSvc.Spec.Ports[0].NodePort
		} else {
			s.AddEvent(corev1.EventTypeWarning, "NoIngressPortFound", fmt.Sprintf("got ingress port: %d", len(state.IngressSvc.Spec.Ports)))
		}
	}

	// update connect config
	s.Status.Protocol = state.Template.Data.IngressProtocol
	switch state.Template.Data.IngressProtocol {
	case hackathonv1.ExperimentIngressVNC:
		s.Status.VNC = state.Template.Data.VNC
	case hackathonv1.ExperimentIngressSSH:
		s.Status.SSH = state.Template.Data.SSH
	default:
		s.AddEvent(corev1.EventTypeWarning, "NoIngressConfig", fmt.Sprintf("ingress protoco %s not supported", state.Template.Data.IngressProtocol))
	}

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
