package service

import (
	"context"

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
	return nil
}
