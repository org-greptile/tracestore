package combiner

import (
	"example.com/acme/tracestore/pkg/api"
	"example.com/acme/tracestore/pkg/collector"
	"example.com/acme/tracestore/pkg/tracestorepb"
	"go.uber.org/atomic"
)

var (
	_ GRPCCombiner[*tracestorepb.SearchTagsResponse]   = (*genericCombiner[*tracestorepb.SearchTagsResponse])(nil)
	_ GRPCCombiner[*tracestorepb.SearchTagsV2Response] = (*genericCombiner[*tracestorepb.SearchTagsV2Response])(nil)
)

func NewSearchTags(limitBytes int) Combiner {
	d := collector.NewDistinctStringWithDiff(limitBytes)
	inspectedBytes := atomic.NewUint64(0)

	c := &genericCombiner[*tracestorepb.SearchTagsResponse]{
		httpStatusCode: 200,
		new:            func() *tracestorepb.SearchTagsResponse { return &tracestorepb.SearchTagsResponse{} },
		current:        &tracestorepb.SearchTagsResponse{TagNames: make([]string, 0)},
		combine: func(partial, _ *tracestorepb.SearchTagsResponse, _ PipelineResponse) error {
			for _, v := range partial.TagNames {
				d.Collect(v)
			}
			if partial.Metrics != nil {
				inspectedBytes.Add(partial.Metrics.InspectedBytes)
			}
			return nil
		},
		finalize: func(response *tracestorepb.SearchTagsResponse) (*tracestorepb.SearchTagsResponse, error) {
			response.TagNames = d.Strings()
			// return metrics with final results
			// TODO: merge with other metrics as well, when we have them, return only InspectedBytes for now
			response.Metrics = &tracestorepb.MetadataMetrics{InspectedBytes: inspectedBytes.Load()}
			return response, nil
		},
		quit: func(_ *tracestorepb.SearchTagsResponse) bool {
			return d.Exceeded()
		},
		diff: func(response *tracestorepb.SearchTagsResponse) (*tracestorepb.SearchTagsResponse, error) {
			resp, err := d.Diff()
			if err != nil {
				return nil, err
			}

			response.TagNames = resp
			// TODO: merge with other metrics as well, when we have them, return only InspectedBytes for now
			// return metrics with diff results
			response.Metrics = &tracestorepb.MetadataMetrics{InspectedBytes: inspectedBytes.Load()}
			return response, nil
		},
	}
	initHTTPCombiner(c, api.HeaderAcceptJSON)
	return c
}

func NewTypedSearchTags(limitBytes int) GRPCCombiner[*tracestorepb.SearchTagsResponse] {
	return NewSearchTags(limitBytes).(GRPCCombiner[*tracestorepb.SearchTagsResponse])
}

func NewSearchTagsV2(limitBytes int) Combiner {
	// Distinct collector map to collect scopes and scope values
	distinctValues := collector.NewScopedDistinctStringWithDiff(limitBytes)
	inspectedBytes := atomic.NewUint64(0)

	c := &genericCombiner[*tracestorepb.SearchTagsV2Response]{
		httpStatusCode: 200,
		new:            func() *tracestorepb.SearchTagsV2Response { return &tracestorepb.SearchTagsV2Response{} },
		current:        &tracestorepb.SearchTagsV2Response{Scopes: make([]*tracestorepb.SearchTagsV2Scope, 0)},
		combine: func(partial, _ *tracestorepb.SearchTagsV2Response, _ PipelineResponse) error {
			for _, res := range partial.GetScopes() {
				for _, tag := range res.Tags {
					distinctValues.Collect(res.Name, tag)
				}
			}
			if partial.Metrics != nil {
				inspectedBytes.Add(partial.Metrics.InspectedBytes)
			}
			return nil
		},
		finalize: func(final *tracestorepb.SearchTagsV2Response) (*tracestorepb.SearchTagsV2Response, error) {
			collected := distinctValues.Strings()
			final.Scopes = make([]*tracestorepb.SearchTagsV2Scope, 0, len(collected))

			for scope, vals := range collected {
				final.Scopes = append(final.Scopes, &tracestorepb.SearchTagsV2Scope{
					Name: scope,
					Tags: vals,
				})
			}
			// return metrics with final results
			// TODO: merge with other metrics as well, when we have them, return only InspectedBytes for now
			final.Metrics = &tracestorepb.MetadataMetrics{InspectedBytes: inspectedBytes.Load()}
			return final, nil
		},
		quit: func(_ *tracestorepb.SearchTagsV2Response) bool {
			return distinctValues.Exceeded()
		},
		diff: func(response *tracestorepb.SearchTagsV2Response) (*tracestorepb.SearchTagsV2Response, error) {
			collected, err := distinctValues.Diff()
			if err != nil {
				return nil, err
			}
			response.Scopes = make([]*tracestorepb.SearchTagsV2Scope, 0, len(collected))

			for scope, vals := range collected {
				response.Scopes = append(response.Scopes, &tracestorepb.SearchTagsV2Scope{
					Name: scope,
					Tags: vals,
				})
			}
			// TODO: merge with other metrics as well, when we have them, return only InspectedBytes for now
			// also return metrics with diff results
			response.Metrics = &tracestorepb.MetadataMetrics{InspectedBytes: inspectedBytes.Load()}
			return response, nil
		},
	}
	initHTTPCombiner(c, api.HeaderAcceptJSON)
	return c
}

func NewTypedSearchTagsV2(limitBytes int) GRPCCombiner[*tracestorepb.SearchTagsV2Response] {
	return NewSearchTagsV2(limitBytes).(GRPCCombiner[*tracestorepb.SearchTagsV2Response])
}
