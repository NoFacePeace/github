package core

import (
	"context"
	"fmt"
	"time"

	umov1 "localhost/umo/api/v1"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	forceDeleteGracePeriodSeconds = 0
	deleteGracePeriodSeconds      = 5
)

type PodManager struct {
	config *PodManagerConfig
	client.Client
}

type PodManagerConfig struct {
	PollInterval  time.Duration
	PollTimeout   time.Duration
	PollImmediate bool
}

func (p *PodManager) GetClusterPods(ctx context.Context, cls *umov1.Middleware) ([]corev1.Pod, error) {
	list := &corev1.PodList{}
	opts := []client.ListOption{
		client.InNamespace(cls.GetNamespace()),
		client.MatchingLabels(map[string]string{}),
	}
	if err := p.Client.List(ctx, list, opts...); err != nil {
		return nil, fmt.Errorf("k8s client list error: [%w]", err)
	}
	return list.Items, nil
}

func (p *PodManager) GetPod(ctx context.Context, ns, name string) (*corev1.Pod, error) {
	pod, err := p.getPod(ctx, ns, name)
	if err != nil {
		return nil, fmt.Errorf("pod manager get pod error: [%w]", err)
	}
	return pod, nil
}

func (p *PodManager) GetPodIfExists(ctx context.Context, ns, name string) (*corev1.Pod, bool, error) {
	pod, err := p.getPod(ctx, ns, name)
	if err == nil {
		return pod, true, nil
	}
	if k8sErrors.IsNotFound(err) {
		return pod, false, nil
	}
	return pod, false, fmt.Errorf("pod manager get pod error: [%w]", err)
}

func (p *PodManager) DeletePod(ctx context.Context, pod *corev1.Pod, force bool) error {
	if force {
		if err := p.Client.Delete(ctx, pod, client.GracePeriodSeconds(forceDeleteGracePeriodSeconds), client.PropagationPolicy(metav1.DeletePropagationForeground)); err != nil {
			return fmt.Errorf("k8s client force delete error: [%w]", err)
		}
		if err := p.deletePvc(); err != nil {
			return fmt.Errorf("pod manager delete pvc error: [%w]", err)
		}
	} else {
		if err := p.Client.Delete(ctx, pod, client.GracePeriodSeconds(deleteGracePeriodSeconds)); err != nil {
			return fmt.Errorf("k8s client delete error: [%w]", err)
		}
	}
	if err := p.isPodDeleted(ctx, pod); err != nil {
		return fmt.Errorf("pod manager is pod deleted error: [%w]", err)
	}
	return nil
}

func (p *PodManager) CreatePod(ctx context.Context, cls *umov1.Middleware, pod *corev1.Pod) error {
	if err := ctrl.SetControllerReference(cls, pod, p.Scheme()); err != nil {
		return fmt.Errorf("controller runtime set controller reference error: [%w]", err)
	}
	if err := p.Client.Create(ctx, pod); err != nil {
		return fmt.Errorf("k8s client create error: [%w]", err)
	}
	if err := p.isPodReady(pod.GetNamespace(), pod.GetName()); err != nil {
		p.setPodNotReady()
		return fmt.Errorf("pod manager is pod ready error: [%w]", err)
	}
	pod, err := p.getPod(ctx, pod.Namespace, pod.Name)
	if err != nil {
		return fmt.Errorf("pod manager get pod error: [%w]", err)
	}

	return nil
}

func (p *PodManager) isPodReady(ns, name string) error {
	return wait.PollUntilContextTimeout(context.Background(), p.config.PollInterval, p.config.PollTimeout, p.config.PollImmediate, func(ctx context.Context) (bool, error) {
		pod, err := p.getPod(ctx, ns, name)
		if err != nil {
			if k8sErrors.IsNotFound(err) {
				return false, nil
			}
			return false, fmt.Errorf("pod manager get pod error: [%w]", err)
		}
		if pod.Status.Phase != corev1.PodRunning {
			return false, nil
		}
		ready := 0
		for _, status := range pod.Status.ContainerStatuses {
			if status.Ready {
				ready++
			}
		}
		return ready == len(pod.Status.ContainerStatuses), nil
	})
}

func (p *PodManager) getPod(ctx context.Context, ns, name string) (*corev1.Pod, error) {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: ns,
		},
	}

	if err := p.Client.Get(ctx, client.ObjectKeyFromObject(pod), pod); err != nil {
		return nil, err
	}
	return pod, nil
}

func (p *PodManager) setPodNotReady() {

}

func (p *PodManager) updatePodAnnotation(ctx context.Context, pod *corev1.Pod) {
	// newPod := pod.DeepCopy()
	// newPod :=
}

func (p *PodManager) deletePvc() error {
	return nil
}

func (p *PodManager) isPodDeleted(ctx context.Context, pod *corev1.Pod) error {
	return wait.PollUntilContextTimeout(ctx, p.config.PollInterval, p.config.PollTimeout, p.config.PollImmediate, func(ctx context.Context) (bool, error) {
		_, err := p.getPod(ctx, pod.GetNamespace(), pod.Name)
		if err == nil {
			return false, nil
		}
		if k8sErrors.IsNotFound(err) {
			return true, nil
		}
		return false, fmt.Errorf("pod manager get pod error: [%w]", err)
	})
}
