package core

import (
	"context"
	"errors"
	"fmt"
	"time"

	umov1 "nofacepeace.github.io/controller/api/v1"
	apiclient "nofacepeace.github.io/controller/pkg/client"
	"nofacepeace.github.io/controller/pkg/config"
	"nofacepeace.github.io/controller/pkg/extensions/event"
	"nofacepeace.github.io/controller/pkg/metrics"
	"nofacepeace.github.io/controller/pkg/model"

	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type Reconciler struct {
	client.Client
	nameToEventChan map[string]chan *ReconcilerEvent
	nm              *NodeManager
	cm              *ClusterManager
	em              *EventManager
	om              *OperationManager
	sm              *StatusManager
	svm             *ServiceManager
	pm              *PodManager
}

type ReconcilerEvent struct {
	ctx     context.Context
	cluster *umov1.Middleware
}

func NewReconciler() *Reconciler {
	r := &Reconciler{}
	return r
}

func (c *Reconciler) Reconcile(ctx context.Context, cls *umov1.Middleware) error {
	logger := logf.FromContext(ctx)
	name := cls.GetName()
	ch := c.nameToEventChan[name]
	// 如果通道不存在，则创建通道
	if ch == nil {
		ch = make(chan *ReconcilerEvent, 100)
		c.nameToEventChan[name] = ch
		go func() {
			for event := range ch {
				for i := 0; i < 5; i++ {
					err := c.processEvent(event)
					retry := i > 0
					metrics.RecordClusterReconcile(cls.Spec.MiddlewareType, config.Get().Eks.Id, cls.Name, retry, err)
					if err == nil {
						break
					}
					logger.Error(err, "reconcile process event error")
					time.Sleep(10 * time.Second)
				}
			}
		}()
	}
	select {
	case ch <- &ReconcilerEvent{
		ctx:     ctx,
		cluster: cls,
	}:
		logger.Info("")
	default:
		logger.Error(errors.New("channel full error"), "")
	}
	return nil
}

func (c *Reconciler) processEvent(e *ReconcilerEvent) error {
	ctx := e.ctx
	cls := e.cluster
	logger := logf.FromContext(ctx)
	logger.Info("cluster reconciler start process event")
	// 检查集群是否存在
	found, err := c.clusterExists(ctx, cls)
	if err != nil {
		return fmt.Errorf("cluster reconciler cluster exists error: [%w]", err)
	}
	if !found {
		ch, ok := c.nameToEventChan[cls.GetName()]
		if ok {
			delete(c.nameToEventChan, cls.GetName())
			close(ch)
			c.em.Dispatch(ctx, cls.GetName(), model.EventTypeClusterDelete, nil)
		}
		apiclient.DeleteClusterPublishStatus(cls.GetName())
		return nil
	}
	return c.checkCluster(e)
}

func (c *Reconciler) checkCluster(e *ReconcilerEvent) (err error) {
	skip, err := c.cm.Skip(e.ctx, e.cluster)
	if err != nil {
		return fmt.Errorf("cluster manager skip error: [%w]", err)
	}
	if skip {
		return nil
	}

	defer func() {
		c.om.NewOperationBuilder(e.ctx, e.cluster.Name, OperationTypeCheckCluster).WithError(err).Report()
		if subErr := c.sm.UpdateClusterStatus(e.ctx, e.cluster); subErr != nil {
			err = errors.Join(err, subErr)
		}
	}()

	if e.cluster.Status.Phase == "" {
		if err := c.sm.updateClusterPhase(e.ctx, e.cluster, umov1.MiddlewarePending); err != nil {
			return fmt.Errorf("status manager update cluster phase error: [%w]", err)
		}
	}
	if err := c.svm.CheckService(e.ctx, e.cluster); err != nil {
		return fmt.Errorf("service manager check service error: [%w]", err)
	}
	if err := c.nm.checkNodes(e.ctx, e.cluster); err != nil {
		return fmt.Errorf("check nodes error: [%w]", err)
	}
	pods, err := c.pm.GetClusterPods(e.ctx, e.cluster)
	if err != nil {
		return fmt.Errorf("pod manager get cluster pods error: [%w]", err)
	}
	if e.cluster.Status.Phase == umov1.MiddlewarePending {
		if err := c.sm.updateClusterPhase(e.ctx, e.cluster, umov1.MiddlewareRunning); err != nil {
			return fmt.Errorf("status manager update cluster phase error: [%w]", err)
		}
		nodes := make([]*event.Node, 0, len(pods))
		for _, pod := range pods {
			nodes = append(nodes, &event.Node{
				Name:   pod.Name,
				OldPod: nil,
				NewPod: &pod,
			})
		}
		c.em.Dispatch(e.ctx, e.cluster.GetName(), model.EventTypeClusterCreate, nodes)
	}
	if err := c.nm.ScaleDown(e.ctx, e.cluster, pods); err != nil {
		return fmt.Errorf("node manager scale down error: [%w]", err)
	}
	if err := c.sm.updateClusterChecksum(e.ctx, e.cluster); err != nil {
		return fmt.Errorf("status manager update cluster checksum error: [%w]", err)
	}
	c.em.Dispatch(e.ctx, e.cluster.GetName(), model.EventTypeClusterReconcile, nil)
	return nil
}

func (c *Reconciler) clusterExists(ctx context.Context, cls *umov1.Middleware) (bool, error) {
	_, found, err := c.cm.Exist(ctx, cls.Namespace, cls.Name)
	if err != nil {
		return false, fmt.Errorf("cluster manager exist error: [%w]", err)
	}
	return found, nil
}
