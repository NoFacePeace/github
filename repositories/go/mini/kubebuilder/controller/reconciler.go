package controller

import (
	"context"

	"github.com/NoFacePeace/github/repositories/go/mini/kubebuilder/manager"
)

type Reconciler interface {
	Reconcile(context.Context) error
}

type ExampleReconciler struct{}

func (er *ExampleReconciler) Reconcile(context.Context) error {
	return nil
}

func (er *ExampleReconciler) SetupWithManager(mgr manager.Manager) error {
	return NewControllerManagedBy(mgr).Complete(er)
}
