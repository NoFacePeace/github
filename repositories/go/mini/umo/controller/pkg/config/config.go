package config

import "time"

var config = &Config{}

type Config struct {
	Eks
	MiddlewareType      string               `json:"middlewareType" yaml:"middlewareType"`
	ClusterFilterPolicy *ClusterFilterPolicy `json:"cluster_filter_policy" yaml:"clusterFilterPolicy"`
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

type ClusterFilterPolicy struct {
	Mode        string   `json:"mode" yaml:"mode"`
	ClusterList []string `json:"cluster_list" yaml:"clusterList"`
}

func (c *ClusterFilterPolicy) Skip(cls string) bool {
	return false
}
