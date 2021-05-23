package cluster

import (
	v1 "cloudengine/api/v1"
	"cloudengine/pkg/types"
	"fmt"
	"k8s.io/klog"
)

type Server struct {
}

func (s *Server) HandleHeartbeat(hb *types.Heartbeat) (resp *types.HeartbeatResponse, err error) {
	resp = &types.HeartbeatResponse{OK: true}
	var (
		cluster       *v1.CustomCluster
		clusterStatus = hb.Cluster
	)
	cluster, clusterStatus, err = s.GetClusterInfo(clusterStatus)
	resp.Cluster = clusterStatus

	cmd, err := s.BuildLatestCommand()
	if err != nil {
		resp.OK = false
		resp.Message = fmt.Sprintf("sync resource failed: %s", err.Error())
		klog.Errorf("build command failed: %s", err.Error())
		return resp, err
	}
	resp.Command = cmd

	_ = cluster

	return resp, nil
}

func (s *Server) BuildLatestCommand() (*types.Command, error) {
	return nil, nil
}

func (s *Server) GetClusterInfo(status types.ClusterStatus) (*v1.CustomCluster, types.ClusterStatus, error) {
	return nil, types.ClusterStatus{}, nil
}

func (s *Server) resourceStatusHandler(status types.ResourceStatus) {
}
