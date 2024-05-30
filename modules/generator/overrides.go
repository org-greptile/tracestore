package generator

import (
	"time"

	"example.com/acme/tracestore/modules/generator/registry"
	"example.com/acme/tracestore/modules/generator/storage"
	"example.com/acme/tracestore/modules/overrides"
	"example.com/acme/tracestore/pkg/sharedconfig"
	filterconfig "example.com/acme/tracestore/pkg/spanfilter/config"
	"example.com/acme/tracestore/tracestoredb/backend"
)

type metricsGeneratorOverrides interface {
	registry.Overrides
	storage.Overrides

	MetricsGeneratorIngestionSlack(userID string) time.Duration
	MetricsGeneratorProcessors(userID string) map[string]struct{}
	MetricsGeneratorProcessorServiceGraphsHistogramBuckets(userID string) []float64
	MetricsGeneratorProcessorServiceGraphsDimensions(userID string) []string
	MetricsGeneratorProcessorServiceGraphsPeerAttributes(userID string) []string
	MetricsGeneratorProcessorSpanMetricsHistogramBuckets(userID string) []float64
	MetricsGeneratorProcessorSpanMetricsDimensions(userID string) []string
	MetricsGeneratorProcessorSpanMetricsIntrinsicDimensions(userID string) map[string]bool
	MetricsGeneratorProcessorSpanMetricsFilterPolicies(userID string) []filterconfig.FilterPolicy
	MetricsGeneratorProcessorLocalBlocksMaxLiveTraces(userID string) uint64
	MetricsGeneratorProcessorLocalBlocksMaxBlockDuration(userID string) time.Duration
	MetricsGeneratorProcessorLocalBlocksMaxBlockBytes(userID string) uint64
	MetricsGeneratorProcessorLocalBlocksTraceIdlePeriod(userID string) time.Duration
	MetricsGeneratorProcessorLocalBlocksFlushCheckPeriod(userID string) time.Duration
	MetricsGeneratorProcessorLocalBlocksCompleteBlockTimeout(userID string) time.Duration
	MetricsGeneratorProcessorSpanMetricsDimensionMappings(userID string) []sharedconfig.DimensionMappings
	MetricsGeneratorProcessorSpanMetricsEnableTargetInfo(userID string) bool
	MetricsGeneratorProcessorServiceGraphsEnableClientServerPrefix(userID string) bool
	MetricsGeneratorProcessorServiceGraphsEnableMessagingSystemLatencyHistogram(userID string) bool
	MetricsGeneratorProcessorServiceGraphsEnableVirtualNodeLabel(userID string) bool
	MetricsGeneratorProcessorSpanMetricsTargetInfoExcludedDimensions(userID string) []string
	DedicatedColumns(userID string) backend.DedicatedColumns
	MaxBytesPerTrace(userID string) int
	UnsafeQueryHints(userID string) bool
}

var _ metricsGeneratorOverrides = (overrides.Interface)(nil)
