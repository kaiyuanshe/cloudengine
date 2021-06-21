package controllers

import (
	hackathonv1 "cloudengine/api/v1"
	"cloudengine/pkg/experiment"
	"cloudengine/pkg/utils/k8stools"
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"
)

var _ = Describe("test-experiment-reconcile", func() {
	var (
		expr *hackathonv1.Experiment
		tpl  = &hackathonv1.Template{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-template-for-exper",
				Namespace: "default",
			},
			Data: hackathonv1.TemplateData{
				Type: hackathonv1.PodTemplateType,
				PodTemplate: &hackathonv1.PodTemplate{
					Image:   "bosybox",
					Command: []string{"sh", "-c", "sleep 100000000"},
				},
			},
		}

		cluster = &hackathonv1.CustomCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cluster-for-exper",
				Namespace: "default",
			},
			Spec: hackathonv1.CustomClusterSpec{
				ClusterTimeoutSeconds: 3600,
				PublishIps:            []string{},
				PrivateIps:            []string{},
				EnablePrivateIP:       false,
			},
		}
	)

	BeforeEach(func() {
		By("init cluster and template")
		time.Sleep(5 * time.Millisecond)
		Expect(k8sClient.Create(context.TODO(), tpl)).ToNot(HaveOccurred())
		Expect(k8sClient.Create(context.TODO(), cluster)).ToNot(HaveOccurred())
		expr = &hackathonv1.Experiment{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-expr",
				Namespace: "default",
			},
			Spec: hackathonv1.ExperimentSpec{
				Pause:       false,
				Template:    tpl.Name,
				ClusterName: cluster.Name,
			},
		}
	})

	Context("reconcile-experiment", func() {
		It("create new", func() {
			By("create new experiment cr")
			Expect(k8sClient.Create(context.TODO(), expr)).ToNot(HaveOccurred())

			timeout := 30
			interval := 5
			created := &hackathonv1.Experiment{}
			Eventually(func() hackathonv1.ExperimentEnvStatus {
				Expect(k8sClient.Get(context.TODO(), client.ObjectKey{
					Namespace: expr.Namespace,
					Name:      expr.Name,
				}, created)).ToNot(HaveOccurred())
				return created.Status.Status
			}, timeout, interval).Should(Equal(hackathonv1.ExperimentRunning))

			podList := &corev1.PodList{}
			selector := labels.NewSelector()
			match, _ := labels.NewRequirement(experiment.LabelKeyExperimentName, selection.Equals, []string{expr.Name})
			selector = selector.Add(*match)
			Expect(k8sClient.List(context.TODO(), podList, client.MatchingLabelsSelector{Selector: selector}))

			Expect(len(podList.Items)).Should(Equal(1))
			envPod := podList.Items[0]
			Expect(k8stools.IsPodReady(&envPod)).Should(BeTrue())
		})
	})
	AfterEach(func() {
		_ = k8sClient.Delete(context.TODO(), expr)
		_ = k8sClient.Delete(context.TODO(), tpl)
		_ = k8sClient.Delete(context.TODO(), cluster)

		pv := &corev1.PersistentVolume{
			ObjectMeta: metav1.ObjectMeta{
				Name:      fmt.Sprintf("pv-%s", expr.Name),
				Namespace: "default",
			},
		}
		_ = k8sClient.Delete(context.TODO(), pv)
	})
})
