package controller

import (
	"fmt"

	"github.com/NoFacePeace/github/repositories/go/mini/kubebuilder/manager"
)

type Builder struct {
	mgr manager.Manager
	ctr Controller
}

func NewControllerManagedBy(m manager.Manager) *Builder {
	return &Builder{
		mgr: m,
	}
}

func (blder *Builder) Complete(r Reconciler) error {
	_, err := blder.Build(r)
	if err != nil {
		return fmt.Errorf("builder build error: [%w]", err)
	}
	return nil
}

func (blder *Builder) Build(r Reconciler) (Controller, error) {
	if err := blder.doController(r); err != nil {
		return nil, fmt.Errorf("builder do controller error: [%w]", err)
	}
	if err := blder.doWatch(); err != nil {
		return nil, fmt.Errorf("builder do watch error: [%w]", err)
	}
	return nil, nil
}

func (blder *Builder) doController(r Reconciler) error {
	opt := Options{}
	opt.Reconciler = r

	ctrl, err := New(blder.mgr, opt)
	if err != nil {
		return fmt.Errorf("new controller error: [%w]", err)
	}
	blder.ctr = ctrl
	return nil
}

func (blder *Builder) doWatch() error {
	return nil
}
