package annotations

import (
	hackathonv1 "cloudengine/api/v1"
	"k8s.io/apimachinery/pkg/api/meta"
)

const (
	hackathonCustomClusterId string = "hackathon.kaiyuanshe.cn/cluster-id"
)

func UpdateClusterAnnotations(cluster *hackathonv1.CustomCluster) error {
	accessor, err := meta.Accessor(cluster)
	if err != nil {
		return err
	}

	anns := accessor.GetAnnotations()

	// update cluster id
	if cluster.Status.ClusterID != "" {
		anns[hackathonCustomClusterId] = cluster.Status.ClusterID
	}

	accessor.SetAnnotations(anns)
	return nil
}
