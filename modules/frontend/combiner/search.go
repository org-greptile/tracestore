package combiner

import (
	"sort"

	"example.com/acme/tracestore/pkg/api"
	"example.com/acme/tracestore/pkg/search"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"example.com/acme/tracestore/pkg/traceql"
)

var _ GRPCCombiner[*tracestorepb.SearchResponse] = (*genericCombiner[*tracestorepb.SearchResponse])(nil)

// NewSearch returns a search combiner
func NewSearch(limit int) Combiner {
	metadataCombiner := traceql.NewMetadataCombiner()
	diffTraces := map[string]struct{}{}

	c := &genericCombiner[*tracestorepb.SearchResponse]{
		httpStatusCode: 200,
		new:            func() *tracestorepb.SearchResponse { return &tracestorepb.SearchResponse{} },
		current:        &tracestorepb.SearchResponse{Metrics: &tracestorepb.SearchMetrics{}},
		combine: func(partial *tracestorepb.SearchResponse, final *tracestorepb.SearchResponse, _ PipelineResponse) error {
			for _, t := range partial.Traces {
				// if we've reached the limit and this is NOT a new trace then skip it
				if limit > 0 &&
					metadataCombiner.Count() >= limit &&
					!metadataCombiner.Exists(t.TraceID) {
					continue
				}

				metadataCombiner.AddMetadata(t)
				// record modified traces
				diffTraces[t.TraceID] = struct{}{}
			}

			if partial.Metrics != nil {
				// there is a coordination with the search sharder here. normal responses
				// will never have total jobs set, but they will have valid Inspected* values
				// a special response is sent back from the sharder with no traces but valid Total* values
				// if TotalJobs is nonzero then assume its the special response
				if partial.Metrics.TotalJobs == 0 {
					final.Metrics.CompletedJobs++

					final.Metrics.InspectedBytes += partial.Metrics.InspectedBytes
					final.Metrics.InspectedTraces += partial.Metrics.InspectedTraces
				} else {
					final.Metrics.TotalBlocks += partial.Metrics.TotalBlocks
					final.Metrics.TotalJobs += partial.Metrics.TotalJobs
					final.Metrics.TotalBlockBytes += partial.Metrics.TotalBlockBytes
				}
			}

			return nil
		},
		finalize: func(final *tracestorepb.SearchResponse) (*tracestorepb.SearchResponse, error) {
			// metrics are already combined on the passed in final
			final.Traces = metadataCombiner.Metadata()

			addRootSpanNotReceivedText(final.Traces)
			return final, nil
		},
		diff: func(current *tracestorepb.SearchResponse) (*tracestorepb.SearchResponse, error) {
			// wipe out any existing traces and recreate from the map
			diff := &tracestorepb.SearchResponse{
				Traces:  make([]*tracestorepb.TraceSearchMetadata, 0, len(diffTraces)),
				Metrics: current.Metrics,
			}

			for _, tr := range metadataCombiner.Metadata() {
				// if not in the map, skip. we haven't seen an update
				if _, ok := diffTraces[tr.TraceID]; !ok {
					continue
				}

				diff.Traces = append(diff.Traces, tr)
			}

			sort.Slice(diff.Traces, func(i, j int) bool {
				return diff.Traces[i].StartTimeUnixNano > diff.Traces[j].StartTimeUnixNano
			})

			addRootSpanNotReceivedText(diff.Traces)

			// wipe out diff traces for the next time
			clear(diffTraces)

			return diff, nil
		},
		// search combiner doesn't use current in the way i would have expected. it only tracks metrics through current and uses the results map for the actual traces.
		//  should we change this?
		quit: func(_ *tracestorepb.SearchResponse) bool {
			if limit <= 0 {
				return false
			}

			return metadataCombiner.Count() >= limit
		},
	}
	initHTTPCombiner(c, api.HeaderAcceptJSON)
	return c
}

func addRootSpanNotReceivedText(results []*tracestorepb.TraceSearchMetadata) {
	for _, tr := range results {
		if tr.RootServiceName == "" {
			tr.RootServiceName = search.RootSpanNotYetReceivedText
		}
	}
}

func NewTypedSearch(limit int) GRPCCombiner[*tracestorepb.SearchResponse] {
	return NewSearch(limit).(GRPCCombiner[*tracestorepb.SearchResponse])
}
