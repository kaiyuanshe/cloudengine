package experiment

import (
	"cloudengine/pkg/common/results"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	hostPathDir   = "/opt/open-hackathon/cloud-engine/data"
	containerPath = "/data"
	volumeSizeGi  = 10
)

type DataVolume struct {
	client        client.Client
	status        *Status
	resourceState *ResourceState
	logger        logr.Logger
}

func (v *DataVolume) Reconcile() *results.Results {
	return nil
}
