package combiner

import (
	"example.com/acme/tracestore/pkg/api"
	"example.com/acme/tracestore/pkg/collector"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"go.uber.org/atomic"
)

var (
	_ GRPCCombiner[*tracestorepb.SearchTagValuesResponse]   = (*genericCombiner[*tracestorepb.SearchTagValuesResponse])(nil)
	_ GRPCCombiner[*tracestorepb.SearchTagValuesV2Response] = (*genericCombiner[*tracestorepb.SearchTagValuesV2Response])(nil)
)

func NewSearchTagValues(limitBytes int) Combiner {
	// Distinct collector with no limit
	d := collector.NewDistinctStringWithDiff(limitBytes)
	inspectedBytes := atomic.NewUint64(0)

	c := &genericCombiner[*tracestorepb.SearchTagValuesResponse]{
		httpStatusCode: 200,
		new:            func() *tracestorepb.SearchTagValuesResponse { return &tracestorepb.SearchTagValuesResponse{} },
		current:        &tracestorepb.SearchTagValuesResponse{TagValues: make([]string, 0)},
		combine: func(partial, _ *tracestorepb.SearchTagValuesResponse, _ PipelineResponse) error {
			for _, v := range partial.TagValues {
				d.Collect(v)
			}
			if partial.Metrics != nil {
				inspectedBytes.Add(partial.Metrics.InspectedBytes)
			}
			return nil
		},
		finalize: func(final *tracestorepb.SearchTagValuesResponse) (*tracestorepb.SearchTagValuesResponse, error) {
			final.TagValues = d.Strings()
			// return metrics in final response
			// TODO: merge with other metrics as well, when we have them, return only InspectedBytes for now
			final.Metrics = &tracestorepb.MetadataMetrics{InspectedBytes: inspectedBytes.Load()}
			return final, nil
		},
		quit: func(_ *tracestorepb.SearchTagValuesResponse) bool {
			return d.Exceeded()
		},
		diff: func(response *tracestorepb.SearchTagValuesResponse) (*tracestorepb.SearchTagValuesResponse, error) {
			resp, err := d.Diff()
			if err != nil {
				return nil, err
			}
			response.TagValues = resp
			// also return latest metrics along with diff
			// TODO: merge with other metrics as well, when we have them, return only InspectedBytes for now
			response.Metrics = &tracestorepb.MetadataMetrics{InspectedBytes: inspectedBytes.Load()}
			return response, nil
		},
	}
	initHTTPCombiner(c, api.HeaderAcceptJSON)
	return c
}

func NewTypedSearchTagValues(limitBytes int) GRPCCombiner[*tracestorepb.SearchTagValuesResponse] {
	return NewSearchTagValues(limitBytes).(GRPCCombiner[*tracestorepb.SearchTagValuesResponse])
}

func NewSearchTagValuesV2(limitBytes int) Combiner {
	// Distinct collector with no limit and diff enabled
	d := collector.NewDistinctValueWithDiff(limitBytes, func(tv tracestorepb.TagValue) int { return len(tv.Type) + len(tv.Value) })
	inspectedBytes := atomic.NewUint64(0)

	c := &genericCombiner[*tracestorepb.SearchTagValuesV2Response]{
		httpStatusCode: 200,
		current:        &tracestorepb.SearchTagValuesV2Response{TagValues: []*tracestorepb.TagValue{}},
		new:            func() *tracestorepb.SearchTagValuesV2Response { return &tracestorepb.SearchTagValuesV2Response{} },
		combine: func(partial, _ *tracestorepb.SearchTagValuesV2Response, _ PipelineResponse) error {
			for _, v := range partial.TagValues {
				d.Collect(*v)
			}
			if partial.Metrics != nil {
				inspectedBytes.Add(partial.Metrics.InspectedBytes)
			}
			return nil
		},
		finalize: func(final *tracestorepb.SearchTagValuesV2Response) (*tracestorepb.SearchTagValuesV2Response, error) {
			values := d.Values()
			final.TagValues = make([]*tracestorepb.TagValue, 0, len(values))
			for _, v := range values {
				v2 := v
				final.TagValues = append(final.TagValues, &v2)
			}
			// load Inspected Bytes here and return along with final response
			// TODO: merge with other metrics as well, when we have them, return only InspectedBytes for now
			final.Metrics = &tracestorepb.MetadataMetrics{InspectedBytes: inspectedBytes.Load()}
			return final, nil
		},
		quit: func(_ *tracestorepb.SearchTagValuesV2Response) bool {
			return d.Exceeded()
		},
		diff: func(response *tracestorepb.SearchTagValuesV2Response) (*tracestorepb.SearchTagValuesV2Response, error) {
			diff, err := d.Diff()
			if err != nil {
				return nil, err
			}
			response.TagValues = make([]*tracestorepb.TagValue, 0, len(diff))
			for _, v := range diff {
				v2 := v
				response.TagValues = append(response.TagValues, &v2)
			}
			// also return metrics along with diffs
			// TODO: merge with other metrics as well, when we have them, return only InspectedBytes for now
			response.Metrics = &tracestorepb.MetadataMetrics{InspectedBytes: inspectedBytes.Load()}
			return response, nil
		},
	}
	initHTTPCombiner(c, api.HeaderAcceptJSON)
	return c
}

func NewTypedSearchTagValuesV2(limitBytes int) GRPCCombiner[*tracestorepb.SearchTagValuesV2Response] {
	return NewSearchTagValuesV2(limitBytes).(GRPCCombiner[*tracestorepb.SearchTagValuesV2Response])
}
