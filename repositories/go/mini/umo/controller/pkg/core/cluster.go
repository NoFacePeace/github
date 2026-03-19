package core

import (
	"context"
	"fmt"
	"strings"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	umov1 "nofacepeace.github.io/controller/api/v1"
	"nofacepeace.github.io/controller/pkg/config"
	"nofacepeace.github.io/controller/pkg/model"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type ClusterManager struct {
	client.Client
	sm *StatusManager
}

func (c *ClusterManager) Exist(ctx context.Context, ns, name string) (*umov1.Middleware, bool, error) {
	cls, err := c.get(ctx, ns, name)
	if err == nil {
		return cls, true, nil
	}
	if k8sErrors.IsNotFound(err) {
		return nil, false, nil
	}
	return nil, false, fmt.Errorf("cluster manager get error: [%w]", err)
}

func (c *ClusterManager) get(ctx context.Context, ns, name string) (*umov1.Middleware, error) {
	cls := &umov1.Middleware{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: ns,
			Name:      name,
		},
	}
	return cls, c.Client.Get(ctx, client.ObjectKeyFromObject(cls), cls)
}

func (c *ClusterManager) Skip(ctx context.Context, cls *umov1.Middleware) (bool, error) {
	logger := logf.FromContext(ctx)
	// 1. config skip
	if config.Get().ClusterFilterPolicy.Skip(cls.GetName()) {
		logger.Info("cluster manager skip cluster", "cluster", cls.GetName())
		return true, nil
	}
	labels := cls.GetLabels()
	if labels == nil {
		logger.Info("cluster manager no labels", "cluster", cls.GetName())
		return false, nil
	}
	// 2. label skip
	if strings.EqualFold(labels[model.LabelClusterIgnore], "true") {
		logger.Info("cluster manager ignore cluster", "cluster", cls.GetName())
		return true, nil
	}
	if strings.EqualFold(labels[model.LabelManualManagement], "true") {
		logger.Info("cluster manager manual management", "cluster", cls.GetName())
		return true, c.sm.UpdateStatus(ctx, cls)
	}
	return false, nil
}
