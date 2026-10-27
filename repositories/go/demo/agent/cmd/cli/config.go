package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/NoFacePeace/github/repositories/go/demo/agent/loop"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/model"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/model/chatcompletion"
	protocol "github.com/NoFacePeace/github/repositories/go/demo/agent/protocol/chatcompletion"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/provider"
	"github.com/NoFacePeace/github/repositories/go/demo/agent/service"
	"gopkg.in/yaml.v3"
)

type config struct {
	Service   service.Config            `yaml:"service"`
	Providers map[string]providerConfig `yaml:"providers"`
}

type providerConfig struct {
	BaseURL     string      `yaml:"base_url"`
	APIKey      string      `yaml:"api_key"`
	Headers     http.Header `yaml:"headers"`
	Protocols   []string    `yaml:"protocols"`
	ModelSource string      `yaml:"model_source"`
	Models      []string    `yaml:"models"`
}

func buildService(path string) (service.Service, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read configuration: %w", err)
	}
	var cfg config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		// 不回显解析细节，避免格式错误时将凭据原文写入终端。
		return nil, fmt.Errorf("invalid YAML configuration")
	}
	providers := make(map[string]service.ProviderEntry, len(cfg.Providers))
	for name, entry := range cfg.Providers {
		connection := provider.Config{
			BaseURL: entry.BaseURL, APIKey: entry.APIKey, Headers: entry.Headers,
		}
		factories := make(map[string]service.ModelFactory, len(entry.Protocols))
		for _, name := range entry.Protocols {
			switch name {
			case "chat_completions":
				client := protocol.New(connection)
				factories[name] = func(name string) model.Model {
					return chatcompletion.New(client, name)
				}
			default:
				return nil, fmt.Errorf("unsupported configured protocol: %s", name)
			}
		}
		providers[name] = service.ProviderEntry{
			Provider: provider.New(connection), Factories: factories,
		}
	}
	return service.NewDefault(cfg.Service, providers, &loop.DefaultLoop{}, nil), nil
}
