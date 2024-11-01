package combiner

import (
	"sort"
	"strings"

	"example.com/acme/tracestore/pkg/api"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"example.com/acme/tracestore/pkg/traceql"
)

var _ GRPCCombiner[*tracestorepb.QueryRangeResponse] = (*genericCombiner[*tracestorepb.QueryRangeResponse])(nil)

// NewQueryRange returns a query range combiner.
func NewQueryRange(req *tracestorepb.QueryRangeRequest, trackDiffs bool) (Combiner, error) {
	combiner, err := traceql.QueryRangeCombinerFor(req, traceql.AggregateModeFinal, trackDiffs)
	if err != nil {
		return nil, err
	}

	c := &genericCombiner[*tracestorepb.QueryRangeResponse]{
		httpStatusCode: 200,
		new:            func() *tracestorepb.QueryRangeResponse { return &tracestorepb.QueryRangeResponse{} },
		current:        &tracestorepb.QueryRangeResponse{Metrics: &tracestorepb.SearchMetrics{}},
		combine: func(partial *tracestorepb.QueryRangeResponse, _ *tracestorepb.QueryRangeResponse, _ PipelineResponse) error {
			if partial.Metrics != nil {
				// this is a coordination between the sharder and combiner. the sharder returns one response with summary metrics
				// only. the combiner correctly takes and accumulates that job. however, if the response has no jobs this is
				// an indicator this is a "real" response so we set CompletedJobs to 1 to increment in the combiner.
				if partial.Metrics.TotalJobs == 0 {
					partial.Metrics.CompletedJobs = 1
				}
			}

			combiner.Combine(partial)

			return nil
		},
		finalize: func(_ *tracestorepb.QueryRangeResponse) (*tracestorepb.QueryRangeResponse, error) {
			resp := combiner.Response()
			if resp == nil {
				resp = &tracestorepb.QueryRangeResponse{}
			}
			sortResponse(resp)
			return resp, nil
		},
		diff: func(_ *tracestorepb.QueryRangeResponse) (*tracestorepb.QueryRangeResponse, error) {
			resp := combiner.Diff()
			if resp == nil {
				resp = &tracestorepb.QueryRangeResponse{}
			}
			sortResponse(resp)
			return resp, nil
		},
	}

	initHTTPCombiner(c, api.HeaderAcceptJSON)

	return c, nil
}

func NewTypedQueryRange(req *tracestorepb.QueryRangeRequest, trackDiffs bool) (GRPCCombiner[*tracestorepb.QueryRangeResponse], error) {
	c, err := NewQueryRange(req, trackDiffs)
	if err != nil {
		return nil, err
	}
	return c.(GRPCCombiner[*tracestorepb.QueryRangeResponse]), nil
}

func sortResponse(res *tracestorepb.QueryRangeResponse) {
	// Sort all output, series alphabetically, samples by time
	sort.SliceStable(res.Series, func(i, j int) bool {
		return strings.Compare(res.Series[i].PromLabels, res.Series[j].PromLabels) == -1
	})
	for _, series := range res.Series {
		sort.Slice(series.Samples, func(i, j int) bool {
			return series.Samples[i].TimestampMs < series.Samples[j].TimestampMs
		})
		sort.Slice(series.Exemplars, func(i, j int) bool {
			return series.Exemplars[i].TimestampMs < series.Exemplars[j].TimestampMs
		})
	}
}
