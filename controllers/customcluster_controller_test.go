package controllers

import (
	hackathonv1 "cloudengine/api/v1"
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"time"
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

	Describe("test-custom-cluster-healthy", func() {
		Context("when-cluster-timeout", func() {
			created := &hackathonv1.CustomCluster{}
			timeout := 30
			interval := 5
			BeforeEach(func() {
				By("new timeout cluster")
				Expect(k8sClient.Create(context.TODO(), cluster)).Should(Succeed())

			})

			JustBeforeEach(func() {
				By("wait for init")
				Eventually(func() bool {
					if err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: "default",
						Name:      "test-cluster",
					}, created); err != nil {
						return false
					}
					return created.Status.Status == hackathonv1.ClusterCreated
				}, timeout, interval).Should(BeTrue())
			})

			It("need-be-lost", func() {
				By("update cluster status: like send heartbeat")
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "default", Name: "test-cluster"}, created)).Should(Succeed())
				created.Status.Status = hackathonv1.ClusterReady
				created.Status.Conditions = append(created.Status.Conditions, hackathonv1.ClusterCondition{
					Type:               hackathonv1.ClusterInit,
					Status:             hackathonv1.ClusterStatusTrue,
					LastProbeTime:      metav1.Time{Time: time.Now().Add(-24 * time.Hour)},
					LastTransitionTime: metav1.Time{Time: time.Now().Add(-24 * time.Hour)},
				})
				created.Status.Conditions = append(created.Status.Conditions, hackathonv1.ClusterCondition{
					Type:               hackathonv1.ClusterFirstConnect,
					Status:             hackathonv1.ClusterStatusTrue,
					LastProbeTime:      metav1.Time{Time: time.Now().Add(-24 * time.Hour)},
					LastTransitionTime: metav1.Time{Time: time.Now().Add(-24 * time.Hour)},
				})
				created.Status.Conditions = append(created.Status.Conditions, hackathonv1.ClusterCondition{
					Type:               hackathonv1.ClusterHeartbeat,
					Status:             hackathonv1.ClusterStatusTrue,
					LastProbeTime:      metav1.Time{Time: time.Now().Add(-5 * time.Minute)},
					LastTransitionTime: metav1.Time{Time: time.Now().Add(-5 * time.Minute)},
				})
				Expect(k8sClient.Status().Update(context.TODO(), created)).Should(Succeed())

				By("current cluster is ready")
				latest := &hackathonv1.CustomCluster{}
				Expect(k8sClient.Get(context.TODO(), types.NamespacedName{
					Namespace: cluster.Namespace,
					Name:      cluster.Name,
				}, latest)).Should(Succeed())
				Expect(latest.Status.Status).Should(Equal(hackathonv1.ClusterReady))

				By("trigger a reconcile")
				latest.Spec.ClusterTimeoutSeconds = 1
				Expect(k8sClient.Update(context.TODO(), latest)).Should(Succeed())

				Eventually(func() hackathonv1.ClusterStatus {
					if err := k8sClient.Get(context.TODO(), types.NamespacedName{
						Namespace: "default",
						Name:      "test-cluster",
					}, latest); err != nil {
						return ""
					}
					return latest.Status.Status
				}, timeout, interval).Should(Equal(hackathonv1.ClusterLost))

				Expect(hackathonv1.CheckClusterCondition(latest.Status.Conditions, hackathonv1.ClusterHeartbeat, hackathonv1.ClusterStatusFalse)).Should(BeTrue())
			})
		})

		AfterEach(func() {
			By("clean up custom cluster")
			Expect(k8sClient.Delete(context.TODO(), cluster)).Should(Succeed())
		})
	})

})
