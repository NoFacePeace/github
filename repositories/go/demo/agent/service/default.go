package service

import (
	"context"
	"fmt"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/event"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/loop"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/model"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/provider"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/tool"
)

type Config struct {
	Defaults Defaults   `yaml:"defaults"`
	Loop     LoopConfig `yaml:"loop"`
}

type Defaults struct {
	Provider string `yaml:"provider"`
	Protocol string `yaml:"protocol"`
	Model    string `yaml:"model"`
}

type LoopConfig struct {
	MaxSteps int `yaml:"max_steps"`
}

type ModelFactory func(name string) model.Model

type ProviderEntry struct {
	Provider  provider.Provider
	Factories map[string]ModelFactory // 协议名称对应的模型工厂，闭包复用已初始化的协议客户端
}

type DefaultService struct {
	config    Config
	providers map[string]ProviderEntry
	loop      loop.Loop
	tools     []tool.Tool
}

func NewDefault(config Config, providers map[string]ProviderEntry, agentLoop loop.Loop, tools []tool.Tool) *DefaultService {
	return &DefaultService{
		config:    config,
		providers: providers,
		loop:      agentLoop,
		tools:     tools,
	}
}

func (s *DefaultService) Run(ctx context.Context, req Request, onEvent func(event.Event)) error {
	if req.Provider == "" {
		req.Provider = s.config.Defaults.Provider
	}
	if req.Protocol == "" {
		req.Protocol = s.config.Defaults.Protocol
	}
	if req.Model == "" {
		req.Model = s.config.Defaults.Model
	}

	entry, ok := s.providers[req.Provider]
	if !ok {
		return fmt.Errorf("unknown provider: %s", req.Provider)
	}
	factory, ok := entry.Factories[req.Protocol]
	if !ok {
		return fmt.Errorf("unsupported protocol: %s", req.Protocol)
	}
	if req.Model == "" {
		return fmt.Errorf("model is required")
	}
	if req.SessionID != "" {
		return fmt.Errorf("sessions are not supported yet")
	}

	return s.loop.Run(ctx, loop.Request{
		Model: factory(req.Model),
		Messages: []model.Message{
			{Role: model.RoleUser, Content: req.Input},
		},
		Tools:    s.tools,
		MaxSteps: s.config.Loop.MaxSteps,
	}, onEvent)
}
