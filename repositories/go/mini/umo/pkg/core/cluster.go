package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	umov1 "localhost/umo/api/v1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
)

type ClusterReconciler struct {
	client.Client
	nameToEventChan map[string]chan *ClusterReconcilerEvent
	nom             *NodeManager
	cm              *ClusterManager
}

type ClusterReconcilerEvent struct {
	ctx     context.Context
	cluster *umov1.Middleware
}

func NewClusterReconciler() *ClusterReconciler {
	r := &ClusterReconciler{}
	return r
}

func (c *ClusterReconciler) Reconcile(ctx context.Context, cls *umov1.Middleware) error {
	logger := logf.FromContext(ctx)
	name := cls.GetName()
	ch := c.nameToEventChan[name]
	if ch == nil {
		ch = make(chan *ClusterReconcilerEvent, 100)
		c.nameToEventChan[name] = ch
		go func() {
			for event := range ch {
				for i := 0; i < 5; i++ {
					if err := c.processEvent(event); err != nil {
						logger.Error(err, "reconcile process event error")
						time.Sleep(10 * time.Second)
						continue
					}
					break
				}
			}
		}()
	}
	select {
	case ch <- &ClusterReconcilerEvent{
		ctx:     ctx,
		cluster: cls,
	}:
		logger.Info("")
	default:
		logger.Error(errors.New("channel full error"), "")
	}
	return nil
}

func (c *ClusterReconciler) processEvent(e *ClusterReconcilerEvent) error {
	ctx := e.ctx
	cls := e.cluster
	logger := logf.FromContext(ctx)
	logger.Info("cluster reconciler start process event")
	found, err := c.clusterExists(ctx, cls)
	if err != nil {
		return fmt.Errorf("cluster reconciler cluster exists error: [%w]", err)
	}
	if !found {
		ch, ok := c.nameToEventChan[cls.GetName()]
		if ok {
			delete(c.nameToEventChan, cls.GetName())
			close(ch)
		}
	}
	c.checkCluster(e)
	return nil
}

func (c *ClusterReconciler) checkCluster(e *ClusterReconcilerEvent) error {
	c.checkClusterNodes(e)
	return nil
}

func (c *ClusterReconciler) checkClusterNodes(e *ClusterReconcilerEvent) error {
	return nil
}

func (c *ClusterReconciler) clusterExists(ctx context.Context, cls *umov1.Middleware) (bool, error) {
	_, found, err := c.cm.Exist(ctx, cls.Namespace, cls.Name)
	if err != nil {
		return false, fmt.Errorf("cluster manager exist error: [%w]", err)
	}
	return found, nil
}

type ClusterManager struct {
	client.Client
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
