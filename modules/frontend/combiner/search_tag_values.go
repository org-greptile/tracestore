package combiner

import (
	"example.com/acme/tracestore/pkg/tracestorepb"
	"example.com/acme/tracestore/pkg/util"
)

var (
	_ GRPCCombiner[*tracestorepb.SearchTagValuesResponse]   = (*genericCombiner[*tracestorepb.SearchTagValuesResponse])(nil)
	_ GRPCCombiner[*tracestorepb.SearchTagValuesV2Response] = (*genericCombiner[*tracestorepb.SearchTagValuesV2Response])(nil)
)

func NewSearchTagValues(limitBytes int) Combiner {
	// Distinct collector with no limit
	d := util.NewDistinctStringCollector(limitBytes)

	return &genericCombiner[*tracestorepb.SearchTagValuesResponse]{
		httpStatusCode: 200,
		new:            func() *tracestorepb.SearchTagValuesResponse { return &tracestorepb.SearchTagValuesResponse{} },
		current:        &tracestorepb.SearchTagValuesResponse{TagValues: make([]string, 0)},
		combine: func(partial, final *tracestorepb.SearchTagValuesResponse, _ PipelineResponse) error {
			for _, v := range partial.TagValues {
				d.Collect(v)
			}
			return nil
		},
		finalize: func(final *tracestorepb.SearchTagValuesResponse) (*tracestorepb.SearchTagValuesResponse, error) {
			final.TagValues = d.Strings()
			return final, nil
		},
		quit: func(_ *tracestorepb.SearchTagValuesResponse) bool {
			return d.Exceeded()
		},
		diff: func(response *tracestorepb.SearchTagValuesResponse) (*tracestorepb.SearchTagValuesResponse, error) {
			response.TagValues = d.Diff()
			return response, nil
		},
	}
}

func NewTypedSearchTagValues(limitBytes int) GRPCCombiner[*tracestorepb.SearchTagValuesResponse] {
	return NewSearchTagValues(limitBytes).(GRPCCombiner[*tracestorepb.SearchTagValuesResponse])
}

func NewSearchTagValuesV2(limitBytes int) Combiner {
	// Distinct collector with no limit
	d := util.NewDistinctValueCollector(limitBytes, func(tv tracestorepb.TagValue) int { return len(tv.Type) + len(tv.Value) })

	return &genericCombiner[*tracestorepb.SearchTagValuesV2Response]{
		current: &tracestorepb.SearchTagValuesV2Response{TagValues: []*tracestorepb.TagValue{}},
		new:     func() *tracestorepb.SearchTagValuesV2Response { return &tracestorepb.SearchTagValuesV2Response{} },
		combine: func(partial, final *tracestorepb.SearchTagValuesV2Response, _ PipelineResponse) error {
			for _, v := range partial.TagValues {
				d.Collect(*v)
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
			return final, nil
		},
		quit: func(_ *tracestorepb.SearchTagValuesV2Response) bool {
			return d.Exceeded()
		},
		diff: func(response *tracestorepb.SearchTagValuesV2Response) (*tracestorepb.SearchTagValuesV2Response, error) {
			diff := d.Diff()
			response.TagValues = make([]*tracestorepb.TagValue, 0, len(diff))
			for _, v := range diff {
				v2 := v
				response.TagValues = append(response.TagValues, &v2)
			}
			return response, nil
		},
	}
}

func NewTypedSearchTagValuesV2(limitBytes int) GRPCCombiner[*tracestorepb.SearchTagValuesV2Response] {
	return NewSearchTagValuesV2(limitBytes).(GRPCCombiner[*tracestorepb.SearchTagValuesV2Response])
}
