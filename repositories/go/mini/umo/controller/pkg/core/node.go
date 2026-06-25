package core

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"sort"
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
	"nofacepeace.github.io/controller/pkg/extensions/checker"
	"nofacepeace.github.io/controller/pkg/metrics"
	"nofacepeace.github.io/controller/pkg/model"
)

const (
	LabelClusterDeleteStage       = "delete-stage"
	ClusterDeleteStageNodeOffline = "node-offline"
)

type NodeAction int

const (
	NodeActionNone NodeAction = iota
	NodeActionPatchOnly
	NodeActionInPlaceUpdate
	NodeActionRecreate
	NodeActionMigrate
)

type NodeManager struct {
	fm             *FeatureManager
	pom            *PodManager
	tm             *TplManager
	Schema         *runtime.Scheme
	em             *EventManager
	pvm            *PvcManager
	stm            *StatusManager
	preCheckers    []checker.Checker
	postCheckers   []checker.Checker
	inPostCheckers []checker.Checker
	op             *OperationManager
	sm             *ServiceManager
}

func (n *NodeManager) checkNodes(ctx context.Context, cls *umov1.Middleware) error {
	if n.isClusterOffline(cls) {
		return nil
	}

	nodeGroupMap := map[string]*nodeGroup{}
	for _, nodeset := range cls.Spec.Normal {
		for eks, cnts := range nodeset.NodeCounts {
			// 只检查当前 EKS 的节点
			if eks != config.Get().Eks.Id {
				continue
			}

			// 遍历节点集中的所有节点
			for idx := cnts.Offset; idx < cnts.Offset+cnts.Count; idx++ {
				nodeName := generateNodeName(cls.GetName(), nodeset.Name, idx)
				pod, _, err := n.pom.GetPodIfExists(ctx, cls.GetNamespace(), nodeName)
				if err != nil {
					return fmt.Errorf("pod manager get pod if exists error: [%w]", err)
				}

				spec, filter := getFinalNodeSet(cls, &nodeset, pod, idx)
				groupName := generateGroupName(filter)
				group := nodeGroupMap[groupName]
				if group == nil {
					group = &nodeGroup{}
					group.updateStrategy = n.getGrayFilterUpdateStrategy(cls, filter)
					nodeGroupMap[groupName] = group
				}
				group.nodes = append(group.nodes, &node{
					spec: spec,
					name: nodeName,
					pod:  pod,
					idx:  idx,
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

// isClusterOffline 检查集群是否离线
func (n *NodeManager) isClusterOffline(cls *umov1.Middleware) bool {
	labels := cls.GetLabels()
	if labels == nil {
		return false
	}
	return labels[LabelClusterDeleteStage] == ClusterDeleteStageNodeOffline && n.fm.IsEnabled(model.FeatureDeleteCluster)
}

func (n *NodeManager) getGrayFilterUpdateStrategy(cls *umov1.Middleware, filter *umov1.GrayFilter) *umov1.UpdateStrategy {
	s := filter.UpdateStrategy
	def := n.getClusterUpdateStrategy(cls)
	return n.fixUpdateStrategy(&s, def)
}

func (n *NodeManager) getClusterUpdateStrategy(cls *umov1.Middleware) *umov1.UpdateStrategy {
	s := cls.Spec.UpdateStrategy
	mid := n.getMiddlewareTypeUpdateStrategy(cls)
	return n.fixUpdateStrategy(&s, mid)
}

func (n *NodeManager) getMiddlewareTypeUpdateStrategy(cls *umov1.Middleware) *umov1.UpdateStrategy {
	s, ok := config.Get().ReconcilePolicy.UpdateStrategys[cls.Spec.MiddlewareType]
	def := n.getDefaultUpdateStrategy()
	if !ok {
		return def
	}
	return n.fixUpdateStrategy(&s, def)
}

func (n *NodeManager) fixUpdateStrategy(a *umov1.UpdateStrategy, b *umov1.UpdateStrategy) *umov1.UpdateStrategy {
	if a == nil {
		return b
	}
	if a.Concurrency <= 0 {
		a.Concurrency = b.Concurrency
	}
	if a.PodUpdateIntervalMs <= 0 {
		a.PodUpdateIntervalMs = b.PodUpdateIntervalMs
	}
	if a.PodExecTimeoutMs <= 0 {
		a.PodExecTimeoutMs = b.PodExecTimeoutMs
	}
	if a.SkipChecker {
		a.SkipChecker = b.SkipChecker
	}
	if a.OnFailure == "" {
		a.OnFailure = b.OnFailure
	}
	return a
}

func (n *NodeManager) getDefaultUpdateStrategy() *umov1.UpdateStrategy {
	return config.DefaultUpdateStrategy
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

func (n *NodeManager) checkNode(ctx context.Context, cls *umov1.Middleware, node *node, strategy *umov1.UpdateStrategy) error {
	if apiClient.IsClusterPublishAbort(cls.GetName(), cls.Spec.PublishId) {
		return nil
	}
	if err := n.reconcileNode(ctx, cls, node, strategy); err != nil {
		return fmt.Errorf("node manager reconcile error: [%w]", err)
	}
	n.op.NewOperationBuilder(ctx, cls.GetName(), "checkNode").Report()
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
		if err = n.updateNode(ctx, cls, node, strategy); err != nil {
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

func (n *NodeManager) SkipSleep(ctx context.Context) bool {
	value, ok := ctx.Value(model.CtxKeySkipSleep).(bool)
	if !ok {
		return false
	}
	return value
}

func (n *NodeManager) createNode(ctx context.Context, cls *umov1.Middleware, node *node, strategy *umov1.UpdateStrategy) error {
	vars := n.tm.createTplVar()
	pod, err := n.generatePod(cls, node, vars)
	if err != nil {
		return fmt.Errorf("node manager generate pod error: [%w]", err)
	}
	if !strategy.SkipChecker {
		if err := n.preCheck(ctx, cls, node); err != nil {
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
	if err := n.postCheck(ctx, cls, node, strategy, model.PostCeckActionCreateNode); err != nil {
		return fmt.Errorf("node manager post check error: [%w]", err)
	}
	return nil
}

// generatePod 生成 pod
func (n *NodeManager) generatePod(cls *umov1.Middleware, node *node, vars *TplVar) (*corev1.Pod, error) {
	// 生成模板名称
	tplName := generateTplName(cls.Spec.MiddlewareType, node.spec.TplVersion)

	// 生成 pod
	pod, err := n.tm.generatePod(tplName, vars)
	if err != nil {
		return nil, fmt.Errorf("template manager generate pod error: [%w]", err)
	}

	// 设置 pod 名称
	pod.SetName(node.name)
	// 设置 pod 命名空间
	pod.SetNamespace(cls.GetNamespace())
	// 设置 pod 主机名
	pod.Spec.Hostname = node.name
	// 设置 pod 子域名
	pod.Spec.Subdomain = cls.GetName()

	// 设置 pod 主容器资源
	for i, v := range pod.Spec.Containers {
		if v.Name == model.ContainerNameMain {
			pod.Spec.Containers[i].Resources = node.spec.Resources.ResourceRequirements
			break
		}
	}

	// 设置 pod 标签
	n.setPodLabels(cls, node, pod)
	// 设置 pod 注解
	if err := n.setPodAnnotations(cls, node, pod, vars); err != nil {
		return nil, fmt.Errorf("node manager set pod annotations error: [%w]", err)
	}

	// 绑定 pod pvc
	n.bindPodPvc(node, pod)
	// 添加节点亲和性
	n.addNodeAffinityForUnhealthyNodes(pod)
	// 设置 pod 配置
	if err := n.setPodConfig(node, pod, vars); err != nil {
		return nil, fmt.Errorf("node manager set pod config error: [%w]", err)
	}
	return pod, nil
}

func (n *NodeManager) setPodLabels(cls *umov1.Middleware, node *node, pod *corev1.Pod) {
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
}

func (n *NodeManager) setPodAnnotations(cls *umov1.Middleware, node *node, pod *corev1.Pod, vars *TplVar, args ...any) error {
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

	cns := []string{}
	for _, container := range pod.Spec.Containers {
		cns = append(cns, container.Name)
	}
	sort.Strings(cns)
	pod.Annotations[model.AnnotationContainers] = strings.Join(cns, ",")

	if n.fm.IsEnabled(model.FeatureInplaceVpa) {
		m := map[string]corev1.ResourceRequirements{}
		m[model.ContainerNameMain] = node.spec.Resources.ResourceRequirements
		raw, err := json.Marshal(m)
		if err != nil {
			return fmt.Errorf("json marshal error: [%w]", err)
		}
		pod.Annotations[model.AnnotationInplaceVpa] = string(raw)
	}
	return nil
}

func (n *NodeManager) bindPodPvc(node *node, pod *corev1.Pod) {
	for _, vct := range node.spec.VolumeClaimTemplates {
		name := n.pvm.genPvcName(pod.Name, vct.Metadata.Name)
		volume := corev1.Volume{
			Name: name,
			VolumeSource: corev1.VolumeSource{
				PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
					ClaimName: name,
					ReadOnly:  false,
				},
			},
		}
		pod.Spec.Volumes = append(pod.Spec.Volumes, volume)
		for i := range pod.Spec.Containers {
			if pod.Spec.Containers[i].Name == model.ContainerNameMain {
				pod.Spec.Containers[i].VolumeMounts = append(pod.Spec.Containers[i].VolumeMounts, corev1.VolumeMount{
					Name:      name,
					MountPath: vct.MountPath,
					ReadOnly:  false,
				})
				break
			}
		}
	}
}

// addNodeAffinityForUnhealthyNodes 添加节点亲和性，避免使用不健康的节点
func (n *NodeManager) addNodeAffinityForUnhealthyNodes(pod *corev1.Pod) {
	if pod.Spec.Affinity == nil {
		pod.Spec.Affinity = &corev1.Affinity{}
	}
	if pod.Spec.Affinity.NodeAffinity == nil {
		pod.Spec.Affinity.NodeAffinity = &corev1.NodeAffinity{}
	}
	if pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = &corev1.NodeSelector{}
	}

	requirement := corev1.NodeSelectorRequirement{
		Key:      model.LabelNodeHealthStatus,
		Operator: corev1.NodeSelectorOpNotIn,
		Values:   []string{model.LabelNodeHealthStatusValueUnhealthy},
	}
	if len(pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms) == 0 {
		pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = []corev1.NodeSelectorTerm{
			{MatchExpressions: []corev1.NodeSelectorRequirement{requirement}},
		}
	} else {
		for i := range pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
			pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i].MatchExpressions = append(
				pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms[i].MatchExpressions,
				requirement,
			)
		}
	}
}

func (n *NodeManager) setPodConfig(node *node, pod *corev1.Pod, vars *TplVar, args ...any) error {
	items := []corev1.DownwardAPIVolumeFile{}
	// files
	for k, v := range node.spec.Config.Files {
		value, err := n.tm.parseTpl(v, vars)
		if err != nil {
			return fmt.Errorf("template manager parse tpl error: [%w]", err)
		}
		pod.Annotations[k] = value
		items = append(items, corev1.DownwardAPIVolumeFile{
			Path: k,
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: fmt.Sprintf("metadata.annotations['%s']", k),
			},
		})
	}
	// version
	pod.Annotations[model.AnnotationTplVersion] = generateVersion()
	items = append(items, corev1.DownwardAPIVolumeFile{
		Path: model.AnnotationTplVersion,
		FieldRef: &corev1.ObjectFieldSelector{
			FieldPath: fmt.Sprintf("metadata.annotations['%s']", model.AnnotationTplVersion),
		},
	})
	volum := corev1.Volume{
		Name: model.VolumeNameConfigFiles,
		VolumeSource: corev1.VolumeSource{
			DownwardAPI: &corev1.DownwardAPIVolumeSource{
				Items: items,
			},
		},
	}
	pod.Spec.Volumes = append(pod.Spec.Volumes, volum)

	for i, container := range pod.Spec.Containers {
		mounted := false
		for _, volume := range container.VolumeMounts {
			if volume.Name == model.VolumeNameConfigFiles {
				mounted = true
				break
			}
		}
		if !mounted {
			pod.Spec.Containers[i].VolumeMounts = append(pod.Spec.Containers[i].VolumeMounts, corev1.VolumeMount{
				Name:      model.VolumeNameConfigFiles,
				MountPath: model.VolumeConfigFilesMountPath,
				ReadOnly:  false,
			})
		}
	}

	// env
	envs := []corev1.EnvVar{}
	for k, v := range node.spec.Config.Envs {
		value, err := n.tm.parseTpl(v, vars)
		if err != nil {
			return fmt.Errorf("template manager parse tpl error: [%w]", err)
		}
		pod.Annotations[k] = value
		envs = append(envs, corev1.EnvVar{
			Name: k,
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					FieldPath: fmt.Sprintf("metadata.annotations['%s']", k),
				},
			},
		})
	}
	for i := range pod.Spec.Containers {
		pod.Spec.Containers[i].Env = append(pod.Spec.Containers[i].Env, envs...)
	}

	for k, v := range node.spec.Config.Variables {
		value, err := n.tm.parseTpl(v, vars)
		if err != nil {
			return fmt.Errorf("template manager parse variables tpl error: [%w]", err)
		}
		pod.Annotations[k] = value
	}

	for k, v := range node.spec.Resources.Variables {
		value, err := n.tm.parseTpl(v, vars)
		if err != nil {
			return fmt.Errorf("template manager parse resources variables tpl error: [%w]", err)
		}
		pod.Annotations[k] = value
	}
	return nil
}

func (n *NodeManager) preCheck(ctx context.Context, cls *umov1.Middleware, node *node, args ...any) (err error) {
	if len(n.preCheckers) == 0 {
		return nil
	}
	defer func() {
		if err != nil {
			metrics.Inc(ctx, cls.Name)
			n.op.NewOperationBuilder(ctx, cls.GetName(), "preCheck").WithError(err).WithNodeName(node.name).Report()
		}
	}()
	n.op.NewOperationBuilder(ctx, cls.GetName(), OperationTypePreCheck).WithNodeName(node.name).Report()
	for _, checker := range n.preCheckers {
		res, msg, err := checker.Check(ctx, node.pod)
		n.op.NewOperationBuilder(ctx, cls.GetName(), checker.GetName()).WithNodeName(node.name).WithError(err).Report()
		if err != nil {
			return fmt.Errorf("checker %s check error: [%w]", checker.GetName(), err)
		}
		if res != model.CheckerResultOk {
			return fmt.Errorf("checker %s check result not ok: [%s]", checker.GetName(), msg)
		}
	}
	return nil
}

func (n *NodeManager) postCheck(ctx context.Context, cls *umov1.Middleware, node *node, strategy *umov1.UpdateStrategy, action string, args ...any) (err error) {
	logger := logf.FromContext(ctx)
	defer func() {
		failed := err != nil
		n.op.NewOperationBuilder(context.Background(), cls.GetName(), OperationTypePostCheck).WithNodeName(node.name).WithError(err).Report()
		if err := n.pom.updatePodAnnotation(ctx, node.pod, map[string]string{
			model.AnnotationPostCheckFailed: strconv.FormatBool(failed),
			model.AnnotationPostCheckAction: action,
		}); err != nil {
			logger.Error(err, "pod manager update pod annotation error")
		}
	}()
	// 如果跳过检查，直接返回成功
	if strategy.SkipChecker {
		return nil
	}
	// 如果是演练模式，直接返回成功
	if config.InDryRunMode(cls.Name) {
		return nil
	}
	checkers := n.inPostCheckers
	checkers = append(checkers, n.postCheckers...)
	for _, checker := range checkers {
		res, msg, retErr := checker.Check(ctx, node.pod)
		if retErr != nil {
			err = fmt.Errorf("checker %s check error: [%w]", checker.GetName(), retErr)
			break
		}
		if res != model.CheckerResultOk {
			err = fmt.Errorf("checker %s check result not ok: [%s]", checker.GetName(), msg)
			break
		}
	}
	return nil
}

func (n *NodeManager) updateNode(ctx context.Context, cls *umov1.Middleware, node *node, strategy *umov1.UpdateStrategy, args ...any) error {
	pod := node.pod
	if isTrue(pod.Labels[model.LabelIsEndpoint]) {
		// create node maybe timeout
		if err := n.sm.UpdateEndpoint(); err != nil {
			return fmt.Errorf("service manager update endpoint error: [%w]", err)
		}
	}
	if isTrue(pod.Labels[model.LabelManualManagement]) {
		return nil
	}
	if err := n.checkPostCheck(ctx, cls, node, strategy); err != nil {
		return fmt.Errorf("node manager check post check error: [%w]", err)
	}
	return n.reconcileNodeSpec(ctx, cls, node, strategy)
}

func (n *NodeManager) checkPostCheck(ctx context.Context, cls *umov1.Middleware, node *node, strategy *umov1.UpdateStrategy) error {
	pod := node.pod
	if !isTrue(pod.Annotations[model.AnnotationPostCheckFailed]) || pod.Annotations[model.AnnotationPublishId] != cls.Spec.PublishId {
		return nil
	}
	action := pod.Annotations[model.AnnotationPostCheckAction]
	return n.postCheck(ctx, cls, node, strategy, action)
}

func (n *NodeManager) reconcileNodeSpec(ctx context.Context, cls *umov1.Middleware, node *node, strategy *umov1.UpdateStrategy, args ...any) error {
	vars := n.tm.createTplVar()
	parsePod, err := n.generatePod(cls, node, vars)
	if err != nil {
		return fmt.Errorf("node manager generate pod error: [%w]", err)
	}
	newPod := node.pod.DeepCopy()
	action, err := n.getNodeAction(node, newPod, parsePod, strategy)
	if err != nil {
		return fmt.Errorf("node manager check pod error: [%w]", err)
	}
	n.handlePod(action)
	n.postCheckForNodeChange()
	return nil
}

func (n *NodeManager) getNodeAction(node *node, newPod *corev1.Pod, parsePod *corev1.Pod, strategy *umov1.UpdateStrategy) (NodeAction, error) {
	// 如果策略中指定了更新方式，则直接使用策略中的更新方式
	if strategy.NodeAction == model.NodeActionInPlaceUpdate {
		return NodeActionInPlaceUpdate, nil
	}
	if strategy.NodeAction == model.NodeActionRecreate {
		return NodeActionRecreate, nil
	}

	if n.needMigrateNode(newPod) {
		return NodeActionMigrate, nil
	}
	if n.needRecreateNode(node, newPod, parsePod) {
		return NodeActionRecreate, nil
	}
	if n.needInPlaceUpdateNode(newPod) {
		return NodeActionInPlaceUpdate, nil
	}
	// if n.needSidecarInPlaceUpdateNode() {

	// }
	if n.needPatchOnlyNode(newPod) {
		return NodeActionPatchOnly, nil
	}
	return NodeActionNone, nil
}

func (n *NodeManager) needMigrateNode(pod *corev1.Pod, args ...any) bool {
	if status, ok := pod.Labels[model.LabelHealthStatus]; ok && status == model.LabelHealthStatusValueMigrate {
		return true
	}

	return false
}

func (n *NodeManager) needRecreateNode(node *node, newPod *corev1.Pod, parsePod *corev1.Pod) bool {
	// 如果健康状态是需要重建，说明节点不健康，需要重建
	if status, ok := newPod.Labels[model.LabelHealthStatus]; ok && status == model.LabelHealthStatusValueRecreate {
		return true
	}

	// 如果 PVC 发生了变化，说明存储发生了变化，需要重建节点
	if n.isPvcChanged(node) {
		return true
	}

	// 如果模板版本发生变化，说明模板发生了变化，需要重建节点
	if ver, ok := newPod.Annotations[model.AnnotationTplVersion]; ok && node.spec.TplVersion != ver {
		return true
	}

	// 如果 init container 发生了变化，说明初始化配置发生了变化，需要重建节点
	if n.isInitContainersChanged(newPod, parsePod) {
		return true
	}

	// 如果容器发生了变化，说明应用配置发生了变化，需要重建节点
	if n.isContainersChanged(newPod, parsePod) {
		return true
	}

	return false
}

func (n *NodeManager) needInPlaceUpdateNode(pod *corev1.Pod, args ...any) bool {
	if status, ok := pod.Labels[model.LabelHealthStatus]; ok && status == model.LabelHealthStatusValueInPlaceUpdate {
		pod.Labels[model.LabelHealthStatus] = ""
		return true
	}
	return false
}

func (n *NodeManager) isPvcChanged(node *node) bool {
	for _, vct := range node.spec.VolumeClaimTemplates {
		name := n.pvm.genPvcName(node.pod.Name, vct.Metadata.Name)
		exist := false
		for _, volume := range node.pod.Spec.Volumes {
			// 如果 volume 中没有对应的 pvc，说明 pvc 发生了变化
			if volume.Name == name {
				exist = true
				break
			}
		}
		if !exist {
			return true
		}
		for _, container := range node.pod.Spec.Containers {
			if container.Name != model.ContainerNameMain {
				continue
			}
			mounted := false
			for _, vm := range container.VolumeMounts {
				// 如果 volume mount 中没有对应的 pvc，说明 pvc 发生了变化
				if vm.Name == name && vm.MountPath == vct.MountPath {
					mounted = true
					break
				}
			}
			if !mounted {
				return true
			}
		}
	}
	return false
}

func (n *NodeManager) isInitContainersChanged(newPod *corev1.Pod, parsePod *corev1.Pod, args ...any) bool {
	m := map[string]*corev1.Container{}
	for i := range parsePod.Spec.InitContainers {
		name := parsePod.Spec.InitContainers[i].Name
		m[name] = &parsePod.Spec.InitContainers[i]
	}
	for _, container := range newPod.Spec.InitContainers {
		if parseContainer, ok := m[container.Name]; ok {
			if container.Image != parseContainer.Image {
				return true
			}
			if !reflect.DeepEqual(container.Command, parseContainer.Command) {
				return true
			}
			if !reflect.DeepEqual(container.Args, parseContainer.Args) {
				return true
			}
		}

	}
	return false
}

func (n *NodeManager) isContainersChanged(newPod *corev1.Pod, parsePod *corev1.Pod, args ...any) bool {
	if len(newPod.Spec.Containers) != len(parsePod.Spec.Containers) {
		return true
	}
	m := map[string]*corev1.Container{}
	for i := range newPod.Spec.Containers {
		name := newPod.Spec.Containers[i].Name
		m[name] = &newPod.Spec.Containers[i]
	}
	for _, container := range parsePod.Spec.Containers {
		old, ok := m[container.Name]
		if !ok {
			return true
		}
		

	}
	return false
}

func (n *NodeManager) isMainContainerImageChanged(newPod *corev1.Pod, parsePod *corev1.Pod) bool {
	m := map[string]*corev1.Container{}
	for i := range newPod.Spec.Containers {
		name := newPod.Spec.Containers[i].Name
		m[name] = &newPod.Spec.Containers[i]
	}
	for _, container := range parsePod.Spec.Containers {
		if container.Name != model.ContainerNameMain {
			continue
		}
		mainContainer := m[container.Name]
		if mainContainer.Image != container.Image {
			return true
		}
	}
	return false

}

func (n *NodeManager) isOtherContainersImageChanged(newPod *corev1.Pod, parsePod *corev1.Pod) bool {
	m := map[string]*corev1.Container{}
	for i := range newPod.Spec.Containers {
		name := newPod.Spec.Containers[i].Name
		m[name] = &newPod.Spec.Containers[i]
	}
	for _, container := range parsePod.Spec.Containers {
		if container.Name == model.ContainerNameMain {
			continue
		}
		if _, ok := m[container.Name]; !ok {
			continue
		}
		other := m[container.Name]
		if other.Image != container.Image {
			return true
		}
	}
	return false
}

func 

func (n *NodeManager) needPatchOnlyNode(args ...any) bool {
	return false
}

func (n *NodeManager) handlePod(args ...any) error {
	return nil
}

func (n *NodeManager) postCheckForNodeChange() error {
	return nil
}

func (n *NodeManager) ScaleDown(ctx context.Context, cls *umov1.Middleware, pods []corev1.Pod) error {
	return nil
}

func (n *NodeManager) getUpdateStrategy(cls *umov1.Middleware) umov1.UpdateStrategy {
	return umov1.UpdateStrategy{}
}

type nodeGroup struct {
	nodes          []*node
	updateStrategy *umov1.UpdateStrategy
}

type node struct {
	spec *umov1.NodeSetSpec
	pod  *corev1.Pod
	name string
	idx  int
}
