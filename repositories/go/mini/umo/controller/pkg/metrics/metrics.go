package metrics

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	// ClusterReconcileTotal 统计 cluster 协调总次数，按 result(success/error) 和 namespace 分组
	ClusterReconcileTotal *prometheus.CounterVec
)

func init() {
	ClusterReconcileTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "umo_cluster_reconcile_total",
			Help: "Total number of cluster reconciliation attempts by result and namespace",
		},
		[]string{"middleware", "eks", "cluster", "retry", "error"},
	)
}

func RecordClusterReconcile(middleware, eks, cluster string, retry bool, err error) {
	ClusterReconcileTotal.WithLabelValues(middleware, eks, cluster, strconv.FormatBool(retry), err.Error()).Inc()
}
