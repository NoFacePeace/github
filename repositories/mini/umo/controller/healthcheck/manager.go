package healthcheck

import (
	"context"
	"errors"
	"log"
	"net/http"
	"slices"
	"sync"
	"time"

	"umo/controller/k8s"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Manager 定义健康检查管理器对外提供的接口。
type Manager interface {
	// Run 阻塞运行检查流程，每个管理器只调用一次，通过取消上下文通知退出。
	Run(ctx context.Context)
	// SubmitEvent 提交外部事件，成功仅表示入队，不表示处理完成。
	SubmitEvent(ctx context.Context, event Event) error
}

// Config 表示管理器的运行配置，字段随具体功能按需补充。
type Config struct {
	// K8sClient 用于读写 Kubernetes 资源。
	K8sClient client.Client
	// CheckInterval 表示检查间隔，为零时默认一分钟，不能为负数。
	CheckInterval time.Duration
	// Policies 定义异常状态的处理策略。
	Policies []Policy
}

// Policy 定义异常状态与处理动作的对应关系。
type Policy struct {
	Action Action
	Status []Status
	// ReconcileInterval 表示累计异常统计窗口及相同动作的去重冷却时间。
	ReconcileInterval time.Duration
	// Threshold 表示累计异常比例上限，取值范围为 0 到 1。
	Threshold float64
	// Webhooks 接收此策略的处理结果，为空时不发送通知。
	Webhooks []string
}

// manager 实现管理器接口，字段随初始化与运行逻辑按需补充。
type manager struct {
	k8sClient     client.Client
	webhookClient *http.Client
	checkers      []Checker
	config        Config
	events        chan Event
	// podActionMatchedAt 记录 Pod 名称、动作对应的最近策略命中时间。
	podActionMatchedAt map[string]map[Action]time.Time
	// podActionExecutedAt 记录 Pod 名称、动作对应的最近处理时间。
	podActionExecutedAt map[string]map[Action]time.Time
}

var _ Manager = (*manager)(nil)

// NewManager 创建管理器并初始化检查间隔，不启动后台协程。
// 其他配置校验随具体功能补充。
func NewManager(checkers []Checker, cfg Config) (Manager, error) {
	if cfg.CheckInterval < 0 {
		return nil, errors.New("check interval must not be negative")
	}
	if cfg.CheckInterval == 0 {
		cfg.CheckInterval = time.Minute
	}
	return &manager{
		k8sClient:           cfg.K8sClient,
		checkers:            slices.Clone(checkers),
		config:              cfg,
		webhookClient:       defaultWebhookClient,
		events:              make(chan Event, eventQueueCapacity),
		podActionMatchedAt:  make(map[string]map[Action]time.Time),
		podActionExecutedAt: make(map[string]map[Action]time.Time),
	}, nil
}

// Run 串行处理定时检查和外部事件，收到取消信号后退出。
func (m *manager) Run(ctx context.Context) {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.runOnce(ctx)
		case event := <-m.events:
			m.processEvent(ctx, event)
		}
	}
}

// runOnce 采集本轮状态并交给统一处理入口，不向调用方返回状态。
func (m *manager) runOnce(ctx context.Context) {
	results := make([]map[string]Status, len(m.checkers))
	var wg sync.WaitGroup
	for i, checker := range m.checkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			statuses, err := checker.Check(ctx)
			if err != nil {
				log.Printf("healthcheck checker %q failed: %v", checker.GetName(), err)
			}
			// 每个协程写入独立位置，等待全部完成后再合并。
			results[i] = statuses
		}()
	}
	wg.Wait()

	statuses := make(map[string][]Status)
	for _, result := range results {
		for podName, status := range result {
			if !slices.Contains(statuses[podName], status) {
				statuses[podName] = append(statuses[podName], status)
			}
		}
	}
	m.processStatuses(ctx, statuses)
}

// processStatuses 读取 Pod，匹配策略并在比例和冷却检查后写入动作标签。
func (m *manager) processStatuses(ctx context.Context, statuses map[string][]Status) {
	if pods := m.listPods(ctx); pods != nil {
		m.processPodStatuses(ctx, statuses, pods)
	}
}

// listPods 读取 Pod 并按名称建立索引。
func (m *manager) listPods(ctx context.Context) map[string]corev1.Pod {
	pods := &corev1.PodList{}
	if err := m.k8sClient.List(ctx, pods); err != nil {
		log.Printf("healthcheck list pods failed: %v", err)
		return nil
	}
	podMap := make(map[string]corev1.Pod, len(pods.Items))
	for _, pod := range pods.Items {
		podMap[pod.Name] = pod
	}
	return podMap
}

// processPodStatuses 使用同一份 Pod 快照处理检查和事件结果。
func (m *manager) processPodStatuses(ctx context.Context, statuses map[string][]Status, podMap map[string]corev1.Pod) {
	matchedPods := make([][]string, len(m.config.Policies))
	for podName, podStatuses := range statuses {
		if _, exists := podMap[podName]; !exists {
			continue
		}
		for i, policy := range m.config.Policies {
			if slices.ContainsFunc(podStatuses, func(status Status) bool {
				return slices.Contains(policy.Status, status)
			}) {
				matchedPods[i] = append(matchedPods[i], podName)
				break
			}
		}
	}
	m.clearExpiredPods()

	// 先记录全部匹配结果，超限而未执行的 Pod 也计入统计窗口。
	now := time.Now()
	for i, podNames := range matchedPods {
		action := m.config.Policies[i].Action
		for _, podName := range podNames {
			if m.podActionMatchedAt[podName] == nil {
				m.podActionMatchedAt[podName] = make(map[Action]time.Time)
			}
			m.podActionMatchedAt[podName][action] = now
		}
	}
	blockedActions := make(map[Action]bool)
	for _, policy := range m.config.Policies {
		if m.exceedsThreshold(policy, podMap) {
			blockedActions[policy.Action] = true
			log.Printf("healthcheck action %q skipped: matched pod ratio exceeds threshold %g or no pods available", policy.Action, policy.Threshold)
		}
	}
	resultsByPolicy := make([][]webhookPodResult, len(m.config.Policies))
	for i, podNames := range matchedPods {
		policy := m.config.Policies[i]
		for _, podName := range podNames {
			pod := podMap[podName]
			result := "skipped_threshold"
			if !blockedActions[policy.Action] {
				result = m.applyPodAction(ctx, policy.Action, pod)
			}
			if len(policy.Webhooks) > 0 {
				resultsByPolicy[i] = append(resultsByPolicy[i], webhookPodResult{
					Namespace: pod.Namespace,
					Name:      pod.Name,
					Result:    result,
				})
			}
		}
	}
	for i, results := range resultsByPolicy {
		if len(results) == 0 {
			continue
		}
		policy := m.config.Policies[i]
		go m.postWebhooks(ctx, slices.Clone(policy.Webhooks), webhookPayload{Action: policy.Action, Pods: results})
	}
}

// applyPodAction 检查冷却和人工管理状态，写入动作标签并返回本轮结果。
func (m *manager) applyPodAction(ctx context.Context, action Action, pod corev1.Pod) string {
	if _, executed := m.podActionExecutedAt[pod.Name][action]; executed {
		return "skipped_cooldown"
	}
	var labelKey, labelValue string
	switch action {
	case ActionMigrate, ActionReCreate, ActionInPlaceUpdate:
		if pod.Labels[k8s.LabelManualManagement] == "true" {
			return "skipped_manual"
		}
		labelKey, labelValue = k8s.LabelHealthStatus, string(action)
	case ActionManual:
		labelKey, labelValue = k8s.LabelManualManagement, "true"
	case ActionAutomatic:
		labelKey, labelValue = k8s.LabelManualManagement, "false"
	default:
		log.Printf("healthcheck unknown action %q skipped", action)
		return "skipped_unknown_action"
	}
	before := pod.DeepCopy()
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}
	pod.Labels[labelKey] = labelValue
	// 带版本检查的局部更新保留其他字段，冲突留待下轮重试。
	if err := m.k8sClient.Patch(ctx, &pod, client.MergeFromWithOptions(before, client.MergeFromWithOptimisticLock{})); err != nil {
		log.Printf("healthcheck action %q for pod %q failed: %v", action, pod.Name, err)
		return "label_write_failed"
	}
	if m.podActionExecutedAt[pod.Name] == nil {
		m.podActionExecutedAt[pod.Name] = make(map[Action]time.Time)
	}
	m.podActionExecutedAt[pod.Name][action] = time.Now()
	return "label_written"
}

// exceedsThreshold 只判断累计命中比例，不更新记录；无 Pod 时暂停处理。
func (m *manager) exceedsThreshold(policy Policy, pods map[string]corev1.Pod) bool {
	if len(pods) == 0 {
		return true
	}
	matchedCount := 0
	for podName := range pods {
		if _, exists := m.podActionMatchedAt[podName][policy.Action]; exists {
			matchedCount++
		}
	}
	return float64(matchedCount)/float64(len(pods)) > policy.Threshold
}

// clearExpiredPods 按策略窗口清理过期记录，并删除没有动作记录的 Pod 条目。
func (m *manager) clearExpiredPods() {
	now := time.Now()
	for _, records := range []map[string]map[Action]time.Time{
		m.podActionMatchedAt,
		m.podActionExecutedAt,
	} {
		for podName, actionTimes := range records {
			for _, policy := range m.config.Policies {
				at, exists := actionTimes[policy.Action]
				if exists && now.Sub(at) > policy.ReconcileInterval {
					delete(actionTimes, policy.Action)
				}
			}
			if len(actionTimes) == 0 {
				delete(records, podName)
			}
		}
	}
}
