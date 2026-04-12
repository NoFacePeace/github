package model

const (
	Domain            = "nofacepeace.github.io"
	ContainerNameMain = "main"
)

// cluster label
const (
	LabelClusterId                      = Domain + "/cluster-id"
	LabelMiddlewareType                 = Domain + "/middleware-type"
	LabelEksId                          = Domain + "/eks-id"
	LabelClusterIgnore                  = Domain + "/cluster-ignore"
	LabelManualManagement               = Domain + "/manual-management"
	LabelManagedBy                      = Domain + "/managed-by"
	LabelNodeSetName                    = Domain + "/nodeset-name"
	LabelIsEndpoint                     = Domain + "/is-endpoint"
	LabelPodName                        = Domain + "/pod-name"
	LabelNodeName                       = Domain + "/node-name"
	LabelClusterName                    = Domain + "/cluster-name"
	LabelAvailabilityZone               = Domain + "/availability-zone"
	LabelRegionZone                     = Domain + "/region-zone"
	LabelNodeSetDomain                  = Domain + "/nodeset-domain"
	LabelNodeHealthStatus               = Domain + "/node-health-status"
	LabelNodeHealthStatusValueUnhealthy = "unhealthy"
)

// annotation
const (
	AnnotationTplVersion     = Domain + "/tpl-version"
	AnnotationPublishId      = Domain + "/publish-id"
	AnnotationNodeGeneration = Domain + "/node-generation"
	AnnotationContainers     = Domain + "/containers"
	AnnotationInplaceVpa     = Domain + "/inplace-vpa"
	AnnotationVersion        = Domain + "/version"
)

type EventType int

const (
	EventTypeClusterCreate EventType = iota
	EventTypeClusterDelete
	EventTypeClusterReconcile
)

// feature
const (
	FeatureDeleteCluster = "delete_cluster"
	FeatureInplaceVpa    = "inplace_vpa"
)

const (
	VolumeNameConfigFiles      = "config-files"
	VolumeConfigFilesMountPath = "/etc/podconfig"
)

const (
	CtxKeySkipSleep = "skip-sleep"
)

type CheckerResult int

const (
	CheckerResultOk CheckerResult = iota
	CheckerResultSuspend
)
