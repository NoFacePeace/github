package alert

import (
	"context"
	"fmt"
	"strings"

	"github.com/NoFacePeace/github/repositories/go/mini/umo/agent/chat"
)

type Analyzer interface {
	Analyze(ctx context.Context, alert *Alert, binding *Binding) error
}

type AnalysisStatus string

const (
	AnalysisStatusPending AnalysisStatus = "pending"
	AnalysisStatusFailed  AnalysisStatus = "failed"
	AnalysisStatusSuccess AnalysisStatus = "success"
)

type LLMAnalyzer struct {
	chat chat.Chat
}

func (a *LLMAnalyzer) Analyze(ctx context.Context, alert *Alert, binding *Binding) (string, string, error) {
	if a.chat == nil {
		return "", "", fmt.Errorf("chat is not initialized, alert %v, binding %v", alert.ID, binding.ID)
	}
	prompt, err := a.buildPrompt(alert, binding)
	if err != nil {
		return "", "", fmt.Errorf("build prompt: [%w]", err)
	}

	// 准备回合
	turn, err := a.chat.Prepare(prompt)
	if err != nil {
		return "", "", fmt.Errorf("chat prepare: [%w]", err)
	}

	// 执行回合
	result, err := a.chat.Execute(turn)
	if err != nil {
		return "", "", fmt.Errorf("chat execute: [%w]", err)
	}
	return turn.Session.ID, strings.TrimSpace(result.Output), nil
}

func (a *LLMAnalyzer) buildPrompt(alert *Alert, binding *Binding) (string, error) {
	return "", nil
}
