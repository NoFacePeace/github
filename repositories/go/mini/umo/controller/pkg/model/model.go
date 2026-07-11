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
	LabelHealthStatus                   = Domain + "/health-status"
	LabelHealthStatusValueMigrate       = "migrate"
	LabelHealthStatusValueRecreate      = "recreate"
	LabelHealthStatusValueInPlaceUpdate = "inplace-update"
)

// annotation
const (
	AnnotationTplVersion      = Domain + "/tpl-version"
	AnnotationPublishID       = Domain + "/publish-id"
	AnnotationNodeGeneration  = Domain + "/node-generation"
	AnnotationContainers      = Domain + "/containers"
	AnnotationInplaceVpa      = Domain + "/inplace-vpa"
	AnnotationVersion         = Domain + "/version"
	AnnotationPostCheckFailed = Domain + "/post-check-failed"
	AnnotationPostCheckAction = Domain + "/post-check-action"
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

// action
const (
	ActionPostCheckCreateNode        = "action-post-check-create-node"
	ActionPostCheckMigrateNode       = "action-post-check-migrate-node"
	ActionPreCheckMigrateNode        = "action-pre-check-migrate-node"
	ActionPreCheckCreateNode         = "action-pre-check-create-node"
	ActionPreCheckRecreateNode       = "action-pre-check-recreate-node"
	ActionPreCheckInPlaceNode        = "action-pre-check-in-place-node"
	ActionPreCheckSidecarInPlaceNode = "action-pre-check-sidecar-in-place-node"
)

const (
	NodeActionDefault int = iota
	NodeActionInPlaceUpdate
	NodeActionRecreate
)
