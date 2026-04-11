package core

import corev1 "k8s.io/api/core/v1"

type TplVar struct {
	NodeGeneration int
}

type TplManager struct {
}

func (t *TplManager) createTplVar(args ...any) *TplVar {
	return &TplVar{}
}

func (t *TplManager) generatePod(args ...any) (*corev1.Pod, error) {
	return nil, nil
}
