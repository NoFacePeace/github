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
	"log/slog"
	"sort"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/builder"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	umov1 "localhost/umo-controller/api/v1"
	"localhost/umo-controller/pkg/core"
)

const (
	LabelMiddlewareType = "MiddlewareType"
)

// UmoClusterReconciler reconciles a UmoCluster object
type UmoClusterReconciler struct {
	client.Client
	Scheme               *runtime.Scheme
	clusterIdToNamespace sync.Map
	sync.Mutex
	ClusterReconciler *core.ClusterReconciler
	config            *Config
}

type Config struct {
	MiddlewareType string
}

// +kubebuilder:rbac:groups=umo.localhost,resources=umoclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=umo.localhost,resources=umoclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=umo.localhost,resources=umoclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the UmoCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.22.1/pkg/reconcile
func (r *UmoClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := logf.FromContext(ctx)
	clusterId := req.Name
	ns := req.Namespace

	logger = logger.WithValues("namespace", ns, "cluster id", clusterId, "component", "umo ")

	// 记录 cluster id 与 namespace 映射，详见 LoopCallReconcile
	if _, ok := r.clusterIdToNamespace.Load(clusterId); !ok {
		r.clusterIdToNamespace.Store(clusterId, ns)
	}
	cls := &umov1.UmoCluster{
		ObjectMeta: v1.ObjectMeta{
			Name:      clusterId,
			Namespace: ns,
		},
	}
	r.Lock()
	defer r.Unlock()

	logf.IntoContext(ctx, logger)
	if err := r.reconcileCluster(ctx, cls); err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *UmoClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&umov1.UmoCluster{}).
		Owns(&corev1.Pod{}, builder.WithPredicates(filterPodByLabel(LabelMiddlewareType, r.config.MiddlewareType))).
		WithEventFilter(filterEvent()).
		Named("umocluster").
		Complete(r)
}

// LoopCallReconcile 定时循环去调用协调函数
func (r *UmoClusterReconciler) LoopCallReconcile() {
	sleep := time.Second * 600
	for {
		clusters := []*umov1.UmoCluster{}
		r.clusterIdToNamespace.Range(func(key, value any) bool {
			cls := &umov1.UmoCluster{
				ObjectMeta: v1.ObjectMeta{
					Name:      key.(string),
					Namespace: value.(string),
				},
			}
			if err := r.Client.Get(context.Background(), types.NamespacedName{
				Namespace: cls.GetNamespace(),
				Name:      cls.GetName(),
			}, cls); err != nil {
				if k8sErrors.IsNotFound(err) {
					r.clusterIdToNamespace.Delete(key)
				} else {
					slog.Error("loop call reconcile client get error", "error", err)
				}
			} else {
				clusters = append(clusters, cls)
			}
			return true
		})
		if len(clusters) == 0 {
			time.Sleep(sleep)
			continue
		}
		sort.Slice(clusters, func(i, j int) bool {
			return clusters[i].Name < clusters[j].Name
		})
		interval := sleep / time.Duration((2 * len(clusters)))
		for _, cls := range clusters {
			r.Lock()
			logger := logf.Log.WithValues("namespace", cls.GetNamespace(), "cluster_id", cls.Name)
			ctx := context.Background()
			logf.IntoContext(ctx, logger)
			if err := r.reconcileCluster(ctx, cls); err != nil {
				logger.Error(err, "loop call reconcile cluster reconcile error")
			}
			r.Unlock()
			time.Sleep(interval)
		}
		time.Sleep(sleep / 2)
	}
}

func (r *UmoClusterReconciler) reconcileCluster(ctx context.Context, cls *umov1.UmoCluster) error {
	return r.ClusterReconciler.Reconcile(ctx, cls)
}

func filterPodByLabel(key, value string) predicate.Predicate {
	return predicate.NewPredicateFuncs(func(object client.Object) bool {
		if object == nil {
			return false
		}
		pod, ok := object.(*corev1.Pod)
		if !ok {
			return false
		}
		return pod.Labels[key] == value
	})
}

func filterEvent() predicate.Predicate {
	return predicate.Funcs{
		UpdateFunc: func(e event.TypedUpdateEvent[client.Object]) bool {
			switch e.ObjectOld.(type) {
			case *corev1.Pod:
				allow := e.ObjectOld.GetResourceVersion() != e.ObjectNew.GetResourceVersion()
				if time.Since(e.ObjectNew.GetCreationTimestamp().Time) < time.Minute {
					allow = false
				}
				return allow
			}
			return true
		},
		CreateFunc: func(e event.TypedCreateEvent[client.Object]) bool {
			switch e.Object.(type) {
			case *corev1.Pod:
				return false
			}
			return true
		},
		DeleteFunc: func(e event.TypedDeleteEvent[client.Object]) bool {
			switch e.Object.(type) {
			case *corev1.Pod:
				return !e.DeleteStateUnknown
			}
			return true
		},
		GenericFunc: func(e event.TypedGenericEvent[client.Object]) bool {
			switch e.Object.(type) {
			case *corev1.Pod:
				return false
			}
			return true
		},
	}
}
