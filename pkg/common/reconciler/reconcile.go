package reconciler

import (
	"context"
	"fmt"
	"github.com/kaiyuanshe/cloudengine/pkg/utils/logtool"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func ReconcileResource(ctx context.Context, config *ResourceConfig) error {
	logger := config.Logger
	defer logtool.SpendTimeRecord(logger, "reconcile resource")()
	if err := validate(config); err != nil {
		logger.Error(err, "config invalid")
		return fmt.Errorf("config invalid")
	}
	metaObj, err := meta.Accessor(config.Expected)
	if err != nil {
		logger.Error(err, "build meta accessor failed")
		return err
	}
	gvk, err := apiutil.GVKForObject(config.Expected, scheme.Scheme)
	if err != nil {
		logger.Error(err, "get resource GVK failed")
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
	logger.Info("do reconcile action",
		"needRecreate", needRecreate,
		"needUpdate", needUpdate,
		"kind", kind,
	)

	if config.Owner != nil {
		config.Logger.Info("set owner ref")
		if err = controllerutil.SetControllerReference(config.Owner, metaObj, scheme.Scheme); err != nil {
			logger.Error(err, "set owner reference failed")
			return err
		}
	}

	defer func() {
		_ = config.Client.Get(ctx, types.NamespacedName{
			Namespace: metaObj.GetNamespace(),
			Name:      metaObj.GetName(),
		}, config.Reconciled)
	}()

	if err = config.Client.Get(ctx, types.NamespacedName{
		Namespace: metaObj.GetNamespace(),
		Name:      metaObj.GetName(),
	}, config.Reconciled); err != nil {
		if errors.IsNotFound(err) {
			logger.Info("resource not found, do create", "query", types.NamespacedName{Namespace: metaObj.GetNamespace(), Name: metaObj.GetName()}.String())
			return create(ctx, config)
		}
		logger.Error(err, "query resource failed")
	}

	rMetaObj, err := meta.Accessor(config.Reconciled)
	if err != nil {
		logger.Error(err, "build reconciled resource meta accessor failed")
		return err
	}

	if needRecreate {
		logger.Info("need delete and recreate resource")
		rUid := rMetaObj.GetUID()
		rVersion := rMetaObj.GetResourceVersion()
		pcOpt := client.Preconditions{
			UID:             &rUid,
			ResourceVersion: &rVersion,
		}
		if err = config.Client.Delete(ctx, config.Reconciled, pcOpt); err != nil {
			if !errors.IsNotFound(err) {
				logger.Error(err, "delete old resource failed")
				return err
			}
		}
		return create(ctx, config)
	}

	if !needUpdate {
		return nil
	}

	crtVersion := rMetaObj.GetResourceVersion()
	logger.Info("need update resource", "currentVersion", crtVersion)

	if config.PreUpdateHook != nil {
		if err = config.PreUpdateHook(); err != nil {
			logger.Error(err, "do pre update hook failed")
			return err
		}
	}

	rMetaObj.SetResourceVersion(crtVersion)
	if err = config.Client.Update(ctx, config.Reconciled); err != nil {
		logger.Error(err, "update reconciled resource failed")
		return err
	}

	if config.PostUpdateHook != nil {
		if err = config.PreUpdateHook(); err != nil {
			logger.Error(err, "post update hook failed")
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

	err := config.Client.Create(ctx, config.Expected)
	if err != nil {
		config.Logger.Error(err, "create resource failed")
		return err
	}
	return nil
}

func validate(config *ResourceConfig) error {
	return nil
}
