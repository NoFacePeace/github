/*
Copyright 2025.

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
	"fmt"
	"reflect"
	"time"

	v1 "localhost/learn/api/v1"
	webappv1 "localhost/learn/api/v1"

	appsv1 "k8s.io/api/apps/v1"

	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

//

const (
	ConditionTypeAvailable = "Available"
)

// LearnReconciler reconciles a Learn object
type LearnReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=webapp.localhost,resources=learns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=webapp.localhost,resources=learns/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=webapp.localhost,resources=learns/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Learn object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.18.4/pkg/reconcile
func (r *LearnReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// _ = log.FromContext(ctx)

	// TODO(user): your logic here
	// log 带有上下文
	logger := log.FromContext(ctx)

	cr := &v1.Learn{}
	// 检查 cr 是否存在
	if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
		// 删除 cr 导致不存在
		if errors.IsNotFound(err) {
			logger.Info(fmt.Sprintf("cr %s is not found", req.String()))
			return ctrl.Result{}, nil
		}
		logger.Error(err, fmt.Sprintf("cr %s get error", req.String()))
		return ctrl.Result{}, err
	}
	logger.Info(fmt.Sprintf("cr %s is found", req.String()))

	// 设置 cr 状态
	if len(cr.Status.Conditions) == 0 {
		meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
			Type:    ConditionTypeAvailable,
			Status:  metav1.ConditionUnknown,
			Reason:  "Reconciling",
			Message: "Starting reconciliation",
		})
		if err := r.Status().Update(ctx, cr); err != nil {
			logger.Error(err, fmt.Sprintf("cr %s status update error", req.String()))
			return ctrl.Result{}, err
		}
		// 重新获取，防止后续更新报错，旧的 CR 版本已经落后
		if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
			logger.Error(err, "cr %s get error", req.String())
			return ctrl.Result{}, err
		}
		logger.Info(fmt.Sprintf("cr %s set status condition success", req.String()))
	}
	deploy := &appsv1.Deployment{}
	// 不存在则创建
	if err := r.Get(ctx, req.NamespacedName, deploy); err != nil {
		if !errors.IsNotFound(err) {
			logger.Error(err, fmt.Sprintf("deployment %s get error", req.String()))
			return ctrl.Result{}, err
		}
		deploy, err := r.deploymentForCR(cr)
		if err != nil {
			// 生成 deploy 失败，设置状态
			logger.Error(err, fmt.Sprintf("deployment %s deployment for cr error", req.String()))
			meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
				Type:    ConditionTypeAvailable,
				Status:  metav1.ConditionFalse,
				Reason:  "Reconciling",
				Message: fmt.Sprintf("deployment %s deployment for cr error %s", req.String(), err),
			})
			if err := r.Status().Update(ctx, cr); err != nil {
				logger.Error(err, fmt.Sprintf("cr %s status update error", req.String()))
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
		logger.Info(fmt.Sprintf("deployment %s create", req.String()))
		if err := r.Create(ctx, deploy); err != nil {
			logger.Error(err, fmt.Sprintf("deployment %s create errror", req.String()))
			return ctrl.Result{}, err
		}
		// 重新 reconcile，查看创建状态
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}
	// 存在就更新
	size := cr.Spec.Size
	if *deploy.Spec.Replicas != size {
		deploy.Spec.Replicas = &size
		if err := r.Update(ctx, deploy); err != nil {
			logger.Error(err, fmt.Sprintf("deployment %s update error", req.String()))
			// 更新失败，设置状态
			if err := r.Get(ctx, req.NamespacedName, cr); err != nil {
				logger.Error(err, fmt.Sprintf("cr %s get error", req.String()))
				return ctrl.Result{}, err
			}
			meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
				Type:    ConditionTypeAvailable,
				Status:  metav1.ConditionFalse,
				Reason:  "Resizing",
				Message: fmt.Sprintf("deployment %s update error: %s", req.String(), err),
			})
			if err := r.Status().Update(ctx, cr); err != nil {
				logger.Error(err, fmt.Sprintf("cr %s status update error", req.String()))
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, err
		}
		// 重新 reconcile，查看更新状态
		return ctrl.Result{Requeue: true}, nil
	}
	meta.SetStatusCondition(&cr.Status.Conditions, metav1.Condition{
		Type:    ConditionTypeAvailable,
		Status:  metav1.ConditionTrue,
		Reason:  "Reconciling",
		Message: fmt.Sprintf("Deployment for custom resource (%s) with %d replicas created successfully", req.String(), size),
	})
	if err := r.Status().Update(ctx, cr); err != nil {
		logger.Error(err, fmt.Sprintf("%s status update error", req.String()))
		return ctrl.Result{}, nil
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *LearnReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		// crd 类型
		For(&webappv1.Learn{}).
		// 子资源
		Owns(&appsv1.Deployment{}).
		// 过滤事件
		WithEventFilter(eventFilter()).
		Complete(r)
}

func eventFilter() predicate.Predicate {
	basic := func(obj client.Object) []any {
		return []any{
			"namespace",
			obj.GetNamespace(),
			"name",
			obj.GetName(),
			"kind",
			reflect.TypeOf(obj),
		}
	}
	return predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			log.Log.Info("update event", basic(e.ObjectOld)...)
			return true
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			log.Log.Info("delete event", basic(e.Object)...)
			return true
		},
		CreateFunc: func(e event.CreateEvent) bool {
			log.Log.Info("create event", basic(e.Object)...)
			return true
		},
		GenericFunc: func(e event.GenericEvent) bool {
			log.Log.Info("generic event", basic(e.Object)...)
			return true
		},
	}
}

func (r *LearnReconciler) deploymentForCR(learn *v1.Learn) (*appsv1.Deployment, error) {
	deploy := &appsv1.Deployment{
		// 必须设置，没有设置报错 cluster-scoped resource must not have a namespace-scoped owner, owner's namespace default
		ObjectMeta: metav1.ObjectMeta{
			Namespace: learn.Namespace,
			Name:      learn.Name,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &learn.Spec.Size,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app.kubernetes.io/name": "project",
				},
			},
			Template: coreV1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app.kubernetes.io/name": "project",
					},
				},
				Spec: coreV1.PodSpec{
					Containers: []coreV1.Container{{
						// require
						Image: "nginx",
						// require
						Name: "nginx",
						Ports: []coreV1.ContainerPort{{
							ContainerPort: 80,
							Name:          "nginx",
						}},
					}},
				},
			},
		},
	}
	if err := ctrl.SetControllerReference(learn, deploy, r.Scheme); err != nil {
		return nil, fmt.Errorf("controller runtime set controller reference error: [%s]", err)
	}
	return deploy, nil
}
