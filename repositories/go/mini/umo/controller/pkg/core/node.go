package core

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"strings"
	"sync"
	"time"

	umov1 "nofacepeace.github.io/controller/api/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	apiClient "nofacepeace.github.io/controller/pkg/client"
	"nofacepeace.github.io/controller/pkg/config"
	"nofacepeace.github.io/controller/pkg/model"
)

const (
	LabelClusterDeleteStage       = "delete-stage"
	ClusterDeleteStageNodeOffline = "node-offline"

	// feature
	FeatureDeleteCluster = "cluster_delete"
)

type NodeManager struct {
	fm     *FeatureManager
	pom    *PodManager
	tm     *TplManager
	Schema *runtime.Scheme
	em     *EventManager
	pvm    *PvcManager
	stm    *StatusManager
}

type nodeGroup struct {
	nodes          []*node
	updateStrategy *umov1.UpdateStrategy
}

type node struct {
	spec umov1.NodeSetSpec
	pod  *corev1.Pod
	name string
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
				if err := n.checkNode(ctx, cls, node, group.updateStrategy); err != nil {
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
	logger := logf.FromContext(ctx)
	if len(errs) > 0 {
		logger.Error(errorsToError(errs), "node manager check node groups error")
	}
	return nil
}

func (n *NodeManager) checkNode(ctx context.Context, cls *umov1.Middleware, node *node, strategy *umov1.UpdateStrategy, args ...any) error {
	if apiClient.IsClusterPublishAbort(cls.GetName(), cls.Spec.PublishId) {
		return nil
	}
	if err := n.reconcileNode(ctx, cls, node, strategy); err != nil {
		return fmt.Errorf("node manager reconcile error: [%w]", err)
	}
	if n.SkipSleep(ctx) {
		return nil
	}
	time.Sleep(strategy.PodUpdateInterval())
	return nil
}

func (n *NodeManager) reconcileNode(ctx context.Context, cls *umov1.Middleware, node *node, strategy *umov1.UpdateStrategy, args ...any) error {
	logger := logf.FromContext(ctx)
	logger = logger.WithValues("node", node.name)
	if err := n.pvm.checkPvc(); err != nil {
		return fmt.Errorf("pvc manager check pvc error: [%w]", err)
	}
	var err error
	if node.pod == nil {
		if err = n.createNode(ctx, cls, node, strategy); err != nil {
			err = fmt.Errorf("node manager create node error: [%w]", err)
		}
	} else {
		if err = n.updateNode(); err != nil {
			err = fmt.Errorf("node manager update node error: [%w]", err)
		}
	}
	if err == nil {
		err = n.stm.UpdateClusterStatus()
		// 非 endpoint 节点，忽略报错
		if !node.spec.IsEndpoint {
			logger.Error(err, "node manager update cluster status error")
			err = nil
		}
	}
	switch strategy.OnFailure {
	case umov1.OnFailureActionTerminate:
		logger.Error(err, "node manager reconcile node terminate failure", "error", err, "node", node.name)
		return err
	case umov1.OnFailureActionIgnore:
		logger.Info("node manager reconcile node ignore failure", "error", err, "node", node.name)
		return nil
	default:
		return err
	}
}

func (n *NodeManager) createNode(ctx context.Context, cls *umov1.Middleware, node *node, strategy *umov1.UpdateStrategy) error {
	vars := n.tm.createTplVar()
	pod, err := n.generatePod(cls, node, vars)
	if err != nil {
		return fmt.Errorf("node manager generate pod error: [%w]", err)
	}
	if !strategy.SkipChecker {
		if err := n.preCheck(); err != nil {
			return fmt.Errorf("node manager pre check error: [%w]", err)
		}
	}
	if err := ctrl.SetControllerReference(cls, pod, n.Schema); err != nil {
		return fmt.Errorf("controller runtime set controller reference error: [%w]", err)
	}
	if err := n.pom.CreatePod(ctx, cls, pod); err != nil {
		return fmt.Errorf("pod manager create pod error: [%w]", err)
	}
	n.em.Dispatch()
	if err := n.postCheck(); err != nil {
		return fmt.Errorf("node manager post check error: [%w]", err)
	}
	return nil
}

func (n *NodeManager) generatePod(cls *umov1.Middleware, node *node, vars *TplVar) (*corev1.Pod, error) {
	tplName := generateTplName()
	pod, err := n.tm.generatePod(tplName, vars)
	if err != nil {
		return nil, fmt.Errorf("template manager generate pod error: [%w]", err)
	}

	pod.SetName(node.name)
	pod.SetNamespace(cls.GetNamespace())
	pod.Spec.Hostname = node.name
	pod.Spec.Subdomain = cls.GetName()

	for i, v := range pod.Spec.Containers {
		if v.Name == model.ContainerNameMain {
			pod.Spec.Containers[i].Resources = node.spec.Resources.ResourceRequirements
			break
		}
	}
	if pod.Labels == nil {
		pod.Labels = map[string]string{}
	}

	// copy pod labels
	for k, v := range cls.Labels {
		if strings.HasPrefix(k, "pod.") {
			pod.Labels[k[4:]] = v
		}
	}
	maps.Copy(pod.Labels, node.spec.Labels)
	// set pod labels
	pod.Labels[model.LabelClusterId] = cls.GetName()
	pod.Labels[model.LabelEksId] = config.Get().Eks.Id
	pod.Labels[model.LabelMiddlewareType] = cls.Spec.MiddlewareType
	pod.Labels[model.LabelManagedBy] = config.Get().ControllerName
	pod.Labels[model.LabelNodeSetName] = node.spec.Name
	pod.Labels[model.LabelIsEndpoint] = strconv.FormatBool(node.spec.IsEndpoint)
	pod.Labels[model.LabelPodName] = pod.GetName()
	pod.Labels[model.LabelNodeName] = pod.Spec.NodeName
	pod.Labels[model.LabelClusterName] = cls.GetName()
	pod.Labels[model.LabelAvailabilityZone] = config.Get().Eks.Az
	pod.Labels[model.LabelRegionZone] = config.Get().Eks.Rz
	pod.Labels[model.LabelNodeSetDomain] = node.spec.Domain

	// set pod annotations
	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}
	for k, v := range cls.Annotations {
		if strings.HasPrefix(k, "pod.") {
			pod.Annotations[k[4:]] = v
		}
	}
	maps.Copy(pod.Annotations, node.spec.Annotations)
	pod.Annotations[model.AnnotationTplVersion] = node.spec.TplVersion
	pod.Annotations[model.AnnotationPublishId] = cls.Spec.PublishId
	pod.Annotations[model.AnnotationNodeGeneration] = strconv.Itoa(vars.NodeGeneration)

	return pod, nil
}

func (n *NodeManager) updateNode(args ...any) error {

	return nil
}

func (n *NodeManager) SkipSleep(ctx context.Context) bool {
	return false
}
func (n *NodeManager) ScaleDown(ctx context.Context, cls *umov1.Middleware, pods []corev1.Pod) error {
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

func (n *NodeManager) preCheck(args ...any) error {
	return nil
}

func (n *NodeManager) postCheck(args ...any) error {
	return nil
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

func generateTplName(args ...any) string {
	return ""
}
