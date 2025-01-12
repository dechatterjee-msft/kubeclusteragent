package metrcis

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

var reg = prometheus.NewRegistry()

const (
	Success                   = "success"
	Failed                    = "failed"
	ClusterCreated            = "cluster-created"
	ClusterFailed             = "cluster-failed"
	DeleteCompleted           = "delete-done"
	DeleteFailed              = "delete-failed"
	ClusterAlreadyInitialized = "cluster-already-initialized"
	UpgradeDone               = "upgrade-done"
	UpgradeFailed             = "upgrade-failed"
	ResetDone                 = "reset-done"
	ResetFailed               = "reset-failed"
)

type LiveMetrics struct{}

var prometheusMetrics PrometheusMetrics = &LiveMetrics{}

var clusterStatsHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "http_request",
	Name:      "request_frequency",
	Help:      "Histogram of http_request",
	Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 50, 100, 150, 200, 250, 300, 350, 400, 450, 500, 550, 600},
}, []string{"cluster_type", "version", "code", "method", "path"})

var clusterStatusSummary = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "cluster",
	Name:      "installation_status",
	Help:      "last when the cluster is created",
}, []string{"action", "status"})

type PrometheusMetrics interface {
	AgentAppMetrics() prometheus.Collector
	AgentProcessMetrics() prometheus.Collector
	ClusterLcmStatus(status string)
	ClusterStatsHistogram(time float64, hist ...string)
}

func (l LiveMetrics) ClusterStatsHistogram(time float64, hist ...string) {
	clusterStatsHistogram.WithLabelValues(hist...).Observe(time)
}

func (l LiveMetrics) ClusterLcmStatus(status string) {
	clusterStatusSummary.WithLabelValues(status).SetToCurrentTime()
}

func MetricsRegistry() *prometheus.Registry {
	reg.MustRegister(prometheusMetrics.AgentAppMetrics())
	reg.MustRegister(prometheusMetrics.AgentProcessMetrics())
	reg.MustRegister(clusterStatsHistogram)
	reg.MustRegister(clusterStatusSummary)
	return reg
}

func (l LiveMetrics) AgentAppMetrics() prometheus.Collector {
	return collectors.NewGoCollector()
}

func (l LiveMetrics) AgentProcessMetrics() prometheus.Collector {
	return collectors.NewProcessCollector(collectors.ProcessCollectorOpts(prometheus.ProcessCollectorOpts{Namespace: "war_machine_agent"}))
}
