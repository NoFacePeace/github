package model

const (
	Domain            = "nofacepeace.github.io"
	ContainerNameMain = "main"
)

// cluster label
const (
	LabelClusterId        = Domain + "/cluster-id"
	LabelMiddlewareType   = Domain + "/middleware-type"
	LabelEksId            = Domain + "/eks-id"
	LabelClusterIgnore    = Domain + "/cluster-ignore"
	LabelManualManagement = Domain + "/manual-management"
	LabelManagedBy        = Domain + "/managed-by"
	LabelNodeSetName      = Domain + "/nodeset-name"
	LabelIsEndpoint       = Domain + "/is-endpoint"
	LabelPodName          = Domain + "/pod-name"
	LabelNodeName         = Domain + "/node-name"
	LabelClusterName      = Domain + "/cluster-name"
	LabelAvailabilityZone = Domain + "/availability-zone"
	LabelRegionZone       = Domain + "/region-zone"
	LabelNodeSetDomain    = Domain + "/nodeset-domain"
)

// annotation
const (
	AnnotationTplVersion     = Domain + "/tpl-version"
	AnnotationPublishId      = Domain + "/publish-id"
	AnnotationNodeGeneration = Domain + "/node-generation"
)

type EventType int

const (
	EventTypeClusterCreate EventType = iota
	EventTypeClusterDelete
	EventTypeClusterReconcile
)
