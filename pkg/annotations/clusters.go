package annotations

import (
	hackathonv1 "cloudengine/api/v1"
	"context"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	hackathonCustomClusterId string = "hackathon.kaiyuanshe.cn/cluster-id"
)

func UpdateClusterAnnotations(ctx context.Context, cluster *hackathonv1.CustomCluster, client client.Client) error {
	accessor, err := meta.Accessor(cluster)
	if err != nil {
		return err
	}

	anns := accessor.GetAnnotations()
	if anns == nil {
		anns = map[string]string{}
	}

	needUpdate := false
	// update cluster id
	if cluster.Status.ClusterID != "" {
		needUpdate = true
		anns[hackathonCustomClusterId] = cluster.Status.ClusterID
	}

	if needUpdate {
		accessor.SetAnnotations(anns)
		return client.Update(ctx, cluster)
	}
	return nil
}
