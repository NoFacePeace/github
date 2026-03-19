package core

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	umov1 "nofacepeace.github.io/controller/api/v1"
	"nofacepeace.github.io/controller/pkg/config"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

type PvcManager struct {
	client.Client
}

func (p *PvcManager) EnsurePvc(ctx context.Context, cls *umov1.Middleware, nodeSetSpec *umov1.NodeSetSpec, pod *corev1.Pod) error {
	logger := logf.FromContext(ctx)
	ns := cls.GetNamespace()
	for _, vct := range nodeSetSpec.VolumeClaimTemplates {
		pvcName := p.genPvcName(pod.Name, vct.Metadata.Name)
		pvc := &corev1.PersistentVolumeClaim{}
		err := p.Client.Get(ctx, types.NamespacedName{Namespace: ns, Name: pvcName}, pvc)
		if err == nil {
			logger.Info("pvc already exists", "pvc", pvcName)
			p.resizePvcIfNeeded(ctx, cls.GetName(), &vct, pvc)
			continue
		}
		if !k8sErrors.IsNotFound(err) {
			return fmt.Errorf("k8s client get pvc %s error: [%w]", pvcName, err)
		}
		if err := p.createPvc(ctx, cls); err != nil {
			return fmt.Errorf("pvc manager create pvc %s error: [%w]", pvcName, err)
		}
	}
	return nil
}

func (p *PvcManager) genPvcName(podName, tplName string) string {
	return fmt.Sprintf("%s-%s", podName, tplName)
}

func (p *PvcManager) resizePvcIfNeeded(ctx context.Context, cls string, vct *umov1.VolumeClaimTemplate, pvc *corev1.PersistentVolumeClaim) error {
	logger := logf.FromContext(ctx)
	want := vct.Spec.Resources.Requests[corev1.ResourceStorage]
	have := pvc.Spec.Resources.Requests[corev1.ResourceStorage]
	if have.Equal(want) {
		return nil
	}
	dryRun := config.InDryRunMode(cls)
	if dryRun {
		logger.Info("[dry run] pvc resized", "cls", cls, "pvc", pvc.Name, "want", want, "have", have)
		return nil
	}
	pvc.Spec.Resources.Requests[corev1.ResourceStorage] = want
	if err := p.Client.Update(ctx, pvc); err != nil {
		return fmt.Errorf("k8s client update pvc %s error: [%w]", pvc.Name, err)
	}
	logger.Info("pvc resized", "pvc", pvc.Name, "want", want, "have", have)
	return nil
}

func (p *PvcManager) createPvc(ctx context.Context, cls *umov1.Middleware) error {
	pvc := &corev1.PersistentVolumeClaim{}
	if err := ctrl.SetControllerReference(cls, pvc, p.Scheme()); err != nil {
		return fmt.Errorf("controller runtime set controller reference error: [%w]", err)
	}
	if err := p.Client.Create(ctx, pvc); err != nil {
		return fmt.Errorf("k8s client create pvc %s error: [%w]", pvc.Name, err)
	}
	return nil
}

func (p *PvcManager) getPvcType(vct *umov1.VolumeClaimTemplate) string {
	return ""
}
