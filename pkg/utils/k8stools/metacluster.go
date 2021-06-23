package k8stools

import (
	hackathonv1 "cloudengine/api/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var (
	MetaClusterMark      = "hackathon.kaiyuanshe.cn/meta-cluster"
	MetaClusterName      = "meta-cluster"
	MetaClusterNameSpace = "default"
)

func NewMetaCluster() *hackathonv1.CustomCluster {
	return &hackathonv1.CustomCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MetaClusterName,
			Namespace: MetaClusterNameSpace,
			Labels: map[string]string{
				MetaClusterMark: "",
			},
		},
		Spec: hackathonv1.CustomClusterSpec{
			ClusterTimeoutSeconds: 60,
		},
	}
}
