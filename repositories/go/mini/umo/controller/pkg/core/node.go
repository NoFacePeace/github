package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	umov1 "nofacepeace.github.io/controller/api/v1"

	corev1 "k8s.io/api/core/v1"
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
	updateStrategy := n.getUpdateStrategy(cls)
	for _, nodeset := range cls.Spec.Normal {
		var wg sync.WaitGroup
		batchCount := &atomic.Int32{}
		errCh := make(chan error, updateStrategy.Concurrency)
		for eks, nodeCount := range nodeset.NodeCounts {
			if eks != GetConfig().Eks.Id {
				continue
			}
			for idx := nodeCount.Offset; idx < nodeCount.Offset+nodeCount.Count; idx++ {
				batchFull := false
				if idx == nodeCount.Offset+nodeCount.Count-1 {
					batchFull = true
				}
				nodeName := genNodeName(cls.GetName(), nodeset.Name, idx)
				pod, _, err := n.pom.GetPodIfExists(ctx, cls.GetNamespace(), nodeName)
				if err != nil {
					return fmt.Errorf("pod manager get pod if exists error: [%w]", err)
				}
				wg.Add(1)
				go func() {
					n.asyncCheckNode(ctx, pod)
					wg.Done()
				}()
				if batchCount.Add(1) >= int32(updateStrategy.Concurrency) {
					batchFull = true
				}
				if batchFull {
					wg.Wait()
					close(errCh)
					arr := []error{}
					for err := range errCh {
						arr = append(arr, err)
					}
					errCh = make(chan error, updateStrategy.Concurrency)
					err := errors.Join(arr...)
					if err != nil {
						return err
					}
				}
				batchCount.Store(0)
			}
		}
	}
	return nil
}

func (n *NodeManager) asyncCheckNode(ctx context.Context, pod *corev1.Pod) {

}

func (n *NodeManager) checkNode() {

}

func (n *NodeManager) reconcile() {

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
