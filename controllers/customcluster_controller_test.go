package controllers

import (
	hackathonv1 "cloudengine/api/v1"
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("test-custom-cluster-reconcile", func() {
	var (
		cluster *hackathonv1.CustomCluster
	)

	BeforeEach(func() {
		cluster = &hackathonv1.CustomCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-cluster",
			},
			Spec: hackathonv1.CustomClusterSpec{},
		}
	})

	Context("when-new-cluster", func() {
		It("cluster-init", func() {
			err := k8sClient.Create(context.TODO(), cluster)
			Expect(err).ToNot(HaveOccurred())

			created := &hackathonv1.CustomCluster{}
			Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
				Namespace: "default",
				Name:      "test-cluster",
			}, created)).Should(Succeed())

			Expect(created.Name).To(Equal(cluster.Name))
			Expect(created.Status.Status).To(Equal(hackathonv1.ClusterCreated))

			Expect(hackathonv1.CheckClusterCondition(created.Status.Conditions, hackathonv1.ClusterInit, hackathonv1.ClusterStatusTrue)).Should(BeTrue())
		})
	})
})
