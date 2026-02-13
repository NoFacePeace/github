package config

import "time"

var config *Config

type Config struct {
	Eks
}

type Eks struct {
	Id string `json:"id" yaml:"id"`
}

func Get() *Config {
	return config
}

func GetPollInterval() time.Duration {
	return time.Second
}

func GetPollTimeout() time.Duration {
	return time.Second
}

func GetPollImmediate() bool {
	return false
}

func InDryRunMode(cls string) bool {
	return false
}
