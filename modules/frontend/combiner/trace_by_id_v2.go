package combiner

import (
	"fmt"

	"example.com/acme/tracestore/pkg/model/trace"
	"example.com/acme/tracestore/pkg/tracestorepb"
)

func NewTraceByIDV2(maxBytes int, marshalingFormat string) Combiner {
	combiner := trace.NewCombiner(maxBytes, true)
	var partialTrace bool
	gc := &genericCombiner[*tracestorepb.TraceByIDResponse]{
		combine: func(partial *tracestorepb.TraceByIDResponse, _ *tracestorepb.TraceByIDResponse, _ PipelineResponse) error {
			if partial.Status == tracestorepb.TraceByIDResponse_PARTIAL {
				partialTrace = true
			}
			_, err := combiner.Consume(partial.Trace)
			return err
		},
		finalize: func(resp *tracestorepb.TraceByIDResponse) (*tracestorepb.TraceByIDResponse, error) {
			traceResult, _ := combiner.Result()
			if traceResult == nil {
				traceResult = &tracestorepb.Trace{}
			}

			// dedupe duplicate span ids
			deduper := newDeduper()
			traceResult = deduper.dedupe(traceResult)
			resp.Trace = traceResult

			if partialTrace || combiner.IsPartialTrace() {
				resp.Status = tracestorepb.TraceByIDResponse_PARTIAL
				resp.Message = fmt.Sprintf("Trace exceeds maximum size of %d bytes, a partial trace is returned", maxBytes)
			}

			return resp, nil
		},
		new:     func() *tracestorepb.TraceByIDResponse { return &tracestorepb.TraceByIDResponse{} },
		current: &tracestorepb.TraceByIDResponse{},
	}
	initHTTPCombiner(gc, marshalingFormat)
	return gc
}
