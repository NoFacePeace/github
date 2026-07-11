package alert

import (
	"context"
	"fmt"
)

type Service interface {
	ListAlerts() ([]Alert, error)
	ListRules() ([]Rule, error)
}

type service struct {
	analyzer Analyzer
	store    Store
}

func (s *service) AnalyzeAlert(ctx context.Context, id string) (string, error) {
	// 判断 analyzer 是否为 nil
	if s.analyzer == nil {
		return "", fmt.Errorf("alert id %s analyzer is nil", id)
	}

	// 获取 alert 和 binding
	alert, err := s.store.GetAlert(ctx, id)
	if err != nil {
		return "", fmt.Errorf("get alert: [%w]", err)
	}
	binding, err := s.store.GetBinding(ctx, id)
	if err != nil {
		return "", fmt.Errorf("get binding: [%w]", err)
	}

	// 创建分析结果
	err = s.store.CreateAnalysis(ctx, &Analysis{})
	if err != nil {
		return "", fmt.Errorf("create analysis: [%w]", err)
	}

	status := AnalysisStatusSuccess
	// 调用 analyzer 进行分析
	err = s.analyzer.Analyze(ctx, alert, binding)
	if err != nil {
		status = AnalysisStatusFailed
	}

	// 发送分析完成事件
	s.emitCompletion()
	// 更新分析结果
	if err := s.store.UpdateAnalysis(ctx, id, map[string]any{
		"status": status,
		"error":  err,
	}); err != nil {
		return "", fmt.Errorf("update analysis: [%w]", err)
	}
	return "", nil
}

func (s *service) emitCompletion() {
}

type Rule struct{}

type Alert struct {
	ID string
}

type Binding struct {
	ID string
}

type Analysis struct {
}
