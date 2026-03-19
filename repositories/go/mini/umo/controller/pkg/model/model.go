package model

const (
	Domain = "nofacepeace.github.io"
)

// cluster label
const (
	LabelClusterId        = Domain + "/cluster-id"
	LabelMiddlewareType   = Domain + "/middleware-type"
	LabelEksId            = Domain + "/eks-id"
	LabelClusterIgnore    = Domain + "/cluster-ignore"
	LabelManualManagement = Domain + "/manual-management"
)

type EventType int

const (
	EventTypeClusterCreate EventType = iota
	EventTypeClusterDelete
	EventTypeClusterReconcile
)
