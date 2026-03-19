package core

import (
	"context"
	"fmt"

	umov1 "nofacepeace.github.io/controller/api/v1"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	apiclient "nofacepeace.github.io/controller/pkg/client"
	"nofacepeace.github.io/controller/pkg/config"
)

const (
	LabelClusterDeleteStage       = "delete-stage"
	ClusterDeleteStageNodeOffline = "node-offline"

	// feature
	FeatureDeleteCluster = "cluster_delete"
)

type NodeManager struct {
	fm  *FeatureManager
	pom *PodManager
}

func (n *NodeManager) checkNodes(ctx context.Context, cls *umov1.Middleware) error {
	if n.isClusterOffline(cls) {
		return nil
	}
	m := map[string]bool{}
	for _, nodeset := range cls.Spec.Normal {
		for eks, cnts := range nodeset.NodeCounts {
			if eks != config.Get().Eks.Id {
				continue
			}
			for idx := cnts.Offset; idx < cnts.Offset+cnts.Count; idx++ {
				nodeName := genNodeName(cls.GetName(), nodeset.Name, idx)
				pod, _, err := n.pom.GetPodIfExists(ctx, cls.GetNamespace(), nodeName)
				if err != nil {
					return fmt.Errorf("pod manager get pod if exists error: [%w]", err)
				}
				spec, filter := GetFinalNodeSet(ctx, cls, pod, idx)
				
			}
		}
	}
	return nil
}

func (n *NodeManager) ScaleDown(ctx context.Context, cls *umov1.Middleware, pods []corev1.Pod) error {
	return nil
}

func (n *NodeManager) asyncCheckNode(ctx context.Context, cls *umov1.Middleware, pod *corev1.Pod) {
	logger := logf.FromContext(ctx)
	if apiclient.IsClusterPublishAbort(cls.GetName(), cls.Spec.PublishId) {
		logger.Info("cluster publish abort", "cluster", cls.GetName(), "publish id", cls.Spec.PublishId)
		return
	}
	if err := n.reconcile(); err != nil {

	}
}

func (n *NodeManager) checkNode() {

}

func (n *NodeManager) reconcile() error {
	return nil
}

func (n *NodeManager) isClusterOffline(cls *umov1.Middleware) bool {
	labels := cls.GetLabels()
	if labels == nil {
		return false
	}
	return labels[LabelClusterDeleteStage] == ClusterDeleteStageNodeOffline && n.fm.IsEnabled(FeatureDeleteCluster)
}

func (n *NodeManager) getUpdateStrategy(cls *umov1.Middleware) umov1.UpdateStrategy {
	return umov1.UpdateStrategy{}
}

func genNodeName(cls, nodeset string, idx int) string {
	return fmt.Sprintf("%s-%s-%d", cls, nodeset, idx)
}
