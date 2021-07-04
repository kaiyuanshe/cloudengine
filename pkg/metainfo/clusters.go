package metainfo

import (
	"context"
	"fmt"
	hackathonv1 "github.com/kaiyuanshe/cloudengine/api/v1"
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
		err = client.Update(ctx, cluster)
		if err != nil {
			return fmt.Errorf("updatee cluster annotations failed")
		}
	}
	return nil
}
