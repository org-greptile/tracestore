package instrumentation

import (
	"github.com/cristalhq/hedgedhttp"
	"example.com/acme/tracestore/pkg/hedgedmetrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var hedgedRequestsMetrics = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "tracestoredb",
		Name:      "backend_hedged_roundtrips_total",
		Help:      "Total number of hedged backend requests.",
	},
)

// PublishHedgedMetrics flushes metrics from hedged requests every 10 seconds
func PublishHedgedMetrics(s *hedgedhttp.Stats) {
	hedgedmetrics.Publish(s, hedgedRequestsMetrics)
}
