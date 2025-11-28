package core

import (
	"context"
	"errors"
	"time"

	umov1 "localhost/umo-controller/api/v1"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type ClusterReconciler struct {
	chanHub map[string]chan *ClusterReconcilerEvent
	nom     *NodeManager
}

type ClusterReconcilerEvent struct {
	ctx     context.Context
	cluster *umov1.UmoCluster
}

func (c *ClusterReconciler) Reconcile(ctx context.Context, cls *umov1.UmoCluster) error {
	logger := logf.FromContext(ctx)
	id := cls.GetName()
	ch := c.chanHub[id]
	if ch == nil {
		ch = make(chan *ClusterReconcilerEvent, 100)
		c.chanHub[id] = ch
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
	_ = logf.FromContext(e.ctx)
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
