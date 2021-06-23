package reconciler

import (
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DecideFn func() bool
type HookFn func() error

type ResourceConfig struct {
	Client     client.Client
	Owner      metav1.Object
	Expected   runtime.Object
	Reconciled runtime.Object

	NeedUpdate   DecideFn
	NeedRecreate DecideFn

	PreCreateHook  HookFn
	PreUpdateHook  HookFn
	PostUpdateHook HookFn

	Logger logr.Logger
}

