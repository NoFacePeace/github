/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	corev1 "k8s.io/api/core/v1"
	examplev1 "linkinstars.com/op-ex/api/v1"
)

// ExampleAReconciler reconciles a ExampleA object
type ExampleAReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=example.linkinstars.com,resources=examplea,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=example.linkinstars.com,resources=examplea/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=example.linkinstars.com,resources=examplea/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ExampleA object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *ExampleAReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)

	// TODO(user): your logic here
	logger := log.FromContext(ctx)
	logger.Info("开始调用Reconcile方法")

	var exp examplev1.ExampleA
	if err := r.Get(ctx, req.NamespacedName, &exp); err != nil {
		logger.Error(err, "未找到对应的CRD资源")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	exp.Status.UnderControl = false

	var podList corev1.PodList
	if err := r.List(ctx, &podList); err != nil {
		logger.Error(err, "无法获取pod列表")
	} else {
		for _, item := range podList.Items {
			if item.GetLabels()["group"] == exp.Spec.GroupName {
				logger.Info("找到对应的pod资源", "name", item.GetName())
				exp.Status.UnderControl = true
			}
		}
	}

	if err := r.Status().Update(ctx, &exp); err != nil {
		logger.Error(err, "无法更新CRD资源状态")
		return ctrl.Result{}, err
	}
	logger.Info("已更新CRD资源状态", "status", exp.Status.UnderControl)
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ExampleAReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&examplev1.ExampleA{}).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(r.podChangeHandler),
		).
		Complete(r)
}

func (r *ExampleAReconciler) podChangeHandler(ctx context.Context, obj client.Object) []reconcile.Request {
	logger := log.FromContext(ctx)

	var req []reconcile.Request
	var list examplev1.ExampleAList
	if err := r.Client.List(ctx, &list); err != nil {
		logger.Error(err, "无法获取到资源")
	} else {
		for _, item := range list.Items {
			if item.Spec.GroupName == obj.GetLabels()["group"] {
				req = append(req, reconcile.Request{
					NamespacedName: types.NamespacedName{Name: item.Name, Namespace: item.Namespace},
				})
			}
		}
	}
	return req
}
