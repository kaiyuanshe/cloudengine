package k8stools

import (
	"context"
	hackathonv1 "github.com/kaiyuanshe/cloudengine/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	MetaClusterMark      = "hackathon.kaiyuanshe.cn/meta-cluster"
	MetaClusterName      = "meta-cluster"
	MetaClusterNameSpace = "default"
)

func NewMetaCluster(client client.Client) (*hackathonv1.CustomCluster, error) {
	nodeLists := &corev1.NodeList{}
	err := client.List(context.TODO(), nodeLists)
	if err != nil {
		return nil, err
	}

	publicIps, privateIps := GetClusterPublicAndPrivateIps(nodeLists.Items)
	return &hackathonv1.CustomCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      MetaClusterName,
			Namespace: MetaClusterNameSpace,
			Labels: map[string]string{
				MetaClusterMark: "",
			},
		},
		Spec: hackathonv1.CustomClusterSpec{
			PublishIps:            publicIps,
			PrivateIps:            privateIps,
			EnablePrivateIP:       true,
			ClusterTimeoutSeconds: 60,
		},
	}, nil
}

func GetClusterPublicAndPrivateIps(nodes []corev1.Node) (publicIps, privateIps []string) {
	for _, no := range nodes {
		for _, addr := range no.Status.Addresses {
			switch addr.Type {
			case corev1.NodeExternalIP:
				publicIps = append(publicIps, addr.Address)
			case corev1.NodeInternalIP:
				privateIps = append(privateIps, addr.Address)
			}
		}
	}
	return
}
