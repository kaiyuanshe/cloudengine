package experiment

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	hackathonv1 "github.com/kaiyuanshe/cloudengine/api/v1"
	"github.com/kaiyuanshe/cloudengine/pkg/common/event"
	"github.com/kaiyuanshe/cloudengine/pkg/common/results"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type IngressService struct {
	client        client.Client
	status        *Status
	resourceState *ResourceState
	logger        logr.Logger
}

func (r *IngressService) Reconcile(ctx context.Context) *results.Results {
	var (
		old     = r.resourceState.IngressSvc
		expr    = r.status.Experiment
		tmpl    = r.resourceState.Template
		cluster = r.resourceState.Cluster
	)
	result := results.NewResults(ctx)

	if old != nil {
		old.Spec.Type = corev1.ServiceTypeNodePort
		old.Spec.ExternalIPs = cluster.Spec.PublishIps
		for i := range old.Spec.Ports {
			if old.Spec.Ports[i].Name == string(tmpl.Data.IngressProtocol) {
				old.Spec.Ports[i].Protocol = corev1.ProtocolTCP
				old.Spec.Ports[i].Port = tmpl.Data.IngressPort
				old.Spec.Ports[i].TargetPort = intstr.FromInt(int(tmpl.Data.EndpointPort))
			}
		}
		return result.WithError(r.client.Update(ctx, old))
	}

	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ingressServiceName(expr.Name, tmpl.Data.IngressProtocol),
			Namespace: expr.Namespace,
			Labels: map[string]string{
				LabelKeyClusterName:    expr.Spec.ClusterName,
				LabelKeyExperimentName: expr.Name,
			},
		},
		Spec: corev1.ServiceSpec{
			Type:        corev1.ServiceTypeNodePort,
			ExternalIPs: cluster.Spec.PublishIps,
			Ports: []corev1.ServicePort{
				{
					Name:       string(tmpl.Data.IngressProtocol),
					Protocol:   corev1.ProtocolTCP,
					Port:       tmpl.Data.IngressPort,
					TargetPort: intstr.FromInt(int(tmpl.Data.EndpointPort)),
				},
			},
		},
	}

	err := controllerutil.SetControllerReference(expr, service.GetObjectMeta(), scheme.Scheme)
	if err != nil {
		r.status.AddEvent(corev1.EventTypeWarning, event.ReasonCreated, "create ingress service failed")
		return result.WithError(fmt.Errorf("set ingress service owner ref failed: %s", err.Error()))
	}

	if err = r.client.Create(ctx, service); err != nil {
		r.status.AddEvent(corev1.EventTypeWarning, event.ReasonCreated, fmt.Sprintf("create ingress service failed: %s", err.Error()))
		return result.WithError(err)
	}
	r.status.AddEvent(corev1.EventTypeNormal, event.ReasonCreated, "create ingress service")
	return result
}

func ingressServiceName(exprName string, protocol hackathonv1.ExperimentIngressProtocol) string {
	return fmt.Sprintf("%s-%s-service", exprName, protocol)
}
