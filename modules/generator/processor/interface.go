package processor

import (
	"context"

	"example.com/acme/tracestore/pkg/tracestorepb"
)

type Processor interface {
	// Name returns the name of the processor.
	Name() string

	// PushSpans processes a batch of spans and updates the metrics registered in RegisterMetrics.
	PushSpans(ctx context.Context, req *tracestorepb.PushSpansRequest)

	// Shutdown releases any resources allocated by the processor. Once the processor is shut down,
	// PushSpans should not be called anymore.
	Shutdown(ctx context.Context)
}
