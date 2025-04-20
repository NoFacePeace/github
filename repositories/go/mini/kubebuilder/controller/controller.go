package controller

import (
	"context"
	"fmt"
	"sync"

	"github.com/NoFacePeace/github/repositories/go/mini/kubebuilder/manager"
)

type Controller interface {
	Start(ctx context.Context) error
}

type controller struct {
	Name string
	Do   Reconciler
}

func New(mgr manager.Manager, options Options) (Controller, error) {
	c, err := NewUnmanaged(mgr, options)
	if err != nil {
		return nil, fmt.Errorf("controller new un managed error: [%w]", err)
	}
	return c, mgr.Add(c)
}

type Options struct {
	Reconciler Reconciler
}

func NewUnmanaged(mgr manager.Manager, options Options) (Controller, error) {
	return &controller{
		Do: options.Reconciler,
	}, nil
}

func (c *controller) Start(ctx context.Context) error {
	var wg sync.WaitGroup
	go func() {
		defer wg.Done()
		for c.processNextWorkItem(ctx) {

		}
	}()
	<-ctx.Done()
	wg.Done()
	return nil
}

func (c *controller) processNextWorkItem(ctx context.Context) bool {
	c.reconcileHandler(ctx)
	return true
}

func (c *controller) reconcileHandler(ctx context.Context) {
	c.Do.Reconcile(ctx)
}
