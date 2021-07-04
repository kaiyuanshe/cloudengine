package customcluster

import (
	hackathonv1 "github.com/kaiyuanshe/cloudengine/api/v1"
	"github.com/kaiyuanshe/cloudengine/pkg/common/event"
	"reflect"
)

type Status struct {
	*event.Recorder
	Cluster *hackathonv1.CustomCluster
	Status  *hackathonv1.CustomClusterStatus
}

func (s *Status) Apply() ([]event.Event, *hackathonv1.CustomCluster) {
	pre, crt := s.Cluster.Status, s.Status
	if reflect.DeepEqual(pre, crt) {
		return s.Events, nil
	}
	cluster := s.Cluster
	cluster.Status = *crt
	return s.Events, cluster
}

func NewStatus(cluster *hackathonv1.CustomCluster) *Status {
	return &Status{
		Recorder: event.NewEventRecorder(),
		Cluster:  cluster,
		Status:   cluster.Status.DeepCopy(),
	}
}
