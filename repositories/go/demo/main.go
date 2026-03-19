package main

import "fmt"

func main() {
	cfg := &Config{
		SubConfig: &SubConfig{
			Name: "test",
		},
	}
	sub := NewSub(cfg.SubConfig)
	sub.PrintName()
	*cfg = Config{
		SubConfig: &SubConfig{
			Name: "test2",
		},
	}
	sub.PrintName()
}

type Config struct {
	SubConfig *SubConfig
}

type SubConfig struct {
	Name string
}

type Sub struct {
	config *SubConfig
}

func NewSub(config *SubConfig) *Sub {
	return &Sub{
		config: config,
	}
}

func (s *Sub) PrintName() {
	fmt.Println(s.config.Name)
}
