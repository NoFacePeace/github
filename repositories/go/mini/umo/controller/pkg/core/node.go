package core

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	umov1 "nofacepeace.github.io/controller/api/v1"

	corev1 "k8s.io/api/core/v1"
	apiClient "nofacepeace.github.io/controller/pkg/client"
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

type nodeGroup struct {
	nodes          []*node
	updateStrategy umov1.UpdateStrategy
}

type node struct {
	spec umov1.NodeSetSpec
	pod  *corev1.Pod
}

func (n *NodeManager) checkNodes(ctx context.Context, cls *umov1.Middleware) error {
	if n.isClusterOffline(cls) {
		return nil
	}
	nodeGroupMap := map[string]*nodeGroup{}
	for _, nodeset := range cls.Spec.Normal {
		for eks, cnts := range nodeset.NodeCounts {
			if eks != config.Get().Eks.Id {
				continue
			}
			for idx := cnts.Offset; idx < cnts.Offset+cnts.Count; idx++ {
				nodeName := generateNodeName(cls.GetName(), nodeset.Name, idx)
				pod, _, err := n.pom.GetPodIfExists(ctx, cls.GetNamespace(), nodeName)
				if err != nil {
					return fmt.Errorf("pod manager get pod if exists error: [%w]", err)
				}
				spec, filter := GetFinalNodeSet(ctx, cls, pod, idx)
				groupName := generateGroupName(filter)
				group := nodeGroupMap[groupName]
				if group == nil {
					group = &nodeGroup{}
					nodeGroupMap[groupName] = group
				}
				group.nodes = append(group.nodes, &node{
					spec: spec,
				})
			}
		}
	}
	groups := []*nodeGroup{}
	for _, group := range nodeGroupMap {
		groups = append(groups, group)
	}
	return n.checkNodeGroups(ctx, cls, groups)
}

// checkNodeGroups 检查节点组
func (n *NodeManager) checkNodeGroups(ctx context.Context, cls *umov1.Middleware, groups []*nodeGroup) error {
	wg := sync.WaitGroup{}
	cnt := 0
	errs := []error{}
	mu := sync.Mutex{}
	for _, group := range groups {
		for i, node := range group.nodes {
			batchFull := false
			if i == len(group.nodes)-1 {
				batchFull = true
			}
			wg.Add(1)
			cnt++
			if cnt == group.updateStrategy.Concurrency {
				batchFull = true
			}
			go func() {
				defer wg.Done()
				if err := n.checkNode(ctx, cls, group.updateStrategy, node); err != nil {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
				}
			}()
			if !batchFull {
				continue
			}
			wg.Wait()
			cnt = 0
			if len(errs) > config.Get().ReconcilePolicy.NodeCheckErrorTolerance {
				return fmt.Errorf("node manager check node groups error: [%w]", errorsToError(errs))
			}
		}
	}

	return nil
}

func (n *NodeManager) checkNode(ctx context.Context, cls *umov1.Middleware, strategy umov1.UpdateStrategy, args ...any) error {
	if apiClient.IsClusterPublishAbort(cls.GetName(), cls.Spec.PublishId) {
		return nil
	}
	if err := n.reconcile(); err != nil {
		return fmt.Errorf("node manager reconcile error: [%w]", err)
	}
	if n.SkipSleep(ctx) {
		return nil
	}
	time.Sleep(strategy.PodUpdateInterval())
	return nil
}

func (n *NodeManager) SkipSleep(ctx context.Context) bool {
	return false
}
func (n *NodeManager) ScaleDown(ctx context.Context, cls *umov1.Middleware, pods []corev1.Pod) error {
	return nil
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

func generateNodeName(cls, nodeset string, idx int) string {
	return fmt.Sprintf("%s-%s-%d", cls, nodeset, idx)
}

func generateGroupName(filter umov1.GrayFilter) string {
	return ""
}

func errorsToError(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	var ret error
	for _, err := range errs {
		ret = errors.Join(ret, err)
	}
	return ret
}
