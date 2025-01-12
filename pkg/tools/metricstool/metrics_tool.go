package metricstool

import (
	"kubeclusteragent/pkg/util/metrcis"
	"time"
)

type PrometheusMetricsTool struct {
	prometheusMetrics metrcis.PrometheusMetrics
	MetricsLabels     []string
}

func (p *PrometheusMetricsTool) PopulateToPrometheusMetrics(startTime time.Time) {
	p.prometheusMetrics = new(metrcis.LiveMetrics)
	p.prometheusMetrics.ClusterStatsHistogram(time.Since(startTime).Seconds(), p.MetricsLabels...)
}
