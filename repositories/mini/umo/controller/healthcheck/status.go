package healthcheck

// Status 表示健康检查发现的异常类型，不等同于 Pod 的运行阶段。
type Status string

// 节点异常状态。
const (
	StatusNodeNotReady           Status = "node_not_ready"
	StatusNodeReadyUnknown       Status = "node_ready_unknown"
	StatusNodeUnreachable        Status = "node_unreachable"
	StatusNodeNetworkUnavailable Status = "node_network_unavailable"
	StatusNodeDiskPressure       Status = "node_disk_pressure"
	StatusNodeMemoryPressure     Status = "node_memory_pressure"
	StatusNodePIDPressure        Status = "node_pid_pressure"
	StatusNodeManualOffline      Status = "manual_offline"
)

// 磁盘异常状态，包括检测程序本身失败。
const (
	StatusDiskReadOnly             Status = "disk_read_only"
	StatusDiskNoSpace              Status = "disk_no_space"
	StatusDiskReadFailed           Status = "disk_read_failed"
	StatusDiskWriteFailed          Status = "disk_write_failed"
	StatusDiskReadLatencyAbnormal  Status = "disk_read_latency_abnormal"
	StatusDiskWriteLatencyAbnormal Status = "disk_write_latency_abnormal"
	StatusDiskDetectionFailed      Status = "disk_detection_failed"
)
