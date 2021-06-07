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
				Name:      "test-cluster",
				Namespace: "default",
			},
			Spec: hackathonv1.CustomClusterSpec{},
		}
	})

	Context("when-new-cluster", func() {
		It("cluster-init", func() {
			By("apply new custom cluster")
			err := k8sClient.Create(context.TODO(), cluster)
			Expect(err).ToNot(HaveOccurred())

			created := &hackathonv1.CustomCluster{}
			timeout := 30
			interval := 5
			Eventually(func() bool {
				if err := k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: "default",
					Name:      "test-cluster",
				}, created); err != nil {
					return false
				}
				return created.Status.Status == hackathonv1.ClusterCreated
			}, timeout, interval).Should(BeTrue())

			Expect(hackathonv1.CheckClusterCondition(created.Status.Conditions, hackathonv1.ClusterInit, hackathonv1.ClusterStatusTrue)).Should(BeTrue())
		})
		AfterEach(func() {
			By("clean up custom cluster")
			Expect(k8sClient.Delete(context.TODO(), cluster)).Should(Succeed())
		})
	})
})
