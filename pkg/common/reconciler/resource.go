package reconciler

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
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

func DoReconcile(ctx context.Context, config *ResourceConfig) error {
	if err := validate(config); err != nil {
		return fmt.Errorf("config invalid")
	}
	metaObj, err := meta.Accessor(config.Expected)
	if err != nil {
		return err
	}
	gvk, err := apiutil.GVKForObject(config.Expected, scheme.Scheme)
	if err != nil {
		return err
	}
	kind := gvk.Kind

	needRecreate, needUpdate := false, false
	if config.NeedRecreate != nil {
		needRecreate = config.NeedRecreate()
	}
	if config.NeedUpdate != nil {
		needUpdate = config.NeedUpdate()
	}
	config.Logger.Info("do reconcile action",
		"needRecreate", needRecreate,
		"needUpdate", needUpdate,
		"kind", kind,
		"name", metaObj.GetName(),
		"namespace", metaObj.GetNamespace(),
	)

	if config.Owner != nil {
		config.Logger.Info("set owner ref")
		if err = controllerutil.SetControllerReference(config.Owner, metaObj, scheme.Scheme); err != nil {
			return err
		}
	}

	defer func() {
		_ = config.Client.Get(ctx, client.ObjectKey{
			Namespace: metaObj.GetNamespace(),
			Name:      metaObj.GetName(),
		}, config.Reconciled)
	}()

	if err = config.Client.Get(ctx, client.ObjectKey{
		Namespace: metaObj.GetNamespace(),
		Name:      metaObj.GetName(),
	}, config.Reconciled); err != nil {
		if errors.IsNotFound(err) {
			return create(ctx, config)
		}
		config.Logger.Error(err, "query resource failed")
	}

	rMetaObj, err := meta.Accessor(config.Reconciled)
	if err != nil {
		return err
	}

	if needRecreate {
		config.Logger.Info("need delete and recreate resource")
		rUid := rMetaObj.GetUID()
		rVersion := rMetaObj.GetResourceVersion()
		pcOpt := client.Preconditions{
			UID:             &rUid,
			ResourceVersion: &rVersion,
		}
		if err = config.Client.Delete(ctx, config.Reconciled, pcOpt); err != nil {
			if !errors.IsNotFound(err) {
				config.Logger.Error(err, "delete old resource failed")
				return err
			}
		}
		return create(ctx, config)
	}

	if !needUpdate {
		return nil
	}

	crtVersion := rMetaObj.GetResourceVersion()
	config.Logger.Info("need update resource", "currentVersion", crtVersion)

	if config.PreUpdateHook != nil {
		if err = config.PreUpdateHook(); err != nil {
			config.Logger.Error(err, "pre update hook failed")
			return err
		}
	}

	rMetaObj.SetResourceVersion(crtVersion)
	if err = config.Client.Update(ctx, config.Reconciled); err != nil {
		config.Logger.Error(err, "update reconciled resource failed")
		return err
	}

	if config.PostUpdateHook != nil {
		if err = config.PreUpdateHook(); err != nil {
			config.Logger.Error(err, "post update hook failed")
			return err
		}
	}

	return nil
}

func create(ctx context.Context, config *ResourceConfig) error {
	config.Logger.Info("create resource")

	if config.PreCreateHook != nil {
		if err := config.PreCreateHook(); err != nil {
			config.Logger.Error(err, "pre create hook failed")
			return err
		}
	}

	return config.Client.Create(ctx, config.Expected)
}

func validate(config *ResourceConfig) error {
	return nil
}
