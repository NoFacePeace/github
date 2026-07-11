package config

import (
	"time"

	umov1 "nofacepeace.github.io/controller/api/v1"
)

var (
	config                = &Config{}
	DefaultUpdateStrategy = &umov1.UpdateStrategy{
		Concurrency:         1,
		PodUpdateIntervalMs: 1000,
		PodExecTimeoutMs:    60000,
		SkipChecker:         false,
		OnFailure:           umov1.OnFailureActionTerminate,
	}
)

type Config struct {
	Eks
	MiddlewareType           string               `json:"middlewareType" yaml:"middlewareType"`
	ClusterFilterPolicy      *ClusterFilterPolicy `json:"cluster_filter_policy" yaml:"clusterFilterPolicy"`
	ReconcilePolicy          *ReconcilePolicy     `json:"reconcile_policy" yaml:"reconcilePolicy"`
	ControllerName           string               `json:"controller_name" yaml:"controllerName"`
	PatchOnlyKeyPrefixes     []string             `json:"patch_only_key_prefixes" yaml:"patchOnlyKeyPrefixes"`
	InPlaceUpdateKeyPrefixes []string             `json:"in_place_update_key_prefixes" yaml:"inPlaceUpdateKeyPrefixes"`
}

type Eks struct {
	Id string `json:"id" yaml:"id"`
	Az string `json:"az" yaml:"az"`
	Rz string `json:"rz" yaml:"rz"`
}

type ReconcilePolicy struct {
	NodeCheckErrorTolerance int                             `json:"node_check_error_tolerance" yaml:"nodeCheckErrorTolerance"`
	UpdateStrategys         map[string]umov1.UpdateStrategy `json:"update_strategy" yaml:"updateStrategy"`
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
